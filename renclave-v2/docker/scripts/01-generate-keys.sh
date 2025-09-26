#!/bin/bash

# Step 1: Generate Keys for Genesis Boot
# This script generates 3 P256 key pairs and creates the Genesis Boot request

set -e

echo "🔑 Step 1: Generating Keys for Genesis Boot"
echo "==========================================="

# Clean up any existing files
rm -f genesis_boot_request.json
rm -f member_keys.json
rm -rf member_keys/

# Generate the Genesis Boot request with 3 member keys
echo "📋 Generating Genesis Boot request with 3 member keys..."
cd /app
./target/release/genesis-key-generator

if [ $? -ne 0 ]; then
    echo "❌ Failed to generate Genesis Boot request"
    exit 1
fi

echo "✅ Genesis Boot request generated: genesis_boot_request.json"

# Display the generated request
echo ""
echo "📊 Generated Genesis Boot Request:"
echo "=================================="
cat genesis_boot_request.json | jq '{
  namespace_name: .namespace_name,
  namespace_nonce: .namespace_nonce,
  manifest_members: .manifest_members | length,
  manifest_threshold: .manifest_threshold,
  share_members: .share_members | length,
  share_threshold: .share_threshold,
  pivot_hash: .pivot_hash,
  pivot_args: .pivot_args
}'

echo ""
echo "🔑 Member Public Keys Generated:"
echo "================================"
cat genesis_boot_request.json | jq -r '.share_members[] | "\(.alias): \(.pub_key | length) bytes"'

echo ""
echo "📁 Files Created:"
echo "  - genesis_boot_request.json (Genesis Boot API request)"
echo ""
echo "✅ Step 1 Complete: Keys generated successfully!"
echo "   Next: Run Step 2 - Start TEE and Host"
