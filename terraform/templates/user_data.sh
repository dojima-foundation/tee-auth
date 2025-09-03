#!/bin/bash

# GitHub Actions Runner Setup Script for OVH Cloud
# This script installs and configures a GitHub Actions runner

set -e

# Variables
GITHUB_TOKEN="${github_token}"
GITHUB_ORG="dojima-foundation"
GITHUB_REPO="tee-auth"
RUNNER_LABELS="${runner_labels:-ovh,self-hosted,ubuntu-22.04}"
RUNNER_NAME="${runner_name:-runner-2}"
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
    jq

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

# Switch to runner user directory
cd /home/github-runner

# Download and install GitHub Actions runner
log "Downloading GitHub Actions runner..."
RUNNER_VERSION="v2.328.0"
wget -O actions-runner-linux-x64-2.328.0.tar.gz https://github.com/actions/runner/releases/download/${RUNNER_VERSION}/actions-runner-linux-x64-2.328.0.tar.gz

# Extract runner
log "Extracting runner..."
tar xzf ./actions-runner-linux-x64-2.328.0.tar.gz
rm -f actions-runner-linux-x64-2.328.0.tar.gz

# Set proper ownership
chown -R github-runner:github-runner /home/github-runner

# Get runner registration token
log "Getting runner registration token..."
RUNNER_TOKEN=$(curl -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/orgs/dojima-foundation/actions/runners/registration-token | \
  jq -r '.token')

if [ "$RUNNER_TOKEN" = "null" ] || [ -z "$RUNNER_TOKEN" ]; then
    error "Failed to get runner registration token"
    exit 1
fi

log "Successfully obtained runner registration token"

# Configure runner
log "Configuring GitHub Actions runner..."

# Use organization-level runner for dojima-foundation
RUNNER_URL="https://github.com/dojima-foundation"

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
