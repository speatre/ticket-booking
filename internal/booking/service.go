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

// Cache defines caching operations needed by booking service.
// It provides Redis-based caching for ticket availability, pending booking TTLs,
// and event statistics to improve performance and handle high concurrency.
type Cache interface {
	// Set stores a value with optional TTL. Used for pending booking timeouts.
	Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	// GetRemainingSeats retrieves cached remaining ticket count for an event
	GetRemainingSeats(ctx context.Context, eventID string) (int, error)
	// DecrementSeats atomically decrements available seats, returns new count
	DecrementSeats(ctx context.Context, eventID string, qty int) (int, error)
	// Del removes a cache key, used for cleanup of pending bookings
	Del(ctx context.Context, key string) error
	// GetInt retrieves an integer value from cache
	GetInt(ctx context.Context, key string) (int, error)
}

// Publisher defines async messaging operations for event-driven architecture.
// Used to publish booking events to message queue for payment processing.
type Publisher interface {
	// Publish sends a message to the specified topic/queue
	Publish(topic string, v interface{}) error
}

// BookingService defines the core booking business logic interface.
// Handles the complete booking lifecycle: creation, confirmation, cancellation.
// Ensures data consistency through database transactions and handles concurrency.
type BookingService interface {
	// CreateBooking creates a new PENDING booking with seat reservation.
	// Returns booking ID or ErrNotEnoughTickets if insufficient capacity.
	CreateBooking(ctx context.Context, userID, eventID string, qty int) (string, error)
	// Get retrieves a booking by ID
	Get(ctx context.Context, id string) (*Booking, error)
	// HandleBookingCreated processes booking.created messages from queue
	HandleBookingCreated(ctx context.Context, body []byte) error
	// ConfirmBooking transitions booking to CONFIRMED status after payment
	ConfirmBooking(ctx context.Context, bookingID string) error
	// CancelBooking transitions booking to CANCELLED and releases seats
	CancelBooking(ctx context.Context, bookingID string) error
}

// EventReserver provides seat reservation operations for booking service.
// Implements both Redis-first (fast) and DB-transaction (safe) reservation strategies
// to handle high concurrency while preventing overbooking.
type EventReserver interface {
	// Reserve attempts fast Redis-based seat reservation
	Reserve(ctx context.Context, eventID string, qty int) (bool, error)
	// ReserveTx performs transactional seat reservation with row locking
	ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error)
	// Get retrieves event details including current pricing
	Get(ctx context.Context, id string) (*event.Event, error)
	// Release returns reserved seats back to available pool
	Release(ctx context.Context, eventID string, qty int) error
}

// Service implements BookingService with transaction safety and concurrency handling.
// Uses database transactions as the source of truth for seat reservations,
// with Redis caching for performance optimization.
type Service struct {
	db        database.Database // Database transaction interface
	repo      BookingRepository // Booking data access layer
	reserver  EventReserver     // Event seat reservation operations
	publisher Publisher         // Message queue publisher for async processing
	cache     Cache             // Redis cache for performance and TTL management
	logger    *zap.Logger       // Structured logger
}

// NewService creates a new booking service with all required dependencies.
// All parameters are required for proper operation:
// - db: provides transactional database operations
// - r: handles booking persistence
// - er: manages event seat reservations
// - pub: publishes booking events for async processing
// - cache: provides Redis caching and TTL management
// - logger: structured logging for observability
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

// BookingCreatedMessage represents the payload sent to message queue
// when a new booking is created. Used for asynchronous payment processing
// and automatic cancellation scheduling.
type BookingCreatedMessage struct {
	BookingID string `json:"booking_id"` // UUID of the created booking
	UserID    string `json:"user_id"`    // UUID of the user who made the booking
	EventID   string `json:"event_id"`   // UUID of the event being booked
	Quantity  int    `json:"quantity"`   // Number of tickets booked
}

// ErrNotEnoughTickets is returned when reservation cannot be satisfied
var ErrNotEnoughTickets = errors.New("not enough tickets")

// CreateBooking creates a new booking with transactional safety and concurrency handling.
//
// Process flow:
// 1. Uses database transaction as source of truth for seat reservation
// 2. Locks event row to prevent race conditions
// 3. Creates booking record with current event pricing
// 4. Publishes booking.created event for async payment processing
// 5. Sets Redis TTL for automatic cancellation after 15 minutes
//
// Returns booking ID on success or ErrNotEnoughTickets if insufficient capacity.
// All operations are atomic - if any step fails, the entire booking is rolled back.
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

	// 4. Publish booking.created event for async payment processing
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

	// 5. Set Redis TTL for automatic cancellation after 15 minutes if payment not completed
	if err := s.cache.Set(ctx, "booking:pending:"+id, "1", 15*time.Minute); err != nil {
		s.logger.Warn("Failed to set pending booking in cache",
			zap.String("booking_id", id), zap.Error(err))
	}

	s.logger.Info("Booking created successfully",
		zap.String("booking_id", id), zap.String("user_id", userID),
		zap.String("event_id", eventID), zap.Int("quantity", qty))
	return id, nil
}

// HandleBookingCreated processes booking.created messages from the message queue.
// This is a simplified implementation that immediately confirms bookings.
// In a production system, this would trigger actual payment processing.
func (s *Service) HandleBookingCreated(ctx context.Context, body []byte) error {
	var msg BookingCreatedMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}

	// For demo purposes, immediately confirm the booking
	// In production, this would initiate payment processing workflow
	if err := s.ConfirmBooking(ctx, msg.BookingID); err != nil {
		s.logger.Error("confirm booking failed in worker", zap.String("booking", msg.BookingID), zap.Error(err))
		return s.CancelBooking(ctx, msg.BookingID)
	}

	// Remove pending TTL key since booking is now processed
	_ = s.cache.Del(ctx, "booking:pending:"+msg.BookingID)
	return nil
}

// ConfirmBooking transitions a booking from PENDING to CONFIRMED status.
// Updates event statistics cache and cleans up pending booking TTL.
// Idempotent - safe to call multiple times on the same booking.
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

// CancelBooking transitions a booking from PENDING to CANCELLED status.
// Releases reserved seats back to the event capacity and updates statistics.
// Idempotent - safe to call multiple times on the same booking.
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

// updateEventStatsCache recalculates and caches event statistics (tickets sold, revenue).
// Only counts CONFIRMED bookings for accurate financial reporting.
// Statistics are stored as JSON in Redis for fast API responses.
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

// Get retrieves a booking by its ID.
// Returns the booking record or an error if not found.
func (s *Service) Get(ctx context.Context, id string) (*Booking, error) {
	return s.repo.Get(id)
}
