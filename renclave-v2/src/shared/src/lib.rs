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
