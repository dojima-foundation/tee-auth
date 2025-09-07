#!/bin/bash

# Update GitHub Actions Runner Go Environment and Protobuf Plugins
# This script fixes the issues identified in the gauth.yml workflow

set -e

echo "🔧 Updating GitHub Actions Runner Go Environment..."

# Set up Go environment
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/home/ubuntu/.cache/go-build"
export GOTMPDIR="/home/ubuntu/.cache/go-tmp"

# Create necessary directories
mkdir -p /home/ubuntu/go /home/ubuntu/.cache/go-build /home/ubuntu/.cache/go-tmp

# Update system-wide profile
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /etc/profile
echo 'export GOPATH="/home/ubuntu/go"' >> /etc/profile
echo 'export GOROOT="/usr/local/go"' >> /etc/profile
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /etc/profile
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /etc/profile

# Update ubuntu user's bashrc
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'export GOPATH="/home/ubuntu/go"' >> /home/ubuntu/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/ubuntu/.bashrc
echo 'export GOCACHE="/home/ubuntu/.cache/go-build"' >> /home/ubuntu/.bashrc
echo 'export GOTMPDIR="/home/ubuntu/.cache/go-tmp"' >> /home/ubuntu/.bashrc

# Install/update protobuf plugins
echo "Installing protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify installation
echo "Verifying Go environment:"
echo "Go version: $(go version)"
echo "Go path: $(which go)"
echo "GOPATH: $GOPATH"
echo "GOROOT: $GOROOT"

echo "Verifying protobuf plugins:"
which protoc-gen-go && echo "✅ protoc-gen-go found" || echo "❌ protoc-gen-go not found"
which protoc-gen-go-grpc && echo "✅ protoc-gen-go-grpc found" || echo "❌ protoc-gen-go-grpc not found"

# Test protobuf generation
echo "Testing protobuf generation..."
cd /tmp
cat > test.proto << 'EOF'
syntax = "proto3";
package test;
option go_package = "./test";

message TestMessage {
  optional string name = 1;
}
EOF

protoc --experimental_allow_proto3_optional --go_out=. test.proto
if [ -f "test/test.pb.go" ]; then
    echo "✅ Protobuf generation test successful"
    rm -rf test test.proto
else
    echo "❌ Protobuf generation test failed"
fi

echo "🎉 Go environment and protobuf plugins updated successfully!"
echo "The runner is now ready for the gauth.yml workflow."
