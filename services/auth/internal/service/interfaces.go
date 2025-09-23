package service

import (
	"evidence-wall/shared/auth"
	"evidence-wall/shared/models"

	"github.com/google/uuid"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByGoogleID(googleID string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	EmailExists(email string) (bool, error)
	GoogleIDExists(googleID string) (bool, error)
}

// JWTManagerInterface defines the interface for JWT operations
type JWTManagerInterface interface {
	GenerateToken(userID uuid.UUID, email, name string) (string, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
	RefreshToken(tokenString string) (string, error)
}

// AuthServiceInterface defines the interface for authentication service operations
type AuthServiceInterface interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	RefreshToken(tokenString string) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*models.UserResponse, error)
	UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*models.UserResponse, error)
	GetGoogleLoginURL(state string) (string, error)
	GoogleCallback(code string) (*AuthResponse, error)
}
