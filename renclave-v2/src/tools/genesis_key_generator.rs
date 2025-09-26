use p256::ecdsa::SigningKey;
use p256::elliptic_curve::sec1::ToEncodedPoint;
use p256::PublicKey;
use rand::rngs::OsRng;
use std::env;

fn main() {
    println!("🔑 Generating dynamic P256 keys for Genesis Boot (matching QoS)...");

    // Get threshold from command line args or default to 2
    let threshold = env::args()
        .nth(1)
        .and_then(|s| s.parse::<usize>().ok())
        .unwrap_or(2);

    println!(
        "📊 Generating {} members with threshold {}",
        threshold, threshold
    );

    // Generate dynamic P256 key pairs
    let mut members = Vec::new();

    for i in 1..=threshold {
        let signing_key = SigningKey::random(&mut OsRng);
        let verifying_key = signing_key.verifying_key();
        let public_key = PublicKey::from(verifying_key);
        let public_bytes = public_key.to_encoded_point(false).as_bytes().to_vec();

        let member = serde_json::json!({
            "alias": format!("member{}", i),
            "pub_key": public_bytes
        });

        println!("✅ Generated member {}: {} bytes", i, public_bytes.len());
        members.push(member);
    }

    // Create Genesis Boot request with dynamic keys (matching QoS)
    let genesis_request = serde_json::json!({
        "namespace_name": "qos-namespace",
        "namespace_nonce": 12345,
        "manifest_members": members,
        "manifest_threshold": threshold,
        "share_members": members, // Same members for shares
        "share_threshold": threshold,
        "pivot_hash": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32],
        "pivot_args": ["arg1", "arg2"],
        "dr_key": null
    });

    // Write to file
    std::fs::write(
        "/tmp/genesis_request.json",
        serde_json::to_string_pretty(&genesis_request).unwrap(),
    )
    .unwrap();

    println!("✅ Dynamic Genesis Boot request generated:");
    println!("   📁 Saved to: /tmp/genesis_request.json");
    println!("   👥 Members: {}", threshold);
    println!("   🔐 Threshold: {}", threshold);
    println!("   🏷️  Namespace: qos-namespace");
}
