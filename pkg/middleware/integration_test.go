package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/test"
)

// Integration tests for RateLimiterMiddleware
func TestRateLimiterMiddleware_Integration(t *testing.T) {
	// Setup in-memory storage for testing
	store := test.NewMemoryStorage()
	config := ratelimiter.NewConfig()
	config.MaxRequestsPerSecond = 2
	config.BlockDuration = time.Second * 2
	config.SetTokenLimit("test-token", 5, time.Second*3)

	limiter := ratelimiter.New(store, config)
	middleware := New(limiter, config)

	// Simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := middleware.Handler(handler)

	t.Run("IP-based rate limiting allows requests within limit", func(t *testing.T) {
		// First request should pass
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "success" {
			t.Errorf("Expected 'success', got %s", w.Body.String())
		}
	})

	t.Run("IP-based rate limiting blocks after exceeding limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"

		// Make requests up to the limit
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Request %d should pass, got status %d", i+1, w.Code)
			}
		}

		// Next request should be blocked
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", w.Code)
		}

		var errorResp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
			t.Errorf("Failed to parse error response: %v", err)
		}
		if errorResp.Error == "" {
			t.Error("Expected error message in response")
		}
	})

	t.Run("Token-based rate limiting takes precedence over IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"
		req.Header.Set("API_KEY", "test-token")

		// Make requests up to token limit (5)
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Token request %d should pass, got status %d", i+1, w.Code)
			}
		}

		// Next request should be blocked
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", w.Code)
		}
	})

	t.Run("Different IPs have separate rate limits", func(t *testing.T) {
		// IP 1 - exhaust limit
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.4:12345"

		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req1)
		}

		// IP 1 should be blocked
		w1 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w1, req1)
		if w1.Code != http.StatusTooManyRequests {
			t.Error("IP 1 should be blocked")
		}

		// IP 2 should still work
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.5:12345"
		w2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Error("IP 2 should not be blocked")
		}
	})

	t.Run("Rate limit resets after block duration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"

		// Exhaust limit
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
		}

		// Should be blocked
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Error("Should be blocked initially")
		}

		// Wait for block duration to pass
		time.Sleep(time.Second * 3)

		// Should work again
		w = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Error("Should work after block duration")
		}
	})

	t.Run("X-Forwarded-For header is respected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

		// Should use first IP from X-Forwarded-For
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Request %d should pass", i+1)
			}
		}

		// Should be blocked based on X-Forwarded-For IP
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Error("Should be blocked based on X-Forwarded-For")
		}
	})

	t.Run("X-Real-IP header is respected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		req.Header.Set("X-Real-IP", "203.0.113.2")

		// Should use X-Real-IP
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Request %d should pass", i+1)
			}
		}

		// Should be blocked based on X-Real-IP
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Error("Should be blocked based on X-Real-IP")
		}
	})
}

func TestRateLimiterMiddleware_ErrorHandling(t *testing.T) {
	// Create a mock limiter that returns errors
	mockLimiter := &mockErrorLimiter{}
	config := ratelimiter.NewConfig()
	middleware := New(mockLimiter, config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	wrappedHandler := middleware.Handler(handler)

	t.Run("Internal error returns 500", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

// Mock limiter that always returns errors
type mockErrorLimiter struct{}

func (m *mockErrorLimiter) IsAllowed(ctx context.Context, key string, isToken bool) (bool, error) {
	return false, &mockError{"test error"}
}

func (m *mockErrorLimiter) GetRemainingRequests(ctx context.Context, key string, isToken bool) (int, error) {
	return 0, &mockError{"test error"}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}