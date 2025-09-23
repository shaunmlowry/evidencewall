package service

import (
	"errors"
	"testing"

	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBoardRepository is a mock implementation of BoardRepository
type MockBoardRepository struct {
	mock.Mock
}

func (m *MockBoardRepository) Create(board *models.Board) error {
	args := m.Called(board)
	return args.Error(0)
}

func (m *MockBoardRepository) GetByID(id uuid.UUID) (*models.Board, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Board), args.Error(1)
}

func (m *MockBoardRepository) GetByIDWithPermission(boardID, userID uuid.UUID) (*models.Board, models.PermissionLevel, error) {
	args := m.Called(boardID, userID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.Board), args.Get(1).(models.PermissionLevel), args.Error(2)
}

func (m *MockBoardRepository) ListByUser(userID uuid.UUID, offset, limit int) ([]models.Board, int64, error) {
	args := m.Called(userID, offset, limit)
	return args.Get(0).([]models.Board), args.Get(1).(int64), args.Error(2)
}

func (m *MockBoardRepository) ListPublic(offset, limit int) ([]models.Board, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]models.Board), args.Get(1).(int64), args.Error(2)
}

func (m *MockBoardRepository) Update(board *models.Board) error {
	args := m.Called(board)
	return args.Error(0)
}

func (m *MockBoardRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockBoardUserRepository is a mock implementation of BoardUserRepository
type MockBoardUserRepository struct {
	mock.Mock
}

func (m *MockBoardUserRepository) Create(boardUser *models.BoardUser) error {
	args := m.Called(boardUser)
	return args.Error(0)
}

func (m *MockBoardUserRepository) GetByBoardAndUser(boardID, userID uuid.UUID) (*models.BoardUser, error) {
	args := m.Called(boardID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardUser), args.Error(1)
}

func (m *MockBoardUserRepository) Update(boardUser *models.BoardUser) error {
	args := m.Called(boardUser)
	return args.Error(0)
}

func (m *MockBoardUserRepository) Delete(boardID, userID uuid.UUID) error {
	args := m.Called(boardID, userID)
	return args.Error(0)
}

func (m *MockBoardUserRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardUser, error) {
	args := m.Called(boardID)
	return args.Get(0).([]models.BoardUser), args.Error(1)
}

// MockBoardItemRepository is a mock implementation of BoardItemRepository
type MockBoardItemRepository struct {
	mock.Mock
}

func (m *MockBoardItemRepository) Create(item *models.BoardItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockBoardItemRepository) GetByID(id uuid.UUID) (*models.BoardItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardItem), args.Error(1)
}

func (m *MockBoardItemRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardItem, error) {
	args := m.Called(boardID)
	return args.Get(0).([]models.BoardItem), args.Error(1)
}

func (m *MockBoardItemRepository) Update(item *models.BoardItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockBoardItemRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBoardItemRepository) DeleteByBoard(boardID uuid.UUID) error {
	args := m.Called(boardID)
	return args.Error(0)
}

// MockBoardConnectionRepository is a mock implementation of BoardConnectionRepository
type MockBoardConnectionRepository struct {
	mock.Mock
}

func (m *MockBoardConnectionRepository) Create(connection *models.BoardConnection) error {
	args := m.Called(connection)
	return args.Error(0)
}

func (m *MockBoardConnectionRepository) GetByID(id uuid.UUID) (*models.BoardConnection, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardConnection), args.Error(1)
}

func (m *MockBoardConnectionRepository) ListByBoard(boardID uuid.UUID) ([]models.BoardConnection, error) {
	args := m.Called(boardID)
	return args.Get(0).([]models.BoardConnection), args.Error(1)
}

func (m *MockBoardConnectionRepository) Update(connection *models.BoardConnection) error {
	args := m.Called(connection)
	return args.Error(0)
}

func (m *MockBoardConnectionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBoardConnectionRepository) DeleteByBoard(boardID uuid.UUID) error {
	args := m.Called(boardID)
	return args.Error(0)
}

func (m *MockBoardConnectionRepository) DeleteByItem(itemID uuid.UUID) error {
	args := m.Called(itemID)
	return args.Error(0)
}

