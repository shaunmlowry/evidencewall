package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionLevel represents the level of access a user has to a board
type PermissionLevel string

const (
	PermissionRead      PermissionLevel = "read"
	PermissionReadWrite PermissionLevel = "read_write"
	PermissionAdmin     PermissionLevel = "admin"
)

// BoardVisibility represents who can access the board
type BoardVisibility string

const (
	VisibilityPrivate BoardVisibility = "private"
	VisibilityShared  BoardVisibility = "shared"
	VisibilityPublic  BoardVisibility = "public"
)

// Board represents an evidence board
type Board struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string          `json:"title" gorm:"not null"`
	Description string          `json:"description"`
	Visibility  BoardVisibility `json:"visibility" gorm:"default:'private'"`
	OwnerID     uuid.UUID       `json:"owner_id" gorm:"type:uuid;not null"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`

	// Relationships
	Owner BoardUser   `json:"owner,omitempty" gorm:"foreignKey:OwnerID;references:UserID"`
	Users []BoardUser `json:"users,omitempty" gorm:"foreignKey:BoardID"`
	Items []BoardItem `json:"items,omitempty" gorm:"foreignKey:BoardID"`
}

// BoardUser represents the many-to-many relationship between boards and users
type BoardUser struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BoardID    uuid.UUID       `json:"board_id" gorm:"type:uuid;not null"`
	UserID     uuid.UUID       `json:"user_id" gorm:"type:uuid;not null"`
	Permission PermissionLevel `json:"permission" gorm:"not null"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`

	// Relationships
	Board Board `json:"board,omitempty" gorm:"foreignKey:BoardID"`
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BoardItem represents an item on the board (post-it note or suspect card)
type BoardItem struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BoardID   uuid.UUID      `json:"board_id" gorm:"type:uuid;not null"`
	Type      string         `json:"type" gorm:"not null"` // "post-it" or "suspect-card"
	X         float64        `json:"x" gorm:"not null"`
	Y         float64        `json:"y" gorm:"not null"`
	Width     float64        `json:"width" gorm:"default:200"`
	Height    float64        `json:"height" gorm:"default:200"`
	Rotation  float64        `json:"rotation" gorm:"default:0"`
	ZIndex    int            `json:"z_index" gorm:"default:1"`
	Content   string         `json:"content"`
	Style     string         `json:"style"` // JSON string for styling properties
	CreatedBy uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Board       Board             `json:"board,omitempty" gorm:"foreignKey:BoardID"`
	Creator     User              `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Connections []BoardConnection `json:"connections,omitempty" gorm:"many2many:board_item_connections;"`
}

// BoardConnection represents a connection between two board items
type BoardConnection struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BoardID    uuid.UUID      `json:"board_id" gorm:"type:uuid;not null"`
	FromItemID uuid.UUID      `json:"from_item_id" gorm:"type:uuid;not null"`
	ToItemID   uuid.UUID      `json:"to_item_id" gorm:"type:uuid;not null"`
	Style      string         `json:"style"` // JSON string for connection styling
	CreatedBy  uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Board    Board     `json:"board,omitempty" gorm:"foreignKey:BoardID"`
	FromItem BoardItem `json:"from_item,omitempty" gorm:"foreignKey:FromItemID"`
	ToItem   BoardItem `json:"to_item,omitempty" gorm:"foreignKey:ToItemID"`
	Creator  User      `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// BeforeCreate hooks
func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (bu *BoardUser) BeforeCreate(tx *gorm.DB) error {
	if bu.ID == uuid.Nil {
		bu.ID = uuid.New()
	}
	return nil
}

func (bi *BoardItem) BeforeCreate(tx *gorm.DB) error {
	if bi.ID == uuid.Nil {
		bi.ID = uuid.New()
	}
	return nil
}

func (bc *BoardConnection) BeforeCreate(tx *gorm.DB) error {
	if bc.ID == uuid.Nil {
		bc.ID = uuid.New()
	}
	return nil
}

// BoardResponse represents the board data returned to clients
type BoardResponse struct {
	ID          uuid.UUID           `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Visibility  BoardVisibility     `json:"visibility"`
	OwnerID     uuid.UUID           `json:"owner_id"`
	Permission  PermissionLevel     `json:"permission,omitempty"` // User's permission level
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Items       []BoardItem         `json:"items,omitempty"`
	Connections []BoardConnection   `json:"connections,omitempty"`
	Users       []BoardUserResponse `json:"users,omitempty"`
}

// BoardUserResponse represents board user data returned to clients
type BoardUserResponse struct {
	User       UserResponse    `json:"user"`
	Permission PermissionLevel `json:"permission"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ToResponse converts Board to BoardResponse
func (b *Board) ToResponse(userPermission PermissionLevel) BoardResponse {
	response := BoardResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Visibility:  b.Visibility,
		OwnerID:     b.OwnerID,
		Permission:  userPermission,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		Items:       b.Items,
		Connections: b.Connections,
	}

	// Convert users
	for _, bu := range b.Users {
		response.Users = append(response.Users, BoardUserResponse{
			User:       bu.User.ToResponse(),
			Permission: bu.Permission,
			CreatedAt:  bu.CreatedAt,
		})
	}

	return response
}


