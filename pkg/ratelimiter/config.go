package ratelimiter

import "time"

// Config holds the rate limiter configuration
type Config struct {
	// MaxRequestsPerSecond is the maximum number of requests allowed per second
	MaxRequestsPerSecond int

	// BlockDuration is how long to block after exceeding the limit
	BlockDuration time.Duration

	// TokenHeader is the header key used for API tokens (default: "API_KEY")
	TokenHeader string

	// TokenLimits holds specific limits for tokens
	TokenLimits map[string]TokenConfig
}

// TokenConfig holds configuration for specific tokens
type TokenConfig struct {
	MaxRequestsPerSecond int
	BlockDuration       time.Duration
}

// NewConfig creates a new rate limiter configuration with default values
func NewConfig() *Config {
	return &Config{
		MaxRequestsPerSecond: 10,
		BlockDuration:       time.Minute * 5,
		TokenHeader:        "API_KEY",
		TokenLimits:       make(map[string]TokenConfig),
	}
}

// SetTokenLimit sets the rate limit for a specific token
func (c *Config) SetTokenLimit(token string, maxRequests int, blockDuration time.Duration) {
	c.TokenLimits[token] = TokenConfig{
		MaxRequestsPerSecond: maxRequests,
		BlockDuration:       blockDuration,
	}
}