func TestBoardService_CreateBoard(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name         string
		request      CreateBoardRequest
		createErr    error
		boardUserErr error
		expectedErr  error
	}{
		{
			name: "successful board creation",
			request: CreateBoardRequest{
				Title:       "Test Board",
				Description: "Test Description",
				Visibility:  models.VisibilityPrivate,
			},
			createErr:    nil,
			boardUserErr: nil,
			expectedErr:  nil,
		},
		{
			name: "board creation error",
			request: CreateBoardRequest{
				Title:       "Test Board",
				Description: "Test Description",
				Visibility:  models.VisibilityPrivate,
			},
			createErr:    errors.New("database error"),
			boardUserErr: nil,
			expectedErr:  errors.New("failed to create board: database error"),
		},
		{
			name: "board user creation error (non-fatal)",
			request: CreateBoardRequest{
				Title:       "Test Board",
				Description: "Test Description",
				Visibility:  models.VisibilityPrivate,
			},
			createErr:    nil,
			boardUserErr: errors.New("board user error"),
			expectedErr:  nil, // Should still succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("Create", mock.AnythingOfType("*models.Board")).Return(tt.createErr)
			if tt.createErr == nil {
				mockBoardUserRepo.On("Create", mock.AnythingOfType("*models.BoardUser")).Return(tt.boardUserErr)
			}

			// Call method
			result, err := service.CreateBoard(userID, tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Title, result.Title)
				assert.Equal(t, tt.request.Description, result.Description)
				assert.Equal(t, tt.request.Visibility, result.Visibility)
				assert.Equal(t, userID, result.OwnerID)
			}

			mockBoardRepo.AssertExpectations(t)
			if tt.createErr == nil {
				mockBoardUserRepo.AssertExpectations(t)
			}
		})
	}
}

func TestBoardService_GetBoard(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		boardID     uuid.UUID
		userID      uuid.UUID
		board       *models.Board
		permission  models.PermissionLevel
		repoErr     error
		expectedErr error
	}{
		{
			name:        "successful board retrieval",
			boardID:     boardID,
			userID:      userID,
			board:       &models.Board{ID: boardID, Title: "Test Board", OwnerID: userID},
			permission:  models.PermissionAdmin,
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "board not found",
			boardID:     boardID,
			userID:      userID,
			board:       nil,
			permission:  "",
			repoErr:     nil,
			expectedErr: ErrBoardNotFound,
		},
		{
			name:        "repository error",
			boardID:     boardID,
			userID:      userID,
			board:       nil,
			permission:  "",
			repoErr:     errors.New("database error"),
			expectedErr: errors.New("failed to get board: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByIDWithPermission", tt.boardID, tt.userID).Return(tt.board, tt.permission, tt.repoErr)

			// Call method
			result, err := service.GetBoard(tt.boardID, tt.userID)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.board.ID, result.ID)
				assert.Equal(t, tt.board.Title, result.Title)
				assert.Equal(t, tt.permission, result.Permission)
			}

			mockBoardRepo.AssertExpectations(t)
		})
	}
}

func TestBoardService_GetPublicBoard(t *testing.T) {
	boardID := uuid.New()

	tests := []struct {
		name        string
		boardID     uuid.UUID
		board       *models.Board
		repoErr     error
		expectedErr error
	}{
		{
			name:        "successful public board retrieval",
			boardID:     boardID,
			board:       &models.Board{ID: boardID, Title: "Public Board", Visibility: models.VisibilityPublic},
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "board not found",
			boardID:     boardID,
			board:       nil,
			repoErr:     nil,
			expectedErr: ErrBoardNotFound,
		},
		{
			name:        "private board access denied",
			boardID:     boardID,
			board:       &models.Board{ID: boardID, Title: "Private Board", Visibility: models.VisibilityPrivate},
			repoErr:     nil,
			expectedErr: ErrUnauthorized,
		},
		{
			name:        "repository error",
			boardID:     boardID,
			board:       nil,
			repoErr:     errors.New("database error"),
			expectedErr: errors.New("failed to get board: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByID", tt.boardID).Return(tt.board, tt.repoErr)

			// Call method
			result, err := service.GetPublicBoard(tt.boardID)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.board.ID, result.ID)
				assert.Equal(t, tt.board.Title, result.Title)
			}

			mockBoardRepo.AssertExpectations(t)
		})
	}
}

