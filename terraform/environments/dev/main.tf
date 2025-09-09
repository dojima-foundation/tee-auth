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
  
  backend "local" {
    path = "terraform.tfstate"
  }
}

# Configure OVH Provider
provider "ovh" {
  endpoint           = var.ovh_endpoint
  application_key    = var.ovh_application_key
  application_secret = var.ovh_application_secret
  consumer_key       = var.ovh_consumer_key
}

# OVH Infrastructure Module
module "ovh_infrastructure" {
  source = "../../modules/ovh-infrastructure"
  
  # OVH Configuration
  ovh_endpoint           = var.ovh_endpoint
  ovh_application_key    = var.ovh_application_key
  ovh_application_secret = var.ovh_application_secret
  ovh_consumer_key       = var.ovh_consumer_key
  
  # Project Configuration
  project_id = var.project_id
  region     = var.region
  
  # Tags
  tags = var.tags
}

# GitHub Runners Module (Single runner for dev)
module "github_runners" {
  source = "../../modules/github-runner"
  
  # Dependencies
  depends_on = [module.ovh_infrastructure]
  
  # GitHub Configuration
  github_token = var.github_token
  github_org   = var.github_org
  github_repo  = var.github_repo
  
  # Runner Configuration (Single runner for dev)
  runner_count     = 1
  runner_labels    = concat(var.runner_labels, ["dev"])
  runner_image_id  = var.runner_image_id
  runner_flavor_id = "b2-7"  # Smaller instance for dev
  
  # Infrastructure Configuration
  project_id = var.project_id
  region     = var.region
  
  # OpenStack Credentials
  openstack_username = module.ovh_infrastructure.openstack_username
  openstack_password = module.ovh_infrastructure.openstack_password
  
  # SSH Configuration
  ssh_public_key  = module.ovh_infrastructure.ssh_public_key
  ssh_private_key = module.ovh_infrastructure.ssh_private_key
  
  # Network Configuration
  vlan_id         = var.vlan_id
  subnet_network  = var.subnet_network
  subnet_start    = var.subnet_start
  subnet_end      = var.subnet_end
  
  # Optional Configuration
  create_load_balancer    = false  # No load balancer for dev
  docker_registry_mirror  = var.docker_registry_mirror
  
  # Tags
  tags = var.tags
}

# Monitoring Module (Basic monitoring for dev)
module "monitoring" {
  source = "../../modules/monitoring"
  
  # Dependencies
  depends_on = [module.github_runners]
  
  # Runner Configuration
  runner_ips = module.github_runners.runner_ips
  
  # Monitoring Configuration (Basic for dev)
  enable_status_pages = true
  enable_health_checks = true
  enable_grafana = false
  enable_prometheus = false
  enable_alerting = false
  
  # Tags
  tags = var.tags
}
