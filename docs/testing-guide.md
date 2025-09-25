# Testing Guide

This document provides comprehensive testing procedures for the Renclave system, including unit tests, integration tests, performance tests, and security validation.

## Testing Overview

The Renclave system includes multiple layers of testing to ensure reliability, security, and performance:

- **Unit Tests**: Component-level testing
- **Integration Tests**: End-to-end testing
- **Performance Tests**: Load and stress testing
- **Security Tests**: Security validation and penetration testing

## Test Environment Setup

### Prerequisites
```bash
# Install required tools
sudo apt-get update
sudo apt-get install -y docker.io docker-compose
sudo apt-get install -y kvm libvirt-daemon-system

# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
```

### Environment Configuration
```bash
# Set up environment variables
export RUST_LOG=debug
export TEE_TEST_MODE=true
export DOCKER_NETWORK=renclave-test-net

# Create test network
docker network create $DOCKER_NETWORK
```

## Unit Tests

### Rust Unit Tests

#### Running All Unit Tests
```bash
cd renclave-v2
cargo test
```

#### Running Specific Test Categories
```bash
# Test quorum operations
cargo test quorum

# Test seed generation
cargo test seed_generator

# Test encryption/decryption
cargo test encryption

# Test TEE communication
cargo test tee_communication
```

#### Test Coverage
```bash
# Install tarpaulin for coverage
cargo install cargo-tarpaulin

# Run tests with coverage
cargo tarpaulin --out Html --output-dir coverage/
```

### Go Unit Tests

#### Running All Unit Tests
```bash
cd gauth
go test ./...
```

#### Running with Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Integration Tests

### TEE Integration Tests

#### Single TEE Flow Test
```bash
cd renclave-v2
./test_complete_tee_flow.sh
```

**Test Steps:**
1. Start TEE container
2. Perform Genesis Boot
3. Inject shares
4. Generate seeds
5. Verify operations

#### Multi-TEE Communication Test
```bash
cd renclave-v2
./test_tee_to_tee_key_sharing.sh
```

**Test Steps:**
1. Start TEE1 and TEE2 containers
2. Complete TEE1 setup
3. Share manifest from TEE1 to TEE2
4. Generate TEE2 attestation
5. Export key from TEE1 to TEE2
6. Inject key into TEE2
7. Test TEE2 seed generation

#### TEE Chain Test
```bash
cd renclave-v2
./test_tee_chain.sh
```

**Test Steps:**
1. Start TEE1, TEE2, and TEE3 containers
2. Set up TEE1 → TEE2 → TEE3 chain
3. Verify each TEE can generate seeds
4. Test independent operation

### API Integration Tests

#### REST API Tests
```bash
cd gauth
go test -tags=integration ./test/integration/
```

#### gRPC API Tests
```bash
cd gauth
go test -tags=grpc ./test/grpc/
```

## Performance Tests

### Load Testing

#### Seed Generation Performance
```bash
cd renclave-v2/benchmarks
cargo bench seed_generation
```

**Expected Performance:**
- Single seed generation: < 100ms
- Batch seed generation (100 seeds): < 5 seconds
- Concurrent requests (10 parallel): < 200ms

#### TEE-to-TEE Communication Performance
```bash
cd renclave-v2/benchmarks
cargo bench tee_communication
```

**Expected Performance:**
- Manifest sharing: < 50ms
- Attestation generation: < 100ms
- Key export: < 150ms
- Key injection: < 100ms

### Stress Testing

#### High-Load Seed Generation
```bash
# Test script for high-load seed generation
#!/bin/bash
for i in {1..1000}; do
  curl -s -X POST http://localhost:9000/generate-seed \
    -H "Content-Type: application/json" \
    -d "{\"seed_type\": \"test-$i\", \"seed_data\": \"data-$i\"}" &
done
wait
```

