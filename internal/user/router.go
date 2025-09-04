package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/users/register", h.Register)
	r.POST("/users/login", h.Login)
	r.POST("/users/refresh", h.RefreshToken)
	// PUT /users/:id is protected in central router with Authn
}

func RegisterProtectedRoutes(rg *gin.RouterGroup, h *Handler) {
	rg.PUT("/users/:id", h.UpdateProfile)
}
