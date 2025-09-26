#!/bin/bash

# Step 2: Start TEE and Host Services
# This script starts the enclave and host processes

set -e

echo "🚀 Step 2: Starting TEE and Host Services"
echo "========================================="

# Kill any existing processes
echo "🧹 Cleaning up existing processes..."
pkill -f "enclave" 2>/dev/null || true
pkill -f "host" 2>/dev/null || true
sleep 2

# Start the enclave process
echo "🔒 Starting TEE Enclave..."
cd /app
./target/release/enclave > enclave.log 2>&1 &
ENCLAVE_PID=$!
echo "   Enclave PID: $ENCLAVE_PID"

# Wait a moment for enclave to initialize
sleep 3

# Start the host process
echo "🌐 Starting Host Service..."
./target/release/host > host.log 2>&1 &
HOST_PID=$!
echo "   Host PID: $HOST_PID"

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check if services are running
if ! kill -0 $ENCLAVE_PID 2>/dev/null; then
    echo "❌ Enclave process failed to start"
    echo "Enclave logs:"
    cat enclave.log
    exit 1
fi

if ! kill -0 $HOST_PID 2>/dev/null; then
    echo "❌ Host process failed to start"
    echo "Host logs:"
    cat host.log
    kill $ENCLAVE_PID 2>/dev/null || true
    exit 1
fi

# Test if host is responding
echo "🔍 Testing host connectivity..."
for i in {1..10}; do
    if curl -s http://localhost:3000/health > /dev/null 2>&1; then
        echo "✅ Host is responding on port 3000"
        break
    fi
    if [ $i -eq 10 ]; then
        echo "❌ Host is not responding after 10 attempts"
        echo "Host logs:"
        cat host.log
        kill $ENCLAVE_PID $HOST_PID 2>/dev/null || true
        exit 1
    fi
    echo "   Attempt $i/10: Waiting for host..."
    sleep 2
done

echo ""
echo "✅ Step 2 Complete: Services started successfully!"
echo "   - Enclave PID: $ENCLAVE_PID"
echo "   - Host PID: $HOST_PID"
echo "   - Host URL: http://localhost:3000"
echo ""
echo "📋 Next: Run Step 3 - Genesis Boot"
echo "   Command: ./scripts/03-genesis-boot.sh"
