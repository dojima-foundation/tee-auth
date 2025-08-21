use anyhow::{Context, Result};
use log::{debug, info, warn};
use std::process::Command;
use std::time::{Duration, Instant};

/// Connectivity tester for network interfaces
pub struct ConnectivityTester {
    timeout: Duration,
}

impl ConnectivityTester {
    pub fn new(timeout: Duration) -> Self {
        Self { timeout }
    }

    /// Test HTTP connectivity to external services
    pub async fn test_http_connectivity(&self) -> Result<HttpConnectivityResult> {
        info!("🌐 Testing HTTP connectivity");

        let test_urls = [
            "http://httpbin.org/ip",
            "http://ifconfig.me/ip",
            "http://api.ipify.org",
        ];

        let start_time = Instant::now();

        for url in &test_urls {
            debug!("🔍 Testing HTTP connectivity to: {}", url);

            let result = Command::new("curl")
                .args(["-s", "--connect-timeout", "5", "--max-time", "10", url])
                .output()
                .context("Failed to execute curl command")?;

            if result.status.success() {
                let response = String::from_utf8_lossy(&result.stdout);
                let duration = start_time.elapsed();

                info!("✅ HTTP connectivity working via: {}", url);
                debug!("Response: {}", response.trim());

                return Ok(HttpConnectivityResult {
                    success: true,
                    url: url.to_string(),
                    response: response.trim().to_string(),
                    duration,
                });
            } else {
                warn!("⚠️  HTTP connectivity failed for: {}", url);
                debug!("Error: {}", String::from_utf8_lossy(&result.stderr));
            }
        }

        Ok(HttpConnectivityResult {
            success: false,
            url: "none".to_string(),
            response: "No successful connections".to_string(),
            duration: start_time.elapsed(),
        })
    }

    /// Test DNS resolution
    pub async fn test_dns_resolution(&self, hostname: &str) -> Result<DnsResult> {
        info!("🔍 Testing DNS resolution for: {}", hostname);

        let start_time = Instant::now();

        let result = Command::new("nslookup")
            .args([hostname])
            .output()
            .context("Failed to execute nslookup")?;

        let duration = start_time.elapsed();

        if result.status.success() {
            let output = String::from_utf8_lossy(&result.stdout);
            info!("✅ DNS resolution successful for: {}", hostname);
            debug!("DNS output: {}", output);

            Ok(DnsResult {
                success: true,
                hostname: hostname.to_string(),
                duration,
                output: output.to_string(),
            })
        } else {
            warn!("⚠️  DNS resolution failed for: {}", hostname);
            let error = String::from_utf8_lossy(&result.stderr);
            debug!("DNS error: {}", error);

            Ok(DnsResult {
                success: false,
                hostname: hostname.to_string(),
                duration,
                output: error.to_string(),
            })
        }
    }

    /// Test ping connectivity
    pub async fn test_ping(&self, target: &str, count: u32) -> Result<PingResult> {
        info!(
            "🏓 Testing ping connectivity to: {} ({} packets)",
            target, count
        );

        let start_time = Instant::now();

        let result = Command::new("ping")
            .args([
                "-c",
                &count.to_string(),
                "-W",
                "5", // 5 second timeout
                target,
            ])
            .output()
            .context("Failed to execute ping")?;

        let duration = start_time.elapsed();

        if result.status.success() {
            let output = String::from_utf8_lossy(&result.stdout);
            info!("✅ Ping successful to: {}", target);

            // Parse ping statistics
            let stats = self.parse_ping_stats(&output);

            Ok(PingResult {
                success: true,
                target: target.to_string(),
                duration,
                packets_sent: count,
                packets_received: stats.packets_received,
                avg_time_ms: stats.avg_time_ms,
                output: output.to_string(),
            })
        } else {
            warn!("⚠️  Ping failed to: {}", target);
            let error = String::from_utf8_lossy(&result.stderr);
            debug!("Ping error: {}", error);

            Ok(PingResult {
                success: false,
                target: target.to_string(),
                duration,
                packets_sent: count,
                packets_received: 0,
                avg_time_ms: 0.0,
                output: error.to_string(),
            })
        }
    }

    /// Parse ping statistics from output
    fn parse_ping_stats(&self, output: &str) -> PingStats {
        let mut packets_received = 0;
        let mut avg_time_ms = 0.0;

        // Look for statistics line like: "1 packets transmitted, 1 received, 0% packet loss"
        for line in output.lines() {
            if line.contains("packets transmitted") && line.contains("received") {
                if let Some(received_part) = line.split("received").next() {
                    if let Some(received_str) = received_part.split(',').nth(1) {
                        if let Ok(received) = received_str.trim().parse::<u32>() {
                            packets_received = received;
                        }
                    }
                }
            }

            // Look for average time like: "rtt min/avg/max/mdev = 1.234/2.345/3.456/0.123 ms"
            if line.contains("rtt min/avg/max") {
                if let Some(times_part) = line.split('=').nth(1) {
                    if let Some(avg_str) = times_part.split('/').nth(1) {
                        if let Ok(avg) = avg_str.trim().parse::<f64>() {
                            avg_time_ms = avg;
                        }
                    }
                }
            }
        }

        PingStats {
            packets_received,
            avg_time_ms,
        }
    }

    /// Run comprehensive connectivity test
    pub async fn run_comprehensive_test(&self) -> Result<ConnectivityReport> {
        info!("🔍 Running comprehensive connectivity test");

        let start_time = Instant::now();

        // Test ping to gateway
        let gateway_ping = self.test_ping("192.168.100.1", 3).await?;

        // Test ping to external IPs
        let external_ping = self.test_ping("8.8.8.8", 3).await?;

        // Test DNS resolution
        let dns_test = self.test_dns_resolution("google.com").await?;

        // Test HTTP connectivity
        let http_test = self.test_http_connectivity().await?;

        let total_duration = start_time.elapsed();

        let report = ConnectivityReport {
            gateway_ping,
            external_ping,
            dns_test,
            http_test,
            total_duration,
        };

        info!(
            "✅ Comprehensive connectivity test completed in {:?}",
            total_duration
        );

        Ok(report)
    }
}

impl Default for ConnectivityTester {
    fn default() -> Self {
        Self::new(Duration::from_secs(10))
    }
}

#[derive(Debug, Clone)]
pub struct HttpConnectivityResult {
    pub success: bool,
    pub url: String,
    pub response: String,
    pub duration: Duration,
}

#[derive(Debug, Clone)]
pub struct DnsResult {
    pub success: bool,
    pub hostname: String,
    pub duration: Duration,
    pub output: String,
}

#[derive(Debug, Clone)]
pub struct PingResult {
    pub success: bool,
    pub target: String,
    pub duration: Duration,
    pub packets_sent: u32,
    pub packets_received: u32,
    pub avg_time_ms: f64,
    pub output: String,
}

#[derive(Debug, Clone)]
struct PingStats {
    packets_received: u32,
    avg_time_ms: f64,
}

#[derive(Debug, Clone)]
pub struct ConnectivityReport {
    pub gateway_ping: PingResult,
    pub external_ping: PingResult,
    pub dns_test: DnsResult,
    pub http_test: HttpConnectivityResult,
    pub total_duration: Duration,
}
