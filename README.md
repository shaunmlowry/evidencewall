# Evidence Wall - Collaborative Investigation Board

A modern, microservices-based collaborative evidence board application built with Go and React. Inspired by detective investigation boards, this application allows teams to collaboratively organize evidence, suspects, and connections in real-time.

## 🎯 Features

### Core Functionality

- **Multiple Boards**: Create and manage multiple investigation boards with unique IDs
- **Real-time Collaboration**: Multiple users can edit boards simultaneously with live updates
- **Evidence Items**: Add post-it notes and suspect cards with drag-and-drop functionality
- **Connections**: Draw string connections between evidence items to show relationships
- **Permissions System**: Granular access control (read, read/write, admin)

### User Management

- **Authentication**: Email/password login with JWT tokens
- **Google OAuth2**: Single sign-on with Google accounts
- **User Profiles**: Customizable user profiles with avatars
- **Board Sharing**: Share boards privately, with specific users, or publicly

### Technical Features

- **Microservices Architecture**: Scalable, maintainable service-oriented design
- **WebSocket Real-time**: Live cursor tracking and instant updates
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Docker Deployment**: Easy deployment with Docker Compose
- **Development Environment**: Complete devcontainer setup with all tools

## 🔒 Security & Network Architecture

The production deployment uses a secure, multi-layered network architecture:

- **Network Isolation**: Database and cache services are completely isolated from external access
- **Reverse Proxy**: NGINX proxy handles all external traffic with rate limiting and security headers
- **No Direct Access**: Backend services are only accessible through the proxy
- **SSL/TLS Ready**: HTTPS configuration with security headers and modern cipher suites

For detailed information, see [NETWORK_ARCHITECTURE.md](./NETWORK_ARCHITECTURE.md).

### Production Security Features

- PostgreSQL and Redis: No public port exposure
- Backend APIs: Internal network access only
- Rate limiting: 5-50 requests/second per endpoint
- Security headers: X-Frame-Options, CSP, HSTS
- CORS: Properly configured for production domains

## 🏗️ Architecture

This application follows a microservices architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │    │ Boards Service  │    │Realtime Service │
│     (Go)        │    │     (Go)        │    │     (Go)        │
│   Port: 8001    │    │   Port: 8002    │    │   Port: 8003    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────────┐
         │                  Frontend (React)                      │
         │                   Port: 3000                           │
         └─────────────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────────────┐
         │              Infrastructure Layer                      │
         │  ┌─────────────────┐    ┌─────────────────┐           │
         │  │   PostgreSQL    │    │     Redis       │           │
         │  │   Port: 5432    │    │   Port: 6379    │           │
         │  └─────────────────┘    └─────────────────┘           │
         └─────────────────────────────────────────────────────────┘
