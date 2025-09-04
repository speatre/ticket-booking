package event

import "time"

type Event struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name             string    `gorm:"type:text;not null" json:"name"`
	Description      *string   `gorm:"type:text" json:"description,omitempty"`
	StartsAt         time.Time `json:"starts_at"`
	EndsAt           time.Time `json:"ends_at"`
	Capacity         int       `gorm:"not null" json:"capacity"`
	Remaining        int       `gorm:"not null" json:"remaining"`
	TicketPriceCents int64     `gorm:"column:ticket_price_cents;not null;default:0" json:"ticket_price_cents"` // store cents
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
