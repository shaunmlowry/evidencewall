package repository

import (
	"testing"

	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create tables manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			avatar TEXT,
			password TEXT NOT NULL,
			google_id TEXT UNIQUE,
			verified INTEGER DEFAULT 0,
			active INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE boards (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			visibility TEXT DEFAULT 'private',
			owner_id TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE board_users (
			id TEXT PRIMARY KEY,
			board_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			permission TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE board_items (
			id TEXT PRIMARY KEY,
			board_id TEXT NOT NULL,
			type TEXT NOT NULL,
			x REAL NOT NULL,
			y REAL NOT NULL,
			width REAL DEFAULT 200,
			height REAL DEFAULT 200,
			rotation REAL DEFAULT 0,
			z_index INTEGER DEFAULT 1,
			content TEXT,
			color TEXT,
			metadata TEXT,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE board_connections (
			id TEXT PRIMARY KEY,
			board_id TEXT NOT NULL,
			from_item_id TEXT NOT NULL,
			to_item_id TEXT NOT NULL,
			style TEXT,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestBoardRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	board := &models.Board{
		ID:          uuid.New(),
		Title:       "Test Board",
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     uuid.New(),
	}

	err := repo.Create(board)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, board.ID)
	assert.NotZero(t, board.CreatedAt)
	assert.NotZero(t, board.UpdatedAt)

	// Verify board was created
	var foundBoard models.Board
	err = db.First(&foundBoard, board.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, board.Title, foundBoard.Title)
	assert.Equal(t, board.Description, foundBoard.Description)
	assert.Equal(t, board.Visibility, foundBoard.Visibility)
}

func TestBoardRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	// Create a test board
	board := &models.Board{
		ID:          uuid.New(),
		Title:       "Test Board",
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     uuid.New(),
	}
	err := db.Create(board).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		boardID  uuid.UUID
		expected *models.Board
		hasError bool
	}{
		{
			name:     "existing board",
			boardID:  board.ID,
			expected: board,
			hasError: false,
		},
		{
			name:     "non-existing board",
			boardID:  uuid.New(),
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.boardID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.Title, result.Title)
				}
			}
		})
	}
}

func TestBoardRepository_GetByIDWithPermission(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	ownerID := uuid.New()
	userID := uuid.New()
	boardID := uuid.New()

	// Create a test board
	board := &models.Board{
		ID:          boardID,
		Title:       "Test Board",
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     ownerID,
	}
	err := db.Create(board).Error
	assert.NoError(t, err)

	// Create a board user relationship
	boardUser := &models.BoardUser{
		ID:         uuid.New(),
		BoardID:    boardID,
		UserID:     userID,
		Permission: models.PermissionWrite,
	}
	err = db.Create(boardUser).Error
	assert.NoError(t, err)

	tests := []struct {
		name       string
		boardID    uuid.UUID
		userID     uuid.UUID
		expected   *models.Board
		permission models.PermissionLevel
		hasError   bool
	}{
		{
			name:       "owner access",
			boardID:    boardID,
			userID:     ownerID,
			expected:   board,
			permission: models.PermissionAdmin,
			hasError:   false,
		},
		{
			name:       "user with permission",
			boardID:    boardID,
			userID:     userID,
			expected:   board,
			permission: models.PermissionWrite,
			hasError:   false,
		},
		{
			name:       "user without permission",
			boardID:    boardID,
			userID:     uuid.New(),
			expected:   nil,
			permission: "",
			hasError:   false,
		},
		{
			name:       "non-existing board",
			boardID:    uuid.New(),
			userID:     userID,
			expected:   nil,
			permission: "",
			hasError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, permission, err := repo.GetByIDWithPermission(tt.boardID, tt.userID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
					assert.Equal(t, models.PermissionLevel(""), permission)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.permission, permission)
				}
			}
		})
	}
}

func TestBoardRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	ownerID := uuid.New()
	userID := uuid.New()

	// Create test boards
	board1 := &models.Board{
		ID:          uuid.New(),
		Title:       "Board 1",
		Description: "Description 1",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     ownerID,
	}
	err := db.Create(board1).Error
	assert.NoError(t, err)

	board2 := &models.Board{
		ID:          uuid.New(),
		Title:       "Board 2",
		Description: "Description 2",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     userID,
	}
	err = db.Create(board2).Error
	assert.NoError(t, err)

	// Create board user relationship
	boardUser := &models.BoardUser{
		ID:         uuid.New(),
		BoardID:    board1.ID,
		UserID:     userID,
		Permission: models.PermissionRead,
	}
	err = db.Create(boardUser).Error
	assert.NoError(t, err)

	// Test listing boards for user
	boards, total, err := repo.ListByUser(userID, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, boards, 2)

	// Verify boards are ordered by updated_at DESC
	assert.Equal(t, board2.ID, boards[0].ID) // board2 should be first (more recent)
	assert.Equal(t, board1.ID, boards[1].ID) // board1 should be second
}

func TestBoardRepository_ListPublic(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	// Create test boards
	board1 := &models.Board{
		ID:          uuid.New(),
		Title:       "Public Board 1",
		Description: "Description 1",
		Visibility:  models.VisibilityPublic,
		OwnerID:     uuid.New(),
	}
	err := db.Create(board1).Error
	assert.NoError(t, err)

	board2 := &models.Board{
		ID:          uuid.New(),
		Title:       "Private Board",
		Description: "Description 2",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     uuid.New(),
	}
	err = db.Create(board2).Error
	assert.NoError(t, err)

	board3 := &models.Board{
		ID:          uuid.New(),
		Title:       "Public Board 2",
		Description: "Description 3",
		Visibility:  models.VisibilityPublic,
		OwnerID:     uuid.New(),
	}
	err = db.Create(board3).Error
	assert.NoError(t, err)

	// Test listing public boards
	boards, total, err := repo.ListPublic(0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, boards, 2)

	// Verify only public boards are returned
	for _, board := range boards {
		assert.Equal(t, models.VisibilityPublic, board.Visibility)
	}
}

func TestBoardRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	// Create a test board
	board := &models.Board{
		ID:          uuid.New(),
		Title:       "Original Title",
		Description: "Original Description",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     uuid.New(),
	}
	err := db.Create(board).Error
	assert.NoError(t, err)

	// Update the board
	board.Title = "Updated Title"
	board.Description = "Updated Description"
	board.Visibility = models.VisibilityShared

	err = repo.Update(board)
	assert.NoError(t, err)

	// Verify board was updated
	var foundBoard models.Board
	err = db.First(&foundBoard, board.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", foundBoard.Title)
	assert.Equal(t, "Updated Description", foundBoard.Description)
	assert.Equal(t, models.VisibilityShared, foundBoard.Visibility)
}

func TestBoardRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBoardRepository(db)

	// Create a test board
	board := &models.Board{
		ID:          uuid.New(),
		Title:       "Test Board",
		Description: "Test Description",
		Visibility:  models.VisibilityPrivate,
		OwnerID:     uuid.New(),
	}
	err := db.Create(board).Error
	assert.NoError(t, err)

	// Delete the board
	err = repo.Delete(board.ID)
	assert.NoError(t, err)

	// Verify board was deleted (hard delete)
	var foundBoard models.Board
	err = db.Unscoped().First(&foundBoard, board.ID).Error
	if err == nil {
		assert.NotNil(t, foundBoard.DeletedAt)
	}

	// Verify board is not found in normal query
	err = db.First(&foundBoard, board.ID).Error
	assert.Error(t, err)
}
