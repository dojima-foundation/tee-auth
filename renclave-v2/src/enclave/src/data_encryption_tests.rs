//! Comprehensive unit tests for DataEncryption module
//! Tests ECDH key agreement, AES-GCM encryption, and data protection

use anyhow::Result;
use renclave_enclave::data_encryption::{
    DataEncryption, P256EncryptPair, P256EncryptPublic, Envelope
};
use renclave_enclave::quorum::P256Pair;
use std::collections::HashSet;

/// Test utilities for data encryption
mod test_utils {
    use super::*;
    
    pub fn create_test_data_encryption() -> DataEncryption {
        let quorum_key = P256Pair::generate().unwrap();
        DataEncryption::new(quorum_key)
    }
    
    pub fn create_test_encrypt_pair() -> P256EncryptPair {
        P256EncryptPair::generate()
    }
    
    pub fn create_test_encrypt_public() -> P256EncryptPublic {
        let pair = P256EncryptPair::generate();
        P256EncryptPublic::from_private(&pair)
    }
    
    pub fn assert_valid_envelope(envelope: &Envelope) {
        assert_eq!(envelope.nonce.len(), 12, "Nonce should be 12 bytes");
        assert_eq!(envelope.ephemeral_sender_public.len(), 65, "Public key should be 65 bytes");
        assert!(!envelope.encrypted_message.is_empty(), "Encrypted message should not be empty");
    }
    
    pub fn assert_valid_encrypted_data(encrypted: &[u8]) {
        assert!(!encrypted.is_empty(), "Encrypted data should not be empty");
        // Should be longer than original due to envelope structure
        assert!(encrypted.len() > 12, "Encrypted data should include nonce");
    }
}

#[tokio::test]
async fn test_data_encryption_creation() {
    let quorum_key = P256Pair::generate().unwrap();
    let data_encryption = DataEncryption::new(quorum_key);
    
    // Should be created successfully
    assert!(data_encryption.quorum_key.public_key().to_bytes().len() == 65);
}

#[tokio::test]
async fn test_p256_encrypt_pair_generation() {
    let pair1 = P256EncryptPair::generate();
    let pair2 = P256EncryptPair::generate();
    
    // Generated pairs should be different
    assert_ne!(pair1.private_key_bytes(), pair2.private_key_bytes());
    
    // Should have valid private key bytes
    assert_eq!(pair1.private_key_bytes().len(), 32);
    assert_eq!(pair2.private_key_bytes().len(), 32);
}

#[tokio::test]
async fn test_p256_encrypt_pair_from_bytes() {
    let original_pair = P256EncryptPair::generate();
    let private_bytes = original_pair.private_key_bytes();
    
    // Create new pair from bytes
    let reconstructed_pair = P256EncryptPair::from_bytes(&private_bytes).unwrap();
    
    // Should be identical
    assert_eq!(original_pair.private_key_bytes(), reconstructed_pair.private_key_bytes());
}

#[tokio::test]
async fn test_p256_encrypt_pair_from_invalid_bytes() {
    // Test with wrong length bytes
    let invalid_bytes = vec![1, 2, 3]; // Too short
    assert!(P256EncryptPair::from_bytes(&invalid_bytes).is_err());
    
    // Test with all zeros
    let zero_bytes = vec![0; 32];
    // This might succeed or fail depending on implementation
    let _ = P256EncryptPair::from_bytes(&zero_bytes);
}

#[tokio::test]
async fn test_p256_encrypt_public_creation() {
    let pair = P256EncryptPair::generate();
    let public = P256EncryptPublic::from_private(&pair);
    
    // Should have valid public key
    assert_eq!(public.to_bytes().len(), 65);
    assert_eq!(public.to_bytes()[0], 0x04); // Uncompressed format
}

#[tokio::test]
async fn test_p256_encrypt_public_from_bytes() {
    let pair = P256EncryptPair::generate();
    let original_public = P256EncryptPublic::from_private(&pair);
    let public_bytes = original_public.to_bytes();
    
    // Create public key from bytes
    let reconstructed_public = P256EncryptPublic::from_bytes(&public_bytes).unwrap();
    
    // Should be identical
    assert_eq!(original_public.to_bytes(), reconstructed_public.to_bytes());
}

#[tokio::test]
async fn test_p256_encrypt_public_from_invalid_bytes() {
    // Test with wrong length bytes
    let invalid_bytes = vec![1, 2, 3]; // Too short
    assert!(P256EncryptPublic::from_bytes(&invalid_bytes).is_err());
    
    // Test with all zeros
    let zero_bytes = vec![0; 65];
    // This might succeed or fail depending on implementation
    let _ = P256EncryptPublic::from_bytes(&zero_bytes);
}

#[tokio::test]
async fn test_basic_encryption_decryption() {
    let data_encryption = test_utils::create_test_data_encryption();
    let test_data = b"test data for encryption";
    let recipient_public = vec![1, 2, 3, 4, 5]; // Mock recipient public key
    
    // Encrypt data
    let encrypted = data_encryption.encrypt_data(test_data, &recipient_public).unwrap();
    test_utils::assert_valid_encrypted_data(&encrypted);
    
    // Decrypt data
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(test_data.to_vec(), decrypted);
}

