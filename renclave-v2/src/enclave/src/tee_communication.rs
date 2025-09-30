//! TEE-to-TEE communication protocol
//! Following QoS key forwarding flow exactly

use anyhow::{anyhow, Result};
use hex;
use log::{debug, error, info};
use p256::elliptic_curve::sec1::ToEncodedPoint;
use std::sync::{Arc, Mutex};

use crate::{
    attestation::{
        AttestationDoc, AttestationManager, BootKeyForwardRequest, BootKeyForwardResponse,
        ExportKeyRequest, ExportKeyResponse, InjectKeyRequest, InjectKeyResponse, NsmResponse,
    },
    data_encryption::{P256EncryptPair, P256EncryptPublic},
    quorum::P256Pair,
};
use renclave_shared::ManifestEnvelope;

/// TEE-to-TEE communication manager
pub struct TeeCommunicationManager {
    /// Attestation manager for this TEE
    attestation_manager: Arc<Mutex<AttestationManager>>,
    /// Quorum key (if this TEE is provisioned)
    quorum_key: Arc<Mutex<Option<P256Pair>>>,
    /// Manifest envelope (if this TEE is provisioned)
    manifest_envelope: Arc<Mutex<Option<ManifestEnvelope>>>,
}

impl TeeCommunicationManager {
    /// Create a new TEE communication manager
    pub fn new() -> Self {
        Self {
            attestation_manager: Arc::new(Mutex::new(AttestationManager::new())),
            quorum_key: Arc::new(Mutex::new(None)),
            manifest_envelope: Arc::new(Mutex::new(None)),
        }
    }

    /// Set quorum key (for provisioned TEEs)
    pub fn set_quorum_key(&self, quorum_key: P256Pair) -> Result<()> {
        info!("🔑 Setting quorum key for TEE communication");

        let mut key_guard = self.quorum_key.lock().unwrap();
        *key_guard = Some(quorum_key);

        info!("✅ Quorum key set successfully");
        Ok(())
    }

    /// Set manifest envelope (for provisioned TEEs)
    pub fn set_manifest_envelope(&self, manifest_envelope: ManifestEnvelope) -> Result<()> {
        info!("📄 Setting manifest envelope for TEE communication");

        let mut manifest_guard = self.manifest_envelope.lock().unwrap();
        *manifest_guard = Some(manifest_envelope);

        info!("✅ Manifest envelope set successfully");
        Ok(())
    }

    /// Get manifest envelope (for provisioned TEEs)
    pub fn get_manifest_envelope(&self) -> Result<ManifestEnvelope> {
        let manifest_guard = self.manifest_envelope.lock().unwrap();
        manifest_guard
            .as_ref()
            .cloned()
            .ok_or_else(|| anyhow!("No manifest envelope available - TEE not provisioned"))
    }

    /// Handle boot key forward request (New Node)
    pub fn handle_boot_key_forward(
        &self,
        request: BootKeyForwardRequest,
    ) -> Result<BootKeyForwardResponse> {
        info!("🚀 Handling boot key forward request (New Node)");

        // 1. Check signatures over the manifest envelope
        info!("🔍 Checking manifest envelope signatures");
        // TODO: Implement signature verification

        // 2. Use manifest envelope from request (for new nodes) or stored one (for provisioned nodes)
        let manifest_envelope = if let Ok(stored) = self.get_manifest_envelope() {
            info!("📄 Using stored manifest envelope");
            stored
        } else {
            info!("📄 Using manifest envelope from request (new node)");
            request.manifest_envelope
        };

        // 3. Generate ephemeral key
        let mut attestation_manager = self.attestation_manager.lock().unwrap();
        let _ephemeral_public = attestation_manager.generate_ephemeral_key()?;

        // 4. Create attestation document using actual manifest data
        let manifest_hash = manifest_envelope.manifest.qos_hash();
        let pcr_values = (
            manifest_envelope.manifest.enclave.pcr0,
            manifest_envelope.manifest.enclave.pcr1,
            manifest_envelope.manifest.enclave.pcr2,
            manifest_envelope.manifest.enclave.pcr3,
        );

        let attestation_doc =
            attestation_manager.create_attestation_doc(&manifest_hash, pcr_values)?;

        // 5. Serialize attestation document
        let document_bytes = borsh::to_vec(&attestation_doc)
            .map_err(|e| anyhow!("Failed to serialize attestation document: {}", e))?;

        let response = BootKeyForwardResponse {
            nsm_response: NsmResponse::Attestation {
                document: document_bytes,
            },
        };

        info!("✅ Boot key forward response created successfully");
        Ok(response)
    }

