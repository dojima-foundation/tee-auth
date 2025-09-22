//! Complete Genesis Boot flow implementation
//!
//! This module implements the complete Genesis Boot flow that integrates all components:
//! 1. P256Pair::generate() (ephemeral key)
//! 2. Extract master seed (32 bytes)
//! 3. Split master seed into SSS shares
//! 4. Encrypt shares to quorum members
//! 5. Clear master seed from memory
//! 6. Store ephemeral key in TEE
//! 7. Generate manifest with quorum public key
//! 8. TEE waits for share reconstruction

use anyhow::Result;
use log::{debug, info};

use crate::manifest::generate_manifest_with_quorum_key;
use crate::quorum::{boot_genesis, GenesisOutput, GenesisSet, P256Pair, P256Public};
use crate::storage::TeeStorage;
use crate::tee_waiting::{TeeWaitingManager, WaitingConfig, WaitingState};
use renclave_shared::{ManifestEnvelope, QuorumMember};

/// Configuration for the Genesis Boot flow
#[derive(Debug, Clone)]
pub struct GenesisBootConfig {
    /// Namespace name for the manifest
    pub namespace_name: String,
    /// Namespace nonce
    pub namespace_nonce: u64,
    /// Manifest members and threshold
    pub manifest_members: Vec<QuorumMember>,
    pub manifest_threshold: u32,
    /// Share members and threshold
    pub share_members: Vec<QuorumMember>,
    pub share_threshold: u32,
    /// Pivot configuration
    pub pivot_hash: [u8; 32],
    pub pivot_args: Vec<String>,
    /// TEE waiting configuration
    pub waiting_config: WaitingConfig,
    /// Optional DR key for disaster recovery
    pub dr_key: Option<Vec<u8>>,
}

/// Result of the Genesis Boot flow
#[derive(Debug, Clone)]
pub struct GenesisBootResult {
    /// The generated quorum key (public key)
    pub quorum_public_key: Vec<u8>,
    /// The genesis output with encrypted shares
    pub genesis_output: GenesisOutput,
    /// The generated manifest envelope
    pub manifest_envelope: ManifestEnvelope,
    /// The ephemeral key that was stored
    pub ephemeral_key: Vec<u8>,
    /// Final state of the TEE waiting process
    pub waiting_state: WaitingState,
}

/// Complete Genesis Boot flow implementation
pub struct GenesisBootFlow {
    config: GenesisBootConfig,
    storage: TeeStorage,
}

impl GenesisBootFlow {
    /// Create a new Genesis Boot flow
    pub fn new(config: GenesisBootConfig) -> Self {
        Self {
            config,
            storage: TeeStorage::new(),
        }
    }

    /// Create a new Genesis Boot flow with custom storage (for testing)
    pub fn new_with_storage(config: GenesisBootConfig, storage: TeeStorage) -> Self {
        Self { config, storage }
    }

    /// Execute the complete Genesis Boot flow
    pub async fn execute(&mut self) -> Result<GenesisBootResult> {
        info!("🌱 Starting complete Genesis Boot flow");

        // Step 1: Generate ephemeral key (P256Pair::generate())
        debug!("🔑 Step 1: Generating ephemeral key...");
        let ephemeral_key = self.generate_ephemeral_key().await?;
        info!("✅ Step 1: Ephemeral key generated");

        // Step 2: Generate quorum key and perform genesis ceremony
        debug!("🎭 Step 2: Performing genesis ceremony...");
        let genesis_output = self.perform_genesis_ceremony(&ephemeral_key).await?;
        info!("✅ Step 2: Genesis ceremony completed");

        // Step 3: Store ephemeral key in TEE
        debug!("💾 Step 3: Storing ephemeral key in TEE...");
        self.storage.put_ephemeral_key(&ephemeral_key)?;
        info!("✅ Step 3: Ephemeral key stored in TEE");

        // Step 4: Generate manifest with quorum public key
        debug!("📋 Step 5: Generating manifest with quorum public key...");
        let manifest_envelope = self.generate_manifest(&genesis_output).await?;
        debug!("📋 Step 5: Storing manifest envelope...");
        self.storage.put_manifest_envelope(&manifest_envelope)?;
        info!("✅ Step 5: Manifest generated and stored");

        // Step 5: Return encrypted shares to caller (NO TEE storage)
        debug!("📤 Step 5: Returning encrypted shares to caller for external distribution...");
        info!("✅ Step 5: Encrypted shares ready for external distribution");

        let result = GenesisBootResult {
            quorum_public_key: genesis_output.quorum_key.clone(),
            genesis_output,
            manifest_envelope,
            ephemeral_key: ephemeral_key.public_key().to_bytes(),
            waiting_state: crate::tee_waiting::WaitingState::GenesisBooted,
        };

        info!(
            "🎉 Genesis Boot flow completed successfully - shares ready for external distribution"
        );
        Ok(result)
    }

