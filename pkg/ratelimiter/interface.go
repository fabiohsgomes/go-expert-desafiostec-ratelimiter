package ratelimiter

import "context"

// RateLimiterInterface defines the interface for rate limiting functionality
type RateLimiterInterface interface {
	IsAllowed(ctx context.Context, key string, isToken bool) (bool, error)
	GetRemainingRequests(ctx context.Context, key string, isToken bool) (int, error)
}