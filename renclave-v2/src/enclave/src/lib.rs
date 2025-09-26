//! Renclave Enclave Library
//!
//! This library provides the core enclave functionality for secure seed generation
//! and cryptographic operations.

pub mod attestation;
pub mod data_encryption;
pub mod genesis_boot;
pub mod manifest;
pub mod nitro;
pub mod quorum;
pub mod seed_generator;
pub mod storage;
pub mod tee_communication;
pub mod tee_waiting;

// Re-export main types for convenience
pub use genesis_boot::{
    create_default_genesis_boot_config, create_test_genesis_boot_config, GenesisBootConfig,
    GenesisBootFlow, GenesisBootResult,
};
pub use manifest::{
    calculate_manifest_hash, create_default_manifest, generate_manifest_with_quorum_key,
    validate_manifest_envelope,
};
pub use quorum::{
    boot_genesis, shares_generate, shares_reconstruct, EncryptedQuorumKey, GenesisOutput,
    GenesisSet, P256Pair, P256Public,
};
pub use seed_generator::AddressDerivationResult;
pub use seed_generator::KeyDerivationResult;
pub use seed_generator::SeedGenerator;
pub use seed_generator::SeedResult;
pub use storage::{StorageState, TeeStorage};
pub use tee_waiting::{
    simulate_tee_waiting_with_mock_shares, TeeWaitingManager, WaitingConfig, WaitingState,
};
