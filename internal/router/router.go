package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"ticket-booking/internal/auth"
	"ticket-booking/internal/booking"
	"ticket-booking/internal/event"
	"ticket-booking/internal/user"
	"ticket-booking/pkg/config"

	_ "ticket-booking/docs" // swagger docs
)

// Deps aggregates all handlers and cross-cutting dependencies
type Deps struct {
	UserH    *user.Handler
	EventH   *event.Handler
	BookingH *booking.Handler
	Cfg      *config.Security
	AuthM    *auth.Middleware
}

// New creates a new Gin router with middleware, rate limiting, and route registration.
// Sets up a complete API server with authentication, authorization, and observability.
func New(d Deps) *gin.Engine {
	r := gin.New()

	// Global middlewares
	r.Use(gin.Recovery())      // Panic recovery
	r.Use(d.AuthM.RequestID()) // Request tracing
	r.Use(d.AuthM.AccessLog()) // Structured access logging

	// API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 root group
	api := r.Group("/api/v1")

	// Public routes (no authentication required)
	user.RegisterRoutes(api, d.UserH)
	event.RegisterPublicRoutes(api, d.EventH)

	// Rate limiting for all subsequent routes
	api.Use(d.AuthM.RateLimit(auth.RatePlan{
		AnonRPS:   2,  // Anonymous: 2 requests per second
		AnonBurst: 5,  // Anonymous: burst of 5 requests
		UserRPS:   10, // Authenticated: 10 requests per second
		UserBurst: 20, // Authenticated: burst of 20 requests
	}))

	// Protected routes (JWT authentication required)
	protected := api.Group("")
	protected.Use(d.AuthM.Authn())

	booking.RegisterRoutes(protected, d.BookingH)
	user.RegisterProtectedRoutes(protected, d.UserH)

	// Admin-only routes (authentication + admin role required)
	admin := api.Group("/admin")
	admin.Use(d.AuthM.Authn(), d.AuthM.Authorize(auth.RoleAdmin))
	event.RegisterAdminRoutes(admin, d.EventH)

	return r
}
