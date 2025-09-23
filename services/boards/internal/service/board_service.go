package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"

	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrBoardNotFound      = errors.New("board not found")
	ErrItemNotFound       = errors.New("item not found")
	ErrConnectionNotFound = errors.New("connection not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInputTooLong       = errors.New("input too long")
	ErrInvalidCharacters  = errors.New("invalid characters in input")
)

// Input validation constants
const (
	MaxTitleLength    = 200
	MaxContentLength  = 5000
	MaxDescriptionLength = 1000
	MaxNameLength     = 100
)

// HTML tag regex for sanitization
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// Input validation and sanitization functions
func validateAndSanitizeString(input string, maxLength int, fieldName string) (string, error) {
	if len(input) > maxLength {
		return "", fmt.Errorf("%s: %w (max %d characters)", fieldName, ErrInputTooLong, maxLength)
	}
	
	// Remove HTML tags and escape HTML entities
	sanitized := htmlTagRegex.ReplaceAllString(input, "")
	sanitized = html.EscapeString(sanitized)
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized, nil
}

func validateTitle(title string) (string, error) {
	return validateAndSanitizeString(title, MaxTitleLength, "title")
}

func validateContent(content string) (string, error) {
	return validateAndSanitizeString(content, MaxContentLength, "content")
}

func validateDescription(description string) (string, error) {
	return validateAndSanitizeString(description, MaxDescriptionLength, "description")
}

func validateName(name string) (string, error) {
	return validateAndSanitizeString(name, MaxNameLength, "name")
}

// BoardService handles board business logic
type BoardService struct {
	boardRepo      BoardRepositoryInterface
	boardUserRepo  BoardUserRepositoryInterface
	boardItemRepo  BoardItemRepositoryInterface
	connectionRepo BoardConnectionRepositoryInterface
	redis          *redis.Client
}

// NewBoardService creates a new board service
func NewBoardService(
	boardRepo BoardRepositoryInterface,
	boardUserRepo BoardUserRepositoryInterface,
	boardItemRepo BoardItemRepositoryInterface,
	connectionRepo BoardConnectionRepositoryInterface,
	redis *redis.Client,
) *BoardService {
	return &BoardService{
		boardRepo:      boardRepo,
		boardUserRepo:  boardUserRepo,
		boardItemRepo:  boardItemRepo,
		connectionRepo: connectionRepo,
		redis:          redis,
	}
}

// CreateBoardRequest represents a board creation request
type CreateBoardRequest struct {
	Title       string                 `json:"title" binding:"required,min=1,max=200"`
	Description string                 `json:"description" binding:"max=1000"`
	Visibility  models.BoardVisibility `json:"visibility" binding:"required,oneof=private shared public"`
}

// UpdateBoardRequest represents a board update request
type UpdateBoardRequest struct {
	Title       string                 `json:"title" binding:"omitempty,min=1,max=200"`
	Description string                 `json:"description" binding:"omitempty,max=1000"`
	Visibility  models.BoardVisibility `json:"visibility" binding:"omitempty,oneof=private shared public"`
}

// CreateBoard creates a new board
func (s *BoardService) CreateBoard(userID uuid.UUID, req CreateBoardRequest) (*models.Board, error) {
	// Validate and sanitize input
	title, err := validateTitle(req.Title)
	if err != nil {
		return nil, fmt.Errorf("title validation failed: %w", err)
	}
	
	description, err := validateDescription(req.Description)
	if err != nil {
		return nil, fmt.Errorf("description validation failed: %w", err)
	}

	board := &models.Board{
		Title:       title,
		Description: description,
		Visibility:  req.Visibility,
		OwnerID:     userID,
	}

	if err := s.boardRepo.Create(board); err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}

	// Ensure creator has admin permission explicitly in board_users for consistency
	if err := s.boardUserRepo.Create(&models.BoardUser{
		BoardID:    board.ID,
		UserID:     userID,
		Permission: models.PermissionAdmin,
	}); err != nil {
		// Non-fatal; log in real app. Continue returning created board.
	}

	return board, nil
}

// GetBoard retrieves a board by ID with permission check
func (s *BoardService) GetBoard(boardID, userID uuid.UUID) (*models.BoardResponse, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil || permission == "" {
		return nil, ErrBoardNotFound
	}

	response := board.ToResponse(permission)
	return &response, nil
}

