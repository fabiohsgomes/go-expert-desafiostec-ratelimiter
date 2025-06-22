package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
	"github.com/stretchr/testify/suite"
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

type MiddlewareTestSuite struct {
	suite.Suite
	config      *ratelimiter.Config
	nextHandler http.Handler
}

func (s *MiddlewareTestSuite) SetupTest() {
	s.config = &ratelimiter.Config{
		MaxRequestsPerSecond: 10,
		BlockDuration:        time.Minute,
		TokenHeader:          "API_KEY",
	}
	s.nextHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (s *MiddlewareTestSuite) createMiddleware(allowed bool, err error) *RateLimiterMiddleware {
	mockLimiter := &mockLimiter{
		allowed: allowed,
		err:     err,
	}
	return &RateLimiterMiddleware{
		limiter: mockLimiter,
		config:  s.config,
	}
}

func (s *MiddlewareTestSuite) TestRateLimiterMiddleware() {
	tests := []struct {
		name           string
		limiterAllowed bool
		limiterErr     error
		token          string
		headers        map[string]string
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
		{
			name:       "Limiter error with token",
			limiterErr: errors.New("storage error"),
			token:      "test-token",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Limiter error with IP",
			limiterErr: errors.New("storage error"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:           "Multiple IPs in X-Forwarded-For",
			limiterAllowed: true,
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			middleware := s.createMiddleware(tt.limiterAllowed, tt.limiterErr)

			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			if tt.token != "" {
				req.Header.Set("API_KEY", tt.token)
			}
			if tt.headers != nil {
				for k, v := range tt.headers {
					req.Header.Set(k, v)
				}
			}

			w := httptest.NewRecorder()
			middleware.Handler(s.nextHandler).ServeHTTP(w, req)

			s.Equal(tt.wantStatus, w.Code, "handler returned wrong status code")

			if tt.wantError != "" {
				var response ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				s.NoError(err, "Failed to decode response body")
				s.Equal(tt.wantError, response.Error, "handler returned wrong error message")
			}
		})
	}
}

type GetClientIPTestSuite struct {
	suite.Suite
}

func (s *GetClientIPTestSuite) TestGetClientIP() {
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
		{
			name: "Multiple IPs in X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteAddr: "10.0.0.1:1234",
			want:       "192.168.1.1",
		},
		{
			name: "Empty X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "",
			},
			remoteAddr: "192.168.1.3:1234",
			want:       "192.168.1.3",
		},
		{
			name:       "Invalid RemoteAddr format",
			headers:    map[string]string{},
			remoteAddr: "invalid",
			want:       "invalid",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getClientIP(req)
			s.Equal(tt.want, got, "getClientIP() returned unexpected value")
		})
	}
}

func TestMiddleware(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
	suite.Run(t, new(GetClientIPTestSuite))
}
