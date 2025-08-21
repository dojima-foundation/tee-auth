# gauth - Go Authentication Service

A modern, secure authentication and organization management service built in Go, inspired by [Turnkey's architecture](https://whitepaper.turnkey.com/architecture/). This service provides robust user management, policy-based authorization, and secure communication with trusted execution environments.

## 🏗️ Architecture

The service follows Turnkey's proven patterns for secure key management and authentication:

- **Organizations**: Data containers encapsulating users, policies, and cryptographic resources
- **Users & Authentication**: Multi-method authentication with API keys, passkeys, and OAuth
- **Activities & Queries**: Distinction between critical operations (activities) and read-only requests (queries)
- **Policy Engine**: Flexible authorization with quorum-based approvals
- **Secure Enclaves**: Communication with renclave-v2 for cryptographic operations

## 🚀 Features

### Core Functionality
- **Organization Management**: Create and manage isolated organizational contexts
- **User Management**: Multi-factor authentication and role-based access control
- **Policy Engine**: Flexible authorization rules with quorum support
- **Activity Tracking**: Comprehensive audit trail for all critical operations
- **Session Management**: Secure session handling with Redis-backed storage

### Security Features
- **TEE Integration**: Secure communication with renclave-v2 for seed generation
- **JWT Authentication**: Stateless authentication with configurable expiration
- **Rate Limiting**: Protection against brute force attacks
- **Audit Logging**: Comprehensive security event logging
- **Encryption**: AES encryption for sensitive data at rest

### Scalability & Reliability
- **gRPC API**: High-performance, type-safe communication
- **PostgreSQL**: ACID-compliant data storage with migrations
- **Redis Caching**: Session storage and distributed locking
- **Health Checks**: Comprehensive service health monitoring
- **Graceful Shutdown**: Clean service termination

## 📋 Requirements

- **Go 1.21+**
- **PostgreSQL 15+**
- **Redis 7+**
- **Protocol Buffers Compiler** (protoc)
- **Docker & Docker Compose** (for containerized deployment)

## 🛠️ Quick Start

### 1. Clone and Setup

```bash
# Clone the repository
cd /path/to/tee-auth/gauth

# Install dependencies
make deps

# Copy environment configuration
cp env.example .env
# Edit .env with your configuration
```

### 2. Database Setup

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Run database migrations
make migrate-up
```

### 3. Build and Run

```bash
# Generate protobuf code and build
make build-local

# Run the service
make run
```

### 4. Docker Development

```bash
# Start all services including databases
docker-compose up -d

# View logs
docker-compose logs -f gauth

# Stop services
docker-compose down
```

## 🔧 Configuration

The service uses environment variables for configuration. Key settings:

### Database
```env
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=gauth
DB_PASSWORD=your-secure-password
DB_DATABASE=gauth
```

### Redis
```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

### Authentication
```env
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
SESSION_TIMEOUT=30m
DEFAULT_QUORUM_THRESHOLD=1
```

### Renclave Integration
```env
RENCLAVE_HOST=localhost
RENCLAVE_PORT=3000
RENCLAVE_TIMEOUT=30s
```

## 📡 API Reference

The service exposes a gRPC API with the following main services:

### Organization Management
- `CreateOrganization` - Create a new organization
- `GetOrganization` - Retrieve organization details
- `UpdateOrganization` - Update organization settings
- `ListOrganizations` - List all organizations

### User Management
- `CreateUser` - Add user to organization
- `GetUser` - Retrieve user details
- `UpdateUser` - Update user information
- `ListUsers` - List organization users

### Authentication & Authorization
- `Authenticate` - Authenticate user and create session
- `Authorize` - Check permissions for activities

### Cryptographic Operations
- `RequestSeedGeneration` - Generate BIP39 seed phrases via renclave-v2
- `ValidateSeed` - Validate seed phrase format and checksum
- `GetEnclaveInfo` - Retrieve enclave status and capabilities

### System Operations
- `Health` - Service health check
- `Status` - Service status and metrics

## 🔐 Security Model

Following Turnkey's security principles:

### Trusted Components
- **Enclave Applications**: renclave-v2 for cryptographic operations
- **Quorum Sets**: Multi-signature approval mechanisms

### Untrusted Components
- **API Gateway**: Request routing and load balancing
- **Database**: Encrypted data storage
- **Cache Layer**: Session and temporary data storage

### Security Measures
- **Request Authentication**: Cryptographic signatures and JWT tokens
- **Authorization Policies**: Flexible rule-based access control
- **Audit Logging**: Comprehensive security event tracking
- **Rate Limiting**: Protection against abuse
- **Session Management**: Secure, expiring sessions

## 🧪 Testing

```bash
# Run all tests
make test

# Run short tests (unit tests only)
make test-short

# Run benchmarks
make bench

# Run linter
make lint

# Run security checks
make security-check
```

## 📦 Deployment

### Production Deployment

1. **Build Production Image**
```bash
make docker-build
```

2. **Configure Environment**
```bash
# Set production environment variables
export JWT_SECRET="your-production-secret-32-chars-min"
export ENCRYPTION_KEY="your-32-byte-production-encryption-key"
export DB_PASSWORD="your-secure-database-password"
```

3. **Deploy with Docker Compose**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes Deployment
See `kustomize/` directory for Kubernetes manifests.

## 🔄 Integration with renclave-v2

The service integrates with renclave-v2 for secure cryptographic operations:

```go
// Example: Generate seed phrase
response, err := gauthClient.RequestSeedGeneration(ctx, &pb.SeedGenerationRequest{
    OrganizationId: "org-uuid",
    UserId:         "user-uuid", 
    Strength:       256,
})
```

Communication flow:
1. gauth receives gRPC request
2. Validates user permissions and organization access
3. Forwards request to renclave-v2 HTTP API
4. Returns signed response with cryptographic proof

## 🏥 Health Monitoring

The service provides comprehensive health checks:

```bash
# Check service health
grpcurl -plaintext localhost:9090 gauth.v1.GAuthService/Health

# Check individual components
curl http://localhost:8080/health
```

Health check includes:
- Database connectivity
- Redis connectivity  
- renclave-v2 integration
- Memory and CPU usage
- Active connections

## 🤝 Contributing

1. **Development Setup**
```bash
make dev-setup
make tools
```

2. **Code Standards**
```bash
make fmt    # Format code
make vet    # Run go vet
make lint   # Run linter
```

3. **Testing**
```bash
make test   # Full test suite
make check  # Quick checks
```

## 📚 Documentation

- [API Documentation](./docs/api.md)
- [Architecture Guide](./docs/architecture.md)
- [Security Model](./docs/security.md)
- [Deployment Guide](./docs/deployment.md)

## 🔗 Related Projects

- [renclave-v2](../renclave-v2/) - Trusted execution environment for cryptographic operations
- [qos](../qos/) - Quality of Service framework for secure enclaves

## 📄 License

This project is part of the Dojima Foundation TEE-Auth suite.

## 🆘 Support

For support and questions:
- Create an issue in this repository
- Check the [troubleshooting guide](./docs/troubleshooting.md)
- Review the [FAQ](./docs/faq.md)

---

**Built with ❤️ for secure, scalable authentication in trusted execution environments.**
