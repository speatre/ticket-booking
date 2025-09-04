package event

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	svc    ServiceInterface
	logger *zap.Logger
}

func NewHandler(s ServiceInterface, logger *zap.Logger) *Handler {
	return &Handler{svc: s, logger: logger}
}

// List godoc
// @Summary List events
// @Description Get all available events
// @Tags events
// @Produce json
// @Param limit query int false "Max items to return (default 20, max 100)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} EventResponse
// @Failure 500 {object} ErrorResponse
// @Router /events [get]
func (h *Handler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	evts, err := h.svc.ListPage(c, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	// map to response
	out := make([]EventResponse, 0, len(evts))
	for i := range evts {
		out = append(out, eventToResponse(&evts[i]))
	}
	h.logger.Info("Events listed", zap.Int("count", len(evts)), zap.Int("limit", limit), zap.Int("offset", offset))
	c.JSON(http.StatusOK, out)
}

// Get godoc
// @Summary Get event by ID
// @Description Retrieve a single event by its ID
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} EventResponse
// @Failure 404 {object} ErrorResponse
// @Router /events/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	evt, err := h.svc.Get(c, id)
	if err != nil {
		h.logger.Error("Failed to get event", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}
	h.logger.Info("Event retrieved", zap.String("event_id", id))
	c.JSON(http.StatusOK, eventToResponse(evt))
}

// Stats godoc
// @Summary Event statistics
// @Description Get total tickets sold and estimated revenue for an event
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /events/{id}/stats [get]
func (h *Handler) Stats(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.svc.Get(c, id); err != nil {
		h.logger.Error("Failed to get event for stats", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}
	tickets, revenueCents, err := h.svc.StatsDB(c, id)
	if err != nil {
		h.logger.Error("Failed to compute stats", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tickets_sold": tickets,
		"revenue":      float64(revenueCents) / 100.0,
	})
}

// Create godoc
// @Summary Create event
// @Description Create a new event (Admin only)
// @Tags events
// @Accept json
// @Produce json
// @Param input body CreateEventRequest true "Event data"
// @Success 201 {object} EventResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/events [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid event creation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	e := &Event{
		Name:             req.Name,
		Description:      req.Description,
		StartsAt:         req.StartsAt,
		EndsAt:           req.EndsAt,
		Capacity:         req.Capacity,
		Remaining:        req.Capacity,
		TicketPriceCents: req.TicketPriceCents,
	}
	if err := h.svc.Create(c, e); err != nil {
		h.logger.Error("Failed to create event", zap.String("event_id", e.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	h.logger.Info("Event created", zap.String("event_id", e.ID))
	c.JSON(http.StatusCreated, eventToResponse(e))
}

// Update godoc
// @Summary Update event
// @Description Update event details (Admin only)
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param input body UpdateEventRequest true "Updated event data"
// @Success 200 {object} EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/events/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid event update request", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	// Load existing to avoid resetting remaining to capacity
	existing, err := h.svc.Get(c, id)
	if err != nil {
		h.logger.Error("Failed to get event for update", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}
	e := &Event{
		ID:               id,
		Name:             existing.Name,
		Description:      existing.Description,
		StartsAt:         existing.StartsAt,
		EndsAt:           existing.EndsAt,
		Capacity:         existing.Capacity,
		Remaining:        existing.Remaining,
		TicketPriceCents: existing.TicketPriceCents,
	}
	if req.Name != nil {
		e.Name = *req.Name
	}
	if req.Description != nil {
		e.Description = req.Description
	}
	if req.StartsAt != nil {
		e.StartsAt = *req.StartsAt
	}
	if req.EndsAt != nil {
		e.EndsAt = *req.EndsAt
	}
	if req.Capacity != nil {
		// adjust remaining only if capacity increased and remaining less than new capacity
		if *req.Capacity >= e.Remaining {
			e.Remaining = minInt(e.Remaining, *req.Capacity)
		}
		e.Capacity = *req.Capacity
	}
	if req.TicketPriceCents != nil {
		e.TicketPriceCents = *req.TicketPriceCents
	}
	if err := h.svc.Update(c, e); err != nil {
		h.logger.Error("Failed to update event", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	h.logger.Info("Event updated", zap.String("event_id", id))
	c.JSON(http.StatusOK, eventToResponse(e))
}

// Delete godoc
// @Summary Delete event
// @Description Delete an event (Admin only)
// @Tags events
// @Param id path string true "Event ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/events/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c, id); err != nil {
		h.logger.Error("Failed to delete event", zap.String("event_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	h.logger.Info("Event deleted", zap.String("event_id", id))
	c.Status(http.StatusNoContent)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func eventToResponse(e *Event) EventResponse {
	return EventResponse{
		ID:           e.ID,
		Name:         e.Name,
		Description:  e.Description,
		DateTime:     e.StartsAt,
		TotalTickets: e.Capacity,
		TicketPrice:  float64(e.TicketPriceCents) / 100.0,
		Remaining:    e.Remaining,
	}
}
