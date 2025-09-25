package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"evidence-wall/shared/auth"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development - in production, be more restrictive
		return true
	},
}

// Client represents a WebSocket connection
type Client struct {
	conn      *websocket.Conn
	userID    string
	userEmail string
	boards    map[string]bool // boards this client has joined
	send      chan []byte     // buffered channel for outbound messages
	mutex     sync.RWMutex
}

// Hub manages all WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	boardRooms map[string]map[*Client]bool // boardID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mutex      sync.RWMutex
}

// Message types
type Message struct {
	Type    string      `json:"type"`
	BoardID string      `json:"board_id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Environment variables
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-key")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")

	// JWT manager for authentication
	jwtManager := auth.NewJWTManager(jwtSecret, time.Hour)

	// Redis client
	rdb := redis.NewClient(&redis.Options{Addr: redisHost + ":" + redisPort})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	// Create hub
	hub := &Hub{
		clients:    make(map[*Client]bool),
		boardRooms: make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}

	// Start hub
	go hub.run()

	// Start Redis subscriber
	go subscribeToRedis(rdb, hub)

	// HTTP handlers
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, jwtManager, w, r)
	})

	log.Printf("WebSocket server listening on :8003")
	if err := http.ListenAndServe(":8003", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)

				// Remove from all board rooms
				for boardID := range client.boards {
					if room, exists := h.boardRooms[boardID]; exists {
						delete(room, client)
						if len(room) == 0 {
							delete(h.boardRooms, boardID)
						}
					}
				}

				close(client.send)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling broadcast message: %v", err)
				continue
			}

			h.mutex.RLock()
			if room, exists := h.boardRooms[msg.BoardID]; exists {
				for client := range room {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(room, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func handleWebSocket(hub *Hub, jwtManager *auth.JWTManager, w http.ResponseWriter, r *http.Request) {
	// Get JWT token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("WebSocket connection rejected: no token provided")
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		log.Printf("WebSocket connection rejected: invalid token: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := &Client{
		conn:      conn,
		userID:    claims.UserID.String(),
		userEmail: claims.Email,
		boards:    make(map[string]bool),
		send:      make(chan []byte, 256),
	}

	// Register client
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(hub)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch msg.Type {
		case "join_board":
			c.joinBoard(hub, msg.BoardID)
		case "leave_board":
			c.leaveBoard(hub, msg.BoardID)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) joinBoard(hub *Hub, boardID string) {
	if boardID == "" {
		return
	}

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	// Add to client's boards
	c.mutex.Lock()
	c.boards[boardID] = true
	c.mutex.Unlock()

	// Add to hub's board room
	if hub.boardRooms[boardID] == nil {
		hub.boardRooms[boardID] = make(map[*Client]bool)
	}
	hub.boardRooms[boardID][c] = true

}

func (c *Client) leaveBoard(hub *Hub, boardID string) {
	if boardID == "" {
		return
	}

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	// Remove from client's boards
	c.mutex.Lock()
	delete(c.boards, boardID)
	c.mutex.Unlock()

	// Remove from hub's board room
	if room, exists := hub.boardRooms[boardID]; exists {
		delete(room, c)
		if len(room) == 0 {
			delete(hub.boardRooms, boardID)
		}
	}

}

func subscribeToRedis(rdb *redis.Client, hub *Hub) {
	ctx := context.Background()
	pubsub := rdb.PSubscribe(ctx, "board:*")
	defer pubsub.Close()

	log.Printf("Subscribed to Redis pattern: board:*")

	for msg := range pubsub.Channel() {

		// Extract board ID from channel name (board:uuid)
		parts := strings.Split(msg.Channel, ":")
		if len(parts) != 2 {
			log.Printf("Invalid channel format: %s", msg.Channel)
			continue
		}
		boardID := parts[1]

		// Create WebSocket message
		wsMessage := Message{
			Type:    "board_update",
			BoardID: boardID,
			Data:    json.RawMessage(msg.Payload),
		}

		messageBytes, err := json.Marshal(wsMessage)
		if err != nil {
			log.Printf("Error marshaling WebSocket message: %v", err)
			continue
		}

		// Broadcast to clients in this board
		hub.mutex.RLock()
		if room, exists := hub.boardRooms[boardID]; exists {
			for client := range room {
				select {
				case client.send <- messageBytes:
				default:
					close(client.send)
					delete(hub.clients, client)
					delete(room, client)
				}
			}
		} else {
		}
		hub.mutex.RUnlock()
	}
}
