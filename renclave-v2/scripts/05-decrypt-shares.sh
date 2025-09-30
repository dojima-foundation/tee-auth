#!/bin/bash

# Step 5: Member Decryption
# This script decrypts shares using member private keys (simulated for localhost)

set -e

echo "🔓 Step 5: Member Decryption"
echo "============================"

# Check if distribution files exist
if [ ! -f "share_distribution.json" ]; then
    echo "❌ Share distribution file not found: share_distribution.json"
    echo "   Please run Step 4 first: ./scripts/04-distribute-shares.sh"
    exit 1
fi

if [ ! -d "member_keys" ]; then
    echo "❌ Member keys directory not found: member_keys/"
    echo "   Please run Step 4 first: ./scripts/04-distribute-shares.sh"
    exit 1
fi

echo "🔓 Running member decryption tool..."
cd /app
./target/release/member-decryptor

if [ $? -ne 0 ]; then
    echo "❌ Member decryption failed"
    exit 1
fi

echo "✅ Member decryption completed!"

# Display decryption results
echo ""
echo "📊 Member Decryption Results:"
echo "============================="
if [ -f "member_decryption_results.json" ]; then
    cat member_decryption_results.json | jq '{
      total_attempts: .total_attempts,
      successful_decryptions: .successful_decryptions
    }'
    
    echo ""
    echo "🔓 Decryption Details:"
    echo "====================="
    cat member_decryption_results.json | jq -r '.results[] | "\(.member_alias): \(if .success then "✅ Success" else "❌ Failed" end)"'
fi

# Check if share injection request was created
if [ -f "share_injection_request.json" ]; then
    echo ""
    echo "💉 Share Injection Request Created:"
    echo "==================================="
    cat share_injection_request.json | jq '{
      namespace_name: .namespace_name,
      namespace_nonce: .namespace_nonce,
      shares: (.shares | length)
    }'
    
    echo ""
    echo "🔐 Decrypted Shares:"
    echo "==================="
    cat share_injection_request.json | jq -r '.shares[] | "\(.member_alias): \(.decrypted_share | length) bytes"'
fi

echo ""
echo "📁 Files Created:"
echo "  - member_decryption_results.json (decryption details)"
echo "  - share_injection_request.json (ready for TEE injection)"
echo ""
echo "✅ Step 5 Complete: Member decryption successful!"
echo "   Next: Run Step 6 - Share Injection"
echo "   Command: ./scripts/06-inject-shares.sh"
