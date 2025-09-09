# Migration Guide: Old to New Terraform Structure

This guide helps you migrate from the old monolithic Terraform structure to the new modular, environment-based structure.

## 🏗️ What Changed

### Old Structure
```
terraform/
├── main.tf
├── variables.tf
├── terraform.tfvars
├── user_data_*.sh
└── scripts/
```

### New Structure
```
terraform/
├── environments/
│   ├── dev/
│   ├── staging/
│   └── prod/
├── modules/
│   ├── ovh-infrastructure/
│   ├── github-runner/
│   └── monitoring/
├── scripts/
│   ├── setup/
│   ├── monitoring/
│   └── cleanup/
└── configs/
    ├── templates/
    └── examples/
```

## 🔄 Migration Steps

### 1. Backup Current State

```bash
# Backup your current Terraform state
cp terraform.tfstate terraform.tfstate.backup
cp terraform.tfvars terraform.tfvars.backup
```

### 2. Choose Your Environment

The new structure supports multiple environments. Choose the appropriate one:

- **Development**: Single runner, smaller instance, basic monitoring
- **Staging**: 2 runners, standard instance, enhanced monitoring
- **Production**: 2+ runners, larger instance, full monitoring stack

### 3. Configure Your Environment

```bash
# For production environment
cd environments/prod
cp ../../configs/examples/terraform.tfvars.example terraform.tfvars

# Edit the configuration
nano terraform.tfvars
```

### 4. Update Configuration Values

Copy your existing values from the old `terraform.tfvars`:

```hcl
# OVH Configuration (copy from old file)
ovh_endpoint = "ovh-ca"
ovh_application_key = "your_existing_key"
ovh_application_secret = "your_existing_secret"
ovh_consumer_key = "your_existing_consumer_key"

# Project Configuration
project_id = "5729bde2da4e41c8b8157376c56c6899"
region = "SGP1"

# GitHub Configuration
github_token = "your_existing_token"
github_repo = "dojima-foundation/tee-auth"

# Runner Configuration
runner_count = 2
runner_labels = ["ovh", "self-hosted", "ubuntu-22.04"]
```

### 5. Deploy New Structure

```bash
# From the terraform root directory
make prod  # or make dev, make staging
```

### 6. Verify Migration

```bash
# Check status
make status

# Verify runners are online
make logs

# Check GitHub repository settings
# Go to: https://github.com/dojima-foundation/tee-auth/settings/actions/runners
```

## 🔧 Key Improvements

### 1. Modular Architecture

- **ovh-infrastructure**: Handles OVH Cloud setup, networking, security groups
- **github-runner**: Manages runner instances, configuration, and monitoring
- **monitoring**: Provides observability and health checks

### 2. Environment Separation

- **Development**: Cost-optimized, single runner
- **Staging**: Balanced setup for testing
- **Production**: High availability, full monitoring

### 3. Enhanced Monitoring

- Web-based status pages
- Automated health checks
- Optional Grafana/Prometheus integration
- Comprehensive logging

### 4. Improved Management

- Environment-specific commands (`make dev`, `make prod`)
- Better error handling and validation
- Automated SSH key management
- Enhanced troubleshooting scripts

## 🚨 Breaking Changes

### 1. File Locations

| Old Location | New Location |
|--------------|--------------|
| `main.tf` | `environments/<env>/main.tf` |
| `variables.tf` | `environments/<env>/variables.tf` |
| `terraform.tfvars` | `environments/<env>/terraform.tfvars` |
| `user_data_*.sh` | `modules/github-runner/templates/user_data.sh` |

### 2. Command Changes

| Old Command | New Command |
|-------------|-------------|
| `terraform apply` | `make prod` (or `make dev`, `make staging`) |
| `terraform plan` | `make plan` |
| `terraform destroy` | `make destroy` |

### 3. State Management

- Each environment has its own state file
- State files are located in `environments/<env>/terraform.tfstate`
- Use environment-specific commands to manage state

## 🔍 Troubleshooting Migration

### Issue: State File Conflicts

```bash
# If you get state conflicts, you may need to import existing resources
terraform import module.ovh_infrastructure.data.ovh_cloud_project.github_runners 5729bde2da4e41c8b8157376c56c6899
```

### Issue: Runner Not Appearing

```bash
# Check if the runner is properly configured
make logs

# Verify GitHub token permissions
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/dojima-foundation/tee-auth
```

### Issue: SSH Key Problems

```bash
# Generate new SSH key
make ssh-key

# Test SSH connection
ssh -i runner_private_key.pem ubuntu@<runner-ip>
```

## 📊 Migration Checklist

- [ ] Backup current Terraform state and configuration
- [ ] Choose target environment (dev/staging/prod)
- [ ] Copy configuration values to new structure
- [ ] Deploy new infrastructure
- [ ] Verify runners are online in GitHub
- [ ] Test workflow execution
- [ ] Update CI/CD pipelines if needed
- [ ] Clean up old files (after verification)

## 🎯 Post-Migration Benefits

1. **Better Organization**: Clear separation of concerns
2. **Environment Management**: Easy switching between environments
3. **Scalability**: Easy to add more runners or environments
4. **Monitoring**: Enhanced observability and health checks
5. **Maintenance**: Simplified updates and troubleshooting
6. **Cost Control**: Environment-specific resource sizing

## 🆘 Need Help?

If you encounter issues during migration:

1. Check the troubleshooting section in the main README
2. Review the logs: `make logs`
3. Verify your configuration: `make validate`
4. Check GitHub repository settings
5. Ensure OVH credentials are correct

## 🔄 Rollback Plan

If you need to rollback to the old structure:

1. Stop new runners: `make destroy`
2. Restore old files from backup
3. Run `terraform apply` from the old structure
4. Verify runners are back online

Remember to clean up any orphaned resources in OVH Cloud Console if needed.
