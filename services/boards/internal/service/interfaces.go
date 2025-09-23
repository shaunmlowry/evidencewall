package service

import (
	"evidence-wall/shared/models"

	"github.com/google/uuid"
)

// BoardRepositoryInterface defines the interface for board repository operations
type BoardRepositoryInterface interface {
	Create(board *models.Board) error
	GetByID(id uuid.UUID) (*models.Board, error)
	GetByIDWithPermission(boardID, userID uuid.UUID) (*models.Board, models.PermissionLevel, error)
	ListByUser(userID uuid.UUID, offset, limit int) ([]models.Board, int64, error)
	ListPublic(offset, limit int) ([]models.Board, int64, error)
	Update(board *models.Board) error
	Delete(id uuid.UUID) error
}

// BoardUserRepositoryInterface defines the interface for board user repository operations
type BoardUserRepositoryInterface interface {
	Create(boardUser *models.BoardUser) error
	GetByBoardAndUser(boardID, userID uuid.UUID) (*models.BoardUser, error)
	Update(boardUser *models.BoardUser) error
	Delete(boardID, userID uuid.UUID) error
	ListByBoard(boardID uuid.UUID) ([]models.BoardUser, error)
}

// BoardItemRepositoryInterface defines the interface for board item repository operations
type BoardItemRepositoryInterface interface {
	Create(item *models.BoardItem) error
	GetByID(id uuid.UUID) (*models.BoardItem, error)
	ListByBoard(boardID uuid.UUID) ([]models.BoardItem, error)
	Update(item *models.BoardItem) error
	Delete(id uuid.UUID) error
	DeleteByBoard(boardID uuid.UUID) error
}

// BoardConnectionRepositoryInterface defines the interface for board connection repository operations
type BoardConnectionRepositoryInterface interface {
	Create(connection *models.BoardConnection) error
	GetByID(id uuid.UUID) (*models.BoardConnection, error)
	ListByBoard(boardID uuid.UUID) ([]models.BoardConnection, error)
	Update(connection *models.BoardConnection) error
	Delete(id uuid.UUID) error
	DeleteByBoard(boardID uuid.UUID) error
	DeleteByItem(itemID uuid.UUID) error
}
