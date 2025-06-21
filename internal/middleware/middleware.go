package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
)

type RateLimiterMiddleware struct {
	limiter ratelimiter.RateLimiterInterface
	config  *ratelimiter.Config
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func New(limiter ratelimiter.RateLimiterInterface, config *ratelimiter.Config) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
		config:  config,
	}
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First check for token-based rate limiting
		token := r.Header.Get(m.config.TokenHeader)
		var allowed bool
		var err error

		if token != "" {
			// Token-based rate limiting takes precedence
			allowed, err = m.limiter.IsAllowed(r.Context(), token, true)
		} else {
			// Fall back to IP-based rate limiting
			ip := getClientIP(r)
			allowed, err = m.limiter.IsAllowed(r.Context(), ip, false)
		}

		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP if multiple are present
		return strings.Split(forwardedFor, ",")[0]
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
