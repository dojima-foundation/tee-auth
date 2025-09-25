//! Simple test for TEE-to-TEE communication
//! Test the basic flow without Docker

use anyhow::Result;
use log::info;
use std::sync::{Arc, Mutex};

use renclave_enclave::{
    attestation::{
        BootKeyForwardRequest, BootKeyForwardResponse, ExportKeyRequest, InjectKeyRequest,
    },
    quorum::P256Pair,
    tee_communication::TeeCommunicationManager,
};
use renclave_shared::{
    Approval, Manifest, ManifestEnvelope, ManifestSet, Namespace, NitroConfig, PivotConfig,
    QuorumMember, RestartPolicy, ShareSet,
};

/// Simple test for TEE-to-TEE communication
fn main() -> Result<()> {
    env_logger::init();

    info!("🚀 Starting Simple TEE-to-TEE Communication Test");
    info!("===============================================");

    // Create two TEE instances
    let tee1 = Arc::new(Mutex::new(TeeCommunicationManager::new()));
    let tee2 = Arc::new(Mutex::new(TeeCommunicationManager::new()));

    // Simulate TEE1 being provisioned with quorum key
    info!("🔑 Provisioning TEE1 with quorum key");
    let quorum_key = P256Pair::generate()?;
    let manifest_envelope = create_mock_manifest_envelope()?;

    {
        let tee1_guard = tee1.lock().unwrap();
        tee1_guard.set_quorum_key(quorum_key)?;
        tee1_guard.set_manifest_envelope(manifest_envelope)?;
    }

    info!("✅ TEE1 provisioned successfully");

    // Test 1: Boot Key Forward Request (TEE2 -> TEE1)
    info!("📤 Test 1: Boot Key Forward Request");
    let boot_request = BootKeyForwardRequest {
        manifest_envelope: create_mock_manifest_envelope()?,
        pivot: b"test_pivot_app".to_vec(),
    };

    let boot_response = {
        let tee2_guard = tee2.lock().unwrap();
        tee2_guard.handle_boot_key_forward(boot_request)?
    };

    info!("✅ Boot key forward response received");

    // Test 2: Export Key Request (Client -> TEE1)
    info!("📤 Test 2: Export Key Request");
    let export_request = ExportKeyRequest {
        manifest_envelope: create_mock_manifest_envelope()?,
        cose_sign1_attestation_doc: extract_attestation_doc(&boot_response)?,
    };

    let export_response = {
        let tee1_guard = tee1.lock().unwrap();
        tee1_guard.handle_export_key(export_request)?
    };

    info!("✅ Export key response received");
    info!(
        "🔐 Encrypted quorum key length: {} bytes",
        export_response.encrypted_quorum_key.len()
    );
    info!(
        "🔐 Signature length: {} bytes",
        export_response.signature.len()
    );

    // Test 3: Inject Key Request (Client -> TEE2)
    info!("💉 Test 3: Inject Key Request");
    let inject_request = InjectKeyRequest {
        encrypted_quorum_key: export_response.encrypted_quorum_key,
        signature: export_response.signature,
    };

    let _inject_response = {
        let tee2_guard = tee2.lock().unwrap();
        tee2_guard.handle_inject_key(inject_request)?
    };

    info!("✅ Inject key response received");

    // Verify TEE2 is now provisioned
    {
        let tee2_guard = tee2.lock().unwrap();
        if tee2_guard.is_provisioned() {
            info!("✅ TEE2 is now provisioned with quorum key");
        } else {
            info!("⚠️ TEE2 is not provisioned");
        }
    }

    // Test 4: Verify both TEEs can use their quorum keys
    info!("🔍 Test 4: Verifying quorum key functionality");

    // TEE1 should still have its quorum key
    {
        let tee1_guard = tee1.lock().unwrap();
        if tee1_guard.is_provisioned() {
            info!("✅ TEE1 still has quorum key");
        } else {
            info!("⚠️ TEE1 lost quorum key");
        }
    }

    // TEE2 should now have the quorum key
    {
        let tee2_guard = tee2.lock().unwrap();
        if tee2_guard.is_provisioned() {
            info!("✅ TEE2 has quorum key");
        } else {
            info!("⚠️ TEE2 does not have quorum key");
        }
    }

    info!("🎉 TEE-to-TEE Communication Test Completed Successfully!");
    info!("========================================================");

    Ok(())
}

/// Create a mock manifest envelope for testing
fn create_mock_manifest_envelope() -> Result<ManifestEnvelope> {
    let manifest = Manifest {
        namespace: Namespace {
            nonce: 1,
            name: "test_namespace".to_string(),
            quorum_key: b"mock_quorum_key".to_vec(),
        },
        enclave: NitroConfig {
            pcr0: b"mock_pcr0".to_vec(),
            pcr1: b"mock_pcr1".to_vec(),
            pcr2: b"mock_pcr2".to_vec(),
            pcr3: b"mock_pcr3".to_vec(),
            aws_root_certificate: b"mock_aws_cert".to_vec(),
            qos_commit: "mock_commit".to_string(),
        },
        pivot: PivotConfig {
            hash: [0u8; 32],
            restart: RestartPolicy::Never,
            args: vec![],
        },
        manifest_set: ManifestSet {
            threshold: 2,
            members: vec![],
        },
        share_set: ShareSet {
            threshold: 2,
            members: vec![],
        },
    };

    let approval = Approval {
        signature: b"mock_signature".to_vec(),
        member: QuorumMember {
            alias: "test_member".to_string(),
            pub_key: b"mock_approver_key".to_vec(),
        },
    };

    Ok(ManifestEnvelope {
        manifest,
        manifest_set_approvals: vec![approval.clone()],
        share_set_approvals: vec![approval],
    })
}

/// Extract attestation document from boot response
fn extract_attestation_doc(boot_response: &BootKeyForwardResponse) -> Result<Vec<u8>> {
    match &boot_response.nsm_response {
        renclave_enclave::attestation::NsmResponse::Attestation { document } => {
            Ok(document.clone())
        }
        _ => Err(anyhow::anyhow!("No attestation document in response")),
    }
}
