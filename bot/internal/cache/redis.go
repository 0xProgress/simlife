package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Client wraps the go-redis client with Simlife-specific helpers and centralized
// key management. All Redis operations in the bot should go through this wrapper
// to ensure consistent error handling, logging, and key naming.
type Client struct {
	rdb *redis.Client
	log zerolog.Logger
}

// NewClient initializes the Redis client and verifies connectivity with a PING.
func NewClient(addr string, log zerolog.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	// Verify connectivity at startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Client{
		rdb: rdb,
		log: log.With().Str("component", "cache").Logger(),
	}, nil
}

// Raw returns the underlying redis.Client for advanced operations not covered
// by the wrapper methods. Use sparingly — prefer the typed helpers.
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// Close gracefully closes the Redis connection pool.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Set stores a key-value pair with a TTL. Errors are logged but not propagated
// for non-critical cache writes (cache is a cache, not a source of truth).
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		c.log.Warn().Err(fmt.Errorf("redis set failed: %w", err)).
			Str("key", key).
			Msg("cache write failed")
		return err
	}
	return nil
}

// Get retrieves a value by key. Returns redis.Nil if the key does not exist.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// Del removes one or more keys. Returns the number of keys actually removed.
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Del(ctx, keys...).Result()
}

// IncrBy atomically increments a counter by the given amount.
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

// DecrBy atomically decrements a counter by the given amount.
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.DecrBy(ctx, key, value).Result()
}

// Exists checks if a key exists. Returns true if it does.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// TTL returns the remaining time-to-live for a key.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Expire sets a TTL on an existing key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// Eval executes a Lua script atomically. Used for sliding window rate limiters
// and other operations that require atomic read-modify-write semantics.
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rdb.Eval(ctx, script, keys, args...).Result()
}