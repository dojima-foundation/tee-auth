#!/bin/bash
set -e

echo "🚀 Starting renclave-v2 services"

# Configuration
ENCLAVE_SOCKET="/tmp/enclave.sock"
HOST_PORT="8080"

# Function to cleanup processes
cleanup() {
    echo "🧹 Cleaning up processes..."
    pkill -f "/app/bin/enclave" 2>/dev/null || true
    pkill -f "/app/bin/host" 2>/dev/null || true
    rm -f "$ENCLAVE_SOCKET" 2>/dev/null || true
    sleep 2
    echo "✅ Cleanup completed"
}

# Setup cleanup trap
trap cleanup EXIT INT TERM

# Function to wait for service
wait_for_service() {
    local service_name="$1"
    local check_command="$2"
    local timeout="${3:-30}"
    local interval="${4:-1}"
    
    echo "⏳ Waiting for $service_name to be ready..."
    
    local count=0
    while [ $count -lt $timeout ]; do
        if eval "$check_command" >/dev/null 2>&1; then
            echo "✅ $service_name is ready"
            return 0
        fi
        sleep $interval
        count=$((count + interval))
    done
    
    echo "❌ $service_name failed to start within ${timeout}s"
    return 1
}

# Start enclave
echo "🔒 Starting Nitro Enclave..."
/app/bin/enclave &
ENCLAVE_PID=$!

# Wait for enclave socket
wait_for_service "Enclave" "test -S $ENCLAVE_SOCKET" 30 1

# Start host
echo "🏠 Starting Host API Gateway..."
/app/bin/host &
HOST_PID=$!

# Wait for host HTTP server
wait_for_service "Host HTTP server" "curl -f http://localhost:$HOST_PORT/health" 30 1

echo "🎉 All services started successfully!"
echo "📋 Service Status:"
echo "  - Enclave PID: $ENCLAVE_PID"
echo "  - Host PID: $HOST_PID"
echo "  - HTTP API: http://localhost:$HOST_PORT"

# Keep services running
echo "🔄 Services are running. Press Ctrl+C to stop."
wait
