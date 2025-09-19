#!/bin/bash
set -e

# Multi-architecture build script for secure-packager-keygen
# This script builds the keygen container for both AMD64 and ARM64 architectures

IMAGE_NAME="secure-packager-keygen"
TAG="${1:-latest}"

echo "Building multi-architecture Docker image: $IMAGE_NAME:$TAG"
echo "Architectures: linux/amd64, linux/arm64"

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

# Build for each architecture separately
echo "Building for AMD64..."
docker buildx build \
    --platform linux/amd64 \
    --tag $IMAGE_NAME:$TAG-amd64 \
    --file Dockerfile.keygen \
    --load \
    .

echo "Building for ARM64..."
docker buildx build \
    --platform linux/arm64 \
    --tag $IMAGE_NAME:$TAG-arm64 \
    --file Dockerfile.keygen \
    --load \
    .

echo "✅ Multi-architecture build completed!"
echo ""
echo "Built images:"
echo "  $IMAGE_NAME:$TAG-amd64 (AMD64)"
echo "  $IMAGE_NAME:$TAG-arm64 (ARM64)"
echo ""
echo "To use the appropriate image for your architecture:"
echo ""
echo "For AMD64 systems:"
echo "  docker run --rm -v \$(pwd)/keys:/output $IMAGE_NAME:$TAG-amd64 customer 2048"
echo ""
echo "For ARM64 systems:"
echo "  docker run --rm -v \$(pwd)/keys:/output $IMAGE_NAME:$TAG-arm64 customer 2048"
echo ""
echo "To create a multi-architecture manifest (requires Docker Hub access):"
echo "  docker manifest create $IMAGE_NAME:$TAG $IMAGE_NAME:$TAG-amd64 $IMAGE_NAME:$TAG-arm64"
echo "  docker manifest annotate $IMAGE_NAME:$TAG $IMAGE_NAME:$TAG-amd64 --arch amd64"
echo "  docker manifest annotate $IMAGE_NAME:$TAG $IMAGE_NAME:$TAG-arm64 --arch arm64"
