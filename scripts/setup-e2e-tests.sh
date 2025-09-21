#!/bin/bash

# Evidence Wall - E2E Test Setup Script
# This script sets up the environment for running end-to-end tests

set -e

echo "🎭 Setting up Evidence Wall E2E Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    print_error "Node.js version 18+ is required. Current version: $(node -v)"
    exit 1
fi

print_success "Node.js version check passed: $(node -v)"

# Check if Docker is installed and running
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! docker info &> /dev/null; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_success "Docker check passed"

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

print_success "Docker Compose check passed"

# Install root dependencies (Playwright)
print_status "Installing Playwright dependencies..."
npm install

# Install Playwright browsers
print_status "Installing Playwright browsers..."
npx playwright install

# Install frontend dependencies
print_status "Installing frontend dependencies..."
cd frontend && npm install && cd ..

# Install Go dependencies
print_status "Installing Go dependencies..."
cd shared && go mod tidy && cd ..
cd services/auth && go mod tidy && cd ../..
cd services/boards && go mod tidy && cd ../..

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p bin
mkdir -p test-results
mkdir -p playwright-report

# Set up environment file if it doesn't exist
if [ ! -f .env ]; then
    print_status "Creating .env file from template..."
    cp env.example .env
    print_warning "Please review and update the .env file with appropriate values"
fi

# Check if development infrastructure is running
print_status "Checking development infrastructure..."
if ! docker-compose -f docker-compose.dev.yml ps | grep -q "Up"; then
    print_status "Starting development infrastructure..."
    docker-compose -f docker-compose.dev.yml up -d
    
    print_status "Waiting for database to be ready..."
    sleep 10
    
    # Wait for PostgreSQL to be ready
    until docker-compose -f docker-compose.dev.yml exec -T postgres pg_isready -U evidencewall; do
        print_status "Waiting for PostgreSQL..."
        sleep 2
    done
    
    print_success "Development infrastructure is ready"
else
    print_success "Development infrastructure is already running"
fi

# Build Go services
print_status "Building Go services..."
make build

# Run database migrations (if migration scripts exist)
if [ -f "services/auth/cmd/migrate/main.go" ]; then
    print_status "Running database migrations..."
    make migrate-up || print_warning "Migration failed or already applied"
fi

# Verify test setup
print_status "Verifying test setup..."

# Check if Playwright can run
if npx playwright --version &> /dev/null; then
    print_success "Playwright is ready: $(npx playwright --version)"
else
    print_error "Playwright setup failed"
    exit 1
fi

# Check if Go services can be built
if [ -f "bin/auth-service" ] && [ -f "bin/boards-service" ]; then
    print_success "Go services built successfully"
else
    print_error "Go services build failed"
    exit 1
fi

# Create a simple test script to verify everything works
cat > test-setup-verification.js << 'EOF'
const { test, expect } = require('@playwright/test');

test('setup verification', async ({ page }) => {
  // This is a simple test to verify the setup works
  await page.goto('https://example.com');
  await expect(page).toHaveTitle(/Example/);
});
EOF

print_status "Running setup verification test..."
if npx playwright test test-setup-verification.js --reporter=line; then
    print_success "Setup verification passed"
    rm test-setup-verification.js
else
    print_error "Setup verification failed"
    rm test-setup-verification.js
    exit 1
fi

# Print summary
echo ""
echo "🎉 E2E Test Setup Complete!"
echo ""
echo "Next steps:"
echo "1. Review and update .env file if needed"
echo "2. Start the development services:"
echo "   make run-auth &"
echo "   make run-boards &"
echo "   cd frontend && npm start &"
echo ""
echo "3. Run the tests:"
echo "   make test-e2e                 # Run all tests"
echo "   make test-e2e-ui              # Run with UI"
echo "   make test-e2e-debug           # Run in debug mode"
echo "   npm run test:e2e:headed       # Run with visible browser"
echo ""
echo "4. View test results:"
echo "   npm run test:e2e:report       # View HTML report"
echo ""
echo "For more information, see e2e/README.md"
echo ""
print_success "Happy testing! 🧪"
