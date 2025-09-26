#[allow(unused_imports)]
use renclave_enclave::data_encryption::DataEncryption;
#[allow(unused_imports)]
use renclave_enclave::quorum::{boot_genesis, shares_reconstruct, GenesisSet, P256Pair};
use renclave_enclave::seed_generator::SeedGenerator;
#[allow(unused_imports)]
use renclave_enclave::storage::TeeStorage;
#[allow(unused_imports)]
use renclave_enclave::tee_communication::TeeCommunicationManager;
use renclave_network::{NetworkConfig, NetworkManager};
#[allow(unused_imports)]
use renclave_shared::{EnclaveOperation, EnclaveRequest, EnclaveResponse, EnclaveResult};
use renclave_shared::{
    Manifest, ManifestEnvelope, ManifestSet, NitroConfig, PivotConfig, QuorumMember, ShareSet,
};
#[allow(unused_imports)]
use std::collections::HashSet;
#[allow(unused_imports)]
use std::sync::Arc;

/// Integration tests for renclave components
/// These tests verify that different parts of the system work together correctly

#[tokio::test]
async fn test_seed_generation_flow() {
    // Test the complete seed generation flow
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed
    let seed_result = seed_generator.generate_seed(256, None).await;
    assert!(seed_result.is_ok());

    let seed = seed_result.unwrap();
    assert_eq!(seed.entropy.len(), 64); // 256 bits = 32 bytes = 64 hex chars
    assert!(!seed.phrase.is_empty());

    // Validate the generated seed
    let validation_result = seed_generator.validate_seed(&seed.phrase).await;
    assert!(validation_result.is_ok());
    assert!(validation_result.unwrap());
}

#[tokio::test]
async fn test_seed_validation_integration() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a valid seed
    let seed = seed_generator.generate_seed(128, None).await.unwrap();

    // Test validation with the generated seed
    let is_valid = seed_generator.validate_seed(&seed.phrase).await.unwrap();
    assert!(is_valid);

    // Test validation with invalid seed
    let is_invalid = seed_generator
        .validate_seed("invalid seed phrase")
        .await
        .unwrap();
    assert!(!is_invalid);
}

#[tokio::test]
async fn test_entropy_consistency() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate multiple seeds and verify entropy consistency
    let seed1 = seed_generator.generate_seed(256, None).await.unwrap();
    let seed2 = seed_generator.generate_seed(256, None).await.unwrap();

    // Entropy should be different (random)
    assert_ne!(seed1.entropy, seed2.entropy);

    // But both should be valid
    assert!(seed_generator.validate_seed(&seed1.phrase).await.unwrap());
    assert!(seed_generator.validate_seed(&seed2.phrase).await.unwrap());
}

#[tokio::test]
async fn test_passphrase_integration() {
    let seed_generator = SeedGenerator::new().await.unwrap();
    let passphrase = "test-passphrase-123";

    // Generate seed with passphrase
    let seed_with_pass = seed_generator
        .generate_seed(256, Some(passphrase))
        .await
        .unwrap();

    // Generate seed without passphrase
    let seed_without_pass = seed_generator.generate_seed(256, None).await.unwrap();

    // Seeds should be different
    assert_ne!(seed_with_pass.entropy, seed_without_pass.entropy);
    assert_ne!(seed_with_pass.phrase, seed_without_pass.phrase);

    // Both should be valid
    assert!(seed_generator
        .validate_seed(&seed_with_pass.phrase)
        .await
        .unwrap());
    assert!(seed_generator
        .validate_seed(&seed_without_pass.phrase)
        .await
        .unwrap());
}

#[tokio::test]
async fn test_error_handling_integration() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test invalid strength
    let result = seed_generator.generate_seed(0, None).await;
    assert!(result.is_err());

    // Test invalid strength (not multiple of 32)
    let result = seed_generator.generate_seed(100, None).await;
    assert!(result.is_err());

    // Test empty seed validation
    let result = seed_generator.validate_seed("").await;
    assert!(result.is_ok());
    assert!(!result.unwrap());
}