#[tokio::test]
async fn test_encryption_different_data_sizes() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    let test_cases = vec![
        b"".to_vec(), // Empty data
        b"a".to_vec(), // Single byte
        b"short".to_vec(), // Short data
        b"medium length data".to_vec(), // Medium data
        vec![0u8; 1024], // Large data (1KB)
        vec![0u8; 10240], // Very large data (10KB)
    ];
    
    for test_data in test_cases {
        // Encrypt
        let encrypted = data_encryption.encrypt_data(&test_data, &recipient_public).unwrap();
        test_utils::assert_valid_encrypted_data(&encrypted);
        
        // Decrypt
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}

#[tokio::test]
async fn test_encryption_randomness() {
    let data_encryption = test_utils::create_test_data_encryption();
    let test_data = b"same data encrypted multiple times";
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    let mut encrypted_results = HashSet::new();
    
    // Encrypt same data multiple times
    for _ in 0..10 {
        let encrypted = data_encryption.encrypt_data(test_data, &recipient_public).unwrap();
        assert!(encrypted_results.insert(encrypted), "Encrypted data should be unique");
    }
    
    assert_eq!(encrypted_results.len(), 10, "All encrypted results should be different");
}

#[tokio::test]
async fn test_encryption_consistency() {
    let data_encryption = test_utils::create_test_data_encryption();
    let test_data = b"data for consistency test";
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Encrypt same data multiple times
    let encrypted1 = data_encryption.encrypt_data(test_data, &recipient_public).unwrap();
    let encrypted2 = data_encryption.encrypt_data(test_data, &recipient_public).unwrap();
    
    // Should be different (due to randomness)
    assert_ne!(encrypted1, encrypted2);
    
    // But both should decrypt to same data
    let decrypted1 = data_encryption.decrypt_data(&encrypted1).unwrap();
    let decrypted2 = data_encryption.decrypt_data(&encrypted2).unwrap();
    
    assert_eq!(decrypted1, decrypted2);
    assert_eq!(test_data.to_vec(), decrypted1);
}

#[tokio::test]
async fn test_encryption_with_different_recipients() {
    let data_encryption = test_utils::create_test_data_encryption();
    let test_data = b"data for different recipients";
    
    let recipients = vec![
        vec![1, 2, 3, 4, 5],
        vec![6, 7, 8, 9, 10],
        vec![11, 12, 13, 14, 15],
    ];
    
    let mut encrypted_results = Vec::new();
    
    for recipient in &recipients {
        let encrypted = data_encryption.encrypt_data(test_data, recipient).unwrap();
        encrypted_results.push(encrypted);
    }
    
    // All should be different
    for i in 0..encrypted_results.len() {
        for j in (i + 1)..encrypted_results.len() {
            assert_ne!(encrypted_results[i], encrypted_results[j]);
        }
    }
    
    // All should decrypt to same data
    for encrypted in &encrypted_results {
        let decrypted = data_encryption.decrypt_data(encrypted).unwrap();
        assert_eq!(test_data.to_vec(), decrypted);
    }
}

