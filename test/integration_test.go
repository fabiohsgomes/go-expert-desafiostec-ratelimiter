package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/internal/middleware"
	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
)

// Full integration test simulating a real HTTP server scenario
func TestFullIntegration(t *testing.T) {
	// Setup
	store := NewMemoryStorage()
	config := ratelimiter.NewConfig()
	config.MaxRequestsPerSecond = 3
	config.BlockDuration = time.Second * 2
	config.SetTokenLimit("premium-token", 10, time.Second*1)
	config.SetTokenLimit("basic-token", 5, time.Second*3)

	limiter := ratelimiter.New(store, config)
	rateLimiterMiddleware := middleware.New(limiter, config)

	// Create test server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"message": "Hello, World!",
			"path":    r.URL.Path,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := httptest.NewServer(rateLimiterMiddleware.Handler(handler))
	defer server.Close()

	client := &http.Client{Timeout: time.Second * 5}

	t.Run("Normal requests work within limits", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			resp, err := client.Get(server.URL + "/test")
			if err != nil {
				t.Fatalf("Request %d failed: %v", i+1, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
			}
		}
	})

	t.Run("Rate limiting kicks in after exceeding limit", func(t *testing.T) {
		// Make one more request to exceed the limit
		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", resp.StatusCode)
		}

		// Verify error response format
		var errorResp middleware.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			t.Errorf("Failed to decode error response: %v", err)
		}
		if errorResp.Error == "" {
			t.Error("Expected error message in response")
		}
	})

	t.Run("Premium token has higher limits", func(t *testing.T) {
		// Create request with premium token
		req, _ := http.NewRequest("GET", server.URL+"/premium", nil)
		req.Header.Set("API_KEY", "premium-token")

		// Should allow 10 requests
		for i := 0; i < 10; i++ {
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Premium request %d failed: %v", i+1, err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Premium request %d: expected status 200, got %d", i+1, resp.StatusCode)
			}
		}

		// 11th request should be blocked
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Premium limit test failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected premium token to be rate limited, got status %d", resp.StatusCode)
		}
	})

	t.Run("Basic token has lower limits", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/basic", nil)
		req.Header.Set("API_KEY", "basic-token")

		// Should allow 5 requests
		for i := 0; i < 5; i++ {
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Basic request %d failed: %v", i+1, err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Basic request %d: expected status 200, got %d", i+1, resp.StatusCode)
			}
		}

		// 6th request should be blocked
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Basic limit test failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected basic token to be rate limited, got status %d", resp.StatusCode)
		}
	})

	t.Run("Rate limits reset after block duration", func(t *testing.T) {
		// Wait for block duration to pass
		time.Sleep(time.Second * 3)

		// Should work again
		resp, err := client.Get(server.URL + "/reset-test")
		if err != nil {
			t.Fatalf("Reset test failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected request to work after reset, got status %d", resp.StatusCode)
		}
	})
}

// Benchmark test to measure performance under load
func BenchmarkRateLimiterMiddleware(b *testing.B) {
	store := NewMemoryStorage()
	config := ratelimiter.NewConfig()
	config.MaxRequestsPerSecond = 1000 // High limit for benchmarking
	
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
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i%255+1)
			w := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(w, req)
			i++
		}
	})
}