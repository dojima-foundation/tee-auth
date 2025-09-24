//! Renclave Enclave Library
//!
//! This library provides the core enclave functionality for secure seed generation
//! and cryptographic operations.

pub mod nitro;
pub mod quorum;
pub mod seed_generator;
pub mod storage;
pub mod manifest;
pub mod tee_waiting;
pub mod genesis_boot;
pub mod data_encryption;
pub mod attestation;
pub mod tee_communication;

// Re-export main types for convenience
pub use quorum::{
    boot_genesis, shares_generate, shares_reconstruct, EncryptedQuorumKey,
    GenesisOutput, GenesisSet, P256Pair, P256Public,
};
pub use seed_generator::AddressDerivationResult;
pub use seed_generator::KeyDerivationResult;
pub use seed_generator::SeedGenerator;
pub use seed_generator::SeedResult;
pub use storage::{TeeStorage, StorageState};
pub use manifest::{
    generate_manifest_with_quorum_key, calculate_manifest_hash, 
    validate_manifest_envelope, create_default_manifest
};
pub use tee_waiting::{
    TeeWaitingManager, WaitingConfig, WaitingState, simulate_tee_waiting_with_mock_shares
};
pub use genesis_boot::{
    GenesisBootFlow, GenesisBootConfig, GenesisBootResult,
    create_default_genesis_boot_config, create_test_genesis_boot_config
};
