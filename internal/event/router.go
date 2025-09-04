package event

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/events", h.List)
	r.GET("/events/:id", h.Get)
	// stats per event
	r.GET("/events/:id/stats", h.Stats)
}

func RegisterAdminRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/events", h.Create)
	r.PUT("/events/:id", h.Update)
	r.DELETE("/events/:id", h.Delete)
}
