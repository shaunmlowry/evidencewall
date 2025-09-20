.PHONY: help dev build test clean docker-build docker-up docker-down

# Default target
help:
	@echo "Available targets:"
	@echo "  dev          - Start development environment"
	@echo "  build        - Build all services"
	@echo "  test         - Run all tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker images"
	@echo "  docker-up    - Start with Docker Compose"
	@echo "  docker-down  - Stop Docker Compose"

# Development
dev:
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development infrastructure started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo ""
	@echo "To start the services manually:"
	@echo "  make run-auth &"
	@echo "  make run-boards &"
	@echo "  cd frontend && npm start"

# Build all services
build:
	@echo "Building auth service..."
	cd services/auth && go build -o ../../bin/auth-service ./cmd/server
	@echo "Building boards service..."
	cd services/boards && go build -o ../../bin/boards-service ./cmd/server
	@echo "Building frontend..."
	cd frontend && npm run build

# Test all services
test:
	@echo "Testing Go services..."
	cd services/auth && go test ./...
	cd services/boards && go test ./...
	@echo "Testing frontend..."
	cd frontend && npm test -- --coverage --watchAll=false

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/build/
	rm -rf services/auth/tmp/
	rm -rf services/boards/tmp/

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Individual service commands
run-auth:
	cd services/auth && air

run-boards:
	cd services/boards && air

# Database commands
migrate-up:
	cd services/auth && go run cmd/migrate/main.go up
	cd services/boards && go run cmd/migrate/main.go up

migrate-down:
	cd services/auth && go run cmd/migrate/main.go down
	cd services/boards && go run cmd/migrate/main.go down

# Setup commands
setup:
	@echo "Setting up development environment..."
	mkdir -p bin
	cd shared && go mod tidy
	cd services/auth && go mod tidy
	cd services/boards && go mod tidy
	cd frontend && npm install
	@echo "Setup complete!"

# Linting
lint:
	@echo "Linting Go code..."
	cd services/auth && golangci-lint run
	cd services/boards && golangci-lint run
	@echo "Linting frontend..."
	cd frontend && npm run lint

