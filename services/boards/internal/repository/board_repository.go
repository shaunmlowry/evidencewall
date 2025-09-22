package repository

import (
	"errors"
	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BoardRepository handles board data operations
type BoardRepository struct {
	db *gorm.DB
}

// NewBoardRepository creates a new board repository
func NewBoardRepository(db *gorm.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

// Create creates a new board
func (r *BoardRepository) Create(board *models.Board) error {
	return r.db.Create(board).Error
}

// GetByID retrieves a board by ID with all related data
func (r *BoardRepository) GetByID(id uuid.UUID) (*models.Board, error) {
	var board models.Board
	err := r.db.Preload("Users.User").
		Preload("Items").
		Preload("Connections").
		Where("id = ?", id).
		First(&board).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &board, nil
}

// GetByIDWithPermission retrieves a board by ID and checks user permission
func (r *BoardRepository) GetByIDWithPermission(boardID, userID uuid.UUID) (*models.Board, models.PermissionLevel, error) {
	var board models.Board
	err := r.db.Preload("Users.User").
		Preload("Items").
		Preload("Connections").
		Where("id = ?", boardID).
		First(&board).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", nil
		}
		return nil, "", err
	}

	// Check if user is owner
	if board.OwnerID == userID {
		return &board, models.PermissionAdmin, nil
	}

	// Check if board is public
	if board.Visibility == models.VisibilityPublic {
		return &board, models.PermissionRead, nil
	}

	// Check user permissions
	var boardUser models.BoardUser
	err = r.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&boardUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", nil // No permission
		}
		return nil, "", err
	}

	return &board, boardUser.Permission, nil
}

// ListByUser retrieves boards accessible by a user
func (r *BoardRepository) ListByUser(userID uuid.UUID, offset, limit int) ([]models.Board, int64, error) {
	var boards []models.Board
	var total int64

	// Get boards where user is owner or has explicit permission
	query := r.db.Model(&models.Board{}).
		Joins("LEFT JOIN board_users ON boards.id = board_users.board_id").
		Where("boards.owner_id = ? OR board_users.user_id = ?", userID, userID).
		Group("boards.id")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get boards with pagination
	err := query.Preload("Users.User").
		Offset(offset).
		Limit(limit).
		Order("boards.updated_at DESC").
		Find(&boards).Error

	return boards, total, err
}

// ListPublic retrieves public boards
func (r *BoardRepository) ListPublic(offset, limit int) ([]models.Board, int64, error) {
	var boards []models.Board
	var total int64

	// Get total count
	if err := r.db.Model(&models.Board{}).Where("visibility = ?", models.VisibilityPublic).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get boards with pagination
	err := r.db.Where("visibility = ?", models.VisibilityPublic).
		Preload("Users.User").
		Offset(offset).
		Limit(limit).
		Order("updated_at DESC").
		Find(&boards).Error

	return boards, total, err
}

// Update updates a board
func (r *BoardRepository) Update(board *models.Board) error {
	return r.db.Save(board).Error
}

// Delete permanently deletes a board
func (r *BoardRepository) Delete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.Board{}).Error
}

// BoardUserRepository handles board user relationships
type BoardUserRepository struct {
	db *gorm.DB
}

// NewBoardUserRepository creates a new board user repository
func NewBoardUserRepository(db *gorm.DB) *BoardUserRepository {
	return &BoardUserRepository{db: db}
}

// Create creates a new board user relationship
func (r *BoardUserRepository) Create(boardUser *models.BoardUser) error {
	return r.db.Create(boardUser).Error
}

// GetByBoardAndUser retrieves a board user relationship
func (r *BoardUserRepository) GetByBoardAndUser(boardID, userID uuid.UUID) (*models.BoardUser, error) {
	var boardUser models.BoardUser
	err := r.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&boardUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &boardUser, nil
}

// Update updates a board user relationship
func (r *BoardUserRepository) Update(boardUser *models.BoardUser) error {
	return r.db.Save(boardUser).Error
}

// Delete deletes a board user relationship
func (r *BoardUserRepository) Delete(boardID, userID uuid.UUID) error {
	return r.db.Where("board_id = ? AND user_id = ?", boardID, userID).Delete(&models.BoardUser{}).Error
}

// ListByBoard retrieves all users for a board
func (r *BoardUserRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardUser, error) {
	var boardUsers []models.BoardUser
	err := r.db.Preload("User").Where("board_id = ?", boardID).Find(&boardUsers).Error
	return boardUsers, err
}


