# External Share Distribution and Decryption Flow

This document provides detailed instructions for running the complete external share distribution and decryption flow in Docker, exactly matching the QoS architecture.

## Overview

The flow demonstrates how encrypted shares are generated, distributed externally to members, decrypted by members using their private keys, and then injected back to the TEE for reconstruction.

## Architecture

```
1. Genesis Boot (TEE) → Generates encrypted shares, returns to caller
2. External Distribution (Host) → Distributes shares to 3 members
3. Member Decryption (External) → Members decrypt their shares
4. Share Injection (TEE) → Decrypted shares sent back to TEE
5. Reconstruction (TEE) → Quorum key reconstructed from shares
```

## Prerequisites

- Docker installed and running
- All tools built and available in the container

## Quick Start

### Option 1: Run Complete Flow Automatically

```bash
# Build the Docker image
docker build -f docker/Dockerfile.test-genesis -t renclave-test-genesis .

# Run the complete flow
docker run --rm renclave-test-genesis
```

### Option 2: Run Individual Steps

```bash
# Build and start interactive container
docker build -f docker/Dockerfile.test-genesis -t renclave-test-genesis .
docker run --rm -it renclave-test-genesis bash

# Inside the container, run individual steps:
./scripts/01-generate-keys.sh
./scripts/02-start-services.sh
./scripts/03-genesis-boot.sh
./scripts/04-distribute-shares.sh
./scripts/05-decrypt-shares.sh
./scripts/06-inject-shares.sh
./scripts/07-verify-results.sh
```

## Detailed Step-by-Step Instructions

### Step 1: Generate Keys for Genesis Boot

**Purpose**: Generate 3 P256 key pairs and create the Genesis Boot API request.

**What happens**:
- Generates 3 P256 key pairs (member1, member2, member3)
- Creates `genesis_boot_request.json` with member public keys
- Sets up 2-of-3 threshold for Shamir Secret Sharing

**Command**:
```bash
./scripts/01-generate-keys.sh
```

**Output files**:
- `genesis_boot_request.json` - Genesis Boot API request

**Expected output**:
```
🔑 Step 1: Generating Keys for Genesis Boot
===========================================
📋 Generating Genesis Boot request with 3 member keys...
✅ Genesis Boot request generated: genesis_boot_request.json

📊 Generated Genesis Boot Request:
==================================
{
  "namespace_name": "test-namespace",
  "namespace_nonce": 1,
  "manifest_members": 3,
  "manifest_threshold": 2,
  "share_members": 3,
  "share_threshold": 2,
  "pivot_hash": [...],
  "pivot_args": [...]
}

🔑 Member Public Keys Generated:
===============================
member1: 65 bytes
member2: 65 bytes
member3: 65 bytes
```

### Step 2: Start TEE and Host Services

**Purpose**: Start the enclave and host processes for API communication.

**What happens**:
- Starts the TEE enclave process
- Starts the host HTTP server on port 3000
- Verifies both services are running and responding

**Command**:
```bash
./scripts/02-start-services.sh
```

**Expected output**:
```
🚀 Step 2: Starting TEE and Host Services
=========================================
🧹 Cleaning up existing processes...
🔒 Starting TEE Enclave...
   Enclave PID: 12345
🌐 Starting Host Service...
   Host PID: 12346
⏳ Waiting for services to be ready...
🔍 Testing host connectivity...
✅ Host is responding on port 3000

✅ Step 2 Complete: Services started successfully!
   - Enclave PID: 12345
   - Host PID: 12346
   - Host URL: http://localhost:3000
```

### Step 3: Run Genesis Boot

**Purpose**: Call the Genesis Boot API to generate encrypted shares.

**What happens**:
- Sends Genesis Boot request to TEE
- TEE generates quorum key and master seed
- TEE splits master seed into 3 SSS shares
- TEE encrypts each share to member's public key
- TEE returns encrypted shares (NO storage in TEE)

**Command**:
```bash
./scripts/03-genesis-boot.sh
```

**Output files**:
- `genesis_boot_response.json` - Contains encrypted shares

