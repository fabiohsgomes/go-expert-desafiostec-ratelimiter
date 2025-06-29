package main

import (
        "log"
        "net/http"

        "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/middleware"
        "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
        "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

func main() {
        // Load configuration
        cfg, err := middleware.LoadConfig()
        if err != nil {
                log.Fatal(err)
        }

        // Load token limits from environment variables
        cfg.LoadTokenLimitsFromEnv()

        // Initialize Redis storage
        addr, password, db := middleware.LoadRedisConfig()
        store, err := storage.NewRedisStorage(addr, password, db)
        if err != nil {
                log.Fatal(err)
        }
        defer store.Close()

        // Create rate limiter
        limiter := ratelimiter.New(store, cfg)

        // Create middleware
        rateLimiterMiddleware := middleware.New(limiter, cfg)

        // Create a simple handler
        handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte("Hello, World!"))
        })

        // Wrap the handler with the rate limiter middleware
        http.Handle("/", rateLimiterMiddleware.Handler(handler))

        // Start the server
        log.Println("Server starting on :8080")
        if err := http.ListenAndServe(":8080", nil); err != nil {
                log.Fatal(err)
        }
}
