#!/bin/bash

# 🚀 CI/CD Test Runner for Renclave
# Simplified version for continuous integration

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TIMEOUT_SECONDS=600

print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")
            echo -e "${YELLOW}ℹ️  $message${NC}"
            ;;
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    print_status "INFO" "Checking prerequisites..."
    
    if ! command -v cargo &> /dev/null; then
        print_status "ERROR" "Rust/cargo is not installed"
        exit 1
    fi
    
    local rust_version=$(rustc --version)
    print_status "INFO" "Using Rust: $rust_version"
}

# Run tests with timeout
run_tests() {
    local test_type=$1
    local command=$2
    
    print_status "INFO" "Running $test_type tests..."
    
    if timeout "$TIMEOUT_SECONDS" bash -c "$command"; then
        print_status "SUCCESS" "$test_type tests passed"
        return 0
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            print_status "ERROR" "$test_type tests timed out after ${TIMEOUT_SECONDS}s"
        else
            print_status "ERROR" "$test_type tests failed with exit code $exit_code"
        fi
        return $exit_code
    fi
}

# Main execution
main() {
    print_status "INFO" "Starting CI test suite..."
    cd "$PROJECT_ROOT"
    
    check_prerequisites
    
    # Clean previous builds
    print_status "INFO" "Cleaning previous builds..."
    cargo clean
    
    # Check code formatting
    print_status "INFO" "Checking code formatting..."
    if ! cargo fmt -- --check; then
        print_status "ERROR" "Code formatting check failed"
        exit 1
    fi
    
    # Check clippy
    print_status "INFO" "Running clippy checks..."
    if ! cargo clippy --workspace --all-targets --all-features -- -D warnings; then
        print_status "ERROR" "Clippy checks failed"
        exit 1
    fi
    
    # Run unit tests
    if ! run_tests "unit" "cargo test --workspace --lib"; then
        exit 1
    fi
    
    # Run integration tests (if they exist)
    if [ -d "tests" ]; then
        if ! run_tests "integration" "cargo test --test integration_tests"; then
            exit 1
        fi
    fi
    
    # Build all crates
    print_status "INFO" "Building all crates..."
    if ! cargo build --workspace; then
        print_status "ERROR" "Build failed"
        exit 1
    fi
    
    print_status "SUCCESS" "All CI tests passed! 🎉"
}

# Handle interruption
trap 'print_status "ERROR" "CI test run interrupted"; exit 1' INT TERM

main "$@"