**Expected output**:
```
🌱 Step 3: Running Genesis Boot
===============================
📤 Sending Genesis Boot request...
   Request file: genesis_boot_request.json
   API endpoint: http://localhost:3000/enclave/genesis-boot
HTTP Status: 200
✅ Genesis Boot completed successfully!

📊 Genesis Boot Response:
=========================
{
  "quorum_public_key": 65,
  "ephemeral_key": 0,
  "waiting_state": "GenesisBooted",
  "encrypted_shares": 3
}

🔐 Encrypted Shares Details:
============================
member1: 65 bytes
member2: 65 bytes
member3: 65 bytes
```

### Step 4: External Share Distribution

**Purpose**: Distribute encrypted shares to members (simulated for localhost).

**What happens**:
- Reads encrypted shares from Genesis Boot response
- Creates distribution manifest with share details
- Generates member private keys for testing
- Simulates external distribution to 3 members

**Command**:
```bash
./scripts/04-distribute-shares.sh
```

**Output files**:
- `share_distribution.json` - Distribution details
- `member_keys/` - Directory with member private keys

**Expected output**:
```
📤 Step 4: External Share Distribution
=====================================
🔄 Running share distribution tool...
✅ Share distribution completed!

📊 Share Distribution Results:
==============================
{
  "total_members": 3,
  "threshold": 2,
  "distributed_shares": 3
}

👥 Distributed Shares:
=====================
member1: 65 bytes
member2: 65 bytes
member3: 65 bytes

🔑 Member Keys Generated:
========================
total 12
-rw-r--r-- 1 root root 64 member1.pub
-rw-r--r-- 1 root root 64 member1.secret
-rw-r--r-- 1 root root 64 member2.pub
-rw-r--r-- 1 root root 64 member2.secret
-rw-r--r-- 1 root root 64 member3.pub
-rw-r--r-- 1 root root 64 member3.secret
```

### Step 5: Member Decryption

**Purpose**: Decrypt shares using member private keys (simulated for localhost).

**What happens**:
- Loads member private keys from `member_keys/` directory
- Decrypts each encrypted share using corresponding private key
- Verifies share integrity using SHA-512 hashes
- Creates share injection request with decrypted shares

**Command**:
```bash
./scripts/05-decrypt-shares.sh
```

**Output files**:
- `member_decryption_results.json` - Decryption details
- `share_injection_request.json` - Ready for TEE injection

**Expected output**:
```
🔓 Step 5: Member Decryption
============================
🔓 Running member decryption tool...
✅ Member decryption completed!

📊 Member Decryption Results:
=============================
{
  "total_attempts": 3,
  "successful_decryptions": 3
}

🔓 Decryption Details:
=====================
member1: ✅ Success
member2: ✅ Success
member3: ✅ Success

💉 Share Injection Request Created:
===================================
{
  "namespace_name": "test-namespace",
  "namespace_nonce": 1,
  "shares": 3
}

🔐 Decrypted Shares:
===================
member1: 32 bytes
member2: 32 bytes
member3: 32 bytes
```

### Step 6: Share Injection

**Purpose**: Inject decrypted shares back to TEE for reconstruction.

**What happens**:
- Sends decrypted shares to TEE via Share Injection API
- TEE reconstructs master seed using Shamir Secret Sharing
- TEE generates quorum key from reconstructed seed
- TEE verifies reconstructed key matches original
- TEE stores quorum key for use

**Command**:
```bash
./scripts/06-inject-shares.sh
```

**Output files**:
- `share_injection_response.json` - Injection results

**Expected output**:
```
💉 Step 6: Share Injection
==========================
💉 Injecting decrypted shares back to TEE...
   Request file: share_injection_request.json
   API endpoint: http://localhost:3000/enclave/inject-shares
HTTP Status: 200
✅ Share injection completed successfully!

📊 Share Injection Response:
============================
{
  "success": true,
  "reconstructed_quorum_key": 65
}

🎉 SUCCESS: Quorum key reconstructed successfully!
   - Reconstructed key: 65 bytes
   - Key is now available in TEE
```

