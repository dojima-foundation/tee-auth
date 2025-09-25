# Renclave Test Suite

This directory contains comprehensive unit, integration, and end-to-end tests for the renclave system.

## Test Structure

### Unit Tests
- **Location**: `src/enclave/src/*_tests.rs`
- **Purpose**: Test individual modules in isolation
- **Coverage**: All core modules including:
  - `seed_generator_tests.rs` - Seed generation and validation
  - `quorum_tests.rs` - Quorum key generation and Shamir Secret Sharing
  - `data_encryption_tests.rs` - Data encryption and decryption
  - `tee_communication_tests.rs` - TEE-to-TEE communication

### Integration Tests
- **Location**: `tests/integration_tests.rs`
- **Purpose**: Test interactions between different modules
- **Coverage**: Component workflows and data flow

### End-to-End Tests
- **Location**: `tests/e2e_tests.rs`
- **Purpose**: Test complete system workflows
- **Coverage**: Real-world usage scenarios

### Test Utilities
- **Location**: `tests/test_runner.rs`, `tests/test_config.rs`
- **Purpose**: Test configuration and execution utilities

## Running Tests

### Run All Tests
```bash
cargo test
```

### Run Specific Test Types
```bash
# Unit tests only
cargo test --test unit_tests

# Integration tests only
cargo test --test integration_tests

# End-to-end tests only
cargo test --test e2e_tests

# Performance tests only
cargo test --test performance_tests

# Stress tests only
cargo test --test stress_tests
```

### Run Tests with Coverage
```bash
# Generate coverage report
cargo test --coverage

# View coverage in browser
open coverage/html/index.html
```

### Run Tests in Docker
```bash
# Build test image
docker build -f docker/Dockerfile.test -t renclave-tests .

# Run tests in container
docker run --rm renclave-tests
```

## Test Configuration

### Environment Variables
- `RENCLAVE_TEST_LOG_LEVEL` - Log level for tests (default: info)
- `RENCLAVE_TEST_TIMEOUT` - Test timeout in seconds (default: 300)
- `RENCLAVE_TEST_CONCURRENT` - Max concurrent tests (default: 10)
- `RENCLAVE_TEST_ITERATIONS` - Performance test iterations (default: 100)

### Test Configuration File
Tests can be configured using `tests/test_config.rs`:
- Unit test settings
- Integration test settings
- E2E test settings
- Performance test settings
- Stress test settings

## Test Categories

### 1. Unit Tests

#### Seed Generator Tests
- **test_seed_generator_creation** - Test generator initialization
- **test_generate_seed_*_bits** - Test seed generation with different strengths
- **test_generate_seed_with_passphrase** - Test passphrase handling
- **test_seed_validation_*** - Test seed validation
- **test_concurrent_seed_generation** - Test concurrent operations
- **test_entropy_quality** - Test entropy randomness
- **test_key_derivation** - Test key derivation from seeds

#### Quorum Tests
- **test_p256_pair_generation** - Test P256 key pair generation
- **test_shares_generate_*** - Test Shamir Secret Sharing
- **test_shares_reconstruct_*** - Test share reconstruction
- **test_boot_genesis_*** - Test quorum key genesis ceremony
- **test_qos_compatibility** - Test QoS compatibility

#### Data Encryption Tests
- **test_data_encryption_creation** - Test encryption service creation
- **test_p256_encrypt_pair_*** - Test encryption key pairs
- **test_basic_encryption_decryption** - Test basic encryption/decryption
- **test_encryption_different_data_sizes** - Test various data sizes
- **test_encryption_randomness** - Test encryption randomness
- **test_concurrent_encryption** - Test concurrent operations

#### TEE Communication Tests
- **test_tee_communication_manager_creation** - Test manager initialization
- **test_set_quorum_key** - Test quorum key setting
- **test_set_manifest_envelope** - Test manifest envelope handling
- **test_boot_key_forward_*** - Test boot key forwarding
- **test_export_key_request** - Test key export
- **test_inject_key_request** - Test key injection

### 2. Integration Tests

#### Component Integration
- **test_seed_generation_integration** - Seed generation workflow
- **test_quorum_key_generation_integration** - Quorum key generation workflow
- **test_data_encryption_integration** - Data encryption workflow
- **test_tee_communication_integration** - TEE communication workflow
- **test_storage_integration** - Storage operations
- **test_network_integration** - Network operations

#### Workflow Integration
- **test_complete_workflow_integration** - Complete system workflow
- **test_concurrent_operations_integration** - Concurrent operations
- **test_error_handling_integration** - Error handling
- **test_memory_usage_integration** - Memory usage
- **test_performance_integration** - Performance testing
- **test_stress_test_integration** - Stress testing

