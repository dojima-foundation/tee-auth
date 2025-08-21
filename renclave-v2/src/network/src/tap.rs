use anyhow::{Context, Result};
use log::{debug, info, warn};
use std::process::Command;

/// TAP interface management
pub struct TapInterface {
    name: String,
}

impl TapInterface {
    pub fn new(name: String) -> Self {
        Self { name }
    }

    /// Create TAP interface (usually done by QEMU)
    pub fn create(&self) -> Result<()> {
        info!("🔧 Creating TAP interface: {}", self.name);

        // Note: In production, QEMU creates the TAP interface
        // This is mainly for testing/development
        let result = Command::new("ip")
            .args(["tuntap", "add", "dev", &self.name, "mode", "tap"])
            .output()
            .context("Failed to create TAP interface")?;

        if result.status.success() {
            info!("✅ TAP interface created: {}", self.name);
        } else {
            warn!(
                "⚠️  TAP interface creation failed: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Configure TAP interface
    pub fn configure(&self, ip: &str, netmask: &str) -> Result<()> {
        info!(
            "🔧 Configuring TAP interface: {} with IP {}/{}",
            self.name, ip, netmask
        );

        // Bring interface up
        let result = Command::new("ip")
            .args(["link", "set", &self.name, "up"])
            .output()
            .context("Failed to bring up TAP interface")?;

        if !result.status.success() {
            warn!(
                "⚠️  Failed to bring up TAP interface: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        // Set IP address
        let ip_with_cidr = format!("{}/24", ip); // Assuming /24 for simplicity
        let result = Command::new("ip")
            .args(["addr", "add", &ip_with_cidr, "dev", &self.name])
            .output()
            .context("Failed to configure TAP interface IP")?;

        if result.status.success() {
            info!("✅ TAP interface configured: {}", ip_with_cidr);
        } else {
            debug!(
                "ℹ️  TAP interface IP configuration: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Remove TAP interface
    pub fn remove(&self) -> Result<()> {
        info!("🗑️  Removing TAP interface: {}", self.name);

        let result = Command::new("ip")
            .args(["link", "delete", &self.name])
            .output()
            .context("Failed to remove TAP interface")?;

        if result.status.success() {
            info!("✅ TAP interface removed: {}", self.name);
        } else {
            warn!(
                "⚠️  TAP interface removal failed: {}",
                String::from_utf8_lossy(&result.stderr)
            );
        }

        Ok(())
    }

    /// Check if TAP interface exists
    pub fn exists(&self) -> bool {
        let result = Command::new("ip")
            .args(["link", "show", &self.name])
            .output();

        match result {
            Ok(output) => output.status.success(),
            Err(_) => false,
        }
    }

    /// Get TAP interface statistics
    pub fn get_stats(&self) -> Result<TapStats> {
        let result = Command::new("cat")
            .args([&format!("/sys/class/net/{}/statistics/rx_bytes", self.name)])
            .output()
            .context("Failed to read RX bytes")?;

        let rx_bytes = if result.status.success() {
            String::from_utf8_lossy(&result.stdout)
                .trim()
                .parse()
                .unwrap_or(0)
        } else {
            0
        };

        let result = Command::new("cat")
            .args([&format!("/sys/class/net/{}/statistics/tx_bytes", self.name)])
            .output()
            .context("Failed to read TX bytes")?;

        let tx_bytes = if result.status.success() {
            String::from_utf8_lossy(&result.stdout)
                .trim()
                .parse()
                .unwrap_or(0)
        } else {
            0
        };

        Ok(TapStats { rx_bytes, tx_bytes })
    }
}

#[derive(Debug, Clone)]
pub struct TapStats {
    pub rx_bytes: u64,
    pub tx_bytes: u64,
}
