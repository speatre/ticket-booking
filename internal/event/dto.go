package event

import (
	"time"
)

// CreateEventRequest input for creating a new event
type CreateEventRequest struct {
	Name             string    `json:"name" binding:"required" example:"Tech Conference 2025"`
	Description      *string   `json:"description" example:"A conference about future tech"`
	StartsAt         time.Time `json:"starts_at" example:"2025-09-01T09:00:00Z"`
	EndsAt           time.Time `json:"ends_at" example:"2025-09-01T17:00:00Z"`
	Capacity         int       `json:"capacity" binding:"required,min=1" example:"100"`
	TicketPriceCents int64     `json:"ticket_price_cents" binding:"required,min=0" example:"5000"`
}

// UpdateEventRequest input for updating event info
type UpdateEventRequest struct {
	Name             *string    `json:"name" example:"Updated Conference"`
	Description      *string    `json:"description" example:"Updated description"`
	StartsAt         *time.Time `json:"starts_at" example:"2025-09-02T09:00:00Z"`
	EndsAt           *time.Time `json:"ends_at" example:"2025-09-02T17:00:00Z"`
	Capacity         *int       `json:"capacity" binding:"gte=0" example:"150"`
	TicketPriceCents *int64     `json:"ticket_price_cents" binding:"gte=0" example:"6000"`
}

// EventResponse represents event output
type EventResponse struct {
	ID           string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string    `json:"name" example:"Tech Conference 2025"`
	Description  *string   `json:"description" example:"A conference about future tech"`
	DateTime     time.Time `json:"date_time" example:"2025-09-02T09:00:00+07:00"`
	TotalTickets int       `json:"total_tickets" example:"100"`
	TicketPrice  float64   `json:"ticket_price" example:"50.00"`
	Remaining    int       `json:"remaining" example:"95"`
}

// ErrorResponse standard error model
type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}
