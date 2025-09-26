//! Member Decryption Tool
//!
//! This tool allows individual members to decrypt their encrypted shares
//! using their private keys. It simulates the member-side decryption process.

use anyhow::Result;
use p256::ecdsa::SigningKey;
use serde::{Deserialize, Serialize};
use sha2::Digest;
use std::fs;
use std::path::Path;

use renclave_shared::DecryptedShare;

#[derive(Debug, Serialize, Deserialize)]
pub struct MemberDecryptionResult {
    pub member_alias: String,
    pub decrypted_share: Vec<u8>,
    pub share_hash_verified: bool,
    pub success: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DecryptionBatchResult {
    pub results: Vec<MemberDecryptionResult>,
    pub successful_decryptions: usize,
    pub total_attempts: usize,
}

/// Decrypt a single share for a member
pub fn decrypt_member_share(
    member_alias: &str,
    private_key_path: &str,
    encrypted_share: &[u8],
    expected_hash: &[u8],
) -> Result<MemberDecryptionResult> {
    println!("🔓 Decrypting share for member: {}", member_alias);

    // Load private key
    let private_key_hex = fs::read_to_string(private_key_path)?;
    let private_key_bytes = hex::decode(private_key_hex.trim())?;
    let private_key_array: [u8; 32] = private_key_bytes.try_into().unwrap();
    let private_key = SigningKey::from_bytes(&private_key_array)
        .map_err(|e| anyhow::anyhow!("Failed to create signing key: {}", e))?;

    // For this simulation, we'll create a mock decryption
    // In a real implementation, this would use ECIES decryption
    let decrypted_share = simulate_decryption(encrypted_share, &private_key)?;

    // Verify share integrity using SHA-512
    let computed_hash = sha2::Sha512::digest(&decrypted_share);
    let share_hash_verified = computed_hash.as_slice() == expected_hash;

    if share_hash_verified {
        println!("✅ Share decrypted and verified successfully");
    } else {
        println!("❌ Share decryption failed - hash verification failed");
    }

    Ok(MemberDecryptionResult {
        member_alias: member_alias.to_string(),
        decrypted_share,
        share_hash_verified,
        success: share_hash_verified,
    })
}

/// Simulate ECIES decryption (in real implementation, use proper ECIES)
fn simulate_decryption(encrypted_share: &[u8], private_key: &SigningKey) -> Result<Vec<u8>> {
    // This is a simulation - in reality, you would use proper ECIES decryption
    // For now, we'll generate a mock decrypted share of the correct length
    let mut decrypted = vec![0u8; 32]; // Standard share size

    // Use the private key to generate deterministic "decrypted" data
    let key_bytes = private_key.to_bytes();
    for (i, byte) in key_bytes.iter().enumerate() {
        if i < decrypted.len() {
            decrypted[i] = *byte;
        }
    }

    // Add some variation based on the encrypted data
    for (i, byte) in encrypted_share.iter().take(32).enumerate() {
        if i < decrypted.len() {
            decrypted[i] ^= byte;
        }
    }

    Ok(decrypted)
}

/// Decrypt shares for multiple members
pub fn decrypt_shares_batch(
    distribution_result_path: &str,
    member_keys_dir: &str,
) -> Result<DecryptionBatchResult> {
    println!("🔓 Starting batch decryption of shares");

    // Load distribution result
    let distribution_json = fs::read_to_string(distribution_result_path)?;
    let distribution: serde_json::Value = serde_json::from_str(&distribution_json)?;

    let distributed_shares = &distribution["distributed_shares"];
    let mut results = Vec::new();
    let mut successful_decryptions = 0;

    let shares_array = distributed_shares.as_array().unwrap();
    for share_data in shares_array {
        let member_alias = share_data["member_alias"].as_str().unwrap();

        // Handle both hex string and byte array formats
        let encrypted_share = if let Some(hex_str) = share_data["encrypted_share"].as_str() {
            hex::decode(hex_str)?
        } else if let Some(bytes) = share_data["encrypted_share"].as_array() {
            bytes
                .iter()
                .map(|v| v.as_u64().unwrap() as u8)
                .collect::<Vec<u8>>()
        } else {
            return Err(anyhow::anyhow!("Invalid encrypted_share format"));
        };

        let share_hash = if let Some(hex_str) = share_data["share_hash"].as_str() {
            hex::decode(hex_str)?
        } else if let Some(bytes) = share_data["share_hash"].as_array() {
            bytes
                .iter()
                .map(|v| v.as_u64().unwrap() as u8)
                .collect::<Vec<u8>>()
        } else {
            return Err(anyhow::anyhow!("Invalid share_hash format"));
        };

        let private_key_path = format!("{}/{}.secret", member_keys_dir, member_alias);

        if Path::new(&private_key_path).exists() {
            match decrypt_member_share(
                member_alias,
                &private_key_path,
                &encrypted_share,
                &share_hash,
            ) {
                Ok(result) => {
                    if result.success {
                        successful_decryptions += 1;
                    }
                    results.push(result);
                }
                Err(e) => {
                    println!("❌ Failed to decrypt share for {}: {}", member_alias, e);
                    results.push(MemberDecryptionResult {
                        member_alias: member_alias.to_string(),
                        decrypted_share: vec![],
                        share_hash_verified: false,
                        success: false,
                    });
                }
            }
        } else {
            println!("❌ Private key not found for member: {}", member_alias);
            results.push(MemberDecryptionResult {
                member_alias: member_alias.to_string(),
                decrypted_share: vec![],
                share_hash_verified: false,
                success: false,
            });
        }
    }

    let result = DecryptionBatchResult {
        results,
        successful_decryptions,
        total_attempts: distributed_shares.as_array().unwrap().len(),
    };

    println!("📊 Batch decryption completed:");
    println!(
        "   - Successful: {}/{}",
        result.successful_decryptions, result.total_attempts
    );

    Ok(result)
}

/// Create share injection request from decrypted shares
pub fn create_share_injection_request(
    decryption_result: &DecryptionBatchResult,
    _namespace_name: &str,
    _namespace_nonce: u64,
) -> Result<Vec<DecryptedShare>> {
    println!("📝 Creating share injection request");

    let mut shares = Vec::new();

    for result in &decryption_result.results {
        if result.success {
            shares.push(DecryptedShare {
                member_alias: result.member_alias.clone(),
                decrypted_share: result.decrypted_share.clone(),
            });
            println!("✅ Added share for member: {}", result.member_alias);
        } else {
            println!(
                "⚠️  Skipping failed decryption for member: {}",
                result.member_alias
            );
        }
    }

    println!("📊 Created {} valid shares for injection", shares.len());
    Ok(shares)
}

/// Save share injection request to file
pub fn save_share_injection_request(
    shares: &[DecryptedShare],
    namespace_name: &str,
    namespace_nonce: u64,
    output_path: &str,
) -> Result<()> {
    let request = serde_json::json!({
        "namespace_name": namespace_name,
        "namespace_nonce": namespace_nonce,
        "shares": shares
    });

    let json = serde_json::to_string_pretty(&request)?;
    fs::write(output_path, json)?;
    println!("💾 Share injection request saved to: {}", output_path);
    Ok(())
}

fn main() -> Result<()> {
    println!("🔓 Member Decryption Tool");
    println!("=========================");

    // Check required files
    let distribution_path = "share_distribution.json";
    let member_keys_dir = "member_keys";

    if !Path::new(distribution_path).exists() {
        println!("❌ Distribution file not found: {}", distribution_path);
        println!("   Please run share distributor first.");
        return Ok(());
    }

    if !Path::new(member_keys_dir).exists() {
        println!("❌ Member keys directory not found: {}", member_keys_dir);
        println!("   Please run share distributor first to generate member keys.");
        return Ok(());
    }

    // Perform batch decryption
    let decryption_result = decrypt_shares_batch(distribution_path, member_keys_dir)?;

    // Save decryption results
    let decryption_json = serde_json::to_string_pretty(&decryption_result)?;
    fs::write("member_decryption_results.json", decryption_json)?;
    println!("💾 Decryption results saved to: member_decryption_results.json");

    // Create share injection request
    let shares = create_share_injection_request(&decryption_result, "test-namespace", 1)?;

    if shares.len() >= 2 {
        // Save share injection request
        save_share_injection_request(&shares, "test-namespace", 1, "share_injection_request.json")?;

        println!("\n🎉 Member decryption completed successfully!");
        println!("📁 Files created:");
        println!("   - member_decryption_results.json (decryption details)");
        println!("   - share_injection_request.json (ready for TEE injection)");
        println!("\n📋 Next steps:");
        println!("   1. Send share_injection_request.json to TEE via Share Injection API");
        println!(
            "   2. TEE will reconstruct quorum key from {} decrypted shares",
            shares.len()
        );
    } else {
        println!(
            "❌ Not enough shares decrypted successfully (need at least 2, got {})",
            shares.len()
        );
        println!("   Please check member keys and try again.");
    }

    Ok(())
}
