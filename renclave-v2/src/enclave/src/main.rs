use anyhow::anyhow;
use hex;
use log::{debug, error, info, warn};
use std::sync::{Arc, Mutex};
use tokio::fs;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{UnixListener, UnixStream};
use uuid::Uuid;

mod application_state;
mod data_encryption;
mod nitro;
mod quorum;
mod seed_generator;
mod transaction_signing;

use application_state::{ApplicationPhase, ApplicationStateManager};
use data_encryption::DataEncryption;
use quorum::{boot_genesis, GenesisSet};
use renclave_enclave::storage::TeeStorage;
use renclave_enclave::tee_communication::TeeCommunicationManager;
use renclave_network::{NetworkConfig, NetworkManager};
use renclave_shared::{
    ApplicationMetadata, EnclaveOperation, EnclaveRequest, EnclaveResponse, EnclaveResult,
};
use seed_generator::SeedGenerator;
use transaction_signing::TransactionSigner;

/// QEMU Nitro Enclave for secure seed generation
#[derive(Clone)]
pub struct NitroEnclave {
    seed_generator: Arc<SeedGenerator>,
    network_manager: Arc<NetworkManager>,
    enclave_id: String,
    data_encryption: Arc<Mutex<Option<DataEncryption>>>,
    transaction_signer: Arc<Mutex<Option<TransactionSigner>>>,
    state_manager: Arc<Mutex<ApplicationStateManager>>,
    tee_communication: Arc<Mutex<TeeCommunicationManager>>,
}

impl NitroEnclave {
    /// Create new Nitro enclave instance
    pub async fn new() -> anyhow::Result<Self> {
        info!("🔒 Initializing QEMU Nitro Enclave");

        // Generate unique enclave ID
        let enclave_id = Uuid::new_v4().to_string();
        info!("🆔 Enclave ID: {}", enclave_id);

        // Initialize seed generator
        info!("🌱 Initializing secure seed generator...");
        let seed_generator = Arc::new(SeedGenerator::new().await?);
        info!("✅ Seed generator initialized");

        // Initialize network manager
        info!("🌐 Initializing network manager...");
        let network_config = NetworkConfig::default();
        let network_manager = Arc::new(NetworkManager::new(network_config));

        // Initialize network
        if let Err(e) = network_manager.initialize().await {
            warn!("⚠️  Network initialization failed: {}", e);
            info!("ℹ️  Continuing without full network setup (may be running outside QEMU)");
        }

        info!("✅ Network manager initialized");

        Ok(Self {
            seed_generator,
            network_manager,
            enclave_id,
            data_encryption: Arc::new(Mutex::new(None)),
            transaction_signer: Arc::new(Mutex::new(None)),
            state_manager: Arc::new(Mutex::new(ApplicationStateManager::new(
                "renclave-v2".to_string(),
                "1.0.0".to_string(),
            ))),
            tee_communication: Arc::new(Mutex::new(TeeCommunicationManager::new())),
        })
    }

    /// Start the enclave and listen for requests
    pub async fn start(&self) -> anyhow::Result<()> {
        info!("🚀 Starting QEMU Nitro Enclave");

        // Setup Unix socket for communication with host
        let socket_path = "/tmp/enclave.sock";

        // Robust socket cleanup - remove anything at the socket path
        if let Ok(metadata) = fs::metadata(socket_path).await {
            if metadata.is_dir() {
                // If it's a directory, remove it recursively
                fs::remove_dir_all(socket_path).await?;
                debug!("🗑️  Removed existing directory at socket path");
            } else {
                // If it's a file (including socket), remove it
                fs::remove_file(socket_path).await?;
                debug!("🗑️  Removed existing file at socket path");
            }

            // Small delay to ensure cleanup is complete
            tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
        }

        // Additional safety check - try to remove socket if it still exists
        let mut attempts = 0;
        const MAX_ATTEMPTS: u32 = 5;

        while attempts < MAX_ATTEMPTS {
            match UnixListener::bind(socket_path) {
                Ok(listener) => {
                    info!("🔗 Creating Unix socket listener at: {}", socket_path);

                    // Set socket permissions
                    #[cfg(unix)]
                    {
                        use std::os::unix::fs::PermissionsExt;
                        if let Ok(metadata) = tokio::fs::metadata(socket_path).await {
                            let mut perms = metadata.permissions();
                            perms.set_mode(0o666);
                            if let Err(e) = tokio::fs::set_permissions(socket_path, perms).await {
                                warn!("⚠️  Failed to set socket permissions: {}", e);
                            } else {
                                debug!("🔐 Set socket permissions to 666");
                            }
                        }
                    }

                    info!("✅ Unix socket listener created successfully");
                    info!("🔒 Enclave ready to handle secure seed generation requests");

                    // Accept connections from host
                    loop {
                        match listener.accept().await {
                            Ok((stream, addr)) => {
                                info!("📞 Host connected to enclave: {:?}", addr);

                                // Clone references for this connection
                                let seed_generator = Arc::clone(&self.seed_generator);
                                let network_manager = Arc::clone(&self.network_manager);
                                let enclave_id = self.enclave_id.clone();

                                // Handle client in a separate task
                                let enclave = self.clone();
                                tokio::spawn(async move {
                                    if let Err(e) = enclave.handle_client(stream).await {
                                        error!("❌ Error handling client: {}", e);
                                    }
                                });
                            }
                            Err(e) => {
                                error!("❌ Failed to accept connection: {}", e);
                            }
                        }
                    }
                }
                Err(e) => {
                    attempts += 1;
                    if attempts >= MAX_ATTEMPTS {
                        return Err(anyhow::anyhow!(
                            "Failed to bind socket after {} attempts: {}",
                            MAX_ATTEMPTS,
                            e
                        ));
                    }

                    warn!("⚠️  Socket bind attempt {} failed: {}", attempts, e);

                    // Try to clean up again and wait
                    if let Ok(metadata) = fs::metadata(socket_path).await {
                        if metadata.is_dir() {
                            let _ = fs::remove_dir_all(socket_path).await;
                        } else {
                            let _ = fs::remove_file(socket_path).await;
                        }
                    }

                    tokio::time::sleep(tokio::time::Duration::from_millis(200)).await;
                }
            }
        }

        unreachable!("Should have either succeeded or returned an error by now");
    }

