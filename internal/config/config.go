package config

import (
	"os"
	"strconv"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
	"github.com/joho/godotenv"
)

// LoadConfig loads configuration from environment variables or .env file
func LoadConfig() (*ratelimiter.Config, error) {
	// Try to load .env file if it exists
	godotenv.Load()

	config := ratelimiter.NewConfig()

	// Load general rate limit settings
	if maxReqs := os.Getenv("RATE_LIMIT_MAX_REQUESTS"); maxReqs != "" {
		if val, err := strconv.Atoi(maxReqs); err == nil {
			config.MaxRequestsPerSecond = val
		}
	}

	if blockDuration := os.Getenv("RATE_LIMIT_BLOCK_DURATION"); blockDuration != "" {
		if duration, err := time.ParseDuration(blockDuration); err == nil {
			config.BlockDuration = duration
		}
	}

	if tokenHeader := os.Getenv("RATE_LIMIT_TOKEN_HEADER"); tokenHeader != "" {
		config.TokenHeader = tokenHeader
	}

	return config, nil
}

// LoadRedisConfig loads Redis configuration from environment
func LoadRedisConfig() (addr, password string, db int) {
	addr = os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password = os.Getenv("REDIS_PASSWORD")

	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if val, err := strconv.Atoi(dbStr); err == nil {
			db = val
		}
	}

	return
}