#### Concurrent TEE Operations
```bash
# Test script for concurrent operations
#!/bin/bash
# Start multiple TEE instances
for port in 9000 9001 9002 9003; do
  docker run -d -p $port:8080 --name tee-$port renclave-v2 &
done

# Run concurrent operations
for i in {1..100}; do
  port=$((9000 + (i % 4)))
  curl -s -X POST http://localhost:$port/generate-seed \
    -H "Content-Type: application/json" \
    -d "{\"seed_type\": \"concurrent-$i\", \"seed_data\": \"data-$i\"}" &
done
wait
```

## Security Tests

### Cryptographic Validation

#### Key Generation Tests
```bash
cd renclave-v2
cargo test key_generation_security
```

**Test Cases:**
- Verify key randomness
- Test key uniqueness
- Validate key format
- Check key entropy

#### Encryption Tests
```bash
cd renclave-v2
cargo test encryption_security
```

**Test Cases:**
- Verify encryption strength
- Test decryption accuracy
- Validate key derivation
- Check for timing attacks

### Attestation Tests

#### Attestation Validation
```bash
cd renclave-v2
cargo test attestation_validation
```

**Test Cases:**
- Verify attestation signatures
- Test attestation content
- Validate TEE identity
- Check PCR values

### Penetration Testing

#### API Security Tests
```bash
# Test for common vulnerabilities
cd gauth
go test -tags=security ./test/security/
```

**Test Cases:**
- SQL injection attempts
- XSS prevention
- CSRF protection
- Authentication bypass
- Authorization escalation

#### Network Security Tests
```bash
# Test network security
nmap -sS -O localhost
nmap --script vuln localhost
```

## Automated Testing

### Continuous Integration

#### GitHub Actions Workflow
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Rust
      uses: actions-rs/toolchain@v1
      with:
        toolchain: stable
        
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
        
    - name: Run Rust tests
      run: |
        cd renclave-v2
        cargo test
        
    - name: Run Go tests
      run: |
        cd gauth
        go test ./...
        
    - name: Run integration tests
      run: |
        cd renclave-v2
        ./scripts/run-integration-tests.sh
        
    - name: Run security tests
      run: |
        cd renclave-v2
        cargo test --features security-tests
```

### Test Automation Scripts

#### Complete Test Suite
```bash
#!/bin/bash
# complete_test_suite.sh

set -e

echo "🧪 Running Complete Test Suite"

# Unit Tests
echo "📋 Running Unit Tests..."
cd renclave-v2
cargo test
cd ../gauth
go test ./...

# Integration Tests
echo "🔗 Running Integration Tests..."
cd ../renclave-v2
./test_complete_tee_flow.sh
./test_tee_to_tee_key_sharing.sh

# Performance Tests
echo "⚡ Running Performance Tests..."
cargo bench

# Security Tests
echo "🔒 Running Security Tests..."
cargo test --features security-tests

echo "✅ All tests completed successfully!"
```

## Test Data Management

### Test Fixtures

#### Genesis Boot Test Data
```json
{
  "test_cases": [
    {
      "name": "2-out-of-3 threshold",
      "member_count": 3,
      "threshold": 2,
      "expected_success": true
    },
    {
      "name": "7-out-of-7 threshold",
      "member_count": 7,
      "threshold": 7,
      "expected_success": true
    },
    {
      "name": "Invalid threshold",
      "member_count": 3,
      "threshold": 5,
      "expected_success": false
    }
  ]
}
```

#### Seed Generation Test Data
```json
{
  "seed_types": [
    "wallet-seed",
    "backup-seed",
    "recovery-seed",
    "test-seed"
  ],
  "seed_data": [
    "user-specific-data",
    "application-context",
    "session-identifier",
    "test-data"
  ]
}
```

### Test Database Setup
```sql
-- Test database setup
CREATE DATABASE renclave_test;
CREATE USER renclave_test WITH PASSWORD 'test_password';
GRANT ALL PRIVILEGES ON DATABASE renclave_test TO renclave_test;

