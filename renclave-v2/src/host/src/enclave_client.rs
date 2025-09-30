use anyhow::{anyhow, Context, Result};
use log::{debug, error, info, warn};
use std::time::Duration;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::UnixStream;
use tokio::time::{sleep, timeout};

use renclave_shared::{
    DecryptedShare, EnclaveOperation, EnclaveRequest, EnclaveResponse, EncryptedQuorumKey,
    ManifestEnvelope, QuorumMember,
};

/// Client for communicating with the Nitro Enclave
#[derive(Debug)]
pub struct EnclaveClient {
    socket_path: String,
}

impl EnclaveClient {
    /// Create new enclave client
    pub fn new(socket_path: String) -> Self {
        Self { socket_path }
    }

    /// Wait for enclave to become available
    pub async fn wait_for_enclave(&self, max_wait: Duration) -> Result<()> {
        info!(
            "⏳ Waiting for enclave to become available at: {}",
            self.socket_path
        );

        let start_time = tokio::time::Instant::now();
        let mut attempt = 0;

        while start_time.elapsed() < max_wait {
            attempt += 1;
            debug!("🔍 Attempt {} to connect to enclave", attempt);

            match self.test_connection().await {
                Ok(_) => {
                    info!(
                        "✅ Enclave is available after {:?} (attempt {})",
                        start_time.elapsed(),
                        attempt
                    );
                    return Ok(());
                }
                Err(e) => {
                    debug!("❌ Connection attempt {} failed: {}", attempt, e);
                    sleep(Duration::from_millis(500)).await;
                }
            }
        }

        Err(anyhow!("Enclave not available after {:?}", max_wait))
    }

    /// Test connection to enclave
    async fn test_connection(&self) -> Result<()> {
        let stream = UnixStream::connect(&self.socket_path)
            .await
            .context("Failed to connect to enclave socket")?;

        drop(stream);
        Ok(())
    }

    /// Send request to enclave and get response
    pub async fn send_request(&self, operation: EnclaveOperation) -> Result<EnclaveResponse> {
        let operation_discriminant = std::mem::discriminant(&operation);
        let request = EnclaveRequest::new(operation);
        info!(
            "📤 Sending request to enclave: {} (operation: {:?})",
            request.id, operation_discriminant
        );
        debug!("🔍 Socket path: {}", self.socket_path);

        // Connect to enclave with timeout
        info!("🔗 Connecting to enclave socket...");
        let stream = timeout(
            Duration::from_secs(5),
            UnixStream::connect(&self.socket_path),
        )
        .await
        .context("Timeout connecting to enclave")?
        .context("Failed to connect to enclave socket")?;
        info!("✅ Connected to enclave socket successfully");

        // Send request with timeout
        info!("📡 Sending request to enclave with 30s timeout...");
        let response = timeout(
            Duration::from_secs(30),
            self.send_request_internal(stream, request),
        )
        .await
        .context("Timeout waiting for enclave response")??;

        info!("📨 Received response from enclave: {}", response.id);
        debug!(
            "🔍 Response result: {:?}",
            std::mem::discriminant(&response.result)
        );
        Ok(response)
    }

    /// Internal method to send request and receive response
    async fn send_request_internal(
        &self,
        mut stream: UnixStream,
        request: EnclaveRequest,
    ) -> Result<EnclaveResponse> {
        // Serialize and send request
        let request_json =
            serde_json::to_string(&request).context("Failed to serialize request")?;

        stream
            .write_all(request_json.as_bytes())
            .await
            .context("Failed to write request to socket")?;
        stream
            .write_all(b"\n")
            .await
            .context("Failed to write newline to socket")?;

        debug!("✅ Request sent to enclave");

        // Read response
        let mut reader = BufReader::new(stream);
        let mut response_line = String::new();

        reader
            .read_line(&mut response_line)
            .await
            .context("Failed to read response from enclave")?;

        if response_line.trim().is_empty() {
            return Err(anyhow!("Received empty response from enclave"));
        }

        debug!("📥 Raw response from enclave: {}", response_line.trim());

        // Deserialize response
        let response: EnclaveResponse = serde_json::from_str(response_line.trim())
            .context("Failed to deserialize response from enclave")?;

        debug!("✅ Response deserialized successfully");
        Ok(response)
    }

