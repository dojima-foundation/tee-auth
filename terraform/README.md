# GitHub Actions Runners on OVH Cloud

This Terraform configuration provides a modular, scalable, and production-ready setup for self-hosted GitHub Actions runners on OVH Cloud. The architecture supports multiple environments, comprehensive monitoring, and automated management.

## 🏗️ Architecture

```
terraform/
├── environments/           # Environment-specific configurations
│   ├── dev/              # Development environment (1 runner)
│   ├── staging/          # Staging environment (2 runners)
│   └── prod/             # Production environment (2+ runners)
├── modules/              # Reusable Terraform modules
│   ├── ovh-infrastructure/  # OVH Cloud infrastructure setup
│   ├── github-runner/       # GitHub Actions runner configuration
│   └── monitoring/          # Monitoring and observability
├── scripts/              # Management and monitoring scripts
│   ├── setup/            # Setup and deployment scripts
│   ├── monitoring/       # Monitoring and health check scripts
│   └── cleanup/          # Cleanup and maintenance scripts
└── configs/              # Configuration templates and examples
    ├── templates/        # Configuration templates
    └── examples/         # Example configurations
```

## ✨ Features

- ✅ **Multi-Environment Support**: Separate configurations for dev, staging, and production
- ✅ **Modular Architecture**: Reusable modules for infrastructure, runners, and monitoring
- ✅ **Automated Setup**: Complete automation of runner installation and configuration
- ✅ **Health Monitoring**: Built-in health checks and monitoring with status pages
- ✅ **Scalable**: Support for multiple runner instances with load balancing
- ✅ **Security**: SSH key management, firewall configuration, and secure communication
- ✅ **Monitoring**: Status pages, health checks, and optional Grafana/Prometheus integration
- ✅ **Cost Optimization**: Environment-specific instance sizing and auto-scaling support

## 🚀 Quick Start

### Prerequisites

