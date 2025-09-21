# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Build-time environment variables for Vite
ARG VITE_API_BASE_URL
ARG VITE_BOARDS_API_URL
ARG VITE_WEBSOCKET_URL

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY frontend/ ./

# Build the application (Vite reads VITE_* from process.env at build time)
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_BOARDS_API_URL=${VITE_BOARDS_API_URL}
ENV VITE_WEBSOCKET_URL=${VITE_WEBSOCKET_URL}
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built app to nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]


