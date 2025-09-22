#!/bin/bash

# Step 4: External Share Distribution
# This script distributes encrypted shares to members (simulated for localhost)

set -e

echo "📤 Step 4: External Share Distribution"
echo "======================================"

# Check if Genesis Boot response exists
if [ ! -f "genesis_boot_response.json" ]; then
    echo "❌ Genesis Boot response file not found: genesis_boot_response.json"
    echo "   Please run Step 3 first: ./scripts/03-genesis-boot.sh"
    exit 1
fi

echo "🔄 Running share distribution tool..."
cd /app
./target/release/share-distributor

if [ $? -ne 0 ]; then
    echo "❌ Share distribution failed"
    exit 1
fi

echo "✅ Share distribution completed!"

# Display distribution results
echo ""
echo "📊 Share Distribution Results:"
echo "=============================="
if [ -f "share_distribution.json" ]; then
    cat share_distribution.json | jq '{
      total_members: .total_members,
      threshold: .threshold,
      distributed_shares: (.distributed_shares | length)
    }'
    
    echo ""
    echo "👥 Distributed Shares:"
    echo "====================="
    cat share_distribution.json | jq -r '.distributed_shares[] | "\(.member_alias): \(.encrypted_share | length) bytes"'
fi

echo ""
echo "🔑 Member Keys Generated:"
echo "========================"
if [ -d "member_keys" ]; then
    ls -la member_keys/
    echo ""
    echo "📋 Member Key Files:"
    for key_file in member_keys/*.secret; do
        if [ -f "$key_file" ]; then
            member_name=$(basename "$key_file" .secret)
            echo "  - $member_name.secret (private key)"
            echo "  - $member_name.pub (public key)"
        fi
    done
fi

echo ""
echo "📁 Files Created:"
echo "  - share_distribution.json (distribution details)"
echo "  - member_keys/ (member private keys directory)"
echo ""
echo "✅ Step 4 Complete: Shares distributed to members!"
echo "   Next: Run Step 5 - Member Decryption"
echo "   Command: ./scripts/05-decrypt-shares.sh"
