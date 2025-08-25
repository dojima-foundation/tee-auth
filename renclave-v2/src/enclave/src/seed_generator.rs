use anyhow::{anyhow, Result};
use bip39::{Language, Mnemonic};
use bitcoin::bip32::{DerivationPath, Xpriv, Xpub};
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
            private_key: hex::encode(&child_key.private_key.secret_bytes()),
            public_key: hex::encode(&public_key.public_key.serialize()),
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
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_seed_generation() {
        let generator = SeedGenerator::new().await.unwrap();

        // Test all supported strengths
        for strength in [128, 160, 192, 224, 256] {
            let result = generator.generate_seed(strength, None).await.unwrap();
            assert_eq!(result.strength, strength);
            assert!(!result.phrase.is_empty());
            assert!(!result.entropy.is_empty());

            // Verify word count
            let expected_words = match strength {
                128 => 12,
                160 => 15,
                192 => 18,
                224 => 21,
                256 => 24,
                _ => panic!("Invalid strength"),
            };
            assert_eq!(result.word_count, expected_words);
        }
    }

    #[tokio::test]
    async fn test_seed_validation() {
        let generator = SeedGenerator::new().await.unwrap();

        // Generate a valid seed
        let result = generator.generate_seed(256, None).await.unwrap();

        // Validate it
        let is_valid = generator.validate_seed(&result.phrase).await.unwrap();
        assert!(is_valid);

        // Test invalid seed
        let is_valid = generator
            .validate_seed("invalid seed phrase")
            .await
            .unwrap();
        assert!(!is_valid);
    }

    #[tokio::test]
    async fn test_entropy_verification() {
        let generator = SeedGenerator::new().await.unwrap();

        // Generate seed
        let result = generator.generate_seed(256, None).await.unwrap();

        // Get entropy from mnemonic
        let entropy_bytes = hex::decode(&result.entropy).unwrap();
        let extracted_entropy = generator.get_entropy_from_mnemonic(&result.phrase).unwrap();

        assert_eq!(entropy_bytes, extracted_entropy);

        // Verify entropy matches mnemonic
        let matches = generator
            .verify_entropy(&entropy_bytes, &result.phrase)
            .await
            .unwrap();
        assert!(matches);
    }
}
