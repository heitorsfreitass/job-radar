// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	RedisAddr   string
	RedisDB     int

	APIPort string

	// FrontendOrigin is the origin allowed to call the API cross-origin
	// (the React dev server by default; set to the deployed frontend's
	// origin in production).
	FrontendOrigin string

	// RemotiveMaxRequestsPerDay caps how many calls the worker is allowed
	// to make to the Remotive API in a 24h window. Remotive's terms of
	// use ask consumers to stay well below their published limits; this
	// value is enforced by internal/scheduler, not left implicit.
	RemotiveMaxRequestsPerDay int
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:               getEnv("DATABASE_URL", "postgres://job_radar:job_radar@localhost:5432/job_radar?sslmode=disable"),
		RedisAddr:                 getEnv("REDIS_ADDR", "localhost:6379"),
		APIPort:                   getEnv("API_PORT", "8080"),
		FrontendOrigin:            getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		RemotiveMaxRequestsPerDay: 4,
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	cfg.RedisDB = redisDB

	if v := os.Getenv("REMOTIVE_MAX_REQUESTS_PER_DAY"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REMOTIVE_MAX_REQUESTS_PER_DAY: %w", err)
		}
		cfg.RemotiveMaxRequestsPerDay = n
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
