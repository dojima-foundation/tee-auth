use serde::{Deserialize, Serialize};
use uuid::Uuid;

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
    GetInfo,
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
    Info {
        version: String,
        enclave_id: String,
        capabilities: Vec<String>,
    },
    Error {
        message: String,
        code: u32,
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
pub struct ErrorResponse {
    pub error: String,
    pub code: u32,
    pub request_id: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InfoResponse {
    pub version: String,
    pub service: String,
    pub enclave_id: String,
    pub capabilities: Vec<String>,
    pub network_status: String,
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
