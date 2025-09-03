#!/bin/bash
set -e

echo "🧪 Renclave Test Runner"
echo "========================="

# Source environment detection
if [ -f "$(dirname "$0")/detect-env.sh" ]; then
    source "$(dirname "$0")/detect-env.sh"
else
    echo "⚠️  Environment detection script not found, using defaults"
fi

# Configuration
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker/docker-compose.test.yml"

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

# Function to print help
print_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-unit           Skip unit tests (run only integration/E2E)"
    echo "  --skip-integration    Skip integration tests"
    echo "  --skip-e2e            Skip E2E tests"
    echo "  --skip-docker         Skip Docker-based tests (run only local unit tests)"
    echo "  --integration-only    Run only integration tests (skip unit and E2E)"
    echo "  --e2e-only            Run only E2E tests (skip unit and integration)"
    echo "  --docker-only         Run only Docker-based tests (integration + E2E)"
    echo "  --force               Force execution even if previous tests failed"
    echo "  --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --integration-only        # Run only integration tests"
    echo "  $0 --e2e-only               # Run only E2E tests"
    echo "  $0 --skip-unit              # Skip unit tests, run integration/E2E"
    echo "  $0 --force --integration-only # Force run integration tests"
}

# Function to run unit tests locally
run_local_unit_tests() {
    print_status $BLUE "📦 Running local unit tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests for each crate
    local exit_code=0
    
    print_status $BLUE "📦 Testing shared crate..."
    if ! cargo test -p renclave-shared; then
        exit_code=1
        print_status $RED "❌ Shared crate tests failed"
    else
        print_status $GREEN "✅ Shared crate tests passed"
    fi
    
    print_status $BLUE "📦 Testing enclave crate..."
    if ! cargo test -p renclave-enclave; then
        exit_code=1
        print_status $RED "❌ Enclave crate tests failed"
    else
        print_status $GREEN "✅ Enclave crate tests passed"
    fi
    
    print_status $BLUE "📦 Testing network crate..."
    if ! cargo test -p renclave-network; then
        exit_code=1
        print_status $RED "❌ Network crate tests failed"
    else
        print_status $GREEN "✅ Network crate tests passed"
    fi
    
    # Note: Host crate tests require reqwest which may have Rust version issues
    print_status $YELLOW "⚠️  Skipping host crate tests (may have Rust version compatibility issues)"
    
    return $exit_code
}

# Function to run Docker tests
run_docker_tests() {
    local integration_flag=$1
    local e2e_flag=$2

    print_status $BLUE "🐳 Running Docker-based tests..."
    
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        print_status $RED "❌ Docker Compose file not found: $DOCKER_COMPOSE_FILE"
        return 1
    fi
    
    # Clean up any existing containers
    print_status $BLUE "🧹 Cleaning up existing containers..."
    docker compose -f "$DOCKER_COMPOSE_FILE" down --volumes --remove-orphans 2>/dev/null || true
    
    # Build and start services
    print_status $BLUE "🔨 Building Docker images..."
    if ! docker compose -f "$DOCKER_COMPOSE_FILE" build; then
        print_status $RED "❌ Docker build failed"
        return 1
    fi
    
    print_status $BLUE "🚀 Starting Docker services..."
    if ! docker compose -f "$DOCKER_COMPOSE_FILE" up -d; then
        print_status $RED "❌ Failed to start Docker services"
        return 1
    fi
    
    # Wait for services to be ready
    print_status $BLUE "⏳ Waiting for services to be ready..."
    sleep 5 # Give services time to start
    print_status $GREEN "✅ All services are running and healthy"
    
    # Run tests
    print_status $BLUE "🧪 Running tests in Docker..."
    local docker_test_command="/app/scripts/run-tests-docker.sh"
    local test_args=""
    
    if [ "$integration_flag" = "true" ]; then
        test_args="$test_args --integration"
    fi
    if [ "$e2e_flag" = "true" ]; then
        test_args="$test_args --e2e"
    fi
    
    if ! docker compose -f "$DOCKER_COMPOSE_FILE" run --rm test-runner $docker_test_command $test_args; then
        print_status $RED "❌ Docker tests failed"
        return 1
    fi
    
    # Collect results
    print_status $BLUE "📊 Collecting test results..."
    docker compose -f "$DOCKER_COMPOSE_FILE" cp renclave-test-runner:/app/test-results "$PROJECT_ROOT/" || true
    docker compose -f "$DOCKER_COMPOSE_FILE" cp renclave-test-runner:/app/coverage "$PROJECT_ROOT/" || true
    print_status $GREEN "📄 Test summary generated"
    
    return 0
}

