use anyhow::{Context, Result};
use log::{debug, error, info, warn};
use std::fs;
use std::path::Path;
use std::process::Command;

pub mod connectivity;
pub mod tap;

pub use connectivity::*;
pub use tap::*;

/// Network configuration for QEMU TAP interface
#[derive(Debug, Clone)]
pub struct NetworkConfig {
    pub tap_interface: String,
    pub guest_ip: String,
    pub guest_netmask: String,
    pub gateway_ip: String,
    pub dns_servers: Vec<String>,
}

impl Default for NetworkConfig {
    fn default() -> Self {
        Self {
            tap_interface: "tap0".to_string(),
            guest_ip: "192.168.100.2".to_string(),
            guest_netmask: "255.255.255.0".to_string(),
            gateway_ip: "192.168.100.1".to_string(),
            dns_servers: vec![
                "8.8.8.8".to_string(),
                "8.8.4.4".to_string(),
                "1.1.1.1".to_string(),
            ],
        }
    }
}

/// Network manager for QEMU guest
pub struct NetworkManager {
    config: NetworkConfig,
}

impl NetworkManager {
    pub fn new(config: NetworkConfig) -> Self {
        Self { config }
    }

    /// Initialize network interfaces and connectivity
    pub async fn initialize(&self) -> Result<()> {
        info!("🌐 Initializing QEMU network configuration");

        // Check if we're in a QEMU environment
        self.detect_qemu_environment()?;

        // Setup basic network interfaces
        self.setup_loopback().await?;

        // Setup TAP interface
        self.setup_tap_interface().await?;

        // Configure routing
        self.setup_routing().await?;

        // Configure DNS
        self.setup_dns().await?;

        // Test connectivity
        self.test_connectivity().await?;

        info!("✅ Network initialization completed successfully");
        Ok(())
    }

    /// Detect if we're running in a QEMU environment
    fn detect_qemu_environment(&self) -> Result<()> {
        debug!("🔍 Detecting QEMU environment");

        // Check for QEMU-specific files and directories
        let qemu_indicators = ["/dev/net/tun", "/sys/class/net", "/proc/net/dev"];

        let mut found_indicators = 0;
        for indicator in &qemu_indicators {
            if Path::new(indicator).exists() {
                found_indicators += 1;
                debug!("✅ Found QEMU indicator: {}", indicator);
            } else {
                debug!("❌ Missing QEMU indicator: {}", indicator);
            }
        }

        if found_indicators == 0 {
            warn!("⚠️  No QEMU indicators found - may be running in non-QEMU environment");
        } else {
            info!(
                "✅ QEMU environment detected ({}/{} indicators)",
                found_indicators,
                qemu_indicators.len()
            );
        }

        // Check system information
        if let Ok(output) = Command::new("uname").args(["-a"]).output() {
            let system_info = String::from_utf8_lossy(&output.stdout);
            debug!("System info: {}", system_info.trim());
        }

        Ok(())
    }

