#!/bin/bash

# GitHub Actions Runner Setup Script for OVH Cloud
# This script installs and configures a GitHub Actions runner

set -e

# Variables from Terraform
GITHUB_TOKEN="${github_token}"
GITHUB_ORG="${github_org}"
GITHUB_REPO="${github_repo}"
RUNNER_LABELS="${runner_labels}"
RUNNER_NAME="${runner_name}"
DOCKER_REGISTRY_MIRROR="${docker_registry_mirror}"

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
wget -O actions-runner-linux-x64-2.328.0.tar.gz https://github.com/actions/runner/releases/download/v2.328.0/actions-runner-linux-x64-2.328.0.tar.gz

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

# Install nginx for status page
apt-get install -y nginx

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

# Final setup
log "Setting up final configurations..."

# Set proper permissions
chown -R github-runner:github-runner /home/github-runner

# Create a simple monitoring script
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

# Create system information script
cat > /usr/local/bin/runner-info << 'EOF'
#!/bin/bash

echo "=== System Information ==="
echo "Hostname: $(hostname)"
echo "OS: $(lsb_release -d | cut -f2)"
echo "Kernel: $(uname -r)"
echo "Architecture: $(uname -m)"
echo "CPU: $(nproc) cores"
echo "Memory: $(free -h | awk '/^Mem:/{print $2}')"
echo "Disk: $(df -h / | awk 'NR==2{print $4}') available"
echo "Uptime: $(uptime -p)"
echo "========================="
EOF

chmod +x /usr/local/bin/runner-info

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
echo "======================"
