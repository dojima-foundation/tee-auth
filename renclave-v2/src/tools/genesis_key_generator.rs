use p256::ecdsa::SigningKey;
use rand::rngs::OsRng;
use serde_json::{json, Value};

/// Generate P256 key pairs and create Genesis Boot API request
fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("🔑 Genesis Boot Key Generator");
    println!("=============================");

    // Generate 5 key pairs for testing (5-threshold scenario)
    let mut key_pairs = Vec::new();
    let mut public_keys = Vec::new();

    for i in 1..=5 {
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

    // Create Genesis Boot request
    let request = create_genesis_boot_request(&public_keys)?;

    // Save to JSON file
    let json_str = serde_json::to_string_pretty(&request)?;
    std::fs::write("genesis_boot_request.json", &json_str)?;

    println!("\n📋 Genesis Boot Request Created:");
    println!("================================");
    println!("{}", json_str);

    println!("\n🚀 To test the API, run:");
    println!("curl -X POST http://localhost:8080/enclave/genesis-boot \\");
    println!("  -H \"Content-Type: application/json\" \\");
    println!("  -d @genesis_boot_request.json");

    Ok(())
}

fn create_genesis_boot_request(
    public_keys: &[Vec<u8>],
) -> Result<Value, Box<dyn std::error::Error>> {
    if public_keys.len() < 5 {
        return Err("Need at least 5 public keys for 5-threshold scenario".into());
    }

    // Use all 5 keys for manifest members
    let manifest_members = vec![
        json!({
            "alias": "manifest_member_1",
            "pub_key": public_keys[0]
        }),
        json!({
            "alias": "manifest_member_2",
            "pub_key": public_keys[1]
        }),
        json!({
            "alias": "manifest_member_3",
            "pub_key": public_keys[2]
        }),
        json!({
            "alias": "manifest_member_4",
            "pub_key": public_keys[3]
        }),
        json!({
            "alias": "manifest_member_5",
            "pub_key": public_keys[4]
        }),
    ];

    // Use all 5 keys for share members
    let share_members = vec![
        json!({
            "alias": "share_member_1",
            "pub_key": public_keys[0]
        }),
        json!({
            "alias": "share_member_2",
            "pub_key": public_keys[1]
        }),
        json!({
            "alias": "share_member_3",
            "pub_key": public_keys[2]
        }),
        json!({
            "alias": "share_member_4",
            "pub_key": public_keys[3]
        }),
        json!({
            "alias": "share_member_5",
            "pub_key": public_keys[4]
        }),
    ];

    let request = json!({
        "namespace_name": "test-namespace",
        "namespace_nonce": 123,
        "manifest_members": manifest_members,
        "manifest_threshold": 5,
        "share_members": share_members,
        "share_threshold": 5,
        "pivot_hash": [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        "pivot_args": [],
        "dr_key": null
    });

    Ok(request)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_key_generation() {
        let signing_key = SigningKey::random(&mut OsRng);
        let public_key = signing_key.verifying_key();
        let public_key_bytes = public_key.to_encoded_point(false).as_bytes().to_vec();

        assert_eq!(public_key_bytes.len(), 65);
        assert_eq!(public_key_bytes[0], 0x04); // Uncompressed format
    }

    #[test]
    fn test_request_creation() {
        let mut public_keys = Vec::new();
        for _ in 0..3 {
            let signing_key = SigningKey::random(&mut OsRng);
            let public_key = signing_key.verifying_key();
            public_keys.push(public_key.to_encoded_point(false).as_bytes().to_vec());
        }

        let request = create_genesis_boot_request(&public_keys).unwrap();
        assert!(request["namespace_name"].as_str().is_some());
        assert!(request["manifest_members"].as_array().is_some());
        assert!(request["share_members"].as_array().is_some());
    }
}
