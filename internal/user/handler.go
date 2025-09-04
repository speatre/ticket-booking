package user

import (
	"net/http"

	"ticket-booking/internal/auth"
	"ticket-booking/pkg/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles user-related HTTP requests
type Handler struct {
	svc    *Service
	cfg    *config.Security
	logger *zap.Logger
}

// NewHandler creates a new Handler
func NewHandler(s *Service, cfg *config.Security, logger *zap.Logger) *Handler {
	return &Handler{svc: s, cfg: cfg, logger: logger}
}

// ===== Register =====
// @Summary Register new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "User registration"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
		h.logger.Warn("Invalid registration request", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	id, err := h.svc.Register(c, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to register user", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info("User registration successful", zap.String("user_id", id), zap.String("email", req.Email))
	c.JSON(http.StatusCreated, RegisterResponse{UserID: id})
}

// ===== Login =====
// @Summary User login
// @Description Authenticate user and return JWT access & refresh tokens
// @Tags users
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
		h.logger.Warn("Invalid login request", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	u, err := h.svc.VerifyLogin(c, req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login attempt failed", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		return
	}

	tokens, err := auth.GenerateTokens(h.cfg, u.ID, u.Role)
	if err != nil {
		h.logger.Error("Failed to generate tokens", zap.String("user_id", u.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate tokens"})
		return
	}

	h.logger.Info("User login successful", zap.String("user_id", u.ID), zap.String("email", req.Email))
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// ===== RefreshToken =====
// @Summary Refresh access token
// @Description Use refresh token to get a new access token
// @Tags users
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	refreshClaims, err := auth.ValidateRefreshToken(h.cfg, req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid refresh token"})
		return
	}

	tokens, err := auth.GenerateTokens(h.cfg, refreshClaims.UserID, "")
	if err != nil {
		h.logger.Error("Failed to generate new tokens", zap.String("user_id", refreshClaims.UserID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate tokens"})
		return
	}

	h.logger.Info("Token refresh successful", zap.String("user_id", refreshClaims.UserID))
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// ===== UpdateProfile =====
// @Summary Update user profile
// @Description Update own user info (only the authenticated user can update self)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param input body UpdateProfileRequest true "Profile update"
// @Success 200 {object} UpdateProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	targetID := c.Param("id")
	callerID := c.GetString(auth.CtxUserID)

	if callerID == "" || callerID != targetID {
		h.logger.Warn("Unauthorized profile update attempt", zap.String("caller_id", callerID), zap.String("target_id", targetID))
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid profile update request", zap.String("user_id", targetID), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.svc.UpdateProfile(c, callerID, targetID, req.FullName); err != nil {
		h.logger.Error("Failed to update profile", zap.String("user_id", targetID), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info("Profile updated successfully", zap.String("user_id", targetID))
	c.JSON(http.StatusOK, UpdateProfileResponse{OK: true})
}
