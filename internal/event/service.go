package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ticket-booking/pkg/cache"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceInterface defines event management operations with Redis caching and seat reservations.
// Provides both fast Redis-based operations and transactional database operations
// for high-performance ticket booking with data consistency guarantees.
type ServiceInterface interface {
	// Get retrieves a single event by ID
	Get(ctx context.Context, id string) (*Event, error)
	// ReserveTx performs transactional seat reservation with row locking (safe path)
	ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error)
	// Release returns reserved seats back to available pool
	Release(ctx context.Context, eventID string, qty int) error
	// List retrieves all events with Redis caching
	List(ctx context.Context) ([]Event, error)
	// Reserve attempts fast Redis-based seat reservation (fast path)
	Reserve(ctx context.Context, eventID string, qty int) (bool, error)
	// ListPage retrieves paginated events with per-page caching
	ListPage(ctx context.Context, limit, offset int) ([]Event, error)
	// Create creates a new event and initializes cache
	Create(ctx context.Context, e *Event) error
	// Update modifies an event and invalidates relevant cache
	Update(ctx context.Context, e *Event) error
	// Delete removes an event and cleans up cache
	Delete(ctx context.Context, id string) error
	// StatsDB calculates event statistics from database (CONFIRMED bookings only)
	StatsDB(ctx context.Context, eventID string) (tickets int64, revenueCents int64, err error)
}

// Service implements EventInterface with Redis caching for performance.
// Uses dual-path strategy: Redis for speed, database transactions for consistency.
type Service struct {
	db     *gorm.DB        // Database connection for transactions
	repo   EventRepository // Data access layer for events
	cache  cache.Cache     // Redis cache for performance optimization
	logger *zap.Logger     // Structured logger
}

// NewService creates a new event service with required dependencies.
// All parameters are required for proper caching and transaction handling.
func NewService(db *gorm.DB, r EventRepository, cache cache.Cache, logger *zap.Logger) *Service {
	return &Service{db: db, repo: r, cache: cache, logger: logger}
}

// List retrieves all events with Redis caching for improved performance.
// Cache TTL is 30 seconds to balance freshness with performance.
func (s *Service) List(ctx context.Context) ([]Event, error) {
	const cacheKey = "events:list"

	if raw, err := s.cache.Get(ctx, cacheKey); err == nil && raw != "" {
		var evts []Event
		if err := json.Unmarshal([]byte(raw), &evts); err == nil {
			s.logger.Info("Events retrieved from cache", zap.String("cache_key", cacheKey))
			return evts, nil
		}
		s.logger.Warn("Failed to unmarshal cached events", zap.String("cache_key", cacheKey), zap.Error(err))
	} else if err != nil {
		s.logger.Warn("Failed to get events from cache", zap.String("cache_key", cacheKey), zap.Error(err))
	}

	evts, err := s.repo.List()
	if err != nil {
		s.logger.Error("Failed to list events from database", zap.Error(err))
		return nil, err
	}

	if data, err := json.Marshal(evts); err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, 30*time.Second); err != nil {
			s.logger.Warn("Failed to cache events list", zap.String("cache_key", cacheKey), zap.Error(err))
		}
	} else {
		s.logger.Warn("Failed to marshal events for cache", zap.Error(err))
	}

	s.logger.Info("Events retrieved from database", zap.Int("count", len(evts)))
	return evts, nil
}

