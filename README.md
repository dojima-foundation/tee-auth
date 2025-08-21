# TEE-Auth: Trusted Execution Environment Authentication System

TEE-Auth is a secure authentication and key management system leveraging trusted execution environments (TEEs) for cryptographic operations. It consists of two main components: `gauth` (Go Authentication Service) and `renclave-v2` (Rust-based Enclave for seed generation).

## Project Structure

- **gauth/**: The core authentication service written in Go.
- **renclave-v2/**: The secure enclave component written in Rust for generating cryptographic seeds.

## Components Overview

### Gauth (Go Authentication Service)

Gauth is a modern authentication and organization management service inspired by Turnkey's architecture. It handles:

- Organization and user management
- Policy-based authorization
- Secure session handling
- Integration with renclave-v2 for seed generation

Key features:
- gRPC API for high-performance communication
- PostgreSQL for data storage
- Redis for caching and sessions
- JWT authentication
- Rate limiting and audit logging

For detailed setup and usage, see [gauth/README.md](gauth/README.md).

### Renclave-v2 (Rust Enclave)

Renclave-v2 is a secure seed phrase generation system designed for QEMU Nitro Enclaves with TAP networking support. It provides:

- BIP39 seed phrase generation
- Seed validation
- Network connectivity testing
- Health and status endpoints

Key features:
- HTTP API via Axum
- Secure enclave isolation
- Hardware RNG for entropy
- TAP networking for external connectivity

For detailed setup and usage, see [renclave-v2/README.md](renclave-v2/README.md).

## Integration

Gauth integrates with renclave-v2 for secure cryptographic operations:
1. Gauth receives requests via gRPC (e.g., seed generation).
2. Validates permissions and forwards to renclave-v2's HTTP API.
3. Renclave-v2 generates seeds in a secure enclave.
4. Response is returned with cryptographic proofs.

This integration is configured in gauth's environment variables (RENCLAVE_HOST, RENCLAVE_PORT) and demonstrated in the docker-compose.yml.

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
   - Check gauth health: `grpcurl -plaintext localhost:9090 gauth.v1.GAuthService/Health`
   - Check renclave health: `curl http://localhost:3000/health`

## Development

- Follow subproject READMEs for component-specific development.
- Run tests: In gauth/ `make test`; In renclave-v2/ `make test`.

## License

This project is licensed under the MIT License.

For more details, refer to the individual component READMEs.
