use anyhow::{anyhow, Result};
use bip39::{Language, Mnemonic};
use bitcoin::bip32::{DerivationPath, Xpriv, Xpub};
use hex;
use log::{debug, info, warn};
use rand::{RngCore, SeedableRng};
use secp256k1::Secp256k1;
use std::str::FromStr;
use std::sync::Arc;
use tokio::sync::Mutex;

/// Secure seed phrase generator for Nitro Enclave
pub struct SeedGenerator {
    rng: Arc<Mutex<rand::rngs::StdRng>>,
}

#[derive(Debug, Clone)]
pub struct SeedResult {
    pub phrase: String,
    pub entropy: String,
    pub strength: u32,
    pub word_count: usize,
}

#[derive(Debug, Clone)]
pub struct KeyDerivationResult {
    pub private_key: String,
    pub public_key: String,
    pub address: String,
}

#[derive(Debug, Clone)]
pub struct AddressDerivationResult {
    pub address: String,
}

impl SeedGenerator {
    /// Create new seed generator with secure entropy
    pub async fn new() -> Result<Self> {
        info!("🌱 Initializing secure seed generator in Nitro Enclave");

        // Initialize with hardware entropy
        let rng = rand::rngs::StdRng::from_entropy();
        debug!("🔐 Initialized RNG with hardware entropy");

        Ok(Self {
            rng: Arc::new(Mutex::new(rng)),
        })
    }

    /// Generate secure seed phrase
    pub async fn generate_seed(
        &self,
        strength: u32,
        passphrase: Option<&str>,
    ) -> Result<SeedResult> {
        info!(
            "🔑 Generating secure seed phrase (strength: {} bits)",
            strength
        );

        // Validate strength
        let word_count = self.validate_strength(strength)?;
        debug!("📊 Word count for strength {}: {}", strength, word_count);

        // Generate entropy
        let entropy = self.generate_entropy(strength).await?;
        debug!("🎲 Generated {} bytes of entropy", entropy.len());

        // Create BIP39 mnemonic
        let mnemonic = Mnemonic::from_entropy_in(Language::English, &entropy)
            .map_err(|e| anyhow!("Failed to create mnemonic: {}", e))?;

        let phrase = mnemonic.to_string();
        debug!(
            "📝 Generated mnemonic with {} words",
            phrase.split_whitespace().count()
        );

        // Validate word count
        let actual_word_count = phrase.split_whitespace().count();
        if actual_word_count != word_count {
            return Err(anyhow!(
                "Word count mismatch: expected {}, got {}",
                word_count,
                actual_word_count
            ));
        }

        // Apply passphrase if provided
        let final_phrase = if let Some(pass) = passphrase {
            info!("🔐 Applying passphrase to seed phrase");
            format!("{} {}", phrase, pass)
        } else {
            phrase
        };

        let result = SeedResult {
            phrase: final_phrase,
            entropy: hex::encode(&entropy),
            strength,
            word_count: actual_word_count,
        };

        info!("✅ Seed phrase generated successfully");
        info!(
            "📊 Strength: {} bits, Words: {}",
            result.strength, result.word_count
        );

        Ok(result)
    }

    /// Validate existing seed phrase
    pub async fn validate_seed(&self, seed_phrase: &str) -> Result<bool> {
        info!("🔍 Validating seed phrase");

        if seed_phrase.trim().is_empty() {
            warn!("⚠️  Empty seed phrase provided");
            return Ok(false);
        }

        // Split to handle potential passphrase
        let words: Vec<&str> = seed_phrase.split_whitespace().collect();
        if words.is_empty() {
            warn!("⚠️  No words found in seed phrase");
            return Ok(false);
        }

        debug!("🔍 Validating {} words", words.len());

        // Try to parse as BIP39 mnemonic
        match Mnemonic::parse_in_normalized(Language::English, seed_phrase) {
            Ok(_) => {
                info!("✅ Seed phrase is valid BIP39 mnemonic");
                Ok(true)
            }
            Err(e) => {
                debug!("❌ Invalid seed phrase: {}", e);

                // If it fails, try without the last word (might be passphrase)
                if words.len() > 12 {
                    let without_last = words[..words.len() - 1].join(" ");
                    match Mnemonic::parse_in_normalized(Language::English, &without_last) {
                        Ok(_) => {
                            info!("✅ Seed phrase is valid BIP39 mnemonic (with passphrase)");
                            Ok(true)
                        }
                        Err(_) => {
                            info!("❌ Seed phrase is not a valid BIP39 mnemonic");
                            Ok(false)
                        }
                    }
                } else {
                    info!("❌ Seed phrase is not a valid BIP39 mnemonic");
                    Ok(false)
                }
            }
        }
    }