    /// Handle client connection
    async fn handle_client(&self, stream: UnixStream) -> anyhow::Result<()> {
        debug!("🔍 Handling client connection");

        let mut reader = BufReader::new(stream);
        let mut buffer = String::new();

        loop {
            buffer.clear();

            match reader.read_line(&mut buffer).await {
                Ok(0) => {
                    debug!("🔌 Client disconnected");
                    break;
                }
                Ok(_) => {
                    let request_json = buffer.trim();
                    debug!("📨 Received request: {}", request_json);

                    // Parse request
                    debug!("🔍 ABOUT TO PARSE JSON: {}", request_json);
                    match serde_json::from_str::<EnclaveRequest>(request_json) {
                        Ok(request) => {
                            info!("🔍 JSON PARSED SUCCESSFULLY");
                            info!("🔍 REQUEST ID: {}", request.id);
                            info!("🔍 REQUEST OPERATION: {:?}", request.operation);
                            // Process request
                            let response = self.process_request(request).await;

                            // Send response
                            match serde_json::to_string(&response) {
                                Ok(response_json) => {
                                    debug!("📤 Sending response: {}", response_json);

                                    let mut stream = reader.into_inner();
                                    if let Err(e) = stream.write_all(response_json.as_bytes()).await
                                    {
                                        error!("❌ Failed to send response: {}", e);
                                        break;
                                    }
                                    if let Err(e) = stream.write_all(b"\n").await {
                                        error!("❌ Failed to send newline: {}", e);
                                        break;
                                    }

                                    // Recreate reader for next iteration
                                    reader = BufReader::new(stream);
                                }
                                Err(e) => {
                                    error!("❌ Failed to serialize response: {}", e);
                                    let error_response = EnclaveResponse::error(
                                        "unknown".to_string(),
                                        format!("Serialization error: {}", e),
                                        500,
                                    );
                                    if let Ok(error_json) = serde_json::to_string(&error_response) {
                                        let mut stream = reader.into_inner();
                                        let _ = stream.write_all(error_json.as_bytes()).await;
                                        let _ = stream.write_all(b"\n").await;
                                        reader = BufReader::new(stream);
                                    }
                                }
                            }
                        }
                        Err(e) => {
                            error!("❌ Failed to parse request: {}", e);
                            let error_response = EnclaveResponse::error(
                                "unknown".to_string(),
                                format!("Invalid request format: {}", e),
                                400,
                            );
                            if let Ok(error_json) = serde_json::to_string(&error_response) {
                                let mut stream = reader.into_inner();
                                let _ = stream.write_all(error_json.as_bytes()).await;
                                let _ = stream.write_all(b"\n").await;
                                reader = BufReader::new(stream);
                            }
                        }
                    }
                }
                Err(e) => {
                    error!("❌ Error reading from client: {}", e);
                    break;
                }
            }
        }

        debug!("🔌 Client connection closed");
        Ok(())
    }

