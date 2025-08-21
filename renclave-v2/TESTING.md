# 🧪 renclave-v2 Testing Instructions

## Quick Start Tests

### 1. Automated Docker Tests
```bash
# Build and run all tests automatically
docker build -f docker/Dockerfile -t renclave-v2:latest .
docker run --rm --name renclave-test renclave-v2:latest
```

### 2. Interactive Docker Testing
```bash
# Run container with shell access
docker run -it --rm --name renclave-interactive -p 3001:3000 renclave-v2:latest /bin/bash

# Inside container, run individual tests:
# Start enclave in background
/app/bin/enclave &

# Start host in background  
/app/bin/host &

# Wait a moment for services to start
sleep 3

# Test endpoints manually
curl http://localhost:3000/health
curl http://localhost:3000/info | jq
curl http://localhost:3000/network/status | jq
curl -X POST http://localhost:3000/generate-seed \
  -H "Content-Type: application/json" \
  -d '{"strength": 256}' | jq
```

## Manual API Testing

### Health Endpoints
```bash
# Basic health check
curl -s http://localhost:3001/health

# Service information
curl -s http://localhost:3001/info | jq
```

### Seed Generation Tests
```bash
# Generate 128-bit seed (12 words)
curl -X POST http://localhost:3001/generate-seed \
  -H "Content-Type: application/json" \
  -d '{"strength": 128}' | jq

# Generate 256-bit seed (24 words)  
curl -X POST http://localhost:3001/generate-seed \
  -H "Content-Type: application/json" \
  -d '{"strength": 256}' | jq

# Generate with passphrase
curl -X POST http://localhost:3001/generate-seed \
  -H "Content-Type: application/json" \
  -d '{"strength": 192, "passphrase": "test123"}' | jq
```

### Seed Validation Tests
```bash
# Validate a good seed
curl -X POST http://localhost:3001/validate-seed \
  -H "Content-Type: application/json" \
  -d '{"seed_phrase": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"}' | jq

# Validate an invalid seed
curl -X POST http://localhost:3001/validate-seed \
  -H "Content-Type: application/json" \
  -d '{"seed_phrase": "invalid seed phrase here"}' | jq
```

### Network Status
```bash
# Check network connectivity
curl -s http://localhost:3001/network/status | jq
```

### Enclave Information
```bash
# Get enclave details
curl -s http://localhost:3001/enclave/info | jq
```

## Expected Test Results

### ✅ Successful Responses

**Health Check:**
```json
{"status": "healthy", "timestamp": "2025-08-20T09:31:15Z"}
```

**Info Response:**
```json
{
  "version": "0.1.0",
  "service": "QEMU Host API Gateway", 
  "enclave_id": "699aaf02-f09f-4828-b5d7-e11c84ea8652",
  "capabilities": ["seed_generation", "seed_validation", "network_connectivity"],
  "network_status": "connected"
}
```

**Seed Generation (256-bit):**
```json
{
  "seed_phrase": "dragon isolate abstract force pigeon cart bus act acoustic ahead sentence baby sick volume ahead city supply cup number kitchen proud exhibit wasp air",
  "entropy": "420ed003ad8a4a4607b8130200a70f888c7deb81494bd9c6b65dbd8acc9f3de0",
  "strength": 256,
  "word_count": 24
}
```

**Seed Validation (Valid):**
```json
{
  "valid": true,
  "word_count": 12
}
```

### ❌ Error Responses

**Invalid Strength:**
```json
{
  "error": "Invalid strength: 100. Must be 128, 160, 192, 224, or 256",
  "code": 400,
  "request_id": "uuid-here"
}
```

**Enclave Unavailable:**
```json
{
  "error": "Enclave communication failed: Failed to connect to enclave socket",
  "code": 503,
  "request_id": "uuid-here"
}
```

## Architecture Validation

The tests validate:
- ✅ Host-Enclave communication via Unix sockets
- ✅ HTTP API gateway functionality  
- ✅ BIP39 seed generation with multiple strengths
- ✅ Secure entropy generation
- ✅ JSON request/response serialization
- ✅ Error handling and validation
- ✅ Network status reporting
- ✅ QEMU environment detection
- ⚠️ TAP interface setup (requires privileged mode)

## Troubleshooting

### Common Issues

1. **Health check fails**: Normal in host-only mode (enclave not running)
2. **Network setup warnings**: Expected in Docker (no /dev/net/tun)
3. **Permission errors**: Run with `--privileged` for network setup
4. **Port conflicts**: Change port mapping `-p 3002:3000`

### Debug Commands
```bash
# Check running processes
ps aux | grep -E "(enclave|host)"

# Check socket status  
ls -la /tmp/enclave.sock

# View detailed logs
docker logs renclave-test

# Run with privileged mode
docker run --privileged --rm renclave-v2:latest
```

## Test Scenarios Covered

1. **Basic Functionality**: Health, info, status endpoints
2. **Seed Generation**: All supported bit strengths (128, 160, 192, 224, 256)  
3. **Seed Validation**: Valid and invalid seed phrases
4. **Error Handling**: Invalid requests, missing enclave
5. **Host-Enclave Communication**: Socket-based IPC
6. **Network Detection**: QEMU environment indicators
7. **End-to-End Flow**: Complete request-response cycle

## Performance Expectations

- Seed generation: < 100ms
- API response time: < 50ms  
- Enclave startup: < 5 seconds
- Host startup: < 3 seconds

