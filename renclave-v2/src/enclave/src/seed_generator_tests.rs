//! Comprehensive unit tests for SeedGenerator module
//! Tests all functionality including edge cases, error handling, and security properties

use anyhow::Result;
use renclave_enclave::seed_generator::{SeedGenerator, SeedResult};
use std::collections::HashSet;
use tokio::time::{sleep, Duration};

/// Test utilities for seed generation
mod test_utils {
    use super::*;

    pub async fn create_test_generator() -> SeedGenerator {
        SeedGenerator::new()
            .await
            .expect("Failed to create test generator")
    }

    pub fn assert_valid_entropy_length(entropy: &str, strength: u32) {
        let expected_length = (strength / 8) * 2; // Convert bytes to hex chars
        assert_eq!(
            entropy.len(),
            expected_length as usize,
            "Entropy length mismatch for strength {}: expected {}, got {}",
            strength,
            expected_length,
            entropy.len()
        );
    }

    pub fn assert_valid_hex(entropy: &str) {
        assert!(
            entropy.chars().all(|c| c.is_ascii_hexdigit()),
            "Entropy is not valid hex: {}",
            entropy
        );
    }

    pub fn assert_word_count(phrase: &str, expected_words: usize) {
        let actual_words = phrase.split_whitespace().count();
        assert_eq!(
            actual_words, expected_words,
            "Word count mismatch: expected {}, got {}",
            expected_words, actual_words
        );
    }
}

#[tokio::test]
async fn test_seed_generator_creation() {
    let generator = test_utils::create_test_generator().await;
    assert!(generator.rng.try_lock().is_ok());
}

#[tokio::test]
async fn test_generate_seed_128_bits() {
    let generator = test_utils::create_test_generator().await;

    let result = generator.generate_seed(128, None).await;
    assert!(result.is_ok());

    let seed = result.unwrap();
    assert_eq!(seed.strength, 128);
    assert_eq!(seed.word_count, 12);
    test_utils::assert_word_count(&seed.phrase, 12);
    test_utils::assert_valid_entropy_length(&seed.entropy, 128);
    test_utils::assert_valid_hex(&seed.entropy);
}

#[tokio::test]
async fn test_generate_seed_256_bits() {
    let generator = test_utils::create_test_generator().await;

    let result = generator.generate_seed(256, None).await;
    assert!(result.is_ok());

    let seed = result.unwrap();
    assert_eq!(seed.strength, 256);
    assert_eq!(seed.word_count, 24);
    test_utils::assert_word_count(&seed.phrase, 24);
    test_utils::assert_valid_entropy_length(&seed.entropy, 256);
    test_utils::assert_valid_hex(&seed.entropy);
}

#[tokio::test]
async fn test_generate_seed_all_valid_strengths() {
    let generator = test_utils::create_test_generator().await;

    let valid_strengths = vec![128, 160, 192, 224, 256];
    let expected_words = vec![12, 15, 18, 21, 24];

    for (strength, expected_words) in valid_strengths.iter().zip(expected_words.iter()) {
        let result = generator.generate_seed(*strength, None).await;
        assert!(
            result.is_ok(),
            "Failed to generate seed with strength {}",
            strength
        );

        let seed = result.unwrap();
        assert_eq!(seed.strength, *strength);
        assert_eq!(seed.word_count, *expected_words);
        test_utils::assert_word_count(&seed.phrase, *expected_words);
        test_utils::assert_valid_entropy_length(&seed.entropy, *strength);
    }
}

#[tokio::test]
async fn test_generate_seed_with_passphrase() {
    let generator = test_utils::create_test_generator().await;
    let passphrase = "test-passphrase-123";

    let result = generator.generate_seed(256, Some(passphrase)).await;
    assert!(result.is_ok());

    let seed = result.unwrap();
    assert!(seed.phrase.ends_with(passphrase));
    test_utils::assert_word_count(&seed.phrase, 25); // 24 words + 1 passphrase
}

#[tokio::test]
async fn test_generate_seed_with_empty_passphrase() {
    let generator = test_utils::create_test_generator().await;

    let seed_with_empty = generator.generate_seed(256, Some("")).await.unwrap();
    let seed_without = generator.generate_seed(256, None).await.unwrap();

    // Should be different due to passphrase handling
    assert_ne!(seed_with_empty.phrase, seed_without.phrase);
}

#[tokio::test]
async fn test_generate_seed_invalid_strengths() {
    let generator = test_utils::create_test_generator().await;

    let invalid_strengths = vec![0, 64, 96, 100, 288, 320];

    for strength in invalid_strengths {
        let result = generator.generate_seed(strength, None).await;
        assert!(
            result.is_err(),
            "Should fail with invalid strength {}",
            strength
        );

        let error = result.unwrap_err();
        assert!(error.to_string().contains("Invalid strength"));
    }
}