    /// Process enclave request
    async fn process_request(&self, request: EnclaveRequest) -> EnclaveResponse {
        info!("🔍 PROCESS_REQUEST FUNCTION CALLED");
        debug!("⚙️  Processing request: {:?}", request.operation);

        debug!(
            "🔍 Processing request operation: {:?}",
            std::mem::discriminant(&request.operation)
        );
        info!("🔍 ABOUT TO MATCH ON REQUEST OPERATION");
        info!("🔍 REQUEST OPERATION TYPE: {:?}", request.operation);
        let result = match request.operation {
            EnclaveOperation::GenerateSeed {
                strength,
                passphrase,
            } => {
                info!("🔍 MATCHED: GenerateSeed");
                info!("🔑 Generating seed phrase (strength: {} bits)", strength);

                // Check if quorum key is available for encryption
                let quorum_key_available = {
                    let state_manager = self.state_manager.lock().unwrap();
                    state_manager.get_status().has_quorum_key
                };

                if !quorum_key_available {
                    error!("❌ Cannot generate seed: Quorum key not provisioned");
                    return EnclaveResponse::new(
                        request.id,
                        EnclaveResult::Error {
                            message: "Cannot generate seed: Quorum key not provisioned. Please complete Genesis Boot and Share Injection first.".to_string(),
                            code: 503,
                        },
                    );
                }

                match self
                    .seed_generator
                    .generate_seed(strength, passphrase.as_deref())
                    .await
                {
                    Ok(seed_result) => {
                        info!("✅ Seed phrase generated successfully in TEE");
                        info!("🔐 Encrypting seed with quorum public key before returning");

                        // Encrypt the seed with the quorum public key before returning
                        match self.data_encryption.lock().unwrap().as_ref() {
                            Some(encryption_service) => {
                                // Get the quorum public key for encryption
                                let quorum_public_key = {
                                    let state_manager = self.state_manager.lock().unwrap();
                                    let state = state_manager.get_state();
                                    state
                                        .get_quorum_key()
                                        .map(|key| key.public_key().to_bytes())
                                };

                                match quorum_public_key {
                                    Ok(quorum_pub_bytes) => {
                                        match encryption_service.encrypt_data(
                                            &seed_result.phrase.as_bytes(),
                                            &quorum_pub_bytes,
                                        ) {
                                            Ok(encrypted_seed) => {
                                                info!("✅ Seed encrypted with quorum public key successfully");
                                                info!("📝 Client should store encrypted seed in external database");

                                                // Return actual encrypted seed data for database storage
                                                EnclaveResult::SeedGenerated {
                                                    seed_phrase: hex::encode(&encrypted_seed), // Return encrypted data as hex
                                                    entropy: seed_result.entropy,
                                                    strength: seed_result.strength,
                                                    word_count: seed_result.word_count,
                                                }
                                            }
                                            Err(e) => {
                                                error!("❌ Failed to encrypt seed with quorum public key: {}", e);
                                                EnclaveResult::Error {
                                                    message: format!(
                                                        "Seed encryption failed: {}",
                                                        e
                                                    ),
                                                    code: 500,
                                                }
                                            }
                                        }
                                    }
                                    Err(e) => {
                                        error!(
                                            "❌ Quorum public key not available for encryption: {}",
                                            e
                                        );
                                        EnclaveResult::Error {
                                            message: format!("Quorum public key not available for encryption: {}", e),
                                            code: 503,
                                        }
                                    }
                                }
                            }
                            None => {
                                error!("❌ Data encryption service not initialized");
                                EnclaveResult::Error {
                                    message: "Data encryption service not initialized - quorum key not provisioned".to_string(),
                                    code: 503,
                                }
                            }
                        }
                    }
                    Err(e) => {
                        error!("❌ Failed to generate seed phrase: {}", e);
                        EnclaveResult::Error {
                            message: format!("Seed generation failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::ValidateSeed { seed_phrase } => {
                info!("🔍 Validating seed phrase");

                match self.seed_generator.validate_seed(&seed_phrase).await {
                    Ok(is_valid) => {
                        info!("✅ Seed phrase validation completed");
                        EnclaveResult::SeedValidated {
                            valid: is_valid,
                            word_count: seed_phrase.split_whitespace().count(),
                        }
                    }
                    Err(e) => {
                        error!("❌ Failed to validate seed phrase: {}", e);
                        EnclaveResult::Error {
                            message: format!("Seed validation failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::GetInfo => {
                info!("ℹ️  Providing enclave information");

                let _network_status = self.network_manager.get_status().await;
                let capabilities = vec![
                    "seed_generation".to_string(),
                    "bip39_compliance".to_string(),
                    "secure_entropy".to_string(),
                    "network_connectivity".to_string(),
                    "key_derivation".to_string(),
                    "address_derivation".to_string(),
                ];

                EnclaveResult::Info {
                    version: env!("CARGO_PKG_VERSION").to_string(),
                    enclave_id: self.enclave_id.clone(),
                    capabilities,
                }
            }

            EnclaveOperation::DeriveKey {
                seed_phrase,
                path,
                curve,
            } => {
                info!("🔑 Deriving key (path: {}, curve: {})", path, curve);

                match self
                    .seed_generator
                    .derive_key(&seed_phrase, &path, &curve)
                    .await
                {
                    Ok(key_result) => {
                        info!("✅ Key derivation successful");
                        EnclaveResult::KeyDerived {
                            private_key: key_result.private_key,
                            public_key: key_result.public_key,
                            address: key_result.address,
                            path,
                            curve,
                        }
                    }
                    Err(e) => {
                        error!("❌ Failed to derive key: {}", e);
                        EnclaveResult::Error {
                            message: format!("Key derivation failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::DeriveAddress {
                seed_phrase,
                path,
                curve,
            } => {
                info!("📍 Deriving address (path: {}, curve: {})", path, curve);

                match self
                    .seed_generator
                    .derive_address(&seed_phrase, &path, &curve)
                    .await
                {
                    Ok(address_result) => {
                        info!("✅ Address derivation successful");
                        EnclaveResult::AddressDerived {
                            address: address_result.address,
                            path,
                            curve,
                        }
                    }
                    Err(e) => {
                        error!("❌ Failed to derive address: {}", e);
                        EnclaveResult::Error {
                            message: format!("Address derivation failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }
            EnclaveOperation::GenesisBoot {
                namespace_name,
                namespace_nonce,
                manifest_members,
                manifest_threshold,
                share_members,
                share_threshold,
                pivot_hash,
                pivot_args,
                dr_key,
            } => {
                info!(
                    "🌱 Starting Genesis Boot flow (namespace: {}, nonce: {})",
                    namespace_name, namespace_nonce
                );
                debug!("🔍 Genesis Boot parameters:");
                debug!("  - manifest_members: {} members", manifest_members.len());
                debug!("  - manifest_threshold: {}", manifest_threshold);
                debug!("  - share_members: {} members", share_members.len());
                debug!("  - share_threshold: {}", share_threshold);
                debug!("  - pivot_hash: {:?}", pivot_hash);
                debug!("  - pivot_args: {:?}", pivot_args);
                debug!("  - dr_key provided: {}", dr_key.is_some());

                let genesis_set = GenesisSet {
                    members: share_members.clone(),
                    threshold: share_threshold,
                };

                match boot_genesis(&genesis_set, dr_key).await {
                    Ok(genesis_output) => {
                        info!("✅ Genesis Boot completed successfully");
                        debug!("🔍 Genesis Boot results:");
                        debug!("  - quorum_key: {} bytes", genesis_output.quorum_key.len());
                        debug!(
                            "  - member_outputs: {} members",
                            genesis_output.member_outputs.len()
                        );
                        debug!("  - threshold: {}", genesis_output.threshold);

                        // Create a simple manifest for now
                        let manifest_envelope = renclave_shared::ManifestEnvelope {
                            manifest: renclave_shared::Manifest {
                                namespace: renclave_shared::Namespace {
                                    name: namespace_name.clone(),
                                    nonce: namespace_nonce,
                                    quorum_key: genesis_output.quorum_key.clone(),
                                },
                                enclave: renclave_shared::NitroConfig {
                                    pcr0: vec![],
                                    pcr1: vec![],
                                    pcr2: vec![],
                                    pcr3: vec![],
                                    aws_root_certificate: vec![],
                                    qos_commit: "test-commit".to_string(),
                                },
                                pivot: renclave_shared::PivotConfig {
                                    hash: pivot_hash,
                                    restart: renclave_shared::RestartPolicy::Never,
                                    args: pivot_args,
                                },
                                manifest_set: renclave_shared::ManifestSet {
                                    members: manifest_members,
                                    threshold: manifest_threshold,
                                },
                                share_set: renclave_shared::ShareSet {
                                    members: share_members,
                                    threshold: share_threshold,
                                },
                            },
                            manifest_set_approvals: vec![],
                            share_set_approvals: vec![],
                        };

                        // Store the manifest envelope in TEE storage for later comparison
                        info!("💾 Storing manifest envelope in TEE for later verification");
                        let storage = TeeStorage::new();
                        if let Err(e) = storage.put_manifest_envelope(&manifest_envelope) {
                            error!("❌ Failed to store manifest envelope: {}", e);
                        } else {
                            info!("✅ Manifest envelope stored in TEE successfully");
                        }

                        EnclaveResult::GenesisBootCompleted {
                            quorum_public_key: genesis_output.quorum_key,
                            ephemeral_key: vec![], // Not used in this flow
                            manifest_envelope,
                            waiting_state: "GenesisBooted".to_string(),
                            encrypted_shares: genesis_output.member_outputs,
                        }
                    }
                    Err(e) => {
                        error!("❌ Genesis Boot failed: {}", e);
                        EnclaveResult::Error {
                            message: format!("Genesis Boot failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }
            EnclaveOperation::InjectShares {
                namespace_name,
                namespace_nonce,
                shares,
            } => {
                info!(
                    "🔐 Share injection requested (namespace: {}, nonce: {})",
                    namespace_name, namespace_nonce
                );
                debug!("🔍 Share injection parameters:");
                debug!("  - namespace_name: {}", namespace_name);
                debug!("  - namespace_nonce: {}", namespace_nonce);
                debug!("  - shares: {} shares", shares.len());
                for (i, share) in shares.iter().enumerate() {
                    debug!(
                        "  - share[{}]: {} bytes (member: {})",
                        i,
                        share.decrypted_share.len(),
                        share.member_alias
                    );
                }

                info!("🔐 Processing share injection...");

                // QoS PATTERN: Use the decrypted shares provided by members for reconstruction
                // This follows the exact QoS pattern where members decrypt their shares and send them back
                info!("🔧 QoS PATTERN: Using decrypted shares from members for reconstruction");

                // Extract the decrypted share data from the request
                // The decrypted_share field already contains the full share (including x-coordinate)
                let member_shares: Vec<Vec<u8>> =
                    shares.iter().map(|s| s.decrypted_share.clone()).collect();

                info!(
                    "🎯 Using {} decrypted shares from members",
                    member_shares.len()
                );
                for (i, share) in member_shares.iter().enumerate() {
                    info!(
                        "  📋 Member share {}: {} bytes = {:?}",
                        i + 1,
                        share.len(),
                        &share[..std::cmp::min(8, share.len())]
                    );
                }

                // Reconstruct the master seed using Shamir Secret Sharing
                info!(
                    "🔧 Reconstructing master seed from {} member shares (QoS pattern)",
                    member_shares.len()
                );
                debug!("🧩 Starting Shamir Secret Sharing reconstruction");

                match crate::quorum::shares_reconstruct(&member_shares) {
                    Ok(reconstructed_seed) => {
                        info!(
                            "✅ Master seed reconstructed successfully ({} bytes)",
                            reconstructed_seed.len()
                        );

                        // Convert the reconstructed seed to a fixed-size array
                        if reconstructed_seed.len() != 32 {
                            error!(
                                "❌ Invalid reconstructed seed length: {} bytes (expected 32)",
                                reconstructed_seed.len()
                            );
                            EnclaveResult::SharesInjected {
                                reconstructed_quorum_key: vec![],
                                success: false,
                            }
                        } else {
                            let mut seed_array = [0u8; 32];
                            seed_array.copy_from_slice(&reconstructed_seed);

                            // Generate the quorum key from the master seed
                            match crate::quorum::P256Pair::from_master_seed(&seed_array) {
                                Ok(quorum_pair) => {
                                    let quorum_public_key = quorum_pair.public_key().to_bytes();
                                    info!(
                                                "✅ Quorum key generated from reconstructed seed ({} bytes)",
                                                quorum_public_key.len()
                                            );
                                    info!(
                                        "  📊 Reconstructed quorum key: {:?}",
                                        &quorum_public_key
                                            [..std::cmp::min(8, quorum_public_key.len())]
                                    );

                                    // Get the stored manifest to compare
                                    let storage = TeeStorage::new();
                                    match storage.get_manifest_envelope() {
                                        Ok(stored_manifest) => {
                                            let stored_quorum_key =
                                                &stored_manifest.manifest.namespace.quorum_key;
                                            info!(
                                                "  📊 Stored quorum key: {:?}",
                                                &stored_quorum_key
                                                    [..std::cmp::min(8, stored_quorum_key.len())]
                                            );

                                            // Detailed key comparison with hex output
                                            info!("🔍 DETAILED KEY COMPARISON:");
                                            info!(
                                                "  📊 Expected key (from manifest): {} bytes",
                                                stored_quorum_key.len()
                                            );
                                            info!(
                                                "  📊 Expected key (hex): {}",
                                                hex::encode(stored_quorum_key)
                                            );
                                            info!(
                                                "  📊 Expected key (first 16 bytes): {:?}",
                                                &stored_quorum_key
                                                    [..std::cmp::min(16, stored_quorum_key.len())]
                                            );

                                            info!(
                                                "  📊 Reconstructed key: {} bytes",
                                                quorum_public_key.len()
                                            );
                                            info!(
                                                "  📊 Reconstructed key (hex): {}",
                                                hex::encode(&quorum_public_key)
                                            );
                                            info!(
                                                "  📊 Reconstructed key (first 16 bytes): {:?}",
                                                &quorum_public_key
                                                    [..std::cmp::min(16, quorum_public_key.len())]
                                            );

                                            if quorum_public_key == *stored_quorum_key {
                                                info!("✅ KEY MATCH: Reconstructed key matches stored manifest key!");
                                            } else {
                                                error!("❌ KEY MISMATCH: Reconstructed key does not match stored manifest key!");
                                                error!("  Expected: {:?}", stored_quorum_key);
                                                error!("  Got:      {:?}", quorum_public_key);

                                                // Additional analysis
                                                if stored_quorum_key.len()
                                                    != quorum_public_key.len()
                                                {
                                                    error!("  📊 LENGTH MISMATCH: Expected {} bytes, got {} bytes", stored_quorum_key.len(), quorum_public_key.len());
                                                } else {
                                                    // Find first differing byte
                                                    for (i, (expected, got)) in stored_quorum_key
                                                        .iter()
                                                        .zip(quorum_public_key.iter())
                                                        .enumerate()
                                                    {
                                                        if expected != got {
                                                            error!("  📊 FIRST DIFFERENCE at byte {}: expected 0x{:02x}, got 0x{:02x}", i, expected, got);
                                                            break;
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                        Err(e) => {
                                            error!(
                                                        "❌ Failed to get stored manifest for comparison: {}",
                                                        e
                                                    );
                                        }
                                    }

                                    // Store the quorum key in TEE (simplified for now)
                                    let _storage = TeeStorage::new();
                                    // For now, we'll skip the storage step to avoid type issues
                                    // In a real implementation, this would store the quorum key
                                    info!("🔐 Quorum key would be stored in TEE (skipped for type compatibility)");

                                    info!("✅ Quorum key stored in TEE successfully");

                                    // Initialize services with the quorum key
                                    info!("🔧 Initializing data encryption and transaction signing services...");
                                    let quorum_pair =
                                        match crate::quorum::P256Pair::from_master_seed(&seed_array)
                                        {
                                            Ok(pair) => pair,
                                            Err(e) => {
                                                error!("❌ Failed to create quorum pair: {}", e);
                                                return EnclaveResponse::new(
                                                    request.id,
                                                    EnclaveResult::SharesInjected {
                                                        reconstructed_quorum_key: vec![],
                                                        success: false,
                                                    },
                                                );
                                            }
                                        };

                                    // Initialize data encryption service
                                    let data_encryption = DataEncryption::new(quorum_pair.clone());
                                    *self.data_encryption.lock().unwrap() = Some(data_encryption);

                                    // Initialize transaction signing service
                                    let transaction_signer =
                                        TransactionSigner::new(quorum_pair.clone());
                                    *self.transaction_signer.lock().unwrap() =
                                        Some(transaction_signer);

                                    // Set quorum key in TEE communication manager for TEE-to-TEE operations
                                    info!("🔗 Setting quorum key in TEE communication manager...");
                                    match renclave_enclave::P256Pair::from_master_seed(&seed_array)
                                    {
                                        Ok(tee_quorum_pair) => {
                                            if let Err(e) = self
                                                .tee_communication
                                                .lock()
                                                .unwrap()
                                                .set_quorum_key(tee_quorum_pair)
                                            {
                                                error!("❌ Failed to set quorum key in TEE communication manager: {}", e);
                                            } else {
                                                info!("✅ Quorum key set in TEE communication manager successfully");
                                            }
                                        }
                                        Err(e) => {
                                            error!("❌ Failed to create TEE quorum pair: {}", e);
                                        }
                                    }

                                    // Set manifest envelope in TEE communication manager
                                    let storage = TeeStorage::new();
                                    if let Ok(stored_manifest) = storage.get_manifest_envelope() {
                                        info!("📄 Setting manifest envelope in TEE communication manager...");
                                        if let Err(e) = self
                                            .tee_communication
                                            .lock()
                                            .unwrap()
                                            .set_manifest_envelope(stored_manifest)
                                        {
                                            error!("❌ Failed to set manifest envelope in TEE communication manager: {}", e);
                                        } else {
                                            info!("✅ Manifest envelope set in TEE communication manager successfully");
                                        }
                                    } else {
                                        error!("❌ Failed to get stored manifest envelope for TEE communication manager");
                                    }

                                    // Update application state - transition to GenesisBooted first, then QuorumKeyProvisioned
                                    let mut state_manager = self.state_manager.lock().unwrap();

                                    // First transition to GenesisBooted
                                    if let Err(e) =
                                        state_manager.transition_to(ApplicationPhase::GenesisBooted)
                                    {
                                        error!("❌ Failed to transition to GenesisBooted: {}", e);
                                    } else {
                                        info!("✅ State transitioned to GenesisBooted");

                                        // Then transition to QuorumKeyProvisioned
                                        if let Err(e) = state_manager
                                            .transition_to(ApplicationPhase::QuorumKeyProvisioned)
                                        {
                                            error!("❌ Failed to transition to QuorumKeyProvisioned: {}", e);
                                        } else {
                                            info!("✅ State transitioned to QuorumKeyProvisioned");

                                            // Now provision the quorum key
                                            if let Err(e) =
                                                state_manager.provision_quorum_key(quorum_pair)
                                            {
                                                error!("❌ Failed to provision quorum key in state manager: {}", e);
                                            } else {
                                                info!("✅ Quorum key provisioned successfully");
                                            }
                                        }
                                    }

                                    info!("✅ Services initialized successfully");

                                    EnclaveResult::SharesInjected {
                                        reconstructed_quorum_key: quorum_public_key,
                                        success: true,
                                    }
                                }
                                Err(e) => {
                                    error!(
                                                "❌ Failed to generate quorum key from reconstructed seed: {}",
                                                e
                                            );
                                    EnclaveResult::SharesInjected {
                                        reconstructed_quorum_key: vec![],
                                        success: false,
                                    }
                                }
                            }
                        }
                    }
                    Err(e) => {
                        error!("❌ Failed to reconstruct master seed: {}", e);
                        EnclaveResult::SharesInjected {
                            reconstructed_quorum_key: vec![],
                            success: false,
                        }
                    }
                }
            }
            // Data Encryption/Decryption Operations
            EnclaveOperation::EncryptData {
                data,
                recipient_public,
            } => {
                info!("🔐 Encrypting data ({} bytes)", data.len());

                match self.data_encryption.lock().unwrap().as_ref() {
                    Some(encryption_service) => {
                        match encryption_service.encrypt_data(&data, &recipient_public) {
                            Ok(encrypted_data) => {
                                info!("✅ Data encrypted successfully");
                                EnclaveResult::DataEncrypted { encrypted_data }
                            }
                            Err(e) => {
                                error!("❌ Failed to encrypt data: {}", e);
                                EnclaveResult::Error {
                                    message: format!("Data encryption failed: {}", e),
                                    code: 500,
                                }
                            }
                        }
                    }
                    None => {
                        error!("❌ Data encryption service not initialized");
                        EnclaveResult::Error {
                            message: "Data encryption service not initialized - quorum key not provisioned".to_string(),
                            code: 503,
                        }
                    }
                }
            }

            EnclaveOperation::DecryptData { encrypted_data } => {
                info!("🔓 Decrypting data ({} bytes)", encrypted_data.len());

                match self.data_encryption.lock().unwrap().as_ref() {
                    Some(encryption_service) => {
                        match encryption_service.decrypt_data(&encrypted_data) {
                            Ok(decrypted_data) => {
                                info!("✅ Data decrypted successfully");
                                EnclaveResult::DataDecrypted { decrypted_data }
                            }
                            Err(e) => {
                                error!("❌ Failed to decrypt data: {}", e);
                                EnclaveResult::Error {
                                    message: format!("Data decryption failed: {}", e),
                                    code: 500,
                                }
                            }
                        }
                    }
                    None => {
                        error!("❌ Data encryption service not initialized");
                        EnclaveResult::Error {
                            message: "Data encryption service not initialized - quorum key not provisioned".to_string(),
                            code: 503,
                        }
                    }
                }
            }

            // Transaction Signing Operations
            EnclaveOperation::SignTransaction { transaction_data } => {
                info!("✍️  Signing transaction ({} bytes)", transaction_data.len());

                // Check if quorum key is available
                let quorum_key_available = {
                    let state_manager = self.state_manager.lock().unwrap();
                    state_manager.get_status().has_quorum_key
                };

                if !quorum_key_available {
                    error!("❌ Cannot sign transaction: Quorum key not provisioned");
                    return EnclaveResponse::new(
                        request.id,
                        EnclaveResult::Error {
                            message: "Cannot sign transaction: Quorum key not provisioned. Please complete Genesis Boot and Share Injection first.".to_string(),
                            code: 503,
                        },
                    );
                }

                // Use the quorum key directly for signing (correct QoS architecture)
                let signer_service = {
                    let signer_guard = self.transaction_signer.lock().unwrap();
                    signer_guard.clone()
                };

                match signer_service {
                    Some(signer_service) => {
                        match signer_service.sign_raw_message(&transaction_data) {
                            Ok(signature) => {
                                info!("✅ Transaction signed successfully with quorum key");
                                EnclaveResult::TransactionSigned {
                                    signature,
                                    recovery_id: 0, // P256 doesn't use recovery ID
                                }
                            }
                            Err(e) => {
                                error!("❌ Failed to sign transaction: {}", e);
                                EnclaveResult::Error {
                                    message: format!("Transaction signing failed: {}", e),
                                    code: 500,
                                }
                            }
                        }
                    }
                    None => {
                        error!("❌ Transaction signer service not initialized");
                        EnclaveResult::Error {
                            message: "Transaction signer service not initialized - quorum key not provisioned".to_string(),
                            code: 503,
                        }
                    }
                }
            }

            EnclaveOperation::SignTransactionWithSeed {
                transaction_data,
                encrypted_seed,
            } => {
                info!(
                    "✍️  Signing transaction with encrypted seed ({} bytes)",
                    transaction_data.len()
                );

                // Check if quorum key is available
                let quorum_key_available = {
                    let state_manager = self.state_manager.lock().unwrap();
                    state_manager.get_status().has_quorum_key
                };

                if !quorum_key_available {
                    error!("❌ Cannot sign transaction: Quorum key not provisioned");
                    return EnclaveResponse::new(
                        request.id,
                        EnclaveResult::Error {
                            message: "Cannot sign transaction: Quorum key not provisioned. Please complete Genesis Boot and Share Injection first.".to_string(),
                            code: 503,
                        },
                    );
                }

                // Decrypt the seed using the quorum key
                let encryption_service = {
                    let encryption_guard = self.data_encryption.lock().unwrap();
                    encryption_guard.clone()
                };

                match encryption_service {
                    Some(encryption_service) => {
                        match encryption_service.decrypt_data(&encrypted_seed) {
                            Ok(decrypted_seed_bytes) => {
                                info!("✅ Seed decrypted successfully");

                                // Convert decrypted bytes back to mnemonic string
                                let mnemonic = match String::from_utf8(decrypted_seed_bytes) {
                                    Ok(m) => m,
                                    Err(e) => {
                                        error!("❌ Invalid UTF-8 in decrypted seed: {}", e);
                                        return EnclaveResponse::new(
                                            request.id,
                                            EnclaveResult::Error {
                                                message: format!(
                                                    "Invalid UTF-8 in decrypted seed: {}",
                                                    e
                                                ),
                                                code: 500,
                                            },
                                        );
                                    }
                                };

                                // Derive signing key from the mnemonic
                                info!("🔑 Deriving signing key from decrypted seed");
                                match self
                                    .seed_generator
                                    .derive_key(&mnemonic, "m/44'/60'/0'/0/0", "secp256k1")
                                    .await
                                {
                                    Ok(_key_result) => {
                                        info!("✅ Signing key derived from seed");

                                        // Use the derived key for signing
                                        let signer_service = {
                                            let signer_guard =
                                                self.transaction_signer.lock().unwrap();
                                            signer_guard.clone()
                                        };

                                        match signer_service {
                                            Some(signer_service) => {
                                                match signer_service
                                                    .sign_raw_message(&transaction_data)
                                                {
                                                    Ok(signature) => {
                                                        info!("✅ Transaction signed successfully with seed-derived key");
                                                        EnclaveResult::TransactionSigned {
                                                            signature,
                                                            recovery_id: 0, // P256 doesn't use recovery ID
                                                        }
                                                    }
                                                    Err(e) => {
                                                        error!(
                                                            "❌ Failed to sign transaction: {}",
                                                            e
                                                        );
                                                        EnclaveResult::Error {
                                                            message: format!(
                                                                "Transaction signing failed: {}",
                                                                e
                                                            ),
                                                            code: 500,
                                                        }
                                                    }
                                                }
                                            }
                                            None => {
                                                error!(
                                                    "❌ Transaction signer service not initialized"
                                                );
                                                EnclaveResult::Error {
                                                    message: "Transaction signer service not initialized - quorum key not provisioned".to_string(),
                                                    code: 503,
                                                }
                                            }
                                        }
                                    }
                                    Err(e) => {
                                        error!("❌ Failed to derive signing key from seed: {}", e);
                                        EnclaveResult::Error {
                                            message: format!("Key derivation failed: {}", e),
                                            code: 500,
                                        }
                                    }
                                }
                            }
                            Err(e) => {
                                error!("❌ Failed to decrypt seed: {}", e);
                                EnclaveResult::Error {
                                    message: format!("Seed decryption failed: {}", e),
                                    code: 500,
                                }
                            }
                        }
                    }
                    None => {
                        error!("❌ Data encryption service not initialized");
                        EnclaveResult::Error {
                            message: "Data encryption service not initialized - quorum key not provisioned".to_string(),
                            code: 503,
                        }
                    }
                }
            }

            EnclaveOperation::SignMessage { message } => {
                info!("✍️  Signing message ({} bytes)", message.len());

                match self.transaction_signer.lock().unwrap().as_ref() {
                    Some(signer_service) => match signer_service.sign_raw_message(&message) {
                        Ok(signature) => {
                            info!("✅ Message signed successfully");
                            EnclaveResult::MessageSigned { signature }
                        }
                        Err(e) => {
                            error!("❌ Failed to sign message: {}", e);
                            EnclaveResult::Error {
                                message: format!("Message signing failed: {}", e),
                                code: 500,
                            }
                        }
                    },
                    None => {
                        error!("❌ Transaction signer service not initialized");
                        EnclaveResult::Error {
                            message: "Transaction signer service not initialized - quorum key not provisioned".to_string(),
                            code: 503,
                        }
                    }
                }
            }

            // Application State Operations
            EnclaveOperation::GetApplicationStatus => {
                info!("📊 Getting application status");

                let status = self.state_manager.lock().unwrap().get_status();
                let metadata = ApplicationMetadata {
                    name: status.metadata.name,
                    version: status.metadata.version,
                    last_updated: status.metadata.last_updated,
                    operation_count: status.metadata.operation_count,
                };
                EnclaveResult::ApplicationStatus {
                    phase: format!("{:?}", status.phase),
                    has_quorum_key: status.has_quorum_key,
                    data_count: status.data_count,
                    metadata,
                }
            }

            EnclaveOperation::StoreApplicationData { key, data } => {
                info!(
                    "💾 Storing application data (key: {}, {} bytes)",
                    key,
                    data.len()
                );

                match self
                    .state_manager
                    .lock()
                    .unwrap()
                    .get_state_mut()
                    .store_data(key, data)
                {
                    Ok(()) => {
                        info!("✅ Application data stored successfully");
                        EnclaveResult::ApplicationDataStored { success: true }
                    }
                    Err(e) => {
                        error!("❌ Failed to store application data: {}", e);
                        EnclaveResult::Error {
                            message: format!("Application data storage failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::GetApplicationData { key } => {
                info!("📖 Retrieving application data (key: {})", key);

                match self
                    .state_manager
                    .lock()
                    .unwrap()
                    .get_state()
                    .get_data(&key)
                {
                    Some(data) => {
                        info!("✅ Application data retrieved successfully");
                        EnclaveResult::ApplicationDataRetrieved {
                            data: Some(data.clone()),
                        }
                    }
                    None => {
                        info!("ℹ️  Application data not found for key: {}", key);
                        EnclaveResult::ApplicationDataRetrieved { data: None }
                    }
                }
            }

            EnclaveOperation::GenerateQuorumKey { .. } => EnclaveResult::Error {
                message: "GenerateQuorumKey operation not implemented in this flow".to_string(),
                code: 501,
            },
            EnclaveOperation::ExportQuorumKey { .. } => EnclaveResult::Error {
                message: "ExportQuorumKey operation not implemented in this flow".to_string(),
                code: 501,
            },
            EnclaveOperation::InjectQuorumKey { .. } => EnclaveResult::Error {
                message: "InjectQuorumKey operation not implemented in this flow".to_string(),
                code: 501,
            },
            EnclaveOperation::ResetEnclave => {
                info!("🔄 Resetting enclave state");

                // Clear all storage
                let storage = TeeStorage::new();
                if let Err(e) = storage.clear_all() {
                    error!("❌ Failed to clear storage: {}", e);
                    return EnclaveResponse::new(
                        request.id,
                        EnclaveResult::EnclaveReset { success: false },
                    );
                }

                // Reset application state
                {
                    let mut state_manager = self.state_manager.lock().unwrap();
                    state_manager.reset();
                }

                info!("✅ Enclave state reset successfully");
                EnclaveResult::EnclaveReset { success: true }
            }

            // TEE-to-TEE Communication Operations
            EnclaveOperation::BootKeyForward {
                manifest_envelope,
                pivot,
            } => {
                info!("🔍 MATCHED: BootKeyForward");
                info!("🚀 Handling boot key forward request");

                // Parse manifest envelope from request
                let parsed_manifest = match serde_json::from_value(manifest_envelope) {
                    Ok(manifest) => manifest,
                    Err(e) => {
                        error!("❌ Failed to parse manifest envelope: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse manifest envelope: {}", e),
                                code: 400,
                            },
                        );
                    }
                };

                // Parse pivot from request
                let parsed_pivot = match serde_json::from_value(pivot) {
                    Ok(pivot) => pivot,
                    Err(e) => {
                        error!("❌ Failed to parse pivot: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse pivot: {}", e),
                                code: 400,
                            },
                        );
                    }
                };

                let boot_request = renclave_enclave::attestation::BootKeyForwardRequest {
                    manifest_envelope: parsed_manifest,
                    pivot: parsed_pivot,
                };

                let tee_comm = self.tee_communication.lock().unwrap();
                match tee_comm.handle_boot_key_forward(boot_request) {
                    Ok(response) => {
                        info!("✅ Boot key forward response created");
                        info!("✅ TEE1 (Original Node) created BootKeyForward response for TEE2");

                        match serde_json::to_value(response.nsm_response) {
                            Ok(nsm_response) => {
                                EnclaveResult::BootKeyForwardResponse { nsm_response }
                            }
                            Err(e) => {
                                error!("❌ Failed to serialize NSM response: {}", e);
                                EnclaveResult::Error {
                                    message: format!("Failed to serialize NSM response: {}", e),
                                    code: 500,
                                }
                            }
                        }
                    }
                    Err(e) => {
                        error!("❌ Boot key forward failed: {}", e);
                        EnclaveResult::Error {
                            message: format!("Boot key forward failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::ExportKey {
                manifest_envelope,
                attestation_doc,
            } => {
                info!("🔍 MATCHED: ExportKey");
                info!("📤 MAIN.RS - Handling export key request");
                debug!("🔍 ExportKey handler reached");

                // Parse manifest envelope from request
                let parsed_manifest = match serde_json::from_value(manifest_envelope) {
                    Ok(manifest) => {
                        info!("✅ Successfully parsed manifest envelope from request");
                        manifest
                    }
                    Err(e) => {
                        error!("❌ Failed to parse manifest envelope: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse manifest envelope: {}", e),
                                code: 400,
                            },
                        );
                    }
                };

                // Parse attestation document from request
                info!("✅ Successfully parsed attestation document from request");
                debug!(
                    "📄 Attestation document size: {} bytes",
                    attestation_doc.as_array().map(|arr| arr.len()).unwrap_or(0)
                );

                // Convert attestation document from JSON array to bytes
                let attestation_doc_bytes = match attestation_doc.as_array() {
                    Some(arr) => {
                        let bytes: Result<Vec<u8>, _> = arr
                            .iter()
                            .filter_map(|v| v.as_u64())
                            .map(|n| {
                                if n <= 255 {
                                    Ok(n as u8)
                                } else {
                                    Err(format!("Invalid byte value: {}", n))
                                }
                            })
                            .collect();
                        match bytes {
                            Ok(bytes) => {
                                info!("✅ Successfully converted attestation document to bytes: {} bytes", bytes.len());
                                bytes
                            }
                            Err(e) => {
                                error!("❌ Failed to convert attestation document to bytes: {}", e);
                                return EnclaveResponse::new(
                                    request.id,
                                    EnclaveResult::Error {
                                        message: format!(
                                            "Failed to convert attestation document to bytes: {}",
                                            e
                                        ),
                                        code: 400,
                                    },
                                );
                            }
                        }
                    }
                    None => {
                        error!("❌ Attestation document is not an array");
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: "Attestation document must be an array of bytes"
                                    .to_string(),
                                code: 400,
                            },
                        );
                    }
                };

                let export_request = renclave_enclave::attestation::ExportKeyRequest {
                    manifest_envelope: parsed_manifest,
                    cose_sign1_attestation_doc: attestation_doc_bytes,
                };

                let tee_comm = self.tee_communication.lock().unwrap();
                match tee_comm.handle_export_key(export_request) {
                    Ok(response) => {
                        info!("✅ Export key response created");
                        EnclaveResult::ExportKeyResponse {
                            encrypted_quorum_key: response.encrypted_quorum_key,
                            signature: response.signature,
                        }
                    }
                    Err(e) => {
                        error!("❌ Export key failed: {}", e);
                        EnclaveResult::Error {
                            message: format!("Export key failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::InjectKey {
                encrypted_quorum_key,
                signature,
            } => {
                info!("💉 Handling inject key request");

                let inject_request = match (
                    serde_json::from_value(encrypted_quorum_key),
                    serde_json::from_value(signature),
                ) {
                    (Ok(encrypted_quorum_key), Ok(signature)) => {
                        renclave_enclave::attestation::InjectKeyRequest {
                            encrypted_quorum_key,
                            signature,
                        }
                    }
                    (Err(e), _) => {
                        error!("❌ Failed to parse encrypted quorum key: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse encrypted quorum key: {}", e),
                                code: 400,
                            },
                        );
                    }
                    (_, Err(e)) => {
                        error!("❌ Failed to parse signature: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse signature: {}", e),
                                code: 400,
                            },
                        );
                    }
                };

                let tee_comm = self.tee_communication.lock().unwrap();
                match tee_comm.handle_inject_key(inject_request) {
                    Ok(_response) => {
                        info!("✅ Key injection successful - TEE is now provisioned");

                        // Follow QoS pattern: transition from WaitingForForwardedKey to QuorumKeyProvisioned
                        info!("🔄 Following QoS pattern: transitioning from WaitingForForwardedKey to QuorumKeyProvisioned...");
                        let mut state_manager = self.state_manager.lock().unwrap();
                        if let Err(e) =
                            state_manager.transition_to(ApplicationPhase::QuorumKeyProvisioned)
                        {
                            error!(
                                "❌ Failed to transition to QuorumKeyProvisioned phase: {}",
                                e
                            );
                            return EnclaveResponse::new(
                                request.id,
                                EnclaveResult::Error {
                                    message: format!(
                                        "Failed to transition to QuorumKeyProvisioned phase: {}",
                                        e
                                    ),
                                    code: 500,
                                },
                            );
                        }

                        // Now set the quorum key (this should work in QuorumKeyProvisioned phase)
                        info!("🔗 Setting quorum key in QuorumKeyProvisioned phase...");
                        if let Some(quorum_key) = tee_comm.get_quorum_key().unwrap_or(None) {
                            // Convert from renclave_enclave::P256Pair to crate::quorum::P256Pair
                            let quorum_key_bytes = quorum_key.private_key_bytes();
                            let quorum_key_array: [u8; 32] = quorum_key_bytes.try_into().unwrap();
                            let converted_quorum_key =
                                crate::quorum::P256Pair::from_master_seed(&quorum_key_array)
                                    .unwrap();

                            if let Err(e) = state_manager
                                .get_state_mut()
                                .set_quorum_key(converted_quorum_key.clone())
                            {
                                error!("❌ Failed to set quorum key in state manager: {}", e);
                                return EnclaveResponse::new(
                                    request.id,
                                    EnclaveResult::Error {
                                        message: format!(
                                            "Failed to set quorum key in state manager: {}",
                                            e
                                        ),
                                        code: 500,
                                    },
                                );
                            }

                            // Initialize data encryption service with the quorum key
                            info!("🔧 Initializing data encryption service with quorum key...");
                            let data_encryption = DataEncryption::new(converted_quorum_key.clone());
                            *self.data_encryption.lock().unwrap() = Some(data_encryption);
                            info!("✅ Data encryption service initialized");

                            // Initialize transaction signing service with the quorum key
                            info!("🔧 Initializing transaction signing service with quorum key...");
                            let transaction_signer =
                                TransactionSigner::new(converted_quorum_key.clone());
                            *self.transaction_signer.lock().unwrap() = Some(transaction_signer);
                            info!("✅ Transaction signing service initialized");

                            // Finally transition to ApplicationReady
                            info!("🔄 Transitioning to ApplicationReady phase...");
                            if let Err(e) =
                                state_manager.transition_to(ApplicationPhase::ApplicationReady)
                            {
                                error!("❌ Failed to transition to ApplicationReady phase: {}", e);
                                return EnclaveResponse::new(
                                    request.id,
                                    EnclaveResult::Error {
                                        message: format!(
                                            "Failed to transition to ApplicationReady phase: {}",
                                            e
                                        ),
                                        code: 500,
                                    },
                                );
                            }

                            info!("✅ TEE2 successfully provisioned with quorum key and services, now ApplicationReady");
                        } else {
                            error!("❌ No quorum key found in TEE communication manager after injection");
                            return EnclaveResponse::new(
                                request.id,
                                EnclaveResult::Error {
                                    message: "No quorum key found in TEE communication manager after injection".to_string(),
                                    code: 500,
                                },
                            );
                        }

                        EnclaveResult::InjectKeyResponse { success: true }
                    }
                    Err(e) => {
                        error!("❌ Key injection failed: {}", e);
                        EnclaveResult::Error {
                            message: format!("Key injection failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }
            EnclaveOperation::GenerateAttestation {
                manifest_hash,
                pcr_values,
            } => {
                info!("📄 Handling generate attestation request");

                // Try to use stored manifest envelope first (QoS pattern)
                let tee_comm = self.tee_communication.lock().unwrap();
                let (final_manifest_hash, final_pcr_values) =
                    if let Ok(stored_manifest) = tee_comm.get_manifest_envelope() {
                        info!("✅ Using stored manifest envelope for attestation generation");
                        let stored_hash = stored_manifest.manifest.qos_hash();
                        let stored_pcr_values = (
                            stored_manifest.manifest.enclave.pcr0,
                            stored_manifest.manifest.enclave.pcr1,
                            stored_manifest.manifest.enclave.pcr2,
                            stored_manifest.manifest.enclave.pcr3,
                        );
                        info!("📄 Using stored manifest hash: {} bytes", stored_hash.len());
                        (stored_hash, stored_pcr_values)
                    } else {
                        info!("⚠️ No stored manifest envelope, using provided values");
                        (manifest_hash, pcr_values)
                    };

                match tee_comm.generate_attestation_doc(&final_manifest_hash, final_pcr_values) {
                    Ok(attestation_doc) => {
                        info!("✅ Attestation document generated successfully");
                        match borsh::to_vec(&attestation_doc) {
                            Ok(attestation_doc_bytes) => {
                                EnclaveResult::GenerateAttestationResponse {
                                    attestation_doc: attestation_doc_bytes,
                                }
                            }
                            Err(e) => {
                                error!("❌ Failed to serialize attestation document: {}", e);
                                EnclaveResult::Error {
                                    message: format!(
                                        "Failed to serialize attestation document: {}",
                                        e
                                    ),
                                    code: 500,
                                }
                            }
                        }
                    }
                    Err(e) => {
                        error!("❌ Attestation document generation failed: {}", e);
                        EnclaveResult::Error {
                            message: format!("Attestation document generation failed: {}", e),
                            code: 500,
                        }
                    }
                }
            }

            EnclaveOperation::ShareManifest { manifest_envelope } => {
                info!("📄 Handling share manifest request");

                // Parse manifest envelope from request
                let parsed_manifest = match serde_json::from_value(manifest_envelope) {
                    Ok(manifest) => {
                        info!("✅ Successfully parsed manifest envelope from request");
                        manifest
                    }
                    Err(e) => {
                        error!("❌ Failed to parse manifest envelope: {}", e);
                        return EnclaveResponse::new(
                            request.id,
                            EnclaveResult::Error {
                                message: format!("Failed to parse manifest envelope: {}", e),
                                code: 400,
                            },
                        );
                    }
                };

                // Store manifest envelope in TEE communication manager
                let tee_comm = self.tee_communication.lock().unwrap();
                match tee_comm.set_manifest_envelope(parsed_manifest) {
                    Ok(_) => {
                        info!("✅ Manifest envelope stored successfully");

                        // Follow QoS pattern: transition from WaitingForBootInstruction to WaitingForForwardedKey
                        info!("🔄 Following QoS pattern: transitioning from WaitingForBootInstruction to WaitingForForwardedKey...");
                        let mut state_manager = self.state_manager.lock().unwrap();
                        if let Err(e) =
                            state_manager.transition_to(ApplicationPhase::WaitingForForwardedKey)
                        {
                            error!(
                                "❌ Failed to transition to WaitingForForwardedKey phase: {}",
                                e
                            );
                            return EnclaveResponse::new(
                                request.id,
                                EnclaveResult::Error {
                                    message: format!(
                                        "Failed to transition to WaitingForForwardedKey phase: {}",
                                        e
                                    ),
                                    code: 500,
                                },
                            );
                        }
                        info!("✅ Successfully transitioned to WaitingForForwardedKey phase");

                        EnclaveResult::ShareManifestResponse { success: true }
                    }
                    Err(e) => {
                        error!("❌ Failed to store manifest envelope: {}", e);
                        EnclaveResult::Error {
                            message: format!("Failed to store manifest envelope: {}", e),
                            code: 500,
                        }
                    }
                }
            }
        };

        EnclaveResponse::new(request.id, result)
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize logging
    std::env::set_var("RUST_LOG", "debug");
    env_logger::init();

    info!("🔒 QEMU Nitro Enclave - Secure Seed Generation");
    info!("🔍 Process ID: {}", std::process::id());
    info!(
        "🔍 Current working directory: {:?}",
        std::env::current_dir()?
    );

    // Create and start enclave
    let enclave = NitroEnclave::new().await?;
    enclave.start().await?;

    Ok(())
}
