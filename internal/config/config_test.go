package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantReqs int
		wantDur  time.Duration
		wantHdr  string
	}{
		{
			name: "Default values",
			wantReqs: 10, // Default value from ratelimiter.NewConfig()
			wantDur:  time.Minute * 5,
			wantHdr:  "API_KEY",
		},
		{
			name: "Custom values",
			envVars: map[string]string{
				"RATE_LIMIT_MAX_REQUESTS":  "20",
				"RATE_LIMIT_BLOCK_DURATION": "10m",
				"RATE_LIMIT_TOKEN_HEADER":   "X-API-KEY",
			},
			wantReqs: 20,
			wantDur:  time.Minute * 10,
			wantHdr:  "X-API-KEY",
		},
		{
			name: "Invalid values",
			envVars: map[string]string{
				"RATE_LIMIT_MAX_REQUESTS":  "invalid",
				"RATE_LIMIT_BLOCK_DURATION": "invalid",
			},
			wantReqs: 10, // Should keep default values
			wantDur:  time.Minute * 5,
			wantHdr:  "API_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			// Check values
			if cfg.MaxRequestsPerSecond != tt.wantReqs {
				t.Errorf("MaxRequestsPerSecond = %v, want %v", cfg.MaxRequestsPerSecond, tt.wantReqs)
			}
			if cfg.BlockDuration != tt.wantDur {
				t.Errorf("BlockDuration = %v, want %v", cfg.BlockDuration, tt.wantDur)
			}
			if cfg.TokenHeader != tt.wantHdr {
				t.Errorf("TokenHeader = %v, want %v", cfg.TokenHeader, tt.wantHdr)
			}
		})
	}
}

func TestLoadRedisConfig(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		wantAddr     string
		wantPassword string
		wantDB       int
	}{
		{
			name:         "Default values",
			wantAddr:     "localhost:6379",
			wantPassword: "",
			wantDB:       0,
		},
		{
			name: "Custom values",
			envVars: map[string]string{
				"REDIS_ADDR":     "redis:6379",
				"REDIS_PASSWORD": "secret",
				"REDIS_DB":       "1",
			},
			wantAddr:     "redis:6379",
			wantPassword: "secret",
			wantDB:       1,
		},
		{
			name: "Invalid DB value",
			envVars: map[string]string{
				"REDIS_DB": "invalid",
			},
			wantAddr:     "localhost:6379",
			wantPassword: "",
			wantDB:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load Redis config
			addr, password, db := LoadRedisConfig()

			// Check values
			if addr != tt.wantAddr {
				t.Errorf("addr = %v, want %v", addr, tt.wantAddr)
			}
			if password != tt.wantPassword {
				t.Errorf("password = %v, want %v", password, tt.wantPassword)
			}
			if db != tt.wantDB {
				t.Errorf("db = %v, want %v", db, tt.wantDB)
			}
		})
	}
}