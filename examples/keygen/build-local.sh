#!/bin/bash
set -e

# Local build script for secure-packager-keygen
# This script builds the keygen container for the current architecture

IMAGE_NAME="secure-packager-keygen"
TAG="${1:-latest}"

echo "Building local Docker image: $IMAGE_NAME:$TAG"

# Build for current architecture
docker build \
    --tag $IMAGE_NAME:$TAG \
    --file Dockerfile.keygen \
    .

echo "✅ Local build completed!"
echo "Image: $IMAGE_NAME:$TAG"
echo ""
echo "To use this image:"
echo "  docker run --rm -v \$(pwd)/keys:/output $IMAGE_NAME:$TAG customer 2048"