// ListPage returns paginated events with per-page Redis caching.
// Each page is cached separately to optimize common pagination patterns.
func (s *Service) ListPage(ctx context.Context, limit, offset int) ([]Event, error) {
	cacheKey := fmt.Sprintf("events:list:%d:%d", limit, offset)
	if raw, err := s.cache.Get(ctx, cacheKey); err == nil && raw != "" {
		var evts []Event
		if err := json.Unmarshal([]byte(raw), &evts); err == nil {
			s.logger.Info("Events page retrieved from cache", zap.String("cache_key", cacheKey))
			return evts, nil
		}
		s.logger.Warn("Failed to unmarshal cached events page", zap.String("cache_key", cacheKey), zap.Error(err))
	}

	evts, err := s.repo.ListPage(limit, offset)
	if err != nil {
		s.logger.Error("Failed to list events page from database", zap.Error(err))
		return nil, err
	}
	if data, err := json.Marshal(evts); err == nil {
		if err := s.cache.Set(ctx, cacheKey, data, 30*time.Second); err != nil {
			s.logger.Warn("Failed to cache events page", zap.String("cache_key", cacheKey), zap.Error(err))
		}
	}
	s.logger.Info("Events page retrieved from database", zap.Int("count", len(evts)), zap.Int("limit", limit), zap.Int("offset", offset))
	return evts, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Event, error) {
	event, err := s.repo.Get(id)
	if err != nil {
		s.logger.Error("Failed to get event", zap.String("event_id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Event retrieved", zap.String("event_id", id))
	return event, nil
}

func (s *Service) Create(ctx context.Context, e *Event) error {
	if err := s.repo.Create(e); err != nil {
		s.logger.Error("Failed to create event", zap.String("event_id", e.ID), zap.Error(err))
		return err
	}

	_ = s.cache.Set(ctx, "event:remaining:"+e.ID, e.Capacity, 0)
	_ = s.cache.Set(ctx, "event:revenue:"+e.ID, 0, 0)
	_ = s.cache.Del(ctx, "events:list")

	s.logger.Info("Event created", zap.String("event_id", e.ID))
	return nil
}

func (s *Service) Update(ctx context.Context, e *Event) error {
	if err := s.repo.Update(e); err != nil {
		s.logger.Error("Failed to update event", zap.String("event_id", e.ID), zap.Error(err))
		return err
	}

	_ = s.cache.Set(ctx, "event:remaining:"+e.ID, e.Remaining, 0)
	_ = s.cache.Del(ctx, "events:list")

	s.logger.Info("Event updated", zap.String("event_id", e.ID))
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete event", zap.String("event_id", id), zap.Error(err))
		return err
	}

	_ = s.cache.Del(ctx, "event:remaining:"+id)
	_ = s.cache.Del(ctx, "event:revenue:"+id)
	_ = s.cache.Del(ctx, "events:list")

	s.logger.Info("Event deleted", zap.String("event_id", id))
	return nil
}

// Reserve performs atomic seat reservation in Redis (fast path).
// Uses Redis DECRBY for atomic operations. Automatically rolls back on insufficient seats.
// This is the high-performance path but may have Redis-DB inconsistencies under failure scenarios.
func (s *Service) Reserve(ctx context.Context, eventID string, qty int) (bool, error) {
	remaining, err := s.cache.DecrementSeats(ctx, eventID, qty)
	if err != nil {
		s.logger.Error("Failed to decrement seats in cache", zap.String("event_id", eventID), zap.Int("quantity", qty), zap.Error(err))
		return false, err
	}
	if remaining < 0 {
		_, _ = s.cache.DecrementSeats(ctx, eventID, -qty) // rollback
		s.logger.Warn("Not enough seats in cache", zap.String("event_id", eventID), zap.Int("quantity", qty))
		return false, nil
	}

	s.logger.Info("Seats reserved (cache)", zap.String("event_id", eventID), zap.Int("qty", qty), zap.Int("remaining", remaining))
	return true, nil
}

// ReserveTx performs transactional seat reservation with row locking (safe path).
// Uses database row-level locking to prevent race conditions and ensure data consistency.
// This is the authoritative reservation method used during booking transactions.
func (s *Service) ReserveTx(tx *gorm.DB, eventID string, qty int) (bool, error) {
	ok, err := s.repo.Reserve(tx, eventID, qty)
	if err != nil {
		s.logger.Error("ReserveTx failed", zap.String("event_id", eventID), zap.Int("qty", qty), zap.Error(err))
		return false, err
	}
	if !ok {
		s.logger.Info("Not enough tickets (ReserveTx)", zap.String("event_id", eventID), zap.Int("qty", qty))
		return false, nil
	}

	// sync cache
	if ev, err := s.repo.Get(eventID); err == nil {
		_ = s.cache.Set(context.Background(), "event:remaining:"+eventID, ev.Remaining, 0)
	}
	return true, nil
}

func (s *Service) Stats(ctx context.Context, eventID string) (ticketsSold int, revenue int, err error) {
	ticketsSold, err = s.cache.GetInt(ctx, "event:sold:"+eventID)
	if err != nil {
		ticketsSold = 0
	}
	revenue, err = s.cache.GetInt(ctx, "event:revenue:"+eventID)
	if err != nil {
		revenue = 0
	}
	return
}

func (s *Service) Release(ctx context.Context, eventID string, qty int) error {
	if err := s.db.WithContext(ctx).Exec(
		"UPDATE events SET remaining = LEAST(remaining + ?, capacity) WHERE id = ?",
		qty, eventID,
	).Error; err != nil {
		return err
	}

	ev, err := s.repo.Get(eventID)
	if err == nil {
		_ = s.cache.Set(ctx, "event:remaining:"+eventID, ev.Remaining, 0)
		_ = s.cache.Del(ctx, "events:list")
	}
	return nil
}

// StatsDB computes tickets sold and revenue from DB (CONFIRMED only)
func (s *Service) StatsDB(ctx context.Context, eventID string) (tickets int64, revenueCents int64, err error) {
	if err = s.db.WithContext(ctx).Raw(
		"SELECT COALESCE(SUM(quantity),0) FROM bookings WHERE event_id = ? AND status = ?",
		eventID, "CONFIRMED",
	).Scan(&tickets).Error; err != nil {
		return
	}
	if err = s.db.WithContext(ctx).Raw(
		"SELECT COALESCE(SUM(quantity * unit_price_cents),0) FROM bookings WHERE event_id = ? AND status = ?",
		eventID, "CONFIRMED",
	).Scan(&revenueCents).Error; err != nil {
		return
	}
	return
}
