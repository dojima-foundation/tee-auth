# gauth - Go Authentication Service
# Makefile for building, testing, and deployment

.PHONY: help build test clean proto docker migrate lint fmt vet deps run dev

# Variables
SERVICE_NAME := gauth
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -X main.serviceVersion=$(VERSION) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.gitCommit=$(GIT_COMMIT)

# Docker variables
DOCKER_IMAGE := $(SERVICE_NAME)
DOCKER_TAG := $(VERSION)

# Database variables
DB_HOST := localhost
DB_PORT := 5432
DB_NAME := gauth
DB_USER := gauth
DB_PASSWORD := password
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing migration tool..."
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	protoc --experimental_allow_proto3_optional --go_out=. --go-grpc_out=. --proto_path=. api/proto/gauth.proto
	@echo "Moving generated files to correct location..."
	@if [ -d "github.com/dojima-foundation/tee-auth/gauth/api/proto" ]; then \
		mv github.com/dojima-foundation/tee-auth/gauth/api/proto/*.pb.go api/proto/; \
		rm -rf github.com; \
	fi
	@echo "Protobuf code generated successfully"

build: proto ## Build the service
	@echo "Building $(SERVICE_NAME)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o bin/$(SERVICE_NAME) \
		./cmd/server
	@echo "Build completed: bin/$(SERVICE_NAME)"

build-local: proto ## Build for local development
	@echo "Building $(SERVICE_NAME) for local development..."
	go build \
		-ldflags "$(LDFLAGS)" \
		-o bin/$(SERVICE_NAME) \
		./cmd/server
	@echo "Local build completed: bin/$(SERVICE_NAME)"

test: ## Run all tests with coverage
	@echo "Running comprehensive test suite..."
	./scripts/test-coverage.sh

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@mkdir -p coverage
	go test -v -race -coverprofile=coverage/unit.out ./internal/... ./pkg/...
	go tool cover -html=coverage/unit.out -o coverage/unit.html
	@echo "Unit tests completed. Coverage: coverage/unit.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	INTEGRATION_TESTS=true go test -v -race -timeout=5m ./test/integration/...

test-rest: ## Run REST API tests
	@echo "Running REST API integration tests..."
	INTEGRATION_TESTS=true go test -v -race -timeout=5m ./test/integration/ -run TestRESTIntegrationSuite

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	E2E_TESTS=true go test -v -timeout=10m ./test/e2e/...

test-e2e-wallet: ## Run wallet workflow E2E tests only
	@echo "Running wallet workflow E2E tests..."
	E2E_TESTS=true go test -v -timeout=10m ./test/e2e/... -run TestWalletWorkflowE2ETestSuite

test-short: ## Run short tests
	@echo "Running short tests..."
	go test -short -v ./...

test-coverage: ## Generate detailed coverage report
	@echo "Generating coverage report..."
	./scripts/test-coverage.sh
	@echo "Coverage report generated in coverage/ directory"

test-coverage-open: ## Generate and open coverage report
	@echo "Generating and opening coverage report..."
	OPEN_REPORT=true ./scripts/test-coverage.sh

test-all: ## Run all test types
	@echo "Running all test types..."
	$(MAKE) test-unit
	INTEGRATION_TESTS=true $(MAKE) test-integration
	E2E_TESTS=true $(MAKE) test-e2e

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf coverage.out coverage.html
	go clean -cache -testcache

# Database operations
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration version: $(VERSION)"
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-version: ## Show current migration version
	@echo "Current migration version:"
	migrate -path migrations -database "$(DATABASE_URL)" version

# Development
run: build-local ## Run the service locally
	@echo "Starting $(SERVICE_NAME)..."
	./bin/$(SERVICE_NAME)

dev: ## Run in development mode with live reload
	@echo "Starting development server..."
	air -c .air.toml

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@echo "Then run: make migrate-up && make run"

# Docker operations
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 9090:9090 \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

# Production operations
deploy-staging: ## Deploy to staging
	@echo "Deploying to staging..."
	# Add your staging deployment commands here

deploy-prod: ## Deploy to production
	@echo "Deploying to production..."
	# Add your production deployment commands here

# Utilities
check: lint vet test-short ## Run all checks (lint, vet, test)

security-check: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

mod-update: ## Update all modules
	@echo "Updating modules..."
	go get -u ./...
	go mod tidy

tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4
	go install github.com/cosmtrek/air@latest

generate: proto ## Generate all code
	@echo "Generating all code..."
	go generate ./...

all: clean deps generate build test ## Build everything from scratch

# Help target should be first
.DEFAULT_GOAL := help
