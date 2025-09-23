package repository

import (
	"testing"

	"evidence-wall/shared/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupItemTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create tables manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE board_items (
			id TEXT PRIMARY KEY,
			board_id TEXT NOT NULL,
			type TEXT NOT NULL,
			x REAL NOT NULL,
			y REAL NOT NULL,
			width REAL DEFAULT 200,
			height REAL DEFAULT 200,
			rotation REAL DEFAULT 0,
			z_index INTEGER DEFAULT 1,
			content TEXT,
			color TEXT,
			metadata TEXT,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE board_connections (
			id TEXT PRIMARY KEY,
			board_id TEXT NOT NULL,
			from_item_id TEXT NOT NULL,
			to_item_id TEXT NOT NULL,
			style TEXT,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestBoardItemRepository_Create(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	item := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Width:     200.0,
		Height:    150.0,
		ZIndex:    1,
		Content:   "Test Note",
		Color:     "#ffff00",
		CreatedBy: uuid.New(),
	}

	err := repo.Create(item)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.NotZero(t, item.CreatedAt)
	assert.NotZero(t, item.UpdatedAt)

	// Verify item was created
	var foundItem models.BoardItem
	err = db.First(&foundItem, item.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, item.Content, foundItem.Content)
	assert.Equal(t, item.X, foundItem.X)
	assert.Equal(t, item.Y, foundItem.Y)
	assert.Equal(t, item.Type, foundItem.Type)
}

func TestBoardItemRepository_GetByID(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	// Create a test item
	item := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Width:     200.0,
		Height:    150.0,
		Content:   "Test Note",
		CreatedBy: uuid.New(),
	}
	err := db.Create(item).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		itemID   uuid.UUID
		expected *models.BoardItem
		hasError bool
	}{
		{
			name:     "existing item",
			itemID:   item.ID,
			expected: item,
			hasError: false,
		},
		{
			name:     "non-existing item",
			itemID:   uuid.New(),
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.itemID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.Content, result.Content)
				}
			}
		})
	}
}

func TestBoardItemRepository_ListByBoard(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	boardID := uuid.New()
	userID := uuid.New()

	// Create test items
	item1 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   boardID,
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Width:     200.0,
		Height:    150.0,
		ZIndex:    2,
		Content:   "Item 1",
		CreatedBy: userID,
	}
	err := db.Create(item1).Error
	assert.NoError(t, err)

	item2 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   boardID,
		Type:      models.ItemTypeNote,
		X:         300.0,
		Y:         400.0,
		Width:     200.0,
		Height:    150.0,
		ZIndex:    1,
		Content:   "Item 2",
		CreatedBy: userID,
	}
	err = db.Create(item2).Error
	assert.NoError(t, err)

	// Create item for different board
	item3 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         500.0,
		Y:         600.0,
		Width:     200.0,
		Height:    150.0,
		Content:   "Item 3",
		CreatedBy: userID,
	}
	err = db.Create(item3).Error
	assert.NoError(t, err)

	// Test listing items for board
	items, err := repo.ListByBoard(boardID)
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Verify items are ordered by z_index ASC, created_at ASC
	assert.Equal(t, item2.ID, items[0].ID) // z_index 1
	assert.Equal(t, item1.ID, items[1].ID) // z_index 2
}

func TestBoardItemRepository_Update(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	// Create a test item
	item := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Width:     200.0,
		Height:    150.0,
		Content:   "Original Content",
		CreatedBy: uuid.New(),
	}
	err := db.Create(item).Error
	assert.NoError(t, err)

	// Update the item
	item.Content = "Updated Content"
	item.X = 150.0
	item.Y = 250.0
	item.ZIndex = 5

	err = repo.Update(item)
	assert.NoError(t, err)

	// Verify item was updated
	var foundItem models.BoardItem
	err = db.First(&foundItem, item.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Content", foundItem.Content)
	assert.Equal(t, 150.0, foundItem.X)
	assert.Equal(t, 250.0, foundItem.Y)
	assert.Equal(t, 5, foundItem.ZIndex)
}

