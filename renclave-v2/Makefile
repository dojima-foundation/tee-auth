# Renclave-v2 Makefile
# Build and test automation for QEMU Nitro Enclave

# Variables
CARGO = cargo
DOCKER = docker
DOCKER_COMPOSE = docker compose
PROJECT_NAME = renclave-v2
RUST_VERSION = 1.82

# Directories
SRC_DIR = src
TEST_DIR = tests
DOCKER_DIR = docker
SCRIPTS_DIR = scripts

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Show this help message
	@echo "🚀 Renclave-v2 Build and Test Commands"
	@echo "======================================"
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  build          - Build all crates in release mode"
	@echo "  build-debug    - Build all crates in debug mode"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  test           - Run all tests (unit locally, integration/e2e in Docker)"
	@echo "  test-unit      - Run unit tests locally only"
	@echo "  test-integration - Run integration tests in Docker (skip unit tests)"
	@echo "  test-e2e       - Run E2E tests in Docker (skip unit tests)"
	@echo "  test-integration-only - Run ONLY integration tests (skip unit and E2E)"
	@echo "  test-e2e-only  - Run ONLY E2E tests (skip unit and integration)"
	@echo "  test-docker    - Run only Docker-based tests (integration + E2E)"
	@echo "  test-force     - Force run tests even if previous ones failed"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-build   - Build Docker images for testing"
	@echo "  docker-up      - Start Docker services (enclave, host, test-runner)"
	@echo "  docker-down    - Stop and remove Docker services"
	@echo "  docker-logs    - View logs for Docker services"
	@echo "  docker-clean   - Clean Docker images and volumes"
	@echo ""

# Build targets
build: ## Build all crates in release mode
	@echo "🔨 Building all crates in release mode..."
	$(CARGO) build --release --workspace
	@echo "✅ Build completed!"

build-debug: ## Build all crates in debug mode
	@echo "🔨 Building all crates in debug mode..."
	$(CARGO) build --workspace
	@echo "✅ Build completed!"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	$(CARGO) clean
	@echo "✅ Clean completed!"

# Testing targets
test: ## Run all tests (hybrid: unit local + integration/E2E Docker)
	@echo "🧪 Running all tests with hybrid strategy..."
	@echo "📋 Strategy: Unit tests (local) + Integration/E2E tests (Docker)"
	$(shell pwd)/scripts/run-tests.sh

test-unit: ## Run unit tests locally only
	@echo "🧪 Running unit tests locally..."
	$(shell pwd)/scripts/run-tests.sh --skip-docker

test-integration: ## Run integration tests in Docker (skip unit tests)
	@echo "🔗 Running integration tests in Docker (skipping unit tests)..."
	$(shell pwd)/scripts/run-tests.sh --skip-unit

test-e2e: ## Run E2E tests in Docker (skip unit tests)
	@echo "🚀 Running E2E tests in Docker (skipping unit tests)..."
	$(shell pwd)/scripts/run-tests.sh --skip-unit --skip-integration

test-integration-only: ## Run ONLY integration tests (skip unit and E2E)
	@echo "🔗 Running ONLY integration tests..."
	$(shell pwd)/scripts/run-tests.sh --integration-only

test-e2e-only: ## Run ONLY E2E tests (skip unit and integration)
	@echo "🚀 Running ONLY E2E tests..."
	$(shell pwd)/scripts/run-tests.sh --e2e-only

test-docker: ## Run only Docker-based tests (integration + E2E)
	@echo "🐳 Running Docker-based tests only..."
	$(shell pwd)/scripts/run-tests.sh --docker-only

test-force: ## Force run tests even if previous ones failed
	@echo "⚡ Force running tests..."
	$(shell pwd)/scripts/run-tests.sh --force

# Docker targets
docker-build: ## Build Docker images
	@echo "🐳 Building Docker images..."
	$(DOCKER_COMPOSE) -f $(DOCKER_DIR)/docker-compose.test.yml build
	@echo "✅ Docker build completed!"

docker-up: ## Start Docker services
	@echo "🚀 Starting Docker services..."
	$(DOCKER_COMPOSE) -f $(DOCKER_DIR)/docker-compose.test.yml up -d
	@echo "✅ Docker services started!"

docker-down: ## Stop Docker services
	@echo "🛑 Stopping Docker services..."
	$(DOCKER_COMPOSE) -f $(DOCKER_DIR)/docker-compose.test.yml down
	@echo "✅ Docker services stopped!"

