package booking

import (
	"net/http"

	"ticket-booking/internal/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"errors"
)

type Handler struct {
	svc    BookingService
	logger *zap.Logger
}

func NewHandler(s BookingService, logger *zap.Logger) *Handler {
	return &Handler{svc: s, logger: logger}
}

// Create godoc
// @Summary Create booking
// @Description Create a booking for an event (only authenticated users)
// @Tags bookings
// @Accept json
// @Produce json
// @Param input body CreateBookingRequest true "Booking request"
// @Success 201 {object} CreateBookingResponse
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 409 {object} ErrorResponse "Conflict (e.g., overbooking)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid booking creation request", zap.Error(err), zap.String("event_id", req.EventID))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	userID := c.GetString(auth.CtxUserID)
	if userID == "" {
		h.logger.Warn("Missing user ID for booking creation", zap.String("event_id", req.EventID))
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	id, err := h.svc.CreateBooking(c, userID, req.EventID, req.Quantity)
	if err != nil {
		if errors.Is(err, ErrNotEnoughTickets) {
			h.logger.Warn("Not enough tickets", zap.String("user_id", userID), zap.String("event_id", req.EventID), zap.Int("quantity", req.Quantity))
			c.JSON(http.StatusConflict, ErrorResponse{Error: "not enough tickets"})
			return
		}
		h.logger.Error("Failed to create booking", zap.String("user_id", userID), zap.String("event_id", req.EventID), zap.Int("quantity", req.Quantity), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	h.logger.Info("Booking created", zap.String("booking_id", id), zap.String("user_id", userID), zap.String("event_id", req.EventID), zap.Int("quantity", req.Quantity))
	c.JSON(http.StatusCreated, CreateBookingResponse{BookingID: id, Status: StatusPending})
}

// Get godoc
// @Summary Get booking
// @Description Get booking details by ID (only authenticated users)
// @Tags bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} BookingResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /bookings/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(auth.CtxUserID)
	if userID == "" {
		h.logger.Warn("Missing user ID for booking retrieval", zap.String("booking_id", id))
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	b, err := h.svc.Get(c, id)
	if err != nil {
		h.logger.Error("Failed to get booking", zap.String("booking_id", id), zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}
	h.logger.Info("Booking retrieved", zap.String("booking_id", id), zap.String("user_id", userID), zap.String("event_id", b.EventID))
	c.JSON(http.StatusOK, b)
}
