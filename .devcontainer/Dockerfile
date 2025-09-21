FROM mcr.microsoft.com/devcontainers/go:1.25

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs

# Install additional tools
RUN apt-get update && apt-get install -y \
    chromium=138.0.7204.183-1~deb12u1 \
    postgresql-client \
    redis-tools \
    make \
    curl \
    wget \
    git \
    xdg-utils \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install github.com/cosmtrek/air@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install github.com/swaggo/swag/cmd/swag@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest

# Install global npm packages
RUN npm install -g create-react-app typescript @types/node

# Install chromedriver
RUN wget -O chromedriver-v37.2.6-linux-arm64.zip https://github.com/electron/electron/releases/download/v37.2.6/chromedriver-v37.2.6-linux-arm64.zip && \
    unzip chromedriver-v37.2.6-linux-arm64.zip && \
    mv chromedriver /usr/local/bin/chromedriver

ARG PLAYWRIGHT_VERSION=1.51.1
RUN npm install -g pnpm@latest-10 && npm install -g playwright@${PLAYWRIGHT_VERSION}
RUN npx playwright install --with-deps

# Set up workspace
WORKDIR /workspace

# Configure git (will be overridden by user settings)
RUN git config --global --add safe.directory /workspace


