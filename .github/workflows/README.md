# GitHub Actions Workflows

This directory contains GitHub Actions workflows for the tee-auth project components.

## Workflow Overview

The project uses a modular CI/CD approach with separate workflows for each component and a main workflow for full integration testing.

### Component Workflows

#### 1. `renclave-v2.yml` - Rust Enclave Component
- **Triggers**: Changes to `renclave-v2/**` or workflow file
- **Jobs**:
  - **Test and Build**: Unit tests, integration tests, code formatting, clippy, coverage
  - **Docker**: Docker image building and testing
  - **Security**: Security audits with cargo-audit and cargo-deny
- **Features**:
  - Rust 1.82 toolchain
  - Cargo caching for faster builds
  - Code coverage with cargo-tarpaulin
  - Security scanning

#### 2. `gauth.yml` - Go Authentication Service
- **Triggers**: Changes to `gauth/**` or workflow file
- **Jobs**:
  - **Test and Build**: Unit tests, integration tests, E2E tests, protobuf generation
  - **Docker**: Docker image building and container testing
  - **Database**: Database migration testing
- **Features**:
  - Go 1.23 toolchain
  - PostgreSQL and Redis services for testing
  - Protobuf code generation
  - Database migration validation
  - Security scanning with gosec

#### 3. `web.yml` - Next.js Web Application
- **Triggers**: Changes to `web/**` or workflow file
- **Jobs**:
  - **Test and Build**: Unit tests, code linting, building
  - **E2E**: Playwright end-to-end testing
  - **Performance**: Lighthouse CI performance testing
  - **Accessibility**: Accessibility testing
  - **Mobile**: Mobile device testing
- **Features**:
  - Node.js 20
  - Playwright for E2E testing
  - Performance and accessibility testing
  - Mobile responsiveness testing

### Main Integration Workflow

#### `main.yml` - Full Project Integration
- **Triggers**: 
  - Changes to any component (excluding individual workflow files)
  - Manual workflow dispatch
- **Jobs**:
  - **Renclave-v2**: Basic Rust testing and building
  - **Gauth**: Basic Go testing and building
  - **Web**: Basic Node.js testing and building
  - **Integration**: Full integration tests across components
  - **Summary**: Test results summary and status reporting

## Workflow Features

### Caching
- **Rust**: Cargo dependencies and build artifacts
- **Go**: Module cache and build cache
- **Node.js**: npm dependencies and build cache

### Services
- **PostgreSQL**: Database testing for Gauth
- **Redis**: Cache testing for Gauth
- **Docker**: Container testing for all components

### Testing Strategy
- **Unit Tests**: Fast, isolated component testing
- **Integration Tests**: Component interaction testing
- **E2E Tests**: Full user workflow testing
- **Security Tests**: Vulnerability scanning and code analysis
- **Performance Tests**: Performance benchmarking and monitoring

## Usage

### Automatic Triggers
Workflows automatically run when:
- Code is pushed to `main` or `develop` branches
- Pull requests are opened against `main` or `develop`
- Changes are made to relevant component directories

### Manual Execution
The main workflow can be manually triggered via:
- GitHub Actions UI
- GitHub CLI: `gh workflow run main.yml`

### Path-Based Triggers
Each component workflow only runs when its relevant files change:
- `renclave-v2/**` → `renclave-v2.yml`
- `gauth/**` → `gauth.yml`
- `web/**` → `web.yml`
- Other changes → `main.yml`

## Dependencies

### Required Tools
- **Rust**: 1.82+ toolchain
- **Go**: 1.23+ toolchain
- **Node.js**: 20+ LTS
- **Docker**: For container testing
- **PostgreSQL**: 15+ for database testing
- **Redis**: 7+ for cache testing

### GitHub Actions
- `actions/checkout@v4`
- `actions/setup-go@v5`
- `actions/setup-node@v4`
- `dtolnay/rust-toolchain@stable`
- `arduino/setup-protoc@v1`
- `docker/setup-buildx-action@v3`
- `actions/cache@v4`
- `actions/upload-artifact@v4`
- `codecov/codecov-action@v4`

## Best Practices

### Performance
- Use caching for dependencies and build artifacts
- Run jobs in parallel when possible
- Use matrix builds for multiple targets
- Implement early failure for critical tests

### Security
- Regular security scanning with cargo-audit and gosec
- Dependency vulnerability checking
- Code quality enforcement with linters
- Secure handling of secrets and environment variables

### Reliability
- Health checks for external services
- Proper error handling and reporting
- Artifact retention policies
- Comprehensive test coverage

## Troubleshooting

### Common Issues
1. **Cache Misses**: Clear cache if builds become inconsistent
2. **Service Health**: Check service health commands and timeouts
3. **Dependency Conflicts**: Verify toolchain versions match requirements
4. **Resource Limits**: Monitor GitHub Actions resource usage

### Debugging
- Enable debug logging in workflows
- Check service logs for database/Redis issues
- Verify file paths and working directories
- Review artifact uploads and downloads

## Contributing

When adding new workflows or modifying existing ones:
1. Follow the established naming conventions
2. Include proper error handling and cleanup
3. Add appropriate documentation
4. Test workflows locally when possible
5. Use semantic versioning for action versions
