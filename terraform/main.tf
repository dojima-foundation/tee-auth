terraform {
  required_version = ">= 1.0"
  required_providers {
    ovh = {
      source  = "ovh/ovh"
      version = "~> 0.34"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# Configure OVH Provider
provider "ovh" {
  endpoint           = var.ovh_endpoint
  application_key    = var.ovh_application_key
  application_secret = var.ovh_application_secret
  consumer_key       = var.ovh_consumer_key
}

# Generate SSH key pair for runner instances
resource "tls_private_key" "runner_ssh" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Get existing cloud projects
data "ovh_cloud_projects" "projects" {}

# Get the first available project (or create one manually)
data "ovh_cloud_project" "github_runners" {
  service_name = "5729bde2da4e41c8b8157376c56c6899"
}

# Get available regions for the project
data "ovh_cloud_project_regions" "regions" {
  service_name = data.ovh_cloud_project.github_runners.service_name
}

# Get available flavors for the region
data "ovh_cloud_project_region" "singapore" {
  service_name = data.ovh_cloud_project.github_runners.service_name
  name         = "SGP1"  # Using SGP1 instead of SG-SIN-1
}

# Create OVH Cloud User for runner instances
resource "ovh_cloud_project_user" "runner_user" {
  service_name = data.ovh_cloud_project.github_runners.service_name
  description  = "GitHub Actions Runner User"
}

# Output important information
output "project_id" {
  description = "OVH Cloud Project ID"
  value       = data.ovh_cloud_project.github_runners.service_name
}

output "available_regions" {
  description = "Available regions for the project"
  value       = data.ovh_cloud_project_regions.regions.names
}

output "singapore_region_services" {
  description = "Services available in Singapore region"
  value       = data.ovh_cloud_project_region.singapore.services
}

output "ssh_private_key" {
  description = "SSH private key for runner instances"
  value       = tls_private_key.runner_ssh.private_key_pem
  sensitive   = true
}

output "ssh_public_key" {
  description = "SSH public key for runner instances"
  value       = tls_private_key.runner_ssh.public_key_openssh
}

output "openstack_credentials" {
  description = "OpenStack credentials for runner user"
  value = {
    username = ovh_cloud_project_user.runner_user.username
    password = ovh_cloud_project_user.runner_user.password
  }
  sensitive = true
}

# Note: OVH Cloud instances need to be created manually through the console
# or using OpenStack CLI with the credentials provided above
output "next_steps" {
  description = "Next steps to complete the setup"
  sensitive   = true
  value = <<EOF
OVH Cloud instances need to be created manually. Here are the next steps:

1. Use the OpenStack credentials above to authenticate with OVH Cloud
2. Create instances using OpenStack CLI or the OVH Cloud console
3. Use the SSH public key above for instance access
4. Run the user_data.sh script on each instance

OpenStack CLI commands:
export OS_AUTH_URL=https://auth.cloud.ovh.net/v3
export OS_IDENTITY_API_VERSION=3
export OS_PROJECT_ID=${data.ovh_cloud_project.github_runners.service_name}
export OS_USERNAME=${ovh_cloud_project_user.runner_user.username}
export OS_PASSWORD=${ovh_cloud_project_user.runner_user.password}
export OS_REGION_NAME=SGP1

Then create instances:
openstack server create --image "Ubuntu 22.04" --flavor ${var.runner_flavor_id} --key-name github-runner-key github-runner-1
EOF
}
