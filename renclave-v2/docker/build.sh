#!/bin/bash

# Build script to work around Docker Buildx bake mode issues
# This script builds the Docker image manually and then uses docker compose

set -e

echo "🐳 Building renclave test image manually..."

# Build the image using traditional docker build (no Buildx)
docker build -f docker/Dockerfile.test -t renclave-test-runner:latest .

# Tag the image to avoid Docker Hub pull attempts
docker tag renclave-test-runner:latest localhost/renclave-test-runner:latest

echo "✅ Image built and tagged successfully!"
echo "🚀 Now you can run docker compose commands normally"
