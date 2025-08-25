use anyhow::{anyhow, Context, Result};
use log::{debug, info, warn};
use std::time::Duration;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::UnixStream;
use tokio::time::{sleep, timeout};

use renclave_shared::{EnclaveOperation, EnclaveRequest, EnclaveResponse};

/// Client for communicating with the Nitro Enclave
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
        let request = EnclaveRequest::new(operation);
        debug!("📤 Sending request to enclave: {}", request.id);

        // Connect to enclave with timeout
        let stream = timeout(
            Duration::from_secs(5),
            UnixStream::connect(&self.socket_path),
        )
        .await
        .context("Timeout connecting to enclave")?
        .context("Failed to connect to enclave socket")?;

        // Send request with timeout
        let response = timeout(
            Duration::from_secs(30),
            self.send_request_internal(stream, request),
        )
        .await
        .context("Timeout waiting for enclave response")??;

        debug!("📨 Received response from enclave: {}", response.id);
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
        let response: EnclaveResponse = serde_json::from_str(&response_line.trim())
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
    pub async fn validate_seed(&self, seed_phrase: String) -> Result<EnclaveResponse> {
        info!("🔍 Requesting seed validation");

        let operation = EnclaveOperation::ValidateSeed { seed_phrase };
        self.send_request(operation).await
    }

    /// Get enclave information
    pub async fn get_info(&self) -> Result<EnclaveResponse> {
        debug!("ℹ️  Requesting enclave information");

        let operation = EnclaveOperation::GetInfo;
        self.send_request(operation).await
    }

    /// Derive key from seed phrase via enclave
    pub async fn derive_key(
        &self,
        seed_phrase: String,
        path: String,
        curve: String,
    ) -> Result<EnclaveResponse> {
        info!(
            "🔑 Requesting key derivation (path: {}, curve: {})",
            path, curve
        );

        let operation = EnclaveOperation::DeriveKey {
            seed_phrase,
            path,
            curve,
        };
        self.send_request(operation).await
    }

    /// Derive address from seed phrase via enclave
    pub async fn derive_address(
        &self,
        seed_phrase: String,
        path: String,
        curve: String,
    ) -> Result<EnclaveResponse> {
        info!(
            "📍 Requesting address derivation (path: {}, curve: {})",
            path, curve
        );

        let operation = EnclaveOperation::DeriveAddress {
            seed_phrase,
            path,
            curve,
        };
        self.send_request(operation).await
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
    use tokio::time::sleep;

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
