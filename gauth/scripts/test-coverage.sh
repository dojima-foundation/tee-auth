#!/bin/bash

# Test Coverage Script for gauth service
# Runs comprehensive tests with coverage reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
COVERAGE_XML="coverage.xml"
MIN_COVERAGE=80

# Functions
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Create coverage directory
mkdir -p $COVERAGE_DIR

print_header "GAUTH SERVICE TEST COVERAGE REPORT"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+')
print_success "Using $GO_VERSION"

# Install test dependencies
print_header "Installing Test Dependencies"
go mod download
go mod tidy
print_success "Dependencies installed"

# Install additional tools for coverage (optional)
print_header "Installing Coverage Tools"
if go install github.com/axw/gocov/gocov@latest 2>/dev/null; then
    print_success "gocov installed"
else
    print_warning "gocov installation failed, using built-in Go tools"
fi

if go install github.com/AlekSi/gocov-xml@latest 2>/dev/null; then
    print_success "gocov-xml installed"
else
    print_warning "gocov-xml installation failed, XML reports will be skipped"
fi

print_success "Coverage tools setup completed"

# Clean previous coverage data
rm -f $COVERAGE_DIR/$COVERAGE_FILE
rm -f $COVERAGE_DIR/$COVERAGE_HTML
rm -f $COVERAGE_DIR/$COVERAGE_XML

print_header "Running Unit Tests with Coverage"

# Run unit tests with coverage
go test -v -race -covermode=atomic -coverprofile=$COVERAGE_DIR/$COVERAGE_FILE ./internal/... ./pkg/... 2>&1 | tee $COVERAGE_DIR/unit-tests.log

if [ ${PIPESTATUS[0]} -ne 0 ]; then
    print_error "Unit tests failed"
    exit 1
fi

print_success "Unit tests completed"

# Check if integration tests should be run
if [ "$INTEGRATION_TESTS" = "true" ]; then
    print_header "Running Integration Tests"
    
    # Check if test database is available
    if ! pg_isready -h ${TEST_DB_HOST:-localhost} -p ${TEST_DB_PORT:-5432} -U ${TEST_DB_USER:-gauth} -d ${TEST_DB_NAME:-gauth_test} &> /dev/null; then
        print_warning "Test database not available, skipping integration tests"
    else
        # Run integration tests and merge coverage
        go test -v -race -covermode=atomic -coverprofile=$COVERAGE_DIR/integration.out ./test/integration/... 2>&1 | tee $COVERAGE_DIR/integration-tests.log
        
        if [ -f $COVERAGE_DIR/integration.out ]; then
            # Merge coverage files
            echo "mode: atomic" > $COVERAGE_DIR/merged.out
            tail -n +2 $COVERAGE_DIR/$COVERAGE_FILE >> $COVERAGE_DIR/merged.out
            tail -n +2 $COVERAGE_DIR/integration.out >> $COVERAGE_DIR/merged.out
            mv $COVERAGE_DIR/merged.out $COVERAGE_DIR/$COVERAGE_FILE
            print_success "Integration tests completed and coverage merged"
        fi
    fi
fi

# Check if E2E tests should be run
if [ "$E2E_TESTS" = "true" ]; then
    print_header "Running End-to-End Tests"
    
    # Check if services are running
    if ! curl -f http://${E2E_GAUTH_HOST:-localhost}:${E2E_GAUTH_PORT:-9090}/health &> /dev/null; then
        print_warning "gauth service not available, skipping E2E tests"
    else
        go test -v -timeout=5m ./test/e2e/... 2>&1 | tee $COVERAGE_DIR/e2e-tests.log
        print_success "E2E tests completed"
    fi
fi

# Generate coverage report
print_header "Generating Coverage Reports"

if [ ! -f $COVERAGE_DIR/$COVERAGE_FILE ]; then
    print_error "No coverage data found"
    exit 1
fi

# Generate HTML coverage report
go tool cover -html=$COVERAGE_DIR/$COVERAGE_FILE -o $COVERAGE_DIR/$COVERAGE_HTML
print_success "HTML coverage report generated: $COVERAGE_DIR/$COVERAGE_HTML"

# Generate XML coverage report for CI/CD (if tools available)
if command -v gocov &> /dev/null && command -v gocov-xml &> /dev/null; then
    gocov convert $COVERAGE_DIR/$COVERAGE_FILE | gocov-xml > $COVERAGE_DIR/$COVERAGE_XML
    print_success "XML coverage report generated: $COVERAGE_DIR/$COVERAGE_XML"
    
    # Generate JSON coverage report
    gocov convert $COVERAGE_DIR/$COVERAGE_FILE > $COVERAGE_DIR/coverage.json
    print_success "JSON coverage report generated: $COVERAGE_DIR/coverage.json"
else
    print_warning "gocov tools not available, skipping XML and JSON reports"
    # Create a simple JSON report using go tool cover
    COVERAGE_PERCENT=$(go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep total | grep -oE '[0-9]+\.[0-9]+')
    echo '{"coverage_percent": "'$COVERAGE_PERCENT'"}' > $COVERAGE_DIR/coverage.json
    print_success "Basic JSON coverage report generated: $COVERAGE_DIR/coverage.json"
fi

# Calculate coverage percentage
COVERAGE_PERCENT=$(go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep total | grep -oE '[0-9]+\.[0-9]+')

print_header "Coverage Summary"

echo "📊 Total Coverage: ${COVERAGE_PERCENT}%"

if (( $(echo "$COVERAGE_PERCENT >= $MIN_COVERAGE" | bc -l) )); then
    print_success "Coverage meets minimum requirement (${MIN_COVERAGE}%)"
