package auth

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"ticket-booking/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Context keys
const (
	CtxUserID = "userID"
	CtxRole   = "role"
	CtxReqID  = "requestID"
)

// Middleware holds dependencies for middleware functions
type Middleware struct {
	logger       *zap.Logger      // Application logger for business logic
	accessLogger *zap.Logger      // Access logger for HTTP requests
	cfg          *config.Security
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware(logger *zap.Logger, accessLogger *zap.Logger, cfg *config.Security) *Middleware {
	return &Middleware{
		logger:       logger,
		accessLogger: accessLogger,
		cfg:          cfg,
	}
}

// RequestID sets a unique request ID in the context and response header
func (m *Middleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
			m.logger.Debug("Generated new request ID", zap.String("request_id", id))
		} else {
			m.logger.Debug("Received request ID from header", zap.String("request_id", id))
		}
		c.Set(CtxReqID, id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

// AccessLog logs HTTP requests with structured fields
func (m *Middleware) AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		latency := time.Since(start)
		reqID, _ := c.Get(CtxReqID)
		userID, _ := c.Get(CtxUserID)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("request_id", reqID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		if userID != nil {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		c.Writer.Header().Add("X-Response-Time", latency.String())

		if len(c.Errors) > 0 {
			m.accessLogger.Error("Request completed with errors", fields...)
		} else if status >= 500 {
			m.accessLogger.Error("Request completed with server error", fields...)
		} else if status >= 400 {
			m.accessLogger.Warn("Request completed with client error", fields...)
		} else {
			m.accessLogger.Info("Request completed", fields...)
		}
	}
}

// Authn validates access token from Authorization header
func (m *Middleware) Authn() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID, _ := c.Get(CtxReqID)
		ah := c.GetHeader("Authorization")
		if ah == "" || !strings.HasPrefix(ah, "Bearer ") {
			m.logger.Warn("Missing or invalid Authorization header",
				zap.String("request_id", reqID.(string)),
				zap.String("header", ah))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing/invalid Authorization"})
			return
		}
		token := strings.TrimPrefix(ah, "Bearer ")

		claims, err := ValidateAccessToken(m.cfg, token)
		if err != nil {
			m.logger.Warn("Invalid access token",
				zap.String("request_id", reqID.(string)),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		m.logger.Debug("Access token validated",
			zap.String("request_id", reqID.(string)),
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role))
		c.Next()
	}
}

// Authorize checks if user role is allowed
func (m *Middleware) Authorize(roles ...string) gin.HandlerFunc {
	allow := map[string]struct{}{}
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *gin.Context) {
		reqID, _ := c.Get(CtxReqID)
		userID, _ := c.Get(CtxUserID)
		role, ok := c.Get(CtxRole)
		if !ok || role.(string) == "" {
			m.logger.Warn("No role found in context",
				zap.String("request_id", reqID.(string)),
				zap.Any("user_id", userID))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role"})
			return
		}
		if _, ok := allow[role.(string)]; !ok {
			m.logger.Warn("Unauthorized role",
				zap.String("request_id", reqID.(string)),
				zap.String("user_id", userID.(string)),
				zap.String("role", role.(string)))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		m.logger.Debug("Role authorized",
			zap.String("request_id", reqID.(string)),
			zap.String("user_id", userID.(string)),
			zap.String("role", role.(string)))
		c.Next()
	}
}

// Rate limit: per-user (if authn ran before) else per-IP
var (
	limits   = map[string]*rate.Limiter{}
	limitsMu sync.Mutex
)

func limiter(key string, r rate.Limit, b int) *rate.Limiter {
	limitsMu.Lock()
	defer limitsMu.Unlock()
	if l, ok := limits[key]; ok {
		return l
	}
	l := rate.NewLimiter(r, b)
	limits[key] = l
	return l
}

type RatePlan struct {
	AnonRPS   float64
	AnonBurst int
	UserRPS   float64
	UserBurst int
}

// RateLimit enforces per-user or per-IP request limiting
func (m *Middleware) RateLimit(plan RatePlan) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID, _ := c.Get(CtxReqID)
		if uid, ok := c.Get(CtxUserID); ok && uid.(string) != "" {
			l := limiter("u:"+uid.(string), rate.Limit(plan.UserRPS), plan.UserBurst)
			if !l.Allow() {
				m.logger.Warn("User rate limit exceeded",
					zap.String("request_id", reqID.(string)),
					zap.String("user_id", uid.(string)),
					zap.Float64("rps", plan.UserRPS),
					zap.Int("burst", plan.UserBurst))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit (user)"})
				return
			}
			m.logger.Debug("User rate limit check passed",
				zap.String("request_id", reqID.(string)),
				zap.String("user_id", uid.(string)))
			c.Next()
			return
		}
		ip := c.ClientIP()
		l := limiter("a:"+ip, rate.Limit(plan.AnonRPS), plan.AnonBurst)
		if !l.Allow() {
			m.logger.Warn("Anonymous rate limit exceeded",
				zap.String("request_id", reqID.(string)),
				zap.String("client_ip", ip),
				zap.Float64("rps", plan.AnonRPS),
				zap.Int("burst", plan.AnonBurst))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit (anon)"})
			return
		}
		m.logger.Debug("Anonymous rate limit check passed",
			zap.String("request_id", reqID.(string)),
			zap.String("client_ip", ip))
		c.Next()
	}
}
