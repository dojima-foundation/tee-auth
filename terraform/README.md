# GitHub Actions Runners on OVH Cloud

This Terraform configuration sets up self-hosted GitHub Actions runners on OVH Cloud instances. The setup includes automatic runner registration, monitoring, health checks, and graceful shutdown capabilities.

## Features

- ✅ **Automated Setup**: Complete automation of runner installation and configuration
- ✅ **Multiple Runners**: Support for multiple runner instances
- ✅ **Organization & Repository Level**: Support for both organization and repository-level runners
- ✅ **Health Monitoring**: Built-in health checks and monitoring
- ✅ **Graceful Shutdown**: Proper cleanup when instances are terminated
- ✅ **Docker Support**: Pre-installed Docker with optional registry mirroring
- ✅ **Security**: SSH key management and firewall configuration
- ✅ **Status Page**: Web-based status monitoring page
- ✅ **Load Balancing**: Optional load balancer setup

## Prerequisites

### OVH Cloud Account
1. Create an OVH Cloud account at [https://www.ovhcloud.com/](https://www.ovhcloud.com/)
2. Generate API credentials:
   - Go to [https://api.ovh.com/createToken/](https://api.ovh.com/createToken/)
   - Create a new application with the following rights:
     - GET /cloud/project
     - GET /cloud/project/*
     - POST /cloud/project/*
     - PUT /cloud/project/*
     - DELETE /cloud/project/*
   - Note down the Application Key, Application Secret, and Consumer Key

### GitHub Account
1. Create a GitHub Personal Access Token:
   - Go to [https://github.com/settings/tokens](https://github.com/settings/tokens)
   - Generate a new token with the following permissions:
     - `repo` (Full control of private repositories)
     - `admin:org` (Full control of organizations and teams)
     - `workflow` (Update GitHub Action workflows)

## Quick Start

### 1. Clone and Configure

```bash
# Navigate to the terraform directory
cd terraform

# Copy the example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit the configuration with your values
nano terraform.tfvars
```

### 2. Configure Variables

Edit `terraform.tfvars` with your specific values:

```hcl
# OVH Configuration
ovh_endpoint = "ovh-eu"
ovh_application_key = "your_ovh_application_key"
ovh_application_secret = "your_ovh_application_secret"
ovh_consumer_key = "your_ovh_consumer_key"

# GitHub Configuration
github_token = "ghp_your_github_personal_access_token"

# For organization-level runners:
github_org = "your_organization_name"
github_repo = ""

# For repository-level runners:
# github_org = ""
# github_repo = "owner/repository_name"

# Runner Configuration
runner_count = 2
region = "GRA11"
runner_flavor_id = "b2-7"  # 2 vCPUs, 7GB RAM
```

### 3. Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the configuration
terraform apply
```

### 4. Verify Deployment

After deployment, you can:

1. **Check runner status in GitHub**:
   - Go to your repository/organization settings
   - Navigate to Actions → Runners
   - Verify that your runners are online

2. **Access status pages**:
   - Each runner has a status page at `http://<runner-ip>/runner-status.html`
   - Use the output from Terraform to get the IP addresses

3. **Monitor runners**:
   ```bash
   # SSH into a runner (use the private key from Terraform output)
   ssh -i runner_private_key.pem ubuntu@<runner-ip>
   
   # Check runner status
   runner-monitor
   
   # Check system information
   runner-info
   ```

## Configuration Options

### Runner Types

#### Organization-Level Runners
```hcl
github_org = "your-organization"
github_repo = ""
```

#### Repository-Level Runners
```hcl
github_org = ""
github_repo = "owner/repository-name"
```

### Instance Sizes

Available OVH Cloud flavors:

| Flavor ID | vCPUs | RAM | Storage | Use Case |
|-----------|-------|-----|---------|----------|
| `b2-7`    | 2     | 7GB | 50GB    | Small workloads |
| `b2-15`   | 4     | 15GB| 100GB   | Medium workloads |
| `b2-30`   | 8     | 30GB| 200GB   | Large workloads |
| `c2-7`    | 2     | 7GB | 50GB    | CPU optimized |
| `c2-15`   | 4     | 15GB| 100GB   | CPU optimized |
| `c2-30`   | 8     | 30GB| 200GB   | CPU optimized |
| `r2-15`   | 2     | 15GB| 50GB    | Memory optimized |
| `r2-30`   | 4     | 30GB| 100GB   | Memory optimized |
| `r2-60`   | 8     | 60GB| 200GB   | Memory optimized |

### Regions

Available OVH Cloud regions:

- `GRA11` - Gravelines, France
- `SBG5` - Strasbourg, France
- `BHS5` - Beauharnois, Canada
- `WAW1` - Warsaw, Poland
- `UK1` - London, UK
- `DE1` - Frankfurt, Germany
- `US-EAST-VA-1` - Virginia, USA
- `US-WEST-OR-1` - Oregon, USA
- `CA-ON-1` - Toronto, Canada
- `AU-SYD-1` - Sydney, Australia
- `SG-SIN-1` - Singapore

## Monitoring and Maintenance

### Health Checks
- Automatic health checks run every 5 minutes
- Failed runners are automatically restarted
- Health check logs are available at `/var/log/github-runner-health.log`

### Status Monitoring
- Web-based status page at `http://<runner-ip>/runner-status.html`
- Command-line monitoring with `runner-monitor`
- System information with `runner-info`

### Logs
- Runner logs: `/home/github-runner/_diag/`
- Health check logs: `/var/log/github-runner-health.log`
- System logs: `journalctl -u github-runner`

## Security Features

- **SSH Key Management**: Automatic generation and management of SSH keys
- **Firewall Configuration**: UFW firewall with minimal required ports
- **User Isolation**: Dedicated `github-runner` user with limited permissions
- **Secure Communication**: All GitHub communication uses HTTPS
- **Graceful Shutdown**: Proper cleanup of runners on instance termination

## Cost Optimization

### Auto-scaling Considerations
- Consider using OVH Cloud's auto-scaling features
- Monitor usage patterns and adjust `runner_count` accordingly
- Use appropriate instance sizes based on workload requirements

### Docker Registry Mirroring
```hcl
docker_registry_mirror = "https://registry-1.docker.io"
```

## Troubleshooting

### Common Issues

1. **Runner not appearing in GitHub**:
   - Check the GitHub token permissions
   - Verify the organization/repository name
   - Check the runner logs: `journalctl -u github-runner`

2. **Runner offline**:
   - SSH into the instance and run `runner-monitor`
   - Check health check logs: `tail -f /var/log/github-runner-health.log`
   - Restart the service: `sudo systemctl restart github-runner`

3. **Docker issues**:
   - Verify Docker is running: `sudo systemctl status docker`
   - Check Docker daemon configuration: `cat /etc/docker/daemon.json`

### Debugging Commands

```bash
# Check runner service status
sudo systemctl status github-runner

# View runner logs
sudo journalctl -u github-runner -f

# Check health check logs
tail -f /var/log/github-runner-health.log

# Monitor runner processes
ps aux | grep -E "(run.sh|Runner)"

# Check Docker status
sudo systemctl status docker
docker info
```

## Cleanup

To destroy the infrastructure:

```bash
# Destroy all resources
terraform destroy

# Note: This will also remove runners from GitHub automatically
```

## Advanced Configuration

### Custom Runner Labels
```hcl
runner_labels = "ovh,self-hosted,ubuntu-22.04,custom-label"
```

### Load Balancer Setup
```hcl
create_load_balancer = true
```

### Custom Network Configuration
```hcl
vlan_id = 1234
subnet_network = "192.168.1.0/24"
subnet_start = "192.168.1.10"
subnet_end = "192.168.1.100"
```

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Review the runner logs and health check logs
3. Verify your OVH and GitHub credentials
4. Check the [OVH Cloud documentation](https://docs.ovh.com/gb/en/public-cloud/)
5. Check the [GitHub Actions documentation](https://docs.github.com/en/actions/hosting-your-own-runners)

## License

This project is licensed under the MIT License.
