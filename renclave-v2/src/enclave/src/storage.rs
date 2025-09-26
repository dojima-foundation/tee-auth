//! Storage module for persistent data in the TEE
//!
//! This module handles storing and retrieving persistent data like ephemeral keys,
//! manifests, and quorum keys in the TEE filesystem.

use anyhow::Result;
use log::{debug, error, info};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;

use crate::quorum::P256Pair;
use renclave_shared::ManifestEnvelope;

/// Paths for persistent storage in the TEE
pub const EPHEMERAL_KEY_PATH: &str = "/tmp/renclave.ephemeral.key";
pub const MANIFEST_PATH: &str = "/tmp/renclave.manifest";
pub const QUORUM_KEY_PATH: &str = "/tmp/renclave.quorum.key";
pub const PIVOT_PATH: &str = "/tmp/renclave.pivot";
pub const SHARES_PATH: &str = "/tmp/renclave.shares";

/// Storage manager for TEE persistent data
#[derive(Clone)]
pub struct TeeStorage {
    ephemeral_key_path: String,
    manifest_path: String,
    quorum_key_path: String,
    pivot_path: String,
    shares_path: String,
}

impl TeeStorage {
    /// Create a new TEE storage manager
    pub fn new() -> Self {
        Self {
            ephemeral_key_path: EPHEMERAL_KEY_PATH.to_string(),
            manifest_path: MANIFEST_PATH.to_string(),
            quorum_key_path: QUORUM_KEY_PATH.to_string(),
            pivot_path: PIVOT_PATH.to_string(),
            shares_path: SHARES_PATH.to_string(),
        }
    }

    /// Clear all stored data (for testing purposes)
    pub fn clear_all(&self) -> Result<()> {
        info!("🧹 Clearing all TEE storage for testing");

        let paths = [
            &self.ephemeral_key_path,
            &self.manifest_path,
            &self.quorum_key_path,
            &self.pivot_path,
            &self.shares_path,
        ];

        for path in &paths {
            if Path::new(path).exists() {
                fs::remove_file(path)?;
                info!("🗑️  Removed: {}", path);
            } else {
                debug!("ℹ️  File does not exist: {}", path);
            }
        }

        info!("✅ All TEE storage cleared");
        Ok(())
    }

    /// Store the ephemeral key pair
    pub fn put_ephemeral_key(&self, pair: &P256Pair) -> Result<()> {
        info!("🔑 Storing ephemeral key to TEE storage");
        info!(
            "🔍 Checking if ephemeral key exists at: {}",
            self.ephemeral_key_path
        );

        let path_exists = Path::new(&self.ephemeral_key_path).exists();
        info!("🔍 Path exists check result: {}", path_exists);

        if path_exists {
            error!(
                "❌ Ephemeral key already exists at: {}",
                self.ephemeral_key_path
            );
            // Let's see what's in the file
            match fs::read_to_string(&self.ephemeral_key_path) {
                Ok(content) => error!("🔍 Existing file content: {}", content),
                Err(e) => error!("🔍 Could not read existing file: {}", e),
            }
            return Err(anyhow::anyhow!("Ephemeral key already exists"));
        }

        info!("✅ Ephemeral key path is clear, proceeding with storage");
        let key_data = pair.to_master_seed_hex();
        self.write_as_read_only(&self.ephemeral_key_path, key_data.as_bytes())?;

        info!("✅ Ephemeral key stored successfully");
        Ok(())
    }

    /// Retrieve the ephemeral key pair
    pub fn get_ephemeral_key(&self) -> Result<P256Pair> {
        debug!("🔍 Retrieving ephemeral key from TEE storage");

        if !Path::new(&self.ephemeral_key_path).exists() {
            return Err(anyhow::anyhow!("Ephemeral key not found"));
        }

        let key_data = fs::read_to_string(&self.ephemeral_key_path)?;
        let pair = P256Pair::from_master_seed_hex(&key_data)?;

        debug!("✅ Ephemeral key retrieved successfully");
        Ok(pair)
    }

    /// Check if ephemeral key exists
    pub fn ephemeral_key_exists(&self) -> bool {
        Path::new(&self.ephemeral_key_path).exists()
    }