    /// Step 1: Generate ephemeral key
    async fn generate_ephemeral_key(&self) -> Result<P256Pair> {
        debug!("🔑 Generating ephemeral key");
        P256Pair::generate()
    }

    /// Step 2: Perform genesis ceremony (extract master seed, split into shares, encrypt)
    async fn perform_genesis_ceremony(&self, _ephemeral_key: &P256Pair) -> Result<GenesisOutput> {
        debug!("🎭 Performing genesis ceremony");

        // Create genesis set from share members
        let genesis_set = GenesisSet {
            members: self.config.share_members.clone(),
            threshold: self.config.share_threshold,
        };

        // Perform the genesis ceremony
        let genesis_output = boot_genesis(&genesis_set, self.config.dr_key.clone()).await?;

        // Clear master seed from memory (it's already cleared in boot_genesis)
        debug!("🧹 Master seed cleared from memory");

        Ok(genesis_output)
    }

    /// Step 4: Generate manifest with quorum public key
    async fn generate_manifest(&self, genesis_output: &GenesisOutput) -> Result<ManifestEnvelope> {
        debug!("📋 Generating manifest with quorum public key");

        let quorum_public = P256Public::from_bytes(&genesis_output.quorum_key)?;

        let manifest_envelope = generate_manifest_with_quorum_key(
            &self.config.namespace_name,
            self.config.namespace_nonce,
            &quorum_public,
            self.config.manifest_members.clone(),
            self.config.manifest_threshold,
            self.config.share_members.clone(),
            self.config.share_threshold,
            self.config.pivot_hash,
            self.config.pivot_args.clone(),
        )?;

        Ok(manifest_envelope)
    }

    /// Step 5: TEE waits for share reconstruction
    async fn wait_for_share_reconstruction(&self) -> Result<WaitingState> {
        debug!("⏳ Starting TEE waiting for share reconstruction");

        let waiting_manager =
            TeeWaitingManager::new(self.config.waiting_config.clone(), self.storage.clone());

        // Start the waiting process in the background
        // The Genesis Boot should return immediately after setting up the waiting state
        // The actual waiting will happen when shares are injected via the InjectShares operation
        let initial_state = WaitingState::WaitingForShares;

        // Start the waiting manager in the background (this will be handled by the InjectShares operation)
        tokio::spawn(async move {
            let mut waiting_manager = waiting_manager;
            let _ = waiting_manager.start_waiting().await;
        });

        Ok(initial_state)
    }

    /// Get the current storage state
    pub fn get_storage_state(&self) -> crate::storage::StorageState {
        self.storage.get_storage_state()
    }

    /// Check if the Genesis Boot flow is complete
    pub fn is_complete(&self) -> bool {
        let state = self.get_storage_state();
        state.ephemeral_key_exists && state.manifest_exists
    }
}

/// Create a default Genesis Boot configuration for testing
pub fn create_default_genesis_boot_config() -> GenesisBootConfig {
    // Generate proper P256 key pairs for testing
    let pair1 = P256Pair::generate().unwrap();
    let pair2 = P256Pair::generate().unwrap();

    let test_member1 = QuorumMember {
        alias: "test_member_1".to_string(),
        pub_key: pair1.public_key().to_bytes(),
    };
    let test_member2 = QuorumMember {
        alias: "test_member_2".to_string(),
        pub_key: pair2.public_key().to_bytes(),
    };

    GenesisBootConfig {
        namespace_name: "renclave-test".to_string(),
        namespace_nonce: 1,
        manifest_members: vec![test_member1.clone(), test_member2.clone()],
        manifest_threshold: 2,
        share_members: vec![test_member1, test_member2],
        share_threshold: 2,
        pivot_hash: [0u8; 32],
        pivot_args: vec![],
        waiting_config: WaitingConfig {
            max_wait_time: 300,      // 5 minutes for production
            check_interval_ms: 1000, // 1 second for production
            min_shares: 2,
        },
        dr_key: None,
    }
}

