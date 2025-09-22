# GitHub Actions Self-Hosted Runners Usage Guide

## Current Setup ✅

**Runners Status:**
- **Runner 1**: 51.79.143.4 (runner-1-bhaagi-spark-park-cricket) - Active for `bhaagiKenpachi/spark-park-cricket`
- **Runner 2**: 15.235.203.215 (runner-2-bhaagi-spark-park-cricket) - Active for `bhaagiKenpachi/spark-park-cricket`

**Available Organizations:**
- `dojima-foundation` - Organization with access to runners
- `dojimanetwork` - Organization with access to runners
- `bhaagiKenpachi` - User account with repositories (currently configured)

## How to Use Runners with Different Repositories

### Option 1: Use Existing Runners (Recommended)

The current runners are configured for `dojima-foundation/tee-auth` but can be used by any repository in the `dojima-foundation` organization by using the appropriate labels.

#### For dojima-foundation repositories:
```yaml
name: CI/CD Pipeline
on: [push, pull_request]
jobs:
  build:
    runs-on: [self-hosted, ovh, ubuntu-22.04]
    steps:
      - uses: actions/checkout@v4
      - name: Build and Test
        run: |
          echo "Running on dojima-foundation runner"
          # Your build commands here
```

#### For dojimanetwork repositories:
```yaml
name: CI/CD Pipeline
on: [push, pull_request]
jobs:
  build:
    runs-on: [self-hosted, ovh, ubuntu-22.04]
    steps:
      - uses: actions/checkout@v4
      - name: Build and Test
        run: |
          echo "Running on dojimanetwork runner"
          # Your build commands here
```

#### For bhaagiKenpachi repositories (including spark-park-cricket):
```yaml
name: CI/CD Pipeline
on: [push, pull_request]
jobs:
  build:
    runs-on: [self-hosted, ovh, ubuntu-22.04]
    steps:
      - uses: actions/checkout@v4
      - name: Build and Test
        run: |
          echo "Running on bhaagiKenpachi runner"
          # Your build commands here
```

### Option 2: Add Runners to Specific Repositories

If you need dedicated runners for specific repositories, you can add the existing runners to those repositories manually through the GitHub UI.

#### Steps to add runners to a repository:

1. Go to the repository settings
2. Navigate to Actions → Runners
3. Click "New runner"
4. Select "Repository" (not Organization)
5. Copy the registration token
6. SSH into the runner and configure it for the specific repository

## Available Tools on Runners

The runners are equipped with comprehensive development tools:

### Go Development
- Go 1.23.0
- protoc-gen-go
- protoc-gen-go-grpc
- golangci-lint
- goimports
- gosec
- golang-migrate

### Rust Development
- Rust 1.82.0
- Cargo
- cargo-audit
- cargo-deny
- cargo-tarpaulin

### Node.js Development
- Node.js 20.x
- npm
- Playwright (with browsers)
- Lighthouse CI
- TypeScript

### Database Tools
- PostgreSQL client (psql, pg_isready)
- Redis CLI
- Database migration tools

### CI/CD Tools
- GitHub CLI
- Kubernetes CLI
- Terraform
- AWS CLI
- Docker Compose

### System Tools
- Docker
- Git
- Make
- CMake
- Build tools (gcc, g++, clang)

## Testing the Setup

### Test Workflow for spark-park-cricket

Create this workflow in your `spark-park-cricket` repository:

```yaml
name: Test Self-Hosted Runner
on: [push]
jobs:
  test:
    runs-on: [self-hosted, ovh, ubuntu-22.04]
    steps:
      - uses: actions/checkout@v4
      - name: Test Runner
        run: |
          echo "Testing runner for spark-park-cricket"
          echo "Runner OS: $(uname -a)"
          echo "Available tools:"
          which go && go version || echo "Go not available"
          which node && node --version || echo "Node not available"
          which docker && docker --version || echo "Docker not available"
          which rustc && rustc --version || echo "Rust not available"
          which psql && psql --version || echo "PostgreSQL not available"
          which redis-cli && redis-cli --version || echo "Redis not available"
```

### Test Workflow for dojima-foundation repositories

```yaml
name: Test Dojima Foundation Runner
on: [push]
jobs:
  test:
    runs-on: [self-hosted, ovh, ubuntu-22.04]
    steps:
      - uses: actions/checkout@v4
      - name: Test Runner
        run: |
          echo "Testing runner for dojima-foundation"
          echo "Runner OS: $(uname -a)"
          echo "Available tools:"
          which go && go version || echo "Go not available"
          which node && node --version || echo "Node not available"
          which docker && docker --version || echo "Docker not available"
```

## Built-in Utility Scripts

Each runner comes with comprehensive utility scripts installed in `/usr/local/bin/`:

### Core Utilities
- **`runner-monitor`**: Check runner status and health
- **`runner-info`**: Display comprehensive system information
- **`runner-reconfigure`**: Reconfigure runner for different org/repo

### Database Testing Utilities
- **`test-postgres-connection`**: Test PostgreSQL connectivity
- **`test-redis-connection`**: Test Redis connectivity
- **`db-monitor`**: Monitor database services status

### Usage Examples
```bash
# SSH into runner
ssh -i runner_private_key.pem ubuntu@51.79.143.4

# Check runner status
runner-monitor

# View comprehensive system information
runner-info

# Reconfigure for organization
runner-reconfigure <github_token> org dojima-foundation

# Reconfigure for repository
runner-reconfigure <github_token> repo user/repo

# Test database connections
test-postgres-connection [host] [port] [user] [password] [database]
test-redis-connection [host] [port] [password]
db-monitor
```

## Monitoring and Troubleshooting

### Check Runner Status
- **GitHub UI**: https://github.com/bhaagiKenpachi/spark-park-cricket/settings/actions/runners
- **SSH Access**: `ssh -i runner_private_key.pem ubuntu@51.79.143.4`
- **Status Pages**: http://51.79.143.4/ and http://15.235.203.215/
- **Built-in Scripts**: Use `runner-monitor` and `runner-info` on the runners

### Common Issues

1. **Runner not appearing in GitHub**:
   - Check if the runner service is running
   - Verify network connectivity
   - Check GitHub token permissions

2. **Workflow not using the runner**:
   - Ensure the correct labels are specified
   - Check if the runner is online
   - Verify the repository has access to the runner

3. **Permission issues**:
   - Check file permissions on the runner
   - Verify user permissions
   - Check systemd service status

### Useful Commands

```bash
# Check runner status
ssh -i runner_private_key.pem ubuntu@51.79.143.4 'systemctl status actions.runner.*'

# Check runner logs
ssh -i runner_private_key.pem ubuntu@51.79.143.4 'journalctl -u actions.runner.* -f'

# Restart runner service
ssh -i runner_private_key.pem ubuntu@51.79.143.4 'sudo systemctl restart actions.runner.*'
```

## Security Notes

- Keep the SSH private key secure
- Regularly rotate GitHub tokens
- Monitor runner access and usage
- Use organization-level permissions when possible
- Keep runners updated with security patches

## Cost Optimization

- Use appropriate instance sizes for workloads
- Monitor resource usage
- Scale runners based on demand
- Use spot instances for non-critical workloads

## Next Steps

1. **Test the setup** with the provided workflow examples
2. **Monitor runner usage** through GitHub Actions
3. **Scale as needed** based on workload requirements
4. **Set up monitoring** for runner health and performance
5. **Document** any custom configurations for your team

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review runner logs and status
3. Verify GitHub token permissions
4. Check network connectivity
5. Contact the infrastructure team
