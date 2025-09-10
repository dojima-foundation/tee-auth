#!/bin/bash

# GitHub Actions Runner Tools Installation Script
# This script installs all required development tools for the tee-auth project
# Usage: ./install_tools.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Set non-interactive mode for all package operations
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install system packages
install_system_packages() {
    log "Installing system packages..."
    
    # Update package lists
    apt-get update -y
    
    # Install essential packages
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
        postgresql-common \
        redis-tools \
        netcat-openbsd \
        sudo \
        protobuf-compiler \
        telnet \
        dnsutils \
        iputils-ping \
        pkg-config \
        libssl-dev \
        libffi-dev \
        python3-dev \
        python3-pip \
        python3-venv \
        make \
        cmake \
        gcc \
        g++ \
        clang \
        llvm
    
    log "System packages installed successfully"
}

# Function to install Docker
install_docker() {
    log "Installing Docker..."
    
    # Install Docker if not already installed
    if ! command_exists docker; then
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
        apt-get update -y
        apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        
        # Start and enable Docker
        systemctl start docker
        systemctl enable docker
        
        # Add ubuntu user to docker group
        usermod -aG docker ubuntu
        chmod 666 /var/run/docker.sock 2>/dev/null || true
    else
        info "Docker is already installed"
    fi
    
    log "Docker installation completed"
}

# Function to install Go
install_go() {
    log "Installing Go 1.23.0..."
    
    if ! command_exists go; then
        # Download and install Go
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
        
        # Add Go environment to system-wide profile
        echo 'export PATH="/usr/local/go/bin:$PATH"' | tee -a /etc/profile
        echo 'export GOPATH="/home/ubuntu/go"' | tee -a /etc/profile
        echo 'export GOROOT="/usr/local/go"' | tee -a /etc/profile
        echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' | tee -a /etc/profile
        echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' | tee -a /etc/profile
        echo 'export GOMODCACHE="/home/ubuntu/go/pkg/mod"' | tee -a /etc/profile
        echo 'export GOSUMDB="sum.golang.org"' | tee -a /etc/profile
        echo 'export GOPROXY="https://proxy.golang.org,direct"' | tee -a /etc/profile
        
        # Add Go environment to ubuntu user's bashrc
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

        # Add Go environment to runner user's bashrc (for GitHub Actions workflows)
        echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/runner/.bashrc
        echo 'export GOPATH="/home/runner/go"' >> /home/runner/.bashrc
        echo 'export GOROOT="/usr/local/go"' >> /home/runner/.bashrc
        echo 'export GOCACHE="/opt/go-cache/go-build"' >> /home/runner/.bashrc
        echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /home/runner/.bashrc
        echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /home/runner/.bashrc
        echo 'export GOSUMDB="sum.golang.org"' >> /home/runner/.bashrc
        echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/runner/.bashrc
        
        # Create necessary directories
        mkdir -p /home/ubuntu/go/pkg /home/ubuntu/.cache/go-build /home/ubuntu/.cache/go-tmp
        mkdir -p /home/runner/go/bin
        chown -R runner:runner /home/runner/go
        
        # Add runner user to ubuntu group for cache access
        usermod -aG ubuntu runner
        
        # Create github-runner user and directories
        useradd -m -s /bin/bash github-runner 2>/dev/null || echo "User already exists"
        usermod -aG docker github-runner
        usermod -aG sudo github-runner
        usermod -aG ubuntu github-runner
        mkdir -p /home/github-runner/go
        chown -R github-runner:github-runner /home/github-runner
        
        # Create Go cache directory for CI/CD workflows
        mkdir -p /opt/go-cache/go-build /opt/go-cache/go-tmp /opt/go-cache/go-mod/cache /opt/go-cache/go-sumdb
        
        # Set permissive permissions for GitHub Actions runner user
        chown -R runner:ubuntu /opt/go-cache
        chmod -R 777 /opt/go-cache
        
        # Ensure runner user can write to github-runner's Go bin directory
        chown -R runner:ubuntu /home/github-runner/go
        chmod -R 775 /home/github-runner/go
        
        # Clean up
        rm -f go1.23.0.linux-amd64.tar.gz
    else
        info "Go is already installed: $(go version)"
    fi
    
    log "Go installation completed"
}

