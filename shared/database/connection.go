package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"evidence-wall/shared/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the PostgreSQL database
func Connect(databaseURL string) (*gorm.DB, error) {
	// Configure GORM logger
	logLevel := logger.Silent
	if os.Getenv("ENVIRONMENT") == "development" {
		logLevel = logger.Info
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Board{},
		&models.BoardUser{},
		&models.BoardItem{},
		&models.BoardConnection{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional database indexes for performance
func CreateIndexes(db *gorm.DB) error {
	log.Println("Creating database indexes...")

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_boards_owner_id ON boards(owner_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_boards_visibility ON boards(visibility)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_users_board_id ON board_users(board_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_users_user_id ON board_users(user_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_items_board_id ON board_items(board_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_connections_board_id ON board_connections(board_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_connections_from_item_id ON board_connections(from_item_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_connections_to_item_id ON board_connections(to_item_id)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// Continue with other indexes even if one fails
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}


