package ratelimiter

import (
	"context"
	"testing"
	"time"
)

// mockStorage implements the storage.Storage interface for testing
type mockStorage struct {
	requestCounts map[string]int64
	blocked       map[string]bool
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		requestCounts: make(map[string]int64),
		blocked:       make(map[string]bool),
	}
}

func (m *mockStorage) GetRequestCount(ctx context.Context, key string) (int64, error) {
	return m.requestCounts[key], nil
}

func (m *mockStorage) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.requestCounts[key]++
	return m.requestCounts[key], nil
}

func (m *mockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *mockStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	m.blocked[key] = true
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestRateLimiter_IsAllowed(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		key          string
		isToken      bool
		requestCount int
		wantAllowed  bool
	}{
		{
			name: "IP under limit",
			config: &Config{
				MaxRequestsPerSecond: 5,
				BlockDuration:        time.Minute,
			},
			key:          "192.168.1.1",
			isToken:      false,
			requestCount: 3,
			wantAllowed:  true,
		},
		{
			name: "IP over limit",
			config: &Config{
				MaxRequestsPerSecond: 5,
				BlockDuration:        time.Minute,
			},
			key:          "192.168.1.2",
			isToken:      false,
			requestCount: 6,
			wantAllowed:  false,
		},
		{
			name: "Token under custom limit",
			config: &Config{
				MaxRequestsPerSecond: 5,
				BlockDuration:        time.Minute,
				TokenLimits: map[string]TokenConfig{
					"test-token": {
						MaxRequestsPerSecond: 10,
						BlockDuration:        time.Minute,
					},
				},
			},
			key:          "test-token",
			isToken:      true,
			requestCount: 8,
			wantAllowed:  true,
		},
		{
			name: "Token over custom limit",
			config: &Config{
				MaxRequestsPerSecond: 5,
				BlockDuration:        time.Minute,
				TokenLimits: map[string]TokenConfig{
					"test-token": {
						MaxRequestsPerSecond: 10,
						BlockDuration:        time.Minute,
					},
				},
			},
			key:          "test-token",
			isToken:      true,
			requestCount: 11,
			wantAllowed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockStorage()
			limiter := New(store, tt.config)
			ctx := context.Background()

			var lastAllowed bool
			for i := 0; i < tt.requestCount; i++ {
				var err error
				lastAllowed, err = limiter.IsAllowed(ctx, tt.key, tt.isToken)
				if err != nil {
					t.Fatalf("IsAllowed error: %v", err)
				}
			}

			if lastAllowed != tt.wantAllowed {
				t.Errorf("IsAllowed() = %v, want %v", lastAllowed, tt.wantAllowed)
			}
		})
	}
}

func TestRateLimiter_GetRemainingRequests(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		key           string
		isToken       bool
		requestCount  int
		wantRemaining int
	}{
		{
			name: "IP requests remaining",
			config: &Config{
				MaxRequestsPerSecond: 5,
			},
			key:           "192.168.1.1",
			isToken:       false,
			requestCount:  2,
			wantRemaining: 3,
		},
		{
			name: "Token requests remaining",
			config: &Config{
				MaxRequestsPerSecond: 5,
				TokenLimits: map[string]TokenConfig{
					"test-token": {
						MaxRequestsPerSecond: 10,
					},
				},
			},
			key:           "test-token",
			isToken:       true,
			requestCount:  4,
			wantRemaining: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockStorage()
			limiter := New(store, tt.config)
			ctx := context.Background()

			// Simulate requests
			for i := 0; i < tt.requestCount; i++ {
				_, err := limiter.IsAllowed(ctx, tt.key, tt.isToken)
				if err != nil {
					t.Fatalf("IsAllowed error: %v", err)
				}
			}

			remaining, err := limiter.GetRemainingRequests(ctx, tt.key, tt.isToken)
			if err != nil {
				t.Fatalf("GetRemainingRequests error: %v", err)
			}

			if remaining != tt.wantRemaining {
				t.Errorf("GetRemainingRequests() = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}