    /// Validate seed strength and return expected word count
    fn validate_strength(&self, strength: u32) -> Result<usize> {
        let word_count = match strength {
            128 => 12,
            160 => 15,
            192 => 18,
            224 => 21,
            256 => 24,
            _ => {
                return Err(anyhow!(
                    "Invalid strength: {}. Must be 128, 160, 192, 224, or 256 bits",
                    strength
                ))
            }
        };

        debug!("✅ Strength {} validated -> {} words", strength, word_count);
        Ok(word_count)
    }

    /// Generate cryptographically secure entropy
    async fn generate_entropy(&self, strength: u32) -> Result<Vec<u8>> {
        let entropy_bytes = (strength / 8) as usize;
        let mut entropy = vec![0u8; entropy_bytes];

        debug!(
            "🎲 Generating {} bytes of entropy for {} bits",
            entropy_bytes, strength
        );

        // Use secure RNG to generate entropy
        {
            let mut rng = self.rng.lock().await;
            rng.fill_bytes(&mut entropy);
        }

        // Verify entropy is not all zeros (extremely unlikely but good practice)
        if entropy.iter().all(|&b| b == 0) {
            warn!("⚠️  Generated entropy is all zeros, regenerating...");
            let mut rng = self.rng.lock().await;
            rng.fill_bytes(&mut entropy);
        }

        debug!("✅ Generated {} bytes of secure entropy", entropy.len());
        Ok(entropy)
    }

    /// Get entropy from existing mnemonic (for testing/verification)
    #[allow(dead_code)]
    pub fn get_entropy_from_mnemonic(&self, mnemonic: &str) -> Result<Vec<u8>> {
        let mnemonic_obj = Mnemonic::parse_in_normalized(Language::English, mnemonic)
            .map_err(|e| anyhow!("Invalid mnemonic: {}", e))?;

        Ok(mnemonic_obj.to_entropy().to_vec())
    }

    /// Derive seed from mnemonic and passphrase
    pub async fn derive_seed(&self, mnemonic: &str, passphrase: Option<&str>) -> Result<Vec<u8>> {
        info!("🌱 Deriving seed from mnemonic");

        let mnemonic_obj = Mnemonic::parse_in_normalized(Language::English, mnemonic)
            .map_err(|e| anyhow!("Invalid mnemonic: {}", e))?;

        let passphrase = passphrase.unwrap_or("");

        // Derive 64-byte seed using PBKDF2
        let seed = mnemonic_obj.to_seed(passphrase);

        info!("✅ Derived {}-byte seed from mnemonic", seed.len());
        Ok(seed.to_vec())
    }

    /// Verify entropy matches mnemonic
    #[allow(dead_code)]
    pub async fn verify_entropy(&self, entropy: &[u8], mnemonic: &str) -> Result<bool> {
        let expected_mnemonic = Mnemonic::from_entropy_in(Language::English, entropy)
            .map_err(|e| anyhow!("Invalid entropy: {}", e))?;

        Ok(expected_mnemonic.to_string() == mnemonic)
    }

