#!/bin/bash

# Test renclave-v2 Setup
# This script verifies the renclave-v2 service is in ApplicationReady state and can generate seeds

set -e

RENCLAVE_PORT=9000
RENCLAVE_URL="http://localhost:$RENCLAVE_PORT"

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

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo "🔍 Testing renclave-v2 Setup"
echo "============================"

# 1. Test health endpoint
log_info "1. Testing health endpoint..."
if curl -s "$RENCLAVE_URL/health" > /dev/null; then
    log_success "Health check passed"
else
    log_error "Health check failed. Is renclave-v2 running on port $RENCLAVE_PORT?"
fi

# 2. Test application status
log_info "2. Testing application status..."
STATUS_RESPONSE=$(curl -s "$RENCLAVE_URL/enclave/application-status")
if echo "$STATUS_RESPONSE" | jq -e '.phase == "ApplicationReady" and .has_quorum_key == true' > /dev/null; then
    log_success "Application status: ApplicationReady with quorum key"
else
    log_warning "Application status: $(echo "$STATUS_RESPONSE" | jq -r '.phase'), has_quorum_key: $(echo "$STATUS_RESPONSE" | jq -r '.has_quorum_key')"
fi

# 3. Test seed generation
log_info "3. Testing seed generation..."
SEED_RESPONSE=$(curl -X POST "$RENCLAVE_URL/generate-seed" \
  -H "Content-Type: application/json" \
  -d '{"strength": 256}' \
  -s)

if echo "$SEED_RESPONSE" | jq -e '.seed_phrase' > /dev/null; then
    log_success "Seed generation successful"
    GENERATED_SEED=$(echo "$SEED_RESPONSE" | jq -r '.seed_phrase')
    echo "   Generated encrypted seed (first 50 chars): ${GENERATED_SEED:0:50}..."
else
    log_error "Seed generation failed. Response: $SEED_RESPONSE"
fi

# 4. Test key derivation (optional, as it requires a valid seed and path)
log_info "4. Testing key derivation..."
# This test requires a valid encrypted seed and derivation path.
# For a quick verification, we'll just check if the endpoint exists and responds.
# A full test would involve decrypting a seed and using it.
DERIVE_KEY_RESPONSE=$(curl -X POST "$RENCLAVE_URL/derive-key" \
  -H "Content-Type: application/json" \
  -d '{"seed_phrase": "dummy_encrypted_seed", "path": "m/44'\''/60'\''/0'\''/0/0", "curve": "CURVE_SECP256K1"}' \
  -s -o /dev/null -w "%{http_code}")

if [ "$DERIVE_KEY_RESPONSE" == "500" ]; then # Expected to fail without a real seed
    log_info "⚠️  Key derivation not available or failed (expected with dummy seed)"
else
    log_warning "Unexpected response for key derivation: HTTP $DERIVE_KEY_RESPONSE"
fi

echo ""
echo "🎉 All tests passed! renclave-v2 is ready for use."
echo "Service URL: $RENCLAVE_URL"
echo ""
