# Makefile for renclave-v2

.PHONY: help build test clean docker-build docker-test docker-clean run-host run-enclave

# Default target
help:
	@echo "renclave-v2 - QEMU Nitro Enclave Seed Generation"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build all Rust binaries"
	@echo "  test          - Run Rust tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-test   - Run tests in Docker"
	@echo "  docker-clean  - Clean Docker images and containers"
	@echo "  run-host      - Run host binary"
	@echo "  run-enclave   - Run enclave binary"
	@echo "  run-both      - Run both host and enclave"

# Rust build targets
build:
	@echo "🔨 Building renclave-v2..."
	cargo build --release --workspace

test:
	@echo "🧪 Running Rust tests..."
	cargo test --workspace

clean:
	@echo "🧹 Cleaning build artifacts..."
	cargo clean

# Docker targets
docker-build:
	@echo "🐳 Building Docker image..."
	docker-compose -f docker/docker-compose.yml build

docker-test:
	@echo "🐳 Running tests in Docker..."
	docker-compose -f docker/docker-compose.yml up --abort-on-container-exit

docker-test-host-only:
	@echo "🐳 Running host-only tests in Docker..."
	docker-compose -f docker/docker-compose.yml --profile host-only up --abort-on-container-exit

docker-test-network:
	@echo "🐳 Running network tests in Docker..."
	docker-compose -f docker/docker-compose.yml --profile network-test up --abort-on-container-exit

docker-clean:
	@echo "🐳 Cleaning Docker resources..."
	docker-compose -f docker/docker-compose.yml down --volumes --remove-orphans
	docker system prune -f

# Local execution targets
run-enclave: build
	@echo "🔒 Starting enclave..."
	./target/release/enclave

run-host: build
	@echo "🏠 Starting host..."
	./target/release/host

run-both: build
	@echo "🚀 Starting both enclave and host..."
	@echo "Starting enclave in background..."
	./target/release/enclave &
	@sleep 3
	@echo "Starting host..."
	./target/release/host

# Development targets
dev-build:
	@echo "🔨 Building in development mode..."
	cargo build --workspace

dev-test:
	@echo "🧪 Running tests in development mode..."
	cargo test --workspace -- --nocapture

dev-watch:
	@echo "👀 Watching for changes..."
	cargo watch -x "build --workspace"

# Linting and formatting
fmt:
	@echo "🎨 Formatting code..."
	cargo fmt --all

clippy:
	@echo "📎 Running clippy..."
	cargo clippy --workspace --all-targets --all-features

# Documentation
docs:
	@echo "📚 Building documentation..."
	cargo doc --workspace --no-deps

docs-open:
	@echo "📚 Opening documentation..."
	cargo doc --workspace --no-deps --open

# Benchmarks (if any)
bench:
	@echo "⚡ Running benchmarks..."
	cargo bench --workspace

# Security audit
audit:
	@echo "🔒 Running security audit..."
	cargo audit

# Check all
check-all: fmt clippy test audit
	@echo "✅ All checks completed"

# Install dependencies
install-deps:
	@echo "📦 Installing dependencies..."
	cargo install cargo-watch cargo-audit

# Quick start
quick-start: docker-build docker-test
	@echo "🎉 Quick start completed!"

# Production build
prod-build:
	@echo "🏭 Building for production..."
	cargo build --release --workspace
	strip target/release/enclave target/release/host

# Create release
release: clean prod-build test
	@echo "📦 Creating release..."
	mkdir -p release
	cp target/release/enclave release/
	cp target/release/host release/
	cp -r docker release/
	cp README.md release/ 2>/dev/null || true
	tar -czf renclave-v2-release.tar.gz release/
	@echo "✅ Release created: renclave-v2-release.tar.gz"