func TestBoardService_UpdateBoard(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		boardID     uuid.UUID
		userID      uuid.UUID
		request     UpdateBoardRequest
		board       *models.Board
		permission  models.PermissionLevel
		repoErr     error
		updateErr   error
		expectedErr error
	}{
		{
			name:    "successful board update",
			boardID: boardID,
			userID:  userID,
			request: UpdateBoardRequest{
				Title:       "Updated Title",
				Description: "Updated Description",
				Visibility:  models.VisibilityShared,
			},
			board:       &models.Board{ID: boardID, Title: "Original Title", OwnerID: userID},
			permission:  models.PermissionAdmin,
			repoErr:     nil,
			updateErr:   nil,
			expectedErr: nil,
		},
		{
			name:        "board not found",
			boardID:     boardID,
			userID:      userID,
			request:     UpdateBoardRequest{Title: "Updated Title"},
			board:       nil,
			permission:  "",
			repoErr:     nil,
			updateErr:   nil,
			expectedErr: ErrBoardNotFound,
		},
		{
			name:        "unauthorized - read permission",
			boardID:     boardID,
			userID:      userID,
			request:     UpdateBoardRequest{Title: "Updated Title"},
			board:       &models.Board{ID: boardID, Title: "Original Title"},
			permission:  models.PermissionRead,
			repoErr:     nil,
			updateErr:   nil,
			expectedErr: ErrUnauthorized,
		},
		{
			name:        "update error",
			boardID:     boardID,
			userID:      userID,
			request:     UpdateBoardRequest{Title: "Updated Title"},
			board:       &models.Board{ID: boardID, Title: "Original Title", OwnerID: userID},
			permission:  models.PermissionAdmin,
			repoErr:     nil,
			updateErr:   errors.New("update error"),
			expectedErr: errors.New("failed to update board: update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByIDWithPermission", tt.boardID, tt.userID).Return(tt.board, tt.permission, tt.repoErr)
			if tt.board != nil && tt.permission == models.PermissionAdmin {
				mockBoardRepo.On("Update", mock.AnythingOfType("*models.Board")).Return(tt.updateErr)
			}

			// Call method
			result, err := service.UpdateBoard(tt.boardID, tt.userID, tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.request.Title != "" {
					assert.Equal(t, tt.request.Title, result.Title)
				}
				if tt.request.Description != "" {
					assert.Equal(t, tt.request.Description, result.Description)
				}
				if tt.request.Visibility != "" {
					assert.Equal(t, tt.request.Visibility, result.Visibility)
				}
			}

			mockBoardRepo.AssertExpectations(t)
		})
	}
}

