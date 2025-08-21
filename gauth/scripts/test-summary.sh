#!/bin/bash

# Test Summary Script for gauth service
# Provides a quick overview of test status and coverage

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Print header
print_header() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    GAUTH TEST SUMMARY                        ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Print section
print_section() {
    echo -e "${CYAN}▶ $1${NC}"
    echo "  ────────────────────────────────────────────────────────────"
}

# Print success
print_success() {
    echo -e "  ${GREEN}✓ $1${NC}"
}

# Print warning
print_warning() {
    echo -e "  ${YELLOW}⚠ $1${NC}"
}

# Print error
print_error() {
    echo -e "  ${RED}✗ $1${NC}"
}

# Print info
print_info() {
    echo -e "  ${PURPLE}ℹ $1${NC}"
}

# Check if coverage file exists
check_coverage() {
    if [ -f "coverage/unit.out" ]; then
        return 0
    else
        return 1
    fi
}

# Get coverage percentage
get_coverage() {
    if check_coverage; then
        go tool cover -func=coverage/unit.out | grep total | grep -oE '[0-9]+\.[0-9]+' || echo "0.0"
    else
        echo "0.0"
    fi
}

# Count test files
count_test_files() {
    find . -name "*_test.go" | wc -l | tr -d ' '
}

# Count test functions
count_test_functions() {
    grep -r "^func Test" --include="*_test.go" . | wc -l | tr -d ' '
}

# Count benchmark functions
count_benchmark_functions() {
    grep -r "^func Benchmark" --include="*_test.go" . | wc -l | tr -d ' '
}

# Check if dependencies are available
check_dependencies() {
    local deps_available=0
    
    # Check PostgreSQL
    if pg_isready -h ${TEST_DB_HOST:-localhost} -p ${TEST_DB_PORT:-5432} -U ${TEST_DB_USER:-gauth} &>/dev/null; then
        print_success "PostgreSQL available for integration tests"
        deps_available=$((deps_available + 1))
    else
        print_warning "PostgreSQL not available (integration tests will be skipped)"
    fi
    
    # Check Redis
    if redis-cli -h ${TEST_REDIS_HOST:-localhost} -p ${TEST_REDIS_PORT:-6379} ping &>/dev/null; then
        print_success "Redis available for integration tests"
        deps_available=$((deps_available + 1))
    else
        print_warning "Redis not available (integration tests will be skipped)"
    fi
    
    # Check gauth service (for E2E tests)
    if curl -f http://${E2E_GAUTH_HOST:-localhost}:${E2E_GAUTH_PORT:-9090}/health &>/dev/null; then
        print_success "gauth service available for E2E tests"
        deps_available=$((deps_available + 1))
    else
        print_warning "gauth service not available (E2E tests will be skipped)"
    fi
    
    return $deps_available
}

# Get test environment info
get_env_info() {
    echo "    Go Version: $(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+')"
    echo "    OS: $(uname -s) $(uname -r)"
    echo "    Architecture: $(uname -m)"
    echo "    Working Directory: $(pwd)"
    echo "    Git Branch: $(git branch --show-current 2>/dev/null || echo 'unknown')"
    echo "    Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
}

# Analyze test coverage by package
analyze_coverage() {
    if ! check_coverage; then
        print_warning "No coverage data available. Run 'make test-unit' first."
        return
    fi
    
    echo "    Package Coverage Analysis:"
    echo ""
    
    go tool cover -func=coverage/unit.out | grep -v "total:" | while read line; do
        PERCENT=$(echo "$line" | grep -oE '[0-9]+\.[0-9]+%' | sed 's/%//')
        PACKAGE=$(echo "$line" | awk '{print $1}' | sed 's|github.com/dojima-foundation/tee-auth/gauth/||')
        
        if (( $(echo "$PERCENT >= 90" | bc -l) )); then
            echo -e "      ${GREEN}$PACKAGE: $PERCENT%${NC}"
        elif (( $(echo "$PERCENT >= 70" | bc -l) )); then
            echo -e "      ${YELLOW}$PACKAGE: $PERCENT%${NC}"
        else
            echo -e "      ${RED}$PACKAGE: $PERCENT%${NC}"
        fi
    done
}

