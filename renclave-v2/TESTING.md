# 🧪 Renclave Testing Guide

This document provides comprehensive information about testing the renclave project, including unit tests, integration tests, and end-to-end tests.

## 📋 Test Overview

The renclave project includes three types of tests:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test the complete system

## 🚀 Quick Start

### Run All Tests
```bash
# From project root
./scripts/run-tests.sh

# Or with coverage
./scripts/run-tests.sh --coverage
```

### Run Specific Test Types
```bash
# Unit tests only
./scripts/run-tests.sh --unit-only

# Integration tests only
./scripts/run-tests.sh --integration-only

# E2E tests only
./scripts/run-tests.sh --e2e-only
```

### CI/CD Testing
```bash
# Run CI test suite
./scripts/ci-test.sh
```

## 🧩 Unit Tests

Unit tests are embedded within each crate and test individual functions and components.

### Shared Library Tests
```bash
cargo test -p renclave-shared --lib
```

**Test Coverage:**
- EnclaveRequest/EnclaveResponse creation and serialization
- EnclaveOperation enum variants
- EnclaveResult enum variants
- HTTP request/response structures
- Error handling and conversion
- Serialization round-trips

### Enclave Tests
```bash
cargo test -p renclave-enclave --lib
```

**Test Coverage:**
- SeedGenerator initialization
- Seed generation with different strengths (128, 160, 192, 224, 256 bits)
- Seed validation (valid and invalid seeds)
- Passphrase handling
- Entropy generation and verification
- Strength validation
- Concurrent operations
- Error handling

### Host API Tests
```bash
cargo test -p renclave-host --lib
```

**Test Coverage:**
- Health check endpoint
- Service information endpoint
- Seed generation endpoint
- Seed validation endpoint
- Network status endpoint
- Error handling
- Request validation
- Response structures

## 🔗 Integration Tests

Integration tests verify that different components work together correctly.

### Running Integration Tests
```bash
cargo test --test integration_tests
```

**Test Coverage:**
- Complete seed generation flow
- Seed validation with generated seeds
- Entropy consistency across multiple generations
- Passphrase integration
- Error handling scenarios
- Concurrent seed generation
- Strength consistency validation
- Serialization round-trips

## 🌐 End-to-End Tests

E2E tests verify the complete system from HTTP API to enclave and back.

### Prerequisites
- Enclave service running (creates `/tmp/enclave.sock`)
- Host service running (listens on port 3000)

### Running E2E Tests
```bash
cargo test --test e2e_tests
```

**Test Coverage:**
- System startup and service readiness
- Health endpoint functionality
- Service information endpoint
- Network status endpoint
- Seed generation endpoint (all strengths)
- Seed generation with passphrases
- Seed validation endpoint
- Error handling
- Concurrent requests
- Performance testing
- System stability
- Service cleanup

## 📊 Test Configuration

### Test Dependencies
The following dependencies are added to each crate for testing:

```toml
[dev-dependencies]
tokio = { workspace = true, features = ["test-util"] }
tokio-test = { workspace = true }
reqwest = { workspace = true, features = ["json"] }
```

### Test Timeouts
- **Unit Tests**: No timeout (fast execution)
- **Integration Tests**: No timeout (fast execution)
- **E2E Tests**: 30 seconds per test
- **CI Tests**: 10 minutes total

## 🛠️ Test Utilities

### TestUtils Class (E2E Tests)
The E2E tests include a `TestUtils` class that provides:

- Service readiness checking
- HTTP request/response handling
- Enclave socket verification
- Configurable timeouts and retries

### Mock Structures (Unit Tests)
Unit tests use mock implementations for:

- `MockEnclaveClient` - Simulates enclave responses
- `MockNetworkManager` - Simulates network status

## 📈 Coverage and Benchmarks

### Generate Coverage Report
```bash
./scripts/run-tests.sh --coverage
```

Coverage reports are generated using `grcov` and saved to `coverage/` directory.

### Run Benchmarks
```bash
./scripts/run-tests.sh --benchmark
```

Benchmark results are saved to `test-results/benchmarks.log`.

## 🔧 Test Scripts

