# Generate user data scripts for each runner
locals {
  user_data_scripts = [
    for i in range(var.runner_count) : templatefile("${path.module}/templates/user_data.sh", {
      github_token              = var.github_token
      github_org                = var.github_org
      github_repo               = var.github_repo
      runner_labels             = join(",", var.runner_labels)
      runner_name               = "runner-${i + 1}"
      docker_registry_mirror    = var.docker_registry_mirror
      enable_health_checks      = var.enable_health_checks
      enable_status_pages       = var.enable_status_pages
      project_id                = var.project_id
      region                    = var.region
    })
  ]
}

# Create runner instances using OpenStack
resource "null_resource" "runner_instances" {
  count = var.runner_count
  
  triggers = {
    user_data = local.user_data_scripts[count.index]
    runner_name = "runner-${count.index + 1}"
  }
  
  provisioner "local-exec" {
    command = <<-EOT
      # Set OpenStack credentials
      export OS_AUTH_URL=https://auth.cloud.ovh.net/v3
      export OS_IDENTITY_API_VERSION=3
      export OS_PROJECT_ID=${var.project_id}
      export OS_USERNAME=${var.openstack_username}
      export OS_PASSWORD=${var.openstack_password}
      export OS_REGION_NAME=${var.region}
      
      # Create instance
      openstack server create \
        --image "${var.runner_image_id}" \
        --flavor ${var.runner_flavor_id} \
        --key-name terraform-runner-key \
        --user-data "${base64encode(local.user_data_scripts[count.index])}" \
        --security-group ${var.security_group_id} \
        --network public \
        --wait \
        "github-runner-${count.index + 1}"
    EOT
  }
  
  provisioner "local-exec" {
    when = destroy
    command = <<-EOT
      # Set OpenStack credentials
      export OS_AUTH_URL=https://auth.cloud.ovh.net/v3
      export OS_IDENTITY_API_VERSION=3
      export OS_PROJECT_ID=${var.project_id}
      export OS_USERNAME=${var.openstack_username}
      export OS_PASSWORD=${var.openstack_password}
      export OS_REGION_NAME=${var.region}
      
      # Delete instance
      openstack server delete "github-runner-${count.index + 1}" || true
    EOT
  }
}

# Get runner IP addresses
data "external" "runner_ips" {
  count = var.runner_count
  
  depends_on = [null_resource.runner_instances]
  
  program = ["bash", "-c", <<-EOT
    # Set OpenStack credentials
    export OS_AUTH_URL=https://auth.cloud.ovh.net/v3
    export OS_IDENTITY_API_VERSION=3
    export OS_PROJECT_ID=${var.project_id}
    export OS_USERNAME=${var.openstack_username}
    export OS_PASSWORD=${var.openstack_password}
    export OS_REGION_NAME=${var.region}
    
    # Get instance IP
    IP=$(openstack server show "github-runner-${count.index + 1}" -f value -c addresses | grep -oE '([0-9]{1,3}\.){3}[0-9]{1,3}' | head -1)
    
    echo "{\"ip\": \"$IP\"}"
  EOT
  ]
}

# Create SSH key file
resource "local_file" "ssh_private_key" {
  content  = var.ssh_private_key
  filename = "${path.module}/../../runner_private_key.pem"
  file_permission = "0600"
}

# Health check script
resource "local_file" "health_check_script" {
  count = var.enable_health_checks ? 1 : 0
  
  content = templatefile("${path.module}/templates/health_check.sh", {
    runner_ips = jsonencode([for i in range(var.runner_count) : data.external.runner_ips[i].result.ip])
    ssh_key_path = local_file.ssh_private_key.filename
  })
  
  filename = "${path.module}/../../scripts/monitoring/health_check.sh"
  file_permission = "0755"
}

# Monitoring script
resource "local_file" "monitoring_script" {
  content = templatefile("${path.module}/templates/monitoring.sh", {
    runner_ips = jsonencode([for i in range(var.runner_count) : data.external.runner_ips[i].result.ip])
    ssh_key_path = local_file.ssh_private_key.filename
  })
  
  filename = "${path.module}/../../scripts/monitoring/monitoring.sh"
  file_permission = "0755"
}
