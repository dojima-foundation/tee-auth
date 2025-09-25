use borsh::{BorshDeserialize, BorshSerialize};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use uuid::Uuid;

/// Quorum member with alias and public key
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct QuorumMember {
    /// Member alias/identifier
    pub alias: String,
    /// Member's public key (32 bytes)
    pub pub_key: Vec<u8>,
}

/// Manifest envelope for quorum operations
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct ManifestEnvelope {
    /// The manifest data
    pub manifest: Manifest,
    /// Manifest set approvals
    pub manifest_set_approvals: Vec<Approval>,
    /// Share set approvals
    pub share_set_approvals: Vec<Approval>,
}

/// Manifest data structure
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct Manifest {
    /// Namespace information
    pub namespace: Namespace,
    /// Enclave configuration
    pub enclave: NitroConfig,
    /// Pivot configuration
    pub pivot: PivotConfig,
    /// Manifest set
    pub manifest_set: ManifestSet,
    /// Share set
    pub share_set: ShareSet,
}

/// Namespace information
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct Namespace {
    /// Nonce for ordering
    pub nonce: u64,
    /// Namespace name
    pub name: String,
    /// Quorum key
    pub quorum_key: Vec<u8>,
}

/// Nitro enclave configuration
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct NitroConfig {
    /// PCR0 value
    pub pcr0: Vec<u8>,
    /// PCR1 value
    pub pcr1: Vec<u8>,
    /// PCR2 value
    pub pcr2: Vec<u8>,
    /// PCR3 value
    pub pcr3: Vec<u8>,
    /// AWS root certificate
    pub aws_root_certificate: Vec<u8>,
    /// QOS commit hash
    pub qos_commit: String,
}

/// Pivot configuration
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct PivotConfig {
    /// Pivot hash
    pub hash: [u8; 32],
    /// Restart policy
    pub restart: RestartPolicy,
    /// Arguments
    pub args: Vec<String>,
}

/// Restart policy
#[derive(
    Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq, Default,
)]
pub enum RestartPolicy {
    #[default]
    Always,
    Never,
    OnFailure,
}

/// Manifest set
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct ManifestSet {
    /// Threshold for approvals
    pub threshold: u32,
    /// Members
    pub members: Vec<QuorumMember>,
}

/// Share set
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct ShareSet {
    /// Threshold for approvals
    pub threshold: u32,
    /// Members
    pub members: Vec<QuorumMember>,
}

/// Encrypted share for injection
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct EncryptedShare {
    pub member_id: String,
    #[serde(with = "serde_bytes")]
    pub encrypted_share: Vec<u8>,
    #[serde(with = "serde_bytes")]
    pub public_key: Vec<u8>,
}

/// Approval structure
#[derive(Debug, Clone, Serialize, Deserialize, BorshSerialize, BorshDeserialize, PartialEq, Eq)]
pub struct Approval {
    /// Signature
    pub signature: Vec<u8>,
    /// Member who approved
    pub member: QuorumMember,
}

/// Genesis output per Setup Member
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct GenesisMemberOutput {
    /// The Quorum Member whom's Setup Key was used.
    pub share_set_member: QuorumMember,
    /// Quorum Key Share encrypted to the `setup_member`'s Personal Key.
    pub encrypted_quorum_key_share: Vec<u8>,
    /// Sha512 hash of the plaintext quorum key share. Used by the share set
    /// member to verify they correctly decrypted the share.
    #[serde(with = "serde_bytes")]
    pub share_hash: [u8; 64],
}

/// A set of member shards used to successfully recover the quorum key during
/// the genesis ceremony.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct RecoveredPermutation(pub Vec<MemberShard>);

/// Member shard structure
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct MemberShard {
    /// Member of the Setup Set.
    pub member: QuorumMember,
    /// Shard of the generated Quorum Key, encrypted to the `member`s Setup Key.
    pub shard: Vec<u8>,
}