### Main Test Runner (`run-tests.sh`)
Comprehensive test runner with options for:
- Selective test execution
- Coverage generation
- Benchmark execution
- Detailed logging
- Service management

**Options:**
- `--unit-only` - Run only unit tests
- `--integration-only` - Run only integration tests
- `--e2e-only` - Run only E2E tests
- `--coverage` - Generate coverage report
- `--benchmark` - Run benchmarks
- `--help` - Show help

### CI Test Runner (`ci-test.sh`)
Simplified runner for continuous integration:
- Code formatting checks
- Clippy linting
- Unit and integration tests
- Build verification

## 📁 Test Structure

```
renclave-v2/
├── src/
│   ├── shared/src/lib.rs          # Unit tests embedded
│   ├── enclave/src/seed_generator.rs  # Unit tests embedded
│   └── host/src/api_handlers.rs   # Unit tests embedded
├── tests/
│   ├── integration_tests.rs        # Integration tests
│   └── e2e_tests.rs               # End-to-end tests
├── scripts/
│   ├── run-tests.sh               # Main test runner
│   └── ci-test.sh                 # CI test runner
├── test-results/                  # Test output and logs
└── coverage/                      # Coverage reports
```

## 🚨 Troubleshooting

### Common Issues

1. **Enclave Socket Not Found**
   ```bash
   # Check if enclave service is running
   ls -la /tmp/enclave.sock
   
   # Start enclave service
   cargo run -p renclave-enclave
   ```

2. **Port Already in Use**
   ```bash
   # Check what's using port 3000
   lsof -i :3000
   
   # Kill process or change port
   ```

3. **Permission Denied**
   ```bash
   # Make scripts executable
   chmod +x scripts/*.sh
   ```

4. **Test Timeouts**
   - Increase timeout in test configuration
   - Check system resources
   - Verify service responsiveness

### Debug Commands
```bash
# View detailed test output
cargo test -- --nocapture

# Run specific test
cargo test test_name

# Run tests with logging
RUST_LOG=debug cargo test

# Check test compilation
cargo test --no-run
```

## 📊 Test Results

### Test Summary
After running tests, a summary is generated in `test-results/test-summary.txt`:

```
🧪 Renclave Test Summary
=========================

Test Run: 2025-01-20 10:30:00
Project: renclave-v2

Test Results:
-------------
✅ Unit Tests: 45 passed, 0 failed
✅ Integration Tests: 12 passed, 0 failed
✅ E2E Tests: 8 passed, 0 failed

Coverage:
---------
📊 Coverage report generated in: coverage/

Test Artifacts:
---------------
📁 Test results: test-results
📁 Coverage: coverage
```

### Log Files
- `test-results/unit-*.log` - Individual crate unit test results
- `test-results/integration-tests.log` - Integration test results
- `test-results/e2e-tests.log` - E2E test results
- `test-results/enclave.log` - Enclave service logs
- `test-results/host.log` - Host service logs

## 🔄 Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      - name: Run tests
        run: ./scripts/ci-test.sh
```

### GitLab CI Example
```yaml
test:
  stage: test
  image: rust:latest
  script:
    - ./scripts/ci-test.sh
  artifacts:
    paths:
      - test-results/
      - coverage/
```

## 📚 Additional Resources

- [Rust Testing Book](https://doc.rust-lang.org/book/ch11-00-testing.html)
- [Tokio Testing Guide](https://tokio.rs/tokio/tutorial/testing)
- [Axum Testing](https://docs.rs/axum/latest/axum/testing/index.html)
- [Cargo Test Documentation](https://doc.rust-lang.org/cargo/commands/cargo-test.html)

## 🤝 Contributing

When adding new tests:

1. **Unit Tests**: Add to the appropriate `#[cfg(test)]` module
2. **Integration Tests**: Add to `tests/integration_tests.rs`
3. **E2E Tests**: Add to `tests/e2e_tests.rs`
4. **Update Documentation**: Modify this file as needed

### Test Naming Convention
- Unit tests: `test_function_name_scenario`
- Integration tests: `test_component_interaction_scenario`
- E2E tests: `test_system_feature_scenario`

### Test Organization
- Group related tests in modules
- Use descriptive test names
- Include setup and teardown logic
- Handle errors gracefully
- Clean up resources after tests

