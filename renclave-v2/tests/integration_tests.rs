use renclave_shared::{
    EnclaveOperation, EnclaveRequest, EnclaveResponse, EnclaveResult, RenclaveError,
};
use renclave_enclave::seed_generator::SeedGenerator;
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
    let is_invalid = seed_generator.validate_seed("invalid seed phrase").await.unwrap();
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
    let seed_with_pass = seed_generator.generate_seed(256, Some(passphrase)).await.unwrap();
    
    // Generate seed without passphrase
    let seed_without_pass = seed_generator.generate_seed(256, None).await.unwrap();
    
    // Seeds should be different
    assert_ne!(seed_with_pass.entropy, seed_without_pass.entropy);
    assert_ne!(seed_with_pass.phrase, seed_without_pass.phrase);
    
    // Both should be valid
    assert!(seed_generator.validate_seed(&seed_with_pass.phrase).await.unwrap());
    assert!(seed_generator.validate_seed(&seed_without_pass.phrase).await.unwrap());
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
        (EnclaveOperation::GenerateSeed { strength: s1, passphrase: p1 }, 
         EnclaveOperation::GenerateSeed { strength: s2, passphrase: p2 }) => {
            assert_eq!(s1, s2);
            assert_eq!(p1, p2);
        },
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
    
    let response_deserialized: EnclaveResponse = serde_json::from_str(&response_serialized.unwrap()).unwrap();
    assert_eq!(response.id, response_deserialized.id);
}

#[tokio::test]
async fn test_network_integration() {
    // Test network-related functionality
    use renclave_network::{NetworkConfig, NetworkManager};
    
    let config = NetworkConfig::default();
    let network_manager = NetworkManager::new(config);
    
    // Test network initialization (should not fail)
    let init_result = network_manager.initialize().await;
    // On non-QEMU systems, this might fail, which is expected
    if init_result.is_err() {
        println!("Network initialization failed (expected on non-QEMU systems): {:?}", init_result.err());
    }
}

#[tokio::test]
async fn test_concurrent_seed_generation() {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());
    
    // Generate multiple seeds concurrently
    let mut handles = vec![];
    
    for _ in 0..5 {
        let generator = Arc::clone(&seed_generator);
        let handle = tokio::spawn(async move {
            generator.generate_seed(256, None).await
        });
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
        assert!(result.is_ok(), "Failed to generate seed with strength {}", strength);
        
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
        assert!(phrases.insert(seed.phrase.clone()), "Duplicate seed phrase generated");
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
    let seed3 = seed_generator.generate_seed(256, Some(&long_passphrase)).await.unwrap();
    
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
        let handle = tokio::spawn(async move {
            generator.validate_seed(&phrase).await
        });
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