# Show recent test results
show_recent_results() {
    if [ -f "coverage/unit-tests.log" ]; then
        local passed=$(grep -c "PASS:" coverage/unit-tests.log 2>/dev/null || echo "0")
        local failed=$(grep -c "FAIL:" coverage/unit-tests.log 2>/dev/null || echo "0")
        local total=$((passed + failed))
        
        if [ $total -gt 0 ]; then
            print_info "Recent unit test results: $passed passed, $failed failed (total: $total)"
        fi
    fi
    
    if [ -f "coverage/integration-tests.log" ]; then
        local passed=$(grep -c "PASS:" coverage/integration-tests.log 2>/dev/null || echo "0")
        local failed=$(grep -c "FAIL:" coverage/integration-tests.log 2>/dev/null || echo "0")
        local total=$((passed + failed))
        
        if [ $total -gt 0 ]; then
            print_info "Recent integration test results: $passed passed, $failed failed (total: $total)"
        fi
    fi
    
    if [ -f "coverage/e2e-tests.log" ]; then
        local passed=$(grep -c "PASS:" coverage/e2e-tests.log 2>/dev/null || echo "0")
        local failed=$(grep -c "FAIL:" coverage/e2e-tests.log 2>/dev/null || echo "0")
        local total=$((passed + failed))
        
        if [ $total -gt 0 ]; then
            print_info "Recent E2E test results: $passed passed, $failed failed (total: $total)"
        fi
    fi
}

# Show available test commands
show_commands() {
    echo "    Available Commands:"
    echo ""
    echo "      make test-unit              # Run unit tests with coverage"
    echo "      make test-integration       # Run integration tests (requires DB)"
    echo "      make test-e2e              # Run end-to-end tests (requires services)"
    echo "      make test-coverage         # Generate comprehensive coverage report"
    echo "      make test-coverage-open    # Generate and open coverage report"
    echo "      make test-all              # Run all test types"
    echo "      make bench                 # Run benchmarks"
    echo ""
    echo "    Environment Variables:"
    echo ""
    echo "      INTEGRATION_TESTS=true     # Enable integration tests"
    echo "      E2E_TESTS=true            # Enable end-to-end tests"
    echo "      OPEN_REPORT=true          # Auto-open coverage report"
    echo "      RUN_BENCHMARKS=true       # Include benchmarks in coverage"
}

# Show test file structure
show_structure() {
    echo "    Test Structure:"
    echo ""
    if [ -d "test" ]; then
        tree test/ 2>/dev/null || find test/ -type f | sed 's/^/      /'
    else
        echo "      test/ directory not found"
    fi
    echo ""
    echo "    Unit Test Files:"
    find . -name "*_test.go" -not -path "./test/*" | head -10 | sed 's/^/      /' || echo "      No unit test files found"
    
    local remaining=$(find . -name "*_test.go" -not -path "./test/*" | wc -l | tr -d ' ')
    if [ $remaining -gt 10 ]; then
        echo "      ... and $((remaining - 10)) more files"
    fi
}

# Main execution
main() {
    print_header
    
    # Basic statistics
    print_section "Test Statistics"
    local test_files=$(count_test_files)
    local test_functions=$(count_test_functions)
    local benchmark_functions=$(count_benchmark_functions)
    local coverage=$(get_coverage)
    
    print_info "Test Files: $test_files"
    print_info "Test Functions: $test_functions"
    print_info "Benchmark Functions: $benchmark_functions"
    
    if (( $(echo "$coverage >= 80.0" | bc -l) )); then
        print_success "Overall Coverage: $coverage%"
    elif (( $(echo "$coverage >= 60.0" | bc -l) )); then
        print_warning "Overall Coverage: $coverage%"
    else
        print_error "Overall Coverage: $coverage%"
    fi
    echo ""
    
    # Dependencies check
    print_section "Test Dependencies"
    check_dependencies
    deps_count=$?
    
    if [ $deps_count -eq 3 ]; then
        print_success "All test dependencies available"
    elif [ $deps_count -eq 0 ]; then
        print_warning "No test dependencies available (unit tests only)"
    else
        print_warning "$deps_count out of 3 test dependencies available"
    fi
    echo ""
    
    # Environment information
    print_section "Environment Information"
    get_env_info
    echo ""
    
    # Coverage analysis
    print_section "Coverage Analysis"
    analyze_coverage
    echo ""
    
    # Recent results
    print_section "Recent Test Results"
    show_recent_results
    echo ""
    
    # Test structure
    print_section "Test Structure"
    show_structure
    echo ""
    
    # Available commands
    print_section "Available Commands"
    show_commands
    echo ""
    
    # Final recommendations
    print_section "Recommendations"
    
    if (( $(echo "$coverage < 80.0" | bc -l) )); then
        print_warning "Consider increasing test coverage to 80%+"
        print_info "Focus on service layer and database operations"
    fi
    
    if [ $deps_count -lt 3 ]; then
        print_warning "Set up test dependencies for comprehensive testing"
        print_info "See test/README.md for setup instructions"
    fi
    
    if [ ! -f "coverage/unit.out" ]; then
        print_info "Run 'make test-unit' to generate coverage data"
    fi
    
    print_success "Test suite ready for development!"
    echo ""
    
    # Footer
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  For detailed test documentation, see: test/README.md        ║${NC}"
    echo -e "${BLUE}║  For coverage reports, see: coverage/                        ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
}

# Run main function
main "$@"