    /// Store the manifest envelope
    pub fn put_manifest_envelope(&self, manifest: &ManifestEnvelope) -> Result<()> {
        info!("📋 Storing manifest envelope to TEE storage");

        let manifest_data = serde_json::to_vec(manifest)?;
        self.write_as_read_only(&self.manifest_path, &manifest_data)?;

        info!("✅ Manifest envelope stored successfully");
        Ok(())
    }

    /// Retrieve the manifest envelope
    pub fn get_manifest_envelope(&self) -> Result<ManifestEnvelope> {
        debug!("🔍 Retrieving manifest envelope from TEE storage");

        if !Path::new(&self.manifest_path).exists() {
            return Err(anyhow::anyhow!("Manifest envelope not found"));
        }

        let manifest_data = fs::read(&self.manifest_path)?;
        let manifest = serde_json::from_slice(&manifest_data)?;

        debug!("✅ Manifest envelope retrieved successfully");
        Ok(manifest)
    }

    /// Check if manifest envelope exists
    pub fn manifest_envelope_exists(&self) -> bool {
        Path::new(&self.manifest_path).exists()
    }

    /// Store the quorum key pair
    pub fn put_quorum_key(&self, pair: &P256Pair) -> Result<()> {
        info!("🔐 Storing quorum key to TEE storage");

        if Path::new(&self.quorum_key_path).exists() {
            return Err(anyhow::anyhow!("Quorum key already exists"));
        }

        let key_data = pair.to_master_seed_hex();
        self.write_as_read_only(&self.quorum_key_path, key_data.as_bytes())?;

        info!("✅ Quorum key stored successfully");
        Ok(())
    }

    /// Retrieve the quorum key pair
    pub fn get_quorum_key(&self) -> Result<P256Pair> {
        debug!("🔍 Retrieving quorum key from TEE storage");

        if !Path::new(&self.quorum_key_path).exists() {
            return Err(anyhow::anyhow!("Quorum key not found"));
        }

        let key_data = fs::read_to_string(&self.quorum_key_path)?;
        let pair = P256Pair::from_master_seed_hex(&key_data)?;

        debug!("✅ Quorum key retrieved successfully");
        Ok(pair)
    }

    /// Check if quorum key exists
    pub fn quorum_key_exists(&self) -> bool {
        Path::new(&self.quorum_key_path).exists()
    }

    /// Store the pivot binary
    pub fn put_pivot(&self, pivot: &[u8]) -> Result<()> {
        info!("📦 Storing pivot binary to TEE storage");

        if Path::new(&self.pivot_path).exists() {
            return Err(anyhow::anyhow!("Pivot binary already exists"));
        }

        self.write_as_read_only(&self.pivot_path, pivot)?;

        // Make the pivot binary executable
        let mut perms = fs::metadata(&self.pivot_path)?.permissions();
        perms.set_mode(0o755);
        fs::set_permissions(&self.pivot_path, perms)?;

        info!("✅ Pivot binary stored successfully");
        Ok(())
    }

    /// Retrieve the pivot binary
    pub fn get_pivot(&self) -> Result<Vec<u8>> {
        debug!("🔍 Retrieving pivot binary from TEE storage");

        if !Path::new(&self.pivot_path).exists() {
            return Err(anyhow::anyhow!("Pivot binary not found"));
        }

        let pivot_data = fs::read(&self.pivot_path)?;

        debug!("✅ Pivot binary retrieved successfully");
        Ok(pivot_data)
    }

    /// Check if pivot binary exists
    pub fn pivot_exists(&self) -> bool {
        Path::new(&self.pivot_path).exists()
    }

    /// Store shares for reconstruction
    pub fn put_shares(&self, shares: &[Vec<u8>]) -> Result<()> {
        debug!("💾 Storing {} shares to {}", shares.len(), self.shares_path);

        let serialized = borsh::to_vec(shares)?;
        self.write_as_read_only(&self.shares_path, &serialized)?;

        info!("✅ Shares stored successfully");
        Ok(())
    }

