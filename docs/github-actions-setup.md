# GitHub Actions Environment Setup

This document explains how to set up GitHub Actions with environment-specific secrets and variables for the TEE Auth project.

## Environment Configuration

The GitHub Actions workflow is configured to use a `dev` environment by default. This allows you to:

1. **Separate secrets by environment** (dev, staging, production)
2. **Use environment-specific variables**
3. **Control deployment permissions per environment**
4. **Track deployments per environment**

## Required Secrets

You need to configure the following secrets in your GitHub repository for the `dev` environment:

### 1. Docker Hub Secrets

| Secret Name | Description | Example Value |
|-------------|-------------|---------------|
| `DOCKER_USERNAME` | Docker Hub username | `dojima0foundation` |
| `DOCKER_PASSWORD` | Docker Hub personal access token | `dckr_pat_...` |

### 2. GitHub Token (Automatic)

The `GITHUB_TOKEN` is automatically provided by GitHub Actions and doesn't need to be configured manually.

## Setting Up the Dev Environment

### Step 1: Create the Environment

1. Go to your repository: `https://github.com/dojima-foundation/tee-auth`
2. Click **Settings** tab
3. In the left sidebar, click **Environments**
4. Click **New environment**
5. Name it `dev`
6. Click **Configure environment**

### Step 2: Configure Environment Secrets

1. In the `dev` environment settings, scroll down to **Environment secrets**
2. Click **Add secret**
3. Add each secret:

   **DOCKER_USERNAME:**
   - Name: `DOCKER_USERNAME`
   - Value: `dojima0foundation`

   **DOCKER_PASSWORD:**
   - Name: `DOCKER_PASSWORD`
   - Value: Your Docker Hub personal access token

### Step 3: Configure Environment Variables (Optional)

You can also set environment-specific variables:

1. Scroll down to **Environment variables**
2. Click **Add variable**
3. Add variables like:

   | Variable Name | Value | Description |
   |---------------|-------|-------------|
   | `ENVIRONMENT` | `dev` | Current environment |
   | `DOCKER_REGISTRY` | `dojima0foundation` | Docker registry name |

### Step 4: Configure Protection Rules (Optional)

For the `dev` environment, you can set:

- **Required reviewers**: None (for dev)
- **Wait timer**: 0 minutes
- **Deployment branches**: Allow all branches

## Workflow Configuration

The workflow is configured with:

```yaml
jobs:
  renclave-v2:
    name: Renclave-v2
    runs-on: ubuntu-latest
    environment: dev  # Uses dev environment secrets
```

## Docker Image Tags

With the dev environment, Docker images will be tagged as:

- `dojima0foundation/renclave-v2:dev`
- `dojima0foundation/renclave-v2:abc1234` (commit SHA)
- `dojima0foundation/gauth:dev`
- `dojima0foundation/gauth:abc1234` (commit SHA)

## Adding More Environments

To add staging or production environments:

1. Create new environments in GitHub (staging, production)
2. Configure environment-specific secrets
3. Update the workflow to use different environments based on branch:

```yaml
environment: ${{ github.ref == 'refs/heads/main' && 'production' || 'dev' }}
```

## Testing the Setup

1. Push to the `test/local-testing-github-actions-clean` branch
2. Check the Actions tab to see the workflow running
3. Verify that Docker images are pushed with the correct tags
4. Check that all tests pass

## Troubleshooting

### Common Issues

1. **"Environment not found"**: Make sure the `dev` environment exists in your repository settings
2. **"Secret not found"**: Verify that `DOCKER_USERNAME` and `DOCKER_PASSWORD` are configured in the dev environment
3. **"Docker push failed"**: Check that your Docker Hub token has the correct permissions

### Verification Commands

You can verify your setup by checking:

```bash
# Check if environment exists
gh api repos/:owner/:repo/environments

# Check environment secrets (requires GitHub CLI)
gh secret list --env dev
```

## Security Best Practices

1. **Use Personal Access Tokens** instead of passwords
2. **Rotate secrets regularly**
3. **Use different tokens for different environments**
4. **Limit token permissions** to only what's needed
5. **Monitor secret usage** in the Actions logs

## Next Steps

After setting up the dev environment:

1. Test the workflow by pushing to your branch
2. Create staging and production environments as needed
3. Set up branch protection rules
4. Configure deployment notifications
5. Set up monitoring and alerting
