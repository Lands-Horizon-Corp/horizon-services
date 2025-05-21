package horizon

import (
	"context"
	"time"
)

// Cache defines the interface for Redis operations
type Cache interface {
	// Start initializes the Redis connection pool
	Start(ctx context.Context) error

	// Stop gracefully shuts down all Redis connections
	Stop(ctx context.Context) error

	// Ping checks Redis server health
	Ping(ctx context.Context) error

	// Get retrieves a value by key from Redis
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with TTL expiration
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error
}
