package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ticket-booking/internal/database"
	"ticket-booking/internal/event"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Cache defines caching operations needed by booking service
type Cache interface {
	Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	GetRemainingSeats(ctx context.Context, eventID string) (int, error)
	DecrementSeats(ctx context.Context, eventID string, qty int) (int, error)
	Del(ctx context.Context, key string) error
	GetInt(ctx context.Context, key string) (int, error)
}

// Publisher defines async messaging operations
type Publisher interface {
	Publish(topic string, v interface{}) error
}

// BookingService defines the interface for booking service operations
type BookingService interface {
	CreateBooking(ctx context.Context, userID, eventID string, qty int) (string, error)
	Get(ctx context.Context, id string) (*Booking, error)
	HandleBookingCreated(ctx context.Context, body []byte) error
	ConfirmBooking(ctx context.Context, bookingID string) error
	CancelBooking(ctx context.Context, bookingID string) error
}

// EventReserver is a contract for booking to interact with event service/repo
type EventReserver interface {
	Reserve(ctx context.Context, eventID string, qty int) (bool, error) // Redis-first
	ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error)       // DB-first fallback
	Get(ctx context.Context, id string) (*event.Event, error)
	Release(ctx context.Context, eventID string, qty int) error
}

// Service is the concrete booking service
type Service struct {
	db        database.Database
	repo      BookingRepository
	reserver  EventReserver
	publisher Publisher
	cache     Cache
	logger    *zap.Logger
}

func NewService(db database.Database, r BookingRepository, er EventReserver, pub Publisher, cache Cache, logger *zap.Logger) *Service {
	return &Service{
		db:        db,
		repo:      r,
		reserver:  er,
		publisher: pub,
		cache:     cache,
		logger:    logger,
	}
}

// Ensure *Service implements BookingService
var _ BookingService = (*Service)(nil)

// --- BookingCreatedMessage payload sent via MQ ---
type BookingCreatedMessage struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	EventID   string `json:"event_id"`
	Quantity  int    `json:"quantity"`
}

// ErrNotEnoughTickets is returned when reservation cannot be satisfied
var ErrNotEnoughTickets = errors.New("not enough tickets")

// --- CreateBooking: create PENDING booking and publish ---
// Redis-first reservation, fallback to DB-tx if Redis not reliable
func (s *Service) CreateBooking(ctx context.Context, userID, eventID string, qty int) (string, error) {
	var id string

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reserve using DB transaction as the source of truth
		okDB, errDB := s.reserver.ReserveTx(tx, eventID, qty)
		if errDB != nil {
			s.logger.Error("DB reservation failed",
				zap.String("event_id", eventID), zap.Int("quantity", qty), zap.Error(errDB))
			return errDB
		}
		if !okDB {
			s.logger.Warn("Not enough tickets (DB check)",
				zap.String("event_id", eventID), zap.Int("quantity", qty))
			return ErrNotEnoughTickets
		}

		// 2. Get event to copy current ticket price (cents)
		ev, err := s.reserver.Get(ctx, eventID)
		if err != nil {
			s.logger.Error("Failed to load event for booking",
				zap.String("event_id", eventID), zap.Error(err))
			return err
		}

		// 3. Create booking record with unit price cents
		b := &Booking{
			UserID:         userID,
			EventID:        eventID,
			Quantity:       qty,
			UnitPriceCents: ev.TicketPriceCents,
			Status:         StatusPending,
		}
		if err := s.repo.Create(tx, b); err != nil {
			s.logger.Error("Failed to create booking",
				zap.String("user_id", userID), zap.String("event_id", eventID),
				zap.Int("quantity", qty), zap.Error(err))
			return err
		}
		id = b.ID
		return nil
	})
	if err != nil {
		return "", err
	}

	// 4. Publish booking.created event
	msg := BookingCreatedMessage{
		BookingID: id,
		UserID:    userID,
		EventID:   eventID,
		Quantity:  qty,
	}
	if err := s.publisher.Publish("booking.created", msg); err != nil {
		s.logger.Error("Failed to publish booking created message",
			zap.String("booking_id", id), zap.Error(err))
		return "", err
	}

	// 5. Set TTL for pending booking in Redis (simulate payment timeout)
	if err := s.cache.Set(ctx, "booking:pending:"+id, "1", 15*time.Minute); err != nil {
		s.logger.Warn("Failed to set pending booking in cache",
			zap.String("booking_id", id), zap.Error(err))
	}

	s.logger.Info("Booking created successfully",
		zap.String("booking_id", id), zap.String("user_id", userID),
		zap.String("event_id", eventID), zap.Int("quantity", qty))
	return id, nil
}

