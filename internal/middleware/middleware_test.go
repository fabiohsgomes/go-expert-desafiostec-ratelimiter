package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
)

// mockLimiter implements a simple mock of the rate limiter for testing
type mockLimiter struct {
	allowed bool
	err     error
}

func (m *mockLimiter) IsAllowed(ctx context.Context, key string, isToken bool) (bool, error) {
	return m.allowed, m.err
}

func (m *mockLimiter) GetRemainingRequests(ctx context.Context, key string, isToken bool) (int, error) {
	return 0, nil
}

func TestRateLimiterMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		limiterAllowed bool
		limiterErr     error
		token          string
		wantStatus     int
		wantError      string
	}{
		{
			name:           "Request allowed",
			limiterAllowed: true,
			wantStatus:     http.StatusOK,
		},
		{
			name:           "Request blocked",
			limiterAllowed: false,
			wantStatus:     http.StatusTooManyRequests,
			wantError:      "you have reached the maximum number of requests or actions allowed within a certain time frame",
		},
		{
			name:           "Request with token allowed",
			limiterAllowed: true,
			token:          "test-token",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "Request with token blocked",
			limiterAllowed: false,
			token:          "test-token",
			wantStatus:     http.StatusTooManyRequests,
			wantError:      "you have reached the maximum number of requests or actions allowed within a certain time frame",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock limiter
			mockLimiter := &mockLimiter{
				allowed: tt.limiterAllowed,
				err:     tt.limiterErr,
			}

			// Create config
			config := &ratelimiter.Config{
				MaxRequestsPerSecond: 10,
				BlockDuration:        time.Minute,
				TokenHeader:          "API_KEY",
			}

			// Create middleware
			middleware := &RateLimiterMiddleware{
				limiter: mockLimiter,
				config:  config,
			}

			// Create test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create test request
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			if tt.token != "" {
				req.Header.Set("API_KEY", tt.token)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute middleware
			middleware.Handler(nextHandler).ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					w.Code, tt.wantStatus)
			}

			// Check error message for blocked requests
			if tt.wantError != "" {
				var response ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response body: %v", err)
				}
				if response.Error != tt.wantError {
					t.Errorf("handler returned wrong error message: got %v want %v",
						response.Error, tt.wantError)
				}
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			remoteAddr: "10.0.0.1:1234",
			want:       "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			remoteAddr: "10.0.0.1:1234",
			want:       "192.168.1.2",
		},
		{
			name:       "RemoteAddr only",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.3:1234",
			want:       "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
