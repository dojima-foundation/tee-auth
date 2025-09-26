# GitHub Actions Runner Templates

This directory contains all templates and scripts for GitHub Actions runner setup and management.

## Files Overview

### 1. `user_data.sh` - Complete Runner Setup Script
A comprehensive Terraform user data script that sets up the entire runner environment.

**Features:**
- ✅ Complete system setup with all development tools
- ✅ GitHub Actions runner installation and configuration
- ✅ Docker with proper configuration
- ✅ Go 1.23.0 with development tools
- ✅ Rust 1.82.0 with Cargo and security tools
- ✅ Node.js 20.x with npm and development tools
- ✅ PostgreSQL and Redis client tools
- ✅ Playwright with browsers (Chromium, Firefox, WebKit)
- ✅ Additional CI/CD tools (GitHub CLI, kubectl, terraform, etc.)
- ✅ Built-in utility scripts for monitoring and management
- ✅ Health monitoring and status pages
- ✅ Comprehensive logging and verification

### 2. `user_data_simple.sh` - Simplified Runner Setup Script
A streamlined version for faster setup with essential tools only.

**Features:**
- ✅ Essential system packages and tools
- ✅ GitHub Actions runner installation and configuration
- ✅ Docker
- ✅ Go 1.23.0 with development tools
- ✅ Rust 1.82.0
- ✅ Node.js 20.x with Playwright
- ✅ Basic monitoring and utility scripts
- ✅ Health monitoring and status pages

### 3. `install_tools.sh` - Standalone Tools Installation Script
A comprehensive script for installing development tools on existing runners.

**Features:**
- ✅ Complete development environment setup
- ✅ All tools from user_data.sh without runner configuration
- ✅ Detailed logging and verification
- ✅ Utility scripts for testing and monitoring

### 4. `quick_install_tools.sh` - Quick Tools Installation Script
A simplified version for faster tool installation.

### 5. `check_runner_logs.sh` - Remote Runner Monitoring Script
A script for checking runner logs and status remotely via SSH.

### 6. `runner_usage_guide.md` - Comprehensive Usage Guide
Complete documentation for using and managing the self-hosted runners.

### 7. `workflow-examples.yml` - GitHub Actions Workflow Examples
Comprehensive examples of GitHub Actions workflows for different organizations and repositories.

## Usage

### For New Runner Setup (Terraform)
The `user_data.sh` script is automatically used by Terraform when creating new runner instances.

### For Existing Runner Updates

#### Option 1: Complete Installation
```bash
# Copy the script to the runner
scp -i runner_private_key.pem templates/install_tools.sh ubuntu@<runner-ip>:~/

# SSH into the runner and execute
ssh -i runner_private_key.pem ubuntu@<runner-ip>
chmod +x install_tools.sh
./install_tools.sh
```

#### Option 2: Quick Installation
```bash
# Copy the script to the runner
scp -i runner_private_key.pem templates/quick_install_tools.sh ubuntu@<runner-ip>:~/

# SSH into the runner and execute
ssh -i runner_private_key.pem ubuntu@<runner-ip>
chmod +x quick_install_tools.sh
./quick_install_tools.sh
```

### Using Terraform Make Commands
```bash
# From the terraform directory
cd /path/to/terraform

# Copy script to runner and execute
make install-tools
```

## Installed Tools

### Go Development Environment
- **Go 1.23.0** - Latest stable Go version
- **protoc-gen-go** - Protocol buffer compiler plugin for Go
- **protoc-gen-go-grpc** - gRPC plugin for Go
- **golangci-lint** - Fast Go linters runner
- **goimports** - Updates Go import lines
- **gosec** - Security analyzer for Go code
- **golang-migrate** - Database migration tool

### Rust Toolchain
- **Rust 1.82.0** - Latest stable Rust version
- **Cargo** - Rust package manager
- **cargo-audit** - Security audit tool
- **cargo-deny** - Lint tool for Cargo projects
- **cargo-tarpaulin** - Code coverage tool

### Node.js Environment
- **Node.js 20.x** - Latest LTS version
- **npm** - Node package manager
- **Playwright** - End-to-end testing framework
- **Lighthouse CI** - Performance testing
- **TypeScript** - Type-safe JavaScript
- **ts-node** - TypeScript execution engine

### Database Tools
- **PostgreSQL client** (psql, pg_isready)
- **Redis CLI** (redis-cli)
- **Database migration tools**

### CI/CD Tools
- **GitHub CLI** (gh)
- **Kubernetes CLI** (kubectl)
- **Terraform**
- **AWS CLI**
- **Azure CLI**
- **Google Cloud CLI**
- **Docker Compose**

### Built-in Utility Scripts
Each runner comes with comprehensive utility scripts installed in `/usr/local/bin/`:

#### Core Utilities
- **`runner-monitor`** - Check runner status and health
- **`runner-info`** - Display comprehensive system information
- **`runner-reconfigure`** - Reconfigure runner for different org/repo

#### Database Testing Utilities
- **`test-postgres-connection`** - Test PostgreSQL connectivity
- **`test-redis-connection`** - Test Redis connectivity
- **`db-monitor`** - Monitor database services status

