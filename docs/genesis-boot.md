# Genesis Boot Process

The Genesis Boot is the fundamental initialization process for Trusted Execution Environment (TEE) instances in the Renclave system. It establishes the cryptographic foundation and quorum-based security model for secure operations.

## Overview

Genesis Boot creates a **quorum-based key management system** where a master secret is split into multiple shares using Shamir Secret Sharing (SSS). The system requires a configurable threshold of shares to reconstruct the master secret, providing both security and fault tolerance.

## Key Concepts

### Quorum Configuration
- **Members**: Total number of participants in the quorum
- **Threshold**: Minimum number of shares required to reconstruct the secret
- **Examples**: 3-out-of-5, 7-out-of-7, 2-out-of-3

### Genesis Set
The Genesis Set contains:
- **Manifest Members**: Participants who will receive the manifest (TEE configuration)
- **Share Members**: Participants who will receive encrypted shares of the master secret

## Genesis Boot Request Structure

```json
{
  "type": "ACTIVITY_TYPE_GENESIS_BOOT_V2",
  "timestampMs": "1703123456789",
  "organizationId": "your-org-id",
  "parameters": {
    "namespace_name": "qos-namespace",
    "namespace_nonce": 12345,
    "pivot": {
      "hash": [1, 2, 3, ..., 32],
      "args": ["arg1", "arg2"]
    },
    "manifest_members": [
      {
        "member_alias": "member1",
        "pub_key": [/* P256 public key bytes */]
      }
    ],
    "share_members": [
      {
        "member_alias": "member1", 
        "pub_key": [/* P256 public key bytes */]
      }
    ],
    "threshold": 3
  }
}
```

## Step-by-Step Genesis Boot Process

### 1. Container Startup
```bash
# Start TEE container
docker run -d \
  --name renclave-v2-tee1 \
  --network docker_renclave-net \
  -p 9000:8080 \
  -e RUST_LOG=debug \
  --privileged \
  docker-renclave-v2
```

### 2. Generate Dynamic Keys
Use the genesis key generator tool to create valid P256 public keys:

```bash
# Generate keys for 7-threshold configuration
cargo run --bin genesis_key_generator -- \
  --member-count 7 \
  --threshold 7 \
  --output /tmp/genesis_request.json
```

### 3. Perform Genesis Boot
```bash
curl -X POST http://localhost:9000/enclave/genesis-boot \
  -H "Content-Type: application/json" \
  -d @/tmp/genesis_request.json
```

### 4. Response Analysis
The Genesis Boot response contains:

```json
{
  "id": "request-uuid",
  "result": {
    "GenesisBootResponse": {
      "manifest_envelope": {
        "manifest": {
          "namespace": "qos-namespace",
          "nonce": 12345,
          "enclave": {
            "pcr0": [/* PCR values */],
            "pcr1": [/* PCR values */],
            "pcr2": [/* PCR values */],
            "pcr3": [/* PCR values */]
          },
          "quorum_public_key": [/* 32-byte public key */]
        }
      },
      "encrypted_shares": [
        {
          "member_alias": "member1",
          "encrypted_quorum_key_share": [/* encrypted share bytes */]
        }
      ]
    }
  }
}
```

## Share Injection Process

After Genesis Boot, shares must be injected to reconstruct the quorum key:

### 1. Extract Encrypted Shares
```bash
SHARE1=$(echo "$GENESIS_RESPONSE" | jq -r '.encrypted_shares[0].encrypted_quorum_key_share')
SHARE2=$(echo "$GENESIS_RESPONSE" | jq -r '.encrypted_shares[1].encrypted_quorum_key_share')
# ... extract all shares
```

### 2. Inject Shares
```bash
curl -X POST http://localhost:9000/enclave/inject-shares \
  -H "Content-Type: application/json" \
  -d '{
    "namespace_name": "qos-namespace",
    "namespace_nonce": 12345,
    "shares": [
      {
        "member_alias": "member1",
        "decrypted_share": '$SHARE1'
      }
    ]
  }'
```

### 3. Verify Key Reconstruction
The system verifies that the reconstructed quorum key matches the one stored in the manifest:

```rust
// In InjectShares handler
if quorum_public_key == *stored_quorum_key {
    info!("✅ KEY MATCH: Reconstructed key matches stored manifest key!");
} else {
    error!("❌ KEY MISMATCH: Reconstructed key does not match!");
}
```

## Threshold Configurations

### 2-out-of-3 Threshold
- **Use Case**: Basic fault tolerance
- **Security**: Good balance of security and availability
- **Members**: 3 total, 2 required

### 7-out-of-7 Threshold  
- **Use Case**: Maximum security
- **Security**: Highest possible, no fault tolerance
- **Members**: 7 total, 7 required
- **Example**: High-value transactions, root keys

### 3-out-of-5 Threshold
- **Use Case**: Enterprise deployments
- **Security**: High security with some fault tolerance
- **Members**: 5 total, 3 required

## Cryptographic Details

### Shamir Secret Sharing (SSS)
- **Algorithm**: Uses `vsss-rs` crate (same as QoS)
- **Secret Size**: 32 bytes (256-bit)
- **Share Format**: Each share includes x-coordinate and y-value
- **Reconstruction**: Requires minimum threshold shares

### Key Generation Process
1. **Master Secret**: 32-byte random value generated
2. **Polynomial Creation**: Degree (threshold-1) polynomial
3. **Share Generation**: Evaluate polynomial at x-coordinates
4. **Encryption**: Each share encrypted with member's public key

### Verification Process
1. **Share Validation**: Verify share format and member association
2. **Decryption**: Decrypt shares using member private keys
3. **Reconstruction**: Use SSS to reconstruct master secret
4. **Key Matching**: Verify reconstructed key matches manifest key

## Error Handling

### Common Genesis Boot Errors
- **Invalid Public Keys**: Ensure keys are valid P256 format
- **Threshold Validation**: Threshold must be ≤ number of members
- **Namespace Conflicts**: Use unique namespace names and nonces

### Share Injection Errors
- **Key Mismatch**: Reconstructed key doesn't match manifest
- **Insufficient Shares**: Not enough shares for threshold
- **Invalid Format**: Share format doesn't match expected structure

## Testing Genesis Boot

### Single Instance Test
```bash
# Complete single TEE setup
./test_complete_tee_flow.sh
```

### Multi-Instance Test
```bash
# Test TEE-to-TEE communication
./test_tee_to_tee_key_sharing.sh
```

### Threshold Testing
```bash
# Test different threshold configurations
./test_threshold_configurations.sh
```

## Security Considerations

### Manifest Security
- **Integrity**: Manifest contains cryptographic hashes for verification
- **Authenticity**: Attestation documents prove TEE identity
- **Confidentiality**: Sensitive data encrypted with quorum key

### Share Security
- **Encryption**: Each share encrypted with member's public key
- **Distribution**: Shares distributed securely to authorized members
- **Storage**: Shares stored encrypted until needed for reconstruction

### Quorum Key Security
- **Generation**: Cryptographically secure random generation
- **Storage**: Stored securely within TEE memory
- **Usage**: Used for encrypting sensitive operations and data

## Next Steps

After successful Genesis Boot:
1. **Seed Generation**: Test cryptographic seed generation
2. **TEE-to-TEE Communication**: Set up multi-instance scenarios
3. **Application Integration**: Integrate with application workflows
4. **Monitoring**: Set up monitoring and logging systems

For more details, see:
- [TEE Instance Management](./tee-instances.md)
- [Key Management](./key-management.md)
- [TEE-to-TEE Key Sharing](./tee-to-tee-sharing.md)
