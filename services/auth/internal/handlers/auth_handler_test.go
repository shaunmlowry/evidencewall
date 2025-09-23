package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"evidence-wall/auth-service/internal/service"
	"evidence-wall/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req service.RegisterRequest) (*service.AuthResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(req service.LoginRequest) (*service.AuthResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(token string) (*service.AuthResponse, error) {
	args := m.Called(token)
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetProfile(userID uuid.UUID) (*models.UserResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) UpdateProfile(userID uuid.UUID, req service.UpdateProfileRequest) (*models.UserResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockAuthService) GetGoogleLoginURL(state string) (string, error) {
	args := m.Called(state)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GoogleCallback(code string) (*service.AuthResponse, error) {
	args := m.Called(code)
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    service.RegisterRequest
		mockResponse   *service.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful registration",
			requestBody: service.RegisterRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockResponse: &service.AuthResponse{
				User: models.UserResponse{
					ID:    uuid.New(),
					Email: "test@example.com",
					Name:  "Test User",
				},
				Token: "mock-jwt-token",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "email already exists",
			requestBody: service.RegisterRequest{
				Email:    "existing@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockResponse:   nil,
			mockError:      service.ErrEmailExists,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.mockResponse != nil {
				mockService.On("Register", tt.requestBody).Return(tt.mockResponse, tt.mockError)
			} else {
				mockService.On("Register", tt.requestBody).Return((*service.AuthResponse)(nil), tt.mockError)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.Register(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response service.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.User.Email, response.User.Email)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    service.LoginRequest
		mockResponse   *service.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: service.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockResponse: &service.AuthResponse{
				User: models.UserResponse{
					ID:    uuid.New(),
					Email: "test@example.com",
					Name:  "Test User",
				},
				Token: "mock-jwt-token",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: service.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockResponse:   nil,
			mockError:      service.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid request body",
			requestBody: service.LoginRequest{
				Email: "invalid-email",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.mockResponse != nil {
				mockService.On("Login", tt.requestBody).Return(tt.mockResponse, tt.mockError)
			} else if tt.mockError != nil {
				mockService.On("Login", tt.requestBody).Return((*service.AuthResponse)(nil), tt.mockError)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.Login(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response service.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.User.Email, response.User.Email)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]string
		mockResponse   *service.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful token refresh",
			requestBody: map[string]string{
				"token": "valid-refresh-token",
			},
			mockResponse: &service.AuthResponse{
				User: models.UserResponse{
					ID:    uuid.New(),
					Email: "test@example.com",
					Name:  "Test User",
				},
				Token: "new-jwt-token",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			requestBody: map[string]string{
				"token": "invalid-token",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid token"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing token",
			requestBody: map[string]string{
				"token": "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.mockResponse != nil {
				mockService.On("RefreshToken", tt.requestBody["token"]).Return(tt.mockResponse, tt.mockError)
			} else if tt.mockError != nil {
				mockService.On("RefreshToken", tt.requestBody["token"]).Return((*service.AuthResponse)(nil), tt.mockError)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.RefreshToken(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response service.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.User.Email, response.User.Email)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tests := []struct {
		name           string
		userID         uuid.UUID
		userIDExists   bool
		mockResponse   *models.UserResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:         "successful profile retrieval",
			userID:       userID,
			userIDExists: true,
			mockResponse: &models.UserResponse{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			userID:         userID,
			userIDExists:   true,
			mockResponse:   nil,
			mockError:      service.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "user not authenticated",
			userID:         userID,
			userIDExists:   false,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.userIDExists && tt.mockResponse != nil {
				mockService.On("GetProfile", tt.userID).Return(tt.mockResponse, tt.mockError)
			} else if tt.userIDExists && tt.mockError != nil {
				mockService.On("GetProfile", tt.userID).Return((*models.UserResponse)(nil), tt.mockError)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/me", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Mock middleware behavior
			if tt.userIDExists {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.GetProfile(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Email, response.Email)
				assert.Equal(t, tt.mockResponse.Name, response.Name)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_UpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tests := []struct {
		name           string
		userID         uuid.UUID
		userIDExists   bool
		requestBody    service.UpdateProfileRequest
		mockResponse   *models.UserResponse
		mockError      error
		expectedStatus int
	}{
		{
			name:         "successful profile update",
			userID:       userID,
			userIDExists: true,
			requestBody: service.UpdateProfileRequest{
				Name:   "Updated Name",
				Avatar: "https://example.com/avatar.jpg",
			},
			mockResponse: &models.UserResponse{
				ID:     userID,
				Email:  "test@example.com",
				Name:   "Updated Name",
				Avatar: "https://example.com/avatar.jpg",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user not found",
			userID:       userID,
			userIDExists: true,
			requestBody: service.UpdateProfileRequest{
				Name: "Updated Name",
			},
			mockResponse:   nil,
			mockError:      service.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "user not authenticated",
			userID:       userID,
			userIDExists: false,
			requestBody: service.UpdateProfileRequest{
				Name: "Updated Name",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid request body",
			userID:       userID,
			userIDExists: true,
			requestBody: service.UpdateProfileRequest{
				Name: "A", // Too short
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.userIDExists && tt.mockResponse != nil {
				mockService.On("UpdateProfile", tt.userID, tt.requestBody).Return(tt.mockResponse, tt.mockError)
			} else if tt.userIDExists && tt.mockError != nil {
				mockService.On("UpdateProfile", tt.userID, tt.requestBody).Return((*models.UserResponse)(nil), tt.mockError)
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Mock middleware behavior
			if tt.userIDExists {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.UpdateProfile(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Email, response.Email)
				assert.Equal(t, tt.mockResponse.Name, response.Name)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_GoogleLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockURL        string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful Google login URL generation",
			mockURL:        "https://accounts.google.com/oauth/authorize?client_id=test&redirect_uri=test&response_type=code&scope=openid+profile+email&state=test",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Google OAuth not configured",
			mockURL:        "",
			mockError:      errors.New("Google OAuth not configured"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			// Mock the GetGoogleLoginURL call
			mockService.On("GetGoogleLoginURL", mock.AnythingOfType("string")).Return(tt.mockURL, tt.mockError)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/auth/google", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.GoogleLogin(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "url")
				assert.Contains(t, response, "state")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_GoogleCallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		code           string
		mockResponse   *service.AuthResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful Google callback",
			code: "valid-auth-code",
			mockResponse: &service.AuthResponse{
				User: models.UserResponse{
					ID:    uuid.New(),
					Email: "test@gmail.com",
					Name:  "Google User",
				},
				Token: "google-jwt-token",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization code",
			code:           "",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Google authentication failed",
			code:           "invalid-code",
			mockResponse:   nil,
			mockError:      errors.New("authentication failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.mockResponse != nil {
				mockService.On("GoogleCallback", tt.code).Return(tt.mockResponse, tt.mockError)
			} else if tt.mockError != nil {
				mockService.On("GoogleCallback", tt.code).Return((*service.AuthResponse)(nil), tt.mockError)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/auth/google/callback", nil)
			if tt.code != "" {
				q := req.URL.Query()
				q.Add("code", tt.code)
				req.URL.RawQuery = q.Encode()
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.GoogleCallback(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response service.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.User.Email, response.User.Email)
				assert.Equal(t, tt.mockResponse.Token, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Create handler
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	// Call handler
	handler.Logout(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}