#### Usage Examples
```bash
# SSH into runner
ssh -i runner_private_key.pem ubuntu@<runner-ip>

# Check runner status
runner-monitor

# View comprehensive system information
runner-info

# Setup multi-organization support
setup-multi-org-runners <github_token>

# Reconfigure for organization
runner-reconfigure <github_token> org dojima-foundation
runner-reconfigure <github_token> org dojimanetwork

# Reconfigure for repository
runner-reconfigure <github_token> repo bhaagiKenpachi/spark-park-cricket
runner-reconfigure <github_token> repo dojima-foundation/tee-auth

# Test database connections
test-postgres-connection [host] [port] [user] [password] [database]
test-redis-connection [host] [port] [password]
db-monitor
```

## Workflow Examples

The `workflow-examples.yml` file contains comprehensive examples for different organizations and repositories:

### Supported Organizations
- **dojima-foundation**: Organization-level workflows
- **dojimanetwork**: Organization-level workflows  
- **bhaagiKenpachi**: User account workflows
- **spark-park-cricket**: Repository-specific workflows

### Workflow Labels
Use these labels in your GitHub Actions workflows:

**Organization-level runners:**
```yaml
runs-on: [self-hosted, ovh, ubuntu-22.04, dojima-foundation]
runs-on: [self-hosted, ovh, ubuntu-22.04, dojimanetwork]
```

**Repository-level runners:**
```yaml
runs-on: [self-hosted, ovh, ubuntu-22.04, bhaagiKenpachi]
runs-on: [self-hosted, ovh, ubuntu-22.04, dojima-foundation]
```

### Quick Setup
1. Copy the appropriate workflow from `workflow-examples.yml`
2. Place it in your repository's `.github/workflows/` directory
3. Modify the workflow to match your specific needs
4. Ensure your runner is configured with the correct labels

## Environment Variables

The scripts set up the following environment variables:

### Go Environment
```bash
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
export GOPATH="/home/ubuntu/go"
export GOROOT="/usr/local/go"
export GOCACHE="/home/ubuntu/.cache/go-build"
export GOTMPDIR="/home/ubuntu/.cache/go-tmp"
export GOMODCACHE="/home/ubuntu/go/pkg/mod"
export GOSUMDB="sum.golang.org"
export GOPROXY="https://proxy.golang.org,direct"
```

### Rust Environment
```bash
export PATH="$HOME/.cargo/bin:$PATH"
source ~/.cargo/env
```

## Verification

After installation, you can verify the tools are working:

```bash
# Check Go tools
go version
protoc-gen-go --version
protoc-gen-go-grpc --version
golangci-lint --version

# Check database tools
psql --version
redis-cli --version
pg_isready --version

# Check Rust tools
rustc --version
cargo --version

# Check Node.js tools
node --version
npm --version
npx playwright --version

# Check Docker
docker --version

# Run comprehensive check
runner-info
```

## Troubleshooting

### Permission Issues
If you encounter permission issues with npm global packages:
```bash
sudo npm install -g <package-name>
```

### Environment Variables Not Loading
After installation, you may need to reload your shell:
```bash
source ~/.bashrc
# or
exec bash
```

### Docker Permission Issues
If Docker commands fail with permission errors:
```bash
sudo usermod -aG docker ubuntu
# Then log out and log back in
```

### Go Tools Not Found
If Go tools are not found in PATH:
```bash
export PATH="/usr/local/go/bin:/home/ubuntu/go/bin:$PATH"
```

### GitHub Actions Cache Permission Issues
If you encounter cache permission errors in GitHub Actions:
```bash
# The scripts now automatically create and configure the cache directory
sudo mkdir -p /opt/go-cache
sudo chown -R ubuntu:ubuntu /opt/go-cache
sudo chmod -R 755 /opt/go-cache
```

### Go Tools Not Available in GitHub Actions
If Go tools are not found in GitHub Actions workflows:
```bash
# The scripts now configure both ubuntu and runner users
# Check that the runner user has the correct PATH:
cat /home/runner/.bashrc | grep PATH
```

### Basic Commands Not Found
If basic system commands are not available:
```bash
# The scripts now use full PATH to ensure all commands are available
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
```

## Customization

You can customize the installation by modifying the scripts:

1. **Add new tools**: Add installation commands to the appropriate function
2. **Change versions**: Update version numbers in the download URLs
3. **Modify environment**: Adjust environment variables as needed
4. **Add verification**: Add custom verification steps

## Security Notes

- The scripts install tools with minimal required permissions
- Docker is configured with proper user groups
- SSH keys are handled securely
- No sensitive information is logged

## Recent Fixes (v2.0)

### 🔧 **PATH Configuration Issues Fixed**
- **Problem**: Go tools not found in GitHub Actions workflows
- **Solution**: Added full PATH configuration for both `ubuntu` and `runner` users
- **Impact**: All Go tools now accessible in CI/CD workflows

### 🔧 **Cache Permission Issues Fixed**
- **Problem**: `/opt/go-cache` permission denied errors
- **Solution**: Created cache directory with proper ownership and permissions
- **Impact**: GitHub Actions cache operations now work correctly

### 🔧 **User Environment Configuration**
- **Problem**: Environment variables not available to GitHub Actions runner
- **Solution**: Configured both `/home/ubuntu/.bashrc` and `/home/runner/.bashrc`
- **Impact**: Consistent environment across all users

### 🔧 **Basic Command Availability**
- **Problem**: Basic system commands not found in minimal environments
- **Solution**: Use full PATH in all script operations
- **Impact**: Scripts work in any environment configuration

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review the installation logs
3. Verify your system meets the requirements
4. Check the GitHub Actions runner logs

## Requirements

- Ubuntu 22.04 LTS
- Internet connectivity
- Sudo access
- At least 4GB RAM
- At least 20GB disk space
