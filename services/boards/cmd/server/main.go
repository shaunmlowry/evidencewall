package main

import (
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

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create indexes
	if err := database.CreateIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
	})

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

	// Add middleware
	router.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "boards"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(jwtManager))
	{
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

		// Board items routes
		items := v1.Group("/boards/:boardId/items")
		{
			items.GET("", boardHandler.ListBoardItems)
			items.POST("", boardHandler.CreateBoardItem)
			items.PUT("/:id", boardHandler.UpdateBoardItem)
			items.DELETE("/:id", boardHandler.DeleteBoardItem)
		}

		// Board connections routes
		connections := v1.Group("/boards/:boardId/connections")
		{
			connections.GET("", boardHandler.ListBoardConnections)
			connections.POST("", boardHandler.CreateBoardConnection)
			connections.PUT("/:id", boardHandler.UpdateBoardConnection)
			connections.DELETE("/:id", boardHandler.DeleteBoardConnection)
		}
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

	log.Printf("Boards service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


