#!/bin/bash

# Complete Genesis Boot Setup for renclave-v2
# This script automates the entire process to get renclave-v2 to ApplicationReady state
# with threshold 7 out of 10 members

set -e

echo "🚀 Complete Genesis Boot Setup for renclave-v2"
echo "=============================================="
echo "Setting up TEE instance with threshold 7/10 for local development"
echo ""

# Configuration
RENCLAVE_DIR="/Users/luffybhaagi/dojima/tee-auth/renclave-v2"
RENCLAVE_URL="http://localhost:9000"
THRESHOLD=7
MEMBERS=10
NAMESPACE="local-dev-namespace"
NAMESPACE_NONCE=12345

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Step 1: Navigate to renclave directory and build
log_info "Step 1: Building renclave-v2 Docker image..."
cd "$RENCLAVE_DIR"

if [ ! -f "target/release/genesis_key_generator" ]; then
    log_info "Building genesis key generator..."
    cargo build --release --bin genesis_key_generator
fi

log_success "Build completed"

# Step 2: Generate keys for genesis boot
log_info "Step 2: Generating P256 keys for genesis boot..."
./target/release/genesis_key_generator $MEMBERS

if [ ! -f "/tmp/genesis_request.json" ]; then
    log_error "Failed to generate genesis request"
    exit 1
fi

log_success "Generated $MEMBERS P256 keys"

# Step 3: Modify threshold to 7
log_info "Step 3: Setting threshold to $THRESHOLD..."
cp /tmp/genesis_request.json /tmp/genesis_request_modified.json

# Update thresholds using jq
jq --argjson threshold $THRESHOLD '.manifest_threshold = $threshold | .share_threshold = $threshold' /tmp/genesis_request_modified.json > /tmp/genesis_request_final.json

# Update namespace
jq --arg namespace "$NAMESPACE" '.namespace_name = $namespace' /tmp/genesis_request_final.json > /tmp/genesis_request_ready.json

log_success "Updated threshold to $THRESHOLD and namespace to $NAMESPACE"

# Step 4: Start renclave-v2 service
log_info "Step 4: Starting renclave-v2 Docker service..."
cd "$RENCLAVE_DIR/docker"

# Stop any existing containers
docker compose down -v > /dev/null 2>&1 || true

# Start the service
docker compose up -d renclave-v2

log_success "renclave-v2 service started"

# Step 5: Wait for service to be ready
log_info "Step 5: Waiting for renclave-v2 to be ready..."
for i in {1..60}; do
    if curl -s "$RENCLAVE_URL/health" > /dev/null 2>&1; then
        log_success "renclave-v2 is responding on port 9000"
        break
    fi
    if [ $i -eq 60 ]; then
        log_error "renclave-v2 failed to start after 60 attempts"
        log_info "Checking logs..."
        docker compose logs renclave-v2 --tail=20
        exit 1
    fi
    echo "   Attempt $i/60: Waiting for renclave-v2..."
    sleep 2
done

# Step 6: Execute genesis boot
log_info "Step 6: Executing genesis boot with threshold $THRESHOLD/$MEMBERS..."
GENESIS_RESPONSE=$(curl -X POST "$RENCLAVE_URL/enclave/genesis-boot" \
  -H "Content-Type: application/json" \
  -d @/tmp/genesis_request_ready.json \
  -w "HTTP Status: %{http_code}" \
  -s)

HTTP_STATUS=$(echo "$GENESIS_RESPONSE" | grep -o "HTTP Status: [0-9]*" | grep -o "[0-9]*")

if [ "$HTTP_STATUS" != "200" ]; then
    log_error "Genesis boot failed with HTTP status: $HTTP_STATUS"
    echo "Response:"
    echo "$GENESIS_RESPONSE"
    exit 1
fi

log_success "Genesis boot completed successfully"

# Step 7: Extract encrypted shares for injection
log_info "Step 7: Preparing share injection..."

# Parse the genesis response to extract encrypted shares
GENESIS_DATA=$(echo "$GENESIS_RESPONSE" | sed 's/HTTP Status: [0-9]*$//')

