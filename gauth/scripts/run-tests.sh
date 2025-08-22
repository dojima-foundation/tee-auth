#!/bin/bash

# Comprehensive test runner for gauth service
# Supports unit tests, integration tests, e2e tests, benchmarks, and coverage

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
RUN_UNIT=false
RUN_INTEGRATION=false
RUN_E2E=false
RUN_BENCHMARK=false
RUN_COVERAGE=false
RUN_ALL=false
COVERAGE_THRESHOLD=80
VERBOSE=false
PARALLEL=true
SHORT=false

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Run tests for the gauth service with various options.

OPTIONS:
    -u, --unit              Run unit tests
    -i, --integration       Run integration tests (requires test database)
    -e, --e2e               Run end-to-end tests (requires test database and Redis)
    -b, --benchmark         Run benchmark tests
    -c, --coverage          Generate test coverage report
    -a, --all               Run all test types
    -t, --threshold N       Set coverage threshold (default: 80)
    -v, --verbose           Enable verbose output
    -s, --short             Run tests in short mode
    --no-parallel           Disable parallel test execution
    -h, --help              Show this help message

EXAMPLES:
    $0 --unit --coverage                    # Run unit tests with coverage
    $0 --integration --e2e                  # Run integration and e2e tests
    $0 --all                                # Run everything
    $0 --benchmark --verbose                # Run benchmarks with verbose output
    $0 --unit --threshold 90                # Run unit tests with 90% coverage threshold

ENVIRONMENT VARIABLES:
    TEST_DB_HOST         Database host for integration/e2e tests (default: localhost)
    TEST_DB_USER         Database user (default: gauth)
    TEST_DB_PASSWORD     Database password (default: password)
    TEST_DB_NAME         Database name (default: gauth_test)
    TEST_REDIS_HOST      Redis host for e2e tests (default: localhost)
    TEST_REDIS_PASSWORD  Redis password (default: "")

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--unit)
            RUN_UNIT=true
            shift
            ;;
        -i|--integration)
            RUN_INTEGRATION=true
            shift
            ;;
        -e|--e2e)
            RUN_E2E=true
            shift
            ;;
        -b|--benchmark)
            RUN_BENCHMARK=true
            shift
            ;;
        -c|--coverage)
            RUN_COVERAGE=true
            shift
            ;;
        -a|--all)
            RUN_ALL=true
            shift
            ;;
        -t|--threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        --no-parallel)
            PARALLEL=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            show_help
            exit 1
            ;;
    esac
done

# If no specific tests selected and not --all, default to unit tests
if [[ "$RUN_ALL" == "false" && "$RUN_UNIT" == "false" && "$RUN_INTEGRATION" == "false" && "$RUN_E2E" == "false" && "$RUN_BENCHMARK" == "false" ]]; then
    RUN_UNIT=true
    RUN_COVERAGE=true
fi

# If --all is selected, enable all test types
if [[ "$RUN_ALL" == "true" ]]; then
    RUN_UNIT=true
    RUN_INTEGRATION=true
    RUN_E2E=true
    RUN_BENCHMARK=true
    RUN_COVERAGE=true
fi

