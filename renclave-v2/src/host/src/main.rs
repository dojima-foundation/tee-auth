use axum::{
    routing::{get, post},
    Router,
};
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;
// Removed tower imports for now
use log::{info, warn};

mod api_handlers;
mod enclave_client;

use enclave_client::EnclaveClient;
use renclave_network::{ConnectivityTester, NetworkConfig, NetworkManager};

/// QEMU Host - HTTP API Gateway for Nitro Enclave
pub struct QemuHost {
    enclave_client: Arc<EnclaveClient>,
    network_manager: Arc<NetworkManager>,
    connectivity_tester: Arc<ConnectivityTester>,
}

impl QemuHost {
    /// Create new QEMU host instance
    pub async fn new() -> anyhow::Result<Self> {
        info!("🏠 Initializing QEMU Host (API Gateway)");

        // Initialize network manager
        info!("🌐 Initializing network manager...");
        let network_config = NetworkConfig::default();
        let network_manager = Arc::new(NetworkManager::new(network_config));

        // Initialize network (non-blocking)
        let network_manager_clone = Arc::clone(&network_manager);
        tokio::spawn(async move {
            if let Err(e) = network_manager_clone.initialize().await {
                warn!("⚠️  Network initialization failed: {}", e);
            }
        });

        // Initialize connectivity tester
        let connectivity_tester = Arc::new(ConnectivityTester::default());

        // Initialize enclave client
        info!("🔗 Initializing enclave client...");
        let enclave_client = Arc::new(EnclaveClient::new("/tmp/enclave.sock".to_string()));

        // Wait for enclave to be available
        info!("⏳ Waiting for enclave to be available...");
        enclave_client
            .wait_for_enclave(Duration::from_secs(30))
            .await?;
        info!("✅ Enclave is available");

        Ok(Self {
            enclave_client,
            network_manager,
            connectivity_tester,
        })
    }

    /// Start the HTTP server
    pub async fn start(&self, bind_addr: SocketAddr) -> anyhow::Result<()> {
        info!("🚀 Starting QEMU Host HTTP server");

        // Create application state
        let app_state = AppState {
            enclave_client: Arc::clone(&self.enclave_client),
            network_manager: Arc::clone(&self.network_manager),
            connectivity_tester: Arc::clone(&self.connectivity_tester),
        };

        // Build router
        let app = Router::new()
            .route("/health", get(api_handlers::health_check))
            .route("/info", get(api_handlers::get_info))
            .route("/generate-seed", post(api_handlers::generate_seed))
            .route("/validate-seed", post(api_handlers::validate_seed))
            .route("/derive-key", post(api_handlers::derive_key))
            .route("/derive-address", post(api_handlers::derive_address))
            .route("/network/status", get(api_handlers::network_status))
            .route("/network/test", post(api_handlers::test_connectivity))
            .route("/enclave/info", get(api_handlers::enclave_info))
            .route("/enclave/genesis-boot", post(api_handlers::genesis_boot))
            .route("/enclave/inject-shares", post(api_handlers::inject_shares))
            .route("/enclave/encrypt-data", post(api_handlers::encrypt_data))
            .route("/enclave/decrypt-data", post(api_handlers::decrypt_data))
            .route(
                "/enclave/sign-transaction",
                post(api_handlers::sign_transaction),
            )
            .route("/enclave/sign-message", post(api_handlers::sign_message))
            .route(
                "/enclave/application-status",
                get(api_handlers::get_application_status),
            )
            .route(
                "/enclave/store-data",
                post(api_handlers::store_application_data),
            )
            .route("/enclave/get-data", get(api_handlers::get_application_data))
            .route("/enclave/reset", post(api_handlers::reset_enclave))
            // TEE-to-TEE Communication endpoints
            .route(
                "/enclave/boot-key-forward",
                post(api_handlers::boot_key_forward),
            )
            .route("/enclave/export-key", post(api_handlers::export_key))
            .route("/enclave/inject-key", post(api_handlers::inject_key))
            .route(
                "/enclave/generate-attestation",
                post(api_handlers::generate_attestation),
            )
            .route(
                "/enclave/share-manifest",
                post(api_handlers::share_manifest),
            )
            .with_state(app_state);

        info!("✅ HTTP router configured with all endpoints");
        info!("🔗 Binding to address: {}", bind_addr);

        // Start server
        let listener = tokio::net::TcpListener::bind(bind_addr).await?;
        info!("🚀 QEMU Host HTTP server started on {}", bind_addr);

        axum::serve(listener, app).await?;

        Ok(())
    }
}

/// Application state shared across handlers
#[derive(Clone)]
pub struct AppState {
    pub enclave_client: Arc<EnclaveClient>,
    pub network_manager: Arc<NetworkManager>,
    pub connectivity_tester: Arc<ConnectivityTester>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize logging
    std::env::set_var("RUST_LOG", "debug");
    env_logger::init();

    info!("🏠 QEMU Host - HTTP API Gateway for Nitro Enclave");
    info!("🔍 Process ID: {}", std::process::id());
    info!(
        "🔍 Current working directory: {:?}",
        std::env::current_dir()?
    );

    // Create and start host
    let host = QemuHost::new().await?;

    // Start HTTP server on all interfaces
    let bind_addr = SocketAddr::from(([0, 0, 0, 0], 8080));
    host.start(bind_addr).await?;

    Ok(())
}