#[tokio::test]
async fn test_seed_uniqueness() {
    let generator = test_utils::create_test_generator().await;
    let mut phrases = HashSet::new();
    let mut entropies = HashSet::new();

    // Generate multiple seeds and ensure uniqueness
    for _ in 0..20 {
        let seed = generator.generate_seed(256, None).await.unwrap();

        assert!(
            phrases.insert(seed.phrase.clone()),
            "Duplicate seed phrase generated: {}",
            seed.phrase
        );
        assert!(
            entropies.insert(seed.entropy.clone()),
            "Duplicate entropy generated: {}",
            seed.entropy
        );
    }
}

#[tokio::test]
async fn test_seed_validation_valid_phrases() {
    let generator = test_utils::create_test_generator().await;

    // Test with generated seeds
    for strength in &[128, 256] {
        let seed = generator.generate_seed(*strength, None).await.unwrap();
        let is_valid = generator.validate_seed(&seed.phrase).await.unwrap();
        assert!(is_valid, "Generated seed should be valid");
    }

    // Test with known valid BIP39 phrases
    let valid_phrases = vec![
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art",
    ];

    for phrase in valid_phrases {
        let is_valid = generator.validate_seed(phrase).await.unwrap();
        assert!(is_valid, "Known valid phrase should be valid: {}", phrase);
    }
}

#[tokio::test]
async fn test_seed_validation_invalid_phrases() {
    let generator = test_utils::create_test_generator().await;

    let invalid_phrases = vec![
        "",
        "   ",
        "invalid seed phrase",
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid",
        "not a valid seed phrase here",
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid",
    ];

    for phrase in invalid_phrases {
        let is_valid = generator.validate_seed(phrase).await.unwrap();
        assert!(!is_valid, "Invalid phrase should be invalid: '{}'", phrase);
    }
}

#[tokio::test]
async fn test_seed_validation_with_passphrase() {
    let generator = test_utils::create_test_generator().await;
    let passphrase = "test-passphrase";

    // Generate seed with passphrase
    let seed = generator
        .generate_seed(256, Some(passphrase))
        .await
        .unwrap();

    // Should validate as valid (with passphrase)
    let is_valid = generator.validate_seed(&seed.phrase).await.unwrap();
    assert!(is_valid);

    // Extract just the mnemonic part (without passphrase)
    let words: Vec<&str> = seed.phrase.split_whitespace().collect();
    let mnemonic_only = words[..24].join(" ");

    // Should also be valid as BIP39 mnemonic
    let is_mnemonic_valid = generator.validate_seed(&mnemonic_only).await.unwrap();
    assert!(is_mnemonic_valid);
}

