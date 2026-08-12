// Package redis holds the Redis-backed adapter used for response caching
// and distributed rate limiting.
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements domain.Cache on top of a Redis client.
type Cache struct {
	client *redis.Client
}

// New builds a Cache from connection settings and verifies it with a ping,
// so callers fail fast on startup instead of on the first command.
func New(ctx context.Context, addr string, db int) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return v, err
}

func (c *Cache) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	return c.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

// Increment atomically increments key and returns the new value. The TTL
// is only applied the moment the key is created (value becomes 1), so a
// caller can implement fixed-window counters (e.g. "N per day") by
// incrementing a time-bucketed key and reading the returned count.
func (c *Cache) Increment(ctx context.Context, key string, ttlSeconds int) (int64, error) {
	n, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := c.client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
			return n, err
		}
	}
	return n, nil
}
