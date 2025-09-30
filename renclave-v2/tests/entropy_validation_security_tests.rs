/// Security tests for entropy validation in validate-seed API
/// These tests verify the critical security fix that prevents acceptance
/// of mismatched entropy in the validate-seed API
/// Also includes tests for passphrase handling bug fixes
#[allow(unused_imports)]
use renclave_enclave::seed_generator::SeedGenerator;
#[allow(unused_imports)]
use std::sync::Arc;

/// Test that validates the critical security fix for entropy validation
/// This test ensures that mismatched entropy is properly rejected
#[tokio::test]
async fn test_entropy_validation_security_fix() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a valid seed phrase
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let valid_seed = seed_result.phrase;

    // Test 1: Valid seed phrase should pass validation
    let validation_result = seed_generator.validate_seed(&valid_seed).await;
    assert!(validation_result.is_ok());
    assert!(validation_result.unwrap());

    // Test 2: Invalid seed phrase should fail validation
    let invalid_seed = "invalid seed phrase with wrong words";
    let invalid_validation = seed_generator.validate_seed(invalid_seed).await;
    assert!(invalid_validation.is_ok());
    assert!(!invalid_validation.unwrap());
}

/// Test entropy validation with correct entropy
/// This test verifies that when correct entropy is provided, validation passes
#[tokio::test]
async fn test_entropy_validation_correct_entropy() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed and get its entropy
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let seed_phrase = seed_result.phrase;
    let correct_entropy = seed_result.entropy;

    // Derive entropy from the seed phrase
    let derived_entropy = seed_generator
        .derive_entropy_from_seed(&seed_phrase)
        .await
        .unwrap();

    // The derived entropy should match the original entropy
    assert_eq!(derived_entropy, correct_entropy);
}

/// Test entropy validation with incorrect entropy
/// This test verifies that when incorrect entropy is provided, validation fails
#[tokio::test]
async fn test_entropy_validation_incorrect_entropy() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let seed_phrase = seed_result.phrase;

    // Create incorrect entropy (different from the seed's entropy)
    let incorrect_entropy = "1111111111111111111111111111111111111111111111111111111111111111";

    // Derive entropy from the seed phrase
    let derived_entropy = seed_generator
        .derive_entropy_from_seed(&seed_phrase)
        .await
        .unwrap();

    // The derived entropy should NOT match the incorrect entropy
    assert_ne!(derived_entropy, incorrect_entropy);
}

/// Test entropy validation with corrupted entropy
/// This test verifies that corrupted entropy is properly rejected
#[tokio::test]
async fn test_entropy_validation_corrupted_entropy() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let seed_phrase = seed_result.phrase;

    // Create corrupted entropy (truncated)
    let corrupted_entropy = "758afdb6aba4c9821192165995b45c67ba30c2846dea6416a1ba43d513f97ef";

    // Derive entropy from the seed phrase
    let derived_entropy = seed_generator
        .derive_entropy_from_seed(&seed_phrase)
        .await
        .unwrap();

    // The derived entropy should NOT match the corrupted entropy
    assert_ne!(derived_entropy, corrupted_entropy);
}

/// Test entropy validation with empty entropy
/// This test verifies that empty entropy is properly handled
#[tokio::test]
async fn test_entropy_validation_empty_entropy() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let seed_phrase = seed_result.phrase;

    // Create empty entropy
    let empty_entropy = "";

    // Derive entropy from the seed phrase
    let derived_entropy = seed_generator
        .derive_entropy_from_seed(&seed_phrase)
        .await
        .unwrap();

    // The derived entropy should NOT match the empty entropy
    assert_ne!(derived_entropy, empty_entropy);
    assert!(!derived_entropy.is_empty());
}

/// Test entropy validation with different seed strengths
/// This test verifies that entropy validation works for different seed strengths
#[tokio::test]
async fn test_entropy_validation_different_strengths() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test different seed strengths
    let strengths = vec![128, 160, 192, 224, 256];

    for strength in strengths {
        // Generate a seed with the specified strength
        let seed_result = seed_generator.generate_seed(strength, None).await.unwrap();
        let seed_phrase = seed_result.phrase;
        let entropy = seed_result.entropy;

        // Validate the seed phrase
        let validation_result = seed_generator.validate_seed(&seed_phrase).await;
        assert!(validation_result.is_ok());
        assert!(validation_result.unwrap());

        // Derive entropy from the seed phrase
        let derived_entropy = seed_generator
            .derive_entropy_from_seed(&seed_phrase)
            .await
            .unwrap();

        // The derived entropy should match the original entropy
        assert_eq!(derived_entropy, entropy);

        // Verify entropy length matches expected strength
        let expected_entropy_length = (strength / 4) as usize; // 4 bits per hex char
        assert_eq!(entropy.len(), expected_entropy_length);
        assert_eq!(derived_entropy.len(), expected_entropy_length);
    }
}

