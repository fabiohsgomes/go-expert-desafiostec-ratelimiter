package ratelimiter

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadTokenLimitsFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	// Clear environment
	os.Clearenv()

	t.Run("Load valid token limits", func(t *testing.T) {
		os.Setenv("TOKEN_LIMIT_ABC123", "100:5m")
		os.Setenv("TOKEN_LIMIT_XYZ789", "50:10m")
		os.Setenv("TOKEN_LIMIT_PREMIUM", "1000:1m")

		config := NewConfig()
		config.LoadTokenLimitsFromEnv()

		// Check ABC123 token
		if limit, exists := config.TokenLimits["ABC123"]; !exists {
			t.Error("ABC123 token limit not loaded")
		} else {
			if limit.MaxRequestsPerSecond != 100 {
				t.Errorf("Expected 100 requests, got %d", limit.MaxRequestsPerSecond)
			}
			if limit.BlockDuration != time.Minute*5 {
				t.Errorf("Expected 5m duration, got %v", limit.BlockDuration)
			}
		}

		// Check XYZ789 token
		if limit, exists := config.TokenLimits["XYZ789"]; !exists {
			t.Error("XYZ789 token limit not loaded")
		} else {
			if limit.MaxRequestsPerSecond != 50 {
				t.Errorf("Expected 50 requests, got %d", limit.MaxRequestsPerSecond)
			}
			if limit.BlockDuration != time.Minute*10 {
				t.Errorf("Expected 10m duration, got %v", limit.BlockDuration)
			}
		}
	})

	t.Run("Ignore invalid formats", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TOKEN_LIMIT_INVALID1", "100")        // Missing duration
		os.Setenv("TOKEN_LIMIT_INVALID2", "abc:5m")     // Invalid requests
		os.Setenv("TOKEN_LIMIT_INVALID3", "100:invalid") // Invalid duration
		os.Setenv("TOKEN_LIMIT_VALID", "200:2m")

		config := NewConfig()
		config.LoadTokenLimitsFromEnv()

		// Should only load the valid one
		if len(config.TokenLimits) != 1 {
			t.Errorf("Expected 1 token limit, got %d", len(config.TokenLimits))
		}

		if limit, exists := config.TokenLimits["VALID"]; !exists {
			t.Error("VALID token limit not loaded")
		} else {
			if limit.MaxRequestsPerSecond != 200 {
				t.Errorf("Expected 200 requests, got %d", limit.MaxRequestsPerSecond)
			}
		}
	})

	t.Run("No token limits in environment", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("OTHER_VAR", "value")

		config := NewConfig()
		config.LoadTokenLimitsFromEnv()

		if len(config.TokenLimits) != 0 {
			t.Errorf("Expected 0 token limits, got %d", len(config.TokenLimits))
		}
	})
}