// GetPublicBoard retrieves a public board (no auth required)
func (s *BoardService) GetPublicBoard(boardID uuid.UUID) (*models.Board, error) {
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if board.Visibility != models.VisibilityPublic {
		return nil, ErrUnauthorized
	}

	return board, nil
}

// ListBoards retrieves boards accessible by a user
func (s *BoardService) ListBoards(userID uuid.UUID, offset, limit int) ([]models.BoardResponse, int64, error) {
	boards, total, err := s.boardRepo.ListByUser(userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list boards: %w", err)
	}

	responses := make([]models.BoardResponse, 0, len(boards))
	for _, b := range boards {
		// Determine permission for current user
		perm := models.PermissionRead
		if b.OwnerID == userID {
			perm = models.PermissionAdmin
		} else {
			for _, bu := range b.Users {
				if bu.UserID == userID {
					perm = bu.Permission
					break
				}
			}
		}
		responses = append(responses, b.ToResponse(perm))
	}

	return responses, total, nil
}

// UpdateBoard updates a board
func (s *BoardService) UpdateBoard(boardID, userID uuid.UUID, req UpdateBoardRequest) (*models.Board, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission != models.PermissionAdmin {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Title != "" {
		board.Title = req.Title
	}
	if req.Description != "" {
		board.Description = req.Description
	}
	if req.Visibility != "" {
		board.Visibility = req.Visibility
	}

	if err := s.boardRepo.Update(board); err != nil {
		return nil, fmt.Errorf("failed to update board: %w", err)
	}

	return board, nil
}

// DeleteBoard deletes a board
func (s *BoardService) DeleteBoard(boardID, userID uuid.UUID) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission != models.PermissionAdmin {
		return ErrUnauthorized
	}

	// Delete all related data
	if err := s.connectionRepo.DeleteByBoard(boardID); err != nil {
		return fmt.Errorf("failed to delete board connections: %w", err)
	}
	if err := s.boardItemRepo.DeleteByBoard(boardID); err != nil {
		return fmt.Errorf("failed to delete board items: %w", err)
	}
	if err := s.boardRepo.Delete(boardID); err != nil {
		return fmt.Errorf("failed to delete board: %w", err)
	}

	return nil
}

// ShareBoardRequest represents a board sharing request
type ShareBoardRequest struct {
	UserID     uuid.UUID              `json:"user_id" binding:"required"`
	Permission models.PermissionLevel `json:"permission" binding:"required,oneof=read write admin"`
}

// ShareBoard shares a board with a user
func (s *BoardService) ShareBoard(boardID, ownerID uuid.UUID, req ShareBoardRequest) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission != models.PermissionAdmin {
		return ErrUnauthorized
	}

	// Check if user already has access
	existing, err := s.boardUserRepo.GetByBoardAndUser(boardID, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to check existing access: %w", err)
	}

	if existing != nil {
		// Update existing permission
		existing.Permission = req.Permission
		return s.boardUserRepo.Update(existing)
	}

	// Create new board user relationship
	boardUser := &models.BoardUser{
		BoardID:    boardID,
		UserID:     req.UserID,
		Permission: req.Permission,
	}

	return s.boardUserRepo.Create(boardUser)
}

// UnshareBoard removes a user's access to a board
func (s *BoardService) UnshareBoard(boardID, ownerID, targetUserID uuid.UUID) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission != models.PermissionAdmin {
		return ErrUnauthorized
	}

	return s.boardUserRepo.Delete(boardID, targetUserID)
}

// UpdateUserPermissionRequest represents a permission update request
type UpdateUserPermissionRequest struct {
	Permission models.PermissionLevel `json:"permission" binding:"required,oneof=read write admin"`
}

// UpdateUserPermission updates a user's permission for a board
func (s *BoardService) UpdateUserPermission(boardID, ownerID, targetUserID uuid.UUID, req UpdateUserPermissionRequest) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission != models.PermissionAdmin {
		return ErrUnauthorized
	}

	boardUser, err := s.boardUserRepo.GetByBoardAndUser(boardID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get board user: %w", err)
	}
	if boardUser == nil {
		return ErrUnauthorized
	}

	boardUser.Permission = req.Permission
	return s.boardUserRepo.Update(boardUser)
}

