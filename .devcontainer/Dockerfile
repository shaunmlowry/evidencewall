FROM mcr.microsoft.com/devcontainers/go:1.21

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs

# Install additional tools
RUN apt-get update && apt-get install -y \
    postgresql-client \
    redis-tools \
    make \
    curl \
    wget \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install github.com/cosmtrek/air@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install github.com/swaggo/swag/cmd/swag@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest

# Install global npm packages
RUN npm install -g create-react-app typescript @types/node

# Set up workspace
WORKDIR /workspace

# Configure git (will be overridden by user settings)
RUN git config --global --add safe.directory /workspace