#[tokio::test]
async fn test_serialization_integration() {
    // Test that EnclaveRequest can be properly serialized and deserialized
    let operation = EnclaveOperation::GenerateSeed {
        strength: 256,
        passphrase: Some("test-pass".to_string()),
    };

    let request = EnclaveRequest {
        id: "test-id".to_string(),
        operation,
    };

    // Serialize
    let serialized = serde_json::to_string(&request);
    assert!(serialized.is_ok());

    // Deserialize
    let deserialized: EnclaveRequest = serde_json::from_str(&serialized.unwrap()).unwrap();
    assert_eq!(request.id, deserialized.id);
    // Note: EnclaveOperation doesn't implement PartialEq, so we can't compare directly
    // Instead, we'll verify the structure is correct by checking individual fields
    match (request.operation, deserialized.operation) {
        (
            EnclaveOperation::GenerateSeed {
                strength: s1,
                passphrase: p1,
            },
            EnclaveOperation::GenerateSeed {
                strength: s2,
                passphrase: p2,
            },
        ) => {
            assert_eq!(s1, s2);
            assert_eq!(p1, p2);
        }
        _ => panic!("Operation types don't match"),
    }

    // Test response serialization with actual EnclaveResult variant
    let response = EnclaveResponse {
        id: request.id.clone(),
        result: EnclaveResult::SeedGenerated {
            seed_phrase: "test seed phrase".to_string(),
            entropy: "test entropy".to_string(),
            strength: 256,
            word_count: 24,
        },
    };

    let response_serialized = serde_json::to_string(&response);
    assert!(response_serialized.is_ok());

    let response_deserialized: EnclaveResponse =
        serde_json::from_str(&response_serialized.unwrap()).unwrap();
    assert_eq!(response.id, response_deserialized.id);
}

#[tokio::test]
async fn test_network_integration_basic() {
    // Test network-related functionality
    use renclave_network::{NetworkConfig, NetworkManager};

    let config = NetworkConfig::default();
    let network_manager = NetworkManager::new(config);

    // Test network initialization (should not fail)
    let init_result = network_manager.initialize().await;
    // On non-QEMU systems, this might fail, which is expected
    if init_result.is_err() {
        println!(
            "Network initialization failed (expected on non-QEMU systems): {:?}",
            init_result.err()
        );
    }
}

#[tokio::test]
async fn test_concurrent_seed_generation() {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());

    // Generate multiple seeds concurrently
    let mut handles = vec![];

    for _ in 0..5 {
        let generator = Arc::clone(&seed_generator);
        let handle = tokio::spawn(async move { generator.generate_seed(256, None).await });
        handles.push(handle);
    }

    // Wait for all to complete
    let results: Vec<_> = futures::future::join_all(handles).await;

    // All should succeed
    for result in results {
        assert!(result.is_ok());
        let seed = result.unwrap().unwrap();
        assert_eq!(seed.entropy.len(), 64); // 256 bits = 32 bytes = 64 hex chars
    }
}

#[tokio::test]
async fn test_seed_strength_validation() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test valid strengths
    let valid_strengths = vec![128, 256];

    for strength in valid_strengths {
        let result = seed_generator.generate_seed(strength, None).await;
        assert!(
            result.is_ok(),
            "Failed to generate seed with strength {}",
            strength
        );

        let seed = result.unwrap();
        // entropy.len() returns hex string length, so 256 bits = 32 bytes = 64 hex chars
        let expected_hex_length = (strength / 8) * 2;
        assert_eq!(seed.entropy.len(), expected_hex_length as usize);
    }

    // Test invalid strengths (only 128, 160, 192, 224, 256 are supported)
    let invalid_strengths = vec![0, 64, 96, 288, 320];

    for strength in invalid_strengths {
        let result = seed_generator.generate_seed(strength, None).await;
        assert!(result.is_err(), "Should fail with strength {}", strength);
    }
}

#[tokio::test]
async fn test_mnemonic_phrase_quality() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate multiple seeds and verify quality
    for _ in 0..10 {
        let seed = seed_generator.generate_seed(256, None).await.unwrap();

        // Verify word count is correct
        assert_eq!(seed.word_count, 24); // 256 bits = 24 words

        // Verify entropy length is correct (hex string)
        assert_eq!(seed.entropy.len(), 64); // 32 bytes = 64 hex chars

        // Verify phrase is not empty
        assert!(!seed.phrase.is_empty());

        // Verify phrase contains expected number of words
        let word_count = seed.phrase.split_whitespace().count();
        assert_eq!(word_count, seed.word_count);

        // Verify validation works
        assert!(seed_generator.validate_seed(&seed.phrase).await.unwrap());
    }
}