# Build test flags
TEST_FLAGS=""
if [[ "$VERBOSE" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [[ "$SHORT" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -short"
fi

if [[ "$PARALLEL" == "true" ]]; then
    TEST_FLAGS="$TEST_FLAGS -parallel 4"
fi

# Coverage flags
COVERAGE_FLAGS=""
if [[ "$RUN_COVERAGE" == "true" ]]; then
    COVERAGE_FLAGS="-coverprofile=coverage.out -covermode=atomic"
fi

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if required services are available
check_services() {
    local need_db=$1
    local need_redis=$2
    
    if [[ "$need_db" == "true" ]]; then
        local db_host=${TEST_DB_HOST:-localhost}
        local db_port=${TEST_DB_PORT:-5432}
        
        print_status $BLUE "Checking database connection..."
        if ! nc -z $db_host $db_port 2>/dev/null; then
            print_status $RED "Database not available at $db_host:$db_port"
            print_status $YELLOW "Please ensure PostgreSQL is running and accessible"
            exit 1
        fi
        print_status $GREEN "Database connection OK"
    fi
    
    if [[ "$need_redis" == "true" ]]; then
        local redis_host=${TEST_REDIS_HOST:-localhost}
        local redis_port=${TEST_REDIS_PORT:-6379}
        
        print_status $BLUE "Checking Redis connection..."
        if ! nc -z $redis_host $redis_port 2>/dev/null; then
            print_status $RED "Redis not available at $redis_host:$redis_port"
            print_status $YELLOW "Please ensure Redis is running and accessible"
            exit 1
        fi
        print_status $GREEN "Redis connection OK"
    fi
}

# Function to run tests with timeout
run_tests_with_timeout() {
    local test_type=$1
    local test_command=$2
    local timeout_duration=${3:-300} # 5 minutes default
    
    print_status $BLUE "Running $test_type tests..."
    
    if timeout $timeout_duration bash -c "$test_command"; then
        print_status $GREEN "$test_type tests PASSED"
        return 0
    else
        local exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            print_status $RED "$test_type tests TIMED OUT after ${timeout_duration}s"
        else
            print_status $RED "$test_type tests FAILED"
        fi
        return $exit_code
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    if [[ ! -f "coverage.out" ]]; then
        print_status $YELLOW "No coverage file found, skipping coverage report"
        return 0
    fi
    
    print_status $BLUE "Generating coverage report..."
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    
    # Calculate coverage percentage
    local coverage_percent=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
    
    print_status $BLUE "Coverage: ${coverage_percent}%"
    
    # Check coverage threshold
    if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
        print_status $GREEN "Coverage ${coverage_percent}% meets threshold of ${COVERAGE_THRESHOLD}%"
    else
        print_status $RED "Coverage ${coverage_percent}% below threshold of ${COVERAGE_THRESHOLD}%"
        return 1
    fi
    
    print_status $BLUE "Coverage report saved to coverage.html"
}

# Main execution
main() {
    print_status $BLUE "Starting gauth test suite..."
    
    # Change to the correct directory
    cd "$(dirname "$0")/.."
    
    # Clean up previous coverage files
    rm -f coverage.out coverage.html
    
    local overall_result=0
    
    # Run unit tests
    if [[ "$RUN_UNIT" == "true" ]]; then
        local unit_command="go test $TEST_FLAGS $COVERAGE_FLAGS ./internal/service/... ./pkg/..."
        if ! run_tests_with_timeout "Unit" "$unit_command" 180; then
            overall_result=1
        fi
    fi
    
    # Run integration tests
    if [[ "$RUN_INTEGRATION" == "true" ]]; then
        check_services true false
        export INTEGRATION_TESTS=true
        
        local integration_command="go test $TEST_FLAGS ./test/integration/..."
        if ! run_tests_with_timeout "Integration" "$integration_command" 300; then
            overall_result=1
        fi
    fi
    
    # Run e2e tests
    if [[ "$RUN_E2E" == "true" ]]; then
        check_services true true
        export E2E_TESTS=true
        
        local e2e_command="go test $TEST_FLAGS ./test/e2e/..."
        if ! run_tests_with_timeout "E2E" "$e2e_command" 600; then
            overall_result=1
        fi
    fi
    
    # Run benchmarks
    if [[ "$RUN_BENCHMARK" == "true" ]]; then
        print_status $BLUE "Running benchmark tests..."
        local benchmark_command="go test -bench=. -benchmem ./test/benchmark/..."
        
        if $benchmark_command; then
            print_status $GREEN "Benchmark tests completed"
        else
            print_status $RED "Benchmark tests failed"
            overall_result=1
        fi
    fi
    
    # Generate coverage report
    if [[ "$RUN_COVERAGE" == "true" ]]; then
        if ! generate_coverage_report; then
            overall_result=1
        fi
    fi
    
    # Print summary
    print_status $BLUE "Test Summary:"
    echo "=============="
    
    if [[ "$RUN_UNIT" == "true" ]]; then
        echo "• Unit tests: $([ $overall_result -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
    fi
    
    if [[ "$RUN_INTEGRATION" == "true" ]]; then
        echo "• Integration tests: $([ $overall_result -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
    fi
    
    if [[ "$RUN_E2E" == "true" ]]; then
        echo "• E2E tests: $([ $overall_result -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
    fi
    
    if [[ "$RUN_BENCHMARK" == "true" ]]; then
        echo "• Benchmarks: $([ $overall_result -eq 0 ] && echo "✅ COMPLETED" || echo "❌ FAILED")"
    fi
    
    if [[ "$RUN_COVERAGE" == "true" ]]; then
        echo "• Coverage: $([ $overall_result -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")"
    fi
    
    if [[ $overall_result -eq 0 ]]; then
        print_status $GREEN "All tests passed! 🎉"
    else
        print_status $RED "Some tests failed! 💥"
    fi
    
    return $overall_result
}

# Run main function
main "$@"
