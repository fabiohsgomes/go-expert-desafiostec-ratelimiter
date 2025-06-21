package storage

import (
	"context"
	"time"
)

// Storage defines the interface for rate limiter storage implementations
type Storage interface {
	// GetRequestCount returns the current request count for a key
	GetRequestCount(ctx context.Context, key string) (int64, error)

	// IncrementRequestCount increments the request count for a key
	IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int64, error)

	// IsBlocked checks if a key is currently blocked
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Block sets a block on a key for the specified duration
	Block(ctx context.Context, key string, duration time.Duration) error

	// Close closes the storage connection
	Close() error
}