/// Create a test Genesis Boot configuration with mock members
pub fn create_test_genesis_boot_config() -> GenesisBootConfig {
    // Generate proper P256 key pairs for testing
    let manifest_pair1 = P256Pair::generate().unwrap();
    let manifest_pair2 = P256Pair::generate().unwrap();
    let share_pair1 = P256Pair::generate().unwrap();
    let share_pair2 = P256Pair::generate().unwrap();
    let share_pair3 = P256Pair::generate().unwrap();

    let manifest_members = vec![
        QuorumMember {
            alias: "manifest_member_1".to_string(),
            pub_key: manifest_pair1.public_key().to_bytes(),
        },
        QuorumMember {
            alias: "manifest_member_2".to_string(),
            pub_key: manifest_pair2.public_key().to_bytes(),
        },
    ];

    let share_members = vec![
        QuorumMember {
            alias: "share_member_1".to_string(),
            pub_key: share_pair1.public_key().to_bytes(),
        },
        QuorumMember {
            alias: "share_member_2".to_string(),
            pub_key: share_pair2.public_key().to_bytes(),
        },
        QuorumMember {
            alias: "share_member_3".to_string(),
            pub_key: share_pair3.public_key().to_bytes(),
        },
    ];

    GenesisBootConfig {
        namespace_name: "renclave-test".to_string(),
        namespace_nonce: 1,
        manifest_members,
        manifest_threshold: 2,
        share_members,
        share_threshold: 2,
        pivot_hash: [0u8; 32],
        pivot_args: vec!["--test".to_string()],
        waiting_config: WaitingConfig {
            max_wait_time: 300,      // 5 minutes for production
            check_interval_ms: 1000, // 1 second for production
            min_shares: 2,
        },
        dr_key: None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_genesis_boot_flow_default() {
        // Initialize logging for the test
        let _ = env_logger::builder()
            .filter_level(log::LevelFilter::Info)
            .is_test(true)
            .try_init();

        println!("🧪 Starting Genesis Boot flow test");
        let config = create_default_genesis_boot_config();
        println!(
            "🧪 Config created with {} manifest members and {} share members",
            config.manifest_members.len(),
            config.share_members.len()
        );

        // Create a test flow with temporary storage
        let test_storage = TeeStorage::new();
        test_storage.clear_all().unwrap(); // Clear for testing
        let mut flow = GenesisBootFlow::new_with_storage(config, test_storage);

        // The flow should complete successfully
        println!("🧪 Executing Genesis Boot flow...");
        let result = flow.execute().await;
        if let Err(e) = &result {
            eprintln!("❌ Genesis Boot flow failed: {}", e);
        }
        assert!(result.is_ok());

        let result = result.unwrap();
        assert!(!result.quorum_public_key.is_empty());
        assert!(!result.ephemeral_key.is_empty());
        assert_eq!(
            result.manifest_envelope.manifest.namespace.name,
            "renclave-test"
        );
        assert_eq!(result.manifest_envelope.manifest.namespace.nonce, 1);
    }

    #[tokio::test]
    async fn test_genesis_boot_flow_with_members() {
        let config = create_test_genesis_boot_config();
        let mut flow = GenesisBootFlow::new(config);

        // The flow should complete successfully
        let result = flow.execute().await;
        assert!(result.is_ok());

        let result = result.unwrap();
        assert!(!result.quorum_public_key.is_empty());
        assert!(!result.ephemeral_key.is_empty());
        assert_eq!(
            result.manifest_envelope.manifest.namespace.name,
            "renclave-test"
        );
        assert_eq!(
            result.manifest_envelope.manifest.manifest_set.members.len(),
            2
        );
        assert_eq!(result.manifest_envelope.manifest.share_set.members.len(), 3);
        assert_eq!(result.manifest_envelope.manifest.manifest_set.threshold, 2);
        assert_eq!(result.manifest_envelope.manifest.share_set.threshold, 2);
    }

    #[tokio::test]
    async fn test_genesis_boot_storage_state() {
        let config = create_default_genesis_boot_config();
        let mut flow = GenesisBootFlow::new(config);

        // Initially should not be complete
        assert!(!flow.is_complete());

        // Execute the flow
        let result = flow.execute().await;
        assert!(result.is_ok());

        // After execution, should be complete
        assert!(flow.is_complete());

        let storage_state = flow.get_storage_state();
        assert!(storage_state.ephemeral_key_exists);
        assert!(storage_state.manifest_exists);
    }

    #[test]
    fn test_default_config() {
        let config = create_default_genesis_boot_config();
        assert_eq!(config.namespace_name, "renclave-test");
        assert_eq!(config.namespace_nonce, 1);
        assert_eq!(config.manifest_threshold, 1);
        assert_eq!(config.share_threshold, 1);
        assert!(config.manifest_members.is_empty());
        assert!(config.share_members.is_empty());
    }

    #[test]
    fn test_test_config() {
        let config = create_test_genesis_boot_config();
        assert_eq!(config.namespace_name, "renclave-test");
        assert_eq!(config.namespace_nonce, 1);
        assert_eq!(config.manifest_threshold, 2);
        assert_eq!(config.share_threshold, 2);
        assert_eq!(config.manifest_members.len(), 2);
        assert_eq!(config.share_members.len(), 3);
        assert_eq!(config.pivot_args, vec!["--test"]);
    }
}
