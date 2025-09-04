package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface để dễ test/mocking
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	Del(ctx context.Context, key string) error
	IncrBy(ctx context.Context, key string, n int) (int, error)
	DecrementSeats(ctx context.Context, eventID string, qty int) (int, error)
	GetRemainingSeats(ctx context.Context, eventID string) (int, error)
	GetEventIDs(ctx context.Context) ([]string, error)
	Close() error
}

// Redis implement Cache interface
type Redis struct {
	client *redis.Client
}

// MustOpen mở kết nối Redis (panic nếu fail)
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

// Atomic decrement seats (prevent oversell)
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
