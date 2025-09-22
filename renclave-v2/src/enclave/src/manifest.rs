//! Manifest generation and management
//! 
//! This module handles creating and managing manifests with quorum public keys
//! for the Genesis Boot flow.

use anyhow::Result;
use log::{info, debug};
use sha2::{Digest, Sha256};

use crate::quorum::P256Public;
use renclave_shared::{
    ManifestEnvelope, Manifest, Namespace, NitroConfig, PivotConfig, 
    ManifestSet, ShareSet, RestartPolicy, QuorumMember
};

/// Generate a manifest envelope with the quorum public key
pub fn generate_manifest_with_quorum_key(
    namespace_name: &str,
    namespace_nonce: u64,
    quorum_public_key: &P256Public,
    manifest_members: Vec<QuorumMember>,
    manifest_threshold: u32,
    share_members: Vec<QuorumMember>,
    share_threshold: u32,
    pivot_hash: [u8; 32],
    pivot_args: Vec<String>,
) -> Result<ManifestEnvelope> {
    info!("📋 Generating manifest with quorum public key");

    // Create namespace with quorum key
    let namespace = Namespace {
        name: namespace_name.to_string(),
        nonce: namespace_nonce,
        quorum_key: quorum_public_key.to_bytes(),
    };

    // Create nitro config (placeholder values for now)
    let nitro_config = NitroConfig {
        pcr0: vec![0; 32], // Placeholder
        pcr1: vec![1; 32], // Placeholder
        pcr2: vec![2; 32], // Placeholder
        pcr3: vec![3; 32], // Placeholder
        aws_root_certificate: vec![], // Placeholder
        qos_commit: "renclave-v2".to_string(),
    };

    // Create pivot config
    let pivot_config = PivotConfig {
        hash: pivot_hash,
        restart: RestartPolicy::Never,
        args: pivot_args,
    };

    // Create manifest set
    let manifest_set = ManifestSet {
        threshold: manifest_threshold,
        members: manifest_members,
    };

    // Create share set
    let share_set = ShareSet {
        threshold: share_threshold,
        members: share_members,
    };

    // Create the manifest
    let manifest = Manifest {
        namespace,
        enclave: nitro_config,
        pivot: pivot_config,
        manifest_set,
        share_set,
    };

    // Create manifest envelope (no approvals initially)
    let manifest_envelope = ManifestEnvelope {
        manifest,
        manifest_set_approvals: vec![],
        share_set_approvals: vec![],
    };

    info!("✅ Manifest generated successfully");
    debug!("Manifest namespace: {}", manifest_envelope.manifest.namespace.name);
    debug!("Manifest nonce: {}", manifest_envelope.manifest.namespace.nonce);
    debug!("Quorum key: {}", hex::encode(&manifest_envelope.manifest.namespace.quorum_key));

    Ok(manifest_envelope)
}

/// Calculate the hash of a manifest for signing/verification
pub fn calculate_manifest_hash(manifest: &Manifest) -> Result<[u8; 32]> {
    // Serialize the manifest to bytes
    let manifest_bytes = serde_json::to_vec(manifest)?;
    
    // Calculate SHA-256 hash
    let mut hasher = Sha256::new();
    hasher.update(&manifest_bytes);
    let hash: [u8; 32] = hasher.finalize().into();
    
    Ok(hash)
}

/// Validate a manifest envelope
pub fn validate_manifest_envelope(manifest_envelope: &ManifestEnvelope) -> Result<()> {
    debug!("🔍 Validating manifest envelope");

    // Check that manifest set threshold is valid
    if manifest_envelope.manifest.manifest_set.threshold == 0 {
        return Err(anyhow::anyhow!("Manifest set threshold cannot be zero"));
    }

    if manifest_envelope.manifest.manifest_set.threshold as usize > manifest_envelope.manifest.manifest_set.members.len() {
        return Err(anyhow::anyhow!("Manifest set threshold cannot be greater than member count"));
    }

    // Check that share set threshold is valid
    if manifest_envelope.manifest.share_set.threshold == 0 {
        return Err(anyhow::anyhow!("Share set threshold cannot be zero"));
    }

    if manifest_envelope.manifest.share_set.threshold as usize > manifest_envelope.manifest.share_set.members.len() {
        return Err(anyhow::anyhow!("Share set threshold cannot be greater than member count"));
    }

    // Check that namespace name is not empty
    if manifest_envelope.manifest.namespace.name.is_empty() {
        return Err(anyhow::anyhow!("Namespace name cannot be empty"));
    }

    // Check that quorum key is not empty
    if manifest_envelope.manifest.namespace.quorum_key.is_empty() {
        return Err(anyhow::anyhow!("Quorum key cannot be empty"));
    }

    debug!("✅ Manifest envelope validation passed");
    Ok(())
}