### 3. End-to-End Tests

#### Complete Workflows
- **test_complete_seed_generation_workflow** - Full seed generation
- **test_complete_quorum_key_generation_workflow** - Full quorum key generation
- **test_complete_data_encryption_workflow** - Full data encryption
- **test_complete_tee_communication_workflow** - Full TEE communication
- **test_complete_storage_workflow** - Full storage operations
- **test_complete_network_workflow** - Full network operations

#### System Scenarios
- **test_complete_system_workflow** - Complete system scenario
- **test_concurrent_workflows** - Concurrent system workflows
- **test_error_recovery_workflows** - Error recovery scenarios
- **test_memory_usage_workflows** - Memory usage scenarios
- **test_performance_workflows** - Performance scenarios
- **test_stress_test_workflows** - Stress test scenarios

## Test Data

### Test Seeds
- Valid BIP39 seeds for testing
- Invalid seeds for error testing
- Seeds with passphrases
- Seeds of different strengths (128, 256 bits)

### Test Keys
- P256 key pairs for testing
- Quorum keys for testing
- Ephemeral keys for testing
- Invalid keys for error testing

### Test Data
- Small data (1 byte)
- Medium data (1KB)
- Large data (1MB)
- Very large data (10MB)
- Binary data
- Unicode data
- Random data

## Performance Benchmarks

### Seed Generation
- **128-bit seeds**: ~1ms per seed
- **256-bit seeds**: ~2ms per seed
- **Concurrent generation**: 10x throughput improvement

### Quorum Key Generation
- **3-of-5 threshold**: ~50ms
- **5-of-10 threshold**: ~100ms
- **Share reconstruction**: ~10ms

### Data Encryption
- **1KB data**: ~5ms
- **1MB data**: ~50ms
- **10MB data**: ~500ms

### TEE Communication
- **Boot key forward**: ~100ms
- **Export key**: ~50ms
- **Inject key**: ~50ms

## Stress Testing

### Seed Generation Stress
- **1000 iterations**: ~2 seconds
- **Memory usage**: <100MB
- **CPU usage**: <50%

### Quorum Key Stress
- **100 iterations**: ~5 seconds
- **Memory usage**: <200MB
- **CPU usage**: <70%

### Data Encryption Stress
- **1000 iterations**: ~5 seconds
- **Memory usage**: <150MB
- **CPU usage**: <60%

## Error Handling

### Expected Errors
- Invalid seed strengths
- Invalid derivation paths
- Invalid encrypted data
- Network timeouts
- Storage failures

### Error Recovery
- Automatic retry mechanisms
- Graceful degradation
- Error logging and reporting
- Cleanup on failure

## Test Reports

### Coverage Reports
- HTML coverage reports in `coverage/html/`
- Line-by-line coverage analysis
- Branch coverage analysis
- Function coverage analysis

### Performance Reports
- Performance metrics in `performance-results/`
- Benchmark comparisons
- Memory usage analysis
- CPU usage analysis

### Test Reports
- Test results in `test-results/`
- Pass/fail statistics
- Duration analysis
- Error analysis

## Continuous Integration

### GitHub Actions
- Automated test execution
- Coverage reporting
- Performance benchmarking
- Security scanning

### Docker Testing
- Containerized test environment
- Isolated test execution
- Consistent test results
- Easy CI/CD integration

## Troubleshooting

### Common Issues
1. **Test timeouts**: Increase timeout values
2. **Memory issues**: Reduce concurrent tests
3. **Network issues**: Check network configuration
4. **Storage issues**: Check storage permissions

### Debug Mode
```bash
# Enable debug logging
RUST_LOG=debug cargo test

# Enable trace logging
RUST_LOG=trace cargo test

# Run single test with debug
cargo test test_name -- --nocapture
```

### Test Isolation
- Each test runs in isolation
- Cleanup after each test
- No shared state between tests
- Deterministic test results

## Contributing

### Adding New Tests
1. Create test file in appropriate directory
2. Follow naming convention: `test_*`
3. Add test documentation
4. Update test configuration
5. Run tests locally
6. Submit pull request

### Test Guidelines
- Write clear, descriptive test names
- Use appropriate assertions
- Test both success and failure cases
- Include performance considerations
- Document test purpose and scope

### Code Coverage
- Aim for >90% line coverage
- Aim for >80% branch coverage
- Test all public APIs
- Test error conditions
- Test edge cases