#[tokio::test]
async fn test_concurrent_seed_generation() {
    let generator = std::sync::Arc::new(test_utils::create_test_generator().await);
    let mut handles = vec![];

    // Generate multiple seeds concurrently
    for i in 0..10 {
        let gen = Arc::clone(&generator);
        let handle = tokio::spawn(async move {
            let strength = if i % 2 == 0 { 128 } else { 256 };
            let passphrase = if i % 3 == 0 {
                Some(format!("pass-{}", i))
            } else {
                None
            };

            let seed = gen
                .generate_seed(strength, passphrase.as_deref())
                .await
                .unwrap();
            assert_eq!(seed.strength, strength);
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
async fn test_concurrent_validation() {
    let generator = std::sync::Arc::new(test_utils::create_test_generator().await);

    // Generate a valid seed
    let seed = generator.generate_seed(256, None).await.unwrap();

    // Validate it concurrently multiple times
    let mut handles = vec![];

    for _ in 0..10 {
        let gen = Arc::clone(&generator);
        let phrase = seed.phrase.clone();
        let handle = tokio::spawn(async move { gen.validate_seed(&phrase).await.unwrap() });
        handles.push(handle);
    }

    // All validations should succeed
    let results: Vec<_> = futures::future::join_all(handles).await;
    for result in results {
        assert!(result.unwrap());
    }
}

#[tokio::test]
async fn test_entropy_quality() {
    let generator = test_utils::create_test_generator().await;

    // Generate multiple entropy samples and verify quality
    let mut has_non_zero = false;
    let mut entropy_samples = Vec::new();

    for _ in 0..10 {
        let seed = generator.generate_seed(256, None).await.unwrap();
        let entropy_bytes = hex::decode(&seed.entropy).unwrap();

        if entropy_bytes.iter().any(|&b| b != 0) {
            has_non_zero = true;
        }

        entropy_samples.push(entropy_bytes);
    }

    assert!(has_non_zero, "All entropy samples were zero");

    // Verify entropy samples are different
    for i in 0..entropy_samples.len() {
        for j in (i + 1)..entropy_samples.len() {
            assert_ne!(
                entropy_samples[i], entropy_samples[j],
                "Entropy samples should be different"
            );
        }
    }
}

#[tokio::test]
async fn test_derive_seed_from_mnemonic() {
    let generator = test_utils::create_test_generator().await;

    // Generate a mnemonic
    let seed_result = generator.generate_seed(256, None).await.unwrap();
    let mnemonic = seed_result.phrase;

    // Derive seed from mnemonic
    let derived_seed = generator.derive_seed(&mnemonic, None).await.unwrap();
    assert_eq!(derived_seed.len(), 64); // BIP39 seed is 64 bytes

    // Derive with passphrase
    let passphrase = "test-passphrase";
    let derived_with_pass = generator
        .derive_seed(&mnemonic, Some(passphrase))
        .await
        .unwrap();
    assert_eq!(derived_with_pass.len(), 64);

    // Should be different
    assert_ne!(derived_seed, derived_with_pass);
}

#[tokio::test]
async fn test_verify_entropy_consistency() {
    let generator = test_utils::create_test_generator().await;

    // Generate entropy and mnemonic
    let seed_result = generator.generate_seed(256, None).await.unwrap();
    let entropy_bytes = hex::decode(&seed_result.entropy).unwrap();

    // Verify entropy matches mnemonic
    let is_consistent = generator
        .verify_entropy(&entropy_bytes, &seed_result.phrase)
        .await
        .unwrap();
    assert!(is_consistent);

    // Test with wrong entropy
    let wrong_entropy = vec![0u8; 32];
    let is_wrong_consistent = generator
        .verify_entropy(&wrong_entropy, &seed_result.phrase)
        .await
        .unwrap();
    assert!(!is_wrong_consistent);
}

#[tokio::test]
async fn test_key_derivation() {
    let generator = test_utils::create_test_generator().await;

    // Generate a seed
    let seed_result = generator.generate_seed(256, None).await.unwrap();

    // Derive key
    let key_result = generator
        .derive_key(&seed_result.phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();

    assert!(!key_result.private_key.is_empty());
    assert!(!key_result.public_key.is_empty());
    assert!(!key_result.address.is_empty());
    assert_eq!(key_result.path, "m/44'/0'/0'/0/0");
    assert_eq!(key_result.curve, "secp256k1");

    // Note: private_key is now encrypted in the actual implementation
    // This test verifies the basic structure, but the private key will be encrypted
    // when used through the main enclave interface
}

#[tokio::test]
async fn test_key_derivation_with_passphrase() {
    let generator = test_utils::create_test_generator().await;

    // Generate a seed with passphrase
    let seed_result = generator
        .generate_seed(256, Some("test-passphrase"))
        .await
        .unwrap();

    // Test that the derive_key method properly handles seed phrases with passphrases
    // This test verifies that the method can extract mnemonic and passphrase correctly
    let key_result = generator
        .derive_key(&seed_result.phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();

    assert!(!key_result.private_key.is_empty());
    assert!(!key_result.public_key.is_empty());
    assert!(!key_result.address.is_empty());
    assert_eq!(key_result.path, "m/44'/0'/0'/0/0");
    assert_eq!(key_result.curve, "secp256k1");

    // Verify that the seed phrase has the expected structure (24 words + passphrase)
    let words: Vec<&str> = seed_result.phrase.split_whitespace().collect();
    assert_eq!(words.len(), 25); // 24 words + 1 passphrase
    assert_eq!(words[24], "test-passphrase"); // Last word should be the passphrase
}

#[tokio::test]
async fn test_mnemonic_and_passphrase_extraction() {
    let generator = test_utils::create_test_generator().await;

    // Test with a 25-word phrase (24 words + passphrase)
    let test_phrase = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about test";

    // This should work with the new extract_mnemonic_and_passphrase method
    let key_result = generator
        .derive_key(test_phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();

    assert!(!key_result.private_key.is_empty());
    assert!(!key_result.public_key.is_empty());
    assert!(!key_result.address.is_empty());

    // Test with a plain 24-word phrase (no passphrase)
    let plain_phrase = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

    let key_result_plain = generator
        .derive_key(plain_phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();

    assert!(!key_result_plain.private_key.is_empty());
    assert!(!key_result_plain.public_key.is_empty());
    assert!(!key_result_plain.address.is_empty());

    // The results should be different because one has a passphrase and one doesn't
    assert_ne!(key_result.private_key, key_result_plain.private_key);
    assert_ne!(key_result.public_key, key_result_plain.public_key);
    assert_ne!(key_result.address, key_result_plain.address);
}

#[tokio::test]
async fn test_address_derivation() {
    let generator = test_utils::create_test_generator().await;

    // Generate a seed
    let seed_result = generator.generate_seed(256, None).await.unwrap();

    // Derive address
    let address_result = generator
        .derive_address(&seed_result.phrase, "m/44'/0'/0'/0/0", "secp256k1")
        .await
        .unwrap();

    assert!(!address_result.address.is_empty());
}

#[tokio::test]
async fn test_derivation_path_validation() {
    let generator = test_utils::create_test_generator().await;
    let seed_result = generator.generate_seed(256, None).await.unwrap();

    // Test valid derivation paths
    let valid_paths = vec![
        "m/44'/0'/0'/0/0",
        "m/44'/0'/0'/0/1",
        "m/44'/0'/0'/1/0",
        "m/84'/0'/0'/0/0",
    ];

    for path in valid_paths {
        let result = generator
            .derive_key(&seed_result.phrase, path, "secp256k1")
            .await;
        assert!(
            result.is_ok(),
            "Valid derivation path should work: {}",
            path
        );
    }

    // Test invalid derivation paths
    let invalid_paths = vec![
        "invalid/path",
        "m/44'/0'/0'/0",     // Missing final component
        "m/44'/0'/0'/0/0/0", // Too many components
        "",                  // Empty path
    ];

    for path in invalid_paths {
        let result = generator
            .derive_key(&seed_result.phrase, path, "secp256k1")
            .await;
        assert!(
            result.is_err(),
            "Invalid derivation path should fail: {}",
            path
        );
    }
}

#[tokio::test]
async fn test_stress_generation() {
    let generator = test_utils::create_test_generator().await;

    // Generate many seeds rapidly to test system stability
    let mut handles = vec![];

    for i in 0..50 {
        let gen = generator.clone();
        let handle = tokio::spawn(async move {
            let strength = if i % 3 == 0 { 128 } else { 256 };
            let passphrase = if i % 5 == 0 {
                Some(format!("stress-{}", i))
            } else {
                None
            };

            let seed = gen
                .generate_seed(strength, passphrase.as_deref())
                .await
                .unwrap();

            // Validate the seed
            let is_valid = gen.validate_seed(&seed.phrase).await.unwrap();
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
async fn test_error_recovery() {
    let generator = test_utils::create_test_generator().await;

    // Test that system can recover from errors
    let mut successful = 0;
    let mut failed = 0;

    for i in 0..20 {
        let strength = if i % 4 == 0 { 0 } else { 256 }; // Some invalid, some valid

        match generator.generate_seed(strength, None).await {
            Ok(seed) => {
                successful += 1;
                assert_eq!(seed.strength, 256);

                // Validate successful seed
                let is_valid = generator.validate_seed(&seed.phrase).await.unwrap();
                assert!(is_valid);
            }
            Err(_) => {
                failed += 1;
                // Expected for invalid strength
            }
        }
    }

    assert!(successful > 0, "Should have some successful generations");
    assert!(failed > 0, "Should have some failed generations");
    assert_eq!(successful + failed, 20);
}

#[tokio::test]
async fn test_memory_usage() {
    let generator = test_utils::create_test_generator().await;

    // Generate many seeds to test memory usage
    let mut seeds = Vec::new();

    for _ in 0..100 {
        let seed = generator.generate_seed(256, None).await.unwrap();
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
async fn test_entropy_distribution() {
    let generator = test_utils::create_test_generator().await;

    // Generate many entropy samples and check distribution
    let mut byte_counts = [0u32; 256];
    let sample_count = 1000;

    for _ in 0..sample_count {
        let seed = generator.generate_seed(256, None).await.unwrap();
        let entropy_bytes = hex::decode(&seed.entropy).unwrap();

        for &byte in &entropy_bytes {
            byte_counts[byte as usize] += 1;
        }
    }

    // Check that we have reasonable distribution (not all zeros)
    let non_zero_count = byte_counts.iter().filter(|&&count| count > 0).count();
    assert!(
        non_zero_count > 100,
        "Entropy distribution seems poor: {} non-zero bytes",
        non_zero_count
    );

    // Check that no single byte value dominates
    let max_count = byte_counts.iter().max().unwrap();
    let expected_max = (sample_count * 32) / 256 * 3; // Allow 3x expected
    assert!(
        *max_count < expected_max as u32,
        "Some byte values appear too frequently: max={}, expected_max={}",
        max_count,
        expected_max
    );
}
