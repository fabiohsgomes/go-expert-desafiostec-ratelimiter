package ratelimiter

import (
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	// Test default values
	if cfg.MaxRequestsPerSecond != 10 {
		t.Errorf("MaxRequestsPerSecond = %v, want %v", cfg.MaxRequestsPerSecond, 10)
	}
	if cfg.BlockDuration != time.Minute*5 {
		t.Errorf("BlockDuration = %v, want %v", cfg.BlockDuration, time.Minute*5)
	}
	if cfg.TokenHeader != "API_KEY" {
		t.Errorf("TokenHeader = %v, want %v", cfg.TokenHeader, "API_KEY")
	}
	if cfg.TokenLimits == nil {
		t.Error("TokenLimits is nil, want initialized map")
	}
	if len(cfg.TokenLimits) != 0 {
		t.Errorf("TokenLimits length = %v, want %v", len(cfg.TokenLimits), 0)
	}
}

func TestSetTokenLimit(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		maxRequests   int
		blockDuration time.Duration
	}{
		{
			name:          "Standard token limit",
			token:         "test-token",
			maxRequests:   100,
			blockDuration: time.Minute * 10,
		},
		{
			name:          "Zero max requests",
			token:         "zero-token",
			maxRequests:   0,
			blockDuration: time.Minute,
		},
		{
			name:          "Zero block duration",
			token:         "no-block-token",
			maxRequests:   50,
			blockDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.SetTokenLimit(tt.token, tt.maxRequests, tt.blockDuration)

			// Check if token limit was set correctly
			limit, exists := cfg.TokenLimits[tt.token]
			if !exists {
				t.Errorf("Token limit not found for token %v", tt.token)
				return
			}

			if limit.MaxRequestsPerSecond != tt.maxRequests {
				t.Errorf("MaxRequestsPerSecond = %v, want %v", limit.MaxRequestsPerSecond, tt.maxRequests)
			}
			if limit.BlockDuration != tt.blockDuration {
				t.Errorf("BlockDuration = %v, want %v", limit.BlockDuration, tt.blockDuration)
			}
		})
	}

	// Test overwriting existing token limit
	t.Run("Overwrite existing token limit", func(t *testing.T) {
		cfg := NewConfig()
		token := "test-token"

		// Set initial limit
		cfg.SetTokenLimit(token, 100, time.Minute)

		// Overwrite limit
		cfg.SetTokenLimit(token, 200, time.Minute*2)

		limit := cfg.TokenLimits[token]
		if limit.MaxRequestsPerSecond != 200 {
			t.Errorf("MaxRequestsPerSecond = %v, want %v", limit.MaxRequestsPerSecond, 200)
		}
		if limit.BlockDuration != time.Minute*2 {
			t.Errorf("BlockDuration = %v, want %v", limit.BlockDuration, time.Minute*2)
		}
	})
}