# Generate SSH key pair for runner instances
resource "tls_private_key" "runner_ssh" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Get existing cloud projects
data "ovh_cloud_projects" "projects" {}

# Get the specified project
data "ovh_cloud_project" "github_runners" {
  service_name = var.project_id
}

# Get available regions for the project
data "ovh_cloud_project_regions" "regions" {
  service_name = data.ovh_cloud_project.github_runners.service_name
}

# Get available flavors for the region
data "ovh_cloud_project_region" "target_region" {
  service_name = data.ovh_cloud_project.github_runners.service_name
  name         = var.region
}

# Create OVH Cloud User for runner instances
resource "ovh_cloud_project_user" "runner_user" {
  service_name = data.ovh_cloud_project.github_runners.service_name
  description  = "GitHub Actions Runner User"
}

# Create private network if VLAN ID is specified
resource "ovh_cloud_project_network_private" "runner_network" {
  count = var.vlan_id != null ? 1 : 0
  
  service_name = data.ovh_cloud_project.github_runners.service_name
  name         = "github-runners-network"
  regions      = [var.region]
  vlan_id      = var.vlan_id
}

# Create subnet for the private network
resource "ovh_cloud_project_network_private_subnet" "runner_subnet" {
  count = var.vlan_id != null ? 1 : 0
  
  service_name = data.ovh_cloud_project.github_runners.service_name
  network_id   = ovh_cloud_project_network_private.runner_network[0].id
  region       = var.region
  start        = var.subnet_start
  end          = var.subnet_end
  network      = var.subnet_network
}

# Create security group for runners
resource "ovh_cloud_project_network_security_group" "runner_sg" {
  service_name = data.ovh_cloud_project.github_runners.service_name
  name         = "github-runners-sg"
  description  = "Security group for GitHub Actions runners"
}

# Security group rules
resource "ovh_cloud_project_network_security_group_rule" "ssh" {
  service_name = ovh_cloud_project_network_security_group.runner_sg.service_name
  security_group_id = ovh_cloud_project_network_security_group.runner_sg.id
  
  protocol = "TCP"
  action   = "ACCEPT"
  port     = 22
  ip_range = "0.0.0.0/0"
}

resource "ovh_cloud_project_network_security_group_rule" "http" {
  service_name = ovh_cloud_project_network_security_group.runner_sg.service_name
  security_group_id = ovh_cloud_project_network_security_group.runner_sg.id
  
  protocol = "TCP"
  action   = "ACCEPT"
  port     = 80
  ip_range = "0.0.0.0/0"
}

resource "ovh_cloud_project_network_security_group_rule" "https" {
  service_name = ovh_cloud_project_network_security_group.runner_sg.service_name
  security_group_id = ovh_cloud_project_network_security_group.runner_sg.id
  
  protocol = "TCP"
  action   = "ACCEPT"
  port     = 443
  ip_range = "0.0.0.0/0"
}

# Create load balancer if requested
resource "ovh_cloud_project_loadbalancer" "runner_lb" {
  count = var.create_load_balancer ? 1 : 0
  
  service_name = data.ovh_cloud_project.github_runners.service_name
  region_name  = var.region
  flavor_id    = "lb"
  name         = "github-runners-lb"
}
