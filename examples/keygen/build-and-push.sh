#!/bin/bash
set -e

# Multi-architecture build and push script for secure-packager-keygen
# This script builds the keygen container for both AMD64 and ARM64 architectures
# and pushes to Docker Hub

# Configuration
DOCKERHUB_USER="${DOCKERHUB_USER:-}"
IMAGE_NAME="secure-packager-keygen"
TAG="${1:-latest}"

# Check if Docker Hub username is provided
if [ -z "$DOCKERHUB_USER" ]; then
    echo "Error: DOCKERHUB_USER environment variable is required"
    echo "Usage: DOCKERHUB_USER=yourusername ./build-and-push.sh [tag]"
    echo "Example: DOCKERHUB_USER=stevef1uk ./build-and-push.sh latest"
    exit 1
fi

FULL_IMAGE_NAME="$DOCKERHUB_USER/$IMAGE_NAME"

echo "Building and pushing multi-architecture Docker image: $FULL_IMAGE_NAME:$TAG"
echo "Architectures: linux/amd64, linux/arm64"
echo "Docker Hub User: $DOCKERHUB_USER"

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "Error: Docker buildx is not available. Please install Docker buildx or update Docker to a newer version."
    echo "For Docker Desktop: Enable buildx in settings"
    echo "For Linux: Install buildx plugin"
    exit 1
fi

# Create a new builder instance if it doesn't exist
BUILDER_NAME="secure-packager-builder"
if ! docker buildx inspect $BUILDER_NAME >/dev/null 2>&1; then
    echo "Creating new buildx builder: $BUILDER_NAME"
    docker buildx create --name $BUILDER_NAME --use
else
    echo "Using existing buildx builder: $BUILDER_NAME"
    docker buildx use $BUILDER_NAME
fi

# Login to Docker Hub
echo "Logging in to Docker Hub..."
docker login

# Build and push multi-architecture image
echo "Building and pushing multi-architecture image..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag $FULL_IMAGE_NAME:$TAG \
    --file Dockerfile.keygen \
    --push \
    .

echo "✅ Multi-architecture build and push completed!"
echo "Image: $FULL_IMAGE_NAME:$TAG"
echo "Architectures: linux/amd64, linux/arm64"
echo ""
echo "To use this image:"
echo "  docker run --rm -v \$(pwd)/keys:/output $FULL_IMAGE_NAME:$TAG customer 2048"
echo ""
echo "To update the Go UI to use your image, update the following files:"
echo "  - examples/go_web_demo/generate_keys.sh"
echo "  - examples/go_web_demo/demo.sh"
echo "  - examples/go_web_demo/docker-compose.yml"
echo ""
echo "Replace 'secure-packager-keygen:latest' with '$FULL_IMAGE_NAME:$TAG'"
