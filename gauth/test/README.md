# gauth Service Test Suite

This directory contains comprehensive test suites for the gauth service, including unit tests, integration tests, and end-to-end tests with coverage reporting.

## 📁 Test Structure

```
test/
├── unit/                    # Unit tests (embedded in source packages)
├── integration/             # Integration tests with real dependencies
├── e2e/                     # End-to-end tests
├── testhelpers/             # Shared test utilities and helpers
└── README.md               # This file
```

## 🧪 Test Types

### Unit Tests
- **Location**: `internal/*/test.go` files
- **Purpose**: Test individual components in isolation
- **Coverage**: Models, services, clients, utilities
- **Dependencies**: Mocked
- **Run**: `make test-unit`

**Current Unit Test Coverage:**
- **Models**: 100% coverage - Complete validation, serialization, table names
- **Service Layer**: 23.5% coverage - Basic functionality, error handling
- **RenclaveClient**: 92.6% coverage - HTTP client, timeout handling, concurrent requests

### Integration Tests
- **Location**: `test/integration/`
- **Purpose**: Test components with real dependencies
- **Coverage**: Database operations, Redis operations, gRPC services
- **Dependencies**: PostgreSQL, Redis, real network calls
- **Run**: `INTEGRATION_TESTS=true make test-integration`

**Integration Test Features:**
- Database CRUD operations
- Transaction handling and rollbacks
- Redis session management and distributed locking
- gRPC service integration
- Performance benchmarks

### End-to-End Tests
- **Location**: `test/e2e/`
- **Purpose**: Test complete workflows
- **Coverage**: Full service integration, real service communication
- **Dependencies**: Running gauth service, renclave-v2 service
- **Run**: `E2E_TESTS=true make test-e2e`

**E2E Test Scenarios:**
- Complete organization and user management workflow
- Seed generation and validation end-to-end
- Concurrent operations testing
- Error scenario validation
- Performance and data consistency testing

## 🚀 Running Tests

### Quick Start
```bash
# Run all unit tests with coverage
make test-unit

# Run integration tests (requires databases)
INTEGRATION_TESTS=true make test-integration

# Run E2E tests (requires running services)
E2E_TESTS=true make test-e2e

# Run comprehensive test suite with coverage
make test
```

### Test Environment Setup

#### For Integration Tests
```bash
# Start PostgreSQL
docker run -d --name gauth-test-postgres \
  -e POSTGRES_USER=gauth \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=gauth_test \
  -p 5432:5432 postgres:15

# Start Redis
docker run -d --name gauth-test-redis \
  -p 6379:6379 redis:7

# Set environment variables
export INTEGRATION_TESTS=true
export TEST_DB_HOST=localhost
export TEST_DB_USER=gauth
export TEST_DB_PASSWORD=password
export TEST_DB_NAME=gauth_test
export TEST_REDIS_HOST=localhost

# Run tests
make test-integration
```

#### For E2E Tests
```bash
# Build and start gauth service
make build-local
export GRPC_PORT=9090
export DB_HOST=localhost
export REDIS_HOST=localhost
export RENCLAVE_HOST=localhost
export RENCLAVE_PORT=3000
./bin/gauth &

# Set E2E environment
export E2E_TESTS=true
export E2E_GAUTH_HOST=localhost
export E2E_GAUTH_PORT=9090

# Run E2E tests
make test-e2e
```

## 📊 Coverage Reporting

### Generate Coverage Reports
```bash
# Generate detailed coverage with HTML report
make test-coverage

# Generate and open coverage report in browser
make test-coverage-open

# Run comprehensive coverage analysis
./scripts/test-coverage.sh
```

### Coverage Thresholds
- **Unit Tests**: Minimum 70% coverage required
- **Total Coverage**: Target 80% coverage
- **Models**: 100% coverage achieved ✅
- **Critical Services**: 80%+ coverage target

### Coverage Reports Generated
- `coverage/coverage.html` - Interactive HTML report
- `coverage/coverage.xml` - XML format for CI/CD
- `coverage/coverage.json` - JSON format for tooling
- `coverage/summary.txt` - Text summary report

## 🏗️ Test Architecture

### Test Helpers (`testhelpers/`)
- **TestDatabase**: Database setup and cleanup utilities
- **TestRedis**: Redis setup and cleanup utilities
- **TestDataGenerator**: Random test data generation
- **Performance Measurement**: Timing and benchmarking utilities

### Mock Strategy
- **Unit Tests**: Use interfaces and dependency injection
- **Integration Tests**: Use real dependencies with test databases
- **E2E Tests**: Use real services with test configuration

### Test Data Management
- **Isolation**: Each test uses separate data/namespaces
- **Cleanup**: Automatic cleanup after each test
- **Deterministic**: Reproducible test data and results

## 🔧 Test Configuration