// --- HandleBookingCreated: worker processing ---
func (s *Service) HandleBookingCreated(ctx context.Context, body []byte) error {
	var msg BookingCreatedMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}

	// Simulate payment succeeded directly here: confirm, otherwise cancel.
	// External worker can still handle real payments; this is a simple path.
	if err := s.ConfirmBooking(ctx, msg.BookingID); err != nil {
		s.logger.Error("confirm booking failed in worker", zap.String("booking", msg.BookingID), zap.Error(err))
		return s.CancelBooking(ctx, msg.BookingID)
	}

	// clean pending key
	_ = s.cache.Del(ctx, "booking:pending:"+msg.BookingID)
	return nil
}

// --- ConfirmBooking update DB, Redis and metrics ---
func (s *Service) ConfirmBooking(ctx context.Context, bookingID string) error {
	b, err := s.repo.Get(bookingID)
	if err != nil {
		s.logger.Error("ConfirmBooking: get booking failed", zap.String("booking_id", bookingID), zap.Error(err))
		return err
	}
	if b.Status == StatusConfirmed {
		return nil
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, StatusConfirmed); err != nil {
		s.logger.Error("ConfirmBooking: update status failed", zap.String("booking_id", bookingID), zap.Error(err))
		return err
	}

	// update event stats cache (tickets sold + revenue)
	if err := s.updateEventStatsCache(ctx, b.EventID); err != nil {
		s.logger.Warn("ConfirmBooking: update stats cache failed", zap.String("event_id", b.EventID), zap.Error(err))
	}

	// delete pending key if exists
	_ = s.cache.Del(ctx, "booking:pending:"+bookingID)

	s.logger.Info("Booking confirmed", zap.String("booking_id", bookingID), zap.String("event_id", b.EventID))
	return nil
}

// --- CancelBooking update DB, Redis and metrics ---
func (s *Service) CancelBooking(ctx context.Context, bookingID string) error {
	b, err := s.repo.Get(bookingID)
	if err != nil {
		s.logger.Error("CancelBooking: get booking failed", zap.String("booking_id", bookingID), zap.Error(err))
		return err
	}
	if b.Status == StatusCancelled {
		return nil
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, StatusCancelled); err != nil {
		s.logger.Error("CancelBooking: update status failed", zap.String("booking_id", bookingID), zap.Error(err))
		return err
	}

	// release seats in DB and sync cache via event reserver
	if err := s.reserver.Release(ctx, b.EventID, b.Quantity); err != nil {
		s.logger.Warn("CancelBooking: failed to release seats via reserver", zap.String("event_id", b.EventID), zap.Int("qty", b.Quantity), zap.Error(err))
	}

	// update stats cache as well
	if err := s.updateEventStatsCache(ctx, b.EventID); err != nil {
		s.logger.Warn("CancelBooking: update stats cache failed", zap.String("event_id", b.EventID), zap.Error(err))
	}

	// remove pending key if any
	_ = s.cache.Del(ctx, "booking:pending:"+bookingID)

	s.logger.Info("Booking cancelled", zap.String("booking_id", bookingID), zap.String("event_id", b.EventID))
	return nil
}

// --- updateEventStatsCache: recalculate tickets sold and revenue, save to Redis ---
func (s *Service) updateEventStatsCache(ctx context.Context, eventID string) error {
	var tickets int64
	var revenueCents int64

	// total tickets sold (CONFIRMED)
	if err := s.db.WithContext(ctx).Model(&Booking{}).
		Where("event_id = ? AND status = ?", eventID, StatusConfirmed).
		Select("COALESCE(SUM(quantity),0)").Scan(&tickets).Error; err != nil {
		s.logger.Error("Failed to calculate tickets sold", zap.String("event_id", eventID), zap.Error(err))
		return err
	}

	// total revenue in cents = SUM(quantity * unit_price_cents)
	if err := s.db.WithContext(ctx).Model(&Booking{}).
		Where("event_id = ? AND status = ?", eventID, StatusConfirmed).
		Select("COALESCE(SUM(quantity * unit_price_cents),0)").Scan(&revenueCents).Error; err != nil {
		s.logger.Error("Failed to calculate revenue cents", zap.String("event_id", eventID), zap.Error(err))
		return err
	}

	// convert to float for presentation (optional)
	revenue := float64(revenueCents) / 100.0

	stats := map[string]interface{}{
		"tickets_sold": tickets,
		"revenue":      revenue,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		s.logger.Warn("Failed to marshal event stats for cache", zap.String("event_id", eventID), zap.Error(err))
		return err
	}

	cacheKey := fmt.Sprintf("event:%s:stats", eventID)
	if err := s.cache.Set(ctx, cacheKey, string(data), 0); err != nil {
		s.logger.Warn("Failed to set event stats in cache", zap.String("event_id", eventID), zap.String("cache_key", cacheKey), zap.Error(err))
		return err
	}

	s.logger.Info("Event stats updated in cache", zap.String("event_id", eventID), zap.Int64("tickets_sold", tickets), zap.Float64("revenue", revenue))
	return nil
}

// --- Get booking by id ---
func (s *Service) Get(ctx context.Context, id string) (*Booking, error) {
	return s.repo.Get(id)
}
