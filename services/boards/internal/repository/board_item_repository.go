package repository

import (
	"errors"
	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BoardItemRepository handles board item data operations
type BoardItemRepository struct {
	db *gorm.DB
}

// NewBoardItemRepository creates a new board item repository
func NewBoardItemRepository(db *gorm.DB) *BoardItemRepository {
	return &BoardItemRepository{db: db}
}

// Create creates a new board item
func (r *BoardItemRepository) Create(item *models.BoardItem) error {
	return r.db.Create(item).Error
}

// GetByID retrieves a board item by ID
func (r *BoardItemRepository) GetByID(id uuid.UUID) (*models.BoardItem, error) {
	var item models.BoardItem
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// ListByBoard retrieves all items for a board
func (r *BoardItemRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardItem, error) {
	var items []models.BoardItem
	err := r.db.Where("board_id = ?", boardID).Order("z_index ASC, created_at ASC").Find(&items).Error
	return items, err
}

// Update updates a board item
func (r *BoardItemRepository) Update(item *models.BoardItem) error {
	return r.db.Save(item).Error
}

// Delete permanently deletes a board item
func (r *BoardItemRepository) Delete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.BoardItem{}).Error
}

// DeleteByBoard permanently deletes all items for a board
func (r *BoardItemRepository) DeleteByBoard(boardID uuid.UUID) error {
	return r.db.Unscoped().Where("board_id = ?", boardID).Delete(&models.BoardItem{}).Error
}

// BoardConnectionRepository handles board connection data operations
type BoardConnectionRepository struct {
	db *gorm.DB
}

// NewBoardConnectionRepository creates a new board connection repository
func NewBoardConnectionRepository(db *gorm.DB) *BoardConnectionRepository {
	return &BoardConnectionRepository{db: db}
}

// Create creates a new board connection
func (r *BoardConnectionRepository) Create(connection *models.BoardConnection) error {
	return r.db.Create(connection).Error
}

// GetByID retrieves a board connection by ID
func (r *BoardConnectionRepository) GetByID(id uuid.UUID) (*models.BoardConnection, error) {
	var connection models.BoardConnection
	err := r.db.Where("id = ?", id).First(&connection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &connection, nil
}

// ListByBoard retrieves all connections for a board
func (r *BoardConnectionRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardConnection, error) {
	var connections []models.BoardConnection
	err := r.db.Where("board_id = ?", boardID).Order("created_at ASC").Find(&connections).Error
	return connections, err
}

// Update updates a board connection
func (r *BoardConnectionRepository) Update(connection *models.BoardConnection) error {
	return r.db.Save(connection).Error
}

// Delete permanently deletes a board connection
func (r *BoardConnectionRepository) Delete(id uuid.UUID) error {
	return r.db.Unscoped().Where("id = ?", id).Delete(&models.BoardConnection{}).Error
}

// DeleteByBoard permanently deletes all connections for a board
func (r *BoardConnectionRepository) DeleteByBoard(boardID uuid.UUID) error {
	return r.db.Unscoped().Where("board_id = ?", boardID).Delete(&models.BoardConnection{}).Error
}

// DeleteByItem permanently deletes all connections for a specific item
func (r *BoardConnectionRepository) DeleteByItem(itemID uuid.UUID) error {
	return r.db.Unscoped().Where("from_item_id = ? OR to_item_id = ?", itemID, itemID).Delete(&models.BoardConnection{}).Error
}