func TestBoardItemRepository_Delete(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	// Create a test item
	item := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Width:     200.0,
		Height:    150.0,
		Content:   "Test Item",
		CreatedBy: uuid.New(),
	}
	err := db.Create(item).Error
	assert.NoError(t, err)

	// Delete the item
	err = repo.Delete(item.ID)
	assert.NoError(t, err)

	// Verify item was deleted (hard delete)
	var foundItem models.BoardItem
	err = db.Unscoped().First(&foundItem, item.ID).Error
	if err == nil {
		assert.NotNil(t, foundItem.DeletedAt)
	}

	// Verify item is not found in normal query
	err = db.First(&foundItem, item.ID).Error
	assert.Error(t, err)
}

func TestBoardItemRepository_DeleteByBoard(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardItemRepository(db)

	boardID := uuid.New()
	userID := uuid.New()

	// Create test items for the board
	item1 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   boardID,
		Type:      models.ItemTypeNote,
		X:         100.0,
		Y:         200.0,
		Content:   "Item 1",
		CreatedBy: userID,
	}
	err := db.Create(item1).Error
	assert.NoError(t, err)

	item2 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   boardID,
		Type:      models.ItemTypeNote,
		X:         300.0,
		Y:         400.0,
		Content:   "Item 2",
		CreatedBy: userID,
	}
	err = db.Create(item2).Error
	assert.NoError(t, err)

	// Create item for different board
	item3 := &models.BoardItem{
		ID:        uuid.New(),
		BoardID:   uuid.New(),
		Type:      models.ItemTypeNote,
		X:         500.0,
		Y:         600.0,
		Content:   "Item 3",
		CreatedBy: userID,
	}
	err = db.Create(item3).Error
	assert.NoError(t, err)

	// Delete items for the board
	err = repo.DeleteByBoard(boardID)
	assert.NoError(t, err)

	// Verify only items for the specified board were deleted
	var remainingItems []models.BoardItem
	err = db.Find(&remainingItems).Error
	assert.NoError(t, err)
	assert.Len(t, remainingItems, 1)
	assert.Equal(t, item3.ID, remainingItems[0].ID)
}

func TestBoardConnectionRepository_Create(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	connection := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "red", "thickness": 2}`,
		CreatedBy:  uuid.New(),
	}

	err := repo.Create(connection)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, connection.ID)
	assert.NotZero(t, connection.CreatedAt)
	assert.NotZero(t, connection.UpdatedAt)

	// Verify connection was created
	var foundConnection models.BoardConnection
	err = db.First(&foundConnection, connection.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, connection.BoardID, foundConnection.BoardID)
	assert.Equal(t, connection.FromItemID, foundConnection.FromItemID)
	assert.Equal(t, connection.ToItemID, foundConnection.ToItemID)
	assert.Equal(t, connection.Style, foundConnection.Style)
}

func TestBoardConnectionRepository_GetByID(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	// Create a test connection
	connection := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "blue"}`,
		CreatedBy:  uuid.New(),
	}
	err := db.Create(connection).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		connID   uuid.UUID
		expected *models.BoardConnection
		hasError bool
	}{
		{
			name:     "existing connection",
			connID:   connection.ID,
			expected: connection,
			hasError: false,
		},
		{
			name:     "non-existing connection",
			connID:   uuid.New(),
			expected: nil,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.connID)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.expected.ID, result.ID)
					assert.Equal(t, tt.expected.BoardID, result.BoardID)
				}
			}
		})
	}
}

