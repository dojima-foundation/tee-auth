#!/bin/bash

# Step 3: Run Genesis Boot
# This script calls the Genesis Boot API and receives encrypted shares

set -e

echo "🌱 Step 3: Running Genesis Boot"
echo "==============================="

# Check if request file exists
if [ ! -f "genesis_boot_request.json" ]; then
    echo "❌ Genesis Boot request file not found: genesis_boot_request.json"
    echo "   Please run Step 1 first: ./scripts/01-generate-keys.sh"
    exit 1
fi

# Check if services are running
if ! curl -s http://localhost:3000/health > /dev/null 2>&1; then
    echo "❌ Host service is not responding"
    echo "   Please run Step 2 first: ./scripts/02-start-services.sh"
    exit 1
fi

echo "📤 Sending Genesis Boot request..."
echo "   Request file: genesis_boot_request.json"
echo "   API endpoint: http://localhost:3000/enclave/genesis-boot"

# Make the Genesis Boot API call
curl -X POST http://localhost:3000/enclave/genesis-boot \
  -H "Content-Type: application/json" \
  -d @genesis_boot_request.json \
  -o genesis_boot_response.json \
  -w "HTTP Status: %{http_code}\n" \
  -s

if [ $? -ne 0 ]; then
    echo "❌ Genesis Boot API call failed"
    exit 1
fi

# Check if response file was created
if [ ! -f "genesis_boot_response.json" ]; then
    echo "❌ Genesis Boot response file not created"
    exit 1
fi

# Check HTTP status
HTTP_STATUS=$(curl -X POST http://localhost:3000/enclave/genesis-boot \
  -H "Content-Type: application/json" \
  -d @genesis_boot_request.json \
  -w "%{http_code}" \
  -o /dev/null \
  -s)

if [ "$HTTP_STATUS" != "200" ]; then
    echo "❌ Genesis Boot failed with HTTP status: $HTTP_STATUS"
    echo "Response:"
    cat genesis_boot_response.json
    exit 1
fi

echo "✅ Genesis Boot completed successfully!"

# Display the response
echo ""
echo "📊 Genesis Boot Response:"
echo "========================="
cat genesis_boot_response.json | jq '{
  quorum_public_key: (.quorum_public_key | length),
  ephemeral_key: (.ephemeral_key | length),
  waiting_state: .waiting_state,
  encrypted_shares: (.encrypted_shares | length)
}'

echo ""
echo "🔐 Encrypted Shares Details:"
echo "============================"
cat genesis_boot_response.json | jq -r '.encrypted_shares[] | "\(.share_set_member.alias): \(.encrypted_quorum_key_share | length) bytes"'

echo ""
echo "📁 Files Created:"
echo "  - genesis_boot_response.json (Genesis Boot API response with encrypted shares)"
echo ""
echo "✅ Step 3 Complete: Genesis Boot successful!"
echo "   Next: Run Step 4 - External Share Distribution"
echo "   Command: ./scripts/04-distribute-shares.sh"