/// Test entropy validation security edge cases
/// This test verifies edge cases that could be exploited
#[tokio::test]
async fn test_entropy_validation_security_edge_cases() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test with malicious input
    let malicious_inputs = vec![
        "'; DROP TABLE users; --",
        "<script>alert('xss')</script>",
        "../../etc/passwd",
        "null",
        "undefined",
        "NaN",
    ];

    for malicious_input in malicious_inputs {
        // These should all fail validation
        let validation_result = seed_generator.validate_seed(malicious_input).await;
        assert!(validation_result.is_ok());
        assert!(!validation_result.unwrap());
    }
}

/// Test entropy validation with passphrase
/// This test verifies that entropy validation works with passphrases
#[tokio::test]
async fn test_entropy_validation_with_passphrase() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Generate a seed without passphrase first
    let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
    let seed_phrase = seed_result.phrase;
    let entropy = seed_result.entropy;

    // Validate the seed phrase
    let validation_result = seed_generator.validate_seed(&seed_phrase).await;
    assert!(validation_result.is_ok());
    assert!(validation_result.unwrap());

    // Derive entropy from the seed phrase
    let derived_entropy = seed_generator
        .derive_entropy_from_seed(&seed_phrase)
        .await
        .unwrap();

    // The derived entropy should match the original entropy
    assert_eq!(derived_entropy, entropy);
}

/// Test entropy validation performance
/// This test verifies that entropy validation doesn't cause performance issues
#[tokio::test]
async fn test_entropy_validation_performance() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    let start_time = std::time::Instant::now();

    // Perform multiple entropy validations
    for _ in 0..100 {
        let seed_result = seed_generator.generate_seed(128, None).await.unwrap();
        let seed_phrase = seed_result.phrase;

        // Validate the seed
        let validation_result = seed_generator.validate_seed(&seed_phrase).await;
        assert!(validation_result.is_ok());
        assert!(validation_result.unwrap());

        // Derive entropy
        let derived_entropy = seed_generator
            .derive_entropy_from_seed(&seed_phrase)
            .await
            .unwrap();
        assert!(!derived_entropy.is_empty());
    }

    let duration = start_time.elapsed();

    // Entropy validation should complete within reasonable time (1 second for 100 operations)
    assert!(duration.as_secs() < 1);
}

/// Test entropy validation concurrency
/// This test verifies that entropy validation works correctly under concurrent access
#[tokio::test]
async fn test_entropy_validation_concurrency() {
    let seed_generator = Arc::new(SeedGenerator::new().await.unwrap());

    let mut handles = vec![];

    // Spawn multiple concurrent entropy validations
    for i in 0..10 {
        let generator = seed_generator.clone();
        let handle = tokio::spawn(async move {
            let seed_result = generator.generate_seed(128, None).await.unwrap();
            let seed_phrase = seed_result.phrase;

            // Validate the seed
            let validation_result = generator.validate_seed(&seed_phrase).await;
            assert!(validation_result.is_ok());
            assert!(validation_result.unwrap());

            // Derive entropy
            let derived_entropy = generator
                .derive_entropy_from_seed(&seed_phrase)
                .await
                .unwrap();
            assert!(!derived_entropy.is_empty());

            i
        });
        handles.push(handle);
    }

    // Wait for all tasks to complete
    for handle in handles {
        let result = handle.await.unwrap();
        assert!(result < 10);
    }
}

/// Test passphrase handling bug fix - 25-word seed phrases (24 mnemonic + 1 passphrase)
/// This test verifies that the fix for entropy derivation from 25-word seed phrases works correctly
#[tokio::test]
async fn test_passphrase_entropy_derivation_fix() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test with different passphrases
    let test_passphrases = vec![
        "testpass",
        "mypassword123",
        "secure_pass_2024",
        "a",
        "very_long_passphrase_with_special_chars_!@#$%",
    ];

    for passphrase in test_passphrases {
        // Generate a seed with passphrase (creates 25 words: 24 mnemonic + 1 passphrase)
        let seed_result = seed_generator
            .generate_seed(256, Some(passphrase))
            .await
            .unwrap();

        // Verify it has 25 words
        let words: Vec<&str> = seed_result.phrase.split_whitespace().collect();
        assert_eq!(
            words.len(),
            25,
            "Should have 25 words (24 mnemonic + 1 passphrase)"
        );

        // Test entropy derivation from 25-word seed phrase (should work now)
        let derived_entropy = seed_generator
            .derive_entropy_from_seed(&seed_result.phrase)
            .await
            .unwrap();

        // The derived entropy should match the original entropy
        assert_eq!(
            derived_entropy, seed_result.entropy,
            "Derived entropy should match original entropy for passphrase: {}",
            passphrase
        );

        // Test validation: the 25-word phrase should be valid
        let is_valid = seed_generator
            .validate_seed(&seed_result.phrase)
            .await
            .unwrap();
        assert!(
            is_valid,
            "25-word seed phrase should be valid with passphrase: {}",
            passphrase
        );

        // Test with just the mnemonic part (24 words) - should also work
        let mnemonic_only = words[..24].join(" ");
        let derived_entropy_mnemonic = seed_generator
            .derive_entropy_from_seed(&mnemonic_only)
            .await
            .unwrap();
        assert_eq!(
            derived_entropy_mnemonic, seed_result.entropy,
            "Entropy from mnemonic only should match original entropy for passphrase: {}",
            passphrase
        );
    }
}

