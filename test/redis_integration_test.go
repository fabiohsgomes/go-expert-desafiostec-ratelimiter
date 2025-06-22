// +build integration

package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/internal/middleware"
	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

// Redis integration tests - requires Redis to be running
// Run with: go test -tags=integration -v
func TestRedisIntegration(t *testing.T) {
	// Skip if Redis is not available
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Try to connect to Redis
	store, err := storage.NewRedisStorage(redisAddr, "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer store.Close()

	// Test Redis connectivity
	ctx := context.Background()
	_, err = store.GetRequestCount(ctx, "test-key")
	if err != nil {
		t.Skipf("Redis not responding: %v", err)
	}

	config := ratelimiter.NewConfig()
	config.MaxRequestsPerSecond = 2
	config.BlockDuration = time.Second * 2
	config.SetTokenLimit("redis-token", 5, time.Second*3)

	limiter := ratelimiter.New(store, config)
	rateLimiterMiddleware := middleware.New(limiter, config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := rateLimiterMiddleware.Handler(handler)

	t.Run("Redis-backed rate limiting works", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.100:12345"

		// First requests should pass
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
	})

	t.Run("Redis token-based limiting works", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.101:12345"
		req.Header.Set("API_KEY", "redis-token")

		// Should allow 5 requests
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Token request %d should pass, got status %d", i+1, w.Code)
			}
		}

		// 6th request should be blocked
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected token to be rate limited, got status %d", w.Code)
		}
	})

	t.Run("Redis persistence across middleware instances", func(t *testing.T) {
		// Create first middleware instance
		limiter1 := ratelimiter.New(store, config)
		middleware1 := middleware.New(limiter1, config)
		handler1 := middleware1.Handler(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.102:12345"

		// Use up the limit with first instance
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			handler1.ServeHTTP(w, req)
		}

		// Create second middleware instance (simulating app restart)
		limiter2 := ratelimiter.New(store, config)
		middleware2 := middleware.New(limiter2, config)
		handler2 := middleware2.Handler(handler)

		// Should still be blocked with second instance
		w := httptest.NewRecorder()
		handler2.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Error("Rate limit should persist across middleware instances")
		}
	})

	t.Run("Redis cleanup after block duration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.103:12345"

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

		// Wait for block duration
		time.Sleep(time.Second * 3)

		// Should work again
		w = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Error("Should work after block duration with Redis")
		}
	})
}

// Benchmark Redis performance
func BenchmarkRedisRateLimiter(b *testing.B) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	store, err := storage.NewRedisStorage(redisAddr, "", 0)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
	}
	defer store.Close()

	config := ratelimiter.NewConfig()
	config.MaxRequestsPerSecond = 10000 // High limit for benchmarking
	
	limiter := ratelimiter.New(store, config)
	rateLimiterMiddleware := middleware.New(limiter, config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rateLimiterMiddleware.Handler(handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := httptest.NewRequest("GET", "/bench", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(w, req)
			i++
		}
	})
}