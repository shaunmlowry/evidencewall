package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"evidence-wall/auth-service/internal/config"
	"evidence-wall/auth-service/internal/repository"
	"evidence-wall/shared/auth"
	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrGoogleIDExists     = errors.New("google account already linked")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo    *repository.UserRepository
	jwtManager  *auth.JWTManager
	config      *config.Config
	oauthConfig *oauth2.Config
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager, cfg *config.Config) *AuthService {
	var oauthConfig *oauth2.Config
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		}
	}

	return &AuthService{
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		config:      cfg,
		oauthConfig: oauthConfig,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User  models.UserResponse `json:"user"`
	Token string              `json:"token"`
}

// Register registers a new user
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
		Verified: false, // Email verification can be implemented later
		Active:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: token,
	}, nil
}

// RefreshToken generates a new token from an existing one
func (s *AuthService) RefreshToken(tokenString string) (*AuthResponse, error) {
	// Validate and refresh token
	newToken, err := s.jwtManager.RefreshToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user from token claims
	claims, err := s.jwtManager.ValidateToken(newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate new token: %w", err)
	}

	// Get updated user data
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: newToken,
	}, nil
}

// GetProfile gets user profile by ID
func (s *AuthService) GetProfile(userID uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2,max=100"`
	Avatar string `json:"avatar" binding:"omitempty,url"`
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// GetGoogleLoginURL returns the Google OAuth login URL
func (s *AuthService) GetGoogleLoginURL(state string) string {
	if s.oauthConfig == nil {
		return ""
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GoogleCallback handles Google OAuth callback
func (s *AuthService) GoogleCallback(code string) (*AuthResponse, error) {
	if s.oauthConfig == nil {
		return nil, errors.New("Google OAuth not configured")
	}

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google
	client := s.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	// Check if user exists by Google ID
	user, err := s.userRepo.GetByGoogleID(googleUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Google ID: %w", err)
	}

	if user == nil {
		// Check if user exists by email
		user, err = s.userRepo.GetByEmail(googleUser.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get user by email: %w", err)
		}

		if user != nil {
            // Link Google account to existing user
            user.GoogleID = &googleUser.ID
			if user.Avatar == "" {
				user.Avatar = googleUser.Picture
			}
			if err := s.userRepo.Update(user); err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		} else {
			// Create new user
            user = &models.User{
				Email:    googleUser.Email,
				Name:     googleUser.Name,
				Avatar:   googleUser.Picture,
                GoogleID: func() *string { v := googleUser.ID; return &v }(),
				Password: uuid.New().String(), // Random password for Google users
				Verified: true,                // Google accounts are pre-verified
				Active:   true,
			}

			if err := s.userRepo.Create(user); err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	// Generate JWT token
	jwtToken, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		User:  user.ToResponse(),
		Token: jwtToken,
	}, nil
}


