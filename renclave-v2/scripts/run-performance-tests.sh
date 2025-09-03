#!/bin/bash

# Performance testing script for Renclave
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

print_status $BLUE "🚀 Running Renclave Performance Tests and Benchmarks..."

# Create performance results directory
mkdir -p performance-results

# Check if criterion is available
if ! cargo bench --help >/dev/null 2>&1; then
    print_status $RED "❌ Error: Cargo bench not available. Please ensure criterion is properly configured."
    exit 1
fi

# Run benchmarks
print_status $BLUE "📊 Running seed generation benchmarks..."
cargo bench --bench seed_generation -- --verbose > performance-results/seed_generation.txt 2>&1

print_status $BLUE "📊 Running seed validation benchmarks..."
cargo bench --bench seed_validation -- --verbose > performance-results/seed_validation.txt 2>&1

print_status $BLUE "📊 Running concurrent operations benchmarks..."
cargo bench --bench concurrent_operations -- --verbose > performance-results/concurrent_operations.txt 2>&1

print_status $BLUE "📊 Running stress test benchmarks..."
cargo bench --bench stress_tests -- --verbose > performance-results/stress_tests.txt 2>&1

# Generate performance summary
print_status $BLUE "📋 Generating performance summary..."

cat > performance-results/performance-summary.md << 'EOF'
# Renclave Performance Test Results

## Overview
Performance test results for the Renclave project, including benchmarks and stress tests.

## Test Categories

### 1. Seed Generation Performance
- **128-bit seeds**: Basic seed generation performance
- **256-bit seeds**: Higher entropy seed generation performance  
- **With passphrase**: Seed generation with passphrase overhead

### 2. Seed Validation Performance
- **Valid seeds**: Performance of validating correct seed phrases
- **Invalid seeds**: Performance of rejecting invalid seed phrases
- **With passphrase**: Validation performance for seeds with passphrases

### 3. Concurrent Operations
- **Low concurrency (10)**: Performance with 10 simultaneous operations
- **High concurrency (50)**: Performance with 50 simultaneous operations
- **Mixed operations**: Performance of mixed generation/validation workloads

### 4. Stress Tests
- **Burst operations**: Performance under sudden high load
- **Sustained operations**: Performance under continuous load
- **Memory pressure**: Performance under memory constraints
- **Error conditions**: Performance when handling errors

## Results Files
- `seed_generation.txt`: Seed generation benchmark results
- `seed_validation.txt`: Seed validation benchmark results
- `concurrent_operations.txt`: Concurrency benchmark results
- `stress_tests.txt`: Stress test benchmark results

## Performance Metrics
- **Throughput**: Operations per second
- **Latency**: Time per operation
- **Resource Usage**: CPU and memory consumption
- **Scalability**: Performance under increasing load

## Recommendations
Based on the benchmark results, consider:
1. **Optimization opportunities**: Identify slow operations
2. **Resource scaling**: Determine optimal resource allocation
3. **Load balancing**: Optimize concurrent operation handling
4. **Error handling**: Improve error condition performance

Generated: $(date)
EOF

# Create performance dashboard
cat > performance-results/performance-dashboard.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Renclave Performance Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; text-align: center; }
        .container { max-width: 1200px; margin: 0 auto; }
        .metric-card { background: white; margin: 20px 0; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .metric-title { color: #333; font-size: 18px; font-weight: bold; margin-bottom: 15px; }
        .metric-value { font-size: 24px; color: #667eea; font-weight: bold; }
        .metric-description { color: #666; margin-top: 10px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0; }
        .file-link { color: #667eea; text-decoration: none; }
        .file-link:hover { text-decoration: underline; }
        .status { padding: 5px 10px; border-radius: 15px; font-size: 12px; font-weight: bold; }
        .status.success { background: #d4edda; color: #155724; }
        .status.warning { background: #fff3cd; color: #856404; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Renclave Performance Dashboard</h1>
            <p>Real-time performance metrics and benchmark results</p>
        </div>
        
        <div class="grid">
            <div class="metric-card">
                <div class="metric-title">Seed Generation</div>
                <div class="metric-value">📊</div>
                <div class="metric-description">Performance metrics for seed generation operations</div>
                <a href="seed_generation.txt" class="file-link">View Results</a>
                <span class="status success">✓ Complete</span>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Seed Validation</div>
                <div class="metric-value">🔍</div>
                <div class="metric-description">Performance metrics for seed validation operations</div>
                <a href="seed_validation.txt" class="file-link">View Results</a>
                <span class="status success">✓ Complete</span>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Concurrent Operations</div>
                <div class="metric-value">⚡</div>
                <div class="metric-description">Performance under concurrent load scenarios</div>
                <a href="concurrent_operations.txt" class="file-link">View Results</a>
                <span class="status success">✓ Complete</span>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Stress Tests</div>
                <div class="metric-value">💪</div>
                <div class="metric-description">Performance under extreme load conditions</div>
                <a href="stress_tests.txt" class="file-link">View Results</a>
                <span class="status success">✓ Complete</span>
            </div>
        </div>
        
        <div class="metric-card">
            <div class="metric-title">📋 Performance Summary</div>
            <div class="metric-description">
                Comprehensive analysis of all performance metrics and recommendations for optimization.
            </div>
            <a href="performance-summary.md" class="file-link">View Summary</a>
        </div>
        
        <div class="metric-card">
            <div class="metric-title">📈 Next Steps</div>
            <div class="metric-description">
                <ul>
                    <li>Analyze benchmark results for optimization opportunities</li>
                    <li>Compare performance across different hardware configurations</li>
                    <li>Set performance baselines for regression testing</li>
                    <li>Implement performance monitoring in production</li>
                </ul>
            </div>
        </div>
    </div>
</body>
</html>
EOF

print_status $GREEN "✅ Performance tests completed successfully!"
print_status $BLUE "📁 Results available in: performance-results/"
print_status $BLUE "   - Dashboard: performance-results/performance-dashboard.html"
print_status $BLUE "   - Summary: performance-results/performance-summary.md"
print_status $BLUE "   - Raw results: performance-results/*.txt"

# Show quick summary of results
print_status $BLUE "📊 Quick Results Summary:"
if [ -f "performance-results/seed_generation.txt" ]; then
    echo "   - Seed Generation: $(grep -c "test" performance-results/seed_generation.txt || echo "0") benchmarks"
fi
if [ -f "performance-results/seed_validation.txt" ]; then
    echo "   - Seed Validation: $(grep -c "test" performance-results/seed_validation.txt || echo "0") benchmarks"
fi
if [ -f "performance-results/concurrent_operations.txt" ]; then
    echo "   - Concurrent Ops: $(grep -c "test" performance-results/concurrent_operations.txt || echo "0") benchmarks"
fi
if [ -f "performance-results/stress_tests.txt" ]; then
    echo "   - Stress Tests: $(grep -c "test" performance-results/stress_tests.txt || echo "0") benchmarks"
fi

print_status $GREEN "🎉 Performance testing completed successfully!"