```

### Services Overview

- **Auth Service**: User authentication, JWT tokens, OAuth2 integration
- **Boards Service**: Board CRUD operations, permissions management, item management
- **Real-time Service**: WebSocket connections for live collaboration
- **Frontend**: React application with Material-UI components
- **PostgreSQL**: Primary database for persistent data
- **Redis**: Caching and real-time message broker

## 🚀 Quick Start

### Option 1: Development with DevContainer (Recommended)

1. **Prerequisites**: VS Code with Dev Containers extension
2. **Clone and Open**:
   ```bash
   git clone <repository-url>
   cd evidence-wall
   code .
   ```
3. **Reopen in Container**: When prompted, click "Reopen in Container"
4. **Start Development Environment**:

   ```bash
   # Start infrastructure (PostgreSQL, Redis)
   make dev

   # In separate terminals, start the services:
   make run-auth     # Terminal 1
   make run-boards   # Terminal 2
   cd frontend && npm start  # Terminal 3
   ```

5. **Access the Application**:
   - Frontend: http://localhost:3000
   - Auth API: http://localhost:8001
   - Boards API: http://localhost:8002

### Option 2: Local Development

1. **Prerequisites**:

   - Go 1.21+
   - Node.js 18+
   - Docker and Docker Compose
   - PostgreSQL 15+
   - Redis 7+

2. **Setup Environment**:

   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Start Infrastructure**:

   ```bash
   docker-compose up -d postgres redis
   ```

4. **Start Backend Services**:

   ```bash
   # In separate terminals:
   make run-auth     # Terminal 1
   make run-boards   # Terminal 2
   ```

5. **Start Frontend**:
   ```bash
   cd frontend && npm install && npm start  # Terminal 3
   ```

### Option 3: Production Deployment

1. **Configure Environment**:

   ```bash
   cp env.example .env
   # Set production values, especially:
   # - JWT_SECRET (use a strong random key)
   # - GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET
   # - Database passwords
   ```

2. **Deploy with Docker Compose**:

   ```bash
   docker-compose up -d
   ```

3. **Access the Application**:
   - Application: http://localhost:3000
   - Health checks available at `/health` endpoints

## 📖 API Documentation

### Authentication Service (Port 8001)

- **Swagger UI**: http://localhost:8001/swagger/
- **Base URL**: http://localhost:8001/api/v1

#### Key Endpoints:

- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `GET /auth/google` - Google OAuth login
- `GET /me` - Get user profile (requires auth)
- `PUT /me` - Update user profile (requires auth)

### Boards Service (Port 8002)

- **Swagger UI**: http://localhost:8002/swagger/
- **Base URL**: http://localhost:8002/api/v1

#### Key Endpoints:

- `GET /boards` - List user's boards
- `POST /boards` - Create new board
- `GET /boards/:id` - Get board details
- `PUT /boards/:id` - Update board
- `DELETE /boards/:id` - Delete board
- `POST /boards/:id/share` - Share board with user
- `GET /boards/:boardId/items` - Get board items
- `POST /boards/:boardId/items` - Create board item
- `GET /public/boards/:id` - Get public board (no auth required)

### Real-time Service (Port 8003)

- **WebSocket URL**: ws://localhost:8003/ws
- **Authentication**: JWT token in connection auth

#### WebSocket Events:

- `join_board` - Join a board room
- `leave_board` - Leave a board room
- `item_update` - Real-time item updates
- `connection_update` - Real-time connection updates
- `user_cursor` - Live cursor tracking

## 🛠️ Development

### Project Structure

```
evidence-wall/
├── services/
│   ├── auth/                   # Authentication service
│   │   ├── cmd/server/         # Main application
│   │   ├── internal/           # Internal packages
│   │   │   ├── config/         # Configuration
│   │   │   ├── handlers/       # HTTP handlers
│   │   │   ├── repository/     # Data access layer
│   │   │   └── service/        # Business logic
│   │   ├── Dockerfile          # Production container
│   │   ├── Dockerfile.dev      # Development container
│   │   └── .air.toml          # Hot reload config
│   ├── boards/                 # Boards service (similar structure)
│   └── realtime/              # Real-time service (similar structure)
├── frontend/                   # React application
│   ├── src/
│   │   ├── components/         # Reusable components
│   │   ├── contexts/          # React contexts
│   │   ├── pages/             # Page components
│   │   ├── services/          # API services
│   │   └── types/             # TypeScript types
│   ├── public/                # Static assets
│   ├── Dockerfile             # Production container
│   └── nginx.conf             # Nginx configuration
├── shared/                     # Shared Go packages
│   ├── auth/                  # JWT utilities
│   ├── database/              # Database utilities
│   ├── middleware/            # HTTP middleware
│   └── models/                # Data models
├── migrations/                 # Database migrations
├── .devcontainer/             # VS Code dev container
├── docker-compose.yml         # Production deployment
├── docker-compose.dev.yml     # Development environment
├── Makefile                   # Build and run commands
└── env.example               # Environment variables template
```

### Available Make Commands

```bash
# Development
make dev          # Start development environment
make setup        # Setup development dependencies

# Building
make build        # Build all services
make clean        # Clean build artifacts

# Testing
make test         # Run all tests
make lint         # Run linters

# Database
make migrate-up   # Run database migrations
make migrate-down # Rollback database migrations

# Docker
make docker-build # Build Docker images
make docker-up    # Start with Docker Compose
make docker-down  # Stop Docker Compose

# Individual services
make run-auth     # Run auth service with hot reload
make run-boards   # Run boards service with hot reload
make run-realtime # Run real-time service with hot reload
```

### Environment Variables

Key environment variables (see `env.example` for complete list):

```bash
# Database
DATABASE_URL=postgres://user:password@localhost:5432/evidencewall?sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY=24h

# Google OAuth2 (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Service Ports
AUTH_SERVICE_PORT=8001
BOARDS_SERVICE_PORT=8002
REALTIME_SERVICE_PORT=8003

