# GitHub Configuration
variable "github_token" {
  description = "GitHub Personal Access Token"
  type        = string
  sensitive   = true
}

variable "github_org" {
  description = "GitHub organization name"
  type        = string
  default     = ""
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
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
  default     = ["ovh", "self-hosted"]
}

variable "runner_image_id" {
  description = "OVH Cloud image ID for runner instances"
  type        = string
  default     = "Ubuntu 22.04"
}

variable "runner_flavor_id" {
  description = "OVH Cloud flavor ID for runner instances"
  type        = string
  default     = "b2-7"
}

# Infrastructure Configuration
variable "project_id" {
  description = "OVH Cloud Project ID"
  type        = string
}

variable "region" {
  description = "OVH Cloud region"
  type        = string
}

# OpenStack Credentials
variable "openstack_username" {
  description = "OpenStack username"
  type        = string
}

variable "openstack_password" {
  description = "OpenStack password"
  type        = string
  sensitive   = true
}

# SSH Configuration
variable "ssh_public_key" {
  description = "SSH public key"
  type        = string
}

variable "ssh_private_key" {
  description = "SSH private key"
  type        = string
  sensitive   = true
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

variable "security_group_id" {
  description = "Security group ID"
  type        = string
  default     = "default"
}

# Optional Configuration
variable "create_load_balancer" {
  description = "Whether to create a load balancer"
  type        = bool
  default     = false
}

variable "docker_registry_mirror" {
  description = "Docker registry mirror URL"
  type        = string
  default     = ""
}

variable "enable_health_checks" {
  description = "Enable health checks for runners"
  type        = bool
  default     = true
}

variable "enable_status_pages" {
  description = "Enable status pages for runners"
  type        = bool
  default     = true
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
