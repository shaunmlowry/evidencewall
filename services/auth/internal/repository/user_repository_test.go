package repository

import (
	"testing"

	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create the users table manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			avatar TEXT,
			password TEXT NOT NULL,
			google_id TEXT UNIQUE,
			verified INTEGER DEFAULT 0,
			active INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)

	// Verify user was created
	var foundUser models.User
	err = db.First(&foundUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Name, foundUser.Name)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		userID   uuid.UUID
		expected *models.User
		hasError bool
	}{
		{
			name:     "existing user",
			userID:   user.ID,
			expected: user,
			hasError: false,
		},
		{
			name:     "non-existing user",
			userID:   uuid.New(),
			expected: nil,
			hasError: false,
		},
		{
			name:     "inactive user",
			userID:   user.ID,
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inactive user" {
				// Make user inactive
				user.Active = false
				err := db.Save(user).Error
				assert.NoError(t, err)
			}

			result, err := repo.GetByID(tt.userID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.Email, result.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		expected *models.User
		hasError bool
	}{
		{
			name:     "existing user",
			email:    "test@example.com",
			expected: user,
			hasError: false,
		},
		{
			name:     "non-existing user",
			email:    "nonexistent@example.com",
			expected: nil,
			hasError: false,
		},
		{
			name:     "inactive user",
			email:    "test@example.com",
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inactive user" {
				// Make user inactive
				user.Active = false
				err := db.Save(user).Error
				assert.NoError(t, err)
			}

			result, err := repo.GetByEmail(tt.email)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.Email, result.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetByGoogleID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	googleID := "google123"
	// Create a test user with Google ID
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		GoogleID: &googleID,
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		googleID string
		expected *models.User
		hasError bool
	}{
		{
			name:     "existing user with Google ID",
			googleID: "google123",
			expected: user,
			hasError: false,
		},
		{
			name:     "non-existing Google ID",
			googleID: "nonexistent",
			expected: nil,
			hasError: false,
		},
		{
			name:     "inactive user",
			googleID: "google123",
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inactive user" {
				// Make user inactive
				user.Active = false
				err := db.Save(user).Error
				assert.NoError(t, err)
			}

			result, err := repo.GetByGoogleID(tt.googleID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.GoogleID, result.GoogleID)
				}
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Update user
	user.Name = "Updated Name"
	user.Avatar = "https://example.com/avatar.jpg"

	err = repo.Update(user)
	assert.NoError(t, err)

	// Verify update
	var updatedUser models.User
	err = db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", updatedUser.Avatar)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Delete user
	err = repo.Delete(user.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	var deletedUser models.User
	err = db.Unscoped().First(&deletedUser, user.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, deletedUser.DeletedAt)

	// Verify user is not found in normal query
	err = db.First(&deletedUser, user.ID).Error
	assert.Error(t, err)
}

func TestUserRepository_EmailExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		expected bool
		hasError bool
	}{
		{
			name:     "existing email",
			email:    "test@example.com",
			expected: true,
			hasError: false,
		},
		{
			name:     "non-existing email",
			email:    "nonexistent@example.com",
			expected: false,
			hasError: false,
		},
		{
			name:     "inactive user email",
			email:    "test@example.com",
			expected: false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inactive user email" {
				// Make user inactive
				user.Active = false
				err := db.Save(user).Error
				assert.NoError(t, err)
			}

			result, err := repo.EmailExists(tt.email)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUserRepository_GoogleIDExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	googleID := "google123"
	// Create a test user with Google ID
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		GoogleID: &googleID,
		Active:   true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		googleID string
		expected bool
		hasError bool
	}{
		{
			name:     "existing Google ID",
			googleID: "google123",
			expected: true,
			hasError: false,
		},
		{
			name:     "non-existing Google ID",
			googleID: "nonexistent",
			expected: false,
			hasError: false,
		},
		{
			name:     "inactive user Google ID",
			googleID: "google123",
			expected: false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inactive user Google ID" {
				// Make user inactive
				user.Active = false
				err := db.Save(user).Error
				assert.NoError(t, err)
			}

			result, err := repo.GoogleIDExists(tt.googleID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create test users
	users := []*models.User{
		{
			ID:       uuid.New(),
			Email:    "user1@example.com",
			Name:     "User 1",
			Password: "hashedpassword",
			Active:   true,
		},
		{
			ID:       uuid.New(),
			Email:    "user2@example.com",
			Name:     "User 2",
			Password: "hashedpassword",
			Active:   true,
		},
		{
			ID:       uuid.New(),
			Email:    "user3@example.com",
			Name:     "User 3",
			Password: "hashedpassword",
			Active:   false, // Inactive user
		},
	}

	for _, user := range users {
		// Use raw SQL to insert the user to bypass GORM defaults
		err := db.Exec(`
			INSERT INTO users (id, email, name, password, active, verified, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		`, user.ID, user.Email, user.Name, user.Password, user.Active, user.Verified).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		offset        int
		limit         int
		expectedCount int
		expectedTotal int64
		hasError      bool
	}{
		{
			name:          "list all active users",
			offset:        0,
			limit:         10,
			expectedCount: 2, // Only active users
			expectedTotal: 2,
			hasError:      false,
		},
		{
			name:          "list with pagination",
			offset:        0,
			limit:         1,
			expectedCount: 1,
			expectedTotal: 2,
			hasError:      false,
		},
		{
			name:          "list with offset",
			offset:        1,
			limit:         1,
			expectedCount: 1,
			expectedTotal: 2,
			hasError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, total, err := repo.List(tt.offset, tt.limit)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(users))
				assert.Equal(t, tt.expectedTotal, total)

				// Verify all returned users are active
				for _, user := range users {
					assert.True(t, user.Active)
				}
			}
		})
	}
}
