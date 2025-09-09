# OVH Infrastructure Outputs
output "project_id" {
  description = "OVH Cloud Project ID"
  value       = module.ovh_infrastructure.project_id
}

output "available_regions" {
  description = "Available regions for the project"
  value       = module.ovh_infrastructure.available_regions
}

output "openstack_credentials" {
  description = "OpenStack credentials for runner user"
  value       = module.ovh_infrastructure.openstack_credentials
  sensitive   = true
}

# SSH Configuration Outputs
output "ssh_public_key" {
  description = "SSH public key for runner instances"
  value       = module.ovh_infrastructure.ssh_public_key
}

output "ssh_public_key_formatted" {
  description = "SSH public key formatted for OVH Cloud console"
  value       = module.ovh_infrastructure.ssh_public_key_formatted
}

# GitHub Runners Outputs
output "runner_ips" {
  description = "IP addresses of the GitHub Actions runners"
  value       = module.github_runners.runner_ips
}

output "runner_names" {
  description = "Names of the GitHub Actions runners"
  value       = module.github_runners.runner_names
}

output "runner_status_pages" {
  description = "Status page URLs for the runners"
  value       = module.github_runners.runner_status_pages
}

# Monitoring Outputs
output "monitoring_dashboard_url" {
  description = "URL of the monitoring dashboard"
  value       = module.monitoring.dashboard_url
  sensitive   = true
}

# Setup Information
output "setup_info" {
  description = "Information needed for runner setup"
  value = {
    github_org                = var.github_org
    github_repo               = var.github_repo
    runner_labels             = join(",", var.runner_labels)
    runner_count              = var.runner_count
    region                    = var.region
    image_id                  = var.runner_image_id
    flavor_id                 = var.runner_flavor_id
    project_id                = var.project_id
  }
}

# Next Steps
output "next_steps" {
  description = "Next steps to complete the setup"
  value = <<EOF
GitHub Actions runners have been configured. Next steps:

1. Check your GitHub repository settings:
   - Go to: https://github.com/${var.github_repo}/settings/actions/runners
   - Verify that your runners are online

2. Test your runners:
   - Push a commit to trigger a workflow
   - Check that jobs execute successfully

3. Monitor your runners:
   - Status pages: ${join(", ", module.github_runners.runner_status_pages)}
   - SSH access: ssh -i runner_private_key.pem ubuntu@<runner-ip>

4. Use the helper scripts:
   - ./scripts/monitoring/check_runner_logs.sh <runner-ip>
   - ./scripts/monitoring/find_runner_ip.sh
EOF
}