    /// Retrieve stored shares
    pub fn get_shares(&self) -> Result<Vec<Vec<u8>>> {
        debug!("📖 Retrieving shares from {}", self.shares_path);

        let data = fs::read(&self.shares_path)?;
        let shares: Vec<Vec<u8>> = borsh::from_slice(&data)?;

        info!("✅ Retrieved {} shares", shares.len());
        Ok(shares)
    }

    /// Check if shares exist
    pub fn shares_exists(&self) -> bool {
        Path::new(&self.shares_path).exists()
    }

    /// Rotate the ephemeral key to a new key pair
    /// This happens post-boot to protect key material encrypted to it
    pub fn rotate_ephemeral_key(&self, new_pair: &P256Pair) -> Result<()> {
        info!("🔄 Rotating ephemeral key");

        if !Path::new(&self.ephemeral_key_path).exists() {
            return Err(anyhow::anyhow!("Cannot rotate non-existent ephemeral key"));
        }

        // Remove the old ephemeral key
        fs::remove_file(&self.ephemeral_key_path)?;

        // Store the new ephemeral key
        let key_data = new_pair.to_master_seed_hex();
        self.write_as_read_only(&self.ephemeral_key_path, key_data.as_bytes())?;

        info!("✅ Ephemeral key rotated successfully");
        Ok(())
    }

    /// Write data to a file with read-only permissions
    fn write_as_read_only(&self, path: &str, data: &[u8]) -> Result<()> {
        // Write the data
        fs::write(path, data)?;

        // Set read-only permissions only for production paths
        if !path.contains("test_") {
            let mut perms = fs::metadata(path)?.permissions();
            perms.set_mode(0o444);
            fs::set_permissions(path, perms)?;
        } else {
            // For test files, ensure they have write permissions
            let mut perms = fs::metadata(path)?.permissions();
            perms.set_mode(0o644);
            fs::set_permissions(path, perms)?;
        }

        Ok(())
    }

    /// Get the current state of the TEE storage
    pub fn get_storage_state(&self) -> StorageState {
        StorageState {
            ephemeral_key_exists: self.ephemeral_key_exists(),
            manifest_exists: self.manifest_envelope_exists(),
            quorum_key_exists: self.quorum_key_exists(),
            pivot_exists: self.pivot_exists(),
            shares_exists: self.shares_exists(),
        }
    }
}

/// Current state of the TEE storage
#[derive(Debug, Clone)]
pub struct StorageState {
    pub ephemeral_key_exists: bool,
    pub manifest_exists: bool,
    pub quorum_key_exists: bool,
    pub pivot_exists: bool,
    pub shares_exists: bool,
}

impl Default for TeeStorage {
    fn default() -> Self {
        Self::new()
    }
}

