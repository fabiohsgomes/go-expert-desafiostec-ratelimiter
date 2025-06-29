package ratelimiter

import (
	"os"
	"strconv"
	"strings"
	"time"
)

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

// LoadTokenLimitsFromEnv loads token limits from environment variables
// Format: TOKEN_LIMIT_<TOKEN>=<requests>:<duration>
// Example: TOKEN_LIMIT_ABC123=100:5m
func (c *Config) LoadTokenLimitsFromEnv() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TOKEN_LIMIT_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			
			token := strings.TrimPrefix(parts[0], "TOKEN_LIMIT_")
			limitParts := strings.Split(parts[1], ":")
			if len(limitParts) != 2 {
				continue
			}
			
			requests, err := strconv.Atoi(limitParts[0])
			if err != nil {
				continue
			}
			
			duration, err := time.ParseDuration(limitParts[1])
			if err != nil {
				continue
			}
			
			c.SetTokenLimit(token, requests, duration)
		}
	}
}