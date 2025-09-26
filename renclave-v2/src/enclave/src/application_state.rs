//! Application state management following QoS state.rs patterns
//! Implements protocol phases and state transitions for application lifecycle

use anyhow::{anyhow, Result};
use borsh::{BorshDeserialize, BorshSerialize};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use crate::quorum::P256Pair;

/// Application phase following QoS ProtocolPhase exactly.
#[derive(
    Debug, Copy, Clone, PartialEq, Eq, BorshSerialize, BorshDeserialize, Serialize, Deserialize,
)]
pub enum ApplicationPhase {
    /// The state machine cannot recover. The enclave must be rebooted.
    UnrecoverableError,
    /// Waiting to receive a boot instruction.
    WaitingForBootInstruction,
    /// Genesis service has been booted. No further actions.
    GenesisBooted,
    /// Waiting to receive K quorum shards
    WaitingForQuorumShards,
    /// The enclave has successfully provisioned its quorum key.
    QuorumKeyProvisioned,
    /// Waiting for a forwarded key to be injected
    WaitingForForwardedKey,
    /// Application is ready for normal operations (new phase for our use case)
    ApplicationReady,
}

impl ApplicationPhase {
    /// Get a human-readable description of the phase.
    #[allow(dead_code)]
    pub fn description(&self) -> &'static str {
        match self {
            ApplicationPhase::UnrecoverableError => {
                "Unrecoverable error - enclave must be rebooted"
            }
            ApplicationPhase::WaitingForBootInstruction => "Waiting for boot instruction",
            ApplicationPhase::GenesisBooted => "Genesis service has been booted",
            ApplicationPhase::WaitingForQuorumShards => "Waiting for quorum shards",
            ApplicationPhase::QuorumKeyProvisioned => "Quorum key has been provisioned",
            ApplicationPhase::WaitingForForwardedKey => "Waiting for forwarded key",
            ApplicationPhase::ApplicationReady => "Application is ready for operations",
        }
    }

    /// Check if the phase allows normal operations.
    pub fn allows_operations(&self) -> bool {
        matches!(self, ApplicationPhase::ApplicationReady)
    }

    /// Check if the phase allows quorum operations.
    pub fn allows_quorum_operations(&self) -> bool {
        matches!(
            self,
            ApplicationPhase::QuorumKeyProvisioned | ApplicationPhase::ApplicationReady
        )
    }
}

/// Application state following QoS ProtocolState structure.
#[derive(Clone)]
pub struct ApplicationState {
    /// Current phase of the application
    phase: ApplicationPhase,
    /// Quorum key (if available)
    quorum_key: Option<P256Pair>,
    /// Application-specific data storage
    application_data: HashMap<String, Vec<u8>>,
    /// Metadata about the application
    metadata: ApplicationMetadata,
}

/// Application metadata for tracking state information.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApplicationMetadata {
    /// Application name
    pub name: String,
    /// Application version
    pub version: String,
    /// Last updated timestamp
    pub last_updated: u64,
    /// Number of operations performed
    pub operation_count: u64,
}

impl ApplicationState {
    /// Create a new application state.
    /// Following QoS ProtocolState::new() patterns.
    pub fn new(name: String, version: String) -> Self {
        Self {
            phase: ApplicationPhase::WaitingForBootInstruction,
            quorum_key: None,
            application_data: HashMap::new(),
            metadata: ApplicationMetadata {
                name,
                version,
                last_updated: 0,
                operation_count: 0,
            },
        }
    }

    /// Get the current phase.
    /// Following QoS ProtocolState::get_phase() exactly.
    #[allow(dead_code)]
    pub fn get_phase(&self) -> ApplicationPhase {
        self.phase
    }

    /// Transition to a new phase.
    /// Following QoS ProtocolState::transition() exactly.
    pub fn transition(&mut self, next: ApplicationPhase) -> Result<()> {
        if self.phase == next {
            return Ok(());
        }

        let transitions = self.get_allowed_transitions();

        if !transitions.contains(&next) {
            let prev = self.phase;
            self.phase = ApplicationPhase::UnrecoverableError;
            return Err(anyhow!(
                "Invalid state transition from {:?} to {:?}",
                prev,
                next
            ));
        }

        self.phase = next;
        self.metadata.last_updated = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();

        Ok(())
    }

