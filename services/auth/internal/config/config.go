package config

import (
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
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://evidencewall:evidencewall_password@localhost:5432/evidencewall?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8001/api/v1/auth/google/callback"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "http://localhost:3000"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		LogLevel:           getEnv("LOG_LEVEL", "debug"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}


