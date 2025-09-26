#!/bin/bash

# GitHub Actions Runner Setup Script for OVH Cloud
# This script installs and configures a GitHub Actions runner

set -e

# Variables
GITHUB_TOKEN="${github_token}"
GITHUB_ORG="${github_org}"
GITHUB_REPO="${github_repo}"
RUNNER_LABELS="${runner_labels}"
RUNNER_NAME="${runner_name}"
DOCKER_REGISTRY_MIRROR="${docker_registry_mirror}"

# Colors for output
RED='$${RED}'
GREEN='$${GREEN}'
YELLOW='$${YELLOW}'
NC='$${NC}' # No Color

# Set actual color values
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "$${GREEN}[$$(date +'%Y-%m-%d %H:%M:%S')] $$1$${NC}"
}

warn() {
    echo -e "$${YELLOW}[$$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $$1$${NC}"
}

error() {
    echo -e "$${RED}[$$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $$1$${NC}"
}

# Update system
log "Updating system packages..."
apt-get update
apt-get upgrade -y

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
    postgresql-client \
    netcat-openbsd \
    sudo \
    protobuf-compiler

# Install Docker
log "Installing Docker..."
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Configure Docker registry mirror if provided
if [ ! -z "$DOCKER_REGISTRY_MIRROR" ]; then
    log "Configuring Docker registry mirror: $DOCKER_REGISTRY_MIRROR"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << EOF
{
    "registry-mirrors": ["$DOCKER_REGISTRY_MIRROR"]
}
EOF
    systemctl restart docker
fi

# Create runner user
log "Creating runner user..."
useradd -m -s /bin/bash github-runner
usermod -aG docker github-runner
usermod -aG sudo github-runner

# Configure passwordless sudo for github-runner user (for CI workflows)
log "Configuring sudo access for github-runner..."
echo "github-runner ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/github-runner
chmod 0440 /etc/sudoers.d/github-runner

# Switch to runner user directory
cd /home/github-runner

# Download and install GitHub Actions runner
log "Downloading GitHub Actions runner..."
RUNNER_VERSION="v2.328.0"
wget -O actions-runner-linux-x64-2.328.0.tar.gz https://github.com/actions/runner/releases/download/$${RUNNER_VERSION}/actions-runner-linux-x64-2.328.0.tar.gz

# Extract runner
log "Extracting runner..."
tar xzf ./actions-runner-linux-x64-2.328.0.tar.gz
rm -f actions-runner-linux-x64-2.328.0.tar.gz

# Set proper ownership
chown -R github-runner:github-runner /home/github-runner

# Get runner registration token
log "Getting runner registration token..."

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
    error "Failed to get runner registration token for $RUNNER_URL"
    exit 1
fi

log "Successfully obtained runner registration token for $RUNNER_URL"

# Configure runner
log "Configuring GitHub Actions runner..."

# Configure the runner
sudo -u github-runner ./config.sh \
    --url "$RUNNER_URL" \
    --token "$RUNNER_TOKEN" \
    --name "$RUNNER_NAME" \
    --labels "$RUNNER_LABELS" \
    --unattended \
    --replace

# Copy runsvc.sh to home directory
sudo -u github-runner cp bin/runsvc.sh .
sudo chown github-runner:github-runner runsvc.sh
sudo chmod +x runsvc.sh

# Create systemd service manually
log "Creating systemd service..."
sudo tee /etc/systemd/system/github-runner.service << 'EOF'
[Unit]
Description=GitHub Actions Runner
After=network.target

[Service]
ExecStart=/home/github-runner/runsvc.sh
User=github-runner
WorkingDirectory=/home/github-runner
KillMode=process
KillSignal=SIGTERM
TimeoutStopSec=5min

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
log "Starting GitHub Actions runner service..."
systemctl daemon-reload
systemctl enable github-runner
systemctl start github-runner

# Set up monitoring and health checks
log "Setting up monitoring..."

# Create health check script
cat > /home/github-runner/health_check.sh << 'EOF'
#!/bin/bash

# Check if runner service is running
if ! systemctl is-active --quiet github-runner; then
    echo "Runner service is not running. Restarting..."
    systemctl restart github-runner
