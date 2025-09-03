#!/bin/bash

# Coverage generation script for Renclave
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if we're in the right directory
if [ ! -f "Cargo.toml" ]; then
    print_status $RED "❌ Error: Must run from renclave-v2 directory"
    exit 1
fi

print_status $BLUE "🔍 Generating comprehensive test coverage for Renclave..."

# Create coverage directory
mkdir -p coverage

# Check if llvm-tools-preview is installed
if ! command_exists cargo-llvm-cov; then
    print_status $YELLOW "📦 Installing llvm-tools-preview..."
    cargo install cargo-llvm-cov
fi

# Check if grcov is installed
if ! command_exists grcov; then
    print_status $YELLOW "📦 Installing grcov..."
    cargo install grcov
fi

# Clean previous coverage data
print_status $BLUE "🧹 Cleaning previous coverage data..."
rm -rf coverage/* target/llvm-cov-target

# Generate coverage for each crate
print_status $BLUE "📊 Generating coverage for shared crate..."
cargo llvm-cov --workspace --package renclave-shared --html --output-dir coverage/shared

print_status $BLUE "📊 Generating coverage for enclave crate..."
cargo llvm-cov --workspace --package renclave-enclave --html --output-dir coverage/enclave

print_status $BLUE "📊 Generating coverage for network crate..."
cargo llvm-cov --workspace --package renclave-network --html --output-dir coverage/network

# Generate overall workspace coverage
print_status $BLUE "📊 Generating overall workspace coverage..."
cargo llvm-cov --workspace --html --output-dir coverage/workspace

# Generate coverage report in different formats
print_status $BLUE "📊 Generating coverage reports in multiple formats..."

# HTML report
cargo llvm-cov --workspace --html --output-dir coverage/html

# JSON report
cargo llvm-cov --workspace --json --output-dir coverage/json

# LCOV report
cargo llvm-cov --workspace --lcov --output-dir coverage/lcov

# Generate summary
print_status $BLUE "📋 Generating coverage summary..."
cargo llvm-cov --workspace --text --output-dir coverage/text

# Create coverage index
cat > coverage/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Renclave Test Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .link { color: #0066cc; text-decoration: none; }
        .link:hover { text-decoration: underline; }
        .stats { background: #e8f4f8; padding: 15px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 Renclave Test Coverage Report</h1>
        <p>Comprehensive coverage metrics for the Renclave project</p>
    </div>
    
    <div class="section">
        <h2>📊 Coverage Reports</h2>
        <ul>
            <li><a href="html/index.html" class="link">📈 HTML Coverage Report</a></li>
            <li><a href="workspace/index.html" class="link">🏗️ Workspace Coverage</a></li>
            <li><a href="shared/index.html" class="link">🔗 Shared Crate Coverage</a></li>
            <li><a href="enclave/index.html" class="link">🔐 Enclave Crate Coverage</a></li>
            <li><a href="network/index.html" class="link">🌐 Network Crate Coverage</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>📋 Raw Data</h2>
        <ul>
            <li><a href="json/coverage.json" class="link">📄 JSON Coverage Data</a></li>
            <li><a href="lcov/lcov.info" class="link">📊 LCOV Coverage Data</a></li>
            <li><a href="text/coverage.txt" class="link">📝 Text Coverage Summary</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>📈 Coverage Statistics</h2>
        <div class="stats">
            <p><strong>Generated:</strong> $(date)</p>
            <p><strong>Project:</strong> Renclave v2</p>
            <p><strong>Coverage Tool:</strong> cargo-llvm-cov + grcov</p>
        </div>
    </div>
</body>
</html>
EOF

print_status $GREEN "✅ Coverage generation completed!"
print_status $BLUE "📁 Coverage reports available in:"
print_status $BLUE "   - HTML: coverage/html/index.html"
print_status $BLUE "   - Workspace: coverage/workspace/index.html"
print_status $BLUE "   - Summary: coverage/index.html"

# Show coverage summary
if [ -f "coverage/text/coverage.txt" ]; then
    print_status $BLUE "📊 Coverage Summary:"
    cat coverage/text/coverage.txt
fi

print_status $GREEN "🎉 Coverage generation completed successfully!"
