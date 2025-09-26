# Multi-stage Dockerfile for QEMU Nitro Enclave testing
FROM rust:1.82-bookworm AS builder

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy source code
COPY . .

# Build all workspace members
RUN cargo build --release --workspace

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    curl \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Create directories
RUN mkdir -p /app/bin /app/scripts /tmp

# Copy binaries from builder
COPY --from=builder /workspace/target/release/enclave /app/bin/
COPY --from=builder /workspace/target/release/host /app/bin/

# Copy scripts
COPY docker/scripts/ /app/scripts/

# Set executable permissions
RUN chmod +x /app/bin/* /app/scripts/* || true

WORKDIR /app

# Default command starts the services
CMD ["/app/scripts/start-services.sh"]