func TestBoardConnectionRepository_ListByBoard(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	boardID := uuid.New()
	userID := uuid.New()

	// Create test connections
	conn1 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    boardID,
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "red"}`,
		CreatedBy:  userID,
	}
	err := db.Create(conn1).Error
	assert.NoError(t, err)

	conn2 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    boardID,
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "blue"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn2).Error
	assert.NoError(t, err)

	// Create connection for different board
	conn3 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "green"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn3).Error
	assert.NoError(t, err)

	// Test listing connections for board
	connections, err := repo.ListByBoard(boardID)
	assert.NoError(t, err)
	assert.Len(t, connections, 2)

	// Verify only connections for the specified board are returned
	for _, conn := range connections {
		assert.Equal(t, boardID, conn.BoardID)
	}
}

func TestBoardConnectionRepository_Update(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	// Create a test connection
	connection := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "red"}`,
		CreatedBy:  uuid.New(),
	}
	err := db.Create(connection).Error
	assert.NoError(t, err)

	// Update the connection
	connection.Style = `{"color": "blue", "thickness": 3}`

	err = repo.Update(connection)
	assert.NoError(t, err)

	// Verify connection was updated
	var foundConnection models.BoardConnection
	err = db.First(&foundConnection, connection.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, `{"color": "blue", "thickness": 3}`, foundConnection.Style)
}

func TestBoardConnectionRepository_Delete(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	// Create a test connection
	connection := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "red"}`,
		CreatedBy:  uuid.New(),
	}
	err := db.Create(connection).Error
	assert.NoError(t, err)

	// Delete the connection
	err = repo.Delete(connection.ID)
	assert.NoError(t, err)

	// Verify connection was deleted (hard delete)
	var foundConnection models.BoardConnection
	err = db.Unscoped().First(&foundConnection, connection.ID).Error
	if err == nil {
		assert.NotNil(t, foundConnection.DeletedAt)
	}

	// Verify connection is not found in normal query
	err = db.First(&foundConnection, connection.ID).Error
	assert.Error(t, err)
}

func TestBoardConnectionRepository_DeleteByBoard(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	boardID := uuid.New()
	userID := uuid.New()

	// Create test connections for the board
	conn1 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    boardID,
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "red"}`,
		CreatedBy:  userID,
	}
	err := db.Create(conn1).Error
	assert.NoError(t, err)

	conn2 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    boardID,
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "blue"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn2).Error
	assert.NoError(t, err)

	// Create connection for different board
	conn3 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: uuid.New(),
		ToItemID:   uuid.New(),
		Style:      `{"color": "green"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn3).Error
	assert.NoError(t, err)

	// Delete connections for the board
	err = repo.DeleteByBoard(boardID)
	assert.NoError(t, err)

	// Verify only connections for the specified board were deleted
	var remainingConnections []models.BoardConnection
	err = db.Find(&remainingConnections).Error
	assert.NoError(t, err)
	assert.Len(t, remainingConnections, 1)
	assert.Equal(t, conn3.ID, remainingConnections[0].ID)
}

func TestBoardConnectionRepository_DeleteByItem(t *testing.T) {
	db := setupItemTestDB(t)
	repo := NewBoardConnectionRepository(db)

	itemID1 := uuid.New()
	itemID2 := uuid.New()
	itemID3 := uuid.New()
	userID := uuid.New()

	// Create test connections
	conn1 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: itemID1,
		ToItemID:   itemID2,
		Style:      `{"color": "red"}`,
		CreatedBy:  userID,
	}
	err := db.Create(conn1).Error
	assert.NoError(t, err)

	conn2 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: itemID2,
		ToItemID:   itemID1,
		Style:      `{"color": "blue"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn2).Error
	assert.NoError(t, err)

	conn3 := &models.BoardConnection{
		ID:         uuid.New(),
		BoardID:    uuid.New(),
		FromItemID: itemID3,
		ToItemID:   uuid.New(),
		Style:      `{"color": "green"}`,
		CreatedBy:  userID,
	}
	err = db.Create(conn3).Error
	assert.NoError(t, err)

	// Delete connections involving itemID1
	err = repo.DeleteByItem(itemID1)
	assert.NoError(t, err)

	// Verify only connections involving itemID1 were deleted
	var remainingConnections []models.BoardConnection
	err = db.Find(&remainingConnections).Error
	assert.NoError(t, err)
	assert.Len(t, remainingConnections, 1)
	assert.Equal(t, conn3.ID, remainingConnections[0].ID)
}