# Function to install Go development tools
install_go_tools() {
    log "Installing Go development tools..."
    
    # Source Go environment with full PATH
    export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/home/ubuntu/go/bin"
    export GOPATH="/home/ubuntu/go"
    export GOROOT="/usr/local/go"
    export GOCACHE="/home/ubuntu/.cache/go-build"
    export GOTMPDIR="/home/ubuntu/.cache/go-tmp"
    export GOMODCACHE="/home/ubuntu/go/pkg/mod"
    export GOSUMDB="sum.golang.org"
    export GOPROXY="https://proxy.golang.org,direct"
    
    # Install Go development tools
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4
    
    # Install golang-migrate tool with postgres support
    go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    
    # Copy Go tools to github-runner directory for CI/CD access
    cp /home/ubuntu/go/bin/protoc-gen-go /home/github-runner/go/bin/ 2>/dev/null || true
    cp /home/ubuntu/go/bin/protoc-gen-go-grpc /home/github-runner/go/bin/ 2>/dev/null || true
    cp /home/ubuntu/go/bin/golangci-lint /home/github-runner/go/bin/ 2>/dev/null || true
    cp /home/ubuntu/go/bin/goimports /home/github-runner/go/bin/ 2>/dev/null || true
    cp /home/ubuntu/go/bin/gosec /home/github-runner/go/bin/ 2>/dev/null || true
    cp /home/ubuntu/go/bin/migrate /home/github-runner/go/bin/ 2>/dev/null || true
    chown github-runner:github-runner /home/github-runner/go/bin/* 2>/dev/null || true
    chmod +x /home/github-runner/go/bin/* 2>/dev/null || true
    
    # Update PATH to include Go bin for both users
    echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/ubuntu/.bashrc
    echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/runner/.bashrc
    
    log "Go development tools installed successfully"
}

# Function to install Rust
install_rust() {
    log "Installing Rust toolchain..."
    
    if ! command_exists rustc; then
        # Install Rust using rustup
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain 1.82.0
        
        # Source Rust environment
        source ~/.cargo/env
        
        # Add Rust to system profiles
        echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /etc/profile
        echo 'source ~/.cargo/env 2>/dev/null || true' >> /etc/profile
        echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /home/ubuntu/.bashrc
        echo 'source ~/.cargo/env 2>/dev/null || true' >> /home/ubuntu/.bashrc
        
        # Install Rust development tools
        cargo install cargo-audit
        cargo install cargo-deny
        cargo install cargo-tarpaulin --version 0.27.0 --locked
    else
        info "Rust is already installed: $(rustc --version)"
    fi
    
    log "Rust installation completed"
}

# Function to install Node.js
install_nodejs() {
    log "Installing Node.js 20.x..."
    
    if ! command_exists node; then
        # Install Node.js using NodeSource repository
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
    else
        info "Node.js is already installed: $(node --version)"
    fi
    
    log "Node.js installation completed"
}

# Function to install additional tools
install_additional_tools() {
    log "Installing additional development tools..."
    
    # Install GitHub CLI
    if ! command_exists gh; then
        apt-get install -y gh
    fi
    
    # Install Git LFS
    if ! command_exists git-lfs; then
        apt-get install -y git-lfs
        git lfs install --system
    fi
    
    # Install additional development tools
    apt-get install -y \
        git-flow \
        hub \
        git-crypt \
        git-secrets \
        pre-commit \
        shellcheck \
        yamllint \
        ansible-lint \
        hadolint \
        docker-compose \
        kubectl \
        helm \
        terraform \
        awscli \
        azure-cli \
        gcloud-cli
    
    # Install Python tools for CI/CD
    pip3 install --upgrade pip
    pip3 install \
        requests \
        pyyaml \
        jinja2 \
        ansible \
        docker \
        kubernetes \
        boto3 \
        azure-mgmt-resource \
        google-cloud-storage \
        pytest \
        pytest-cov \
        pytest-xdist \
        black \
        flake8 \
        mypy \
        bandit \
        safety
    
    log "Additional tools installed successfully"
}

# Function to create utility scripts
create_utility_scripts() {
    log "Creating utility scripts..."
    
    # Create database connection testing scripts
    cat > /usr/local/bin/test-postgres-connection << 'EOF'
#!/bin/bash
# PostgreSQL connection test script

HOST=${1:-localhost}
PORT=${2:-5432}
USER=${3:-gauth}
PASSWORD=${4:-password}
DATABASE=${5:-postgres}

echo "Testing PostgreSQL connection to $HOST:$PORT..."
echo "User: $USER, Database: $DATABASE"

# Test with pg_isready
if command -v pg_isready &> /dev/null; then
    if pg_isready -h "$HOST" -p "$PORT" -U "$USER"; then
        echo "✅ PostgreSQL server is ready (pg_isready)"
    else
        echo "❌ PostgreSQL server is not ready (pg_isready)"
        exit 1
    fi
else
    echo "⚠️  pg_isready not available, using alternative test"
fi

# Test with psql
if command -v psql &> /dev/null; then
    if PGPASSWORD="$PASSWORD" psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DATABASE" -c "SELECT 1;" &> /dev/null; then
        echo "✅ PostgreSQL connection successful (psql)"
    else
        echo "❌ PostgreSQL connection failed (psql)"
        exit 1
    fi
else
    echo "⚠️  psql not available"
fi

# Test with netcat as fallback
if command -v nc &> /dev/null; then
    if nc -z "$HOST" "$PORT"; then
        echo "✅ PostgreSQL port is open (netcat)"
    else
        echo "❌ PostgreSQL port is not accessible (netcat)"
        exit 1
    fi
else
    echo "⚠️  netcat not available"
fi

echo "✅ All PostgreSQL connection tests passed!"
EOF

    # Redis connection test script
    cat > /usr/local/bin/test-redis-connection << 'EOF'
#!/bin/bash
# Redis connection test script

HOST=${1:-localhost}
PORT=${2:-6379}
PASSWORD=${3:-}

echo "Testing Redis connection to $HOST:$PORT..."

# Test with redis-cli
if command -v redis-cli &> /dev/null; then
    if [ -n "$PASSWORD" ]; then
        if redis-cli -h "$HOST" -p "$PORT" -a "$PASSWORD" ping | grep -q "PONG"; then
            echo "✅ Redis connection successful (redis-cli with auth)"
        else
            echo "❌ Redis connection failed (redis-cli with auth)"
            exit 1
        fi
    else
        if redis-cli -h "$HOST" -p "$PORT" ping | grep -q "PONG"; then
            echo "✅ Redis connection successful (redis-cli)"
        else
            echo "❌ Redis connection failed (redis-cli)"
            exit 1
        fi
    fi
else
    echo "⚠️  redis-cli not available"
fi

# Test with netcat as fallback
if command -v nc &> /dev/null; then
    if nc -z "$HOST" "$PORT"; then
        echo "✅ Redis port is open (netcat)"
    else
        echo "❌ Redis port is not accessible (netcat)"
        exit 1
    fi
else
    echo "⚠️  netcat not available"
fi

echo "✅ All Redis connection tests passed!"
EOF

    # Database monitoring script
    cat > /usr/local/bin/db-monitor << 'EOF'
#!/bin/bash
# Database monitoring script

echo "=== Database Services Status ==="
echo "PostgreSQL:"
if command -v pg_isready &> /dev/null; then
    pg_isready -h localhost -p 5432 2>/dev/null && echo "  ✅ PostgreSQL (localhost:5432) is ready" || echo "  ❌ PostgreSQL (localhost:5432) is not ready"
else
    echo "  ⚠️  pg_isready not available"
fi

echo "Redis:"
if command -v redis-cli &> /dev/null; then
    redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q "PONG" && echo "  ✅ Redis (localhost:6379) is ready" || echo "  ❌ Redis (localhost:6379) is not ready"
else
    echo "  ⚠️  redis-cli not available"
fi

echo ""
echo "=== Network Connectivity ==="
echo "PostgreSQL port 5432:"
nc -z localhost 5432 2>/dev/null && echo "  ✅ Port 5432 is open" || echo "  ❌ Port 5432 is closed"

echo "Redis port 6379:"
nc -z localhost 6379 2>/dev/null && echo "  ✅ Port 6379 is open" || echo "  ❌ Port 6379 is closed"

echo ""
echo "=== Available Tools ==="
echo "PostgreSQL tools:"
command -v psql &> /dev/null && echo "  ✅ psql" || echo "  ❌ psql"
command -v pg_isready &> /dev/null && echo "  ✅ pg_isready" || echo "  ❌ pg_isready"

echo "Redis tools:"
command -v redis-cli &> /dev/null && echo "  ✅ redis-cli" || echo "  ❌ redis-cli"

echo "Network tools:"
command -v nc &> /dev/null && echo "  ✅ netcat" || echo "  ❌ netcat"
command -v telnet &> /dev/null && echo "  ✅ telnet" || echo "  ❌ telnet"
EOF

    # Runner information script
    cat > /usr/local/bin/runner-info << 'EOF'
#!/bin/bash

echo "=== GitHub Actions Runner Information ==="
echo "Hostname: $(hostname)"
echo "OS: $(lsb_release -d | cut -f2)"
echo "Kernel: $(uname -r)"
echo "Architecture: $(uname -m)"
echo "CPU: $(nproc) cores"
echo "Memory: $(free -h | awk '/^Mem:/{print $2}')"
echo "Disk: $(df -h / | awk 'NR==2{print $4}') available"
echo "Uptime: $(uptime -p)"
echo "Docker Version: $(docker --version 2>/dev/null || echo 'Not installed')"
echo "Runner Service: $(systemctl is-active actions.runner.* 2>/dev/null || echo 'Not running')"
echo ""
echo "=== Development Tools ==="
echo "Go Version: $(go version 2>/dev/null || echo 'Not installed')"
echo "protoc-gen-go: $(protoc-gen-go --version 2>/dev/null || echo 'Not installed')"
echo "protoc-gen-go-grpc: $(protoc-gen-go-grpc --version 2>/dev/null || echo 'Not installed')"
echo "Rust Version: $(rustc --version 2>/dev/null || echo 'Not installed')"
echo "Node.js Version: $(node --version 2>/dev/null || echo 'Not installed')"
echo "NPM Version: $(npm --version 2>/dev/null || echo 'Not installed')"
echo ""
echo "=== Database Tools ==="
echo "PostgreSQL Client: $(psql --version 2>/dev/null || echo 'Not installed')"
echo "Redis CLI: $(redis-cli --version 2>/dev/null || echo 'Not installed')"
echo "Migration Tool: $(migrate -version 2>/dev/null || echo 'Not installed')"
echo ""
echo "=== Build Tools ==="
echo "Make: $(make --version 2>/dev/null | head -n1 || echo 'Not installed')"
echo "CMake: $(cmake --version 2>/dev/null | head -n1 || echo 'Not installed')"
echo "GCC: $(gcc --version 2>/dev/null | head -n1 || echo 'Not installed')"
echo "Python: $(python3 --version 2>/dev/null || echo 'Not installed')"
echo ""
echo "=== Performance & Testing Tools ==="
echo "Lighthouse: $(lighthouse --version 2>/dev/null || echo 'Not installed')"
echo "Lighthouse CI: $(lhci --version 2>/dev/null || echo 'Not installed')"
echo "Playwright: $(npx playwright --version 2>/dev/null || echo 'Not installed')"
echo ""
echo "=== CI/CD Tools ==="
echo "GitHub CLI: $(gh --version 2>/dev/null | head -n1 || echo 'Not installed')"
echo "Kubernetes CLI: $(kubectl version --client --short 2>/dev/null || echo 'Not installed')"
echo "Terraform: $(terraform version 2>/dev/null | head -n1 || echo 'Not installed')"
echo "AWS CLI: $(aws --version 2>/dev/null || echo 'Not installed')"
echo "Docker Compose: $(docker-compose --version 2>/dev/null || echo 'Not installed')"
echo "========================================="
EOF

    # Make scripts executable
    chmod +x /usr/local/bin/test-postgres-connection
    chmod +x /usr/local/bin/test-redis-connection
    chmod +x /usr/local/bin/db-monitor
    chmod +x /usr/local/bin/runner-info
    
    log "Utility scripts created successfully"
}

# Function to verify installation
verify_installation() {
    log "Verifying installation..."
    
    # Source Go environment with full PATH
    export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/home/ubuntu/go/bin"
    export GOPATH="/home/ubuntu/go"
    export GOROOT="/usr/local/go"
    export GOCACHE="/home/ubuntu/.cache/go-build"
    export GOTMPDIR="/home/ubuntu/.cache/go-tmp"
    export GOMODCACHE="/home/ubuntu/go/pkg/mod"
    export GOSUMDB="sum.golang.org"
    export GOPROXY="https://proxy.golang.org,direct"
    
    echo ""
    echo "=== Installation Verification ==="
    
    # Check Go
    if command_exists go; then
        echo "✅ Go: $(go version)"
    else
        echo "❌ Go: Not installed"
    fi
    
    # Check Go tools
    if command_exists protoc-gen-go; then
        echo "✅ protoc-gen-go: $(protoc-gen-go --version)"
    else
        echo "❌ protoc-gen-go: Not installed"
    fi
    
    if command_exists protoc-gen-go-grpc; then
        echo "✅ protoc-gen-go-grpc: $(protoc-gen-go-grpc --version)"
    else
        echo "❌ protoc-gen-go-grpc: Not installed"
    fi
    
    if command_exists golangci-lint; then
        echo "✅ golangci-lint: $(golangci-lint --version)"
    else
        echo "❌ golangci-lint: Not installed"
    fi
    
    # Check database tools
    if command_exists psql; then
        echo "✅ PostgreSQL client: $(psql --version)"
    else
        echo "❌ PostgreSQL client: Not installed"
    fi
    
    if command_exists redis-cli; then
        echo "✅ Redis CLI: $(redis-cli --version)"
    else
        echo "❌ Redis CLI: Not installed"
    fi
    
    # Check Rust
    if command_exists rustc; then
        echo "✅ Rust: $(rustc --version)"
    else
        echo "❌ Rust: Not installed"
    fi
    
    # Check Node.js
    if command_exists node; then
        echo "✅ Node.js: $(node --version)"
    else
        echo "❌ Node.js: Not installed"
    fi
    
    if command_exists npm; then
        echo "✅ NPM: $(npm --version)"
    else
        echo "❌ NPM: Not installed"
    fi
    
    # Check Docker
    if command_exists docker; then
        echo "✅ Docker: $(docker --version)"
    else
        echo "❌ Docker: Not installed"
    fi
    
    echo ""
    log "Installation verification completed"
}

# Main installation function
main() {
    echo "=========================================="
    echo "GitHub Actions Runner Tools Installation"
    echo "=========================================="
    echo ""
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root. Please run as ubuntu user."
        exit 1
    fi
    
    # Check if running on Ubuntu
    if ! grep -q "Ubuntu" /etc/os-release; then
        error "This script is designed for Ubuntu systems only."
        exit 1
    fi
    
    log "Starting tools installation..."
    
    # Install tools in order
    install_system_packages
    install_docker
    install_go
    install_go_tools
    install_rust
    install_nodejs
    install_additional_tools
    create_utility_scripts
    verify_installation
    
    echo ""
    echo "=========================================="
    log "🎉 All tools installed successfully!"
    echo "=========================================="
    echo ""
    echo "Available utility scripts:"
    echo "  - test-postgres-connection: Test PostgreSQL connectivity"
    echo "  - test-redis-connection: Test Redis connectivity"
    echo "  - db-monitor: Monitor database services"
    echo "  - runner-info: Display runner information"
    echo ""
    echo "To use the tools, you may need to:"
    echo "  1. Log out and log back in, or"
    echo "  2. Run: source ~/.bashrc"
    echo ""
    echo "Your GitHub Actions runner is now ready for CI/CD workflows!"
}

# Run main function
main "$@"
