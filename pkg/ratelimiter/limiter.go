package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

// RateLimiter handles the rate limiting logic
type RateLimiter struct {
	storage storage.Storage
	config  *Config
}

// New creates a new RateLimiter instance
func New(storage storage.Storage, config *Config) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		config:  config,
	}
}

// IsAllowed checks if a request should be allowed based on the key (IP or token)
func (r *RateLimiter) IsAllowed(ctx context.Context, key string, isToken bool) (bool, error) {
	// First check if the key is blocked
	blocked, err := r.storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if key is blocked: %w", err)
	}
	if blocked {
		return false, nil
	}

	// Get the appropriate limits for the key
	maxRequests := r.config.MaxRequestsPerSecond
	blockDuration := r.config.BlockDuration

	// If it's a token and we have specific limits for it, use those instead
	if isToken {
		if tokenConfig, exists := r.config.TokenLimits[key]; exists {
			maxRequests = tokenConfig.MaxRequestsPerSecond
			blockDuration = tokenConfig.BlockDuration
		}
	}

	// Increment the request count
	count, err := r.storage.IncrementRequestCount(ctx, key, time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment request count: %w", err)
	}

	// If we've exceeded the limit, block the key
	if count > int64(maxRequests) {
		err = r.storage.Block(ctx, key, blockDuration)
		if err != nil {
			return false, fmt.Errorf("failed to block key: %w", err)
		}
		return false, nil
	}

	return true, nil
}

// GetRemainingRequests returns the number of remaining requests allowed for a key
func (r *RateLimiter) GetRemainingRequests(ctx context.Context, key string, isToken bool) (int, error) {
	count, err := r.storage.GetRequestCount(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	maxRequests := r.config.MaxRequestsPerSecond
	if isToken {
		if tokenConfig, exists := r.config.TokenLimits[key]; exists {
			maxRequests = tokenConfig.MaxRequestsPerSecond
		}
	}

	remaining := maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
