package service

import (
	"evidence-wall/shared/models"

	"github.com/google/uuid"
)

// AuthServiceInterface defines the contract for the authentication service used by HTTP handlers.
type AuthServiceInterface interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	RefreshToken(token string) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*models.UserResponse, error)
	UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*models.UserResponse, error)
	GetGoogleLoginURL(state string) string
	GoogleCallback(code string) (*AuthResponse, error)
}
