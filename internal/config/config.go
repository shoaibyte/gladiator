package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	Port         int
	Environment  string
	AppURL       string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	JWTAccessTTL int // minutes
	JWTRefreshTTL int // minutes (days * 24 * 60)
}

// Load reads configuration from environment and validates it.
func Load() (*Config, error) {
	if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "" {
		_ = godotenv.Load()
	}

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(jwtSecret) < 32 && env == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}
	if env == "production" && (jwtSecret == "dev-secret-key-change-in-production" || len(jwtSecret) < 32) {
		return nil, fmt.Errorf("JWT_SECRET: reject dev secret in production")
	}

	accessTTL := 15   // minutes
	refreshTTL := 7 * 24 * 60 // 7 days in minutes
	if t := os.Getenv("JWT_ACCESS_TTL_MIN"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 {
			accessTTL = v
		}
	}
	if t := os.Getenv("JWT_REFRESH_TTL_DAYS"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 {
			refreshTTL = v * 24 * 60
		}
	}

	return &Config{
		Port:          port,
		Environment:   env,
		AppURL:        appURL,
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		JWTSecret:     jwtSecret,
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
	}, nil
}
