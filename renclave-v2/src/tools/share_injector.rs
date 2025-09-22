use p256::ecdsa::SigningKey;
use rand::{rngs::OsRng, Rng};
use serde_json::{json, Value};

/// Share Injector - Sends encrypted shares back to TEE to complete Genesis Boot
fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("🔐 Genesis Boot Share Injector");
    println!("==============================");
    
    // Generate the same key pairs that were used in the Genesis Boot request
    let mut key_pairs = Vec::new();
    let mut public_keys = Vec::new();
    
    for i in 1..=3 {
        let signing_key = SigningKey::random(&mut OsRng);
        let public_key = signing_key.verifying_key();
        let public_key_bytes = public_key.to_encoded_point(false).as_bytes().to_vec();
        
        key_pairs.push((signing_key, public_key));
        public_keys.push(public_key_bytes.clone());
        
        println!(
            "✅ Generated key pair {}: {} bytes",
            i,
            public_key_bytes.len()
        );
        println!("   Public key: {}", hex::encode(&public_key_bytes));
    }
    
    // Create mock encrypted shares (in real scenario, these would be the actual encrypted shares from Genesis Boot)
    let mock_encrypted_shares = create_mock_encrypted_shares(&public_keys)?;
    
    // Create share injection request
    let request = create_share_injection_request(&mock_encrypted_shares)?;
    
    // Save to JSON file
    let json_str = serde_json::to_string_pretty(&request)?;
    std::fs::write("share_injection_request.json", &json_str)?;
    
    println!("\n📋 Share Injection Request Created:");
    println!("===================================");
    println!("{}", json_str);
    
    println!("\n🚀 To inject shares, run:");
    println!("curl -X POST http://localhost:3000/enclave/inject-shares \\");
    println!("  -H \"Content-Type: application/json\" \\");
    println!("  -d @share_injection_request.json");
    
    Ok(())
}

fn create_mock_encrypted_shares(public_keys: &[Vec<u8>]) -> Result<Vec<Value>, Box<dyn std::error::Error>> {
    let mut shares = Vec::new();
    let mut rng = OsRng;
    
    for (i, pub_key) in public_keys.iter().enumerate() {
        // Create a mock encrypted share (32 bytes of random data)
        let encrypted_share: Vec<u8> = (0..32).map(|_| rng.gen()).collect();
        
        let share = json!({
            "member_id": format!("share_member_{}", i + 1),
            "encrypted_share": encrypted_share,
            "public_key": pub_key
        });
        
        shares.push(share);
        println!("🔒 Created mock encrypted share for member {}: {} bytes", i + 1, 32);
    }
    
    Ok(shares)
}

fn create_share_injection_request(shares: &[Value]) -> Result<Value, Box<dyn std::error::Error>> {
    let request = json!({
        "namespace_name": "test-namespace",
        "namespace_nonce": 123,
        "shares": shares
    });
    
    Ok(request)
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_mock_share_creation() {
        let mut public_keys = Vec::new();
        for _ in 0..3 {
            let signing_key = SigningKey::random(&mut OsRng);
            let public_key = signing_key.verifying_key();
            public_keys.push(public_key.to_encoded_point(false).as_bytes().to_vec());
        }
        
        let shares = create_mock_encrypted_shares(&public_keys).unwrap();
        assert_eq!(shares.len(), 3);
        
        for share in &shares {
            assert!(share["encrypted_share"].as_array().is_some());
            assert_eq!(share["encrypted_share"].as_array().unwrap().len(), 32);
        }
    }
    
    #[test]
    fn test_request_creation() {
        let mut public_keys = Vec::new();
        for _ in 0..3 {
            let signing_key = SigningKey::random(&mut OsRng);
            let public_key = signing_key.verifying_key();
            public_keys.push(public_key.to_encoded_point(false).as_bytes().to_vec());
        }
        
        let shares = create_mock_encrypted_shares(&public_keys).unwrap();
        let request = create_share_injection_request(&shares).unwrap();
        
        assert_eq!(request["namespace_name"], "test-namespace");
        assert_eq!(request["namespace_nonce"], 123);
        assert_eq!(request["shares"].as_array().unwrap().len(), 3);
    }
}
