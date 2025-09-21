package main

import (
	"log"
	"os"
	"time"

	"evidence-wall/auth-service/internal/config"
	"evidence-wall/auth-service/internal/handlers"
	"evidence-wall/auth-service/internal/repository"
	"evidence-wall/auth-service/internal/service"
	"evidence-wall/shared/auth"
	"evidence-wall/shared/database"
	"evidence-wall/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "evidence-wall/auth-service/docs" // Import generated docs
)

// @title Evidence Wall Auth Service API
// @version 1.0
// @description Authentication service for Evidence Wall application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8001
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	log.Printf("auth:starting ENVIRONMENT=%s DB_HOST=%s", cfg.Environment, maskDatabaseHost(cfg.DatabaseURL))

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("auth:db connect error: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create indexes
	if err := database.CreateIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize JWT manager
	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		log.Fatalf("Invalid JWT expiry duration: %v", err)
	}
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, jwtExpiry)

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize service
	authService := service.NewAuthService(userRepo, jwtManager, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	router := gin.Default()

	// Configure trusted proxies (empty string means trust none)
	if cfg.TrustedProxies == "" {
		if err := router.SetTrustedProxies(nil); err != nil {
			log.Printf("Failed to set trusted proxies: %v", err)
		}
	} else {
		// Comma-separated list
		proxies := []string{}
		for _, p := range splitAndTrim(cfg.TrustedProxies) {
			proxies = append(proxies, p)
		}
		if err := router.SetTrustedProxies(proxies); err != nil {
			log.Printf("Failed to set trusted proxies: %v", err)
		}
	}

	// Add middleware
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigins))
	router.Use(middleware.RequestLogger("auth", cfg.Environment))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "auth"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/google", authHandler.GoogleLogin)
			auth.GET("/google/callback", authHandler.GoogleCallback)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/me", authHandler.GetProfile)
			protected.PUT("/me", authHandler.UpdateProfile)
			protected.POST("/logout", authHandler.Logout)
		}
	}

	// Swagger documentation
	if cfg.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Start server
	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("auth:listening port=%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("auth:server start error: %v", err)
	}
}

// splitAndTrim splits a comma-separated string and trims whitespace entries, skipping empties.
func splitAndTrim(s string) []string {
	res := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// trim spaces
			for len(part) > 0 && (part[0] == ' ' || part[0] == '\t') {
				part = part[1:]
			}
			for len(part) > 0 && (part[len(part)-1] == ' ' || part[len(part)-1] == '\t') {
				part = part[:len(part)-1]
			}
			if part != "" {
				res = append(res, part)
			}
			start = i + 1
		}
	}
	return res
}

// maskDatabaseHost extracts and prints only the database host part for logging.
// It avoids logging credentials while still being useful for debugging.
func maskDatabaseHost(dsn string) string {
	// Very small helper: find '@' then take host:port until next '/' or '?'
	at := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at == -1 {
		// no credentials case like postgres://host:5432/db
		start := 0
		for i := 0; i < len(dsn); i++ {
			if dsn[i] == '/' {
				start = i + 2 // skip '//'
				break
			}
		}
		host := ""
		for i := start; i < len(dsn); i++ {
			if dsn[i] == '/' || dsn[i] == '?' {
				break
			}
			host += string(dsn[i])
		}
		return host
	}
	host := ""
	for i := at + 1; i < len(dsn); i++ {
		if dsn[i] == '/' || dsn[i] == '?' {
			break
		}
		host += string(dsn[i])
	}
	return host
}