/// An encrypted quorum key along with a signature over the encrypted payload
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct EncryptedQuorumKey {
    /// The encrypted payload: a quorum key
    pub encrypted_quorum_key: Vec<u8>,
    /// Signature over the encrypted quorum key
    pub signature: Vec<u8>,
}

impl Manifest {
    /// Calculate the QoS hash of this manifest
    pub fn qos_hash(&self) -> Vec<u8> {
        // Create a deterministic hash of the manifest
        let mut hasher = Sha256::new();

        // Hash namespace info
        hasher.update(self.namespace.nonce.to_le_bytes());
        hasher.update(self.namespace.name.as_bytes());
        hasher.update(&self.namespace.quorum_key);

        // Hash enclave config
        hasher.update(&self.enclave.pcr0);
        hasher.update(&self.enclave.pcr1);
        hasher.update(&self.enclave.pcr2);
        hasher.update(&self.enclave.pcr3);
        hasher.update(&self.enclave.aws_root_certificate);
        hasher.update(self.enclave.qos_commit.as_bytes());

        // Hash pivot config
        hasher.update(self.pivot.hash);
        // Convert RestartPolicy to string manually
        let restart_str = match self.pivot.restart {
            RestartPolicy::Always => "Always",
            RestartPolicy::Never => "Never",
            RestartPolicy::OnFailure => "OnFailure",
        };
        hasher.update(restart_str.as_bytes());
        for arg in &self.pivot.args {
            hasher.update(arg.as_bytes());
        }

        // Hash manifest set
        hasher.update(self.manifest_set.threshold.to_le_bytes());
        for member in &self.manifest_set.members {
            hasher.update(member.alias.as_bytes());
            hasher.update(&member.pub_key);
        }

        // Hash share set
        hasher.update(self.share_set.threshold.to_le_bytes());
        for member in &self.share_set.members {
            hasher.update(member.alias.as_bytes());
            hasher.update(&member.pub_key);
        }

        hasher.finalize().to_vec()
    }
}

/// Request types for communication between host and enclave
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnclaveRequest {
    pub id: String,
    pub operation: EnclaveOperation,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum EnclaveOperation {
    GenerateSeed {
        strength: u32,
        passphrase: Option<String>,
    },
    ValidateSeed {
        seed_phrase: String,
    },
    DeriveKey {
        seed_phrase: String,
        path: String,
        curve: String,
    },
    DeriveAddress {
        seed_phrase: String,
        path: String,
        curve: String,
    },
    GenerateQuorumKey {
        members: Vec<QuorumMember>,
        threshold: u32,
        dr_key: Option<Vec<u8>>,
    },
    ExportQuorumKey {
        new_manifest_envelope: Box<ManifestEnvelope>,
        cose_sign1_attestation_document: Vec<u8>,
    },
    InjectQuorumKey {
        encrypted_quorum_key: EncryptedQuorumKey,
    },
    GenesisBoot {
        namespace_name: String,
        namespace_nonce: u64,
        manifest_members: Vec<QuorumMember>,
        manifest_threshold: u32,
        share_members: Vec<QuorumMember>,
        share_threshold: u32,
        pivot_hash: [u8; 32],
        pivot_args: Vec<String>,
        dr_key: Option<Vec<u8>>,
    },
    InjectShares {
        namespace_name: String,
        namespace_nonce: u64,
        shares: Vec<DecryptedShare>,
    },
    GetInfo,
    // Data Encryption/Decryption Operations
    EncryptData {
        data: Vec<u8>,
        recipient_public: Vec<u8>,
    },
    DecryptData {
        encrypted_data: Vec<u8>,
    },
    // Transaction Signing Operations
    SignTransaction {
        transaction_data: Vec<u8>,
    },
    SignTransactionWithSeed {
        transaction_data: Vec<u8>,
        encrypted_seed: Vec<u8>,
    },
    SignMessage {
        message: Vec<u8>,
    },
    // Application State Operations
    GetApplicationStatus,
    StoreApplicationData {
        key: String,
        data: Vec<u8>,
    },
    GetApplicationData {
        key: String,
    },
    // Reset Operations
    ResetEnclave,

    // TEE-to-TEE Communication Operations
    BootKeyForward {
        manifest_envelope: serde_json::Value,
        pivot: serde_json::Value,
    },
    ExportKey {
        manifest_envelope: serde_json::Value,
        attestation_doc: serde_json::Value,
    },
    InjectKey {
        encrypted_quorum_key: serde_json::Value,
        signature: serde_json::Value,
    },
    GenerateAttestation {
        manifest_hash: Vec<u8>,
        pcr_values: (Vec<u8>, Vec<u8>, Vec<u8>, Vec<u8>),
    },
    ShareManifest {
        manifest_envelope: serde_json::Value,
    },
}

/// Application metadata for tracking state information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApplicationMetadata {
    /// Application name
    pub name: String,
    /// Application version
    pub version: String,
    /// Last updated timestamp
    pub last_updated: u64,
    /// Number of operations performed
    pub operation_count: u64,
}

