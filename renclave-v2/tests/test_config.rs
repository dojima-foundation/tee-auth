//! Test configuration and utilities for renclave system
//! Provides centralized configuration for all test types

use serde::{Deserialize, Serialize};
use std::time::Duration;

/// Test configuration for different test types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TestConfig {
    /// Unit test configuration
    pub unit_tests: UnitTestConfig,
    /// Integration test configuration
    pub integration_tests: IntegrationTestConfig,
    /// End-to-end test configuration
    pub e2e_tests: E2ETestConfig,
    /// Performance test configuration
    pub performance_tests: PerformanceTestConfig,
    /// Stress test configuration
    pub stress_tests: StressTestConfig,
    /// Global test settings
    pub global: GlobalTestConfig,
}

/// Unit test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnitTestConfig {
    pub enabled: bool,
    pub timeout_seconds: u64,
    pub max_concurrent: usize,
    pub modules: Vec<String>,
}

/// Integration test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IntegrationTestConfig {
    pub enabled: bool,
    pub timeout_seconds: u64,
    pub max_concurrent: usize,
    pub test_workflows: Vec<String>,
}

/// End-to-end test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct E2ETestConfig {
    pub enabled: bool,
    pub timeout_seconds: u64,
    pub max_concurrent: usize,
    pub test_scenarios: Vec<String>,
}

/// Performance test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceTestConfig {
    pub enabled: bool,
    pub timeout_seconds: u64,
    pub max_concurrent: usize,
    pub iterations: usize,
    pub warmup_iterations: usize,
}

/// Stress test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StressTestConfig {
    pub enabled: bool,
    pub timeout_seconds: u64,
    pub max_concurrent: usize,
    pub iterations: usize,
    pub data_sizes: Vec<usize>,
}

/// Global test configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GlobalTestConfig {
    pub log_level: String,
    pub output_format: String,
    pub generate_reports: bool,
    pub cleanup_after_tests: bool,
    pub parallel_execution: bool,
}

impl Default for TestConfig {
    fn default() -> Self {
        Self {
            unit_tests: UnitTestConfig {
                enabled: true,
                timeout_seconds: 60,
                max_concurrent: 10,
                modules: vec![
                    "seed_generator".to_string(),
                    "quorum".to_string(),
                    "data_encryption".to_string(),
                    "tee_communication".to_string(),
                    "storage".to_string(),
                    "network".to_string(),
                ],
            },
            integration_tests: IntegrationTestConfig {
                enabled: true,
                timeout_seconds: 120,
                max_concurrent: 5,
                test_workflows: vec![
                    "seed_generation_workflow".to_string(),
                    "quorum_key_generation_workflow".to_string(),
                    "data_encryption_workflow".to_string(),
                    "tee_communication_workflow".to_string(),
                    "storage_workflow".to_string(),
                    "network_workflow".to_string(),
                ],
            },
            e2e_tests: E2ETestConfig {
                enabled: true,
                timeout_seconds: 300,
                max_concurrent: 3,
                test_scenarios: vec![
                    "complete_seed_generation_scenario".to_string(),
                    "complete_quorum_key_generation_scenario".to_string(),
                    "complete_data_encryption_scenario".to_string(),
                    "complete_tee_communication_scenario".to_string(),
                    "complete_system_scenario".to_string(),
                ],
            },
            performance_tests: PerformanceTestConfig {
                enabled: true,
                timeout_seconds: 600,
                max_concurrent: 1,
                iterations: 100,
                warmup_iterations: 10,
            },
            stress_tests: StressTestConfig {
                enabled: true,
                timeout_seconds: 1800, // 30 minutes
                max_concurrent: 1,
                iterations: 1000,
                data_sizes: vec![1024, 10240, 102400, 1024000], // 1KB, 10KB, 100KB, 1MB
            },
            global: GlobalTestConfig {
                log_level: "info".to_string(),
                output_format: "json".to_string(),
                generate_reports: true,
                cleanup_after_tests: true,
                parallel_execution: true,
            },
        }
    }
}

