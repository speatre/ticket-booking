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

func New(d Deps) *gin.Engine {
	r := gin.New()

	// ==== Global middlewares ====
	r.Use(gin.Recovery())      // panic recovery
	r.Use(d.AuthM.RequestID()) // request tracing
	r.Use(d.AuthM.AccessLog()) // structured logging

	// ==== Swagger Docs ====
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==== API v1 root ====
	api := r.Group("/api/v1")

	// ---- Public Routes ----
	user.RegisterRoutes(api, d.UserH)         // register, login, refresh
	event.RegisterPublicRoutes(api, d.EventH) // list, get events

	// ---- Rate Limiting (applies to all APIs below) ----
	api.Use(d.AuthM.RateLimit(auth.RatePlan{
		AnonRPS:   2,
		AnonBurst: 5,
		UserRPS:   10,
		UserBurst: 20,
	}))

	// ---- Protected Routes (JWT Access Token required) ----
	protected := api.Group("")
	protected.Use(d.AuthM.Authn()) // Use AuthM.Authn

	booking.RegisterRoutes(protected, d.BookingH)    // create, list bookings
	user.RegisterProtectedRoutes(protected, d.UserH) // update profile, etc.

	// ---- Admin Routes ----
	admin := api.Group("/admin")
	admin.Use(d.AuthM.Authn(), d.AuthM.Authorize(auth.RoleAdmin))
	event.RegisterAdminRoutes(admin, d.EventH) // event CRUD for admins

	return r
}