impl TeeStorage {
    /// Create a new TEE storage manager for testing with temporary paths
    #[cfg(test)]
    pub fn new_for_testing() -> Self {
        use std::env;
        let temp_dir = env::temp_dir();
        let test_id = uuid::Uuid::new_v4();

        Self {
            ephemeral_key_path: temp_dir
                .join(format!("test_ephemeral_{}.key", test_id))
                .to_string_lossy()
                .to_string(),
            manifest_path: temp_dir
                .join(format!("test_manifest_{}", test_id))
                .to_string_lossy()
                .to_string(),
            quorum_key_path: temp_dir
                .join(format!("test_quorum_{}.key", test_id))
                .to_string_lossy()
                .to_string(),
            pivot_path: temp_dir
                .join(format!("test_pivot_{}", test_id))
                .to_string_lossy()
                .to_string(),
            shares_path: temp_dir
                .join(format!("test_shares_{}", test_id))
                .to_string_lossy()
                .to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::tempdir;

    #[test]
    fn test_ephemeral_key_storage() {
        let temp_dir = tempdir().unwrap();
        let storage = TeeStorage {
            ephemeral_key_path: temp_dir
                .path()
                .join("ephemeral.key")
                .to_string_lossy()
                .to_string(),
            manifest_path: temp_dir
                .path()
                .join("manifest")
                .to_string_lossy()
                .to_string(),
            quorum_key_path: temp_dir
                .path()
                .join("quorum.key")
                .to_string_lossy()
                .to_string(),
            pivot_path: temp_dir.path().join("pivot").to_string_lossy().to_string(),
            shares_path: temp_dir.path().join("shares").to_string_lossy().to_string(),
        };

        // Test storing and retrieving ephemeral key
        let pair = P256Pair::generate().unwrap();

        // Initially should not exist
        assert!(!storage.ephemeral_key_exists());

        // Store the key
        storage.put_ephemeral_key(&pair).unwrap();
        assert!(storage.ephemeral_key_exists());

        // Retrieve and verify
        let retrieved_pair = storage.get_ephemeral_key().unwrap();
        assert_eq!(
            pair.public_key().to_bytes(),
            retrieved_pair.public_key().to_bytes()
        );

        // Test that storing again fails
        let result = storage.put_ephemeral_key(&pair);
        assert!(result.is_err());
    }

    #[test]
    fn test_manifest_storage() {
        let temp_dir = tempdir().unwrap();
        let storage = TeeStorage {
            ephemeral_key_path: temp_dir
                .path()
                .join("ephemeral.key")
                .to_string_lossy()
                .to_string(),
            manifest_path: temp_dir
                .path()
                .join("manifest")
                .to_string_lossy()
                .to_string(),
            quorum_key_path: temp_dir
                .path()
                .join("quorum.key")
                .to_string_lossy()
                .to_string(),
            pivot_path: temp_dir.path().join("pivot").to_string_lossy().to_string(),
            shares_path: temp_dir.path().join("shares").to_string_lossy().to_string(),
        };

        // Create a test manifest
        let manifest = ManifestEnvelope {
            manifest: renclave_shared::Manifest {
                namespace: renclave_shared::Namespace {
                    name: "test".to_string(),
                    nonce: 1,
                    quorum_key: vec![1, 2, 3, 4],
                },
                enclave: renclave_shared::NitroConfig {
                    pcr0: vec![0; 32],
                    pcr1: vec![1; 32],
                    pcr2: vec![2; 32],
                    pcr3: vec![3; 32],
                    aws_root_certificate: vec![],
                    qos_commit: "test".to_string(),
                },
                pivot: renclave_shared::PivotConfig {
                    hash: [0; 32],
                    restart: renclave_shared::RestartPolicy::Never,
                    args: vec![],
                },
                manifest_set: renclave_shared::ManifestSet {
                    threshold: 2,
                    members: vec![],
                },
                share_set: renclave_shared::ShareSet {
                    threshold: 2,
                    members: vec![],
                },
            },
            manifest_set_approvals: vec![],
            share_set_approvals: vec![],
        };

        // Initially should not exist
        assert!(!storage.manifest_envelope_exists());

        // Store the manifest
        storage.put_manifest_envelope(&manifest).unwrap();
        assert!(storage.manifest_envelope_exists());

        // Retrieve and verify
        let retrieved_manifest = storage.get_manifest_envelope().unwrap();
        assert_eq!(
            manifest.manifest.namespace.name,
            retrieved_manifest.manifest.namespace.name
        );
    }

    #[test]
    fn test_storage_state() {
        let temp_dir = tempdir().unwrap();
        let storage = TeeStorage {
            ephemeral_key_path: temp_dir
                .path()
                .join("ephemeral.key")
                .to_string_lossy()
                .to_string(),
            manifest_path: temp_dir
                .path()
                .join("manifest")
                .to_string_lossy()
                .to_string(),
            quorum_key_path: temp_dir
                .path()
                .join("quorum.key")
                .to_string_lossy()
                .to_string(),
            pivot_path: temp_dir.path().join("pivot").to_string_lossy().to_string(),
            shares_path: temp_dir.path().join("shares").to_string_lossy().to_string(),
        };

        // Initially all should be false
        let state = storage.get_storage_state();
        assert!(!state.ephemeral_key_exists);
        assert!(!state.manifest_exists);
        assert!(!state.quorum_key_exists);
        assert!(!state.pivot_exists);

        // Store an ephemeral key
        let pair = P256Pair::generate().unwrap();
        storage.put_ephemeral_key(&pair).unwrap();

        // Check state again
        let state = storage.get_storage_state();
        assert!(state.ephemeral_key_exists);
        assert!(!state.manifest_exists);
        assert!(!state.quorum_key_exists);
        assert!(!state.pivot_exists);
    }
}
