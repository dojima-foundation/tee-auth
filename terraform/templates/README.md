# GitHub Actions Runner Tools Installation Scripts

This directory contains scripts to install all required development tools on GitHub Actions runner instances for the tee-auth project.

## Scripts Overview

### 1. `install_tools.sh` - Complete Installation Script
A comprehensive script that installs all required tools with detailed logging and verification.

**Features:**
- ✅ System packages (curl, wget, git, build-essential, etc.)
- ✅ Docker with proper configuration
- ✅ Go 1.23.0 with development tools
- ✅ Rust 1.82.0 with Cargo and security tools
- ✅ Node.js 20.x with npm and development tools
- ✅ PostgreSQL and Redis client tools
- ✅ Playwright with browsers (Chromium, Firefox, WebKit)
- ✅ Additional CI/CD tools (GitHub CLI, kubectl, terraform, etc.)
- ✅ Utility scripts for testing and monitoring
- ✅ Comprehensive verification and logging
- ✅ **Fixed PATH configuration for GitHub Actions**
- ✅ **Fixed cache permission issues**
- ✅ **Configured both ubuntu and runner users**

### 2. `quick_install_tools.sh` - Quick Installation Script
A simplified version for faster installation with essential tools only.

**Features:**
- ✅ Essential system packages
- ✅ Docker
- ✅ Go 1.23.0 with development tools
- ✅ Rust 1.82.0
- ✅ Node.js 20.x with Playwright
- ✅ Basic verification
- ✅ **Fixed PATH configuration for GitHub Actions**
- ✅ **Fixed cache permission issues**
- ✅ **Configured both ubuntu and runner users**

### 3. `user_data.sh` - Terraform User Data Script
The original user data script used by Terraform for initial runner setup.

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

### Utility Scripts
- **test-postgres-connection** - Test PostgreSQL connectivity
- **test-redis-connection** - Test Redis connectivity
- **db-monitor** - Monitor database services
- **runner-info** - Display runner information

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
