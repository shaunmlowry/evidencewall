package config

import (
	"os"
)

// Config holds all configuration for the boards service
type Config struct {
	DatabaseURL    string
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	JWTSecret      string
	AuthServiceURL string
	CORSOrigins    string
	Environment    string
	LogLevel       string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://evidencewall:evidencewall_password@localhost:5432/evidencewall?sslmode=disable"),
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		CORSOrigins:    getEnv("CORS_ORIGINS", "http://localhost:3000"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "debug"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}


