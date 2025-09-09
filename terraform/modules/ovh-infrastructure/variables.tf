# OVH Provider Configuration
variable "ovh_endpoint" {
  description = "OVH API endpoint"
  type        = string
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
}

variable "region" {
  description = "OVH Cloud region"
  type        = string
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
  description = "Whether to create a load balancer"
  type        = bool
  default     = false
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
