package booking

import (
	"context"

	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(tx *gorm.DB, b *Booking) error
	Get(id string) (*Booking, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
	ListConfirmedByEvent(ctx context.Context, eventID string) ([]*Booking, error)
	ListPendingOlderThan(ctx context.Context, cutoff string) ([]*Booking, error)
}

type repo struct{ db *gorm.DB }

func NewBookingRepository(db *gorm.DB) BookingRepository { return &repo{db} }

func (r *repo) Create(tx *gorm.DB, b *Booking) error {
	return tx.Create(b).Error
}

func (r *repo) Get(id string) (*Booking, error) {
	var b Booking
	if err := r.db.First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// UpdateStatus sets booking status to any of [Pending, Confirmed, Cancelled]
func (r *repo) UpdateStatus(ctx context.Context, id string, status Status) error {
	return r.db.WithContext(ctx).
		Model(&Booking{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ListConfirmedByEvent returns all confirmed bookings for a specific event
func (r *repo) ListConfirmedByEvent(ctx context.Context, eventID string) ([]*Booking, error) {
	var bookings []*Booking
	if err := r.db.WithContext(ctx).
		Where("event_id = ? AND status = ?", eventID, StatusConfirmed).
		Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

// ListPendingOlderThan returns PENDING bookings created before cutoff time (ISO string expected)
func (r *repo) ListPendingOlderThan(ctx context.Context, cutoff string) ([]*Booking, error) {
	var bookings []*Booking
	if err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", StatusPending, cutoff).
		Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}
