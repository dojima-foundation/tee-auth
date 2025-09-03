#!/bin/bash
set -e

echo "🐳 Running tests inside Docker container"

# Configuration
ENCLAVE_SOCKET="${ENCLAVE_SOCKET:-/tmp/enclave.sock}"
HOST_URL="${HOST_URL:-http://localhost:3000}"
TEST_MODE="${TEST_MODE:-docker}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run unit tests (for host crate, or if --unit-only is passed)
run_unit_tests() {
    print_status $BLUE "📦 Testing unit tests for host crate..."
    # Run unit tests for the host crate specifically
    cargo test -p renclave-host
    if [ $? -eq 0 ]; then
        print_status $GREEN "✅ Host crate unit tests passed"
    else
        print_status $RED "❌ Host crate unit tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status $BLUE "🧪 Running integration tests..."

    if [ -d "/app/tests" ]; then
        cd /app
        print_status $BLUE "🔍 Setting up integration test environment..."
        
        # Create a proper test directory structure
        mkdir -p /tmp/integration-test/src
        cd /tmp/integration-test
        
        # Create a Cargo.toml for the integration tests
        cat > Cargo.toml << 'EOF'
[package]
name = "renclave-integration-tests"
version = "0.1.0"
edition = "2021"

[dependencies]
renclave-shared = { path = "/app/src/shared" }
renclave-enclave = { path = "/app/src/enclave" }
renclave-network = { path = "/app/src/network" }
tokio = { version = "1.0", features = ["full", "test-util"] }
serde_json = "1.0"
futures = "0.3"
EOF
        
        # Copy the integration test file
        cp /app/tests/integration_tests.rs src/lib.rs
        
        print_status $BLUE "🧪 Running actual integration tests..."
        
        # Actually run the tests and capture the real result
        if cargo test -- --nocapture; then
            print_status $GREEN "✅ Integration tests PASSED"
            return 0
        else
            print_status $RED "❌ Integration tests FAILED"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  Integration tests directory not found: /app/tests"
        return 1
    fi
}

# Function to run E2E tests
run_e2e_tests() {
    print_status $BLUE "🚀 Running E2E tests..."

    if [ -d "/app/tests" ]; then
        cd /app
        print_status $BLUE "🔍 Setting up E2E test environment..."
        
        # Create a proper test directory structure
        mkdir -p /tmp/e2e-test/src
        cd /tmp/e2e-test
        
        # Create a Cargo.toml for the E2E tests
        cat > Cargo.toml << 'EOF'
[package]
name = "renclave-e2e-tests"
version = "0.1.0"
edition = "2021"

[dependencies]
renclave-shared = { path = "/app/src/shared" }
renclave-enclave = { path = "/app/src/enclave" }
renclave-network = { path = "/app/src/network" }
tokio = { version = "1.0", features = ["full", "test-util"] }
serde_json = "1.0"
futures = "0.3"
EOF
        
        # Copy the E2E test file
        cp /app/tests/e2e_tests.rs src/lib.rs
        
        print_status $BLUE "🧪 Running actual E2E tests..."
        
        # Actually run the tests and capture the real result
        if cargo test -- --nocapture; then
            print_status $GREEN "✅ E2E tests PASSED"
            return 0
        else
            print_status $RED "❌ E2E tests FAILED"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  E2E tests directory not found: /app/tests"
        return 1
    fi
}

# Parse arguments
RUN_UNIT_TESTS=false
RUN_INTEGRATION_TESTS=false
RUN_E2E_TESTS=false
UNIT_ONLY=false

if [ "$#" -eq 0 ]; then
    RUN_UNIT_TESTS=true
    RUN_INTEGRATION_TESTS=true
    RUN_E2E_TESTS=true
else
    for arg in "$@"; do
        case $arg in
            --unit-only)
                UNIT_ONLY=true
                RUN_UNIT_TESTS=true
                ;;
            --integration)
                RUN_INTEGRATION_TESTS=true
                ;;
            --e2e)
                RUN_E2E_TESTS=true
                ;;
            *)
                echo "Unknown argument: $arg"
                exit 1
                ;;
        esac
    done
fi

OVERALL_SUCCESS=0

# Run unit tests if requested (specifically for host crate in Docker)
if [ "$RUN_UNIT_TESTS" = "true" ]; then
    run_unit_tests
    if [ $? -ne 0 ]; then
        OVERALL_SUCCESS=1
    fi
fi

# Run integration tests if requested
if [ "$RUN_INTEGRATION_TESTS" = "true" ]; then
    run_integration_tests
    if [ $? -ne 0 ]; then
        OVERALL_SUCCESS=1
    fi
fi

# Run E2E tests if requested
if [ "$RUN_E2E_TESTS" = "true" ]; then
    run_e2e_tests
    if [ $? -ne 0 ]; then
        OVERALL_SUCCESS=1
    fi
fi

# Generate a simple test summary
echo "📊 Generating test summary..."
echo "--- Test Summary ---" > /app/test-results/test-summary.txt
if [ $OVERALL_SUCCESS -eq 0 ]; then
    echo "Docker tests: ✅ PASSED" >> /app/test-results/test-summary.txt
    print_status $GREEN "✅ All tests completed successfully in Docker."
else
    echo "Docker tests: ❌ FAILED" >> /app/test-results/test-summary.txt
    print_status $RED "❌ Some tests failed in Docker."
    exit 1
fi
