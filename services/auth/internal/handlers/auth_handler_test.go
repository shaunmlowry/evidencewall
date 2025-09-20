package handlers

import (
	"bytes"
	"encoding/json"
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

func (m *MockAuthService) GetGoogleLoginURL(state string) string {
	args := m.Called(state)
	return args.String(0)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			handler := NewAuthHandler(mockService)

			if tt.mockResponse != nil {
				mockService.On("Login", tt.requestBody).Return(tt.mockResponse, tt.mockError)
			} else {
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