# Frontend URLs
REACT_APP_API_BASE_URL=http://localhost:8001
REACT_APP_BOARDS_API_URL=http://localhost:8002
REACT_APP_WEBSOCKET_URL=ws://localhost:8003
```

## 🧪 Testing

### Backend Testing

```bash
# Run all Go tests
make test

# Run tests for specific service
cd services/auth && go test ./...
cd services/boards && go test ./...
cd services/realtime && go test ./...

# Run tests with coverage
cd services/auth && go test -cover ./...
```

### Frontend Testing

```bash
cd frontend

# Run all tests
npm test

# Run tests with coverage
npm test -- --coverage --watchAll=false

# Run specific test file
npm test -- LoginPage.test.tsx
```

## 🔧 Technology Stack

### Backend Technologies

- **Go 1.21+**: Modern, performant backend language
- **Gin**: Fast HTTP web framework
- **GORM**: Feature-rich ORM for Go
- **JWT**: Stateless authentication
- **WebSockets**: Real-time bidirectional communication
- **PostgreSQL 15**: Robust relational database
- **Redis 7**: In-memory data store for caching and pub/sub

### Frontend Technologies

- **React 18**: Modern UI library with hooks
- **TypeScript**: Type-safe JavaScript
- **Material-UI v5**: Google's Material Design components
- **React Router v6**: Client-side routing
- **TanStack Query**: Server state management
- **Socket.io Client**: WebSocket client library
- **Axios**: HTTP client for API calls

### DevOps & Tools

- **Docker**: Containerization platform
- **Docker Compose**: Multi-container orchestration
- **Air**: Live reload for Go applications
- **Nginx**: Web server and reverse proxy
- **VS Code DevContainers**: Consistent development environment

### Testing Frameworks

- **Testify**: Go testing toolkit
- **Jest**: JavaScript testing framework
- **React Testing Library**: React component testing
- **Supertest**: HTTP assertion library

## 🔒 Security Features

- **JWT Authentication**: Secure, stateless authentication
- **Password Hashing**: bcrypt for secure password storage
- **CORS Protection**: Configurable cross-origin resource sharing
- **Input Validation**: Comprehensive request validation
- **SQL Injection Prevention**: Parameterized queries with GORM
- **XSS Protection**: Content Security Policy headers
- **Rate Limiting**: Built-in request rate limiting (can be added)

## 📊 Database Schema

### Core Tables

- **users**: User accounts and profiles
- **boards**: Investigation boards
- **board_users**: User permissions for boards
- **board_items**: Post-it notes and suspect cards
- **board_connections**: String connections between items

### Key Relationships

- Users can own multiple boards
- Users can have different permissions on different boards
- Boards contain multiple items and connections
- Items can be connected to other items within the same board

## 🚀 Deployment

### Production Checklist

- [ ] Set strong JWT_SECRET
- [ ] Configure Google OAuth2 credentials
- [ ] Set up SSL/TLS certificates
- [ ] Configure database backups
- [ ] Set up monitoring and logging
- [ ] Configure firewall rules
- [ ] Set up domain and DNS
- [ ] Enable database connection pooling
- [ ] Configure Redis persistence
- [ ] Set up health checks

### Scaling Considerations

- **Horizontal Scaling**: Each service can be scaled independently
- **Database**: Consider read replicas for high read loads
- **Redis**: Use Redis Cluster for high availability
- **Load Balancing**: Use nginx or cloud load balancers
- **CDN**: Serve static assets from CDN
- **Monitoring**: Implement comprehensive monitoring and alerting

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the Repository**
2. **Create a Feature Branch**: `git checkout -b feature/amazing-feature`
3. **Make Your Changes**: Follow the coding standards
4. **Add Tests**: Ensure your changes are tested
5. **Commit Changes**: `git commit -m 'Add amazing feature'`
6. **Push to Branch**: `git push origin feature/amazing-feature`
7. **Open Pull Request**: Describe your changes clearly

### Coding Standards

- **Go**: Follow `gofmt` and `golint` standards
- **TypeScript**: Use ESLint and Prettier configurations
- **Commits**: Use conventional commit messages
- **Tests**: Maintain >80% test coverage
- **Documentation**: Update relevant documentation

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Documentation**: Check the `/docs` folder for detailed guides
- **Community**: Join our discussions in GitHub Discussions

## 🙏 Acknowledgments

- Inspired by detective investigation boards and evidence walls
- Built with modern web technologies and best practices
- Thanks to the open-source community for the amazing tools and libraries
