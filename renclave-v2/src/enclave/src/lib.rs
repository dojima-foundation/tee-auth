//! Renclave Enclave Library
//!
//! This library provides the core enclave functionality for secure seed generation
//! and cryptographic operations.

pub mod nitro;
pub mod seed_generator;

// Re-export main types for convenience
pub use seed_generator::AddressDerivationResult;
pub use seed_generator::KeyDerivationResult;
pub use seed_generator::SeedGenerator;
pub use seed_generator::SeedResult;
