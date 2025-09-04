// Package cache provides Redis-based caching and atomic operations for high-performance
// ticket booking. Implements seat reservation counters, event data caching,
// and TTL-based automatic cleanup for pending bookings.
package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines Redis operations for high-performance caching and atomic operations.
// Provides both general caching and specialized ticket booking operations.
type Cache interface {
	// Set stores a value with optional TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get retrieves a string value from cache
	Get(ctx context.Context, key string) (string, error)
	// GetInt retrieves an integer value from cache
	GetInt(ctx context.Context, key string) (int, error)
	// Del removes a key from cache
	Del(ctx context.Context, key string) error
	// IncrBy atomically increments a key by n
	IncrBy(ctx context.Context, key string, n int) (int, error)
	// DecrementSeats atomically decrements available seats for an event
	DecrementSeats(ctx context.Context, eventID string, qty int) (int, error)
	// GetRemainingSeats retrieves current available seats for an event
	GetRemainingSeats(ctx context.Context, eventID string) (int, error)
	// GetEventIDs retrieves all event IDs that have seat tracking in cache
	GetEventIDs(ctx context.Context) ([]string, error)
	// Close closes the Redis connection
	Close() error
}

// Redis implements the Cache interface using Redis as the backing store.
// Provides atomic operations critical for preventing ticket overbooking.
type Redis struct {
	client *redis.Client // Redis client instance
}

// MustOpen creates a new Redis connection and panics on failure.
// Used during application startup where Redis connectivity is required.
func MustOpen(addr string, db int) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return &Redis{client: rdb}
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *Redis) GetInt(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) IncrBy(ctx context.Context, key string, n int) (int, error) {
	res, err := r.client.IncrBy(ctx, key, int64(n)).Result()
	return int(res), err
}

// DecrementSeats atomically decrements available seats to prevent overbooking.
// Returns the new remaining count. Caller should check if result is negative and rollback if needed.
func (r *Redis) DecrementSeats(ctx context.Context, eventID string, qty int) (int, error) {
	key := "event:remaining:" + eventID
	res, err := r.client.DecrBy(ctx, key, int64(qty)).Result()
	return int(res), err
}

func (r *Redis) GetRemainingSeats(ctx context.Context, eventID string) (int, error) {
	return r.GetInt(ctx, "event:remaining:"+eventID)
}

func (r *Redis) GetEventIDs(ctx context.Context) ([]string, error) {
	var cursor uint64
	var keys []string
	var allKeys []string

	pattern := "event:remaining:*"

	for {
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	// strip prefix "event:remaining:"
	for i := range allKeys {
		allKeys[i] = allKeys[i][len("event:remaining:"):]
	}

	return allKeys, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
