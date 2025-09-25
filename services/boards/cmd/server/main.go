package main

import (
	"context"
	"log"
	"os"

	"evidence-wall/boards-service/internal/config"
	"evidence-wall/boards-service/internal/handlers"
	"evidence-wall/boards-service/internal/repository"
	"evidence-wall/boards-service/internal/service"
	"evidence-wall/shared/auth"
	"evidence-wall/shared/database"
	"evidence-wall/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "evidence-wall/boards-service/docs" // Import generated docs
)

// @title Evidence Wall Boards Service API
// @version 1.0
// @description Boards management service for Evidence Wall application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8002
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

	log.Printf("boards:starting ENVIRONMENT=%s DB_HOST=%s REDIS=%s:%s", cfg.Environment, maskDatabaseHost(cfg.DatabaseURL), cfg.RedisHost, cfg.RedisPort)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("boards:db connect error: %v", err)
	}

	// Run migrations (temporarily disabled to debug)
	// if err := database.AutoMigrate(db); err != nil {
	//	log.Fatalf("Failed to run migrations: %v", err)
	// }

	// Create indexes (temporarily disabled to debug)
	// if err := database.CreateIndexes(db); err != nil {
	//	log.Printf("Warning: Failed to create indexes: %v", err)
	// }

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Simple connectivity check to Redis
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("boards:redis ping error: %v", err)
	} else {
		log.Printf("boards:redis connected addr=%s:%s", cfg.RedisHost, cfg.RedisPort)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 0) // Expiry not used for validation

	// Initialize repositories
	boardRepo := repository.NewBoardRepository(db)
	boardUserRepo := repository.NewBoardUserRepository(db)
	boardItemRepo := repository.NewBoardItemRepository(db)
	boardConnectionRepo := repository.NewBoardConnectionRepository(db)

	// Initialize services
	boardService := service.NewBoardService(boardRepo, boardUserRepo, boardItemRepo, boardConnectionRepo, rdb)

	// Initialize handlers
	boardHandler := handlers.NewBoardHandler(boardService)

	// Setup router
	router := gin.Default()

	// Configure trusted proxies (empty string means trust none)
	if cfg.TrustedProxies == "" {
		if err := router.SetTrustedProxies(nil); err != nil {
			log.Printf("Failed to set trusted proxies: %v", err)
		}
	} else {
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
	router.Use(middleware.RequestLogger("boards", cfg.Environment))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "boards"})
	})

	// (removed debug NoRoute handler)

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(jwtManager))
	{
		// (removed temporary test route)

		// Board routes
		boards := v1.Group("/boards")
		{
			boards.GET("", boardHandler.ListBoards)
			boards.POST("", boardHandler.CreateBoard)
			boards.GET("/:id", boardHandler.GetBoard)
			boards.PUT("/:id", boardHandler.UpdateBoard)
			boards.DELETE("/:id", boardHandler.DeleteBoard)

			// Board sharing routes
			boards.POST("/:id/share", boardHandler.ShareBoard)
			boards.DELETE("/:id/share/:userId", boardHandler.UnshareBoard)
			boards.PUT("/:id/users/:userId/permission", boardHandler.UpdateUserPermission)
		}

		// Board items routes (use consistent board :id and distinct item :itemId)
		items := boards.Group("/:id/items")
		{
			items.GET("", boardHandler.ListBoardItems)
			items.POST("", boardHandler.CreateBoardItem)
			items.PUT("/:itemId", boardHandler.UpdateBoardItem)
			items.DELETE("/:itemId", boardHandler.DeleteBoardItem)
		}

		// Board connections routes (use consistent board :id and distinct connection :connectionId)
		boards.GET("/:id/connections", boardHandler.ListBoardConnections)
		boards.POST("/:id/connections", boardHandler.CreateBoardConnection)
		boards.PUT("/:id/connections/:connectionId", boardHandler.UpdateBoardConnection)
		boards.DELETE("/:id/connections/:connectionId", boardHandler.DeleteBoardConnection)
	}

	// Public routes (for public boards)
	public := router.Group("/api/v1/public")
	public.Use(middleware.OptionalAuthMiddleware(jwtManager))
	{
		public.GET("/boards/:id", boardHandler.GetPublicBoard)
	}

	// Swagger documentation
	if cfg.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Start server
	port := os.Getenv("BOARDS_SERVICE_PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("boards:listening port=%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("boards:server start error: %v", err)
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
	at := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at == -1 {
		start := 0
		for i := 0; i < len(dsn); i++ {
			if dsn[i] == '/' {
				start = i + 2
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
