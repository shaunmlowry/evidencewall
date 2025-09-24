module evidence-wall/realtime-service

go 1.21

require (
	evidence-wall/shared v0.0.0
	github.com/googollee/go-socket.io v1.7.0
	github.com/redis/go-redis/v9 v9.1.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
)

replace evidence-wall/shared => ../../shared
