//! Attestation document system for TEE-to-TEE communication
//! Following QoS patterns exactly

use anyhow::{anyhow, Result};
use borsh::{BorshDeserialize, BorshSerialize};
use log::{debug, info, warn};
use p256::{elliptic_curve::sec1::ToEncodedPoint, PublicKey, SecretKey};
use renclave_shared::ManifestEnvelope;
use serde::{Deserialize, Serialize};
use std::time::{SystemTime, UNIX_EPOCH};

/// Attestation document for TEE-to-TEE communication
/// Following QoS AttestationDoc structure
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct AttestationDoc {
    /// Timestamp in milliseconds since Unix epoch
    pub timestamp_ms: u64,
    /// Hash of the manifest (user_data field)
    pub user_data: Vec<u8>,
    /// Ephemeral public key for ECDH
    pub public_key: Option<Vec<u8>>,
    /// PCR values for verification
    pub pcr0: Vec<u8>,
    pub pcr1: Vec<u8>,
    pub pcr2: Vec<u8>,
    pub pcr3: Vec<u8>,
}

/// NSM Response types for attestation
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub enum NsmResponse {
    /// Attestation document response
    Attestation { document: Vec<u8> },
    /// Other NSM responses
    Other,
}

/// Boot key forward request
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct BootKeyForwardRequest {
    /// Manifest envelope for the new node
    pub manifest_envelope: ManifestEnvelope,
    /// Pivot application data
    pub pivot: Vec<u8>,
}

/// Boot key forward response
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct BootKeyForwardResponse {
    /// NSM response containing attestation document
    pub nsm_response: NsmResponse,
}

/// Export key request from client to original node
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct ExportKeyRequest {
    /// Manifest of the enclave requesting the quorum key
    pub manifest_envelope: ManifestEnvelope,
    /// Attestation document from the requesting enclave
    pub cose_sign1_attestation_doc: Vec<u8>,
}

/// Export key response from original node
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct ExportKeyResponse {
    /// Quorum key encrypted to the ephemeral key from attestation document
    pub encrypted_quorum_key: Vec<u8>,
    /// Signature over the encrypted quorum key
    pub signature: Vec<u8>,
}

/// Inject key request to new node
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct InjectKeyRequest {
    /// Quorum key encrypted to the ephemeral key of the target enclave
    pub encrypted_quorum_key: Vec<u8>,
    /// Signature over the encrypted quorum key
    pub signature: Vec<u8>,
}

/// Inject key response from new node
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize)]
pub struct InjectKeyResponse {
    // Empty response indicating success
}

/// Attestation document manager
pub struct AttestationManager {
    /// Ephemeral key pair for this TEE
    ephemeral_key: Option<SecretKey>,
}

impl AttestationManager {
    /// Create a new attestation manager
    pub fn new() -> Self {
        Self {
            ephemeral_key: None,
        }
    }

    /// Generate ephemeral key for attestation
    pub fn generate_ephemeral_key(&mut self) -> Result<PublicKey> {
        info!("🔑 Generating ephemeral key for attestation");

        let ephemeral_private = SecretKey::random(&mut rand::rngs::OsRng);
        let ephemeral_public = ephemeral_private.public_key();

        self.ephemeral_key = Some(ephemeral_private);

        info!("✅ Ephemeral key generated successfully");
        debug!(
            "🔑 Ephemeral public key: {:?}",
            ephemeral_public.to_encoded_point(false).as_bytes()
        );

        Ok(ephemeral_public)
    }

    /// Create attestation document
    pub fn create_attestation_doc(
        &self,
        manifest_hash: &[u8],
        pcr_values: (Vec<u8>, Vec<u8>, Vec<u8>, Vec<u8>),
    ) -> Result<AttestationDoc> {
        info!("📄 Creating attestation document");

        let ephemeral_public = self
            .ephemeral_key
            .as_ref()
            .ok_or_else(|| anyhow!("No ephemeral key available"))?
            .public_key();

        let timestamp_ms = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|e| anyhow!("Failed to get timestamp: {}", e))?
            .as_millis() as u64;

        let attestation_doc = AttestationDoc {
            timestamp_ms,
            user_data: manifest_hash.to_vec(),
            public_key: Some(ephemeral_public.to_encoded_point(false).as_bytes().to_vec()),
            pcr0: pcr_values.0,
            pcr1: pcr_values.1,
            pcr2: pcr_values.2,
            pcr3: pcr_values.3,
        };

        info!("✅ Attestation document created successfully");
        debug!("📄 Timestamp: {}", attestation_doc.timestamp_ms);
        debug!("📄 User data length: {}", attestation_doc.user_data.len());
        debug!(
            "📄 Public key length: {:?}",
            attestation_doc.public_key.as_ref().map(|k| k.len())
        );

        Ok(attestation_doc)
    }

    /// Get ephemeral private key for decryption
    pub fn get_ephemeral_private_key(&self) -> Result<&SecretKey> {
        self.ephemeral_key
            .as_ref()
            .ok_or_else(|| anyhow!("No ephemeral key available"))
    }

    /// Check if ephemeral key exists
    pub fn has_ephemeral_key(&self) -> bool {
        self.ephemeral_key.is_some()
    }

    /// Verify attestation document
    pub fn verify_attestation_doc(
        &self,
        attestation_doc: &AttestationDoc,
        expected_manifest_hash: &[u8],
        expected_pcr_values: (Vec<u8>, Vec<u8>, Vec<u8>, Vec<u8>),
    ) -> Result<bool> {
        info!("🔍 Verifying attestation document");

        // Check timestamp (should be recent, within 5 minutes)
        let current_time = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|e| anyhow!("Failed to get current timestamp: {}", e))?
            .as_millis() as u64;

        let time_diff = current_time.saturating_sub(attestation_doc.timestamp_ms);
        if time_diff > 300_000 {
            // 5 minutes in milliseconds
            warn!("⚠️ Attestation document is too old: {}ms", time_diff);
            return Ok(false);
        }

        // Check manifest hash
        if attestation_doc.user_data != expected_manifest_hash {
            warn!("⚠️ Manifest hash mismatch");
            return Ok(false);
        }

        // Check PCR values
        if attestation_doc.pcr0 != expected_pcr_values.0
            || attestation_doc.pcr1 != expected_pcr_values.1
            || attestation_doc.pcr2 != expected_pcr_values.2
            || attestation_doc.pcr3 != expected_pcr_values.3
        {
            warn!("⚠️ PCR values mismatch");
            return Ok(false);
        }

        // Check public key is present
        if attestation_doc.public_key.is_none() {
            warn!("⚠️ No ephemeral public key in attestation document");
            return Ok(false);
        }

        info!("✅ Attestation document verification successful");
        Ok(true)
    }

    /// Extract ephemeral public key from attestation document
    pub fn extract_ephemeral_public_key(attestation_doc: &AttestationDoc) -> Result<PublicKey> {
        let public_key_bytes = attestation_doc
            .public_key
            .as_ref()
            .ok_or_else(|| anyhow!("No ephemeral public key in attestation document"))?;

        let public_key = PublicKey::from_sec1_bytes(public_key_bytes)
            .map_err(|e| anyhow!("Failed to parse ephemeral public key: {}", e))?;

        Ok(public_key)
    }
}

impl Default for AttestationManager {
    fn default() -> Self {
        Self::new()
    }
}