-- Test tables
CREATE TABLE test_organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE test_users (
    id UUID PRIMARY KEY,
    organization_id UUID REFERENCES test_organizations(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Test Monitoring and Reporting

### Test Metrics

#### Performance Metrics
- **Test Execution Time**: Track test duration
- **Memory Usage**: Monitor memory consumption during tests
- **CPU Usage**: Track CPU utilization
- **Network Latency**: Measure network performance

#### Quality Metrics
- **Test Coverage**: Percentage of code covered by tests
- **Test Pass Rate**: Percentage of tests passing
- **Bug Detection Rate**: Number of bugs found per test cycle
- **Regression Detection**: Number of regressions caught

### Test Reporting

#### HTML Reports
```bash
# Generate HTML test reports
cd renclave-v2
cargo test -- --nocapture --test-threads=1 > test_results.html

# Generate coverage reports
cargo tarpaulin --out Html --output-dir coverage/
```

#### JSON Reports
```bash
# Generate JSON test reports
cd renclave-v2
cargo test -- --format json > test_results.json

# Generate coverage reports
cargo tarpaulin --out Json --output-dir coverage/
```

## Troubleshooting Tests

### Common Test Issues

#### Container Startup Failures
```bash
# Check Docker daemon
sudo systemctl status docker

# Check TEE hardware support
ls /dev/kvm

# Check network configuration
docker network ls
```

#### Test Timeout Issues
```bash
# Increase test timeout
export RUST_TEST_TIMEOUT=300

# Run tests with verbose output
cargo test -- --nocapture --test-threads=1
```

#### Memory Issues
```bash
# Check memory usage
free -h

# Clean up Docker resources
docker system prune -a

# Restart Docker daemon
sudo systemctl restart docker
```

### Debug Tools

#### Test Debugging
```bash
# Run tests with debug output
RUST_LOG=debug cargo test

# Run specific test with debug
RUST_LOG=debug cargo test test_name

# Attach debugger
gdb --args cargo test test_name
```

#### Container Debugging
```bash
# Check container logs
docker logs container_name

# Execute commands in container
docker exec -it container_name /bin/bash

# Check container status
docker ps -a
```

## Best Practices

### Test Development
1. **Write tests first** (TDD approach)
2. **Keep tests simple** and focused
3. **Use descriptive test names**
4. **Test edge cases** and error conditions
5. **Mock external dependencies**

### Test Maintenance
1. **Regular test updates** with code changes
2. **Remove obsolete tests** and fixtures
3. **Optimize test performance**
4. **Maintain test documentation**
5. **Monitor test metrics**

### Test Security
1. **Never use production data** in tests
2. **Sanitize test data** and inputs
3. **Use isolated test environments**
4. **Validate test results** thoroughly
5. **Protect test credentials** and keys

## Test Environment Management

### Docker Test Environment
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  renclave-test:
    build: ./renclave-v2
    ports:
      - "9000:8080"
    environment:
      - RUST_LOG=debug
      - TEE_TEST_MODE=true
    privileged: true
    devices:
      - /dev/kvm:/dev/kvm
    
  postgres-test:
    image: postgres:15
    environment:
      - POSTGRES_DB=renclave_test
      - POSTGRES_USER=test_user
      - POSTGRES_PASSWORD=test_password
    ports:
      - "5432:5432"
      
  redis-test:
    image: redis:7
    ports:
      - "6379:6379"
```

### Test Data Cleanup
```bash
#!/bin/bash
# cleanup_test_environment.sh

echo "🧹 Cleaning up test environment..."

# Stop and remove test containers
docker-compose -f docker-compose.test.yml down -v

# Remove test images
docker rmi $(docker images -q --filter "reference=*test*")

# Clean up test data
rm -rf test-data/
rm -rf coverage/
rm -rf test-results/

echo "✅ Test environment cleaned up!"
```

## Next Steps

After setting up testing:
1. **Run initial test suite** to establish baseline
2. **Set up CI/CD pipeline** for automated testing
3. **Implement test monitoring** and alerting
4. **Create test documentation** for team members
5. **Establish test maintenance** procedures

For more details, see:
- [Genesis Boot Process](./genesis-boot.md)
- [TEE Instance Management](./tee-instances.md)
- [API Reference](./api-reference.md)
- [Architecture Overview](./architecture.md)