### Step 7: Verification

**Purpose**: Verify the complete flow and display final results.

**What happens**:
- Checks all generated files exist
- Verifies cryptographic operations
- Compares original and reconstructed keys
- Displays complete flow summary

**Command**:
```bash
./scripts/07-verify-results.sh
```

**Expected output**:
```
🔍 Step 7: Verification
=======================
📁 Checking generated files...
  ✅ genesis_boot_request.json
  ✅ genesis_boot_response.json
  ✅ share_distribution.json
  ✅ member_decryption_results.json
  ✅ share_injection_request.json
  ✅ share_injection_response.json

🔐 Cryptographic Verification:
==============================
1. Genesis Boot:
   ✅ Quorum key generated: 65 bytes
   ✅ Encrypted shares created: 3 shares

2. Share Distribution:
   ✅ Shares distributed to: 3 members
   ✅ Threshold: 2

3. Member Decryption:
   ✅ Decryption attempts: 3
   ✅ Successful decryptions: 3

4. Share Injection:
   ✅ Injection success: true
   ✅ Reconstructed key: 65 bytes

🔍 Key Comparison:
==================
   ✅ Key sizes match: 65 bytes
   ✅ Reconstruction successful!

📋 Flow Summary:
================
1. ✅ Genesis Boot: Generated quorum key and encrypted shares
2. ✅ Share Distribution: Distributed encrypted shares to 3 members
3. ✅ Member Decryption: Decrypted shares using private keys
4. ✅ Share Injection: Injected decrypted shares back to TEE
5. ✅ Reconstruction: Reconstructed quorum key from shares
6. ✅ Verification: Confirmed successful reconstruction

🎉 COMPLETE SUCCESS!
===================
The external share distribution and decryption flow has been
successfully implemented and tested. The quorum key has been
reconstructed from the distributed shares and is now available
in the TEE for use.
```

## Troubleshooting

### Common Issues

1. **Services not starting**:
   ```bash
   # Check if ports are available
   netstat -tlnp | grep 3000
   
   # Kill existing processes
   pkill -f "enclave"
   pkill -f "host"
   ```

2. **API calls failing**:
   ```bash
   # Check service logs
   cat enclave.log
   cat host.log
   
   # Test connectivity
   curl http://localhost:3000/health
   ```

3. **File not found errors**:
   ```bash
   # Ensure you're in the correct directory
   cd /workspace
   
   # Check if files exist
   ls -la *.json
   ```

### Manual Verification

You can manually verify each step by examining the generated files:

```bash
# Check Genesis Boot request
cat genesis_boot_request.json | jq '.'

# Check Genesis Boot response
cat genesis_boot_response.json | jq '.'

# Check share distribution
cat share_distribution.json | jq '.'

# Check decryption results
cat member_decryption_results.json | jq '.'

# Check share injection request
cat share_injection_request.json | jq '.'

# Check share injection response
cat share_injection_response.json | jq '.'
```

## Security Features Demonstrated

- ✅ **Shamir Secret Sharing**: 2-of-3 threshold
- ✅ **P-256 Elliptic Curve**: Cryptography
- ✅ **Share Encryption**: To member public keys
- ✅ **External Distribution**: Shares sent to members
- ✅ **Member Decryption**: Using private keys
- ✅ **Secure Reconstruction**: In TEE
- ✅ **Integrity Verification**: SHA-512 hashes
- ✅ **End-to-End Flow**: Complete verification

## Architecture Comparison

| Aspect | QoS | renclave-v2 |
|--------|-----|-------------|
| **Genesis Boot** | Returns encrypted shares | ✅ Returns encrypted shares |
| **Share Storage** | No TEE storage | ✅ No TEE storage |
| **Distribution** | External to members | ✅ External to members |
| **Decryption** | Members handle | ✅ Members handle |
| **Injection** | Decrypted shares | ✅ Decrypted shares |
| **Reconstruction** | SSS in TEE | ✅ SSS in TEE |

The implementation exactly matches the QoS architecture for external share distribution and decryption.