// CreateItemRequest represents a board item creation request
type CreateItemRequest struct {
	Type     models.ItemType        `json:"type" binding:"required,oneof=text image note link"`
	Content  string                 `json:"content" binding:"required"`
	X        float64                `json:"x" binding:"required"`
	Y        float64                `json:"y" binding:"required"`
	Width    float64                `json:"width" binding:"required,min=10"`
	Height   float64                `json:"height" binding:"required,min=10"`
	ZIndex   int                    `json:"z_index"`
	Color    string                 `json:"color"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateBoardItem creates a new board item
func (s *BoardService) CreateBoardItem(boardID, userID uuid.UUID, req CreateItemRequest) (*models.BoardItem, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return nil, ErrUnauthorized
	}

	// Validate and sanitize content
	content, err := validateContent(req.Content)
	if err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	// Determine persisted item type to satisfy DB constraints.
	// Frontend may send req.Type = "note" and specify UI variant in metadata.variant.
	// Map to DB types: 'post-it' | 'suspect-card'. Default to 'post-it'.
	persistedType := models.ItemType("post-it")
	if req.Metadata != nil {
		if v, ok := req.Metadata["variant"]; ok {
			if vs, ok := v.(string); ok {
				switch vs {
				case "suspect-card":
					persistedType = models.ItemType("suspect-card")
				case "post-it":
					persistedType = models.ItemType("post-it")
				}
			}
		}
	}

	item := &models.BoardItem{
		BoardID:   boardID,
		Type:      persistedType,
		Content:   content,
		X:         req.X,
		Y:         req.Y,
		Width:     req.Width,
		Height:    req.Height,
		ZIndex:    req.ZIndex,
		Color:     req.Color,
		Metadata:  metadataJSON,
		CreatedBy: userID,
	}

	if err := s.boardItemRepo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Publish real-time update
	s.publishBoardUpdate(boardID, "item_created", item)

	return item, nil
}

// UpdateItemRequest represents a board item update request
type UpdateItemRequest struct {
	Content  string                 `json:"content"`
	X        *float64               `json:"x"`
	Y        *float64               `json:"y"`
	Width    *float64               `json:"width"`
	Height   *float64               `json:"height"`
	ZIndex   *int                   `json:"z_index"`
	Color    string                 `json:"color"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UpdateBoardItem updates a board item
func (s *BoardService) UpdateBoardItem(boardID, itemID, userID uuid.UUID, req UpdateItemRequest) (*models.BoardItem, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return nil, ErrUnauthorized
	}

	item, err := s.boardItemRepo.GetByID(itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil || item.BoardID != boardID {
		return nil, ErrItemNotFound
	}

	// Update fields if provided
	if req.Content != "" {
		content, err := validateContent(req.Content)
		if err != nil {
			return nil, fmt.Errorf("content validation failed: %w", err)
		}
		item.Content = content
	}
	if req.X != nil {
		item.X = *req.X
	}
	if req.Y != nil {
		item.Y = *req.Y
	}
	if req.Width != nil {
		item.Width = *req.Width
	}
	if req.Height != nil {
		item.Height = *req.Height
	}
	if req.ZIndex != nil {
		item.ZIndex = *req.ZIndex
	}
	if req.Color != "" {
		item.Color = req.Color
	}
	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		item.Metadata = metadataJSON
	}

	if err := s.boardItemRepo.Update(item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	// Publish real-time update
	s.publishBoardUpdate(boardID, "item_updated", item)

	return item, nil
}

// DeleteBoardItem deletes a board item
func (s *BoardService) DeleteBoardItem(boardID, itemID, userID uuid.UUID) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return ErrUnauthorized
	}

	item, err := s.boardItemRepo.GetByID(itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil || item.BoardID != boardID {
		return ErrItemNotFound
	}

	// Delete related connections first
	if err := s.connectionRepo.DeleteByItem(itemID); err != nil {
		return fmt.Errorf("failed to delete item connections: %w", err)
	}

	if err := s.boardItemRepo.Delete(itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	// Publish real-time update
	s.publishBoardUpdate(boardID, "item_deleted", map[string]interface{}{"id": itemID})

	return nil
}

// ListBoardItems retrieves all items for a board
func (s *BoardService) ListBoardItems(boardID, userID uuid.UUID) ([]models.BoardItem, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" {
		return nil, ErrUnauthorized
	}

	items, err := s.boardItemRepo.ListByBoard(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	return items, nil
}

// Helper function to publish real-time updates
func (s *BoardService) publishBoardUpdate(boardID uuid.UUID, event string, data interface{}) {
	if s.redis == nil {
		return
	}

	update := map[string]interface{}{
		"board_id": boardID,
		"event":    event,
		"data":     data,
	}

	updateJSON, _ := json.Marshal(update)
	s.redis.Publish(context.Background(), fmt.Sprintf("board:%s", boardID), updateJSON)
}

// ----- Connections -----

// CreateConnectionRequest represents a request to create a connection
type CreateConnectionRequest struct {
	FromItemID uuid.UUID      `json:"from_item_id" binding:"required"`
	ToItemID   uuid.UUID      `json:"to_item_id" binding:"required"`
	Style      map[string]any `json:"style"`
}

// UpdateConnectionRequest represents a request to update a connection
type UpdateConnectionRequest struct {
	Style map[string]any `json:"style"`
}

// ListBoardConnections returns all connections for a board
func (s *BoardService) ListBoardConnections(boardID, userID uuid.UUID) ([]models.BoardConnection, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" {
		return nil, ErrUnauthorized
	}
	return s.connectionRepo.ListByBoard(boardID)
}