    /// Get allowed transitions from current phase.
    /// Following QoS ProtocolState transition logic exactly.
    fn get_allowed_transitions(&self) -> Vec<ApplicationPhase> {
        match self.phase {
            ApplicationPhase::UnrecoverableError => vec![],
            ApplicationPhase::WaitingForBootInstruction => vec![
                ApplicationPhase::UnrecoverableError,
                ApplicationPhase::GenesisBooted,
                ApplicationPhase::WaitingForQuorumShards,
                ApplicationPhase::WaitingForForwardedKey,
            ],
            ApplicationPhase::GenesisBooted => {
                vec![
                    ApplicationPhase::UnrecoverableError,
                    ApplicationPhase::QuorumKeyProvisioned,
                ]
            }
            ApplicationPhase::WaitingForQuorumShards => {
                vec![
                    ApplicationPhase::UnrecoverableError,
                    ApplicationPhase::QuorumKeyProvisioned,
                ]
            }
            ApplicationPhase::QuorumKeyProvisioned => {
                vec![
                    ApplicationPhase::UnrecoverableError,
                    ApplicationPhase::ApplicationReady,
                ]
            }
            ApplicationPhase::WaitingForForwardedKey => {
                vec![
                    ApplicationPhase::UnrecoverableError,
                    ApplicationPhase::QuorumKeyProvisioned,
                ]
            }
            ApplicationPhase::ApplicationReady => {
                vec![ApplicationPhase::UnrecoverableError]
            }
        }
    }

    /// Set the quorum key.
    /// Following QoS patterns for quorum key management.
    pub fn set_quorum_key(&mut self, quorum_key: P256Pair) -> Result<()> {
        if !self.phase.allows_quorum_operations() {
            return Err(anyhow!("Cannot set quorum key in phase {:?}", self.phase));
        }

        self.quorum_key = Some(quorum_key);
        self.metadata.operation_count += 1;
        Ok(())
    }

    /// Get the quorum key.
    /// Following QoS patterns for quorum key access.
    pub fn get_quorum_key(&self) -> Result<&P256Pair> {
        self.quorum_key
            .as_ref()
            .ok_or_else(|| anyhow!("Quorum key not available in phase {:?}", self.phase))
    }

    /// Check if quorum key is available.
    pub fn has_quorum_key(&self) -> bool {
        self.quorum_key.is_some()
    }

    /// Store application data.
    /// Following QoS patterns for data storage.
    pub fn store_data(&mut self, key: String, data: Vec<u8>) -> Result<()> {
        if !self.phase.allows_operations() {
            return Err(anyhow!("Cannot store data in phase {:?}", self.phase));
        }

        self.application_data.insert(key, data);
        self.metadata.operation_count += 1;
        Ok(())
    }

    /// Retrieve application data.
    /// Following QoS patterns for data retrieval.
    pub fn get_data(&self, key: &str) -> Option<&Vec<u8>> {
        self.application_data.get(key)
    }

    /// Remove application data.
    /// Following QoS patterns for data management.
    #[allow(dead_code)]
    pub fn remove_data(&mut self, key: &str) -> Result<()> {
        if !self.phase.allows_operations() {
            return Err(anyhow!("Cannot remove data in phase {:?}", self.phase));
        }

        self.application_data.remove(key);
        self.metadata.operation_count += 1;
        Ok(())
    }

    /// Get all application data keys.
    #[allow(dead_code)]
    pub fn get_data_keys(&self) -> Vec<String> {
        self.application_data.keys().cloned().collect()
    }

    /// Get application metadata.
    #[allow(dead_code)]
    pub fn get_metadata(&self) -> &ApplicationMetadata {
        &self.metadata
    }

    /// Update application metadata.
    #[allow(dead_code)]
    pub fn update_metadata(&mut self, name: Option<String>, version: Option<String>) {
        if let Some(name) = name {
            self.metadata.name = name;
        }
        if let Some(version) = version {
            self.metadata.version = version;
        }
        self.metadata.last_updated = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();
    }

    /// Get application status information.
    /// Following QoS status patterns.
    pub fn get_status(&self) -> ApplicationStatus {
        ApplicationStatus {
            phase: self.phase,
            has_quorum_key: self.has_quorum_key(),
            data_count: self.application_data.len(),
            metadata: self.metadata.clone(),
        }
    }

    /// Reset application state (for testing).
    /// Following QoS testing patterns.
    pub fn reset(&mut self) {
        self.phase = ApplicationPhase::WaitingForBootInstruction;
        self.quorum_key = None;
        self.application_data.clear();
        self.metadata.operation_count = 0;
    }
}

/// Application status information.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApplicationStatus {
    pub phase: ApplicationPhase,
    pub has_quorum_key: bool,
    pub data_count: usize,
    pub metadata: ApplicationMetadata,
}

