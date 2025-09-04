package booking

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusCancelled Status = "CANCELLED"
)

type Booking struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         string    `gorm:"type:uuid;not null" json:"user_id"`
	EventID        string    `gorm:"type:uuid;not null" json:"event_id"`
	Quantity       int       `gorm:"not null" json:"quantity"`
	UnitPriceCents int64     `gorm:"column:unit_price_cents;not null" json:"unit_price_cents"` // store cents
	Status         Status    `gorm:"type:text;not null" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
