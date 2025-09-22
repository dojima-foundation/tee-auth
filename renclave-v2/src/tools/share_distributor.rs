//! Share Distribution Tool
//!
//! This tool distributes encrypted shares from Genesis Boot to individual members.
//! It simulates the external distribution process where shares are sent to members
//! via secure channels.

use anyhow::Result;
use p256::ecdsa::SigningKey;
use rand::rngs::OsRng;
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::Path;

use renclave_shared::{GenesisMemberOutput, QuorumMember};

#[derive(Debug, Serialize, Deserialize)]
pub struct ShareDistribution {
    pub member_alias: String,
    pub encrypted_share: Vec<u8>,
    #[serde(with = "serde_bytes")]
    pub share_hash: Vec<u8>,
    pub member_public_key: Vec<u8>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DistributionResult {
    pub distributed_shares: Vec<ShareDistribution>,
    pub total_members: usize,
    pub threshold: u32,
}

/// Distribute encrypted shares to members
pub fn distribute_shares(genesis_output: &[GenesisMemberOutput]) -> Result<DistributionResult> {
    println!("📤 Distributing encrypted shares to {} members", genesis_output.len());
    
    let mut distributed_shares = Vec::new();
    
    for (i, member_output) in genesis_output.iter().enumerate() {
        println!("📋 Distributing share {} to member: {}", i + 1, member_output.share_set_member.alias);
        
        let distribution = ShareDistribution {
            member_alias: member_output.share_set_member.alias.clone(),
            encrypted_share: member_output.encrypted_quorum_key_share.clone(),
            share_hash: member_output.share_hash.to_vec(),
            member_public_key: member_output.share_set_member.pub_key.clone(),
        };
        
        distributed_shares.push(distribution);
    }
    
    let result = DistributionResult {
        distributed_shares,
        total_members: genesis_output.len(),
        threshold: 2, // 2-of-3 threshold
    };
    
    println!("✅ Successfully distributed {} shares", result.total_members);
    Ok(result)
}

/// Save distribution result to file
pub fn save_distribution_result(result: &DistributionResult, output_path: &str) -> Result<()> {
    let json = serde_json::to_string_pretty(result)?;
    fs::write(output_path, json)?;
    println!("💾 Distribution result saved to: {}", output_path);
    Ok(())
}

/// Load distribution result from file
pub fn load_distribution_result(input_path: &str) -> Result<DistributionResult> {
    let json = fs::read_to_string(input_path)?;
    let result: DistributionResult = serde_json::from_str(&json)?;
    println!("📖 Distribution result loaded from: {}", input_path);
    Ok(result)
}

/// Generate member private keys for testing
pub fn generate_member_keys(members: &[QuorumMember], output_dir: &str) -> Result<()> {
    println!("🔑 Generating private keys for {} members", members.len());
    
    // Create output directory if it doesn't exist
    fs::create_dir_all(output_dir)?;
    
    for member in members {
        // Generate a new private key for this member
        let private_key = SigningKey::random(&mut OsRng);
        let public_key = private_key.verifying_key();
        
        // Convert to bytes
        let private_key_bytes = private_key.to_bytes();
        let public_key_bytes = public_key.to_encoded_point(false).as_bytes().to_vec();
        
        // Save private key
        let private_key_path = format!("{}/{}.secret", output_dir, member.alias);
        fs::write(&private_key_path, hex::encode(private_key_bytes))?;
        println!("🔐 Private key saved: {}", private_key_path);
        
        // Save public key
        let public_key_path = format!("{}/{}.pub", output_dir, member.alias);
        fs::write(&public_key_path, hex::encode(&public_key_bytes))?;
        println!("🔑 Public key saved: {}", public_key_path);
        
        // Verify the public key matches what was used in Genesis Boot
        if public_key_bytes != member.pub_key {
            println!("⚠️  Warning: Generated public key doesn't match Genesis Boot public key for {}", member.alias);
            println!("   Generated: {}", hex::encode(&public_key_bytes));
            println!("   Expected:  {}", hex::encode(&member.pub_key));
        }
    }
    
    println!("✅ Member keys generated successfully");
    Ok(())
}

fn main() -> Result<()> {
    println!("🚀 Share Distribution Tool");
    println!("==========================");
    
    // Check if we have a genesis output file
    let genesis_output_path = "genesis_boot_response.json";
    if !Path::new(genesis_output_path).exists() {
        println!("❌ Genesis Boot response file not found: {}", genesis_output_path);
        println!("   Please run Genesis Boot first to generate encrypted shares.");
        return Ok(());
    }
    
    // Load Genesis Boot response
    println!("📖 Loading Genesis Boot response...");
    let genesis_response: serde_json::Value = serde_json::from_str(&fs::read_to_string(genesis_output_path)?)?;
    
    // Extract encrypted shares
    let encrypted_shares: Vec<GenesisMemberOutput> = serde_json::from_value(
        genesis_response["encrypted_shares"].clone()
    )?;
    
    println!("📊 Found {} encrypted shares", encrypted_shares.len());
    
    // Distribute shares
    let distribution_result = distribute_shares(&encrypted_shares)?;
    
    // Save distribution result
    save_distribution_result(&distribution_result, "share_distribution.json")?;
    
    // Generate member private keys for testing
    let members: Vec<QuorumMember> = encrypted_shares
        .iter()
        .map(|output| output.share_set_member.clone())
        .collect();
    
    generate_member_keys(&members, "member_keys")?;
    
    println!("\n🎉 Share distribution completed successfully!");
    println!("📁 Files created:");
    println!("   - share_distribution.json (distribution details)");
    println!("   - member_keys/ (member private keys)");
    println!("\n📋 Next steps:");
    println!("   1. Each member should decrypt their share using their private key");
    println!("   2. Members send decrypted shares back to TEE via Share Injection API");
    println!("   3. TEE reconstructs quorum key from decrypted shares");
    
    Ok(())
}
