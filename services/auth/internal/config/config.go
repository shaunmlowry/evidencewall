package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the auth service
type Config struct {
	DatabaseURL        string
	JWTSecret          string
	JWTExpiry          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	CORSOrigins        string
	Environment        string
	LogLevel           string
	TrustedProxies     string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Determine sensible development defaults that work inside the devcontainer.
	// When running services directly in the devcontainer, the database is exposed
	// via host.docker.internal. In production/docker-compose networks, these are
	// expected to be explicitly provided via environment variables.
	environment := getEnv("ENVIRONMENT", "development")
	defaultDatabaseURL := "postgres://evidencewall:evidencewall_password@localhost:5432/evidencewall?sslmode=disable"
	if environment == "development" {
		gateway := getEnv("DEV_DOCKER_GATEWAY", "172.17.0.1")
		defaultDatabaseURL = fmt.Sprintf("postgres://evidencewall:evidencewall_password@%s:5432/evidencewall?sslmode=disable", gateway)
	}

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", defaultDatabaseURL),
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8001/api/v1/auth/google/callback"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "http://localhost:3000"),
		Environment:        environment,
		LogLevel:           getEnv("LOG_LEVEL", "debug"),
		TrustedProxies:     getEnv("TRUSTED_PROXIES", ""),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