    /// Generate seed phrase via enclave
    pub async fn generate_seed(
        &self,
        strength: u32,
        passphrase: Option<String>,
    ) -> Result<EnclaveResponse> {
        info!(
            "🔑 Requesting seed generation (strength: {} bits)",
            strength
        );

        let operation = EnclaveOperation::GenerateSeed {
            strength,
            passphrase,
        };
        self.send_request(operation).await
    }

    /// Validate seed phrase via enclave
    pub async fn validate_seed(
        &self,
        seed_phrase: String,
        encrypted_entropy: Option<String>,
    ) -> Result<EnclaveResponse> {
        info!("🔍 Requesting seed validation");

        let operation = EnclaveOperation::ValidateSeed {
            seed_phrase,
            encrypted_entropy,
        };
        self.send_request(operation).await
    }

    /// Get enclave information
    pub async fn get_info(&self) -> Result<EnclaveResponse> {
        debug!("ℹ️  Requesting enclave information");

        let operation = EnclaveOperation::GetInfo;
        self.send_request(operation).await
    }

    /// Derive key from encrypted seed phrase via enclave
    pub async fn derive_key(
        &self,
        encrypted_seed_phrase: String,
        path: String,
        curve: String,
    ) -> Result<EnclaveResponse> {
        info!(
            "🔑 Requesting key derivation (path: {}, curve: {})",
            path, curve
        );

        let operation = EnclaveOperation::DeriveKey {
            encrypted_seed_phrase,
            path,
            curve,
        };
        self.send_request(operation).await
    }

    /// Derive address from encrypted seed phrase via enclave
    pub async fn derive_address(
        &self,
        encrypted_seed_phrase: String,
        path: String,
        curve: String,
    ) -> Result<EnclaveResponse> {
        info!(
            "📍 Requesting address derivation (path: {}, curve: {})",
            path, curve
        );

        let operation = EnclaveOperation::DeriveAddress {
            encrypted_seed_phrase,
            path,
            curve,
        };
        self.send_request(operation).await
    }

    /// Generate quorum key via enclave
    pub async fn generate_quorum_key(
        &self,
        members: Vec<QuorumMember>,
        threshold: u32,
        dr_key: Option<Vec<u8>>,
    ) -> Result<EnclaveResponse> {
        info!(
            "🔐 Requesting quorum key generation (members: {}, threshold: {})",
            members.len(),
            threshold
        );

        let operation = EnclaveOperation::GenerateQuorumKey {
            members,
            threshold,
            dr_key,
        };
        self.send_request(operation).await
    }

    /// Export quorum key via enclave
    pub async fn export_quorum_key(
        &self,
        new_manifest_envelope: ManifestEnvelope,
        cose_sign1_attestation_document: Vec<u8>,
    ) -> Result<EnclaveResponse> {
        info!("📤 Requesting quorum key export");

        let operation = EnclaveOperation::ExportQuorumKey {
            new_manifest_envelope: Box::new(new_manifest_envelope),
            cose_sign1_attestation_document,
        };
        self.send_request(operation).await
    }

    /// Inject quorum key via enclave
    pub async fn inject_quorum_key(
        &self,
        encrypted_quorum_key: EncryptedQuorumKey,
    ) -> Result<EnclaveResponse> {
        info!("📥 Requesting quorum key injection");

        let operation = EnclaveOperation::InjectQuorumKey {
            encrypted_quorum_key,
        };
        self.send_request(operation).await
    }

