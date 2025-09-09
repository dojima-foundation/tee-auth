# Project Information
output "project_id" {
  description = "OVH Cloud Project ID"
  value       = data.ovh_cloud_project.github_runners.service_name
}

output "available_regions" {
  description = "Available regions for the project"
  value       = data.ovh_cloud_project_regions.regions.names
}

output "target_region_services" {
  description = "Services available in target region"
  value       = data.ovh_cloud_project_region.target_region.services
}

# SSH Configuration
output "ssh_private_key" {
  description = "SSH private key for runner instances"
  value       = tls_private_key.runner_ssh.private_key_pem
  sensitive   = true
}

output "ssh_public_key" {
  description = "SSH public key for runner instances"
  value       = tls_private_key.runner_ssh.public_key_openssh
}

output "ssh_public_key_formatted" {
  description = "SSH public key formatted for OVH Cloud console"
  value       = tls_private_key.runner_ssh.public_key_openssh
}

# OpenStack Credentials
output "openstack_credentials" {
  description = "OpenStack credentials for runner user"
  value = {
    username = ovh_cloud_project_user.runner_user.username
    password = ovh_cloud_project_user.runner_user.password
  }
  sensitive = true
}

output "openstack_username" {
  description = "OpenStack username"
  value       = ovh_cloud_project_user.runner_user.username
}

output "openstack_password" {
  description = "OpenStack password"
  value       = ovh_cloud_project_user.runner_user.password
  sensitive   = true
}

# Network Configuration
output "private_network_id" {
  description = "ID of the private network"
  value       = var.vlan_id != null ? ovh_cloud_project_network_private.runner_network[0].id : null
}

output "subnet_id" {
  description = "ID of the subnet"
  value       = var.vlan_id != null ? ovh_cloud_project_network_private_subnet.runner_subnet[0].id : null
}

output "security_group_id" {
  description = "ID of the security group"
  value       = ovh_cloud_project_network_security_group.runner_sg.id
}

# Load Balancer
output "load_balancer_id" {
  description = "ID of the load balancer"
  value       = var.create_load_balancer ? ovh_cloud_project_loadbalancer.runner_lb[0].id : null
}

output "load_balancer_ip" {
  description = "IP address of the load balancer"
  value       = var.create_load_balancer ? ovh_cloud_project_loadbalancer.runner_lb[0].ipv4 : null
}