fi

# Check runner status
if [ -f "/home/github-runner/.runner" ]; then
    RUNNER_STATUS=$(cat /home/github-runner/.runner | jq -r '.status' 2>/dev/null || echo "unknown")
    if [ "$RUNNER_STATUS" != "online" ]; then
        echo "Runner is not online. Current status: $RUNNER_STATUS"
        # Attempt to restart the service
        systemctl restart github-runner
    fi
fi
EOF

chmod +x /home/github-runner/health_check.sh

# Set up cron job for health checks
echo "*/5 * * * * /home/github-runner/health_check.sh >> /var/log/github-runner-health.log 2>&1" | crontab -

# Create log rotation configuration
cat > /etc/logrotate.d/github-runner << EOF
/var/log/github-runner-health.log {
    daily
    missingok
    rotate 7
    compress
    notifempty
    create 644 root root
}
EOF

# Create a simple status page
cat > /var/www/html/runner-status.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>GitHub Actions Runner Status</title>
    <meta http-equiv="refresh" content="30">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .online { background-color: #d4edda; color: #155724; }
        .offline { background-color: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <h1>GitHub Actions Runner Status</h1>
    <div id="status"></div>
    <script>
        fetch('/api/status')
            .then(response => response.json())
            .then(data => {
                const statusDiv = document.getElementById('status');
                statusDiv.innerHTML = '<div class="status ' + (data.online ? 'online' : 'offline') + '">' +
                    'Runner: ' + (data.online ? 'Online' : 'Offline') + '<br>' +
                    'Last Check: ' + new Date().toLocaleString() + '</div>';
            })
            .catch(error => {
                document.getElementById('status').innerHTML = '<div class="status offline">Error checking status</div>';
            });
    </script>
</body>
</html>
EOF

# Install nginx for status page
apt-get install -y nginx

# Configure nginx
cat > /etc/nginx/sites-available/runner-status << EOF
server {
    listen 80;
    server_name _;
    
    location / {
        root /var/www/html;
        index runner-status.html;
    }
    
    location /api/status {
        default_type application/json;
        return 200 '{"online": true, "timestamp": "$(date -Iseconds)"}';
    }
}
EOF

ln -sf /etc/nginx/sites-available/runner-status /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
systemctl restart nginx

# Install Go development tools
log "Installing Go development tools..."
# Install Go (required for protoc plugins and other tools)
wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
# Export Go path
export PATH="/usr/local/go/bin:$PATH"
echo 'export PATH="/usr/local/go/bin:$PATH"' >> /etc/profile
# Ensure .bashrc exists and add Go environment
touch /home/github-runner/.bashrc
echo 'export PATH="/usr/local/go/bin:$PATH"' >> /home/github-runner/.bashrc
echo 'export GOPATH="$HOME/go"' >> /home/github-runner/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/github-runner/.bashrc

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

# Install Rust toolchain and tools
log "Installing Rust toolchain and security tools..."
# Install Rust using rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain 1.82.0
# Source cargo environment - use full path to avoid expansion issues
if [ -f "$HOME/.cargo/env" ]; then
    source "$HOME/.cargo/env"
fi
# Add to system-wide and user profiles
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /etc/profile
echo 'source $HOME/.cargo/env 2>/dev/null || true' >> /etc/profile
# Ensure .bashrc exists and add Rust environment
touch /home/github-runner/.bashrc
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> /home/github-runner/.bashrc
echo 'source $HOME/.cargo/env 2>/dev/null || true' >> /home/github-runner/.bashrc

# Install Rust security and development tools
cargo install cargo-audit
cargo install cargo-deny
cargo install cargo-tarpaulin --version 0.27.0 --locked

# Install Node.js and Playwright dependencies
log "Installing Node.js and Playwright dependencies..."
# Install Node.js 20.x using NodeSource repository
curl -fsSL https://deb.nodesource.com/setup_20.x | bash - || {
    log "NodeSource setup failed, trying alternative installation..."
    # Fallback: install Node.js using Ubuntu repository
    apt-get update
    apt-get install -y nodejs npm
}
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
npx playwright install --with-deps

# Set up firewall
log "Configuring firewall..."
ufw --force enable
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp

# Create cleanup script for graceful shutdown
cat > /home/github-runner/cleanup.sh << 'EOF'
#!/bin/bash

# Graceful shutdown script for GitHub Actions runner
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log "Starting graceful shutdown..."

# Stop the runner service
if systemctl is-active --quiet github-runner; then
    log "Stopping GitHub Actions runner..."
    cd /home/github-runner
    systemctl stop github-runner
    systemctl disable github-runner
fi

# Remove runner from GitHub
if [ -f "/home/github-runner/.runner" ]; then
    log "Removing runner from GitHub..."
    cd /home/github-runner
    ./config.sh remove --unattended --token "$GITHUB_TOKEN" || true
fi

log "Cleanup completed"
EOF

chmod +x /home/github-runner/cleanup.sh

# Set up systemd service for graceful shutdown
cat > /etc/systemd/system/github-runner-cleanup.service << EOF
[Unit]
Description=Cleanup GitHub Actions Runner on shutdown
DefaultDependencies=no
Before=shutdown.target reboot.target halt.target

[Service]
Type=oneshot
ExecStart=/home/github-runner/cleanup.sh
User=github-runner
Environment=GITHUB_TOKEN=$GITHUB_TOKEN
TimeoutStartSec=0

[Install]
WantedBy=shutdown.target
EOF

systemctl enable github-runner-cleanup.service

# Final setup
log "Setting up final configurations..."

# Set proper permissions
chown -R github-runner:github-runner /home/github-runner

# Create a comprehensive monitoring script
cat > /usr/local/bin/runner-monitor << 'EOF'
#!/bin/bash

echo "=== GitHub Actions Runner Status ==="
echo "Service Status: $(systemctl is-active github-runner)"
echo "Service Enabled: $(systemctl is-enabled github-runner)"
echo "Runner Process: $(ps aux | grep -v grep | grep -c 'run.sh')"
echo "Docker Status: $(systemctl is-active docker)"
echo "Last Health Check: $(tail -1 /var/log/github-runner-health.log 2>/dev/null || echo 'No health check log')"
echo "================================"
EOF

chmod +x /usr/local/bin/runner-monitor

# Create runner reconfiguration script
cat > /usr/local/bin/runner-reconfigure << 'EOF'
#!/bin/bash

# GitHub Actions Runner Reconfiguration Script
# Usage: runner-reconfigure <github_token> <target_type> <target_name> [labels]

set -e

GITHUB_TOKEN="$1"
TARGET_TYPE="$2"  # "org" or "repo"
TARGET_NAME="$3"  # organization or repository name
LABELS="${4:-ovh,self-hosted,ubuntu-22.04}"

if [ -z "$GITHUB_TOKEN" ] || [ -z "$TARGET_TYPE" ] || [ -z "$TARGET_NAME" ]; then
    echo "Usage: $0 <github_token> <target_type> <target_name> [labels]"
    echo "  target_type: 'org' for organization or 'repo' for repository"
    echo "  target_name: organization name (e.g., 'dojima-foundation') or repo name (e.g., 'user/repo')"
    echo "  labels: comma-separated labels (default: ovh,self-hosted,ubuntu-22.04)"
    echo ""
    echo "Examples:"
    echo "  $0 <token> org dojima-foundation"
    echo "  $0 <token> org dojimanetwork"
    echo "  $0 <token> repo bhaagiKenpachi/spark-park-cricket"
    echo "  $0 <token> repo dojima-foundation/tee-auth"
    exit 1
fi

echo "Reconfiguring runner for $TARGET_TYPE: $TARGET_NAME"

# Get registration token
if [ "$TARGET_TYPE" = "org" ]; then
    RUNNER_URL="https://github.com/$TARGET_NAME"
    RUNNER_TOKEN=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/orgs/$TARGET_NAME/actions/runners/registration-token" | \
        jq -r '.token')
else
    RUNNER_URL="https://github.com/$TARGET_NAME"
    RUNNER_TOKEN=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$TARGET_NAME/actions/runners/registration-token" | \
        jq -r '.token')
fi

if [ "$RUNNER_TOKEN" = "null" ] || [ -z "$RUNNER_TOKEN" ]; then
    echo "Failed to get registration token for $RUNNER_URL"
    echo "Please check:"
    echo "  1. GitHub token has correct permissions"
    echo "  2. Target organization/repository exists"
    echo "  3. Token has access to the target"
    exit 1
fi

# Stop current service
systemctl stop github-runner || true

# Remove current configuration
cd /home/github-runner
./config.sh remove --unattended --token "$GITHUB_TOKEN" || true

# Configure for new target
./config.sh \
    --url "$RUNNER_URL" \
    --token "$RUNNER_TOKEN" \
    --name "$(hostname)-$(date +%s)" \
    --labels "$LABELS" \
    --unattended \
    --replace

# Start service
systemctl start github-runner

echo "Runner reconfigured successfully for $RUNNER_URL"
echo "Labels: $LABELS"
echo "Service status: $(systemctl is-active github-runner)"
EOF

chmod +x /usr/local/bin/runner-reconfigure

# Create multi-organization setup script
cat > /usr/local/bin/setup-multi-org-runners << 'EOF'
#!/bin/bash

# Multi-Organization GitHub Actions Runner Setup Script
# This script helps configure runners for multiple organizations and repositories

set -e

GITHUB_TOKEN="${1:-$GITHUB_TOKEN}"

if [ -z "$GITHUB_TOKEN" ]; then
    echo "Usage: $0 <github_token>"
    echo "  github_token: GitHub Personal Access Token with appropriate permissions"
    echo ""
    echo "Required token permissions:"
    echo "  - admin:org (for organization-level runners)"
    echo "  - repo (for repository-level runners)"
    echo "  - actions:write (for runner management)"
    exit 1
fi

echo "=== Multi-Organization Runner Setup ==="
echo "Setting up runners for multiple organizations and repositories"
echo ""

# Function to test GitHub API access
test_github_access() {
    local target="$1"
    local target_type="$2"
    
    echo "Testing access to $target_type: $target"
    
    if [ "$target_type" = "org" ]; then
        response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
            -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/orgs/$target")
    else
        response=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
            -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/repos/$target")
    fi
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        echo "❌ Access denied or not found: $target"
        return 1
    else
        echo "✅ Access confirmed: $target"
        return 0
    fi
}

# Function to get registration token
get_registration_token() {
    local target="$1"
    local target_type="$2"
    
    if [ "$target_type" = "org" ]; then
        curl -s -X POST \
            -H "Authorization: token $GITHUB_TOKEN" \
            -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/orgs/$target/actions/runners/registration-token" | \
            jq -r '.token'
    else
        curl -s -X POST \
            -H "Authorization: token $GITHUB_TOKEN" \
            -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/repos/$target/actions/runners/registration-token" | \
            jq -r '.token'
    fi
}

# Test access to supported organizations and repositories
echo "=== Testing GitHub API Access ==="

# Test organizations
test_github_access "dojima-foundation" "org"
test_github_access "dojimanetwork" "org"

# Test user repositories
test_github_access "bhaagiKenpachi" "org" || echo "Note: bhaagiKenpachi is a user account, not an organization"

# Test specific repositories
test_github_access "bhaagiKenpachi/spark-park-cricket" "repo"
test_github_access "dojima-foundation/tee-auth" "repo"

echo ""
echo "=== Available Configuration Options ==="
echo ""
echo "1. Organization-level runners (recommended for multiple repos):"
echo "   - dojima-foundation: All repos in the organization can use the runner"
echo "   - dojimanetwork: All repos in the organization can use the runner"
echo ""
echo "2. Repository-level runners (for specific repos):"
echo "   - bhaagiKenpachi/spark-park-cricket: Dedicated runner for this repo"
echo "   - dojima-foundation/tee-auth: Dedicated runner for this repo"
echo ""
echo "=== Configuration Commands ==="
echo ""
echo "To configure for organization-level access:"
echo "  runner-reconfigure $GITHUB_TOKEN org dojima-foundation 'ovh,self-hosted,ubuntu-22.04,dojima-foundation'"
echo "  runner-reconfigure $GITHUB_TOKEN org dojimanetwork 'ovh,self-hosted,ubuntu-22.04,dojimanetwork'"
echo ""
echo "To configure for repository-level access:"
echo "  runner-reconfigure $GITHUB_TOKEN repo bhaagiKenpachi/spark-park-cricket 'ovh,self-hosted,ubuntu-22.04,bhaagiKenpachi'"
echo "  runner-reconfigure $GITHUB_TOKEN repo dojima-foundation/tee-auth 'ovh,self-hosted,ubuntu-22.04,dojima-foundation'"
echo ""
echo "=== Workflow Configuration ==="
echo ""
echo "For organization-level runners, use these labels in your workflows:"
echo "  runs-on: [self-hosted, ovh, ubuntu-22.04, dojima-foundation]"
echo "  runs-on: [self-hosted, ovh, ubuntu-22.04, dojimanetwork]"
echo ""
echo "For repository-level runners, use these labels:"
echo "  runs-on: [self-hosted, ovh, ubuntu-22.04, bhaagiKenpachi]"
echo "  runs-on: [self-hosted, ovh, ubuntu-22.04, dojima-foundation]"
echo ""
echo "=== Current Runner Status ==="
echo "Service Status: $(systemctl is-active github-runner)"
echo "Runner Configuration: $(cat /home/github-runner/.runner 2>/dev/null | jq -r '.url // "Not configured"' 2>/dev/null || echo "Not configured")"
echo ""
echo "=== Next Steps ==="
echo "1. Choose your configuration approach (organization vs repository level)"
echo "2. Run the appropriate runner-reconfigure command"
echo "3. Update your workflow files to use the correct labels"
echo "4. Test the setup with a simple workflow"
EOF

chmod +x /usr/local/bin/setup-multi-org-runners

# Create comprehensive system information script
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

chmod +x /usr/local/bin/runner-info

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

# Make database scripts executable
chmod +x /usr/local/bin/test-postgres-connection
chmod +x /usr/local/bin/test-redis-connection
chmod +x /usr/local/bin/db-monitor

log "GitHub Actions runner setup completed successfully!"
log "Runner name: $RUNNER_NAME"
log "Runner URL: $RUNNER_URL"
log "Labels: $RUNNER_LABELS"

# Display final status
echo "=== Setup Complete ==="
echo "Runner Name: $RUNNER_NAME"
echo "Runner URL: $RUNNER_URL"
echo "Service Status: $(systemctl is-active github-runner)"
echo "Health Check: Every 5 minutes"
echo "Status Page: http://$(curl -s ifconfig.me)/runner-status.html"
echo ""
echo "=== Available Utility Scripts ==="
echo "runner-monitor: Check runner status and health"
echo "runner-info: Display comprehensive system information"
echo "runner-reconfigure: Reconfigure runner for different org/repo"
echo "setup-multi-org-runners: Setup runners for multiple organizations"
echo "test-postgres-connection: Test PostgreSQL connectivity"
echo "test-redis-connection: Test Redis connectivity"
echo "db-monitor: Monitor database services status"
echo ""
echo "=== Usage Examples ==="
echo "Check runner status: runner-monitor"
echo "View system info: runner-info"
echo "Setup multi-org support: setup-multi-org-runners <github_token>"
echo "Reconfigure for org: runner-reconfigure <token> org dojima-foundation"
echo "Reconfigure for repo: runner-reconfigure <token> repo bhaagiKenpachi/spark-park-cricket"
echo "Test PostgreSQL: test-postgres-connection [host] [port] [user] [password] [database]"
echo "Test Redis: test-redis-connection [host] [port] [password]"
echo "Monitor databases: db-monitor"
echo ""
echo "=== Multi-Organization Support ==="
echo "This runner supports the following organizations and repositories:"
echo "  - dojima-foundation (organization)"
echo "  - dojimanetwork (organization)"
echo "  - bhaagiKenpachi (user account with repositories)"
echo ""
echo "Use 'setup-multi-org-runners <token>' to configure for multiple targets"
echo "======================"
