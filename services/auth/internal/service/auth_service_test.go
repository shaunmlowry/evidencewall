package service

import (
	"errors"
	"testing"

	"evidence-wall/auth-service/internal/config"
	"evidence-wall/shared/auth"
	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByGoogleID(googleID string) (*models.User, error) {
	args := m.Called(googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GoogleIDExists(googleID string) (bool, error) {
	args := m.Called(googleID)
	return args.Bool(0), args.Error(1)
}

// MockJWTManager is a mock implementation of JWTManagerInterface
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID uuid.UUID, email, name string) (string, error) {
	args := m.Called(userID, email, name)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockJWTManager) RefreshToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name           string
		request        RegisterRequest
		emailExists    bool
		emailExistsErr error
		createErr      error
		generateErr    error
		expectedErr    error
	}{
		{
			name: "successful registration",
			request: RegisterRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			emailExists:    false,
			emailExistsErr: nil,
			createErr:      nil,
			generateErr:    nil,
			expectedErr:    nil,
		},
		{
			name: "email already exists",
			request: RegisterRequest{
				Email:    "existing@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			emailExists:    true,
			emailExistsErr: nil,
			createErr:      nil,
			generateErr:    nil,
			expectedErr:    ErrEmailExists,
		},
		{
			name: "email check error",
			request: RegisterRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			emailExists:    false,
			emailExistsErr: errors.New("database error"),
			createErr:      nil,
			generateErr:    nil,
			expectedErr:    errors.New("failed to check email existence: database error"),
		},
		{
			name: "user creation error",
			request: RegisterRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			emailExists:    false,
			emailExistsErr: nil,
			createErr:      errors.New("database error"),
			generateErr:    nil,
			expectedErr:    errors.New("failed to create user: database error"),
		},
		{
			name: "token generation error",
			request: RegisterRequest{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			emailExists:    false,
			emailExistsErr: nil,
			createErr:      nil,
			generateErr:    errors.New("jwt error"),
			expectedErr:    errors.New("failed to generate token: jwt error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			cfg := &config.Config{}

			service := NewAuthService(mockRepo, mockJWT, cfg)

			// Setup mocks
			mockRepo.On("EmailExists", tt.request.Email).Return(tt.emailExists, tt.emailExistsErr)

			if tt.emailExistsErr == nil && !tt.emailExists {
				// Mock user creation
				mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(tt.createErr)

				if tt.createErr == nil {
					// Mock JWT generation
					mockJWT.On("GenerateToken", mock.AnythingOfType("uuid.UUID"), tt.request.Email, tt.request.Name).Return("test-token", tt.generateErr)
				}
			}

			// Call method
			result, err := service.Register(tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Email, result.User.Email)
				assert.Equal(t, tt.request.Name, result.User.Name)
				assert.Equal(t, "test-token", result.Token)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	userID := uuid.New()
	hashedPassword := "$2a$10$xPBk8LAx8mtkjmcr/K6J8.g4o.UC5xZ8U6IYKk4d8j7.R7uWgBnXC" // "password123"

	tests := []struct {
		name        string
		request     LoginRequest
		user        *models.User
		userErr     error
		generateErr error
		expectedErr error
	}{
		{
			name: "successful login",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			user: &models.User{
				ID:       userID,
				Email:    "test@example.com",
				Name:     "Test User",
				Password: hashedPassword,
			},
			userErr:     nil,
			generateErr: nil,
			expectedErr: nil,
		},
		{
			name: "user not found",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			user:        nil,
			userErr:     nil,
			generateErr: nil,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "database error",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			user:        nil,
			userErr:     errors.New("database error"),
			generateErr: nil,
			expectedErr: errors.New("failed to get user: database error"),
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			user: &models.User{
				ID:       userID,
				Email:    "test@example.com",
				Name:     "Test User",
				Password: hashedPassword,
			},
			userErr:     nil,
			generateErr: nil,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name: "token generation error",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			user: &models.User{
				ID:       userID,
				Email:    "test@example.com",
				Name:     "Test User",
				Password: hashedPassword,
			},
			userErr:     nil,
			generateErr: errors.New("jwt error"),
			expectedErr: errors.New("failed to generate token: jwt error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			cfg := &config.Config{}

			service := NewAuthService(mockRepo, mockJWT, cfg)

			// Setup mocks
			mockRepo.On("GetByEmail", tt.request.Email).Return(tt.user, tt.userErr)

			if tt.userErr == nil && tt.user != nil && (tt.expectedErr == nil || tt.generateErr != nil) {
				// Mock JWT generation when we have a user and either expect success or have a token error
				mockJWT.On("GenerateToken", tt.user.ID, tt.user.Email, tt.user.Name).Return("test-token", tt.generateErr)
			}

			// Call method
			result, err := service.Login(tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Email, result.User.Email)
				assert.Equal(t, "test-token", result.Token)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		token       string
		refreshErr  error
		validateErr error
		user        *models.User
		userErr     error
		expectedErr error
	}{
		{
			name:        "successful token refresh",
			token:       "valid-token",
			refreshErr:  nil,
			validateErr: nil,
			user: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
			userErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "refresh token error",
			token:       "invalid-token",
			refreshErr:  errors.New("invalid token"),
			validateErr: nil,
			user:        nil,
			userErr:     nil,
			expectedErr: errors.New("failed to refresh token: invalid token"),
		},
		{
			name:        "validate token error",
			token:       "valid-token",
			refreshErr:  nil,
			validateErr: errors.New("invalid token"),
			user:        nil,
			userErr:     nil,
			expectedErr: errors.New("failed to validate new token: invalid token"),
		},
		{
			name:        "user not found",
			token:       "valid-token",
			refreshErr:  nil,
			validateErr: nil,
			user:        nil,
			userErr:     nil,
			expectedErr: ErrUserNotFound,
		},
		{
			name:        "user database error",
			token:       "valid-token",
			refreshErr:  nil,
			validateErr: nil,
			user:        nil,
			userErr:     errors.New("database error"),
			expectedErr: errors.New("failed to get user: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			cfg := &config.Config{}

			service := NewAuthService(mockRepo, mockJWT, cfg)

			// Setup mocks
			mockJWT.On("RefreshToken", tt.token).Return("new-token", tt.refreshErr)

			if tt.refreshErr == nil {
				claims := &auth.Claims{UserID: userID}
				mockJWT.On("ValidateToken", "new-token").Return(claims, tt.validateErr)

				if tt.validateErr == nil {
					mockRepo.On("GetByID", userID).Return(tt.user, tt.userErr)
				}
			}

			// Call method
			result, err := service.RefreshToken(tt.token)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "new-token", result.Token)
			}

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetProfile(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		user        *models.User
		userErr     error
		expectedErr error
	}{
		{
			name:   "successful profile retrieval",
			userID: userID,
			user: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
			userErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "user not found",
			userID:      userID,
			user:        nil,
			userErr:     nil,
			expectedErr: ErrUserNotFound,
		},
		{
			name:        "database error",
			userID:      userID,
			user:        nil,
			userErr:     errors.New("database error"),
			expectedErr: errors.New("failed to get user: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			cfg := &config.Config{}

			service := NewAuthService(mockRepo, mockJWT, cfg)

			// Setup mocks
			mockRepo.On("GetByID", tt.userID).Return(tt.user, tt.userErr)

			// Call method
			result, err := service.GetProfile(tt.userID)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.user.Email, result.Email)
				assert.Equal(t, tt.user.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_UpdateProfile(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		request     UpdateProfileRequest
		user        *models.User
		userErr     error
		updateErr   error
		expectedErr error
	}{
		{
			name:   "successful profile update",
			userID: userID,
			request: UpdateProfileRequest{
				Name:   "Updated Name",
				Avatar: "https://example.com/avatar.jpg",
			},
			user: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
			userErr:     nil,
			updateErr:   nil,
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userID: userID,
			request: UpdateProfileRequest{
				Name: "Updated Name",
			},
			user:        nil,
			userErr:     nil,
			updateErr:   nil,
			expectedErr: ErrUserNotFound,
		},
		{
			name:   "database error on get",
			userID: userID,
			request: UpdateProfileRequest{
				Name: "Updated Name",
			},
			user:        nil,
			userErr:     errors.New("database error"),
			updateErr:   nil,
			expectedErr: errors.New("failed to get user: database error"),
		},
		{
			name:   "database error on update",
			userID: userID,
			request: UpdateProfileRequest{
				Name: "Updated Name",
			},
			user: &models.User{
				ID:    userID,
				Email: "test@example.com",
				Name:  "Test User",
			},
			userErr:     nil,
			updateErr:   errors.New("database error"),
			expectedErr: errors.New("failed to update user: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			cfg := &config.Config{}

			service := NewAuthService(mockRepo, mockJWT, cfg)

			// Setup mocks
			mockRepo.On("GetByID", tt.userID).Return(tt.user, tt.userErr)

			if tt.userErr == nil && tt.user != nil {
				mockRepo.On("Update", mock.AnythingOfType("*models.User")).Return(tt.updateErr)
			}

			// Call method
			result, err := service.UpdateProfile(tt.userID, tt.request)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.user.Email, result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetGoogleLoginURL(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		state       string
		expectedURL string
		expectError bool
	}{
		{
			name: "OAuth configured",
			config: &config.Config{
				GoogleClientID:     "test-client-id",
				GoogleClientSecret: "test-client-secret",
				GoogleRedirectURL:  "http://localhost:8080/auth/google/callback",
			},
			state:       "test-state",
			expectedURL: "https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=test-client-id&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fgoogle%2Fcallback&response_type=code&scope=openid+profile+email&state=test-state",
			expectError: false,
		},
		{
			name:        "OAuth not configured",
			config:      &config.Config{},
			state:       "test-state",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)

			service := NewAuthService(mockRepo, mockJWT, tt.config)

			// Call method
			result, err := service.GetGoogleLoginURL(tt.state)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, result, "accounts.google.com")
				assert.Contains(t, result, tt.state)
			}
		})
	}
}
