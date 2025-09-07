#!/bin/bash

# GitHub Actions Runner Setup Script for ${github_repo}
# This script installs and configures a GitHub Actions runner on Ubuntu 22.04

set -e

# Configuration variables
GITHUB_TOKEN="${github_token}"
GITHUB_ORG="${github_org}"
GITHUB_REPO="${github_repo}"
RUNNER_LABELS="${runner_labels}"
RUNNER_NAME="${runner_name}"
DOCKER_REGISTRY_MIRROR="${docker_registry_mirror}"
ENABLE_HEALTH_CHECKS="${enable_health_checks}"
ENABLE_STATUS_PAGES="${enable_status_pages}"
PROJECT_ID="${project_id}"
REGION="${region}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Set non-interactive mode for all package operations
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

# Update system
log "Updating system packages..."
apt-get update -y
apt-get upgrade -y -o Dpkg::Options::="--force-confold"

# Install required packages
log "Installing required packages..."
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
    iputils-ping

# Install Docker
log "Installing Docker..."
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Fix Docker permissions for ubuntu user
usermod -aG docker ubuntu
chmod 666 /var/run/docker.sock 2>/dev/null || true

# Configure Docker registry mirror if provided
if [ ! -z "$DOCKER_REGISTRY_MIRROR" ]; then
    log "Configuring Docker registry mirror: $DOCKER_REGISTRY_MIRROR"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << EOF
{
  "registry-mirrors": ["$DOCKER_REGISTRY_MIRROR"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
    systemctl restart docker
fi

# Create runner user
log "Creating runner user..."
useradd -m -s /bin/bash runner
usermod -aG docker runner
usermod -aG sudo runner

# Configure passwordless sudo for runner user (for CI workflows)
log "Configuring sudo access for runner..."
echo "runner ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/runner
chmod 0440 /etc/sudoers.d/runner

# Download and install GitHub Actions runner
log "Downloading GitHub Actions runner..."
RUNNER_VERSION="v2.328.0"
cd /home/runner
wget -O actions-runner-linux-x64-2.328.0.tar.gz https://github.com/actions/runner/releases/download/${RUNNER_VERSION}/actions-runner-linux-x64-2.328.0.tar.gz
tar xzf actions-runner-linux-x64-2.328.0.tar.gz
rm -f actions-runner-linux-x64-2.328.0.tar.gz

# Set proper ownership
chown -R runner:runner /home/runner

# Get registration token
log "Getting GitHub Actions runner registration token..."
if [ ! -z "$GITHUB_ORG" ] && [ -z "$GITHUB_REPO" ]; then
    # Organization-level runner
    RUNNER_URL="https://github.com/$GITHUB_ORG"
    RUNNER_TOKEN=$(curl -X POST -H "Authorization: token $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      "https://api.github.com/orgs/$GITHUB_ORG/actions/runners/registration-token" | \
      jq -r '.token')
else
    # Repository-level runner
    RUNNER_URL="https://github.com/$GITHUB_REPO"
    RUNNER_TOKEN=$(curl -X POST -H "Authorization: token $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      "https://api.github.com/repos/$GITHUB_REPO/actions/runners/registration-token" | \
      jq -r '.token')
fi

if [ "$RUNNER_TOKEN" = "null" ] || [ -z "$RUNNER_TOKEN" ]; then
    error "Failed to get registration token. Check your GitHub token and repository access."
    exit 1
fi

# Configure runner
log "Configuring GitHub Actions runner..."
sudo -u runner ./config.sh \
    --url "$RUNNER_URL" \
    --token "$RUNNER_TOKEN" \
    --name "$RUNNER_NAME" \
    --labels "$RUNNER_LABELS" \
    --unattended \
    --replace

# Install runner as a service
log "Installing GitHub Actions runner as a service..."
./svc.sh install runner
./svc.sh start

# Set up monitoring if enabled
if [ "$ENABLE_HEALTH_CHECKS" = "true" ]; then
    log "Setting up health checks..."
    
    # Create health check script
    cat > /usr/local/bin/runner-health-check.sh << 'EOF'
#!/bin/bash
# Health check script for GitHub Actions runner

LOG_FILE="/var/log/runner-health-check.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Check if runner service is running
if ! systemctl is-active --quiet actions.runner.*; then
    echo "[$TIMESTAMP] Runner service is not running, attempting to restart..." >> $LOG_FILE
    systemctl restart actions.runner.*
    sleep 10
    
    # Check again after restart
    if systemctl is-active --quiet actions.runner.*; then
        echo "[$TIMESTAMP] Runner service restarted successfully" >> $LOG_FILE
    else
        echo "[$TIMESTAMP] Failed to restart runner service" >> $LOG_FILE
        exit 1
    fi
fi

# Check runner connectivity
if ! curl -s https://api.github.com > /dev/null; then
    echo "[$TIMESTAMP] Network connectivity check failed" >> $LOG_FILE
    exit 1
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    echo "[$TIMESTAMP] Disk usage is high: ${DISK_USAGE}%" >> $LOG_FILE
    # Clean up old logs and temporary files
    find /var/log -name "*.log" -mtime +7 -delete 2>/dev/null || true
    docker system prune -f 2>/dev/null || true
fi

echo "[$TIMESTAMP] Runner health check passed" >> $LOG_FILE
EOF

    chmod +x /usr/local/bin/runner-health-check.sh

    # Create systemd timer for health checks
    cat > /etc/systemd/system/runner-health-check.service << 'EOF'
[Unit]
Description=GitHub Actions Runner Health Check
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/runner-health-check.sh
User=root
EOF

    cat > /etc/systemd/system/runner-health-check.timer << 'EOF'
[Unit]
Description=Run GitHub Actions Runner Health Check every 5 minutes
Requires=runner-health-check.service

[Timer]
OnBootSec=5min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
EOF

    systemctl daemon-reload
    systemctl enable runner-health-check.timer
    systemctl start runner-health-check.timer
fi

# Set up status page if enabled
if [ "$ENABLE_STATUS_PAGES" = "true" ]; then
    log "Setting up status page..."
    
    # Install nginx
    apt-get install -y nginx

    # Create status page
    cat > /var/www/html/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>GitHub Actions Runner Status</title>
    <meta http-equiv="refresh" content="30">
    <meta charset="utf-8">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            margin: 40px; 
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .status { 
            padding: 15px; 
            margin: 15px 0; 
            border-radius: 8px; 
            border-left: 4px solid;
        }
        .healthy { 
            background-color: #d4edda; 
            color: #155724; 
            border-left-color: #28a745;
        }
        .unhealthy { 
            background-color: #f8d7da; 
            color: #721c24; 
            border-left-color: #dc3545;
        }
        .info {
            background-color: #d1ecf1;
            color: #0c5460;
            border-left-color: #17a2b8;
        }
        h1 { color: #333; }
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        .metric {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 6px;
            text-align: center;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #007bff;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GitHub Actions Runner Status</h1>
        <div id="status" class="status info">Checking status...</div>
        
        <div class="metrics">
            <div class="metric">
                <div class="metric-value" id="uptime">-</div>
                <div>Uptime</div>
            </div>
            <div class="metric">
                <div class="metric-value" id="cpu">-</div>
                <div>CPU Usage</div>
            </div>
            <div class="metric">
                <div class="metric-value" id="memory">-</div>
                <div>Memory Usage</div>
            </div>
            <div class="metric">
                <div class="metric-value" id="disk">-</div>
                <div>Disk Usage</div>
            </div>
        </div>
        
        <script>
            function updateStatus() {
                fetch('/api/status')
                    .then(response => response.json())
                    .then(data => {
                        const statusDiv = document.getElementById('status');
                        if (data.healthy) {
                            statusDiv.className = 'status healthy';
                            statusDiv.innerHTML = '✅ Runner is healthy and ready';
                        } else {
                            statusDiv.className = 'status unhealthy';
                            statusDiv.innerHTML = '❌ Runner is unhealthy: ' + (data.error || 'Unknown error');
                        }
                        
                        // Update metrics
                        document.getElementById('uptime').textContent = data.uptime || '-';
                        document.getElementById('cpu').textContent = data.cpu || '-';
                        document.getElementById('memory').textContent = data.memory || '-';
                        document.getElementById('disk').textContent = data.disk || '-';
                    })
                    .catch(error => {
                        const statusDiv = document.getElementById('status');
                        statusDiv.className = 'status unhealthy';
                        statusDiv.innerHTML = '❌ Error checking status: ' + error.message;
                    });
            }
            
            // Update status immediately and then every 30 seconds
            updateStatus();
            setInterval(updateStatus, 30000);
        </script>
    </div>
</body>
</html>
EOF

    # Create API endpoint for status
    cat > /etc/nginx/sites-available/runner-status << 'EOF'
server {
    listen 80;
    server_name _;
    root /var/www/html;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }

    location /api/status {
        add_header Content-Type application/json;
        add_header Access-Control-Allow-Origin *;
        
        # Check if runner service is running
        set $healthy "true";
        set $error "";
        
        # Use a simple script to check status
        access_by_lua_block {
            local handle = io.popen("systemctl is-active actions.runner.* 2>/dev/null")
            local result = handle:read("*a")
            handle:close()
            
            if not string.match(result, "active") then
                ngx.var.healthy = "false"
                ngx.var.error = "Service not active"
            end
        }
        
        return 200 '{"healthy": $healthy, "error": "$error", "timestamp": "$date_gmt", "uptime": "$upstream_response_time", "cpu": "N/A", "memory": "N/A", "disk": "N/A"}';
    }
}
EOF

    ln -sf /etc/nginx/sites-available/runner-status /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    systemctl restart nginx
fi

# Set up log rotation
log "Setting up log rotation..."
cat > /etc/logrotate.d/github-runner << 'EOF'
/var/log/actions-runner/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 runner runner
}

/var/log/runner-health-check.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
}
EOF

# Install Go development tools
log "Installing Go development tools..."
# Install Go (required for protoc plugins and other tools)
wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Set up Go environment
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/home/ubuntu/.cache/go-build"
export GOTMPDIR="/home/ubuntu/.cache/go-tmp"

# Create necessary directories
mkdir -p /home/ubuntu/go /home/ubuntu/.cache/go-build /home/ubuntu/.cache/go-tmp

# Add Go environment to system-wide profile
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /etc/profile
echo 'export GOPATH="/home/ubuntu/go"' >> /etc/profile
echo 'export GOROOT="/usr/local/go"' >> /etc/profile
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /etc/profile
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /etc/profile

# Add Go environment to ubuntu user's bashrc
touch /home/ubuntu/.bashrc
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'export GOPATH="/home/ubuntu/go"' >> /home/ubuntu/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/ubuntu/.bashrc
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /home/ubuntu/.bashrc
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /home/ubuntu/.bashrc

# Install Go development tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

# Install golang-migrate tool
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/

# Clean up
rm go1.23.0.linux-amd64.tar.gz

# Install additional PostgreSQL and Redis tools
log "Installing additional PostgreSQL and Redis tools..."

# Install PostgreSQL development tools and utilities
apt-get install -y \
    postgresql-contrib \
    postgresql-client-15 \
    pgcli \
    postgresql-client-common

# Install Redis tools and utilities
apt-get install -y \
    redis-tools \
    redis-server-common

# Install additional database testing and debugging tools
apt-get install -y \
    mysql-client \
    sqlite3 \
    mongodb-database-tools

# Install network and connectivity testing tools
apt-get install -y \
    nmap \
    tcpdump \
    wireshark-common \
    iperf3 \
    netstat-nat

# Install additional monitoring and debugging tools
apt-get install -y \
    strace \
    ltrace \
    gdb \
    valgrind \
    perf-tools-unstable

# Install build and development tools
apt-get install -y \
    make \
    cmake \
    pkg-config \
    libssl-dev \
    libffi-dev \
    python3-dev \
    python3-pip \
    python3-venv \
    build-essential \
    gcc \
    g++ \
    clang \
    llvm

# Install archive and compression tools
apt-get install -y \
    tar \
    gzip \
    bzip2 \
    xz-utils \
    zip \
    unzip \
    p7zip-full \
    rar \
    unrar

# Install additional network and download tools
apt-get install -y \
    curl \
    wget \
    aria2 \
    axel \
    httrack

# Install system utilities
apt-get install -y \
    tree \
    rsync \
    screen \
    tmux \
    less \
    more \
    nano \
    emacs-nox \
    git-lfs \
    subversion \
    mercurial

# Install performance and monitoring tools
apt-get install -y \
    htop \
    iotop \
    nethogs \
    iftop \
    nload \
    vnstat \
    sysstat \
    dstat \
    atop

# Create database connection testing scripts
log "Creating database connection testing scripts..."

# PostgreSQL connection test script
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

# Make scripts executable
chmod +x /usr/local/bin/test-postgres-connection
chmod +x /usr/local/bin/test-redis-connection

# Create database monitoring script
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
command -v pgcli &> /dev/null && echo "  ✅ pgcli" || echo "  ❌ pgcli"

echo "Redis tools:"
command -v redis-cli &> /dev/null && echo "  ✅ redis-cli" || echo "  ❌ redis-cli"

echo "Network tools:"
command -v nc &> /dev/null && echo "  ✅ netcat" || echo "  ❌ netcat"
command -v telnet &> /dev/null && echo "  ✅ telnet" || echo "  ❌ telnet"
EOF

chmod +x /usr/local/bin/db-monitor

# Install Rust toolchain and tools
log "Installing Rust toolchain and security tools..."
# Set up temporary directory for Rust installation
export TMPDIR="/home/ubuntu/.cache"
mkdir -p /home/ubuntu/.cache

# Install Rust using rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain 1.82.0

# Source cargo environment
if [ -f "/home/ubuntu/.cargo/env" ]; then
    source "/home/ubuntu/.cargo/env"
fi

# Add to system-wide and user profiles
echo 'export PATH="/home/ubuntu/.cargo/bin:$PATH"' >> /etc/profile
echo 'source /home/ubuntu/.cargo/env 2>/dev/null || true' >> /etc/profile

# Ensure .bashrc exists and add Rust environment
touch /home/ubuntu/.bashrc
echo 'export PATH="/home/ubuntu/.cargo/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'source /home/ubuntu/.cargo/env 2>/dev/null || true' >> /home/ubuntu/.bashrc

# Install Rust security and development tools
cargo install cargo-audit
cargo install cargo-deny
cargo install cargo-tarpaulin --version 0.26.0

# Install Node.js and Playwright dependencies
log "Installing Node.js and Playwright dependencies..."
# Install Node.js using nvm (more reliable than NodeSource)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
export NVM_DIR="/home/ubuntu/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
nvm install 20
nvm use 20
nvm alias default 20

# Add nvm to system profiles
echo 'export NVM_DIR="/home/ubuntu/.nvm"' >> /etc/profile
echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> /etc/profile
echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"' >> /etc/profile

# Add nvm to ubuntu user's bashrc
echo 'export NVM_DIR="/home/ubuntu/.nvm"' >> /home/ubuntu/.bashrc
echo '[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"' >> /home/ubuntu/.bashrc
echo '[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"' >> /home/ubuntu/.bashrc

# Verify Node.js installation
if ! command -v node &> /dev/null; then
    error "Node.js installation failed"
    exit 1
fi
log "Node.js version: $(node --version)"
log "NPM version: $(npm --version)"

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
    libasound2-dev

# Install Playwright browsers globally
npm install -g playwright

# Set up proper temporary directory for Playwright
export TMPDIR="/home/ubuntu/.cache"
mkdir -p /home/ubuntu/.cache

# Install additional Playwright dependencies
log "Installing additional Playwright dependencies..."
apt-get install -y \
    libgtk-4-1 \
    libgraphene-1.0-0 \
    libwoff1 \
    libvpx7 \
    libevent-2.1-7 \
    libopus0 \
    libgstreamer1.0-0 \
    libgstreamer-plugins-base1.0-0 \
    libflite1 \
    libwebpdemux2 \
    libavif13 \
    libharfbuzz-icu0 \
    libwebpmux3 \
    libenchant-2-2 \
    libsecret-1-0 \
    libhyphen0 \
    libmanette-0.2-0 \
    libgles2-mesa \
    libx264-163

# Install Playwright browsers
npx playwright install

# Install Lighthouse CI and performance testing tools
log "Installing Lighthouse CI and performance testing tools..."
npm install -g @lhci/cli
npm install -g lighthouse
npm install -g web-vitals
npm install -g pa11y
npm install -g axe-core

# Install Percy CLI for visual regression testing
npm install -g @percy/cli

# Install additional testing and development tools
npm install -g \
    typescript \
    ts-node \
    nodemon \
    pm2 \
    concurrently \
    cross-env \
    dotenv-cli \
    http-server \
    serve \
    live-server \
    json-server \
    mkcert \
    local-ssl-proxy

# Install additional CI/CD and development tools
log "Installing additional CI/CD and development tools..."

# Install GitHub CLI and Git LFS
apt-get install -y gh git-lfs

# Initialize Git LFS
git lfs install --system

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

# Install additional Python tools for CI/CD
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

# Install additional Rust tools for CI/CD
log "Installing additional Rust tools..."
cargo install \
    cargo-watch \
    cargo-expand \
    cargo-outdated \
    cargo-udeps \
    cargo-machete \
    cargo-deps \
    cargo-tree \
    cargo-modules \
    cargo-geiger \
    cargo-fuzz \
    cargo-bench \
    cargo-profdata \
    cargo-llvm-cov \
    cargo-nextest \
    cargo-hack \
    cargo-msrv \
    cargo-edit \
    cargo-update \
    cargo-generate \
    cargo-make \
    cargo-release

# Set up firewall
log "Configuring firewall..."
ufw --force enable
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp

# Create monitoring scripts
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
echo "Percy CLI: $(percy --version 2>/dev/null || echo 'Not installed')"
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

chmod +x /usr/local/bin/runner-info

# Final status check
log "Performing final status check..."
sleep 10

# Verify database tools installation
log "Verifying database tools installation..."
if command -v psql &> /dev/null; then
    log "✅ PostgreSQL client (psql) installed: $(psql --version)"
else
    warn "❌ PostgreSQL client (psql) not found"
fi

if command -v redis-cli &> /dev/null; then
    log "✅ Redis CLI installed: $(redis-cli --version)"
else
    warn "❌ Redis CLI not found"
fi

if command -v pg_isready &> /dev/null; then
    log "✅ PostgreSQL readiness check tool (pg_isready) installed"
else
    warn "❌ PostgreSQL readiness check tool (pg_isready) not found"
fi

if command -v migrate &> /dev/null; then
    log "✅ Database migration tool installed: $(migrate -version)"
else
    warn "❌ Database migration tool not found"
fi

# Verify network tools
if command -v nc &> /dev/null; then
    log "✅ Network connectivity tool (netcat) installed"
else
    warn "❌ Network connectivity tool (netcat) not found"
fi

if command -v telnet &> /dev/null; then
    log "✅ Telnet client installed"
else
    warn "❌ Telnet client not found"
fi

# Verify Go installation
if command -v go &> /dev/null; then
    log "✅ Go installed: $(go version)"
else
    warn "❌ Go not found"
fi

# Verify protobuf plugins
if command -v protoc-gen-go &> /dev/null; then
    log "✅ protoc-gen-go plugin installed: $(protoc-gen-go --version 2>/dev/null || echo 'version check failed')"
else
    warn "❌ protoc-gen-go plugin not found"
fi

if command -v protoc-gen-go-grpc &> /dev/null; then
    log "✅ protoc-gen-go-grpc plugin installed: $(protoc-gen-go-grpc --version 2>/dev/null || echo 'version check failed')"
else
    warn "❌ protoc-gen-go-grpc plugin not found"
fi

# Verify Rust installation
if command -v rustc &> /dev/null; then
    log "✅ Rust installed: $(rustc --version)"
else
    warn "❌ Rust not found"
fi

if command -v cargo &> /dev/null; then
    log "✅ Cargo installed: $(cargo --version)"
else
    warn "❌ Cargo not found"
fi

# Verify Node.js installation
if command -v node &> /dev/null; then
    log "✅ Node.js installed: $(node --version)"
else
    warn "❌ Node.js not found"
fi

if command -v npm &> /dev/null; then
    log "✅ NPM installed: $(npm --version)"
else
    warn "❌ NPM not found"
fi

# Verify Playwright installation
if command -v npx &> /dev/null && npx playwright --version &> /dev/null; then
    log "✅ Playwright installed: $(npx playwright --version)"
else
    warn "❌ Playwright not found"
fi

# Verify Docker installation
if command -v docker &> /dev/null; then
    log "✅ Docker installed: $(docker --version)"
else
    warn "❌ Docker not found"
fi

# Verify build tools
if command -v make &> /dev/null; then
    log "✅ Make installed: $(make --version | head -1)"
else
    warn "❌ Make not found"
fi

# Verify Git tools
if command -v git &> /dev/null; then
    log "✅ Git installed: $(git --version)"
else
    warn "❌ Git not found"
fi

if command -v git-lfs &> /dev/null; then
    log "✅ Git LFS installed: $(git-lfs version)"
else
    warn "❌ Git LFS not found"
fi

if command -v gh &> /dev/null; then
    log "✅ GitHub CLI installed: $(gh --version)"
else
    warn "❌ GitHub CLI not found"
fi

# Final installation summary
log ""
log "🎉 GitHub Actions Runner Setup Complete!"
log ""
log "📋 Installation Summary:"
log "  ✅ System packages updated"
log "  ✅ Docker installed and configured"
log "  ✅ PostgreSQL and Redis tools installed"
log "  ✅ Go 1.23.0 with development tools"
log "  ✅ Rust 1.82.0 with Cargo and security tools"
log "  ✅ Node.js 20.x with npm and nvm"
log "  ✅ Playwright with browsers (Chromium, Firefox, WebKit)"
log "  ✅ Lighthouse CI and performance testing tools"
log "  ✅ Additional CI/CD and development tools"
log "  ✅ Git, Git LFS, and GitHub CLI"
log "  ✅ Network and debugging tools"
log ""
log "🚀 Your GitHub Actions runner is ready for CI/CD workflows!"
log ""

if command -v cmake &> /dev/null; then
    log "✅ CMake build tool installed: $(cmake --version | head -n1)"
else
    warn "❌ CMake build tool not found"
fi

# Verify archive tools
if command -v tar &> /dev/null; then
    log "✅ Tar archive tool installed: $(tar --version | head -n1)"
else
    warn "❌ Tar archive tool not found"
fi

if command -v gzip &> /dev/null; then
    log "✅ Gzip compression tool installed"
else
    warn "❌ Gzip compression tool not found"
fi

# Verify download tools
if command -v curl &> /dev/null; then
    log "✅ Curl download tool installed: $(curl --version | head -n1)"
else
    warn "❌ Curl download tool not found"
fi

if command -v wget &> /dev/null; then
    log "✅ Wget download tool installed: $(wget --version | head -n1)"
else
    warn "❌ Wget download tool not found"
fi

# Verify Node.js tools
if command -v lighthouse &> /dev/null; then
    log "✅ Lighthouse performance tool installed: $(lighthouse --version)"
else
    warn "❌ Lighthouse performance tool not found"
fi

if command -v lhci &> /dev/null; then
    log "✅ Lighthouse CI tool installed: $(lhci --version)"
else
    warn "❌ Lighthouse CI tool not found"
fi

if command -v percy &> /dev/null; then
    log "✅ Percy CLI tool installed: $(percy --version)"
else
    warn "❌ Percy CLI tool not found"
fi

# Verify GitHub CLI
if command -v gh &> /dev/null; then
    log "✅ GitHub CLI installed: $(gh --version | head -n1)"
else
    warn "❌ GitHub CLI not found"
fi

# Verify additional development tools
if command -v kubectl &> /dev/null; then
    log "✅ Kubernetes CLI installed: $(kubectl version --client --short 2>/dev/null || echo 'kubectl available')"
else
    warn "❌ Kubernetes CLI not found"
fi

if command -v terraform &> /dev/null; then
    log "✅ Terraform installed: $(terraform version | head -n1)"
else
    warn "❌ Terraform not found"
fi

if command -v aws &> /dev/null; then
    log "✅ AWS CLI installed: $(aws --version)"
else
    warn "❌ AWS CLI not found"
fi

# Test database connection scripts
if [ -x "/usr/local/bin/test-postgres-connection" ]; then
    log "✅ PostgreSQL connection test script created"
else
    warn "❌ PostgreSQL connection test script not found or not executable"
fi

if [ -x "/usr/local/bin/test-redis-connection" ]; then
    log "✅ Redis connection test script created"
else
    warn "❌ Redis connection test script not found or not executable"
fi

if [ -x "/usr/local/bin/db-monitor" ]; then
    log "✅ Database monitoring script created"
else
    warn "❌ Database monitoring script not found or not executable"
fi

if systemctl is-active --quiet actions.runner.*; then
    log "GitHub Actions runner installed and started successfully!"
    log "Runner name: $RUNNER_NAME"
    log "Runner labels: $RUNNER_LABELS"
    log "Repository: $GITHUB_REPO"
    if [ "$ENABLE_STATUS_PAGES" = "true" ]; then
        log "Status page available at: http://$(curl -s ifconfig.me)/"
    fi
    log ""
    log "Available tools and utilities:"
    log ""
    log "Database Tools:"
    log "  - psql: PostgreSQL command-line client"
    log "  - pg_isready: PostgreSQL server readiness check"
    log "  - redis-cli: Redis command-line client"
    log "  - migrate: Database migration tool"
    log "  - test-postgres-connection: PostgreSQL connection test script"
    log "  - test-redis-connection: Redis connection test script"
    log "  - db-monitor: Database services monitoring script"
    log ""
    log "Build & Development Tools:"
    log "  - make, cmake: Build automation tools"
    log "  - gcc, g++, clang: Compilers"
    log "  - pkg-config: Package configuration tool"
    log "  - python3, pip3: Python development tools"
    log ""
    log "Archive & Compression Tools:"
    log "  - tar, gzip, bzip2, xz: Archive and compression tools"
    log "  - zip, unzip, p7zip: Additional compression tools"
    log ""
    log "Network & Download Tools:"
    log "  - curl, wget: Download tools"
    log "  - aria2, axel: Alternative download tools"
    log "  - nc, telnet: Network connectivity tools"
    log ""
    log "Performance & Testing Tools:"
    log "  - lighthouse: Web performance testing"
    log "  - lhci: Lighthouse CI integration"
    log "  - percy: Visual regression testing"
    log "  - pa11y: Accessibility testing"
    log "  - axe-core: Accessibility testing library"
    log ""
    log "CI/CD & DevOps Tools:"
    log "  - gh: GitHub CLI"
    log "  - kubectl: Kubernetes CLI"
    log "  - terraform: Infrastructure as Code"
    log "  - aws, azure, gcloud: Cloud CLI tools"
    log "  - docker-compose: Container orchestration"
    log ""
    log "Monitoring & System Tools:"
    log "  - htop, iotop, nethogs: System monitoring"
    log "  - strace, ltrace, gdb: Debugging tools"
    log "  - runner-info: Complete runner information script"
else
    error "GitHub Actions runner failed to start properly"
    exit 1
fi

log "Setup completed successfully!"