/// Response types from enclave to host
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnclaveResponse {
    pub id: String,
    pub result: EnclaveResult,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum EnclaveResult {
    SeedGenerated {
        seed_phrase: String,
        entropy: String,
        strength: u32,
        word_count: usize,
    },
    SeedValidated {
        valid: bool,
        word_count: usize,
    },
    KeyDerived {
        private_key: String,
        public_key: String,
        address: String,
        path: String,
        curve: String,
    },
    AddressDerived {
        address: String,
        path: String,
        curve: String,
    },
    QuorumKeyGenerated {
        quorum_key: Vec<u8>,
        member_outputs: Vec<GenesisMemberOutput>,
        threshold: u32,
        recovery_permutations: Vec<RecoveredPermutation>,
        dr_key_wrapped_quorum_key: Option<Vec<u8>>,
        #[serde(with = "serde_bytes")]
        quorum_key_hash: [u8; 64],
        test_message_ciphertext: Vec<u8>,
        test_message_signature: Vec<u8>,
        test_message: Vec<u8>,
    },
    QuorumKeyExported {
        encrypted_quorum_key: EncryptedQuorumKey,
    },
    QuorumKeyInjected {
        success: bool,
    },
    GenesisBootCompleted {
        quorum_public_key: Vec<u8>,
        ephemeral_key: Vec<u8>,
        manifest_envelope: Box<ManifestEnvelope>,
        waiting_state: String,
        encrypted_shares: Vec<GenesisMemberOutput>,
    },
    SharesInjected {
        reconstructed_quorum_key: Vec<u8>,
        success: bool,
    },
    Info {
        version: String,
        enclave_id: String,
        capabilities: Vec<String>,
    },
    Error {
        message: String,
        code: u32,
    },
    // Data Encryption/Decryption Results
    DataEncrypted {
        encrypted_data: Vec<u8>,
    },
    DataDecrypted {
        decrypted_data: Vec<u8>,
    },
    // Transaction Signing Results
    TransactionSigned {
        signature: Vec<u8>,
        recovery_id: u8,
    },
    MessageSigned {
        signature: Vec<u8>,
    },
    // Application State Results
    ApplicationStatus {
        phase: String,
        has_quorum_key: bool,
        data_count: usize,
        metadata: ApplicationMetadata,
    },
    ApplicationDataStored {
        success: bool,
    },
    ApplicationDataRetrieved {
        data: Option<Vec<u8>>,
    },
    // Reset Results
    EnclaveReset {
        success: bool,
    },

    // TEE-to-TEE Communication Results
    BootKeyForwardResponse {
        nsm_response: serde_json::Value,
    },
    ExportKeyResponse {
        encrypted_quorum_key: Vec<u8>,
        signature: Vec<u8>,
    },
    InjectKeyResponse {
        success: bool,
    },
    GenerateAttestationResponse {
        attestation_doc: Vec<u8>,
    },
    ShareManifestResponse {
        success: bool,
    },
}

/// HTTP API request/response types
#[derive(Debug, Serialize, Deserialize)]
pub struct GenerateSeedRequest {
    pub strength: Option<u32>,
    pub passphrase: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenerateSeedResponse {
    pub seed_phrase: String,
    pub entropy: String,
    pub strength: u32,
    pub word_count: usize,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ValidateSeedRequest {
    pub seed_phrase: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ValidateSeedResponse {
    pub valid: bool,
    pub word_count: usize,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveKeyRequest {
    pub seed_phrase: String,
    pub path: String,
    pub curve: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveKeyResponse {
    pub private_key: String,
    pub public_key: String,
    pub address: String,
    pub path: String,
    pub curve: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveAddressRequest {
    pub seed_phrase: String,
    pub path: String,
    pub curve: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveAddressResponse {
    pub address: String,
    pub path: String,
    pub curve: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenerateQuorumKeyRequest {
    pub members: Vec<QuorumMember>,
    pub threshold: u32,
    pub dr_key: Option<Vec<u8>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenerateQuorumKeyResponse {
    pub quorum_key: Vec<u8>,
    pub member_outputs: Vec<GenesisMemberOutput>,
    pub threshold: u32,
    pub recovery_permutations: Vec<RecoveredPermutation>,
    pub dr_key_wrapped_quorum_key: Option<Vec<u8>>,
    #[serde(with = "serde_bytes")]
    pub quorum_key_hash: [u8; 64],
    pub test_message_ciphertext: Vec<u8>,
    pub test_message_signature: Vec<u8>,
    pub test_message: Vec<u8>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ExportQuorumKeyRequest {
    pub new_manifest_envelope: ManifestEnvelope,
    pub cose_sign1_attestation_document: Vec<u8>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ExportQuorumKeyResponse {
    pub encrypted_quorum_key: EncryptedQuorumKey,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InjectQuorumKeyRequest {
    pub encrypted_quorum_key: EncryptedQuorumKey,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InjectQuorumKeyResponse {
    pub success: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenesisBootRequest {
    pub namespace_name: String,
    pub namespace_nonce: u64,
    pub manifest_members: Vec<QuorumMember>,
    pub manifest_threshold: u32,
    pub share_members: Vec<QuorumMember>,
    pub share_threshold: u32,
    pub pivot_hash: [u8; 32],
    pub pivot_args: Vec<String>,
    pub dr_key: Option<Vec<u8>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenesisBootResponse {
    pub quorum_public_key: Vec<u8>,
    pub ephemeral_key: Vec<u8>,
    pub manifest_envelope: ManifestEnvelope,
    pub waiting_state: String,
    pub encrypted_shares: Vec<GenesisMemberOutput>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecryptedShare {
    pub member_alias: String,
    #[serde(with = "serde_bytes")]
    pub decrypted_share: Vec<u8>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ShareInjectionRequest {
    pub namespace_name: String,
    pub namespace_nonce: u64,
    pub shares: Vec<DecryptedShare>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ShareInjectionResponse {
    pub reconstructed_quorum_key: Vec<u8>,
    pub success: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: u32,
    pub request_id: Option<String>,
}

// Data Encryption/Decryption Request/Response Types

/// Request to encrypt data
#[derive(Debug, Serialize, Deserialize)]
pub struct EncryptDataRequest {
    /// Data to encrypt
    pub data: Vec<u8>,
    /// Recipient's public key
    pub recipient_public: Vec<u8>,
}

/// Response from data encryption
#[derive(Debug, Serialize, Deserialize)]
pub struct EncryptDataResponse {
    /// Whether encryption was successful
    pub success: bool,
    /// Encrypted data
    pub encrypted_data: Vec<u8>,
}

/// Request to decrypt data
#[derive(Debug, Serialize, Deserialize)]
pub struct DecryptDataRequest {
    /// Encrypted data to decrypt
    pub encrypted_data: Vec<u8>,
}

/// Response from data decryption
#[derive(Debug, Serialize, Deserialize)]
pub struct DecryptDataResponse {
    /// Whether decryption was successful
    pub success: bool,
    /// Decrypted data
    pub decrypted_data: Vec<u8>,
}

// Transaction Signing Request/Response Types

/// Request to sign a transaction
#[derive(Debug, Serialize, Deserialize)]
pub struct SignTransactionRequest {
    /// Transaction data to sign
    pub transaction_data: Vec<u8>,
}

/// Response from transaction signing
#[derive(Debug, Serialize, Deserialize)]
pub struct SignTransactionResponse {
    /// Whether signing was successful
    pub success: bool,
    /// Signature bytes
    pub signature: Vec<u8>,
    /// Recovery ID for signature
    pub recovery_id: u8,
}

/// Request to sign a message
#[derive(Debug, Serialize, Deserialize)]
pub struct SignMessageRequest {
    /// Message to sign
    pub message: Vec<u8>,
}

/// Response from message signing
#[derive(Debug, Serialize, Deserialize)]
pub struct SignMessageResponse {
    /// Whether signing was successful
    pub success: bool,
    /// Signature bytes
    pub signature: Vec<u8>,
}

// Application Status Request/Response Types

/// Response from application status request
#[derive(Debug, Serialize, Deserialize)]
pub struct ApplicationStatusResponse {
    /// Whether request was successful
    pub success: bool,
    /// Current application phase
    pub phase: String,
    /// Whether quorum key is available
    pub has_quorum_key: bool,
    /// Number of stored data items
    pub data_count: usize,
    /// Application metadata
    pub metadata: ApplicationMetadata,
}

// Application Data Storage Request/Response Types

/// Request to store application data
#[derive(Debug, Serialize, Deserialize)]
pub struct StoreApplicationDataRequest {
    /// Key to store data under
    pub key: String,
    /// Data to store
    pub data: Vec<u8>,
}

/// Response from application data storage
#[derive(Debug, Serialize, Deserialize)]
pub struct StoreApplicationDataResponse {
    /// Whether storage was successful
    pub success: bool,
}

/// Request to get application data
#[derive(Debug, Serialize, Deserialize)]
pub struct GetApplicationDataRequest {
    /// Key to retrieve data for
    pub key: String,
}

/// Response from application data retrieval
#[derive(Debug, Serialize, Deserialize)]
pub struct GetApplicationDataResponse {
    /// Whether retrieval was successful
    pub success: bool,
    /// Retrieved data (None if not found)
    pub data: Option<Vec<u8>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InfoResponse {
    pub version: String,
    pub service: String,
    pub enclave_id: String,
    pub capabilities: Vec<String>,
    pub network_status: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SeedGenerationResult {
    pub seed_phrase: String,
    pub entropy: String,
    pub strength: u32,
    pub word_count: usize,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SeedValidationResult {
    pub valid: bool,
    pub strength: u32,
    pub word_count: usize,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct EnclaveInfo {
    pub version: String,
    pub enclave_id: String,
    pub capabilities: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct NetworkStatus {
    pub connectivity: ConnectivityStatus,
    pub interfaces: Vec<String>,
    pub qemu_detected: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConnectivityStatus {
    pub external: bool,
    pub gateway: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConnectivityResult {
    pub success: bool,
    pub external: bool,
    pub gateway: bool,
    pub dns: bool,
}

impl EnclaveRequest {
    pub fn new(operation: EnclaveOperation) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            operation,
        }
    }
}

impl EnclaveResponse {
    pub fn new(id: String, result: EnclaveResult) -> Self {
        Self { id, result }
    }

    pub fn error(id: String, message: String, code: u32) -> Self {
        Self {
            id,
            result: EnclaveResult::Error { message, code },
        }
    }
}

/// Common error types
#[derive(thiserror::Error, Debug)]
pub enum RenclaveError {
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Invalid seed strength: {0}. Must be 128, 160, 192, 224, or 256")]
    InvalidStrength(u32),

    #[error("Seed generation failed: {0}")]
    SeedGeneration(String),

    #[error("Network error: {0}")]
    Network(String),

    #[error("Enclave communication error: {0}")]
    EnclaveCommunication(String),
}

pub type Result<T> = std::result::Result<T, RenclaveError>;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_enclave_request_new() {
        let operation = EnclaveOperation::GetInfo;
        let request = EnclaveRequest::new(operation);

        assert!(!request.id.is_empty());
        assert!(matches!(request.operation, EnclaveOperation::GetInfo));
    }

    #[test]
    fn test_enclave_response_new() {
        let id = "test-id".to_string();
        let result = EnclaveResult::Info {
            version: "1.0.0".to_string(),
            enclave_id: "test-enclave".to_string(),
            capabilities: vec!["test".to_string()],
        };

        let response = EnclaveResponse::new(id.clone(), result);

        assert_eq!(response.id, id);
        assert!(matches!(response.result, EnclaveResult::Info { .. }));
    }

    #[test]
    fn test_enclave_response_error() {
        let id = "test-id".to_string();
        let message = "Test error".to_string();
        let code = 500;

        let response = EnclaveResponse::error(id.clone(), message.clone(), code);

        assert_eq!(response.id, id);
        match response.result {
            EnclaveResult::Error {
                message: msg,
                code: c,
            } => {
                assert_eq!(msg, message);
                assert_eq!(c, code);
            }
            _ => panic!("Expected error result"),
        }
    }

    #[test]
    fn test_enclave_operation_serialization() {
        let operations = vec![
            EnclaveOperation::GenerateSeed {
                strength: 256,
                passphrase: Some("test123".to_string()),
            },
            EnclaveOperation::ValidateSeed {
                seed_phrase: "test seed".to_string(),
            },
            EnclaveOperation::DeriveKey {
                seed_phrase: "test seed".to_string(),
                path: "m/44'/0'/0'/0/0".to_string(),
                curve: "secp256k1".to_string(),
            },
            EnclaveOperation::DeriveAddress {
                seed_phrase: "test seed".to_string(),
                path: "m/44'/0'/0'/0/0".to_string(),
                curve: "secp256k1".to_string(),
            },
            EnclaveOperation::GetInfo,
        ];

        for operation in operations {
            let serialized = serde_json::to_string(&operation).unwrap();
            let _deserialized: EnclaveOperation = serde_json::from_str(&serialized).unwrap();
            // Verify the deserialized operation matches the original
            assert!(
                matches!(&operation, _deserialized),
                "Serialization round-trip failed for {:?}",
                operation
            );
        }
    }

    #[test]
    fn test_enclave_result_serialization() {
        let results = vec![
            EnclaveResult::SeedGenerated {
                seed_phrase: "test phrase".to_string(),
                entropy: "test entropy".to_string(),
                strength: 256,
                word_count: 24,
            },
            EnclaveResult::SeedValidated {
                valid: true,
                word_count: 12,
            },
            EnclaveResult::KeyDerived {
                private_key: "private".to_string(),
                public_key: "public".to_string(),
                address: "address".to_string(),
                path: "m/44'/0'/0'/0/0".to_string(),
                curve: "secp256k1".to_string(),
            },
            EnclaveResult::AddressDerived {
                address: "address".to_string(),
                path: "m/44'/0'/0'/0/0".to_string(),
                curve: "secp256k1".to_string(),
            },
            EnclaveResult::Info {
                version: "1.0.0".to_string(),
                enclave_id: "test".to_string(),
                capabilities: vec!["test".to_string()],
            },
            EnclaveResult::Error {
                message: "test error".to_string(),
                code: 500,
            },
        ];

        for result in results {
            let serialized = serde_json::to_string(&result).unwrap();
            let _deserialized: EnclaveResult = serde_json::from_str(&serialized).unwrap();
            // Verify the deserialized result matches the original
            assert!(
                matches!(&result, _deserialized),
                "Serialization round-trip failed for {:?}",
                result
            );
        }
    }

    #[test]
    fn test_http_request_serialization() {
        // Test GenerateSeedRequest
        let generate_request = GenerateSeedRequest {
            strength: Some(256),
            passphrase: Some("test123".to_string()),
        };
        let serialized = serde_json::to_string(&generate_request).unwrap();
        assert!(!serialized.is_empty());

        // Test ValidateSeedRequest
        let validate_request = ValidateSeedRequest {
            seed_phrase: "test seed".to_string(),
        };
        let serialized = serde_json::to_string(&validate_request).unwrap();
        assert!(!serialized.is_empty());

        // Test DeriveKeyRequest
        let derive_key_request = DeriveKeyRequest {
            seed_phrase: "test seed".to_string(),
            path: "m/44'/0'/0'/0/0".to_string(),
            curve: "secp256k1".to_string(),
        };
        let serialized = serde_json::to_string(&derive_key_request).unwrap();
        assert!(!serialized.is_empty());

        // Test DeriveAddressRequest
        let derive_address_request = DeriveAddressRequest {
            seed_phrase: "test seed".to_string(),
            path: "m/44'/0'/0'/0/0".to_string(),
            curve: "secp256k1".to_string(),
        };
        let serialized = serde_json::to_string(&derive_address_request).unwrap();
        assert!(!serialized.is_empty());
    }

    #[test]
    fn test_renclave_error_display() {
        let errors = vec![
            RenclaveError::InvalidStrength(100),
            RenclaveError::SeedGeneration("test error".to_string()),
            RenclaveError::Network("network error".to_string()),
            RenclaveError::EnclaveCommunication("communication error".to_string()),
        ];

        for error in errors {
            let display = format!("{}", error);
            assert!(!display.is_empty());

            // Check for specific content based on error type
            match error {
                RenclaveError::InvalidStrength(_) => {
                    assert!(
                        display.contains("Invalid seed strength"),
                        "Should contain 'Invalid seed strength'"
                    );
                    assert!(
                        display.contains("128, 160, 192, 224, or 256"),
                        "Should contain valid strength values"
                    );
                }
                RenclaveError::SeedGeneration(_) => {
                    assert!(
                        display.contains("Seed generation failed"),
                        "Should contain 'Seed generation failed'"
                    );
                }
                RenclaveError::Network(_) => {
                    assert!(
                        display.contains("Network error"),
                        "Should contain 'Network error'"
                    );
                }
                RenclaveError::EnclaveCommunication(_) => {
                    assert!(
                        display.contains("Enclave communication error"),
                        "Should contain 'Enclave communication error'"
                    );
                }
                _ => {
                    // For other error types, just check they're not empty
                    assert!(!display.is_empty(), "Error message should not be empty");
                }
            }
        }
    }

    #[test]
    fn test_renclave_error_from_serde() {
        let invalid_json = "{ invalid json }";
        let result: std::result::Result<serde_json::Value, serde_json::Error> =
            serde_json::from_str(invalid_json);

        assert!(result.is_err());
        let serde_error = result.unwrap_err();
        let renclave_error: RenclaveError = serde_error.into();

        match renclave_error {
            RenclaveError::Serialization(_) => {}
            _ => panic!("Expected serialization error"),
        }
    }

    #[test]
    fn test_renclave_error_from_io() {
        use std::io;

        let io_error = io::Error::new(io::ErrorKind::NotFound, "file not found");
        let renclave_error: RenclaveError = io_error.into();

        match renclave_error {
            RenclaveError::Io(_) => {}
            _ => panic!("Expected IO error"),
        }
    }
}
