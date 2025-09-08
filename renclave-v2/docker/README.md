# Docker Testing Setup for Renclave-v2

This directory contains the Docker configuration for testing the renclave-v2 project.

## Quick Start

### Option 1: Automated Build & Test (Recommended)
```bash
# Build the Docker image
./docker/build.sh

# Run integration tests
docker compose -f docker/docker-compose.test.yml run --rm test-runner /app/scripts/run-tests-docker.sh --integration

# Clean up
docker compose -f docker/docker-compose.test.yml down --volumes --remove-orphans
```

### Option 2: Manual Docker Build
```bash
# Build image directly (from project root)
docker build -f renclave-v2/docker/Dockerfile.test -t renclave-test-runner:latest renclave-v2/

# Run tests
docker run --rm -v $(pwd)/renclave-v2:/app renclave-test-runner /app/scripts/run-tests-docker.sh --integration
```

## Files Overview

- `Dockerfile.test` - Multi-stage Dockerfile for building and testing
- `docker-compose.test.yml` - Docker Compose configuration for test containers
- `build.sh` - Automated build script that works around Docker Buildx issues

## Troubleshooting

### Docker Buildx Issues
If you encounter "load local bake definitions" errors:
1. Use the `build.sh` script instead of `docker compose build`
2. Or use `DOCKER_BUILDKIT=0 docker compose build`

### Docker Registry Pull Issues
If you get "pull access denied" errors:
1. Ensure the image exists locally: `docker images | grep renclave`
2. Use `pull_policy: never` in docker-compose.yml (already configured)
3. Rebuild the image: `./docker/build.sh`

### Disk Space Issues
```bash
# Clean up Docker
docker system prune -f

# Check disk usage
docker system df
```

## Test Results

The integration tests include:
- ✅ Seed generation and validation
- ✅ Network connectivity tests
- ✅ Serialization/deserialization
- ✅ Error handling
- ✅ Concurrent operations

All tests pass successfully with the Docker configuration.
