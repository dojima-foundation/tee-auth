#!/bin/bash

# Fix GitHub Runner User Permissions for Go Operations
# This script ensures the github-runner user has proper access to Go directories

set -e

echo "🔧 Fixing GitHub Runner user permissions for Go operations..."

# Create shared directories with proper ownership
sudo mkdir -p /opt/go-cache/go-build
sudo mkdir -p /opt/go-cache/go-tmp
sudo mkdir -p /opt/go-cache/go-mod
sudo mkdir -p /opt/go-cache/go-sumdb

# Set ownership to github-runner user (the user that runs GitHub Actions)
sudo chown -R github-runner:github-runner /opt/go-cache
sudo chmod -R 755 /opt/go-cache

# Also ensure the ubuntu user's Go directories are accessible
sudo mkdir -p /home/ubuntu/go/pkg/sumdb
sudo chown -R github-runner:github-runner /home/ubuntu/go/pkg/sumdb
sudo chmod -R 755 /home/ubuntu/go/pkg/sumdb

# Create a symlink from ubuntu's sumdb to the shared location
sudo rm -rf /home/ubuntu/go/pkg/sumdb
sudo ln -sf /opt/go-cache/go-sumdb /home/ubuntu/go/pkg/sumdb

# Update github-runner user's environment
if [ -d "/home/github-runner" ]; then
    # Set up Go environment for github-runner user
    echo 'export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"' >> /home/github-runner/.bashrc
    echo 'export GOPATH="/home/ubuntu/go"' >> /home/github-runner/.bashrc
    echo 'export GOROOT="/usr/local/go"' >> /home/github-runner/.bashrc
    echo 'export GOCACHE="/opt/go-cache/go-build"' >> /home/github-runner/.bashrc
    echo 'export GOTMPDIR="/opt/go-cache/go-tmp"' >> /home/github-runner/.bashrc
    echo 'export GOMODCACHE="/opt/go-cache/go-mod"' >> /home/github-runner/.bashrc
    echo 'export GOSUMDB="sum.golang.org"' >> /home/github-runner/.bashrc
    echo 'export GOPROXY="https://proxy.golang.org,direct"' >> /home/github-runner/.bashrc
    echo 'export TMPDIR="/opt/go-cache"' >> /home/github-runner/.bashrc
fi

# Test as github-runner user
echo "Testing Go operations as github-runner user..."
sudo -u github-runner bash -c "
export PATH='/usr/local/go/bin:/home/ubuntu/go/bin:\$PATH'
export GOPATH='/home/ubuntu/go'
export GOROOT='/usr/local/go'
export GOCACHE='/opt/go-cache/go-build'
export GOTMPDIR='/opt/go-cache/go-tmp'
export GOMODCACHE='/opt/go-cache/go-mod'
export GOSUMDB='sum.golang.org'
export GOPROXY='https://proxy.golang.org,direct'

cd /tmp
mkdir -p test-github-runner-go
cd test-github-runner-go
go mod init test-github-runner
go get github.com/stretchr/testify
go mod tidy
echo '✅ Go operations successful as github-runner user'
"

# Clean up test
sudo rm -rf /tmp/test-github-runner-go

echo "🎉 GitHub Runner user permissions fixed successfully!"
echo "The github-runner user can now access Go sumdb and perform all Go operations."
