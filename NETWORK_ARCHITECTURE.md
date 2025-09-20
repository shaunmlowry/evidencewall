# Network Architecture

This document describes the network isolation and security architecture for the Evidence Wall application in production.

## Network Topology

The application uses Docker networks to isolate services and minimize the attack surface:

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      │ Port 80/443 (HTTP/HTTPS)
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  NGINX Proxy                                │
│              (nginx:alpine)                                 │
│  - Rate limiting                                            │
│  - SSL termination                                          │
│  - Security headers                                         │
│  - API routing                                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        │ frontend    │ backend     │
        │ network     │ network     │
        │             │             │
┌───────▼──────┐     ┌▼─────────────▼──────────────────────────┐
│   Frontend   │     │           Backend Services              │
│  (React App) │     │                                         │
│              │     │  ┌─────────────┐ ┌─────────────────────┐│
│              │     │  │Auth Service │ │ Boards Service      ││
│              │     │  │   :8001     │ │    :8002            ││
│              │     │  └─────────────┘ └─────────────────────┘│
│              │     │                                         │
│              │     │  ┌─────────────────────────────────────┐│
│              │     │  │      Realtime Service              ││
│              │     │  │          :8003                     ││
│              │     │  └─────────────────────────────────────┘│
└──────────────┘     └─────────────┬───────────────────────────┘
                                   │
                                   │ backend network
                                   │
                     ┌─────────────▼───────────────┐
                     │        internal network     │
                     │         (isolated)          │
                     │                             │
                     │  ┌─────────────────────────┐│
                     │  │     PostgreSQL          ││
                     │  │       :5432             ││
                     │  └─────────────────────────┘│
                     │                             │
                     │  ┌─────────────────────────┐│
                     │  │       Redis             ││
                     │  │       :6379             ││
                     │  └─────────────────────────┘│
                     └─────────────────────────────┘
```

## Network Isolation

### 1. Internal Network (Completely Isolated)

- **Purpose**: Database and cache services
- **Services**: PostgreSQL, Redis
- **Access**: Only accessible by backend services
- **Security**: `internal: true` prevents external access
- **Ports**: No public port exposure

### 2. Backend Network

- **Purpose**: API services and database access
- **Services**: Auth, Boards, Realtime services + Database services
- **Access**: Only accessible via NGINX proxy
- **Security**: No direct public access
- **Ports**: No public port exposure

### 3. Frontend Network

- **Purpose**: Public-facing services
- **Services**: Frontend React app, NGINX proxy
- **Access**: NGINX proxy routes traffic to appropriate services
- **Security**: Only NGINX proxy has public ports (80/443)

## Security Features

### NGINX Proxy Security

- **Rate Limiting**: API endpoints have rate limits (5-50 req/s)
- **Security Headers**: X-Frame-Options, X-Content-Type-Options, etc.
- **CORS**: Properly configured for production domains
- **SSL Ready**: HTTPS configuration ready for certificates
- **Request Filtering**: Only allows specific API paths

### Service Isolation

- **Database**: No direct external access
- **Backend APIs**: Only accessible through proxy
- **Frontend**: Served through proxy with proper headers
- **Health Checks**: Internal health monitoring

### Environment Configuration

- **Production Secrets**: JWT secrets, database passwords
- **CORS Origins**: Restricted to production domains
- **OAuth Callbacks**: Updated for proxy routing
- **API URLs**: Relative paths through proxy

## Port Exposure

### Public Ports (Exposed to Internet)

- **Port 80**: HTTP traffic (redirects to HTTPS in production)
- **Port 443**: HTTPS traffic (when SSL is configured)

### Internal Ports (Docker Network Only)

- **5432**: PostgreSQL (internal + backend networks)
- **6379**: Redis (internal + backend networks)
- **8001**: Auth Service (backend network)
- **8002**: Boards Service (backend network)
- **8003**: Realtime Service (backend network)
- **80**: Frontend container (frontend network)

## API Routing

All external API access goes through NGINX proxy:

- `GET /` → Frontend React app
- `POST /api/auth/*` → Auth Service (rate limited)
- `GET /api/boards/*` → Boards Service
- `WebSocket /api/realtime/*` → Realtime Service
- `GET /health` → NGINX health check

## SSL/TLS Configuration

### Development

- HTTP only on port 80
- Self-signed certificates can be generated for testing

### Production

- HTTPS on port 443 with valid certificates
- HTTP redirects to HTTPS
- TLS 1.2+ with secure cipher suites
- HSTS headers for security

## Deployment

### Starting the Application

```bash
# Copy environment variables
cp env.example .env
# Edit .env with production values

# Start all services
docker-compose up -d

# Check service health
docker-compose ps
curl http://localhost/health
```

### SSL Certificate Setup

```bash
# Generate self-signed certificates (development)
openssl req -x509 -newkey rsa:4096 -keyout nginx/ssl/key.pem -out nginx/ssl/cert.pem -days 365 -nodes

# Or use Let's Encrypt (production)
# Place certificates in nginx/ssl/ directory
# Uncomment HTTPS server block in nginx.conf
```

### Monitoring

```bash
# View logs
docker-compose logs nginx
docker-compose logs auth-service

# Check network connectivity
docker network ls
docker network inspect evidence-wall_backend
```

## Security Considerations

1. **Database Security**: PostgreSQL and Redis are completely isolated from external access
2. **API Security**: All API calls go through NGINX with rate limiting and security headers
3. **Secret Management**: Use Docker secrets or external secret management in production
4. **SSL/TLS**: Always use HTTPS in production with valid certificates
5. **CORS**: Configure CORS origins for your production domain
6. **Firewall**: Only expose ports 80/443 on the host system
7. **Updates**: Regularly update base images and dependencies
8. **Monitoring**: Implement proper logging and monitoring for security events

## Troubleshooting

### Common Issues

1. **Service can't connect to database**: Check if service is on backend network
2. **CORS errors**: Verify CORS_ORIGINS environment variable
3. **SSL errors**: Check certificate paths and permissions
4. **Rate limiting**: Adjust NGINX rate limits if needed

### Network Debugging

```bash
# Check network connectivity between services
docker-compose exec auth-service ping postgres
docker-compose exec nginx ping auth-service

# Inspect network configuration
docker network inspect evidence-wall_internal
```