    /// Derive key from seed phrase
    pub async fn derive_key(
        &self,
        seed_phrase: &str,
        path: &str,
        curve: &str,
    ) -> Result<KeyDerivationResult> {
        info!("🔑 Deriving key (path: {}, curve: {})", path, curve);

        // Parse derivation path
        let derivation_path = DerivationPath::from_str(path)
            .map_err(|e| anyhow!("Invalid derivation path: {}", e))?;

        // Derive seed from mnemonic
        let seed = self.derive_seed(seed_phrase, None).await?;

        // Create extended private key
        let secp = Secp256k1::new();
        let master_key = Xpriv::new_master(bitcoin::Network::Bitcoin, &seed)
            .map_err(|e| anyhow!("Failed to create master key: {}", e))?;

        // Derive child key
        let child_key = master_key
            .derive_priv(&secp, &derivation_path)
            .map_err(|e| anyhow!("Failed to derive child key: {}", e))?;

        // Get public key
        let public_key = Xpub::from_priv(&secp, &child_key);

        // Generate address (simplified - in production you'd use proper address generation)
        let address = format!(
            "0x{}",
            hex::encode(&public_key.public_key.serialize()[..20])
        );

        let result = KeyDerivationResult {
            private_key: hex::encode(child_key.private_key.secret_bytes()),
            public_key: hex::encode(public_key.public_key.serialize()),
            address,
        };

        info!("✅ Key derivation successful");
        Ok(result)
    }

    /// Derive address from seed phrase
    pub async fn derive_address(
        &self,
        seed_phrase: &str,
        path: &str,
        curve: &str,
    ) -> Result<AddressDerivationResult> {
        info!("📍 Deriving address (path: {}, curve: {})", path, curve);

        // For now, we'll derive the key and return just the address
        // In production, you might want to optimize this to only derive the public key
        let key_result = self.derive_key(seed_phrase, path, curve).await?;

        let result = AddressDerivationResult {
            address: key_result.address,
        };

        info!("✅ Address derivation successful");
        Ok(result)
    }

    /// Derive entropy from a seed phrase using BIP39
    pub async fn derive_entropy_from_seed(&self, seed_phrase: &str) -> Result<String> {
        info!("🔍 Deriving entropy from seed phrase");
        
        // Parse the mnemonic
        let mnemonic = Mnemonic::parse(seed_phrase)
            .map_err(|e| anyhow::anyhow!("Failed to parse mnemonic: {}", e))?;
        
        // Get the entropy from the mnemonic
        let entropy = mnemonic.to_entropy();
        let entropy_hex = hex::encode(entropy);
        
        info!("✅ Entropy derived successfully: {}", entropy_hex);
        Ok(entropy_hex)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::runtime::Runtime;

    fn create_test_runtime() -> Runtime {
        tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()
            .unwrap()
    }

    #[test]
    fn test_seed_generator_new() {
        let runtime = create_test_runtime();
        let result = runtime.block_on(SeedGenerator::new());

        assert!(result.is_ok());
        let generator = result.unwrap();
        assert!(generator.rng.try_lock().is_ok());
    }

    #[test]
    fn test_validate_strength() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        // Test valid strengths
        assert_eq!(generator.validate_strength(128).unwrap(), 12);
        assert_eq!(generator.validate_strength(160).unwrap(), 15);
        assert_eq!(generator.validate_strength(192).unwrap(), 18);
        assert_eq!(generator.validate_strength(224).unwrap(), 21);
        assert_eq!(generator.validate_strength(256).unwrap(), 24);

        // Test invalid strengths
        assert!(generator.validate_strength(100).is_err());
        assert!(generator.validate_strength(300).is_err());
        assert!(generator.validate_strength(0).is_err());
    }

    #[test]
    fn test_generate_seed_128_bits() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime.block_on(generator.generate_seed(128, None));
        assert!(result.is_ok());