1. **OVH Cloud Account**: Create an account at [https://www.ovhcloud.com/](https://www.ovhcloud.com/)
2. **GitHub Personal Access Token**: Generate with `repo` and `admin:org` permissions
3. **Terraform**: Install Terraform >= 1.0
4. **OpenStack CLI**: Install for instance management (optional)

### 1. Configure Environment

Choose your environment and configure the variables:

```bash
# For production environment
cd environments/prod
cp terraform.tfvars.example terraform.tfvars
nano terraform.tfvars
```

### 2. Deploy Infrastructure

```bash
# Deploy to production
make prod

# Or deploy to development
make dev

# Or deploy to staging
make staging
```

### 3. Verify Deployment

```bash
# Check status
make status

# View logs
make logs

# SSH into a runner
ssh -i runner_private_key.pem ubuntu@<runner-ip>
```

## 📋 Configuration

### Environment Variables

Each environment has its own configuration file (`terraform.tfvars`):

```hcl
# OVH Cloud Configuration
ovh_endpoint = "ovh-ca"
ovh_application_key = "your_key"
ovh_application_secret = "your_secret"
ovh_consumer_key = "your_consumer_key"

# Project Configuration
project_id = "your_project_id"
region = "SGP1"

# GitHub Configuration
github_token = "ghp_your_token"
github_repo = "owner/repository"

# Runner Configuration
runner_count = 2
runner_labels = ["ovh", "self-hosted", "ubuntu-22.04"]
runner_flavor_id = "b2-7"
```

### Environment-Specific Settings

| Environment | Runners | Instance Size | Load Balancer | Monitoring |
|-------------|---------|---------------|---------------|------------|
| Development | 1 | b2-7 | No | Basic |
| Staging | 2 | b2-7 | Optional | Enhanced |
| Production | 2+ | b2-15+ | Yes | Full Stack |

## 🛠️ Management Commands

### Environment Management

```bash
# Deploy environments
make dev          # Deploy to development
make staging      # Deploy to staging
make prod         # Deploy to production

# Check status
make dev-status   # Development status
make prod-status  # Production status

# View logs
make dev-logs     # Development logs
make prod-logs    # Production logs

# Destroy environments
make dev-destroy  # Destroy development
make prod-destroy # Destroy production
```

### General Commands

```bash
# Initialize and deploy
make setup        # Complete setup (init + plan + apply)

# Validation and formatting
make validate     # Validate all configurations
make format       # Format Terraform files

# Monitoring
make status       # Show current status
make logs         # Show runner logs
make ssh-key      # Generate SSH key file

# Cleanup
make clean        # Clean up Terraform files
```

## 📊 Monitoring

### Status Pages

Each runner includes a web-based status page:
- **URL**: `http://<runner-ip>/`
- **Features**: Real-time status, system metrics, health checks
- **Auto-refresh**: Updates every 30 seconds

### Health Checks

Automated health checks run every 5 minutes:
- Service status monitoring
- Network connectivity tests
- Resource usage monitoring
- Automatic service restart on failure

### Advanced Monitoring (Optional)

Enable full monitoring stack:

```hcl
# In terraform.tfvars
enable_grafana = true
enable_prometheus = true
enable_alerting = true
```

## 🔧 Troubleshooting

### Common Issues

1. **Runner not appearing in GitHub**:
   ```bash
   # Check GitHub token permissions
   make logs
   
   # Verify repository access
   curl -H "Authorization: token $GITHUB_TOKEN" \
        https://api.github.com/repos/owner/repo
   ```

2. **Runner offline**:
   ```bash
   # Check service status
   ssh -i runner_private_key.pem ubuntu@<ip> 'sudo systemctl status actions.runner.*'
   
   # Restart service
   ssh -i runner_private_key.pem ubuntu@<ip> 'sudo systemctl restart actions.runner.*'
   ```

3. **SSH connection issues**:
   ```bash
   # Generate SSH key
   make ssh-key
   
   # Test connection
   ssh -i runner_private_key.pem ubuntu@<ip>
   ```

### Debug Commands

```bash
# Check runner status
./scripts/monitoring/check_runner_logs.sh <runner-ip>

# View system information
ssh -i runner_private_key.pem ubuntu@<ip> 'runner-info'

# Monitor runner processes
ssh -i runner_private_key.pem ubuntu@<ip> 'runner-monitor'
```

## 💰 Cost Optimization

### Instance Sizing

| Flavor | vCPUs | RAM | Storage | Use Case | Cost/Month |
|--------|-------|-----|---------|----------|------------|
| b2-7 | 2 | 7GB | 50GB | Development | ~$15 |
| b2-15 | 4 | 15GB | 100GB | Production | ~$30 |
| b2-30 | 8 | 30GB | 200GB | High Load | ~$60 |

### Cost-Saving Tips

1. **Use appropriate instance sizes** for each environment
2. **Enable auto-scaling** for variable workloads
3. **Use spot instances** for non-critical workloads
4. **Monitor usage** and adjust resources accordingly

## 🔒 Security

### Security Features

- ✅ **SSH Key Management**: Automatic generation and secure storage
- ✅ **Firewall Configuration**: UFW with minimal required ports
- ✅ **User Isolation**: Dedicated runner user with limited permissions
- ✅ **Secure Communication**: HTTPS for all GitHub communication
- ✅ **Network Security**: Private networks and security groups

### Best Practices

1. **Keep SSH keys secure**: Don't share `runner_private_key.pem`
2. **Regular updates**: Keep Ubuntu and Docker updated
3. **Monitor access**: Check SSH logs regularly
4. **Backup configuration**: Save your Terraform state

## 📚 Advanced Configuration

### Custom Runner Labels

```hcl
runner_labels = ["ovh", "self-hosted", "ubuntu-22.04", "custom-label"]
```

### Load Balancer Setup

```hcl
create_load_balancer = true
```

### Docker Registry Mirroring

```hcl
docker_registry_mirror = "https://registry-1.docker.io"
```

### Custom Network Configuration

```hcl
vlan_id = 1234
subnet_network = "192.168.1.0/24"
subnet_start = "192.168.1.10"
subnet_end = "192.168.1.100"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test in development environment
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the runner logs and health check logs
3. Verify your OVH and GitHub credentials
4. Check the [OVH Cloud documentation](https://docs.ovh.com/gb/en/public-cloud/)
5. Check the [GitHub Actions documentation](https://docs.github.com/en/actions/hosting-your-own-runners)

## 🎯 Roadmap

- [ ] Auto-scaling based on queue length
- [ ] Multi-region deployment support
- [ ] Integration with GitHub Enterprise
- [ ] Advanced monitoring with Grafana dashboards
- [ ] Cost optimization recommendations
- [ ] Automated backup and disaster recovery