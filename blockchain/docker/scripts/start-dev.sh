#!/bin/bash

set -e

echo "🚀 Starting Geth in developer mode..."

# Set default data directory
GETH_DATA_DIR=${GETH_DATA_DIR:-/blockchain/data}

# Create data directory if it doesn't exist
mkdir -p $GETH_DATA_DIR

# Initialize blockchain with genesis configuration if not already done
if [ ! -f "$GETH_DATA_DIR/geth/genesis.json" ]; then
    echo "📋 Initializing blockchain with genesis configuration..."
    geth init /blockchain/config/genesis.json --datadir $GETH_DATA_DIR
fi

# Start Geth in developer mode
echo "🔧 Starting Geth with the following configuration:"
echo "   - HTTP RPC: Enabled on port 8545"
echo "   - WebSocket RPC: Enabled on port 8546"
echo "   - CORS: Enabled for all domains"
echo "   - Gas price: 0 (free transactions)"
echo "   - Block generation: On-demand"
echo "   - Chain ID: 1337"

exec geth \
    --dev \
    --dev.period 3 \
    --datadir $GETH_DATA_DIR \
    --http \
    --http.addr 0.0.0.0 \
    --http.port 8545 \
    --http.api ${GETH_HTTP_API:-eth,web3,net,personal,txpool} \
    --http.corsdomain ${GETH_HTTP_CORS_DOMAIN:-*} \
    --ws \
    --ws.addr 0.0.0.0 \
    --ws.port 8546 \
    --ws.api ${GETH_HTTP_API:-eth,web3,net,personal,txpool} \
    --ws.origins ${GETH_HTTP_CORS_DOMAIN:-*} \
    --networkid 1337 \
    --gasprice 0 \
    --unlock 0x7Aa16266Ba3d309e3cb278B452b1A6307E52Fb62,0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d \
    --password /dev/null \
    --allow-insecure-unlock \
    --verbosity 3