### Environment Variables
```bash
# Test Type Control
INTEGRATION_TESTS=true    # Enable integration tests
E2E_TESTS=true           # Enable E2E tests

# Database Configuration
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=gauth
TEST_DB_PASSWORD=password
TEST_DB_NAME=gauth_test

# Redis Configuration
TEST_REDIS_HOST=localhost
TEST_REDIS_PORT=6379
TEST_REDIS_PASSWORD=

# E2E Configuration
E2E_GAUTH_HOST=localhost
E2E_GAUTH_PORT=9090
E2E_RENCLAVE_HOST=localhost
E2E_RENCLAVE_PORT=3000

# Coverage Configuration
MIN_COVERAGE=80          # Minimum coverage threshold
OPEN_REPORT=true         # Auto-open coverage report
RUN_BENCHMARKS=true      # Include benchmark tests
```

## 🚨 CI/CD Integration

### GitHub Actions
The test suite is integrated with GitHub Actions (`.github/workflows/test.yml`):

- **Unit Tests**: Run on every push/PR
- **Integration Tests**: Run with PostgreSQL and Redis services
- **E2E Tests**: Run with mock renclave service
- **Coverage Reporting**: Upload to Codecov
- **Performance Tests**: Run benchmarks on PRs

### Test Stages
1. **Unit Tests** - Fast feedback (< 2 minutes)
2. **Integration Tests** - Database validation (< 5 minutes)
3. **E2E Tests** - Full system validation (< 10 minutes)
4. **Coverage Analysis** - Generate reports and check thresholds

## 📈 Test Metrics

### Current Test Stats
- **Total Tests**: 50+ test cases
- **Test Files**: 8 test files
- **Coverage**: 
  - Models: 100%
  - Service: 23.5%
  - Overall: 11.8%
- **Test Types**:
  - Unit: 40+ tests
  - Integration: 10+ tests
  - E2E: 5+ test scenarios

### Performance Benchmarks
- Health check: < 100ms
- User creation: < 500ms
- List operations: < 200ms
- Database operations: < 50ms per query

## 🛠️ Development Guidelines

### Adding New Tests
1. **Unit Tests**: Add alongside source code in same package
2. **Integration Tests**: Add to `test/integration/` with real dependencies
3. **E2E Tests**: Add to `test/e2e/` for complete workflows

### Test Naming Convention
```go
// Unit tests
func TestServiceName_MethodName(t *testing.T) {}
func TestServiceName_MethodName_ErrorCase(t *testing.T) {}

// Integration tests  
func TestIntegration_FeatureName(t *testing.T) {}

// E2E tests
func TestE2E_WorkflowName(t *testing.T) {}

// Benchmarks
func BenchmarkServiceName_MethodName(b *testing.B) {}
```

### Test Structure
```go
func TestFeature(t *testing.T) {
    // Arrange
    setup := setupTest(t)
    defer setup.cleanup()
    
    // Act
    result, err := setup.service.DoSomething(ctx, input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## 🔍 Debugging Tests

### Common Issues
1. **Database Connection**: Ensure PostgreSQL is running and accessible
2. **Redis Connection**: Ensure Redis is running on correct port
3. **Port Conflicts**: Check for port conflicts in E2E tests
4. **Race Conditions**: Use atomic operations for concurrent tests

### Debug Commands
```bash
# Run specific test with verbose output
go test -v -run TestSpecificTest ./internal/service/

# Run tests with race detection
go test -race ./...

# Generate CPU profile during tests
go test -cpuprofile=cpu.prof ./...

# Run tests with coverage and open report
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

## 📚 Best Practices

### Test Quality
- ✅ **Fast**: Unit tests complete in < 2 minutes
- ✅ **Isolated**: Tests don't depend on each other
- ✅ **Deterministic**: Same input produces same output
- ✅ **Comprehensive**: Cover happy path, edge cases, and errors
- ✅ **Maintainable**: Clear test names and structure

### Coverage Guidelines
- **Critical Path**: 100% coverage for security-critical code
- **Business Logic**: 90%+ coverage for core business logic
- **Utilities**: 80%+ coverage for utility functions
- **Integration Points**: 70%+ coverage for external integrations

### Performance Testing
- **Load Testing**: Test with realistic data volumes
- **Stress Testing**: Test beyond normal capacity
- **Benchmark Testing**: Track performance over time
- **Memory Testing**: Check for memory leaks

---

## 🎯 Next Steps

### Coverage Improvements
1. **Service Layer**: Increase from 23.5% to 80%+ coverage
2. **Database Layer**: Add comprehensive database operation tests
3. **gRPC Layer**: Add complete gRPC service testing
4. **Error Handling**: Test all error paths and edge cases

### Test Enhancements
1. **Property-Based Testing**: Add fuzzing for complex data structures
2. **Contract Testing**: Add API contract validation
3. **Security Testing**: Add security-focused test cases
4. **Performance Regression**: Add performance regression detection

### Automation Improvements
1. **Test Data Management**: Automated test data generation
2. **Environment Management**: Containerized test environments  
3. **Parallel Execution**: Optimize test execution speed
4. **Reporting**: Enhanced test reporting and metrics

For questions or contributions to the test suite, please refer to the main project documentation or create an issue in the repository.