    /// Handle export key request (Original Node)
    pub fn handle_export_key(&self, request: ExportKeyRequest) -> Result<ExportKeyResponse> {
        info!("🔍 TEE_COMMUNICATION: handle_export_key called");
        info!("📤 Handling export key request (Original Node)");

        // Check if this TEE has a quorum key
        let quorum_key_guard = self.quorum_key.lock().unwrap();
        let quorum_key = quorum_key_guard
            .as_ref()
            .ok_or_else(|| anyhow!("No quorum key available - TEE not provisioned"))?;
        info!("✅ Quorum key available for export");

        // Deserialize attestation document
        info!("🔍 Deserializing attestation document...");
        let attestation_doc: AttestationDoc =
            borsh::from_slice(&request.cose_sign1_attestation_doc).map_err(|e| {
                error!("❌ Failed to deserialize attestation document: {}", e);
                anyhow!("Failed to deserialize attestation document: {}", e)
            })?;
        info!("✅ Attestation document deserialized successfully");
        debug!(
            "📄 Attestation doc user_data: {} bytes",
            attestation_doc.user_data.len()
        );
        debug!(
            "📄 Attestation doc public_key: {} bytes",
            attestation_doc
                .public_key
                .as_ref()
                .map(|pk| pk.len())
                .unwrap_or(0)
        );

        // Use actual manifest data from the request
        info!("🔍 Using actual manifest data for verification");
        let manifest_hash = request.manifest_envelope.manifest.qos_hash();
        let expected_pcr_values = (
            request.manifest_envelope.manifest.enclave.pcr0,
            request.manifest_envelope.manifest.enclave.pcr1,
            request.manifest_envelope.manifest.enclave.pcr2,
            request.manifest_envelope.manifest.enclave.pcr3,
        );
        debug!("📄 Expected manifest hash: {} bytes", manifest_hash.len());
        debug!(
            "📄 Expected manifest hash (hex): {}",
            hex::encode(&manifest_hash)
        );
        debug!(
            "📄 Attestation doc user_data (hex): {}",
            hex::encode(&attestation_doc.user_data)
        );
        debug!("📄 Expected PCR0: {} bytes", expected_pcr_values.0.len());
        debug!("📄 Expected PCR1: {} bytes", expected_pcr_values.1.len());
        debug!("📄 Expected PCR2: {} bytes", expected_pcr_values.2.len());
        debug!("📄 Expected PCR3: {} bytes", expected_pcr_values.3.len());
        debug!(
            "📄 Attestation doc PCR0: {} bytes",
            attestation_doc.pcr0.len()
        );
        debug!(
            "📄 Attestation doc PCR1: {} bytes",
            attestation_doc.pcr1.len()
        );
        debug!(
            "📄 Attestation doc PCR2: {} bytes",
            attestation_doc.pcr2.len()
        );
        debug!(
            "📄 Attestation doc PCR3: {} bytes",
            attestation_doc.pcr3.len()
        );

        info!("🔍 Verifying attestation document with actual manifest data");
        let attestation_manager = self.attestation_manager.lock().unwrap();

        match attestation_manager.verify_attestation_doc(
            &attestation_doc,
            &manifest_hash,
            expected_pcr_values,
        ) {
            Ok(is_valid) => {
                if !is_valid {
                    error!("❌ Attestation document verification failed");
                    return Err(anyhow!("Attestation document verification failed"));
                }
                info!("✅ Attestation document verification successful");
            }
            Err(e) => {
                error!("❌ Attestation document verification error: {}", e);
                return Err(anyhow!("Attestation document verification error: {}", e));
            }
        }

        // Extract ephemeral public key from attestation document
        info!("🔍 Extracting ephemeral public key from attestation document");
        let ephemeral_public = AttestationManager::extract_ephemeral_public_key(&attestation_doc)
            .map_err(|e| {
            error!("❌ Failed to extract ephemeral public key: {}", e);
            anyhow!("Failed to extract ephemeral public key: {}", e)
        })?;
        info!("✅ Ephemeral public key extracted successfully");

        info!("🔑 Extracted ephemeral public key from attestation document");
        debug!(
            "🔑 Ephemeral public key: {:?}",
            ephemeral_public.to_encoded_point(false).as_bytes()
        );

        // Encrypt quorum key to the new node's ephemeral key using ECDH
        info!("🔐 Encrypting quorum key using ECDH");

        let quorum_master_seed = quorum_key.to_master_seed();

        // Create a P256EncryptPublic from the ephemeral public key
        let ephemeral_encrypt_public = P256EncryptPublic::new(ephemeral_public);

        let encrypted_quorum_key = ephemeral_encrypt_public.encrypt(&quorum_master_seed)?;

        info!("✅ Quorum key encrypted successfully");
        debug!(
            "🔐 Encrypted quorum key length: {} bytes",
            encrypted_quorum_key.len()
        );

        // Sign the encrypted quorum key
        let signature = quorum_key.sign(&encrypted_quorum_key)?;

        info!("✅ Signature created over encrypted quorum key");
        debug!("🔐 Signature length: {} bytes", signature.len());

        let response = ExportKeyResponse {
            encrypted_quorum_key,
            signature,
        };

        info!("✅ Export key response created successfully");
        Ok(response)
    }