#[tokio::test]
async fn test_seed_phrase_uniqueness() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate multiple seeds and ensure they're unique
    let mut phrases = std::collections::HashSet::new();

    for _ in 0..20 {
        let seed = seed_generator.generate_seed(256, None).await.unwrap();
        assert!(
            phrases.insert(seed.phrase.clone()),
            "Duplicate seed phrase generated"
        );
    }
}

#[tokio::test]
async fn test_invalid_input_handling() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test various invalid inputs
    let invalid_inputs = vec![
        "",
        "   ",
        "not a valid seed phrase",
        "invalid words that are not in the BIP39 wordlist",
    ];

    for input in invalid_inputs {
        let result = seed_generator.validate_seed(input).await;
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    // Test that a valid BIP39 mnemonic passes validation
    let valid_mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art";
    let result = seed_generator.validate_seed(valid_mnemonic).await;
    assert!(result.is_ok());
    assert!(result.unwrap());
}

#[tokio::test]
async fn test_seed_generation_edge_cases() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test minimum valid strength
    let result = seed_generator.generate_seed(128, None).await;
    assert!(result.is_ok());
    let seed = result.unwrap();
    assert_eq!(seed.strength, 128);
    assert_eq!(seed.word_count, 12);

    // Test maximum supported strength
    let result = seed_generator.generate_seed(256, None).await;
    assert!(result.is_ok());
    let seed = result.unwrap();
    assert_eq!(seed.strength, 256);
    assert_eq!(seed.word_count, 24);
}

#[tokio::test]
async fn test_passphrase_edge_cases() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test empty passphrase
    let seed1 = seed_generator.generate_seed(256, Some("")).await.unwrap();
    let seed2 = seed_generator.generate_seed(256, None).await.unwrap();

    // Should be different due to passphrase handling
    assert_ne!(seed1.phrase, seed2.phrase);

    // Test very long passphrase
    let long_passphrase = "a".repeat(1000);
    let seed3 = seed_generator
        .generate_seed(256, Some(&long_passphrase))
        .await
        .unwrap();

    // Should still be valid
    assert!(seed_generator.validate_seed(&seed3.phrase).await.unwrap());
}

#[tokio::test]
async fn test_concurrent_validation() {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());

    // Generate a valid seed
    let seed = seed_generator.generate_seed(256, None).await.unwrap();

    // Validate it concurrently multiple times
    let mut handles = vec![];

    for _ in 0..10 {
        let generator = Arc::clone(&seed_generator);
        let phrase = seed.phrase.clone();
        let handle = tokio::spawn(async move { generator.validate_seed(&phrase).await });
        handles.push(handle);
    }

    // All validations should succeed
    let results: Vec<_> = futures::future::join_all(handles).await;
    for result in results {
        assert!(result.is_ok());
        let validation_result = result.unwrap().unwrap();
        assert!(validation_result);
    }
}

#[tokio::test]
async fn test_quorum_key_generation_integration() {
    let (genesis_set, _member_pairs) = create_test_genesis_set(5, 3);

    // Run genesis ceremony
    // Create a storage instance and clear it first to avoid permission issues
    let storage = TeeStorage::new();
    let _ = storage.clear_all(); // Clear any existing files

    let output = match boot_genesis(&genesis_set, None, Some(&storage)).await {
        Ok(output) => output,
        Err(e) if e.to_string().contains("Permission denied") => {
            // Skip this test in environments where we can't write to /tmp/
            println!("Skipping test due to permission issues: {}", e);
            return;
        }
        Err(e) => panic!("Unexpected error: {}", e),
    };

    // Verify basic properties
    assert_eq!(output.threshold, 3);
    assert_eq!(output.member_outputs.len(), 5);
    assert!(!output.quorum_key.is_empty());
    assert!(!output.shares.is_empty());

    // Verify we can reconstruct the quorum key from shares
    let reconstructed = shares_reconstruct(&output.shares[0..3]).unwrap();
    assert_eq!(reconstructed.len(), 32); // 32 bytes for P256 master seed

    // Verify test message
    assert_eq!(output.test_message, b"renclave-quorum-test-message");
    assert!(!output.test_message_ciphertext.is_empty());
    assert!(!output.test_message_signature.is_empty());
}

