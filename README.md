# TEE-Auth: Trusted Execution Environment Authentication System

TEE-Auth is a secure authentication and key management system leveraging trusted execution environments (TEEs) for cryptographic operations. It consists of two main components: `gauth` (Go Authentication Service) and `renclave-v2` (Rust-based Enclave for seed generation).

## Project Structure

- **gauth/**: The core authentication service written in Go.
- **renclave-v2/**: The secure enclave component written in Rust for generating cryptographic seeds.

## Architecture Overview

```
                          +-------------+
                          |   Client    |
                          | (gRPC/REST) |
                          +------+------+
                                 |
                                 | Requests (Auth, Seed Gen, etc.)
                                 v
                    +------------+------------+
                    |          Gauth           |
                    | (Go Service)            |
                    | - User/Org Management   |
                    | - Policy Engine         |
                    | - Session Handling      |
                    | - gRPC Server (9091)    |
                    | - REST API (8082)       |
                    +------------+------------+
                                 |
                                 | Seed Requests
                                 v
                    +------------+------------+
                    |       Renclave-v2       |
                    | (Rust Enclave)          |
                    | - Seed Generation       |
                    | - Validation            |
                    | - Secure Isolation      |
                    | - HTTP API (3000)       |
                    +------------+------------+
                       ^                 ^
                       |                 |
                       |                 | Network (TAP)
                       |                 v
+-------------+   +----+----+     +------+------+
| PostgreSQL  |   |  Redis  |     | External Net |
| (Data Store)|   | (Cache) |     | (if needed)  |
+-------------+   +---------+     +-------------+
```

## Components Overview

### Gauth (Go Authentication Service)

Gauth is a modern authentication and organization management service inspired by Turnkey's architecture. It handles:

- Organization and user management
- Policy-based authorization
- Secure session handling
- Integration with renclave-v2 for seed generation

**Key Features:**
- **Dual API Support**: gRPC (port 9091) and REST (port 8082)
- **PostgreSQL**: ACID-compliant data storage with migrations
- **Redis**: Session storage and distributed caching
- **JWT Authentication**: Stateless authentication with configurable expiration
- **Rate Limiting**: Protection against brute force attacks
- **Audit Logging**: Comprehensive security event logging
- **Encryption**: AES encryption for sensitive data at rest

**API Endpoints:**
- **gRPC**: High-performance, type-safe communication
- **REST**: HTTP/JSON interface for web clients
- **Health Checks**: Service health and status monitoring

For detailed setup and usage, see [gauth/README.md](gauth/README.md).

### Renclave-v2 (Rust Enclave)

Renclave-v2 is a secure seed phrase generation system designed for QEMU Nitro Enclaves with TAP networking support. It provides:

- BIP39 seed phrase generation
- Seed validation
- Network connectivity testing
- Health and status endpoints

**Key Features:**
- **HTTP API**: RESTful interface via Axum framework
- **Secure Enclave Isolation**: Hardware-backed security
- **Hardware RNG**: True random number generation
- **TAP Networking**: External connectivity support
- **Health Monitoring**: Service status and capabilities

For detailed setup and usage, see [renclave-v2/README.md](renclave-v2/README.md).

## API Reference

### Gauth Service APIs

#### gRPC API (Port 9091)
```bash
# Health check
grpcurl -plaintext localhost:9091 gauth.v1.GAuthService/Health

# Service status
grpcurl -plaintext localhost:9091 gauth.v1.GAuthService/Status

# Organization management
grpcurl -plaintext -d '{"name": "Test Org"}' localhost:9091 gauth.v1.GAuthService/CreateOrganization

# User management
grpcurl -plaintext -d '{"organization_id": "org-id", "name": "Test User"}' localhost:9091 gauth.v1.GAuthService/CreateUser

# Seed generation
grpcurl -plaintext -d '{"organization_id": "org-id", "strength": 256}' localhost:9091 gauth.v1.GAuthService/RequestSeedGeneration
```

#### REST API (Port 8082)
```bash
# Health check
curl http://localhost:8082/api/v1/health

# Service status
curl http://localhost:8082/api/v1/status

# Create organization
curl -X POST http://localhost:8082/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Organization"}'

# Create user
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"organization_id": "org-id", "name": "Test User"}'

# Generate seed
curl -X POST http://localhost:8082/api/v1/renclave/seed/generate \
  -H "Content-Type: application/json" \
  -d '{"organization_id": "org-id", "strength": 256}'
```

### Renclave-v2 API (Port 3000)
```bash
# Health check
curl http://localhost:3000/health

# Generate seed
curl -X POST http://localhost:3000/seed/generate \
  -H "Content-Type: application/json" \
  -d '{"strength": 256}'

# Validate seed
curl -X POST http://localhost:3000/seed/validate \
  -H "Content-Type: application/json" \
  -d '{"seed": "your seed phrase here"}'
```

## Integration

Gauth integrates with renclave-v2 for secure cryptographic operations:
1. Gauth receives requests via gRPC (e.g., seed generation).
2. Validates permissions and forwards to renclave-v2's HTTP API.
3. Renclave-v2 generates seeds in a secure enclave.
4. Response is returned with cryptographic proofs.

This integration is configured in gauth's environment variables (RENCLAVE_HOST, RENCLAVE_PORT) and demonstrated in the docker-compose.yml.

### Integration Flow

```
Client ─► Gauth ─┬─► Validate Permissions
                 │
                 ├─► Forward to Renclave-v2 (HTTP)
                 │
                 └─► Store in DB/Redis

Renclave-v2 ─► Generate Seed in Enclave
            │
            └─► Return with Proofs ─► Gauth ─► Client
```

## Quick Start

1. Clone the repository:
   ```bash
   git clone git@github.com-bhaagikenpachi:dojima-foundation/tee-auth.git
   cd tee-auth
   ```

2. Start services using Docker Compose (from gauth/):
   ```bash
   cd gauth
   docker-compose up -d
   ```

3. Test the integration:
   - Check gauth gRPC health: `grpcurl -plaintext localhost:9091 gauth.v1.GAuthService/Health`
   - Check gauth REST health: `curl http://localhost:8082/api/v1/health`
   - Check renclave health: `curl http://localhost:3000/health`

## Testing

### Comprehensive Test Suite

The project includes a complete testing framework with multiple test types:

#### Unit Tests
```bash
cd gauth
make test-unit          # Run unit tests only
make test-unit-coverage # Unit tests with coverage
```

#### Integration Tests
```bash
# Requires PostgreSQL and Redis
INTEGRATION_TESTS=true make test-integration
```

#### End-to-End Tests
```bash
# Requires running gauth service
E2E_TESTS=true make test-e2e
```

#### Full Test Suite
```bash
make test              # All tests with coverage
make test-coverage     # Detailed coverage report
make test-coverage-open # Coverage with browser
```

#### Test Coverage
- **Models Package**: 100% coverage ✅
- **RenclaveClient**: 92.6% coverage ✅
- **Service Layer**: 23.5% (improving)
- **Overall Coverage**: 11.8% (with comprehensive integration tests)

### Test Architecture

```
Unit Tests ──► Fast, isolated, mocked dependencies
     │
     ├─► Integration Tests ──► Real databases, real network calls
     │
     └─► E2E Tests ──► Complete service integration
```

### Test Helpers and Utilities

- **Database Setup**: Automatic PostgreSQL test containers
- **Redis Setup**: In-memory Redis for testing
- **Mock Servers**: HTTP client testing with mock responses
- **Test Data**: Comprehensive test data generation
- **Performance Benchmarks**: Regression testing

### CI/CD Integration

- **GitHub Actions**: Automated testing on every push/PR
- **Coverage Reporting**: Codecov integration with PR comments
- **Performance Monitoring**: Benchmark execution tracking
- **Quality Gates**: 80% coverage threshold enforcement

## Development

### Prerequisites
- **Go 1.21+** for gauth service
- **Rust 1.70+** for renclave-v2
- **PostgreSQL 15+** for data storage
- **Redis 7+** for caching
- **Docker & Docker Compose** for containerized development

### Development Setup
```bash
# Setup gauth service
cd gauth
make deps              # Install Go dependencies
make proto             # Generate protobuf code
make build-local       # Build for local development

# Setup renclave-v2
cd ../renclave-v2
make build             # Build Rust components
```

### Environment Configuration
```bash
# Copy environment template
cp gauth/env.example gauth/.env

# Configure required variables
JWT_SECRET=your-secure-jwt-secret-32-chars-min
ENCRYPTION_KEY=your-32-byte-encryption-key
DB_PASSWORD=your-database-password
```

### Running Services
```bash
# Start all services with Docker Compose
cd gauth
docker-compose up -d

# Or run individually
make run              # Run gauth service locally
cd ../renclave-v2
make run-host         # Run renclave-v2 locally
```

## API Testing

### Postman Collections

The project includes ready-to-use Postman collections:

- **REST API Collection**: `gauth/postman/gauth-rest-api.postman_collection.json`
- **REST API Environment**: `gauth/postman/gauth-rest-api.postman_environment.json`

### Command Line Testing

```bash
# gRPC testing with grpcurl
grpcurl -plaintext localhost:9091 list
grpcurl -plaintext localhost:9091 describe gauth.v1.GAuthService

# REST API testing with curl
curl -X GET http://localhost:8082/api/v1/health
curl -X POST http://localhost:8082/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Organization"}'
```

## Monitoring and Health Checks

### Service Health
- **Gauth gRPC**: `grpcurl -plaintext localhost:9091 gauth.v1.GAuthService/Health`
- **Gauth REST**: `curl http://localhost:8082/api/v1/health`
- **Renclave-v2**: `curl http://localhost:3000/health`

### Service Status
- **Gauth gRPC**: `grpcurl -plaintext localhost:9091 gauth.v1.GAuthService/Status`
- **Gauth REST**: `curl http://localhost:8082/api/v1/status`

### Database Health
```bash
# PostgreSQL
docker exec -it gauth-postgres-1 pg_isready

# Redis
docker exec -it gauth-redis-1 redis-cli ping
```

## Security Features

### Authentication & Authorization
- **JWT Tokens**: Stateless authentication with configurable expiration
- **API Keys**: Secure API key management
- **Rate Limiting**: Protection against abuse and brute force attacks
- **Session Management**: Secure session handling with Redis

### Data Protection
- **Encryption**: AES encryption for sensitive data at rest
- **Secure Communication**: TLS/SSL for all external communications
- **Audit Logging**: Comprehensive security event tracking
- **Input Validation**: Strict input sanitization and validation

### Enclave Security
- **Hardware Isolation**: Secure enclave execution environment
- **Cryptographic Operations**: Hardware-backed cryptographic functions
- **Attestation**: Cryptographic proof of enclave integrity
- **Secure Networking**: TAP networking for external connectivity

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check what's using a port
   lsof -i :8082
   lsof -i :9091
   lsof -i :3000
   ```

2. **Database Connection Issues**
   ```bash
   # Check PostgreSQL status
   docker-compose logs postgres
   
   # Run migrations manually
   make migrate-up
   ```

3. **Service Startup Issues**
   ```bash
   # Check service logs
   docker-compose logs gauth
   docker-compose logs renclave-v2
   
   # Validate configuration
   make validate-config
   ```

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
make run
```

## License

This project is licensed under the MIT License.

For more details, refer to the individual component READMEs:
- [gauth/README.md](gauth/README.md) - Go authentication service
- [renclave-v2/README.md](renclave-v2/README.md) - Rust enclave component