func TestBoardService_DeleteBoard(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		boardID       uuid.UUID
		userID        uuid.UUID
		board         *models.Board
		permission    models.PermissionLevel
		repoErr       error
		connectionErr error
		itemErr       error
		deleteErr     error
		expectedErr   error
	}{
		{
			name:          "successful board deletion",
			boardID:       boardID,
			userID:        userID,
			board:         &models.Board{ID: boardID, OwnerID: userID},
			permission:    models.PermissionAdmin,
			repoErr:       nil,
			connectionErr: nil,
			itemErr:       nil,
			deleteErr:     nil,
			expectedErr:   nil,
		},
		{
			name:          "board not found",
			boardID:       boardID,
			userID:        userID,
			board:         nil,
			permission:    "",
			repoErr:       nil,
			connectionErr: nil,
			itemErr:       nil,
			deleteErr:     nil,
			expectedErr:   ErrBoardNotFound,
		},
		{
			name:          "unauthorized - read permission",
			boardID:       boardID,
			userID:        userID,
			board:         &models.Board{ID: boardID},
			permission:    models.PermissionRead,
			repoErr:       nil,
			connectionErr: nil,
			itemErr:       nil,
			deleteErr:     nil,
			expectedErr:   ErrUnauthorized,
		},
		{
			name:          "connection deletion error",
			boardID:       boardID,
			userID:        userID,
			board:         &models.Board{ID: boardID, OwnerID: userID},
			permission:    models.PermissionAdmin,
			repoErr:       nil,
			connectionErr: errors.New("connection error"),
			itemErr:       nil,
			deleteErr:     nil,
			expectedErr:   errors.New("failed to delete board connections: connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByIDWithPermission", tt.boardID, tt.userID).Return(tt.board, tt.permission, tt.repoErr)
			if tt.board != nil && tt.permission == models.PermissionAdmin {
				mockConnectionRepo.On("DeleteByBoard", tt.boardID).Return(tt.connectionErr)
				if tt.connectionErr == nil {
					mockBoardItemRepo.On("DeleteByBoard", tt.boardID).Return(tt.itemErr)
					if tt.itemErr == nil {
						mockBoardRepo.On("Delete", tt.boardID).Return(tt.deleteErr)
					}
				}
			}

			// Call method
			err := service.DeleteBoard(tt.boardID, tt.userID)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockBoardRepo.AssertExpectations(t)
			if tt.board != nil && tt.permission == models.PermissionAdmin {
				mockConnectionRepo.AssertExpectations(t)
				if tt.connectionErr == nil {
					mockBoardItemRepo.AssertExpectations(t)
					if tt.itemErr == nil {
						mockBoardRepo.AssertExpectations(t)
					}
				}
			}
		})
	}
}

func TestBoardService_ShareBoard(t *testing.T) {
	boardID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name         string
		boardID      uuid.UUID
		ownerID      uuid.UUID
		request      ShareBoardRequest
		board        *models.Board
		permission   models.PermissionLevel
		existingUser *models.BoardUser
		repoErr      error
		existingErr  error
		createErr    error
		updateErr    error
		expectedErr  error
	}{
		{
			name:    "successful board sharing - new user",
			boardID: boardID,
			ownerID: ownerID,
			request: ShareBoardRequest{
				UserID:     userID,
				Permission: models.PermissionWrite,
			},
			board:        &models.Board{ID: boardID, OwnerID: ownerID},
			permission:   models.PermissionAdmin,
			existingUser: nil,
			repoErr:      nil,
			existingErr:  nil,
			createErr:    nil,
			updateErr:    nil,
			expectedErr:  nil,
		},
		{
			name:    "successful board sharing - existing user update",
			boardID: boardID,
			ownerID: ownerID,
			request: ShareBoardRequest{
				UserID:     userID,
				Permission: models.PermissionWrite,
			},
			board:        &models.Board{ID: boardID, OwnerID: ownerID},
			permission:   models.PermissionAdmin,
			existingUser: &models.BoardUser{BoardID: boardID, UserID: userID, Permission: models.PermissionRead},
			repoErr:      nil,
			existingErr:  nil,
			createErr:    nil,
			updateErr:    nil,
			expectedErr:  nil,
		},
		{
			name:        "board not found",
			boardID:     boardID,
			ownerID:     ownerID,
			request:     ShareBoardRequest{UserID: userID, Permission: models.PermissionWrite},
			board:       nil,
			permission:  "",
			repoErr:     nil,
			existingErr: nil,
			createErr:   nil,
			updateErr:   nil,
			expectedErr: ErrBoardNotFound,
		},
		{
			name:        "unauthorized - not admin",
			boardID:     boardID,
			ownerID:     ownerID,
			request:     ShareBoardRequest{UserID: userID, Permission: models.PermissionWrite},
			board:       &models.Board{ID: boardID},
			permission:  models.PermissionWrite,
			repoErr:     nil,
			existingErr: nil,
			createErr:   nil,
			updateErr:   nil,
			expectedErr: ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByIDWithPermission", tt.boardID, tt.ownerID).Return(tt.board, tt.permission, tt.repoErr)
			if tt.board != nil && tt.permission == models.PermissionAdmin {
				mockBoardUserRepo.On("GetByBoardAndUser", tt.boardID, tt.request.UserID).Return(tt.existingUser, tt.existingErr)
				if tt.existingUser != nil {
					mockBoardUserRepo.On("Update", mock.AnythingOfType("*models.BoardUser")).Return(tt.updateErr)
				} else {
					mockBoardUserRepo.On("Create", mock.AnythingOfType("*models.BoardUser")).Return(tt.createErr)
				}
			}

			// Call method
			err := service.ShareBoard(tt.boardID, tt.ownerID, tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockBoardRepo.AssertExpectations(t)
			if tt.board != nil && tt.permission == models.PermissionAdmin {
				mockBoardUserRepo.AssertExpectations(t)
			}
		})
	}
}