/// Create a default manifest for testing/development
pub fn create_default_manifest(
    namespace_name: &str,
    quorum_public_key: &P256Public,
) -> Result<ManifestEnvelope> {
    info!("📋 Creating default manifest for development");

    // Create default members (empty for now)
    let manifest_members = vec![];
    let share_members = vec![];

    // Create default pivot hash (all zeros for development)
    let pivot_hash = [0u8; 32];

    generate_manifest_with_quorum_key(
        namespace_name,
        1, // Start with nonce 1
        quorum_public_key,
        manifest_members,
        1, // Default threshold of 1
        share_members,
        1, // Default threshold of 1
        pivot_hash,
        vec![], // No pivot args
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::quorum::P256Pair;

    #[test]
    fn test_generate_manifest_with_quorum_key() {
        let pair = P256Pair::generate().unwrap();
        let public_key = pair.public_key();

        let manifest_members = vec![
            QuorumMember {
                alias: "member1".to_string(),
                pub_key: vec![1, 2, 3, 4],
            },
            QuorumMember {
                alias: "member2".to_string(),
                pub_key: vec![5, 6, 7, 8],
            },
        ];

        let share_members = manifest_members.clone();

        let manifest = generate_manifest_with_quorum_key(
            "test-namespace",
            1,
            &public_key,
            manifest_members,
            2,
            share_members,
            2,
            [0u8; 32],
            vec!["arg1".to_string()],
        ).unwrap();

        assert_eq!(manifest.manifest.namespace.name, "test-namespace");
        assert_eq!(manifest.manifest.namespace.nonce, 1);
        assert_eq!(manifest.manifest.namespace.quorum_key, public_key.to_bytes());
        assert_eq!(manifest.manifest.manifest_set.threshold, 2);
        assert_eq!(manifest.manifest.manifest_set.members.len(), 2);
        assert_eq!(manifest.manifest.share_set.threshold, 2);
        assert_eq!(manifest.manifest.share_set.members.len(), 2);
        assert_eq!(manifest.manifest.pivot.args, vec!["arg1"]);
    }

    #[test]
    fn test_validate_manifest_envelope() {
        let pair = P256Pair::generate().unwrap();
        let public_key = pair.public_key();

        // Test valid manifest
        let valid_manifest = create_default_manifest("test", &public_key).unwrap();
        assert!(validate_manifest_envelope(&valid_manifest).is_ok());

        // Test invalid manifest with zero threshold
        let mut invalid_manifest = valid_manifest.clone();
        invalid_manifest.manifest.manifest_set.threshold = 0;
        assert!(validate_manifest_envelope(&invalid_manifest).is_err());

        // Test invalid manifest with empty namespace name
        let mut invalid_manifest = valid_manifest.clone();
        invalid_manifest.manifest.namespace.name = String::new();
        assert!(validate_manifest_envelope(&invalid_manifest).is_err());

        // Test invalid manifest with empty quorum key
        let mut invalid_manifest = valid_manifest.clone();
        invalid_manifest.manifest.namespace.quorum_key = vec![];
        assert!(validate_manifest_envelope(&invalid_manifest).is_err());
    }

    #[test]
    fn test_calculate_manifest_hash() {
        let pair = P256Pair::generate().unwrap();
        let public_key = pair.public_key();
        let manifest_envelope = create_default_manifest("test", &public_key).unwrap();

        let hash1 = calculate_manifest_hash(&manifest_envelope.manifest).unwrap();
        let hash2 = calculate_manifest_hash(&manifest_envelope.manifest).unwrap();

        // Hash should be deterministic
        assert_eq!(hash1, hash2);

        // Hash should be 32 bytes
        assert_eq!(hash1.len(), 32);
    }

    #[test]
    fn test_create_default_manifest() {
        let pair = P256Pair::generate().unwrap();
        let public_key = pair.public_key();

        let manifest = create_default_manifest("test-namespace", &public_key).unwrap();

        assert_eq!(manifest.manifest.namespace.name, "test-namespace");
        assert_eq!(manifest.manifest.namespace.nonce, 1);
        assert_eq!(manifest.manifest.namespace.quorum_key, public_key.to_bytes());
        assert_eq!(manifest.manifest.manifest_set.threshold, 1);
        assert_eq!(manifest.manifest.share_set.threshold, 1);
        assert_eq!(manifest.manifest.pivot.args.len(), 0);
    }
}


