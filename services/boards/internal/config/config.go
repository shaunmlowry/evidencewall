package config

import (
	"fmt"
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
	TrustedProxies string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Provide devcontainer-friendly defaults when ENVIRONMENT=development
	environment := getEnv("ENVIRONMENT", "development")
	defaultDatabaseURL := "postgres://evidencewall:evidencewall_password@localhost:5432/evidencewall?sslmode=disable"
	defaultRedisHost := "localhost"
	if environment == "development" {
		gateway := getEnv("DEV_DOCKER_GATEWAY", "172.17.0.1")
		defaultDatabaseURL = fmt.Sprintf("postgres://evidencewall:evidencewall_password@%s:5432/evidencewall?sslmode=disable", gateway)
		defaultRedisHost = gateway
	}

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", defaultDatabaseURL),
		RedisHost:      getEnv("REDIS_HOST", defaultRedisHost),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		CORSOrigins:    getEnv("CORS_ORIGINS", "http://localhost:3000"),
		Environment:    environment,
		LogLevel:       getEnv("LOG_LEVEL", "debug"),
		TrustedProxies: getEnv("TRUSTED_PROXIES", ""),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
