package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	sharedauth "evidence-wall/shared/auth"

	socketio "github.com/googollee/go-socket.io"
	"github.com/redis/go-redis/v9"
)

type clientInfo struct {
	userEmail string
	userID    string
}

func main() {
	// Env
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-key")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")

	// JWT validator
	jwtManager := sharedauth.NewJWTManager(jwtSecret, time.Hour)

	// Redis client
	rdb := redis.NewClient(&redis.Options{Addr: redisHost + ":" + redisPort})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	// Socket.IO server
	server := socketio.NewServer(nil)

	// Per-socket joined boards
	socketBoards := make(map[string]map[string]struct{})

	server.OnConnect("/", func(s socketio.Conn) error {
		// Authenticate using token passed in auth
		url := s.URL()
		query := url.Query()
		token, ok := query["token"]
		if !ok || len(token) == 0 {
			s.Close()
			return nil
		}
		claims, err := jwtManager.ValidateToken(token[0])
		if err != nil {
			s.Close()
			return nil
		}
		s.SetContext(&clientInfo{userEmail: claims.Email, userID: claims.UserID.String()})
		socketBoards[s.ID()] = make(map[string]struct{})
		return nil
	})

	server.OnEvent("/", "join_board", func(s socketio.Conn, payload map[string]string) {
		boardID := strings.TrimSpace(payload["board_id"])
		if boardID == "" {
			return
		}
		// Join room and start Redis subscription for this board
		s.Join(boardID)
		if _, ok := socketBoards[s.ID()][boardID]; ok {
			return
		}
		socketBoards[s.ID()][boardID] = struct{}{}

		go func(room string, sid string) {
			sub := rdb.Subscribe(context.Background(), "board:"+room)
			ch := sub.Channel()
			for msg := range ch {
				// Forward raw JSON payload to room
				server.BroadcastToRoom("/", room, "board_update", msg.Payload)
			}
		}(boardID, s.ID())
	})

	server.OnEvent("/", "leave_board", func(s socketio.Conn, payload map[string]string) {
		boardID := strings.TrimSpace(payload["board_id"])
		if boardID == "" {
			return
		}
		s.Leave(boardID)
		delete(socketBoards[s.ID()], boardID)
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Printf("socket error: %v", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		delete(socketBoards, s.ID())
	})

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", server)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8003"
	log.Printf("realtime service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