// CreateBoardConnection creates a new connection between two items
func (s *BoardService) CreateBoardConnection(boardID, userID uuid.UUID, req CreateConnectionRequest) (*models.BoardConnection, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return nil, ErrUnauthorized
	}

	// Validate items belong to the same board
	fromItem, err := s.boardItemRepo.GetByID(req.FromItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from item: %w", err)
	}
	toItem, err := s.boardItemRepo.GetByID(req.ToItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to item: %w", err)
	}
	if fromItem == nil || toItem == nil || fromItem.BoardID != boardID || toItem.BoardID != boardID {
		return nil, ErrInvalidInput
	}
	if fromItem.ID == toItem.ID {
		return nil, ErrInvalidInput
	}

	var styleJSON []byte
	if req.Style != nil {
		styleJSON, _ = json.Marshal(req.Style)
	}

	conn := &models.BoardConnection{
		BoardID:    boardID,
		FromItemID: req.FromItemID,
		ToItemID:   req.ToItemID,
		Style:      string(styleJSON),
		CreatedBy:  userID,
	}
	if err := s.connectionRepo.Create(conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	s.publishBoardUpdate(boardID, "connection_created", conn)
	return conn, nil
}

// UpdateBoardConnection updates connection style
func (s *BoardService) UpdateBoardConnection(boardID, connectionID, userID uuid.UUID, req UpdateConnectionRequest) (*models.BoardConnection, error) {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return nil, ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return nil, ErrUnauthorized
	}

	conn, err := s.connectionRepo.GetByID(connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil || conn.BoardID != boardID {
		return nil, ErrConnectionNotFound
	}

	if req.Style != nil {
		styleJSON, _ := json.Marshal(req.Style)
		conn.Style = string(styleJSON)
	}

	if err := s.connectionRepo.Update(conn); err != nil {
		return nil, fmt.Errorf("failed to update connection: %w", err)
	}

	s.publishBoardUpdate(boardID, "connection_updated", conn)
	return conn, nil
}

// DeleteBoardConnection deletes a connection
func (s *BoardService) DeleteBoardConnection(boardID, connectionID, userID uuid.UUID) error {
	board, permission, err := s.boardRepo.GetByIDWithPermission(boardID, userID)
	if err != nil {
		return fmt.Errorf("failed to get board: %w", err)
	}
	if board == nil {
		return ErrBoardNotFound
	}
	if permission == "" || permission == models.PermissionRead {
		return ErrUnauthorized
	}

	conn, err := s.connectionRepo.GetByID(connectionID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	if conn == nil || conn.BoardID != boardID {
		return ErrConnectionNotFound
	}

	if err := s.connectionRepo.Delete(connectionID); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	s.publishBoardUpdate(boardID, "connection_deleted", map[string]interface{}{"id": connectionID})
	return nil
}
