.PHONY: help init plan apply destroy validate format clean status setup logs dev prod staging

# Default target
help:
	@echo "GitHub Actions Runners on OVH Cloud - Terraform Management"
	@echo ""
	@echo "Available commands:"
	@echo "  init      - Initialize Terraform and download providers"
	@echo "  plan      - Show what Terraform will do"
	@echo "  apply     - Apply the Terraform configuration"
	@echo "  destroy   - Destroy all resources"
	@echo "  validate  - Validate Terraform configuration"
	@echo "  format    - Format Terraform files"
	@echo "  clean     - Clean up Terraform files and state"
	@echo "  status    - Show current status and outputs"
	@echo "  setup     - Complete setup (init + plan + apply)"
	@echo "  logs      - Show runner logs (requires SSH access)"
	@echo ""
	@echo "Environment-specific commands:"
	@echo "  dev       - Deploy to development environment"
	@echo "  staging   - Deploy to staging environment"
	@echo "  prod      - Deploy to production environment"
	@echo ""
	@echo "Usage examples:"
	@echo "  make dev     - Deploy development environment"
	@echo "  make prod    - Deploy production environment"
	@echo "  make status  - Show current status"

# Environment variables
ENVIRONMENT ?= prod
TERRAFORM_DIR = environments/$(ENVIRONMENT)

# Initialize Terraform
init:
	@echo "Initializing Terraform for $(ENVIRONMENT) environment..."
	cd $(TERRAFORM_DIR) && terraform init
	@echo "Terraform initialized successfully!"

# Plan the deployment
plan:
	@echo "Planning Terraform deployment for $(ENVIRONMENT) environment..."
	cd $(TERRAFORM_DIR) && terraform plan
	@echo "Plan completed!"

# Apply the configuration
apply:
	@echo "Applying Terraform configuration for $(ENVIRONMENT) environment..."
	cd $(TERRAFORM_DIR) && terraform apply -auto-approve
	@echo "Configuration applied successfully!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Check your GitHub repository/organization settings"
	@echo "2. Navigate to Actions → Runners to verify runners are online"
	@echo "3. Use 'make status' to see runner IPs and SSH keys"

# Destroy all resources
destroy:
	@echo "WARNING: This will destroy all resources in $(ENVIRONMENT) environment!"
	@read -p "Are you sure? Type 'yes' to continue: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "Destroying resources..."; \
		cd $(TERRAFORM_DIR) && terraform destroy -auto-approve; \
		echo "Resources destroyed successfully!"; \
	else \
		echo "Destroy cancelled."; \
	fi

# Validate Terraform configuration
validate:
	@echo "Validating Terraform configuration..."
	@for env in dev staging prod; do \
		echo "Validating $$env environment..."; \
		cd environments/$$env && terraform validate; \
	done
	@echo "Configuration is valid!"

# Format Terraform files
format:
	@echo "Formatting Terraform files..."
	terraform fmt -recursive
	@echo "Files formatted!"

# Clean up Terraform files and state
clean:
	@echo "Cleaning up Terraform files..."
	@for env in dev staging prod; do \
		echo "Cleaning $$env environment..."; \
		rm -rf environments/$$env/.terraform environments/$$env/.terraform.lock.hcl environments/$$env/terraform.tfstate*; \
	done
	rm -rf .terraform .terraform.lock.hcl terraform.tfstate*
	@echo "Cleanup completed!"

# Show current status and outputs
status:
	@echo "=== GitHub Actions Runners Status ($(ENVIRONMENT)) ==="
	@if [ -f "$(TERRAFORM_DIR)/terraform.tfstate" ]; then \
		cd $(TERRAFORM_DIR) && \
		echo "Project ID: $$(terraform output -raw project_id 2>/dev/null || echo 'Not available')"; \
		echo ""; \
		echo "Runner IPs:"; \
		terraform output -raw runner_ips 2>/dev/null || echo "No runners deployed"; \
		echo ""; \
		echo "SSH Public Key:"; \
		terraform output -raw ssh_public_key 2>/dev/null || echo "Not available"; \
		echo ""; \
		echo "To SSH into runners:"; \
		echo "ssh -i runner_private_key.pem ubuntu@<runner-ip>"; \
		echo ""; \
		echo "Status pages:"; \
		for ip in $$(terraform output -raw runner_ips 2>/dev/null); do \
			echo "http://$$ip/"; \
		done; \
	else \
		echo "No Terraform state found. Run 'make init' first."; \
	fi

# Complete setup (init + plan + apply)
setup: init plan apply status