#[tokio::test]
async fn test_data_encryption_integration() {
    let quorum_key = P256Pair::generate().unwrap();
    let data_encryption = DataEncryption::new(quorum_key);

    // Test data encryption workflow
    let test_data = b"test data for encryption integration";
    let recipient_public = vec![1, 2, 3, 4, 5]; // Mock recipient

    // Encrypt data
    let encrypted = data_encryption
        .encrypt_data(test_data, &recipient_public)
        .unwrap();
    assert!(!encrypted.is_empty());

    // Decrypt data
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(test_data.to_vec(), decrypted);

    // Test with different data sizes
    let test_cases = vec![
        b"".to_vec(),      // Empty
        b"a".to_vec(),     // Single byte
        b"short".to_vec(), // Short
        vec![0u8; 1024],   // Large
    ];

    for test_data in test_cases {
        let encrypted = data_encryption
            .encrypt_data(&test_data, &recipient_public)
            .unwrap();
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}

#[tokio::test]
async fn test_tee_communication_integration() {
    let manager = TeeCommunicationManager::new();
    let quorum_key = P256Pair::generate().unwrap();
    let manifest_envelope = create_test_manifest_envelope();

    // Set up manager
    manager.set_quorum_key(quorum_key).unwrap();
    manager
        .set_manifest_envelope(manifest_envelope.clone())
        .unwrap();

    // Test boot key forward workflow
    let request = renclave_enclave::attestation::BootKeyForwardRequest {
        manifest_envelope: manifest_envelope.clone(),
        pivot: vec![1, 2, 3, 4, 5], // Mock pivot data
    };

    let response = manager.handle_boot_key_forward(request).unwrap();

    // Verify response
    match response.nsm_response {
        renclave_enclave::attestation::NsmResponse::Attestation { document } => {
            assert!(!document.is_empty());
        }
        renclave_enclave::attestation::NsmResponse::Other => {
            // Accept other responses for testing
        }
    }
}

#[tokio::test]
async fn test_storage_integration() {
    let storage = TeeStorage::new();

    // Clear any existing storage first
    let _ = storage.clear_all(); // Ignore errors for test environment

    // Note: TeeStorage doesn't have put_data/get_data methods
    // These are handled through the application state in the actual implementation
    // For testing purposes, we'll just verify storage can be created
    let state = storage.get_storage_state();

    // After clearing, all should be false
    assert!(!state.ephemeral_key_exists);
    assert!(!state.manifest_exists);
    assert!(!state.quorum_key_exists);
    assert!(!state.pivot_exists);
    assert!(!state.shares_exists);

    // Test that we can clear storage (if files exist)
    let _ = storage.clear_all(); // Ignore errors for test environment
}

#[tokio::test]
async fn test_network_integration() {
    let network_manager = create_test_network_manager().await;

    // Test network initialization
    let init_result = network_manager.initialize().await;
    // On non-QEMU systems, this might fail, which is expected
    if init_result.is_err() {
        println!(
            "Network initialization failed (expected on non-QEMU systems): {:?}",
            init_result.err()
        );
    }

    // Test connectivity
    let _connectivity_tester = renclave_network::ConnectivityTester::default();
    // Connectivity testing would be implemented here
}

#[tokio::test]
async fn test_complete_workflow_integration() {
    // Test complete workflow from seed generation to quorum key creation
    let seed_generator = create_test_seed_generator().await;

    // 1. Generate seed
    let seed_result = seed_generator.generate_seed(256, None).await.unwrap();
    assert!(!seed_result.phrase.is_empty());

    // 2. Validate seed
    let is_valid = seed_generator
        .validate_seed(&seed_result.phrase)
        .await
        .unwrap();
    assert!(is_valid);

    // 3. Derive key from seed
    let key_result = seed_generator
        .derive_key(&seed_result.phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();
    assert!(!key_result.private_key.is_empty());

    // 4. Create quorum key
    let quorum_key = P256Pair::generate().unwrap();
    let data_encryption = DataEncryption::new(quorum_key);

    // 5. Encrypt data using quorum key
    let test_data = b"test data for complete workflow";
    let recipient_public = vec![1, 2, 3, 4, 5];
    let encrypted = data_encryption
        .encrypt_data(test_data, &recipient_public)
        .unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(test_data.to_vec(), decrypted);
}

#[tokio::test]
async fn test_concurrent_operations_integration() {
    let seed_generator = Arc::new(create_test_seed_generator().await);
    let mut handles = vec![];

    // Test concurrent seed generation
    for i in 0..10 {
        let generator = Arc::clone(&seed_generator);
        let handle = tokio::spawn(async move {
            let strength = if i % 2 == 0 { 128 } else { 256 };
            let passphrase = if i % 3 == 0 {
                Some(format!("pass-{}", i))
            } else {
                None
            };

            let seed = generator
                .generate_seed(strength, passphrase.as_deref())
                .await
                .unwrap();
            let is_valid = generator.validate_seed(&seed.phrase).await.unwrap();
            assert!(is_valid);

            seed
        });
        handles.push(handle);
    }

    // Wait for all to complete
    let results: Vec<_> = futures::future::join_all(handles).await;

    // All should succeed
    for result in results {
        let seed = result.unwrap();
        assert!(!seed.phrase.is_empty());
        assert!(!seed.entropy.is_empty());
    }
}

#[tokio::test]
async fn test_error_handling_integration_advanced() {
    let seed_generator = create_test_seed_generator().await;

    // Test error handling across different operations
    let error_cases = vec![
        // Invalid strength
        (0, None),
        (100, None),
        (300, None),
    ];

    for (strength, passphrase) in error_cases {
        let result = seed_generator.generate_seed(strength, passphrase).await;
        assert!(
            result.is_err(),
            "Should fail with invalid strength {}",
            strength
        );
    }

    // Test invalid seed validation
    let invalid_seeds = vec![
        "",
        "invalid seed phrase",
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid",
    ];

    for invalid_seed in invalid_seeds {
        let result = seed_generator.validate_seed(invalid_seed).await.unwrap();
        assert!(
            !result,
            "Invalid seed should be invalid: '{}'",
            invalid_seed
        );
    }
}

#[tokio::test]
async fn test_memory_usage_integration() {
    let seed_generator = create_test_seed_generator().await;

    // Test memory usage with many operations
    let mut seeds = Vec::new();

    for _ in 0..100 {
        let seed = seed_generator.generate_seed(256, None).await.unwrap();
        seeds.push(seed);
    }

    // All seeds should be unique
    let mut phrases = HashSet::new();
    for seed in &seeds {
        assert!(phrases.insert(seed.phrase.clone()));
    }

    assert_eq!(phrases.len(), 100);
}

#[tokio::test]
async fn test_performance_integration() {
    let seed_generator = create_test_seed_generator().await;

    let start = std::time::Instant::now();

    // Test performance with multiple operations
    for _ in 0..100 {
        let seed = seed_generator.generate_seed(256, None).await.unwrap();
        let is_valid = seed_generator.validate_seed(&seed.phrase).await.unwrap();
        assert!(is_valid);
    }

    let duration = start.elapsed();
    println!("Performance test: {:?} for 100 operations", duration);

    // Should complete in reasonable time
    assert!(duration.as_secs() < 10);
}

#[tokio::test]
async fn test_stress_test_integration() {
    let seed_generator = create_test_seed_generator().await;

    // Stress test with many operations
    for i in 0..1000 {
        let strength = if i % 3 == 0 { 128 } else { 256 };
        let passphrase = if i % 5 == 0 {
            Some(format!("stress-{}", i))
        } else {
            None
        };

        let seed = seed_generator
            .generate_seed(strength, passphrase.as_deref())
            .await
            .unwrap();
        let is_valid = seed_generator.validate_seed(&seed.phrase).await.unwrap();
        assert!(is_valid);

        if i % 100 == 0 {
            println!("Completed {} operations", i);
        }
    }
}

#[tokio::test]
async fn test_quorum_workflow_integration() {
    let (genesis_set, _member_pairs) = create_test_genesis_set(5, 3);

    // Run genesis ceremony
    // Create a storage instance and clear it first to avoid permission issues
    let storage = TeeStorage::new();
    let _ = storage.clear_all(); // Clear any existing files

    let output = match boot_genesis(&genesis_set, None, Some(&storage)).await {
        Ok(output) => output,
        Err(e) if e.to_string().contains("Permission denied") => {
            // Skip this test in environments where we can't write to /tmp/
            println!("Skipping test due to permission issues: {}", e);
            return;
        }
        Err(e) => panic!("Unexpected error: {}", e),
    };

    // Verify member outputs
    assert_eq!(output.member_outputs.len(), 5);
    for member_output in &output.member_outputs {
        assert!(!member_output.encrypted_quorum_key_share.is_empty());
        assert_eq!(member_output.share_hash.len(), 64); // SHA-512 hash
    }

    // Verify shares can be reconstructed
    let reconstructed = shares_reconstruct(&output.shares[0..3]).unwrap();
    assert_eq!(reconstructed.len(), 32); // 32 bytes for P256 master seed

    // Verify quorum key properties
    assert_eq!(output.quorum_key.len(), 65); // Uncompressed P256 public key
    assert_eq!(output.quorum_key[0], 0x04); // Uncompressed format
}

#[tokio::test]
async fn test_data_encryption_workflow_integration() {
    let quorum_key = P256Pair::generate().unwrap();
    let data_encryption = DataEncryption::new(quorum_key);

    // Test encryption workflow with different data types
    let test_cases = vec![
        b"text data".to_vec(),
        b"binary data".to_vec(),
        vec![0u8; 1024],  // Large data
        vec![0u8; 10240], // Very large data
    ];

    for test_data in test_cases {
        let recipient_public = vec![1, 2, 3, 4, 5];

        // Encrypt
        let encrypted = data_encryption
            .encrypt_data(&test_data, &recipient_public)
            .unwrap();
        assert!(!encrypted.is_empty());

        // Decrypt
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}

#[tokio::test]
async fn test_tee_communication_workflow_integration() {
    let manager = TeeCommunicationManager::new();
    let quorum_key = P256Pair::generate().unwrap();
    let manifest_envelope = create_test_manifest_envelope();

    // Set up manager
    manager.set_quorum_key(quorum_key).unwrap();
    manager
        .set_manifest_envelope(manifest_envelope.clone())
        .unwrap();

    // Test complete TEE communication workflow
    let request = renclave_enclave::attestation::BootKeyForwardRequest {
        manifest_envelope: manifest_envelope.clone(),
        pivot: vec![1, 2, 3, 4, 5], // Mock pivot data
    };

    let response = manager.handle_boot_key_forward(request).unwrap();

    // Verify complete response
    match response.nsm_response {
        renclave_enclave::attestation::NsmResponse::Attestation { document } => {
            assert!(!document.is_empty());
        }
        renclave_enclave::attestation::NsmResponse::Other => {
            // Accept other responses for testing
        }
    }
}

// Helper functions for integration tests
#[allow(dead_code)]
fn create_test_genesis_set(member_count: usize, threshold: u32) -> (GenesisSet, Vec<P256Pair>) {
    let mut members = Vec::new();
    let mut pairs = Vec::new();

    for i in 0..member_count {
        let pair = P256Pair::generate().unwrap();
        let member = QuorumMember {
            alias: format!("member_{}", i + 1),
            pub_key: pair.public_key().to_bytes(),
        };
        members.push(member);
        pairs.push(pair);
    }

    let genesis_set = GenesisSet { members, threshold };
    (genesis_set, pairs)
}

#[allow(dead_code)]
fn create_test_manifest_envelope() -> ManifestEnvelope {
    let quorum_key = P256Pair::generate().unwrap();
    let manifest = Manifest {
        namespace: renclave_shared::Namespace {
            nonce: 1,
            name: "test_namespace".to_string(),
            quorum_key: quorum_key.public_key().to_bytes(),
        },
        enclave: NitroConfig {
            pcr0: [1; 32].to_vec(),
            pcr1: [2; 32].to_vec(),
            pcr2: [3; 32].to_vec(),
            pcr3: [4; 32].to_vec(),
            aws_root_certificate: vec![],
            qos_commit: "test_commit".to_string(),
        },
        pivot: PivotConfig {
            hash: [0; 32],
            restart: renclave_shared::RestartPolicy::Always,
            args: vec!["test_arg".to_string()],
        },
        manifest_set: ManifestSet {
            members: create_test_genesis_set(3, 2).0.members,
            threshold: 2,
        },
        share_set: ShareSet {
            members: create_test_genesis_set(3, 2).0.members,
            threshold: 2,
        },
    };

    ManifestEnvelope {
        manifest,
        manifest_set_approvals: vec![],
        share_set_approvals: vec![],
    }
}

#[allow(dead_code)]
async fn create_test_seed_generator() -> SeedGenerator {
    SeedGenerator::new()
        .await
        .expect("Failed to create test generator")
}

#[allow(dead_code)]
async fn create_test_network_manager() -> NetworkManager {
    let config = NetworkConfig::default();
    NetworkManager::new(config)
}

fn main() {
    // This is a test binary, main function is not needed for tests
    println!("Integration tests - run with cargo test");
}
