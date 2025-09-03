use log::{debug, error, info, warn};
use std::sync::Arc;
use tokio::fs;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{UnixListener, UnixStream};
use uuid::Uuid;

mod nitro;
mod seed_generator;

use renclave_network::{NetworkConfig, NetworkManager};
use renclave_shared::{EnclaveOperation, EnclaveRequest, EnclaveResponse, EnclaveResult};
use seed_generator::SeedGenerator;

/// QEMU Nitro Enclave for secure seed generation
pub struct NitroEnclave {
    seed_generator: Arc<SeedGenerator>,
    network_manager: Arc<NetworkManager>,
    enclave_id: String,
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
                                tokio::spawn(async move {
                                    if let Err(e) = Self::handle_client(
                                        stream,
                                        seed_generator,
                                        network_manager,
                                        enclave_id,
                                    )
                                    .await
                                    {
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
    async fn handle_client(
        stream: UnixStream,
        seed_generator: Arc<SeedGenerator>,
        network_manager: Arc<NetworkManager>,
        enclave_id: String,
    ) -> anyhow::Result<()> {
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
                    match serde_json::from_str::<EnclaveRequest>(request_json) {
                        Ok(request) => {
                            // Process request
                            let response = Self::process_request(
                                request,
                                &seed_generator,
                                &network_manager,
                                &enclave_id,
                            )
                            .await;

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
    async fn process_request(
        request: EnclaveRequest,
        seed_generator: &SeedGenerator,
        network_manager: &NetworkManager,
        enclave_id: &str,
    ) -> EnclaveResponse {
        debug!("⚙️  Processing request: {:?}", request.operation);

        let result = match request.operation {
            EnclaveOperation::GenerateSeed {
                strength,
                passphrase,
            } => {
                info!("🔑 Generating seed phrase (strength: {} bits)", strength);

                match seed_generator
                    .generate_seed(strength, passphrase.as_deref())
                    .await
                {
                    Ok(seed_result) => {
                        info!("✅ Seed phrase generated successfully");
                        EnclaveResult::SeedGenerated {
                            seed_phrase: seed_result.phrase,
                            entropy: seed_result.entropy,
                            strength: seed_result.strength,
                            word_count: seed_result.word_count,
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

                match seed_generator.validate_seed(&seed_phrase).await {
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

                let _network_status = network_manager.get_status().await;
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
                    enclave_id: enclave_id.to_string(),
                    capabilities,
                }
            }

            EnclaveOperation::DeriveKey {
                seed_phrase,
                path,
                curve,
            } => {
                info!("🔑 Deriving key (path: {}, curve: {})", path, curve);

                match seed_generator.derive_key(&seed_phrase, &path, &curve).await {
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

                match seed_generator
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
