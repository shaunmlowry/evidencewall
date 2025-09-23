package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"evidence-wall/boards-service/internal/service"
	"evidence-wall/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBoardService is a mock implementation of BoardService
type MockBoardService struct {
	mock.Mock
}

func (m *MockBoardService) CreateBoard(userID uuid.UUID, req service.CreateBoardRequest) (*models.Board, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Board), args.Error(1)
}

func (m *MockBoardService) GetBoard(boardID, userID uuid.UUID) (*models.BoardResponse, error) {
	args := m.Called(boardID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardResponse), args.Error(1)
}

func (m *MockBoardService) GetPublicBoard(boardID uuid.UUID) (*models.Board, error) {
	args := m.Called(boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Board), args.Error(1)
}

func (m *MockBoardService) ListBoards(userID uuid.UUID, offset, limit int) ([]models.BoardResponse, int64, error) {
	args := m.Called(userID, offset, limit)
	return args.Get(0).([]models.BoardResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockBoardService) UpdateBoard(boardID, userID uuid.UUID, req service.UpdateBoardRequest) (*models.Board, error) {
	args := m.Called(boardID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Board), args.Error(1)
}

func (m *MockBoardService) DeleteBoard(boardID, userID uuid.UUID) error {
	args := m.Called(boardID, userID)
	return args.Error(0)
}

func (m *MockBoardService) ShareBoard(boardID, ownerID uuid.UUID, req service.ShareBoardRequest) error {
	args := m.Called(boardID, ownerID, req)
	return args.Error(0)
}

func (m *MockBoardService) UnshareBoard(boardID, ownerID, targetUserID uuid.UUID) error {
	args := m.Called(boardID, ownerID, targetUserID)
	return args.Error(0)
}

func (m *MockBoardService) UpdateUserPermission(boardID, ownerID, targetUserID uuid.UUID, req service.UpdateUserPermissionRequest) error {
	args := m.Called(boardID, ownerID, targetUserID, req)
	return args.Error(0)
}

func (m *MockBoardService) CreateBoardItem(boardID, userID uuid.UUID, req service.CreateItemRequest) (*models.BoardItem, error) {
	args := m.Called(boardID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardItem), args.Error(1)
}

func (m *MockBoardService) UpdateBoardItem(boardID, itemID, userID uuid.UUID, req service.UpdateItemRequest) (*models.BoardItem, error) {
	args := m.Called(boardID, itemID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardItem), args.Error(1)
}

func (m *MockBoardService) DeleteBoardItem(boardID, itemID, userID uuid.UUID) error {
	args := m.Called(boardID, itemID, userID)
	return args.Error(0)
}

func (m *MockBoardService) ListBoardItems(boardID, userID uuid.UUID) ([]models.BoardItem, error) {
	args := m.Called(boardID, userID)
	return args.Get(0).([]models.BoardItem), args.Error(1)
}

func (m *MockBoardService) ListBoardConnections(boardID, userID uuid.UUID) ([]models.BoardConnection, error) {
	args := m.Called(boardID, userID)
	return args.Get(0).([]models.BoardConnection), args.Error(1)
}

func (m *MockBoardService) CreateBoardConnection(boardID, userID uuid.UUID, req service.CreateConnectionRequest) (*models.BoardConnection, error) {
	args := m.Called(boardID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardConnection), args.Error(1)
}

func (m *MockBoardService) UpdateBoardConnection(boardID, connectionID, userID uuid.UUID, req service.UpdateConnectionRequest) (*models.BoardConnection, error) {
	args := m.Called(boardID, connectionID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BoardConnection), args.Error(1)
}

func (m *MockBoardService) DeleteBoardConnection(boardID, connectionID, userID uuid.UUID) error {
	args := m.Called(boardID, connectionID, userID)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestBoardHandler_CreateBoard(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()

	tests := []struct {
		name           string
		userID         uuid.UUID
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		mockSetup      func(*MockBoardService)
	}{
		{
			name:   "successful board creation",
			userID: userID,
			requestBody: service.CreateBoardRequest{
				Title:       "Test Board",
				Description: "Test Description",
				Visibility:  models.VisibilityPrivate,
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				expectedReq := service.CreateBoardRequest{
					Title:       "Test Board",
					Description: "Test Description",
					Visibility:  models.VisibilityPrivate,
				}
				board := &models.Board{
					ID:          boardID,
					Title:       "Test Board",
					Description: "Test Description",
					Visibility:  models.VisibilityPrivate,
					OwnerID:     userID,
				}
				m.On("CreateBoard", userID, expectedReq).Return(board, nil)
			},
		},
		{
			name:           "invalid request body",
			userID:         userID,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot unmarshal",
			mockSetup:      func(m *MockBoardService) {},
		},
		{
			name:   "service error",
			userID: userID,
			requestBody: service.CreateBoardRequest{
				Title:       "Test Board",
				Description: "Test Description",
				Visibility:  models.VisibilityPrivate,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create board",
			mockSetup: func(m *MockBoardService) {
				expectedReq := service.CreateBoardRequest{
					Title:       "Test Board",
					Description: "Test Description",
					Visibility:  models.VisibilityPrivate,
				}
				m.On("CreateBoard", userID, expectedReq).Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBoardService)
			tt.mockSetup(mockService)

			handler := NewBoardHandler(mockService)
			router := setupTestRouter()

			// Add middleware to set user ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
			})

			router.POST("/boards", handler.CreateBoard)

			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/boards", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_GetBoard(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()

	tests := []struct {
		name           string
		userID         uuid.UUID
		boardID        string
		expectedStatus int
		expectedError  string
		mockSetup      func(*MockBoardService)
	}{
		{
			name:           "successful board retrieval",
			userID:         userID,
			boardID:        boardID.String(),
			expectedStatus: http.StatusOK,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				board := &models.BoardResponse{
					ID:          boardID,
					Title:       "Test Board",
					Description: "Test Description",
					Visibility:  models.VisibilityPrivate,
					OwnerID:     userID,
					Permission:  models.PermissionAdmin,
				}
				m.On("GetBoard", boardID, userID).Return(board, nil)
			},
		},
		{
			name:           "invalid board ID",
			userID:         userID,
			boardID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid board ID",
			mockSetup:      func(m *MockBoardService) {},
		},
		{
			name:           "board not found",
			userID:         userID,
			boardID:        boardID.String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  "Board not found",
			mockSetup: func(m *MockBoardService) {
				m.On("GetBoard", boardID, userID).Return(nil, service.ErrBoardNotFound)
			},
		},
		{
			name:           "service error",
			userID:         userID,
			boardID:        boardID.String(),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get board",
			mockSetup: func(m *MockBoardService) {
				m.On("GetBoard", boardID, userID).Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBoardService)
			tt.mockSetup(mockService)

			handler := NewBoardHandler(mockService)
			router := setupTestRouter()

			// Add middleware to set user ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
			})

			router.GET("/boards/:id", handler.GetBoard)

			// Prepare request
			req := httptest.NewRequest("GET", "/boards/"+tt.boardID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_GetPublicBoard(t *testing.T) {
	boardID := uuid.New()

	tests := []struct {
		name           string
		boardID        string
		expectedStatus int
		expectedError  string
		mockSetup      func(*MockBoardService)
	}{
		{
			name:           "successful public board retrieval",
			boardID:        boardID.String(),
			expectedStatus: http.StatusOK,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				board := &models.Board{
					ID:          boardID,
					Title:       "Public Board",
					Description: "Public Description",
					Visibility:  models.VisibilityPublic,
					OwnerID:     uuid.New(),
				}
				m.On("GetPublicBoard", boardID).Return(board, nil)
			},
		},
		{
			name:           "invalid board ID",
			boardID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid board ID",
			mockSetup:      func(m *MockBoardService) {},
		},
		{
			name:           "board not found",
			boardID:        boardID.String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  "Board not found",
			mockSetup: func(m *MockBoardService) {
				m.On("GetPublicBoard", boardID).Return(nil, service.ErrBoardNotFound)
			},
		},
		{
			name:           "private board access denied",
			boardID:        boardID.String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  "Board not found",
			mockSetup: func(m *MockBoardService) {
				m.On("GetPublicBoard", boardID).Return(nil, service.ErrUnauthorized)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBoardService)
			tt.mockSetup(mockService)

			handler := NewBoardHandler(mockService)
			router := setupTestRouter()

			router.GET("/public/boards/:id", handler.GetPublicBoard)

			// Prepare request
			req := httptest.NewRequest("GET", "/public/boards/"+tt.boardID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_ListBoards(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		userID         uuid.UUID
		queryParams    string
		expectedStatus int
		expectedError  string
		mockSetup      func(*MockBoardService)
	}{
		{
			name:           "successful board listing",
			userID:         userID,
			queryParams:    "?page=1&limit=10",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				boards := []models.BoardResponse{
					{
						ID:          uuid.New(),
						Title:       "Board 1",
						Description: "Description 1",
						Visibility:  models.VisibilityPrivate,
						OwnerID:     userID,
						Permission:  models.PermissionAdmin,
					},
				}
				m.On("ListBoards", userID, 0, 10).Return(boards, int64(1), nil)
			},
		},
		{
			name:           "default pagination",
			userID:         userID,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				boards := []models.BoardResponse{}
				m.On("ListBoards", userID, 0, 20).Return(boards, int64(0), nil)
			},
		},
		{
			name:           "invalid page parameter",
			userID:         userID,
			queryParams:    "?page=invalid&limit=10",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				boards := []models.BoardResponse{}
				m.On("ListBoards", userID, 0, 10).Return(boards, int64(0), nil)
			},
		},
		{
			name:           "service error",
			userID:         userID,
			queryParams:    "?page=1&limit=10",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to list boards",
			mockSetup: func(m *MockBoardService) {
				m.On("ListBoards", userID, 0, 10).Return([]models.BoardResponse{}, int64(0), errors.New("service error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBoardService)
			tt.mockSetup(mockService)

			handler := NewBoardHandler(mockService)
			router := setupTestRouter()

			// Add middleware to set user ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
			})

			router.GET("/boards", handler.ListBoards)

			// Prepare request
			req := httptest.NewRequest("GET", "/boards"+tt.queryParams, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_CreateBoardItem(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()
	itemID := uuid.New()

	tests := []struct {
		name           string
		userID         uuid.UUID
		boardID        string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		mockSetup      func(*MockBoardService)
	}{
		{
			name:    "successful item creation",
			userID:  userID,
			boardID: boardID.String(),
			requestBody: service.CreateItemRequest{
				Type:    models.ItemTypeNote,
				Content: "Test Note",
				X:       100,
				Y:       200,
				Width:   200,
				Height:  150,
				ZIndex:  1,
				Color:   "#ffff00",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
			mockSetup: func(m *MockBoardService) {
				expectedReq := service.CreateItemRequest{
					Type:    models.ItemTypeNote,
					Content: "Test Note",
					X:       100,
					Y:       200,
					Width:   200,
					Height:  150,
					ZIndex:  1,
					Color:   "#ffff00",
				}
				item := &models.BoardItem{
					ID:        itemID,
					BoardID:   boardID,
					Type:      models.ItemTypeNote,
					Content:   "Test Note",
					X:         100,
					Y:         200,
					Width:     200,
					Height:    150,
					ZIndex:    1,
					Color:     "#ffff00",
					CreatedBy: userID,
				}
				m.On("CreateBoardItem", boardID, userID, expectedReq).Return(item, nil)
			},
		},
		{
			name:           "invalid request body",
			userID:         userID,
			boardID:        boardID.String(),
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot unmarshal",
			mockSetup:      func(m *MockBoardService) {},
		},
		{
			name:    "board not found",
			userID:  userID,
			boardID: boardID.String(),
			requestBody: service.CreateItemRequest{
				Type:    models.ItemTypeNote,
				Content: "Test Note",
				X:       100,
				Y:       200,
				Width:   200,
				Height:  150,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Board not found",
			mockSetup: func(m *MockBoardService) {
				expectedReq := service.CreateItemRequest{
					Type:    models.ItemTypeNote,
					Content: "Test Note",
					X:       100,
					Y:       200,
					Width:   200,
					Height:  150,
				}
				m.On("CreateBoardItem", boardID, userID, expectedReq).Return(nil, service.ErrBoardNotFound)
			},
		},
		{
			name:    "unauthorized",
			userID:  userID,
			boardID: boardID.String(),
			requestBody: service.CreateItemRequest{
				Type:    models.ItemTypeNote,
				Content: "Test Note",
				X:       100,
				Y:       200,
				Width:   200,
				Height:  150,
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
			mockSetup: func(m *MockBoardService) {
				expectedReq := service.CreateItemRequest{
					Type:    models.ItemTypeNote,
					Content: "Test Note",
					X:       100,
					Y:       200,
					Width:   200,
					Height:  150,
				}
				m.On("CreateBoardItem", boardID, userID, expectedReq).Return(nil, service.ErrUnauthorized)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBoardService)
			tt.mockSetup(mockService)

			handler := NewBoardHandler(mockService)
			router := setupTestRouter()

			// Add middleware to set user ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
			})

			router.POST("/boards/:id/items", handler.CreateBoardItem)

			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/boards/"+tt.boardID+"/items", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}
