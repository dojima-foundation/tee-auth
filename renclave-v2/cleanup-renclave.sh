#!/bin/bash

# Cleanup Script for renclave-v2 Docker services and temporary files

set -e

RENCLAVE_V2_DIR="/Users/luffybhaagi/dojima/tee-auth/renclave-v2"
GENESIS_REQUEST_FILE="/tmp/genesis_request.json"
SHARE_INJECTION_REQUEST_FILE="/tmp/share_injection_request.json"
GENESIS_BOOT_RESPONSE_FILE="/tmp/genesis_boot_response.json"
SHARE_INJECTION_RESPONSE_FILE="/tmp/share_injection_response.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo "🧹 Cleaning up renclave-v2 Docker services and temporary files"
echo "============================================================"

# Stop and remove Docker containers and networks
log_info "Stopping and removing renclave-v2 Docker containers and networks..."
cd "$RENCLAVE_V2_DIR/docker"
docker compose down -v --remove-orphans || log_error "Failed to stop/remove renclave-v2 services."
log_success "renclave-v2 Docker services and network removed."

# Remove temporary files
log_info "Removing temporary files..."
rm -f "$GENESIS_REQUEST_FILE"
rm -f "$SHARE_INJECTION_REQUEST_FILE"
rm -f "$GENESIS_BOOT_RESPONSE_FILE"
rm -f "$SHARE_INJECTION_RESPONSE_FILE"
rm -f /tmp/genesis_request_*.json
log_success "Temporary files removed."

# Clean up any remaining Docker resources
log_info "Cleaning up Docker system resources..."
docker system prune -f > /dev/null 2>&1 || true
log_success "Docker system cleanup completed."

echo ""
echo "🎉 Cleanup Complete!"
echo "===================="
echo "The renclave-v2 environment is now clean."
echo ""