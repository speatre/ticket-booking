package booking

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/bookings", h.Create)
	r.GET("/bookings/:id", h.Get)
}