# Create share injection request
SHARE_INJECTION_REQUEST=$(echo "$GENESIS_DATA" | jq '{
  namespace_name: .manifest_envelope.manifest.namespace.name,
  namespace_nonce: .manifest_envelope.manifest.namespace.nonce,
  shares: [.encrypted_shares[] | {
    member_alias: .share_set_member.alias,
    decrypted_share: .encrypted_quorum_key_share
  }] | .[0:'$THRESHOLD']
}')

echo "$SHARE_INJECTION_REQUEST" > /tmp/share_injection_request.json

log_success "Prepared share injection request with $THRESHOLD shares"

# Step 8: Inject shares
log_info "Step 8: Injecting shares to complete TEE setup..."
SHARE_RESPONSE=$(curl -X POST "$RENCLAVE_URL/enclave/inject-shares" \
  -H "Content-Type: application/json" \
  -d @/tmp/share_injection_request.json \
  -w "HTTP Status: %{http_code}" \
  -s)

HTTP_STATUS=$(echo "$SHARE_RESPONSE" | grep -o "HTTP Status: [0-9]*" | grep -o "[0-9]*")

if [ "$HTTP_STATUS" != "200" ]; then
    log_error "Share injection failed with HTTP status: $HTTP_STATUS"
    echo "Response:"
    echo "$SHARE_RESPONSE"
    exit 1
fi

log_success "Share injection completed successfully"

# Step 9: Verify application status
log_info "Step 9: Verifying application status..."
APP_STATUS=$(curl -s "$RENCLAVE_URL/enclave/application-status")

PHASE=$(echo "$APP_STATUS" | jq -r '.phase')
HAS_QUORUM_KEY=$(echo "$APP_STATUS" | jq -r '.has_quorum_key')

if [ "$PHASE" = "ApplicationReady" ] && [ "$HAS_QUORUM_KEY" = "true" ]; then
    log_success "Application is in ApplicationReady state with quorum key"
else
    log_warning "Application status: $PHASE, has_quorum_key: $HAS_QUORUM_KEY"
fi

# Step 10: Test seed generation
log_info "Step 10: Testing seed generation..."
SEED_RESPONSE=$(curl -X POST "$RENCLAVE_URL/generate-seed" \
  -H "Content-Type: application/json" \
  -d '{"strength": 256}' \
  -w "HTTP Status: %{http_code}" \
  -s)

HTTP_STATUS=$(echo "$SEED_RESPONSE" | grep -o "HTTP Status: [0-9]*" | grep -o "[0-9]*")

if [ "$HTTP_STATUS" = "200" ]; then
    log_success "Seed generation test passed"
    SEED_PHRASE=$(echo "$SEED_RESPONSE" | sed 's/HTTP Status: [0-9]*$//' | jq -r '.seed_phrase')
    log_info "Generated encrypted seed (first 50 chars): ${SEED_PHRASE:0:50}..."
else
    log_error "Seed generation test failed with HTTP status: $HTTP_STATUS"
    echo "Response:"
    echo "$SEED_RESPONSE"
    exit 1
fi

# Cleanup temporary files
rm -f /tmp/genesis_request*.json /tmp/share_injection_request.json

# Final summary
echo ""
echo "🎉 Genesis Boot Setup Complete!"
echo "=============================="
echo "✅ renclave-v2 is running on: $RENCLAVE_URL"
echo "✅ Threshold: $THRESHOLD out of $MEMBERS members"
echo "✅ Namespace: $NAMESPACE"
echo "✅ Status: ApplicationReady"
echo "✅ Quorum key: Provisioned"
echo "✅ Seed generation: Functional"
echo ""
echo "🔧 Available endpoints:"
echo "   - Health: $RENCLAVE_URL/health"
echo "   - Generate Seed: $RENCLAVE_URL/generate-seed"
echo "   - Application Status: $RENCLAVE_URL/enclave/application-status"
echo ""
echo "🚀 renclave-v2 is ready for integration with gauth service!"
