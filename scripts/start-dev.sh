#!/bin/bash

# Evidence Wall Development Startup Script

echo "🔧 Starting Evidence Wall Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start infrastructure services
echo "🚀 Starting PostgreSQL and Redis..."
docker-compose -f docker-compose.dev.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are healthy
echo "🔍 Checking service health..."
if docker-compose -f docker-compose.dev.yml ps | grep -q "unhealthy"; then
    echo "⚠️  Some services are not healthy. Check docker-compose logs."
    docker-compose -f docker-compose.dev.yml logs
    exit 1
fi

echo "✅ Infrastructure services are ready!"
echo ""
echo "📋 Next steps:"
echo "1. Start the auth service:    make run-auth"
echo "2. Start the boards service:  make run-boards"
echo "3. Start the frontend:        cd frontend && npm start"
echo ""
echo "🌐 Access points:"
echo "- Frontend:     http://localhost:3000"
echo "- Auth API:     http://localhost:8001"
echo "- Boards API:   http://localhost:8002"
echo "- PostgreSQL:   localhost:5432"
echo "- Redis:        localhost:6379"
echo ""
echo "🛑 To stop infrastructure: docker-compose -f docker-compose.dev.yml down"