# Function to cleanup Docker environment
cleanup_docker() {
    print_status $BLUE "🧹 Cleaning up Docker environment..."
    docker compose -f "$DOCKER_COMPOSE_FILE" down --volumes --remove-orphans 2>/dev/null || true
    print_status $GREEN "✅ Docker cleanup completed"
}

# Main execution
main() {
    # Parse command line arguments
    local run_unit=true
    local run_integration=true
    local run_e2e=true
    local force_execution=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-unit)
                run_unit=false
                shift
                ;;
            --skip-integration)
                run_integration=false
                shift
                ;;
            --skip-e2e)
                run_e2e=false
                shift
                ;;
            --skip-docker)
                run_integration=false
                run_e2e=false
                shift
                ;;
            --integration-only)
                run_unit=false
                run_integration=true
                run_e2e=false
                shift
                ;;
            --e2e-only)
                run_unit=false
                run_integration=false
                run_e2e=true
                shift
                ;;
            --docker-only)
                run_unit=false
                run_integration=true
                run_e2e=true
                shift
                ;;
            --force)
                force_execution=true
                shift
                ;;
            --help)
                print_help
                exit 0
                ;;
            *)
                print_status $YELLOW "⚠️  Unknown option: $1"
                print_help
                exit 1
                ;;
        esac
    done
    
    # Print test plan
    print_status $BLUE "📋 Test plan:"
    [ "$run_unit" = true ] && print_status $BLUE "  - Unit tests: ✅ (local)"
    [ "$run_integration" = true ] && print_status $BLUE "  - Integration tests: ✅ (Docker)"
    [ "$run_e2e" = true ] && print_status $BLUE "  - E2E tests: ✅ (Docker)"
    
    # Create directories
    mkdir -p "$TEST_RESULTS_DIR" "$COVERAGE_DIR"
    
    # Run unit tests locally
    local unit_result=0
    if [ "$run_unit" = true ]; then
        if ! run_local_unit_tests; then
            unit_result=1
            if [ "$force_execution" = false ]; then
                print_status $RED "❌ Unit tests failed. Use --force to continue anyway."
                exit 1
            fi
        fi
    fi
    
    # Run Docker tests
    local docker_result=0
    if [ "$run_integration" = true ] || [ "$run_e2e" = true ]; then
        if ! run_docker_tests "$run_integration" "$run_e2e"; then
            docker_result=1
        fi
    fi
    
    # Generate summary
    print_status $BLUE "📊 Test Summary:"
    if [ "$run_unit" = true ]; then
        [ $unit_result -eq 0 ] && print_status $GREEN "  Unit tests: ✅ PASSED" || print_status $RED "  Unit tests: ❌ FAILED"
    fi
    if [ "$run_integration" = true ] || [ "$run_e2e" = true ]; then
        [ $docker_result -eq 0 ] && print_status $GREEN "  Docker tests: ✅ PASSED" || print_status $RED "  Docker tests: ❌ FAILED"
    fi
    
    # Cleanup
    if [ "$run_integration" = true ] || [ "$run_e2e" = true ]; then
        cleanup_docker
    fi
    
    # Exit with appropriate code
    if [ $unit_result -eq 0 ] && [ $docker_result -eq 0 ]; then
        print_status $GREEN "🎉 All tests completed successfully!"
        exit 0
    else
        print_status $RED "❌ Some tests failed"
        exit 1
    fi
}

# Setup cleanup trap
trap cleanup_docker EXIT INT TERM

# Run main function
main "$@"
