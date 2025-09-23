# Security Configuration Guide

## Environment Variables

This application requires several environment variables to be set for secure operation. Copy `env.example` to `.env` and configure the following:

### Required Security Variables

```bash
# Generate a strong JWT secret (minimum 32 characters)
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters-long

# Use a strong database password
POSTGRES_PASSWORD=your-secure-database-password-here

# Configure Redis password for production
REDIS_PASSWORD=your-redis-password-here
```

### Generating Secure Secrets

#### JWT Secret
```bash
# Generate a cryptographically secure JWT secret
openssl rand -base64 32
```

#### Database Password
```bash
# Generate a strong database password
openssl rand -base64 24
```

## Production Security Checklist

- [ ] Change all default passwords
- [ ] Use strong, unique JWT secrets
- [ ] Enable HTTPS with valid certificates
- [ ] Configure proper CORS origins
- [ ] Set up Redis authentication
- [ ] Enable database SSL connections
- [ ] Configure security headers
- [ ] Set up proper logging and monitoring
- [ ] Regular security updates
- [ ] Backup and recovery procedures

## Security Headers

The application includes the following security headers via NGINX:

- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

## Rate Limiting

API endpoints are protected with rate limiting:

- Authentication endpoints: 5 requests/second
- General API endpoints: 10 requests/second

## Network Security

- Database and Redis are isolated in internal networks
- Backend services are not directly accessible from external networks
- All external traffic goes through NGINX proxy
- SSL/TLS termination at proxy level

## Reporting Security Issues

If you discover a security vulnerability, please report it to the development team immediately. Do not create public issues for security vulnerabilities.
