//! Renclave Host Library
//!
//! This library provides the host-side API functionality for communicating with
//! the enclave and handling HTTP requests.

pub mod api_handlers;
pub mod enclave_client;

// Re-export main types for convenience
pub use api_handlers::*;
pub use enclave_client::*;

use enclave_client::EnclaveClient;
use renclave_network::{ConnectivityTester, NetworkManager};
use std::sync::Arc;

/// Application state shared across handlers
#[derive(Clone)]
pub struct AppState {
    pub enclave_client: Arc<EnclaveClient>,
    pub network_manager: Arc<NetworkManager>,
    pub connectivity_tester: Arc<ConnectivityTester>,
}
