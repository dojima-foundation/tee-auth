#!/bin/bash

set -e

echo "🧪 Testing Faucet Service..."

# Test faucet health
echo "1️⃣ Testing faucet health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:3000/health)
echo "   Health Response: $HEALTH_RESPONSE"

if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo "   ✅ Faucet is healthy"
else
    echo "   ❌ Faucet is not healthy"
    exit 1
fi

# Test faucet info
echo ""
echo "2️⃣ Getting faucet information..."
INFO_RESPONSE=$(curl -s http://localhost:3000/faucet/info)
echo "   Info Response: $INFO_RESPONSE"

# Generate test address
echo ""
echo "3️⃣ Generating test address..."
TEST_ADDRESS="0x$(openssl rand -hex 20)"
echo "   Test Address: $TEST_ADDRESS"

# Request tokens from faucet
echo ""
echo "4️⃣ Requesting tokens from faucet..."
FAUCET_RESPONSE=$(curl -s -X POST http://localhost:3000/faucet \
    -H "Content-Type: application/json" \
    -d "{\"address\": \"$TEST_ADDRESS\"}")

echo "   Faucet Response: $FAUCET_RESPONSE"

if echo "$FAUCET_RESPONSE" | grep -q "success"; then
    echo "   ✅ Faucet test successful"
    
    # Extract transaction details
    TX_HASH=$(echo "$FAUCET_RESPONSE" | grep -o '"transactionHash":"[^"]*"' | cut -d'"' -f4)
    AMOUNT=$(echo "$FAUCET_RESPONSE" | grep -o '"amount":"[^"]*"' | cut -d'"' -f4)
    
    echo ""
    echo "📋 Transaction Details:"
    echo "   Hash: $TX_HASH"
    echo "   Amount: $AMOUNT ETH"
    echo "   Recipient: $TEST_ADDRESS"
    
    # Verify transaction on blockchain
    echo ""
    echo "5️⃣ Verifying transaction on blockchain..."
    sleep 5  # Wait for transaction to be mined
    
    # Get transaction details from Geth
    TX_DETAILS=$(curl -s -X POST http://localhost:8545 \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionByHash\",\"params\":[\"$TX_HASH\"],\"id\":1}")
    
    echo "   Transaction Details: $TX_DETAILS"
    
    if echo "$TX_DETAILS" | grep -q "$TEST_ADDRESS"; then
        echo "   ✅ Transaction verified on blockchain"
    else
        echo "   ⚠️  Transaction verification failed"
    fi
    
else
    echo "   ❌ Faucet test failed"
    exit 1
fi

echo ""
echo "🎉 All faucet tests passed!"