    /// Setup loopback interface
    async fn setup_loopback(&self) -> Result<()> {
        info!("🔧 Setting up loopback interface");

        // Bring up loopback interface
        let result = Command::new("ip")
            .args(["link", "set", "lo", "up"])
            .output()
            .context("Failed to bring up loopback interface")?;

        if result.status.success() {
            debug!("✅ Loopback interface brought up successfully");
        } else {
            warn!(
                "⚠️  Failed to bring up loopback: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        // Configure loopback address
        let result = Command::new("ip")
            .args(["addr", "add", "127.0.0.1/8", "dev", "lo"])
            .output()
            .context("Failed to configure loopback address")?;

        if result.status.success() {
            debug!("✅ Loopback address configured successfully");
        } else {
            // Address might already exist
            debug!(
                "ℹ️  Loopback address configuration: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Setup TAP interface
    async fn setup_tap_interface(&self) -> Result<()> {
        info!("🔧 Setting up TAP interface: {}", self.config.tap_interface);

        // Check if TAP interface exists
        let result = Command::new("ip")
            .args(["link", "show", &self.config.tap_interface])
            .output()
            .context("Failed to check TAP interface")?;

        if !result.status.success() {
            warn!(
                "⚠️  TAP interface {} not found - may need to be created by QEMU",
                self.config.tap_interface
            );
            return Ok(());
        }

        info!("✅ TAP interface {} found", self.config.tap_interface);

        // Bring up TAP interface
        let result = Command::new("ip")
            .args(["link", "set", &self.config.tap_interface, "up"])
            .output()
            .context("Failed to bring up TAP interface")?;

        if result.status.success() {
            debug!("✅ TAP interface brought up successfully");
        } else {
            error!(
                "❌ Failed to bring up TAP interface: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        // Configure IP address
        let ip_with_mask = format!("{}/24", self.config.guest_ip);
        let result = Command::new("ip")
            .args([
                "addr",
                "add",
                &ip_with_mask,
                "dev",
                &self.config.tap_interface,
            ])
            .output()
            .context("Failed to configure TAP interface IP")?;

        if result.status.success() {
            info!("✅ TAP interface IP configured: {}", ip_with_mask);
        } else {
            // Address might already exist
            debug!(
                "ℹ️  TAP interface IP configuration: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Setup routing
    async fn setup_routing(&self) -> Result<()> {
        info!(
            "🔧 Setting up routing via gateway: {}",
            self.config.gateway_ip
        );

        // Add default route
        let result = Command::new("ip")
            .args([
                "route",
                "add",
                "default",
                "via",
                &self.config.gateway_ip,
                "dev",
                &self.config.tap_interface,
            ])
            .output()
            .context("Failed to add default route")?;

        if result.status.success() {
            info!("✅ Default route configured via {}", self.config.gateway_ip);
        } else {
            // Route might already exist
            debug!(
                "ℹ️  Default route configuration: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Setup DNS configuration
    async fn setup_dns(&self) -> Result<()> {
        info!("🔧 Setting up DNS configuration");

        let mut resolv_conf = String::new();
        for dns in &self.config.dns_servers {
            resolv_conf.push_str(&format!("nameserver {}\n", dns));
        }

        // Write DNS configuration
        if let Err(e) = fs::write("/etc/resolv.conf", &resolv_conf) {
            warn!("⚠️  Failed to write DNS configuration: {}", e);
        } else {
            info!(
                "✅ DNS configuration written with {} servers",
                self.config.dns_servers.len()
            );
        }

        Ok(())
    }

    /// Test network connectivity
    async fn test_connectivity(&self) -> Result<()> {
        info!("🔍 Testing network connectivity");

        // Test loopback
        self.test_loopback().await?;

        // Test gateway
        self.test_gateway().await?;

        // Test external connectivity
        self.test_external().await?;

        // Test DNS resolution
        self.test_dns().await?;

        Ok(())
    }

    /// Test loopback connectivity
    async fn test_loopback(&self) -> Result<()> {
        debug!("🔍 Testing loopback connectivity");

        let result = Command::new("ping")
            .args(["-c", "1", "-W", "2", "127.0.0.1"])
            .output()
            .context("Failed to ping loopback")?;

        if result.status.success() {
            debug!("✅ Loopback connectivity working");
        } else {
            warn!("⚠️  Loopback connectivity failed");
        }

        Ok(())
    }

    /// Test gateway connectivity
    async fn test_gateway(&self) -> Result<()> {
        debug!(
            "🔍 Testing gateway connectivity: {}",
            self.config.gateway_ip
        );

        let result = Command::new("ping")
            .args(["-c", "1", "-W", "5", &self.config.gateway_ip])
            .output()
            .context("Failed to ping gateway")?;

        if result.status.success() {
            info!(
                "✅ Gateway connectivity working: {}",
                self.config.gateway_ip
            );
        } else {
            warn!("⚠️  Gateway connectivity failed: {}", self.config.gateway_ip);
        }

        Ok(())
    }

    /// Test external connectivity
    async fn test_external(&self) -> Result<()> {
        debug!("🔍 Testing external connectivity");

        let test_ips = ["8.8.8.8", "1.1.1.1"];

        for ip in &test_ips {
            let result = Command::new("ping")
                .args(["-c", "1", "-W", "5", ip])
                .output()
                .context("Failed to ping external IP")?;

            if result.status.success() {
                info!("✅ External connectivity working: {}", ip);
                return Ok(());
            }
        }

        warn!("⚠️  External connectivity failed for all test IPs");
        Ok(())
    }

    /// Test DNS resolution
    async fn test_dns(&self) -> Result<()> {
        debug!("🔍 Testing DNS resolution");

        let result = Command::new("nslookup")
            .args(["google.com"])
            .output()
            .context("Failed to test DNS resolution")?;

        if result.status.success() {
            info!("✅ DNS resolution working");
        } else {
            warn!("⚠️  DNS resolution failed");
        }

        Ok(())
    }

    /// Get network status information
    pub async fn get_status(&self) -> NetworkStatus {
        NetworkStatus {
            tap_interface: self.config.tap_interface.clone(),
            guest_ip: self.config.guest_ip.clone(),
            gateway_ip: self.config.gateway_ip.clone(),
            connectivity: self.check_connectivity().await,
        }
    }

    /// Check connectivity status
    async fn check_connectivity(&self) -> ConnectivityStatus {
        let loopback = self.test_loopback().await.is_ok();
        let gateway = self.test_gateway().await.is_ok();
        let external = self.test_external().await.is_ok();
        let dns = self.test_dns().await.is_ok();

        ConnectivityStatus {
            loopback,
            gateway,
            external,
            dns,
        }
    }
}

#[derive(Debug, Clone)]
pub struct NetworkStatus {
    pub tap_interface: String,
    pub guest_ip: String,
    pub gateway_ip: String,
    pub connectivity: ConnectivityStatus,
}

#[derive(Debug, Clone)]
pub struct ConnectivityStatus {
    pub loopback: bool,
    pub gateway: bool,
    pub external: bool,
    pub dns: bool,
}
