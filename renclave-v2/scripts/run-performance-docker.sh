#!/bin/bash

# Docker-based performance testing script for Renclave
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored status
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_status $RED "❌ Docker is not running or not accessible"
        exit 1
    fi
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! docker compose version >/dev/null 2>&1; then
        print_status $RED "❌ Docker Compose is not available"
        exit 1
    fi
}

# Function to run performance tests in Docker
run_performance_tests() {
    local container_name="renclave-performance-runner"
    
    print_status $BLUE "🚀 Starting performance tests in Docker..."
    
    # Start the performance runner container
    docker compose -f docker/docker-compose.test.yml up -d performance-runner
    
    # Wait for container to be ready
    print_status $BLUE "⏳ Waiting for performance runner to be ready..."
    sleep 5
    
    # Run benchmarks
    print_status $BLUE "📊 Running seed generation benchmarks..."
    docker exec $container_name cargo bench --bench seed_generation -- --verbose > performance-results/seed_generation.txt 2>&1
    
    print_status $BLUE "📊 Running seed validation benchmarks..."
    docker exec $container_name cargo bench --bench seed_validation -- --verbose > performance-results/seed_validation.txt 2>&1
    
    print_status $BLUE "📊 Running concurrent operations benchmarks..."
    docker exec $container_name cargo bench --bench concurrent_operations -- --verbose > performance-results/concurrent_operations.txt 2>&1
    
    print_status $BLUE "📊 Running stress test benchmarks..."
    docker exec $container_name cargo bench --bench stress_tests -- --verbose > performance-results/stress_tests.txt 2>&1
    
    print_status $GREEN "✅ Performance tests completed!"
}

# Function to run coverage generation in Docker
run_coverage_generation() {
    local container_name="renclave-performance-runner"
    
    print_status $BLUE "📊 Generating coverage reports in Docker..."
    
    # Install coverage tools
    print_status $BLUE "📦 Installing coverage tools..."
    docker exec $container_name cargo install cargo-llvm-cov
    docker exec $container_name cargo install grcov
    
    # Generate coverage
    print_status $BLUE "📊 Generating overall workspace coverage..."
    docker exec $container_name cargo llvm-cov --workspace --html --output-dir coverage/workspace
    
    print_status $BLUE "📊 Generating coverage for individual crates..."
    docker exec $container_name cargo llvm-cov --package renclave-shared --html --output-dir coverage/shared
    docker exec $container_name cargo llvm-cov --package renclave-enclave --html --output-dir coverage/enclave
    docker exec $container_name cargo llvm-cov --package renclave-network --html --output-dir coverage/network
    
    print_status $GREEN "✅ Coverage generation completed!"
}

# Function to cleanup
cleanup() {
    print_status $BLUE "🧹 Cleaning up Docker environment..."
    docker compose -f docker/docker-compose.test.yml down
    print_status $GREEN "✅ Cleanup completed!"
}

# Main execution
main() {
    print_status $BLUE "🚀 Renclave Docker Performance Testing & Coverage"
    print_status $BLUE "=================================================="
    
    # Check prerequisites
    check_docker
    check_docker_compose
    
    # Create directories
    mkdir -p performance-results coverage
    
    # Parse command line arguments
    case "${1:-all}" in
        "performance")
            run_performance_tests
            ;;
        "coverage")
            run_coverage_generation
            ;;
        "all")
            run_performance_tests
            run_coverage_generation
            ;;
        *)
            print_status $RED "❌ Unknown option: $1"
            print_status $BLUE "Usage: $0 [performance|coverage|all]"
            exit 1
            ;;
    esac
    
    print_status $GREEN "🎉 All operations completed successfully!"
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