        let seed_result = result.unwrap();
        assert_eq!(seed_result.strength, 128);
        assert_eq!(seed_result.word_count, 12);
        assert_eq!(seed_result.phrase.split_whitespace().count(), 12);
        assert!(!seed_result.entropy.is_empty());
    }

    #[test]
    fn test_generate_seed_256_bits() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime.block_on(generator.generate_seed(256, None));
        assert!(result.is_ok());

        let seed_result = result.unwrap();
        assert_eq!(seed_result.strength, 256);
        assert_eq!(seed_result.word_count, 24);
        assert_eq!(seed_result.phrase.split_whitespace().count(), 24);
        assert!(!seed_result.entropy.is_empty());
    }

    #[test]
    fn test_generate_seed_with_passphrase() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let passphrase = "test123";
        let result = runtime.block_on(generator.generate_seed(192, Some(passphrase)));
        assert!(result.is_ok());

        let seed_result = result.unwrap();
        assert_eq!(seed_result.strength, 192);
        assert_eq!(seed_result.word_count, 18);
        assert!(seed_result.phrase.ends_with(passphrase));
    }

    #[test]
    fn test_generate_seed_invalid_strength() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime.block_on(generator.generate_seed(100, None));
        assert!(result.is_err());

        let error = result.unwrap_err();
        assert!(error.to_string().contains("Invalid strength"));
    }

    #[test]
    fn test_validate_seed_valid() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        // Valid BIP39 seed phrase (12 words)
        let valid_seed = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";
        let result = runtime.block_on(generator.validate_seed(valid_seed));
        assert!(result.is_ok());
        assert!(result.unwrap());
    }

    #[test]
    fn test_validate_seed_valid_24_words() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        // Generate a valid 24-word seed first
        let seed_result = runtime
            .block_on(generator.generate_seed(256, None))
            .unwrap();
        let result = runtime.block_on(generator.validate_seed(&seed_result.phrase));
        assert!(result.is_ok());
        assert!(result.unwrap());
    }

    #[test]
    fn test_validate_seed_invalid() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let invalid_seed = "invalid seed phrase here";
        let result = runtime.block_on(generator.validate_seed(invalid_seed));
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[test]
    fn test_validate_seed_empty() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime.block_on(generator.validate_seed(""));
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[test]
    fn test_validate_seed_whitespace_only() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime.block_on(generator.validate_seed("   \n\t  "));
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[test]
    fn test_generate_entropy() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        // Test different entropy sizes
        let entropy_128 = runtime.block_on(generator.generate_entropy(128)).unwrap();
        assert_eq!(entropy_128.len(), 16); // 128 bits = 16 bytes

        let entropy_256 = runtime.block_on(generator.generate_entropy(256)).unwrap();
        assert_eq!(entropy_256.len(), 32); // 256 bits = 32 bytes
    }

    #[test]
    fn test_entropy_not_all_zeros() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        // Generate multiple entropy samples to ensure they're not all zeros
        let mut has_non_zero = false;
        for _ in 0..10 {
            let entropy = runtime.block_on(generator.generate_entropy(256)).unwrap();
            if entropy.iter().any(|&b| b != 0) {
                has_non_zero = true;
                break;
            }
        }
        assert!(has_non_zero, "All entropy samples were zero");
    }

    #[test]
    fn test_seed_phrase_uniqueness() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let mut phrases = std::collections::HashSet::new();

        // Generate multiple seeds and ensure they're unique
        for _ in 0..5 {
            let result = runtime
                .block_on(generator.generate_seed(256, None))
                .unwrap();
            assert!(
                phrases.insert(result.phrase),
                "Duplicate seed phrase generated"
            );
        }
    }

    #[test]
    fn test_entropy_hex_encoding() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let result = runtime
            .block_on(generator.generate_seed(128, None))
            .unwrap();

        // Verify entropy is valid hex
        assert!(result.entropy.len() == 32); // 128 bits = 16 bytes = 32 hex chars
        assert!(result.entropy.chars().all(|c| c.is_ascii_hexdigit()));
    }

    #[test]
    fn test_word_count_consistency() {
        let runtime = create_test_runtime();
        let generator = runtime.block_on(SeedGenerator::new()).unwrap();

        let strengths = vec![128, 160, 192, 224, 256];
        let expected_words = vec![12, 15, 18, 21, 24];

        for (strength, expected) in strengths.iter().zip(expected_words.iter()) {
            let result = runtime
                .block_on(generator.generate_seed(*strength, None))
                .unwrap();
            assert_eq!(result.word_count, *expected);
            assert_eq!(result.phrase.split_whitespace().count(), *expected);
        }
    }
}
