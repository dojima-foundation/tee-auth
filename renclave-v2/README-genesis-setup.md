# renclave-v2 Local Genesis Boot Setup

This guide provides a comprehensive, automated solution to set up the `renclave-v2` TEE instance locally, perform a genesis boot, and inject shares to bring it to an "ApplicationReady" state with a specified quorum threshold. This setup is essential for local development and testing of services that rely on `renclave-v2` for secure key management and seed generation.

## 🚀 Features

- **Fully Automated**: A single script handles the entire process from building Docker images to verifying the final state.
- **Configurable Threshold**: Sets up the TEE with a 7 out of 10 member threshold for quorum operations.
- **Dynamic Key Generation**: Automatically generates valid P256 cryptographic keys for the genesis boot.
- **Health Checks & Verification**: Includes steps to ensure the service is running and in the correct "ApplicationReady" state.
- **Seed Generation Test**: Verifies that encrypted seed generation is functional after setup.
- **Cleanup Script**: A dedicated script to easily tear down the Docker environment and remove temporary files.
- **Clear Output**: Provides colored, informative output for easy monitoring.

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- [Docker Desktop](https://www.docker.com/products/docker-desktop) (includes Docker Engine and Docker Compose)
- `jq` (a lightweight and flexible command-line JSON processor)
  - On macOS: `brew install jq`
  - On Debian/Ubuntu: `sudo apt-get install jq`
- `curl` (usually pre-installed)
- `rustup` and Rust toolchain (for building `genesis_key_generator`)

## 🛠️ Setup Instructions

Follow these steps to get your `renclave-v2` instance to an "ApplicationReady" state:

1. **Navigate to the `renclave-v2` directory:**
   ```bash
   cd /Users/luffybhaagi/dojima/tee-auth/renclave-v2
   ```

2. **Run the main setup script:**
   This script will build the Docker image, generate keys, start the `renclave-v2` service, perform the genesis boot, inject shares, and verify the setup.

   ```bash
   ./setup-renclave-genesis.sh
   ```
   The script will provide detailed output on each step. Upon successful completion, you will see a "🎉 Genesis Boot Setup Complete!" message.

## 🧪 Verification

To quickly verify that your `renclave-v2` instance is correctly set up and functional, you can run the verification script:

```bash
./test-renclave-setup.sh
```
This script performs:
- A health check on the `renclave-v2` service.
- Checks if the application status is "ApplicationReady" with a provisioned quorum key.
- Tests encrypted seed generation.
- (Optional) Attempts key derivation (expected to fail without a real seed for full decryption).

## 🗑️ Cleanup

When you are finished with your local testing, you can easily stop and remove all `renclave-v2` Docker containers, networks, and temporary files using the cleanup script:

```bash
./cleanup-renclave.sh
```

## 💡 Troubleshooting

-   **Port Conflicts**: If `renclave-v2` fails to start due to port conflicts, ensure no other services are using port `9000`. You can check with `lsof -i :9000`. The `docker-compose.test.yml` in `renclave-v2/docker` uses `9000:8080`.
-   **Docker Build Failures**: Ensure you have the Rust toolchain installed and that Docker Desktop is running.
-   **"ApplicationReady" State Not Reached**: Review the output of `setup-renclave-genesis.sh` for any errors during the genesis boot or share injection steps. You can also inspect `docker logs renclave-v2-testing` for more details.
-   **"Invalid public key bytes: signature error"**: This error during genesis boot indicates that the public keys provided in the `genesis_boot_request.json` were not in the correct P256 byte array format. The `setup-renclave-genesis.sh` script now uses `genesis_key_generator` to create valid keys.
-   **Network Overlap**: If you encounter `failed to create network docker_renclave-net: Error response from daemon: invalid pool request: Pool overlaps with other one on this address space`, it means the subnet `172.22.0.0/16` used by `renclave-net` is already in use. You might need to manually remove conflicting networks (`docker network rm <network_id>`) or adjust the subnet in `renclave-v2/docker/docker-compose.test.yml`.

## 🔧 Available Scripts

- **`setup-renclave-genesis.sh`**: Main setup script that automates the entire genesis boot process
- **`test-renclave-setup.sh`**: Verification script to test the setup
- **`cleanup-renclave.sh`**: Cleanup script to remove containers and temporary files

## 📊 Expected Output

Upon successful completion, you should see:

```
🎉 Genesis Boot Setup Complete!
==============================
✅ renclave-v2 is running on: http://localhost:9000
✅ Threshold: 7 out of 10 members
✅ Namespace: local-dev-namespace
✅ Status: ApplicationReady
✅ Quorum key: Provisioned
✅ Seed generation: Functional

🔧 Available endpoints:
   - Health: http://localhost:9000/health
   - Generate Seed: http://localhost:9000/generate-seed
   - Application Status: http://localhost:9000/enclave/application-status

🚀 renclave-v2 is ready for integration with gauth service!
```

---

This automated setup streamlines the process of getting `renclave-v2` ready for development, allowing you to focus on integrating and testing your applications.
