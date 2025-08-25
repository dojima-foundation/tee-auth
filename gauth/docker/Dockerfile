# Multi-stage build for gauth service
# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.serviceVersion=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o bin/gauth \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S gauth && \
    adduser -u 1001 -S gauth -G gauth

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/gauth .

# Copy configuration files
COPY --from=builder /app/migrations ./migrations

# Change ownership
RUN chown -R gauth:gauth /app

# Switch to non-root user
USER gauth

# Expose ports
EXPOSE 9090 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./gauth"]