/// Test result for individual tests
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TestResult {
    pub test_name: String,
    pub test_type: String,
    pub duration: Duration,
    pub passed: bool,
    pub error_message: Option<String>,
    pub metrics: Option<TestMetrics>,
}

/// Test metrics for performance and stress tests
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TestMetrics {
    pub iterations: usize,
    pub average_duration: Duration,
    pub min_duration: Duration,
    pub max_duration: Duration,
    pub throughput: f64,
    pub memory_usage: Option<usize>,
}

/// Test suite result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TestSuiteResult {
    pub suite_name: String,
    pub total_tests: usize,
    pub passed_tests: usize,
    pub failed_tests: usize,
    pub total_duration: Duration,
    pub results: Vec<TestResult>,
}

impl TestSuiteResult {
    pub fn success_rate(&self) -> f64 {
        if self.total_tests == 0 {
            0.0
        } else {
            self.passed_tests as f64 / self.total_tests as f64
        }
    }
}

/// Test runner configuration
#[derive(Debug, Clone)]
pub struct TestRunnerConfig {
    pub config: TestConfig,
    pub output_dir: Option<String>,
    pub verbose: bool,
    pub fail_fast: bool,
}

impl Default for TestRunnerConfig {
    fn default() -> Self {
        Self {
            config: TestConfig::default(),
            output_dir: None,
            verbose: false,
            fail_fast: false,
        }
    }
}

/// Test utilities
pub mod test_utils {
    use super::*;
    use std::time::Instant;
    use tokio::time::timeout;

    /// Run a test with timeout
    pub async fn run_test_with_timeout<F, Fut>(
        test_name: &str,
        test_fn: F,
        timeout_duration: Duration,
    ) -> TestResult
    where
        F: FnOnce() -> Fut,
        Fut: std::future::Future<Output = Result<()>>,
    {
        let start_time = Instant::now();
        
        match timeout(timeout_duration, test_fn()).await {
            Ok(Ok(())) => {
                let duration = start_time.elapsed();
                TestResult {
                    test_name: test_name.to_string(),
                    test_type: "unit".to_string(),
                    duration,
                    passed: true,
                    error_message: None,
                    metrics: None,
                }
            }
            Ok(Err(e)) => {
                let duration = start_time.elapsed();
                TestResult {
                    test_name: test_name.to_string(),
                    test_type: "unit".to_string(),
                    duration,
                    passed: false,
                    error_message: Some(e.to_string()),
                    metrics: None,
                }
            }
            Err(_) => {
                let duration = start_time.elapsed();
                TestResult {
                    test_name: test_name.to_string(),
                    test_type: "unit".to_string(),
                    duration,
                    passed: false,
                    error_message: Some("Test timed out".to_string()),
                    metrics: None,
                }
            }
        }
    }

    /// Run performance test with metrics
    pub async fn run_performance_test<F, Fut>(
        test_name: &str,
        test_fn: F,
        iterations: usize,
        warmup_iterations: usize,
    ) -> TestResult
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = Result<()>>,
    {
        let start_time = Instant::now();
        let mut durations = Vec::new();
        
        // Warmup iterations
        for _ in 0..warmup_iterations {
            if let Err(e) = test_fn().await {
                return TestResult {
                    test_name: test_name.to_string(),
                    test_type: "performance".to_string(),
                    duration: start_time.elapsed(),
                    passed: false,
                    error_message: Some(e.to_string()),
                    metrics: None,
                };
            }
        }
        
        // Actual test iterations
        for _ in 0..iterations {
            let iteration_start = Instant::now();
            if let Err(e) = test_fn().await {
                return TestResult {
                    test_name: test_name.to_string(),
                    test_type: "performance".to_string(),
                    duration: start_time.elapsed(),
                    passed: false,
                    error_message: Some(e.to_string()),
                    metrics: None,
                };
            }
            durations.push(iteration_start.elapsed());
        }
        
        let total_duration = start_time.elapsed();
        let average_duration = durations.iter().sum::<Duration>() / durations.len() as u32;
        let min_duration = durations.iter().min().copied().unwrap_or(Duration::ZERO);
        let max_duration = durations.iter().max().copied().unwrap_or(Duration::ZERO);
        let throughput = iterations as f64 / total_duration.as_secs_f64();
        
        TestResult {
            test_name: test_name.to_string(),
            test_type: "performance".to_string(),
            duration: total_duration,
            passed: true,
            error_message: None,
            metrics: Some(TestMetrics {
                iterations,
                average_duration,
                min_duration,
                max_duration,
                throughput,
                memory_usage: None,
            }),
        }
    }

