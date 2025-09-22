package handlers

import (
	"net/http"
	"strconv"

	"evidence-wall/boards-service/internal/service"
	"evidence-wall/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BoardHandler handles board HTTP requests
type BoardHandler struct {
	boardService *service.BoardService
}

// NewBoardHandler creates a new board handler
func NewBoardHandler(boardService *service.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
	}
}

// CreateBoard godoc
// @Summary Create a new board
// @Description Create a new board
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateBoardRequest true "Board creation request"
// @Success 201 {object} models.Board
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards [post]
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req service.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	board, err := h.boardService.CreateBoard(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create board"})
		return
	}

	c.JSON(http.StatusCreated, board)
}

// GetBoard godoc
// @Summary Get a board by ID
// @Description Get a board by ID with permission check
// @Tags boards
// @Produce json
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Success 200 {object} models.Board
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id} [get]
func (h *BoardHandler) GetBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	board, err := h.boardService.GetBoard(boardID, userID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get board"})
		}
		return
	}

	c.JSON(http.StatusOK, board)
}

// GetPublicBoard godoc
// @Summary Get a public board by ID
// @Description Get a public board by ID (no authentication required)
// @Tags boards
// @Produce json
// @Param id path string true "Board ID"
// @Success 200 {object} models.Board
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /public/boards/{id} [get]
func (h *BoardHandler) GetPublicBoard(c *gin.Context) {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	board, err := h.boardService.GetPublicBoard(boardID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"}) // Don't reveal private boards
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get board"})
		}
		return
	}

	c.JSON(http.StatusOK, board)
}

// ListBoards godoc
// @Summary List user's boards
// @Description List boards accessible by the current user
// @Tags boards
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards [get]
func (h *BoardHandler) ListBoards(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	boards, total, err := h.boardService.ListBoards(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list boards"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"boards": boards,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// UpdateBoard godoc
// @Summary Update a board
// @Description Update a board (admin permission required)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Param request body service.UpdateBoardRequest true "Board update request"
// @Success 200 {object} models.Board
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id} [put]
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var req service.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	board, err := h.boardService.UpdateBoard(boardID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update board"})
		}
		return
	}

	c.JSON(http.StatusOK, board)
}

// DeleteBoard godoc
// @Summary Delete a board
// @Description Delete a board (admin permission required)
// @Tags boards
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id} [delete]
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	err = h.boardService.DeleteBoard(boardID, userID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete board"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ShareBoard godoc
// @Summary Share a board with a user
// @Description Share a board with a user (admin permission required)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Param request body service.ShareBoardRequest true "Share board request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id}/share [post]
func (h *BoardHandler) ShareBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var req service.ShareBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.boardService.ShareBoard(boardID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share board"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board shared successfully"})
}

// UnshareBoard godoc
// @Summary Remove user access to a board
// @Description Remove user access to a board (admin permission required)
// @Tags boards
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id}/share/{userId} [delete]
func (h *BoardHandler) UnshareBoard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.boardService.UnshareBoard(boardID, userID, targetUserID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unshare board"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Board unshared successfully"})
}

// UpdateUserPermission godoc
// @Summary Update user permission for a board
// @Description Update user permission for a board (admin permission required)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Board ID"
// @Param userId path string true "User ID"
// @Param request body service.UpdateUserPermissionRequest true "Permission update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{id}/users/{userId}/permission [put]
func (h *BoardHandler) UpdateUserPermission(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req service.UpdateUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.boardService.UpdateUserPermission(boardID, userID, targetUserID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission updated successfully"})
}

// CreateBoardItem godoc
// @Summary Create a board item
// @Description Create a new item on a board
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param request body service.CreateItemRequest true "Item creation request"
// @Success 201 {object} models.BoardItem
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{boardId}/items [post]
func (h *BoardHandler) CreateBoardItem(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var req service.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.boardService.CreateBoardItem(boardID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		}
		return
	}

	c.JSON(http.StatusCreated, item)
}

// ListBoardItems godoc
// @Summary List board items
// @Description List all items on a board
// @Tags items
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 200 {array} models.BoardItem
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{boardId}/items [get]
func (h *BoardHandler) ListBoardItems(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	items, err := h.boardService.ListBoardItems(boardID, userID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list items"})
		}
		return
	}

	c.JSON(http.StatusOK, items)
}

// UpdateBoardItem godoc
// @Summary Update a board item
// @Description Update a board item
// @Tags items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param id path string true "Item ID"
// @Param request body service.UpdateItemRequest true "Item update request"
// @Success 200 {object} models.BoardItem
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{boardId}/items/{id} [put]
func (h *BoardHandler) UpdateBoardItem(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req service.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.boardService.UpdateBoardItem(boardID, itemID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrItemNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		}
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteBoardItem godoc
// @Summary Delete a board item
// @Description Delete a board item
// @Tags items
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param id path string true "Item ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /boards/{boardId}/items/{id} [delete]
func (h *BoardHandler) DeleteBoardItem(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

    // Router uses :id for board identifier (see routes setup)
    boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	err = h.boardService.DeleteBoardItem(boardID, itemID, userID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrItemNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Placeholder handlers for board connections (not implemented in service yet)
func (h *BoardHandler) ListBoardConnections(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	connections, err := h.boardService.ListBoardConnections(boardID, userID)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list connections"})
		}
		return
	}

	c.JSON(http.StatusOK, connections)
}

func (h *BoardHandler) CreateBoardConnection(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var req service.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := h.boardService.CreateBoardConnection(boardID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		case service.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection input"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
		}
		return
	}

	c.JSON(http.StatusCreated, conn)
}

func (h *BoardHandler) UpdateBoardConnection(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}
	connectionID, err := uuid.Parse(c.Param("connectionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	var req service.UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := h.boardService.UpdateBoardConnection(boardID, connectionID, userID, req)
	if err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrConnectionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection"})
		}
		return
	}

	c.JSON(http.StatusOK, conn)
}

func (h *BoardHandler) DeleteBoardConnection(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}
	connectionID, err := uuid.Parse(c.Param("connectionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection ID"})
		return
	}

	if err := h.boardService.DeleteBoardConnection(boardID, connectionID, userID); err != nil {
		switch err {
		case service.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		case service.ErrConnectionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		case service.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
