#!/bin/bash

set -e

echo "👤 Creating new Ethereum account..."

# Generate new account
ACCOUNT_INFO=$(geth account new --datadir /blockchain/data --password /dev/null 2>&1)

# Extract address from output
ADDRESS=$(echo "$ACCOUNT_INFO" | grep -o "0x[a-fA-F0-9]\{40\}")

if [ -z "$ADDRESS" ]; then
    echo "❌ Failed to create account"
    exit 1
fi

echo "✅ Account created successfully!"
echo "📍 Address: $ADDRESS"

# Get faucet tokens if faucet is available
if curl -f http://localhost:3000/health >/dev/null 2>&1; then
    echo "🚰 Requesting tokens from faucet..."
    
    RESPONSE=$(curl -s -X POST http://localhost:3000/faucet \
        -H "Content-Type: application/json" \
        -d "{\"address\": \"$ADDRESS\"}")
    
    if echo "$RESPONSE" | grep -q "success"; then
        TX_HASH=$(echo "$RESPONSE" | grep -o '"transactionHash":"[^"]*"' | cut -d'"' -f4)
        AMOUNT=$(echo "$RESPONSE" | grep -o '"amount":"[^"]*"' | cut -d'"' -f4)
        echo "💰 Received $AMOUNT ETH from faucet"
        echo "📋 Transaction: $TX_HASH"
    else
        echo "⚠️  Faucet request failed: $RESPONSE"
    fi
else
    echo "⚠️  Faucet not available. Use the faucet service to get tokens."
fi

echo ""
echo "🔑 Account Details:"
echo "   Address: $ADDRESS"
echo "   Network: Local Development (Chain ID: 1337)"
echo "   RPC URL: http://localhost:8545"
echo "   Faucet: http://localhost:3000/faucet"