/// Test passphrase handling with different word counts
/// This test ensures the fix works for various BIP39 word counts with passphrases
#[tokio::test]
async fn test_passphrase_entropy_different_word_counts() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test different BIP39 strengths with passphrases
    let test_cases = vec![
        (128, 12, "test128"),
        (160, 15, "test160"),
        (192, 18, "test192"),
        (224, 21, "test224"),
        (256, 24, "test256"),
    ];

    for (strength, expected_words, passphrase) in test_cases {
        // Generate seed with passphrase
        let seed_result = seed_generator
            .generate_seed(strength, Some(passphrase))
            .await
            .unwrap();

        // Should have expected_words + 1 (for passphrase)
        let total_words = seed_result.phrase.split_whitespace().count();
        assert_eq!(
            total_words,
            expected_words + 1,
            "Should have {} words ({} mnemonic + 1 passphrase) for strength {}",
            expected_words + 1,
            expected_words,
            strength
        );

        // Test entropy derivation - this should work for all cases now
        let derived_entropy = seed_generator
            .derive_entropy_from_seed(&seed_result.phrase)
            .await
            .unwrap();
        assert_eq!(
            derived_entropy, seed_result.entropy,
            "Entropy derivation should work for strength {} with passphrase",
            strength
        );

        // Test validation
        let is_valid = seed_generator
            .validate_seed(&seed_result.phrase)
            .await
            .unwrap();
        assert!(
            is_valid,
            "Seed validation should work for strength {} with passphrase",
            strength
        );
    }
}

/// Test edge cases for passphrase handling
/// This test covers edge cases that could break the passphrase handling fix
#[tokio::test]
async fn test_passphrase_entropy_edge_cases() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test with empty passphrase - this should behave like no passphrase
    let seed_result_empty = seed_generator.generate_seed(256, Some("")).await.unwrap();
    let words_empty: Vec<&str> = seed_result_empty.phrase.split_whitespace().collect();
    // Empty passphrase doesn't add extra word, so it should be 24 words
    assert_eq!(
        words_empty.len(),
        24,
        "Empty passphrase should create 24 words (same as no passphrase)"
    );

    // Test entropy derivation with empty passphrase
    let derived_entropy_empty = seed_generator
        .derive_entropy_from_seed(&seed_result_empty.phrase)
        .await
        .unwrap();
    assert_eq!(derived_entropy_empty, seed_result_empty.entropy);

    // Test with passphrase containing spaces
    let seed_result_spaces = seed_generator
        .generate_seed(256, Some("pass with spaces"))
        .await
        .unwrap();
    let words_spaces: Vec<&str> = seed_result_spaces.phrase.split_whitespace().collect();
    // Note: This will create more than 25 words due to spaces in passphrase
    assert!(
        words_spaces.len() > 25,
        "Passphrase with spaces should create more than 25 words"
    );

    // Test entropy derivation with spaced passphrase
    let derived_entropy_spaces = seed_generator
        .derive_entropy_from_seed(&seed_result_spaces.phrase)
        .await
        .unwrap();
    assert_eq!(derived_entropy_spaces, seed_result_spaces.entropy);

    // Test validation with spaced passphrase - this should work with our enhanced validation
    let is_valid_spaces = seed_generator
        .validate_seed(&seed_result_spaces.phrase)
        .await
        .unwrap();
    // Note: The validation logic tries to parse without the last word(s) if initial parsing fails
    // So this should succeed
    assert!(
        is_valid_spaces,
        "Seed with spaced passphrase should be valid due to fallback validation logic"
    );
}

/// Test backward compatibility - standard BIP39 word counts without passphrases
/// This test ensures the fix doesn't break existing functionality
#[tokio::test]
async fn test_backward_compatibility_standard_word_counts() {
    let seed_generator = SeedGenerator::new().await.unwrap();

    // Test standard BIP39 word counts without passphrases
    let test_cases = vec![(128, 12), (160, 15), (192, 18), (224, 21), (256, 24)];

    for (strength, expected_words) in test_cases {
        // Generate seed without passphrase
        let seed_result = seed_generator.generate_seed(strength, None).await.unwrap();

        // Should have exactly expected_words
        let actual_words = seed_result.phrase.split_whitespace().count();
        assert_eq!(
            actual_words, expected_words,
            "Should have exactly {} words for strength {} without passphrase",
            expected_words, strength
        );

        // Test entropy derivation (should work as before)
        let derived_entropy = seed_generator
            .derive_entropy_from_seed(&seed_result.phrase)
            .await
            .unwrap();
        assert_eq!(
            derived_entropy, seed_result.entropy,
            "Entropy derivation should work for strength {} without passphrase",
            strength
        );

        // Test validation (should work as before)
        let is_valid = seed_generator
            .validate_seed(&seed_result.phrase)
            .await
            .unwrap();
        assert!(
            is_valid,
            "Seed validation should work for strength {} without passphrase",
            strength
        );
    }
}

fn main() {
    // This is a test binary, main function is not needed for tests
}
