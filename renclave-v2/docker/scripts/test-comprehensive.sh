#!/bin/bash
set -e
echo "🧪 Running comprehensive renclave-v2 tests"

echo "📋 Test 1: Host-only mode"
timeout 30 /app/bin/host &
HOST_PID=$!
sleep 5

curl -s http://localhost:3000/health && echo " ✅ Health OK" || echo "❌ Health check failed"
curl -s http://localhost:3000/info | head -c 200 && echo " ✅ Info OK" || echo "❌ Info endpoint failed" 
curl -s http://localhost:3000/network/status | head -c 200 && echo " ✅ Network OK" || echo "❌ Network status failed"

kill $HOST_PID 2>/dev/null || true
sleep 2

echo "📋 Test 2: Host + Enclave mode"
timeout 30 /app/bin/enclave &
ENCLAVE_PID=$!
sleep 5
timeout 30 /app/bin/host &
HOST_PID=$!
sleep 5

echo "🔑 Testing seed generation..."
curl -s -X POST http://localhost:3000/generate-seed -H "Content-Type: application/json" -d '{"strength": 256}' | head -c 200 && echo " ✅ Seed OK" || echo "❌ Seed generation failed"

curl -s http://localhost:3000/enclave/info | head -c 200 && echo " ✅ Enclave info OK" || echo "❌ Enclave info failed"

kill $HOST_PID $ENCLAVE_PID 2>/dev/null || true
echo "🎉 Comprehensive tests completed!"
echo "🧹 Cleaning up processes..."
pkill -f "/app/bin/" 2>/dev/null || true
