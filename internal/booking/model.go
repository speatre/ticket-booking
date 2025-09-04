// Package booking provides the core booking domain models and business logic
// for the ticket booking system.
package booking

import "time"

// Status represents the lifecycle states of a booking.
// Bookings transition: PENDING -> CONFIRMED (on payment) or CANCELLED (on timeout/failure)
type Status string

const (
	// StatusPending indicates a booking is created but payment not yet processed
	StatusPending Status = "PENDING"
	// StatusConfirmed indicates payment was successful and tickets are secured
	StatusConfirmed Status = "CONFIRMED"
	// StatusCancelled indicates booking was cancelled due to payment failure or timeout
	StatusCancelled Status = "CANCELLED"
)

// Booking represents a ticket reservation for an event.
// Captures pricing at booking time to handle price changes gracefully.
type Booking struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         string    `gorm:"type:uuid;not null" json:"user_id"`                        // Foreign key to users table
	EventID        string    `gorm:"type:uuid;not null" json:"event_id"`                       // Foreign key to events table
	Quantity       int       `gorm:"not null" json:"quantity"`                                 // Number of tickets booked (must be > 0)
	UnitPriceCents int64     `gorm:"column:unit_price_cents;not null" json:"unit_price_cents"` // Price per ticket in cents (captured at booking time)
	Status         Status    `gorm:"type:text;not null" json:"status"`                         // Current booking state
	CreatedAt      time.Time `json:"created_at"`                                               // When booking was created
	UpdatedAt      time.Time `json:"updated_at"`                                               // Last status change timestamp
}