/// Application state manager for handling state transitions and operations.
/// Following QoS patterns for state management.
#[derive(Clone)]
pub struct ApplicationStateManager {
    state: ApplicationState,
}

impl ApplicationStateManager {
    /// Create a new application state manager.
    pub fn new(name: String, version: String) -> Self {
        Self {
            state: ApplicationState::new(name, version),
        }
    }

    /// Get the current state.
    pub fn get_state(&self) -> &ApplicationState {
        &self.state
    }

    /// Get mutable access to the state.
    pub fn get_state_mut(&mut self) -> &mut ApplicationState {
        &mut self.state
    }

    /// Transition to a new phase.
    pub fn transition_to(&mut self, phase: ApplicationPhase) -> Result<()> {
        self.state.transition(phase)
    }

    /// Set the quorum key and transition to ready state.
    pub fn provision_quorum_key(&mut self, quorum_key: P256Pair) -> Result<()> {
        self.state.set_quorum_key(quorum_key)?;
        self.state.transition(ApplicationPhase::ApplicationReady)
    }

    /// Check if the application is ready for operations.
    #[allow(dead_code)]
    pub fn is_ready(&self) -> bool {
        self.state.phase == ApplicationPhase::ApplicationReady && self.state.has_quorum_key()
    }

    /// Get application status.
    pub fn get_status(&self) -> ApplicationStatus {
        self.state.get_status()
    }

    /// Reset application state (for testing).
    /// Following QoS testing patterns.
    pub fn reset(&mut self) {
        self.state.reset();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_application_state_creation() {
        let state = ApplicationState::new("test_app".to_string(), "1.0.0".to_string());
        assert_eq!(
            state.get_phase(),
            ApplicationPhase::WaitingForBootInstruction
        );
        assert!(!state.has_quorum_key());
    }

    #[test]
    fn test_state_transitions() {
        let mut state = ApplicationState::new("test_app".to_string(), "1.0.0".to_string());

        // Valid transition
        assert!(state.transition(ApplicationPhase::GenesisBooted).is_ok());
        assert_eq!(state.get_phase(), ApplicationPhase::GenesisBooted);

        // Invalid transition
        assert!(state
            .transition(ApplicationPhase::ApplicationReady)
            .is_err());
    }

    #[test]
    fn test_quorum_key_management() {
        let mut state = ApplicationState::new("test_app".to_string(), "1.0.0".to_string());

        // Cannot set quorum key in initial phase
        let quorum_key = P256Pair::generate().unwrap();
        assert!(state.set_quorum_key(quorum_key).is_err());

        // Follow proper state transition sequence (QoS protocol)
        state.transition(ApplicationPhase::GenesisBooted).unwrap();

        // Transition to provisioned phase
        state
            .transition(ApplicationPhase::QuorumKeyProvisioned)
            .unwrap();

        // Now can set quorum key
        let quorum_key = P256Pair::generate().unwrap();
        assert!(state.set_quorum_key(quorum_key).is_ok());
        assert!(state.has_quorum_key());
    }

    #[test]
    fn test_data_operations() {
        let mut state = ApplicationState::new("test_app".to_string(), "1.0.0".to_string());

        // Cannot store data in initial phase
        assert!(state.store_data("key".to_string(), vec![1, 2, 3]).is_err());

        // Follow proper state transition sequence (QoS protocol)
        state.transition(ApplicationPhase::GenesisBooted).unwrap();

        state
            .transition(ApplicationPhase::QuorumKeyProvisioned)
            .unwrap();

        state
            .transition(ApplicationPhase::ApplicationReady)
            .unwrap();

        // Now can store data
        assert!(state.store_data("key".to_string(), vec![1, 2, 3]).is_ok());
        assert_eq!(state.get_data("key"), Some(&vec![1, 2, 3]));
    }

    #[test]
    fn test_application_state_manager() {
        let mut manager = ApplicationStateManager::new("test_app".to_string(), "1.0.0".to_string());

        assert!(!manager.is_ready());

        // Follow proper state transition sequence (QoS protocol)
        manager
            .get_state_mut()
            .transition(ApplicationPhase::GenesisBooted)
            .unwrap();
        manager
            .get_state_mut()
            .transition(ApplicationPhase::QuorumKeyProvisioned)
            .unwrap();

        // Provision quorum key
        let quorum_key = P256Pair::generate().unwrap();
        manager.provision_quorum_key(quorum_key).unwrap();

        manager
            .get_state_mut()
            .transition(ApplicationPhase::ApplicationReady)
            .unwrap();

        assert!(manager.is_ready());
        assert!(manager.get_state().has_quorum_key());
    }
}
