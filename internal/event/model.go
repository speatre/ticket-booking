// Package event provides event management functionality including
// CRUD operations, seat reservations, and statistics tracking.
package event

import "time"

// Event represents a ticketed event with capacity management.
// Tracks both total capacity and remaining available tickets for real-time availability.
type Event struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name             string    `gorm:"type:text;not null" json:"name"`                                         // Event title
	Description      *string   `gorm:"type:text" json:"description,omitempty"`                                 // Optional event details
	StartsAt         time.Time `json:"starts_at"`                                                              // Event start time (UTC)
	EndsAt           time.Time `json:"ends_at"`                                                                // Event end time (UTC)
	Capacity         int       `gorm:"not null" json:"capacity"`                                               // Total tickets available (immutable after creation)
	Remaining        int       `gorm:"not null" json:"remaining"`                                              // Current available tickets (decreases with bookings)
	TicketPriceCents int64     `gorm:"column:ticket_price_cents;not null;default:0" json:"ticket_price_cents"` // Price per ticket in cents for precision
	CreatedAt        time.Time `json:"created_at"`                                                             // Event creation timestamp
	UpdatedAt        time.Time `json:"updated_at"`                                                             // Last modification timestamp
}
