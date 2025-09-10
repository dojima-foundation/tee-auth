#!/bin/bash

# Quick GitHub Actions Runner Tools Installation Script
# This is a simplified version for quick installation
# Usage: ./quick_install_tools.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

# Set non-interactive mode
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

log "Starting quick tools installation..."

# Update system
log "Updating system packages..."
apt-get update -y
apt-get upgrade -y -o Dpkg::Options::="--force-confold"

# Install essential packages
log "Installing essential packages..."
apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    apt-transport-https \
    unzip \
    jq \
    htop \
    vim \
    net-tools \
    postgresql-client \
    postgresql-client-common \
    redis-tools \
    netcat-openbsd \
    sudo \
    protobuf-compiler \
    telnet \
    dnsutils \
    iputils-ping \
    pkg-config \
    libssl-dev

# Install Docker
log "Installing Docker..."
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
systemctl start docker
systemctl enable docker
usermod -aG docker ubuntu
chmod 666 /var/run/docker.sock 2>/dev/null || true

# Install Go 1.23.0
log "Installing Go 1.23.0..."
wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Set up Go environment
export PATH="/usr/local/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/home/ubuntu/.cache/go-build"
export GOTMPDIR="/home/ubuntu/.cache/go-tmp"
export GOMODCACHE="/home/ubuntu/go/pkg/mod"
export GOSUMDB="sum.golang.org"
export GOPROXY="https://proxy.golang.org,direct"

# Add Go environment to profiles
echo 'export PATH="/usr/local/go/bin:$PATH"' | tee -a /etc/profile
echo 'export GOPATH="/home/ubuntu/go"' | tee -a /etc/profile
echo 'export GOROOT="/usr/local/go"' | tee -a /etc/profile
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' | tee -a /etc/profile
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' | tee -a /etc/profile
echo 'export GOMODCACHE="/home/ubuntu/go/pkg/mod"' | tee -a /etc/profile
echo 'export GOSUMDB="sum.golang.org"' | tee -a /etc/profile
echo 'export GOPROXY="https://proxy.golang.org,direct"' | tee -a /etc/profile

echo 'export PATH="/usr/local/go/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'export GOPATH="/home/ubuntu/go"' >> /home/ubuntu/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/ubuntu/.bashrc
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /home/ubuntu/.bashrc
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /home/ubuntu/.bashrc
echo 'export GOMODCACHE="/home/ubuntu/go/pkg/mod"' >> /home/ubuntu/.bashrc
echo 'export GOSUMDB="sum.golang.org"' >> /home/ubuntu/.bashrc
echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/ubuntu/.bashrc

# Add Go environment to runner user's bashrc (for GitHub Actions)
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/runner/.bashrc
echo 'export GOPATH="/home/ubuntu/go"' >> /home/runner/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/runner/.bashrc
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /home/runner/.bashrc
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /home/runner/.bashrc
echo 'export GOMODCACHE="/home/ubuntu/go/pkg/mod"' >> /home/runner/.bashrc
echo 'export GOSUMDB="sum.golang.org"' >> /home/runner/.bashrc
echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/runner/.bashrc

# Add Go environment to github-runner user's bashrc (for GitHub Actions)
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/github-runner/.bashrc
echo 'export GOPATH="/home/github-runner/go"' >> /home/github-runner/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/github-runner/.bashrc
echo 'export GOCACHE="/opt/go-cache/go-build"' >> /home/github-runner/.bashrc
echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /home/github-runner/.bashrc
echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /home/github-runner/.bashrc
echo 'export GOSUMDB="sum.golang.org"' >> /home/github-runner/.bashrc
echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/github-runner/.bashrc

# Create necessary directories
mkdir -p /home/ubuntu/go/pkg /home/ubuntu/.cache/go-build /home/ubuntu/.cache/go-tmp
mkdir -p /home/runner/go/bin
chown -R runner:runner /home/runner/go

# Create github-runner user and directories
useradd -m -s /bin/bash github-runner 2>/dev/null || echo "User already exists"
usermod -aG docker github-runner
usermod -aG sudo github-runner
mkdir -p /home/github-runner/go
chown -R github-runner:github-runner /home/github-runner

# Create Go cache directory for CI/CD workflows
mkdir -p /opt/go-cache/go-build /opt/go-cache/go-tmp /opt/go-cache/go-mod
chown -R github-runner:github-runner /opt/go-cache
chmod -R 755 /opt/go-cache

# Install Go development tools
log "Installing Go development tools..."
# Ensure full PATH is available
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/home/ubuntu/go/bin"
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

# Install golang-migrate tool
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/

# Update PATH to include Go bin for both users
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/runner/.bashrc

# Install Rust
log "Installing Rust..."
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain 1.82.0
source ~/.cargo/env
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /etc/profile
echo 'source ~/.cargo/env 2>/dev/null || true' >> /etc/profile
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'source ~/.cargo/env 2>/dev/null || true' >> /home/ubuntu/.bashrc

# Install Node.js 20.x
log "Installing Node.js 20.x..."
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs

# Install Playwright system dependencies
apt-get install -y \
    libnss3-dev \
    libatk-bridge2.0-dev \
    libdrm2 \
    libxkbcommon-dev \
    libxcomposite-dev \
    libxdamage-dev \
    libxrandr-dev \
    libgbm-dev \
    libxss1 \
    libasound2-dev \
    libgtk-4-1 \
    libvpx7 \
    libevent-2.1-7 \
    libflite1 \
    libavif13 \
    libwebpmux3 \
    libenchant-2-2 \
    libsecret-1-0 \
    libhyphen0 \
    libmanette-0.2-0 \
    libgles2-mesa \
    libx264-163

# Install Playwright browsers
npx playwright install

# Install additional Node.js tools
npm install -g @lhci/cli
npm install -g lighthouse
npm install -g typescript
npm install -g ts-node

# Clean up
rm -f go1.23.0.linux-amd64.tar.gz

# Verify installation
log "Verifying installation..."
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/home/ubuntu/go/bin"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"

echo "=== Installation Verification ==="
echo "Go version: $(go version)"
echo "Protoc version: $(protoc --version)"
echo "Protoc-gen-go: $(which protoc-gen-go)"
echo "Protoc-gen-go-grpc: $(which protoc-gen-go-grpc)"
echo "Migration tool: $(migrate -version)"
echo "Docker version: $(docker --version)"
echo "PostgreSQL client: $(psql --version)"
echo "Redis client: $(redis-cli --version)"
echo "✅ All tools verified and ready!"

log "Quick installation completed successfully!"
log "Run 'source ~/.bashrc' to reload environment variables"