func TestBoardService_CreateBoardItem(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		boardID     uuid.UUID
		userID      uuid.UUID
		request     CreateItemRequest
		board       *models.Board
		permission  models.PermissionLevel
		repoErr     error
		createErr   error
		expectedErr error
	}{
		{
			name:    "successful item creation",
			boardID: boardID,
			userID:  userID,
			request: CreateItemRequest{
				Type:     models.ItemTypeNote,
				Content:  "Test Note",
				X:        100,
				Y:        200,
				Width:    200,
				Height:   150,
				ZIndex:   1,
				Color:    "#ffff00",
				Metadata: map[string]interface{}{"variant": "post-it"},
			},
			board:       &models.Board{ID: boardID, OwnerID: userID},
			permission:  models.PermissionWrite,
			repoErr:     nil,
			createErr:   nil,
			expectedErr: nil,
		},
		{
			name:        "board not found",
			boardID:     boardID,
			userID:      userID,
			request:     CreateItemRequest{Type: models.ItemTypeNote, Content: "Test", X: 100, Y: 200, Width: 200, Height: 150},
			board:       nil,
			permission:  "",
			repoErr:     nil,
			createErr:   nil,
			expectedErr: ErrBoardNotFound,
		},
		{
			name:        "unauthorized - read permission",
			boardID:     boardID,
			userID:      userID,
			request:     CreateItemRequest{Type: models.ItemTypeNote, Content: "Test", X: 100, Y: 200, Width: 200, Height: 150},
			board:       &models.Board{ID: boardID},
			permission:  models.PermissionRead,
			repoErr:     nil,
			createErr:   nil,
			expectedErr: ErrUnauthorized,
		},
		{
			name:        "item creation error",
			boardID:     boardID,
			userID:      userID,
			request:     CreateItemRequest{Type: models.ItemTypeNote, Content: "Test", X: 100, Y: 200, Width: 200, Height: 150},
			board:       &models.Board{ID: boardID, OwnerID: userID},
			permission:  models.PermissionWrite,
			repoErr:     nil,
			createErr:   errors.New("create error"),
			expectedErr: errors.New("failed to create item: create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBoardRepo := new(MockBoardRepository)
			mockBoardUserRepo := new(MockBoardUserRepository)
			mockBoardItemRepo := new(MockBoardItemRepository)
			mockConnectionRepo := new(MockBoardConnectionRepository)

			service := NewBoardService(mockBoardRepo, mockBoardUserRepo, mockBoardItemRepo, mockConnectionRepo, nil)

			// Setup mocks
			mockBoardRepo.On("GetByIDWithPermission", tt.boardID, tt.userID).Return(tt.board, tt.permission, tt.repoErr)
			if tt.board != nil && tt.permission != "" && tt.permission != models.PermissionRead {
				mockBoardItemRepo.On("Create", mock.AnythingOfType("*models.BoardItem")).Return(tt.createErr)
			}

			// Call method
			result, err := service.CreateBoardItem(tt.boardID, tt.userID, tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Content, result.Content)
				assert.Equal(t, tt.request.X, result.X)
				assert.Equal(t, tt.request.Y, result.Y)
				assert.Equal(t, tt.request.Width, result.Width)
				assert.Equal(t, tt.request.Height, result.Height)
			}

			mockBoardRepo.AssertExpectations(t)
			if tt.board != nil && tt.permission != "" && tt.permission != models.PermissionRead {
				mockBoardItemRepo.AssertExpectations(t)
			}
		})
	}
}
