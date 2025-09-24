use serde_json;
use std::env;

fn main() {
    println!("💉 Creating dynamic share injection request (matching QoS)...");

    // Get the Genesis Boot response from environment or use default
    let genesis_response = env::var("GENESIS_RESPONSE")
        .unwrap_or_else(|_| {
            println!("❌ GENESIS_RESPONSE environment variable not set");
            println!("💡 Run: export GENESIS_RESPONSE='$(curl -X POST http://localhost:3000/enclave/genesis-boot -H \"Content-Type: application/json\" -d @/tmp/genesis_request.json)'");
            std::process::exit(1);
        });

    // Parse the Genesis Boot response
    let genesis_data: serde_json::Value =
        serde_json::from_str(&genesis_response).expect("Failed to parse Genesis Boot response");

    // Extract required fields dynamically
    let namespace_name = genesis_data["manifest_envelope"]["manifest"]["namespace"]["name"]
        .as_str()
        .unwrap_or("qos-namespace");

    let namespace_nonce = genesis_data["manifest_envelope"]["manifest"]["namespace"]["nonce"]
        .as_u64()
        .unwrap_or(12345);

    let encrypted_shares = genesis_data["encrypted_shares"]
        .as_array()
        .expect("No encrypted_shares found in Genesis Boot response");

    println!("📊 Extracted from Genesis Boot response:");
    println!("   🏷️  Namespace: {}", namespace_name);
    println!("   🔢 Nonce: {}", namespace_nonce);
    println!("   🔐 Encrypted shares: {}", encrypted_shares.len());

    // Create share injection request with dynamic data
    let share_injection_request = serde_json::json!({
        "namespace_name": namespace_name,
        "namespace_nonce": namespace_nonce,
        "shares": encrypted_shares.iter().map(|share| {
            serde_json::json!({
                "member_alias": share["share_set_member"]["alias"],
                "decrypted_share": share["encrypted_quorum_key_share"] // Use encrypted as decrypted for testing
            })
        }).collect::<Vec<_>>()
    });

    // Write to file
    std::fs::write(
        "/tmp/share_injection_request.json",
        serde_json::to_string_pretty(&share_injection_request).unwrap(),
    )
    .unwrap();

    println!("✅ Dynamic share injection request generated:");
    println!("   📁 Saved to: /tmp/share_injection_request.json");
    println!("   👥 Shares: {}", encrypted_shares.len());
    println!("   🏷️  Namespace: {}", namespace_name);
}
