# OVH Provider Configuration
variable "ovh_endpoint" {
  description = "OVH API endpoint (ovh-eu, ovh-ca, ovh-us, kimsufi-eu, kimsufi-ca, soyoustart-eu, soyoustart-ca)"
  type        = string
  default     = "ovh-ca"  # Singapore region
}

variable "ovh_application_key" {
  description = "OVH Application Key"
  type        = string
  sensitive   = true
}

variable "ovh_application_secret" {
  description = "OVH Application Secret"
  type        = string
  sensitive   = true
}

variable "ovh_consumer_key" {
  description = "OVH Consumer Key"
  type        = string
  sensitive   = true
}

# Project Configuration
variable "project_id" {
  description = "OVH Cloud Project ID"
  type        = string
  default     = "5729bde2da4e41c8b8157376c56c6899"
}

variable "region" {
  description = "OVH Cloud region"
  type        = string
  default     = "SGP1"  # Singapore region
}

# GitHub Configuration
variable "github_token" {
  description = "GitHub Personal Access Token with repo and admin:org permissions"
  type        = string
  sensitive   = true
}

variable "github_org" {
  description = "GitHub organization name (leave empty for repository-level runners)"
  type        = string
  default     = ""
}

variable "github_repo" {
  description = "GitHub repository name (format: owner/repo)"
  type        = string
  default     = "dojima-foundation/tee-auth"
}

# Runner Configuration
variable "runner_count" {
  description = "Number of GitHub Actions runners to create"
  type        = number
  default     = 2
}

variable "runner_labels" {
  description = "Labels to apply to the GitHub Actions runners"
  type        = list(string)
  default     = ["ovh", "self-hosted", "ubuntu-22.04"]
}

variable "runner_image_id" {
  description = "OVH Cloud image ID for runner instances"
  type        = string
  default     = "Ubuntu 22.04"
}

variable "runner_flavor_id" {
  description = "OVH Cloud flavor ID for runner instances"
  type        = string
  default     = "b2-7" # 2 vCPUs, 7GB RAM
}

# Network Configuration
variable "vlan_id" {
  description = "VLAN ID for the private network"
  type        = number
  default     = null
}

variable "subnet_network" {
  description = "Subnet network CIDR"
  type        = string
  default     = "10.0.0.0/24"
}

variable "subnet_start" {
  description = "Subnet start IP"
  type        = string
  default     = "10.0.0.10"
}

variable "subnet_end" {
  description = "Subnet end IP"
  type        = string
  default     = "10.0.0.100"
}

# Optional Configuration
variable "create_load_balancer" {
  description = "Whether to create a load balancer for the runners"
  type        = bool
  default     = false
}

variable "docker_registry_mirror" {
  description = "Docker registry mirror URL (optional)"
  type        = string
  default     = ""
}

# Monitoring Configuration
variable "enable_status_pages" {
  description = "Enable status pages for runners"
  type        = bool
  default     = true
}

variable "enable_health_checks" {
  description = "Enable health checks for runners"
  type        = bool
  default     = true
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "production"
    Purpose     = "github-actions-runner"
    ManagedBy   = "terraform"
    Project     = "tee-auth"
  }
}