# Show runner logs (requires SSH access)
logs:
	@echo "=== Runner Logs ($(ENVIRONMENT)) ==="
	@if [ -f "$(TERRAFORM_DIR)/terraform.tfstate" ]; then \
		cd $(TERRAFORM_DIR) && \
		for ip in $$(terraform output -raw runner_ips 2>/dev/null); do \
			echo "Runner $$ip:"; \
			echo "  Service status: $$(ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no ubuntu@$$ip 'sudo systemctl is-active actions.runner.*' 2>/dev/null || echo 'Unreachable')"; \
			echo "  Health check: $$(ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no ubuntu@$$ip 'tail -1 /var/log/runner-health-check.log' 2>/dev/null || echo 'No logs')"; \
			echo ""; \
		done; \
	else \
		echo "No Terraform state found. Run 'make init' first."; \
	fi

# Environment-specific deployments
dev:
	@echo "Deploying to development environment..."
	$(MAKE) ENVIRONMENT=dev setup

staging:
	@echo "Deploying to staging environment..."
	$(MAKE) ENVIRONMENT=staging setup

prod:
	@echo "Deploying to production environment..."
	$(MAKE) ENVIRONMENT=prod setup

# Development environment commands
dev-status:
	@echo "Development environment status:"
	$(MAKE) ENVIRONMENT=dev status

dev-logs:
	@echo "Development environment logs:"
	$(MAKE) ENVIRONMENT=dev logs

dev-destroy:
	@echo "Destroying development environment..."
	$(MAKE) ENVIRONMENT=dev destroy

# Production environment commands
prod-status:
	@echo "Production environment status:"
	$(MAKE) ENVIRONMENT=prod status

prod-logs:
	@echo "Production environment logs:"
	$(MAKE) ENVIRONMENT=prod logs

prod-destroy:
	@echo "Destroying production environment..."
	$(MAKE) ENVIRONMENT=prod destroy

# Utility commands
ssh-key:
	@echo "Creating SSH private key file..."
	@if [ -f "$(TERRAFORM_DIR)/terraform.tfstate" ]; then \
		cd $(TERRAFORM_DIR) && \
		terraform output -raw ssh_private_key > ../../runner_private_key.pem 2>/dev/null || echo "No private key available"; \
		chmod 600 ../../runner_private_key.pem 2>/dev/null || true; \
		echo "SSH private key saved to runner_private_key.pem"; \
	else \
		echo "No Terraform state found. Run 'make init' first."; \
	fi

# Show available OVH regions
regions:
	@echo "=== Available OVH Cloud Regions ==="
	@echo "GRA11 - Gravelines, France"
	@echo "SBG5  - Strasbourg, France"
	@echo "BHS5  - Beauharnois, Canada"
	@echo "WAW1  - Warsaw, Poland"
	@echo "UK1   - London, UK"
	@echo "DE1   - Frankfurt, Germany"
	@echo "US-EAST-VA-1 - Virginia, USA"
	@echo "US-WEST-OR-1 - Oregon, USA"
	@echo "CA-ON-1 - Toronto, Canada"
	@echo "AU-SYD-1 - Sydney, Australia"
	@echo "SGP1  - Singapore"

# Show available OVH flavors
flavors:
	@echo "=== Available OVH Cloud Flavors ==="
	@echo "b2-7  - 2 vCPUs, 7GB RAM, 50GB storage"
	@echo "b2-15 - 4 vCPUs, 15GB RAM, 100GB storage"
	@echo "b2-30 - 8 vCPUs, 30GB RAM, 200GB storage"
	@echo "c2-7  - 2 vCPUs, 7GB RAM, 50GB storage (CPU optimized)"
	@echo "c2-15 - 4 vCPUs, 15GB RAM, 100GB storage (CPU optimized)"
	@echo "c2-30 - 8 vCPUs, 30GB RAM, 200GB storage (CPU optimized)"
	@echo "r2-15 - 2 vCPUs, 15GB RAM, 50GB storage (Memory optimized)"
	@echo "r2-30 - 4 vCPUs, 30GB RAM, 100GB storage (Memory optimized)"
	@echo "r2-60 - 8 vCPUs, 60GB RAM, 200GB storage (Memory optimized)"

# Show project structure
structure:
	@echo "=== Project Structure ==="
	@echo "terraform/"
	@echo "├── environments/"
	@echo "│   ├── dev/          # Development environment"
	@echo "│   ├── staging/      # Staging environment"
	@echo "│   └── prod/         # Production environment"
	@echo "├── modules/"
	@echo "│   ├── ovh-infrastructure/  # OVH Cloud infrastructure"
	@echo "│   ├── github-runner/       # GitHub Actions runner"
	@echo "│   └── monitoring/          # Monitoring and observability"
	@echo "├── scripts/"
	@echo "│   ├── setup/        # Setup scripts"
	@echo "│   ├── monitoring/   # Monitoring scripts"
	@echo "│   └── cleanup/      # Cleanup scripts"
	@echo "└── configs/"
	@echo "    ├── templates/    # Configuration templates"
	@echo "    └── examples/     # Example configurations"