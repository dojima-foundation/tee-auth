# Runner Information
output "runner_ips" {
  description = "IP addresses of the GitHub Actions runners"
  value       = [for i in range(var.runner_count) : data.external.runner_ips[i].result.ip]
}

output "runner_names" {
  description = "Names of the GitHub Actions runners"
  value       = [for i in range(var.runner_count) : "runner-${i + 1}"]
}

output "runner_status_pages" {
  description = "Status page URLs for the runners"
  value       = [for i in range(var.runner_count) : "http://${data.external.runner_ips[i].result.ip}/"]
}

# SSH Configuration
output "ssh_private_key_path" {
  description = "Path to the SSH private key file"
  value       = local_file.ssh_private_key.filename
}

# Health Check Script
output "health_check_script_path" {
  description = "Path to the health check script"
  value       = var.enable_health_checks ? local_file.health_check_script[0].filename : null
}

# Monitoring Script
output "monitoring_script_path" {
  description = "Path to the monitoring script"
  value       = local_file.monitoring_script.filename
}
