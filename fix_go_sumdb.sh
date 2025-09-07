#!/bin/bash

# Fix Go SumDB Permission Issues
# This script fixes permission issues with Go sum database

set -e

echo "🔧 Fixing Go SumDB permission issues..."

# Create sumdb directory in shared cache
sudo mkdir -p /opt/go-cache/go-sumdb
sudo chown -R ubuntu:ubuntu /opt/go-cache/go-sumdb
sudo chmod -R 775 /opt/go-cache/go-sumdb

# Set up Go environment with shared sumdb
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/opt/go-cache/go-build"
export GOTMPDIR="/opt/go-cache/go-tmp"
export GOMODCACHE="/opt/go-cache/go-mod"
export GOSUMDB="sum.golang.org"
export GOPROXY="https://proxy.golang.org,direct"

# Update system-wide profile with sumdb settings
echo 'export GOSUMDB="sum.golang.org"' >> /etc/profile
echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /etc/profile

# Update ubuntu user's bashrc
echo 'export GOSUMDB="sum.golang.org"' >> /home/ubuntu/.bashrc
echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/ubuntu/.bashrc

# Update github-runner user's bashrc
if [ -d "/home/github-runner" ]; then
    echo 'export GOSUMDB="sum.golang.org"' >> /home/github-runner/.bashrc
    echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/github-runner/.bashrc
fi

# Also ensure the local sumdb directory has proper permissions
sudo mkdir -p /home/ubuntu/go/pkg/sumdb
sudo chown -R ubuntu:ubuntu /home/ubuntu/go/pkg/sumdb
sudo chmod -R 775 /home/ubuntu/go/pkg/sumdb

# Test Go module operations
echo "Testing Go module operations with sumdb..."
cd /tmp
mkdir -p test-go-sumdb
cd test-go-sumdb
go mod init test-sumdb
go get github.com/stretchr/testify
go mod tidy
echo "✅ Go module operations with sumdb successful"

# Clean up test
cd /
rm -rf /tmp/test-go-sumdb

echo "🎉 Go SumDB permission issues fixed successfully!"
echo "Go sum database operations should now work correctly."