    /// Execute Genesis Boot flow
    #[allow(clippy::too_many_arguments)]
    pub async fn genesis_boot(
        &self,
        namespace_name: String,
        namespace_nonce: u64,
        manifest_members: Vec<QuorumMember>,
        manifest_threshold: u32,
        share_members: Vec<QuorumMember>,
        share_threshold: u32,
        pivot_hash: [u8; 32],
        pivot_args: Vec<String>,
        dr_key: Option<Vec<u8>>,
    ) -> Result<EnclaveResponse> {
        info!("🌱 Requesting Genesis Boot flow execution");
        debug!("🔍 Genesis Boot parameters:");
        debug!("  - namespace_name: {}", namespace_name);
        debug!("  - namespace_nonce: {}", namespace_nonce);
        debug!("  - manifest_members: {} members", manifest_members.len());
        debug!("  - manifest_threshold: {}", manifest_threshold);
        debug!("  - share_members: {} members", share_members.len());
        debug!("  - share_threshold: {}", share_threshold);
        debug!("  - pivot_hash: {:?}", pivot_hash);
        debug!("  - pivot_args: {:?}", pivot_args);
        debug!(
            "  - dr_key: {}",
            if dr_key.is_some() { "provided" } else { "none" }
        );

        let operation = EnclaveOperation::GenesisBoot {
            namespace_name,
            namespace_nonce,
            manifest_members,
            manifest_threshold,
            share_members,
            share_threshold,
            pivot_hash,
            pivot_args,
            dr_key,
        };

        info!("📡 Sending Genesis Boot operation to enclave...");
        let result = self.send_request(operation).await;
        match &result {
            Ok(response) => {
                info!("✅ Genesis Boot operation completed successfully");
                debug!(
                    "🔍 Response result type: {:?}",
                    std::mem::discriminant(&response.result)
                );
            }
            Err(e) => {
                error!("❌ Genesis Boot operation failed: {}", e);
            }
        }
        result
    }

    /// Inject shares to complete Genesis Boot
    pub async fn inject_shares(
        &self,
        namespace_name: String,
        namespace_nonce: u64,
        shares: Vec<DecryptedShare>,
    ) -> Result<EnclaveResponse> {
        info!("🔐 Requesting share injection for Genesis Boot completion");
        debug!("🔍 Share injection parameters:");
        debug!("  - namespace_name: {}", namespace_name);
        debug!("  - namespace_nonce: {}", namespace_nonce);
        debug!("  - shares: {} shares", shares.len());

        let operation = EnclaveOperation::InjectShares {
            namespace_name,
            namespace_nonce,
            shares,
        };

        info!("📡 Sending share injection operation to enclave...");
        let result = self.send_request(operation).await;
        match &result {
            Ok(response) => {
                info!("✅ Share injection operation completed successfully");
                debug!(
                    "🔍 Response result type: {:?}",
                    std::mem::discriminant(&response.result)
                );
            }
            Err(e) => {
                error!("❌ Share injection operation failed: {}", e);
            }
        }
        result
    }

    /// Check enclave health
    pub async fn health_check(&self) -> Result<bool> {
        debug!("🏥 Performing enclave health check");

        match self.test_connection().await {
            Ok(_) => {
                debug!("✅ Enclave health check passed");
                Ok(true)
            }
            Err(e) => {
                warn!("⚠️  Enclave health check failed: {}", e);
                Ok(false)
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_enclave_client_creation() {
        let client = EnclaveClient::new("/tmp/test_enclave.sock".to_string());
        assert_eq!(client.socket_path, "/tmp/test_enclave.sock");
    }

    #[tokio::test]
    async fn test_wait_for_enclave_timeout() {
        let client = EnclaveClient::new("/tmp/nonexistent_enclave.sock".to_string());
        let result = client.wait_for_enclave(Duration::from_millis(100)).await;
        assert!(result.is_err());
    }
}