else
    print_error "Coverage below minimum requirement (${MIN_COVERAGE}%)"
    COVERAGE_FAILED=true
fi

# Show detailed coverage by package
print_header "Coverage by Package"
go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep -v "total:" | while read line; do
    PERCENT=$(echo "$line" | grep -oE '[0-9]+\.[0-9]+%')
    PACKAGE=$(echo "$line" | awk '{print $1}' | sed 's|github.com/dojima-foundation/tee-auth/gauth/||')
    
    if (( $(echo "$PERCENT" | sed 's/%//' | awk '{print ($1 >= 80)}') )); then
        echo -e "${GREEN}  $PACKAGE: $PERCENT${NC}"
    elif (( $(echo "$PERCENT" | sed 's/%//' | awk '{print ($1 >= 60)}') )); then
        echo -e "${YELLOW}  $PACKAGE: $PERCENT${NC}"
    else
        echo -e "${RED}  $PACKAGE: $PERCENT${NC}"
    fi
done

# Generate coverage badge
print_header "Generating Coverage Badge"
BADGE_COLOR="red"
if (( $(echo "$COVERAGE_PERCENT >= 80" | bc -l) )); then
    BADGE_COLOR="brightgreen"
elif (( $(echo "$COVERAGE_PERCENT >= 60" | bc -l) )); then
    BADGE_COLOR="yellow"
fi

# Create a simple coverage badge URL
BADGE_URL="https://img.shields.io/badge/coverage-${COVERAGE_PERCENT}%25-${BADGE_COLOR}"
echo "Coverage Badge URL: $BADGE_URL" > $COVERAGE_DIR/badge.txt
print_success "Coverage badge URL saved: $COVERAGE_DIR/badge.txt"

# Generate uncovered lines report
print_header "Analyzing Uncovered Code"
go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep -E ":[0-9]+:.*0\.0%" > $COVERAGE_DIR/uncovered.txt || true

if [ -s $COVERAGE_DIR/uncovered.txt ]; then
    print_warning "Found uncovered code (see $COVERAGE_DIR/uncovered.txt):"
    head -10 $COVERAGE_DIR/uncovered.txt
    if [ $(wc -l < $COVERAGE_DIR/uncovered.txt) -gt 10 ]; then
        echo "... and $(( $(wc -l < $COVERAGE_DIR/uncovered.txt) - 10 )) more"
    fi
else
    print_success "No completely uncovered functions found"
fi

# Run benchmarks if requested
if [ "$RUN_BENCHMARKS" = "true" ]; then
    print_header "Running Benchmarks"
    go test -bench=. -benchmem -run=^$ ./... 2>&1 | tee $COVERAGE_DIR/benchmarks.txt
    print_success "Benchmarks completed: $COVERAGE_DIR/benchmarks.txt"
fi

# Generate final report
print_header "Test Coverage Report Summary"
cat > $COVERAGE_DIR/summary.txt << EOF
GAUTH SERVICE TEST COVERAGE REPORT
Generated: $(date)
Go Version: $GO_VERSION

COVERAGE METRICS:
- Total Coverage: ${COVERAGE_PERCENT}%
- Minimum Required: ${MIN_COVERAGE}%
- Status: $(if [ "$COVERAGE_FAILED" = "true" ]; then echo "FAILED"; else echo "PASSED"; fi)

REPORTS GENERATED:
- HTML Report: $COVERAGE_DIR/$COVERAGE_HTML
- XML Report: $COVERAGE_DIR/$COVERAGE_XML
- JSON Report: $COVERAGE_DIR/coverage.json
- Uncovered Lines: $COVERAGE_DIR/uncovered.txt

TEST LOGS:
- Unit Tests: $COVERAGE_DIR/unit-tests.log
$([ -f $COVERAGE_DIR/integration-tests.log ] && echo "- Integration Tests: $COVERAGE_DIR/integration-tests.log")
$([ -f $COVERAGE_DIR/e2e-tests.log ] && echo "- E2E Tests: $COVERAGE_DIR/e2e-tests.log")

RECOMMENDATIONS:
$(if [ "$COVERAGE_FAILED" = "true" ]; then
    echo "- Increase test coverage by adding more unit tests"
    echo "- Focus on packages with low coverage"
    echo "- Add integration tests for database operations"
    echo "- Add E2E tests for complete workflows"
else
    echo "- Coverage goals met ✓"
    echo "- Consider adding more edge case tests"
    echo "- Maintain current coverage levels"
fi)
EOF

print_success "Summary report generated: $COVERAGE_DIR/summary.txt"

# Open HTML report if in interactive mode and requested
if [ "$OPEN_REPORT" = "true" ] && [ -t 1 ]; then
    print_header "Opening Coverage Report"
    if command -v open &> /dev/null; then
        open $COVERAGE_DIR/$COVERAGE_HTML
    elif command -v xdg-open &> /dev/null; then
        xdg-open $COVERAGE_DIR/$COVERAGE_HTML
    else
        print_warning "Cannot open browser automatically. Open $COVERAGE_DIR/$COVERAGE_HTML manually."
    fi
fi

print_header "Coverage Analysis Complete"

if [ "$COVERAGE_FAILED" = "true" ]; then
    print_error "Coverage below minimum threshold"
    exit 1
else
    print_success "All coverage requirements met!"
fi

echo ""
echo "📁 All reports saved in: $COVERAGE_DIR/"
echo "🌐 View HTML report: file://$(pwd)/$COVERAGE_DIR/$COVERAGE_HTML"
echo "📊 Coverage: ${COVERAGE_PERCENT}%"
echo ""