#[tokio::test]
async fn test_encryption_edge_cases() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with empty data
    let empty_data = b"";
    let encrypted = data_encryption.encrypt_data(empty_data, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(empty_data.to_vec(), decrypted);
    
    // Test with single byte
    let single_byte = b"a";
    let encrypted = data_encryption.encrypt_data(single_byte, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(single_byte.to_vec(), decrypted);
    
    // Test with all zeros
    let zeros = vec![0u8; 100];
    let encrypted = data_encryption.encrypt_data(&zeros, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(zeros, decrypted);
    
    // Test with all ones
    let ones = vec![1u8; 100];
    let encrypted = data_encryption.encrypt_data(&ones, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(ones, decrypted);
}

#[tokio::test]
async fn test_encryption_large_data() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with large data (1MB)
    let large_data = vec![0u8; 1024 * 1024];
    let encrypted = data_encryption.encrypt_data(&large_data, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(large_data, decrypted);
}

#[tokio::test]
async fn test_encryption_binary_data() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with binary data (all possible byte values)
    let binary_data: Vec<u8> = (0..256).map(|i| i as u8).collect();
    let encrypted = data_encryption.encrypt_data(&binary_data, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(binary_data, decrypted);
}

#[tokio::test]
async fn test_encryption_unicode_data() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with Unicode data
    let unicode_data = "Hello, 世界! 🌍".as_bytes();
    let encrypted = data_encryption.encrypt_data(unicode_data, &recipient_public).unwrap();
    let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
    assert_eq!(unicode_data.to_vec(), decrypted);
}

#[tokio::test]
async fn test_encryption_error_handling() {
    let data_encryption = test_utils::create_test_data_encryption();
    
    // Test with invalid encrypted data
    let invalid_encrypted = vec![1, 2, 3, 4, 5];
    let result = data_encryption.decrypt_data(&invalid_encrypted);
    assert!(result.is_err());
    
    // Test with empty encrypted data
    let empty_encrypted = vec![];
    let result = data_encryption.decrypt_data(&empty_encrypted);
    assert!(result.is_err());
}

#[tokio::test]
async fn test_concurrent_encryption() {
    let data_encryption = std::sync::Arc::new(test_utils::create_test_data_encryption());
    let recipient_public = vec![1, 2, 3, 4, 5];
    let mut handles = vec![];
    
    // Encrypt different data concurrently
    for i in 0..10 {
        let enc = Arc::clone(&data_encryption);
        let recipient = recipient_public.clone();
        let handle = tokio::spawn(async move {
            let test_data = format!("test data {}", i).into_bytes();
            let encrypted = enc.encrypt_data(&test_data, &recipient).unwrap();
            let decrypted = enc.decrypt_data(&encrypted).unwrap();
            assert_eq!(test_data, decrypted);
            encrypted
        });
        handles.push(handle);
    }
    
    // Wait for all to complete
    let results: Vec<_> = futures::future::join_all(handles).await;
    
    // All should succeed
    for result in results {
        let encrypted = result.unwrap();
        assert!(!encrypted.is_empty());
    }
}

#[tokio::test]
async fn test_encryption_performance() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    let test_data = vec![0u8; 1024]; // 1KB data
    
    let start = std::time::Instant::now();
    
    // Encrypt and decrypt multiple times
    for _ in 0..100 {
        let encrypted = data_encryption.encrypt_data(&test_data, &recipient_public).unwrap();
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
    
    let duration = start.elapsed();
    println!("Encryption/decryption performance: {:?} for 100 operations", duration);
    
    // Should complete in reasonable time (less than 1 second)
    assert!(duration.as_secs() < 1);
}

#[tokio::test]
async fn test_envelope_serialization() {
    let envelope = Envelope {
        nonce: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
        ephemeral_sender_public: [0x04; 65],
        encrypted_message: vec![1, 2, 3, 4, 5],
    };
    
    // Test serialization
    let serialized = borsh::to_vec(&envelope).unwrap();
    assert!(!serialized.is_empty());
    
    // Test deserialization
    let deserialized: Envelope = borsh::from_slice(&serialized).unwrap();
    assert_eq!(envelope.nonce, deserialized.nonce);
    assert_eq!(envelope.ephemeral_sender_public, deserialized.ephemeral_sender_public);
    assert_eq!(envelope.encrypted_message, deserialized.encrypted_message);
}

#[tokio::test]
async fn test_encryption_with_different_quorum_keys() {
    let quorum_key1 = P256Pair::generate().unwrap();
    let quorum_key2 = P256Pair::generate().unwrap();
    
    let data_encryption1 = DataEncryption::new(quorum_key1);
    let data_encryption2 = DataEncryption::new(quorum_key2);
    
    let test_data = b"data for different quorum keys";
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Encrypt with first key
    let encrypted1 = data_encryption1.encrypt_data(test_data, &recipient_public).unwrap();
    
    // Encrypt with second key
    let encrypted2 = data_encryption2.encrypt_data(test_data, &recipient_public).unwrap();
    
    // Should be different
    assert_ne!(encrypted1, encrypted2);
    
    // Each should decrypt with its own key
    let decrypted1 = data_encryption1.decrypt_data(&encrypted1).unwrap();
    let decrypted2 = data_encryption2.decrypt_data(&encrypted2).unwrap();
    
    assert_eq!(test_data.to_vec(), decrypted1);
    assert_eq!(test_data.to_vec(), decrypted2);
    
    // Cross-decryption should fail
    assert!(data_encryption1.decrypt_data(&encrypted2).is_err());
    assert!(data_encryption2.decrypt_data(&encrypted1).is_err());
}

#[tokio::test]
async fn test_encryption_memory_usage() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with various data sizes to check memory usage
    let sizes = vec![1024, 10240, 102400, 1024000]; // 1KB, 10KB, 100KB, 1MB
    
    for size in sizes {
        let test_data = vec![0u8; size];
        let encrypted = data_encryption.encrypt_data(&test_data, &recipient_public).unwrap();
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}

#[tokio::test]
async fn test_encryption_stress_test() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Stress test with many operations
    for i in 0..1000 {
        let test_data = format!("stress test data {}", i).into_bytes();
        let encrypted = data_encryption.encrypt_data(&test_data, &recipient_public).unwrap();
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}

#[tokio::test]
async fn test_encryption_random_data() {
    let data_encryption = test_utils::create_test_data_encryption();
    let recipient_public = vec![1, 2, 3, 4, 5];
    
    // Test with random data
    use rand::Rng;
    let mut rng = rand::thread_rng();
    
    for _ in 0..100 {
        let size = rng.gen_range(1..=1024);
        let test_data: Vec<u8> = (0..size).map(|_| rng.gen()).collect();
        
        let encrypted = data_encryption.encrypt_data(&test_data, &recipient_public).unwrap();
        let decrypted = data_encryption.decrypt_data(&encrypted).unwrap();
        assert_eq!(test_data, decrypted);
    }
}