    /// Run stress test
    pub async fn run_stress_test<F, Fut>(
        test_name: &str,
        test_fn: F,
        iterations: usize,
        data_sizes: Vec<usize>,
    ) -> TestResult
    where
        F: Fn(usize) -> Fut,
        Fut: std::future::Future<Output = Result<()>>,
    {
        let start_time = Instant::now();
        let mut durations = Vec::new();
        
        for i in 0..iterations {
            let data_size = data_sizes[i % data_sizes.len()];
            let iteration_start = Instant::now();
            
            if let Err(e) = test_fn(data_size).await {
                return TestResult {
                    test_name: test_name.to_string(),
                    test_type: "stress".to_string(),
                    duration: start_time.elapsed(),
                    passed: false,
                    error_message: Some(e.to_string()),
                    metrics: None,
                };
            }
            
            durations.push(iteration_start.elapsed());
        }
        
        let total_duration = start_time.elapsed();
        let average_duration = durations.iter().sum::<Duration>() / durations.len() as u32;
        let min_duration = durations.iter().min().copied().unwrap_or(Duration::ZERO);
        let max_duration = durations.iter().max().copied().unwrap_or(Duration::ZERO);
        let throughput = iterations as f64 / total_duration.as_secs_f64();
        
        TestResult {
            test_name: test_name.to_string(),
            test_type: "stress".to_string(),
            duration: total_duration,
            passed: true,
            error_message: None,
            metrics: Some(TestMetrics {
                iterations,
                average_duration,
                min_duration,
                max_duration,
                throughput,
                memory_usage: None,
            }),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_test_config_default() {
        let config = TestConfig::default();
        assert!(config.unit_tests.enabled);
        assert!(config.integration_tests.enabled);
        assert!(config.e2e_tests.enabled);
        assert!(config.performance_tests.enabled);
        assert!(config.stress_tests.enabled);
    }

    #[test]
    fn test_test_result_creation() {
        let result = TestResult {
            test_name: "test_example".to_string(),
            test_type: "unit".to_string(),
            duration: Duration::from_millis(100),
            passed: true,
            error_message: None,
            metrics: None,
        };
        
        assert_eq!(result.test_name, "test_example");
        assert_eq!(result.test_type, "unit");
        assert_eq!(result.duration, Duration::from_millis(100));
        assert!(result.passed);
        assert!(result.error_message.is_none());
        assert!(result.metrics.is_none());
    }

    #[test]
    fn test_test_suite_result_success_rate() {
        let results = vec![
            TestResult {
                test_name: "test1".to_string(),
                test_type: "unit".to_string(),
                duration: Duration::from_millis(100),
                passed: true,
                error_message: None,
                metrics: None,
            },
            TestResult {
                test_name: "test2".to_string(),
                test_type: "unit".to_string(),
                duration: Duration::from_millis(200),
                passed: false,
                error_message: Some("Error".to_string()),
                metrics: None,
            },
        ];
        
        let suite_result = TestSuiteResult {
            suite_name: "test_suite".to_string(),
            total_tests: 2,
            passed_tests: 1,
            failed_tests: 1,
            total_duration: Duration::from_millis(300),
            results,
        };
        
        assert_eq!(suite_result.success_rate(), 0.5);
    }
}