docker-logs: ## Show Docker service logs
	@echo "📋 Showing Docker service logs..."
	$(DOCKER_COMPOSE) -f $(DOCKER_DIR)/docker-compose.test.yml logs -f

docker-clean: ## Clean up Docker environment
	@echo "🧹 Cleaning up Docker environment..."
	$(DOCKER_COMPOSE) -f $(DOCKER_DIR)/docker-compose.test.yml down --volumes --remove-orphans
	@echo "✅ Docker cleanup completed!"

# Development targets
check: ## Check code without building
	@echo "🔍 Checking code..."
	$(CARGO) check --workspace
	@echo "✅ Code check completed!"

fmt: ## Format code with rustfmt
	@echo "✨ Formatting code..."
	$(CARGO) fmt --all
	@echo "✅ Code formatting completed!"

clippy: ## Run clippy linter
	@echo "🔍 Running clippy linter..."
	$(CARGO) clippy --workspace -- -D warnings
	@echo "✅ Clippy check completed!"

# Performance and Benchmark targets
bench: ## Run all benchmarks
	@echo "🚀 Running all benchmarks..."
	$(CARGO) bench --workspace
	@echo "✅ All benchmarks completed!"

bench-seed: ## Run seed generation/validation benchmarks
	@echo "🌱 Running seed generation and validation benchmarks..."
	$(CARGO) bench --bench seed_generation --bench seed_validation
	@echo "✅ Seed benchmarks completed!"

bench-concurrent: ## Run concurrent operations benchmarks
	@echo "⚡ Running concurrent operations benchmarks..."
	$(CARGO) bench --bench concurrent_operations
	@echo "✅ Concurrent operations benchmarks completed!"

bench-stress: ## Run stress test benchmarks
	@echo "💪 Running stress test benchmarks..."
	$(CARGO) bench --bench stress_tests
	@echo "✅ Stress test benchmarks completed!"

performance: ## Run comprehensive performance tests with reports
	@echo "🚀 Running comprehensive performance tests..."
	$(shell pwd)/scripts/run-performance-docker.sh performance
	@echo "✅ Performance tests completed!"

# Coverage targets
coverage: ## Generate comprehensive test coverage reports
	@echo "📊 Generating comprehensive test coverage reports..."
	$(shell pwd)/scripts/run-performance-docker.sh coverage
	@echo "✅ Coverage reports generated!"

coverage-html: ## Generate HTML coverage reports
	@echo "📊 Generating HTML coverage reports..."
	cargo llvm-cov --workspace --html --output-dir coverage/html
	@echo "✅ HTML coverage reports generated!"

coverage-json: ## Generate JSON coverage data
	@echo "📊 Generating JSON coverage data..."
	cargo llvm-cov --workspace --json --output-dir coverage/json
	@echo "✅ JSON coverage data generated!"

coverage-lcov: ## Generate LCOV coverage data
	@echo "📊 Generating LCOV coverage data..."
	cargo llvm-cov --workspace --lcov --output-dir coverage/lcov
	@echo "✅ LCOV coverage data generated!"



# Utility targets
status: ## Show project status
	@echo "📋 Project Status"
	@echo "================="
	@echo "Project: $(PROJECT_NAME)"
	@echo "Rust Version: $(RUST_VERSION)"
	@echo "Source Directory: $(SRC_DIR)"
	@echo "Test Directory: $(TEST_DIR)"
	@echo "Docker Directory: $(DOCKER_DIR)"
	@echo "Scripts Directory: $(SCRIPTS_DIR)"
	@echo ""
	@echo "🔍 Checking Rust version..."
	@rustc --version || echo "❌ Rust not found"
	@echo ""
	@echo "🐳 Checking Docker..."
	@$(DOCKER) --version || echo "❌ Docker not found"
	@echo ""
	@echo "📦 Checking Cargo..."
	@$(CARGO) --version || echo "❌ Cargo not found"

# Phony targets
.PHONY: help build build-debug clean test test-unit test-integration test-e2e \
        test-integration-only test-e2e-only test-docker test-force \
        docker-build docker-up docker-down docker-logs docker-clean \
        check fmt clippy bench bench-seed bench-concurrent bench-stress performance \
        coverage coverage-html coverage-json coverage-lcov status
