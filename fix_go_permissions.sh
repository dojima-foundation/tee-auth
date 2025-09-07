#!/bin/bash

# Fix Go Environment Permissions for GitHub Actions Runner
# This script fixes permission issues for Go module cache and directories

set -e

echo "🔧 Fixing Go environment permissions for GitHub Actions runner..."

# Create directories with proper permissions
sudo mkdir -p /home/ubuntu/go/pkg/mod
sudo mkdir -p /home/ubuntu/.cache/go-build
sudo mkdir -p /home/ubuntu/.cache/go-tmp

# Set ownership to ubuntu user and github-runner group
sudo chown -R ubuntu:ubuntu /home/ubuntu/go
sudo chown -R ubuntu:ubuntu /home/ubuntu/.cache

# Add github-runner user to ubuntu group
sudo usermod -a -G ubuntu github-runner

# Set group permissions for Go directories
sudo chmod -R 775 /home/ubuntu/go
sudo chmod -R 775 /home/ubuntu/.cache

# Create a shared Go cache directory that both users can access
sudo mkdir -p /opt/go-cache
sudo chown -R ubuntu:ubuntu /opt/go-cache
sudo chmod -R 775 /opt/go-cache

# Update Go environment to use shared cache
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/opt/go-cache/go-build"
export GOTMPDIR="/opt/go-cache/go-tmp"
export GOMODCACHE="/opt/go-cache/go-mod"

# Update system-wide profile with shared cache paths
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /etc/profile
echo 'export GOPATH="/home/ubuntu/go"' >> /etc/profile
echo 'export GOROOT="/usr/local/go"' >> /etc/profile
echo 'export GOCACHE="/opt/go-cache/go-build"' >> /etc/profile
echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /etc/profile
echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /etc/profile

# Update ubuntu user's bashrc
echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/ubuntu/.bashrc
echo 'export GOPATH="/home/ubuntu/go"' >> /home/ubuntu/.bashrc
echo 'export GOROOT="/usr/local/go"' >> /home/ubuntu/.bashrc
echo 'export GOCACHE="/opt/go-cache/go-build"' >> /home/ubuntu/.bashrc
echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /home/ubuntu/.bashrc
echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /home/ubuntu/.bashrc

# Update github-runner user's bashrc
if [ -d "/home/github-runner" ]; then
    echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/github-runner/.bashrc
    echo 'export GOPATH="/home/ubuntu/go"' >> /home/github-runner/.bashrc
    echo 'export GOROOT="/usr/local/go"' >> /home/github-runner/.bashrc
    echo 'export GOCACHE="/opt/go-cache/go-build"' >> /home/github-runner/.bashrc
    echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /home/github-runner/.bashrc
    echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /home/github-runner/.bashrc
fi

# Create the shared cache directories
sudo mkdir -p /opt/go-cache/go-build
sudo mkdir -p /opt/go-cache/go-tmp
sudo mkdir -p /opt/go-cache/go-mod

# Set proper ownership and permissions
sudo chown -R ubuntu:ubuntu /opt/go-cache
sudo chmod -R 775 /opt/go-cache

# Verify Go environment
echo "Verifying Go environment:"
echo "Go version: $(go version)"
echo "Go path: $(which go)"
echo "GOPATH: $GOPATH"
echo "GOROOT: $GOROOT"
echo "GOCACHE: $GOCACHE"
echo "GOTMPDIR: $GOTMPDIR"
echo "GOMODCACHE: $GOMODCACHE"

# Test Go module operations
echo "Testing Go module operations..."
cd /tmp
mkdir -p test-go-module
cd test-go-module
go mod init test-module
go get github.com/stretchr/testify
go mod tidy
echo "✅ Go module operations successful"

# Clean up test
cd /
rm -rf /tmp/test-go-module

echo "🎉 Go environment permissions fixed successfully!"
echo "All Go cache directories now use shared paths with proper permissions."