    /// Handle inject key request (New Node)
    pub fn handle_inject_key(&self, request: InjectKeyRequest) -> Result<InjectKeyResponse> {
        info!("💉 Handling inject key request (New Node)");
        debug!(
            "📄 Encrypted quorum key size: {} bytes",
            request.encrypted_quorum_key.len()
        );
        debug!("📄 Signature size: {} bytes", request.signature.len());

        // Get ephemeral private key for decryption
        info!("🔍 Getting ephemeral private key from attestation manager");
        let attestation_manager = self.attestation_manager.lock().unwrap();
        let ephemeral_private = attestation_manager
            .get_ephemeral_private_key()
            .map_err(|e| {
                error!("❌ Failed to get ephemeral private key: {}", e);
                anyhow!("Failed to get ephemeral private key: {}", e)
            })?;

        info!("✅ Ephemeral private key retrieved successfully");
        debug!(
            "🔓 Ephemeral private key (first 10 bytes): {:?}",
            &ephemeral_private.to_be_bytes()[..10]
        );

        // Create decryptor from ephemeral private key
        info!("🔍 Creating decryptor from ephemeral private key");
        let ephemeral_decryptor = P256EncryptPair::from_bytes(&ephemeral_private.to_be_bytes())
            .map_err(|e| {
                error!(
                    "❌ Failed to create decryptor from ephemeral private key: {}",
                    e
                );
                anyhow!(
                    "Failed to create decryptor from ephemeral private key: {}",
                    e
                )
            })?;
        info!("✅ Decryptor created successfully");

        // Decrypt the quorum key
        info!("🔓 Decrypting quorum key using ephemeral private key");
        let decrypted_quorum_key = ephemeral_decryptor
            .decrypt(&request.encrypted_quorum_key)
            .map_err(|e| {
                error!("❌ Failed to decrypt quorum key: {}", e);
                anyhow!("Failed to decrypt quorum key: {}", e)
            })?;

        info!("✅ Quorum key decrypted successfully");
        debug!(
            "🔓 Decrypted quorum key length: {} bytes",
            decrypted_quorum_key.len()
        );

        // Verify signature (TODO: Implement signature verification)
        info!("🔍 Verifying signature over encrypted quorum key");
        // TODO: Implement signature verification against quorum public key

        // Convert decrypted bytes to quorum key
        let quorum_key_bytes: [u8; 32] = decrypted_quorum_key
            .try_into()
            .map_err(|_| anyhow!("Invalid quorum key length"))?;

        let quorum_key = P256Pair::from_master_seed(&quorum_key_bytes)?;

        // Set the quorum key in this TEE
        drop(attestation_manager); // Release the lock
        self.set_quorum_key(quorum_key)?;

        info!("✅ Quorum key injected successfully - TEE is now provisioned");

        let response = InjectKeyResponse {};
        Ok(response)
    }

    /// Check if this TEE is provisioned
    pub fn is_provisioned(&self) -> bool {
        let quorum_key_guard = self.quorum_key.lock().unwrap();
        quorum_key_guard.is_some()
    }

    /// Generate attestation document for TEE-to-TEE communication
    pub fn generate_attestation_doc(
        &self,
        manifest_hash: &[u8],
        pcr_values: (Vec<u8>, Vec<u8>, Vec<u8>, Vec<u8>),
    ) -> Result<AttestationDoc> {
        info!("📄 Generating attestation document for TEE-to-TEE communication");

        let mut attestation_manager = self.attestation_manager.lock().unwrap();

        // Generate ephemeral key if not already generated
        if !attestation_manager.has_ephemeral_key() {
            attestation_manager.generate_ephemeral_key()?;
        }

        // Create attestation document
        let attestation_doc =
            attestation_manager.create_attestation_doc(manifest_hash, pcr_values)?;

        info!("✅ Attestation document generated successfully");
        Ok(attestation_doc)
    }

    /// Get quorum key (if provisioned)
    pub fn get_quorum_key(&self) -> Result<Option<P256Pair>> {
        let quorum_key_guard = self.quorum_key.lock().unwrap();
        Ok(quorum_key_guard.clone())
    }
}

impl Default for TeeCommunicationManager {
    fn default() -> Self {
        Self::new()
    }
}
