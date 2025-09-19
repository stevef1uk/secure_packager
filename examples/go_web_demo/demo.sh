#!/usr/bin/env bash
set -euo pipefail

# Demo script for Go Web Demo
# This script demonstrates the Go web interface for secure_packager

echo "🚀 Secure Packager Go Web Demo"
echo "==============================="
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is not installed or not in PATH"
    echo "Please install Docker and try again"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! command -v docker compose &> /dev/null; then
    echo "❌ Error: Docker Compose is not installed or not in PATH"
    echo "Please install Docker Compose and try again"
    exit 1
fi

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "📁 Working directory: $SCRIPT_DIR"
echo

# Function to check if docker-compose or docker compose is available
check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif command -v docker &> /dev/null && docker compose version &> /dev/null; then
        echo "docker compose"
    else
        echo ""
    fi
}

DOCKER_COMPOSE_CMD=$(check_docker_compose)
if [ -z "$DOCKER_COMPOSE_CMD" ]; then
    echo "❌ Error: Neither 'docker-compose' nor 'docker compose' is available"
    exit 1
fi

echo "🔧 Using: $DOCKER_COMPOSE_CMD"
echo

# Clean up previous containers
echo "🧹 Cleaning up previous containers..."
$DOCKER_COMPOSE_CMD down --remove-orphans 2>/dev/null || true

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p data output keys logs

# Build the keygen container first
echo "🔨 Building keygen container..."
docker build -t secure-packager-keygen:latest -f ../keygen/Dockerfile.keygen ../keygen/

# Generate keys before starting the web app
echo "🔑 Generating keys..."
./generate_keys.sh

echo "🔨 Building web demo container..."
$DOCKER_COMPOSE_CMD build

echo
echo "🚀 Starting services..."
echo "This will:"
echo "  - Start the Go web application"
echo "  - Make keygen container available"
echo "  - Set up volume mounts for data persistence"
echo

# Start the services
$DOCKER_COMPOSE_CMD up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
echo "🔍 Checking service status..."
$DOCKER_COMPOSE_CMD ps

echo
echo "🌐 Web Interface Available!"
echo "=========================="
echo
echo "📍 URL: http://localhost:8080"
echo
echo "🔧 Available Features:"
echo "  - Key Generation (RSA key pairs)"
echo "  - File Creation (sample files)"
echo "  - File Packaging (encryption)"
echo "  - License Token Issuance"
echo "  - File Unpacking (decryption)"
echo "  - File Browser (view files)"
echo "  - Complete Workflow (automated demo)"
echo
echo "📊 API Endpoints:"
echo "  - GET  /health - Health check"
echo "  - POST /api/keys/generate - Generate keys"
echo "  - POST /api/files/create - Create files"
echo "  - POST /api/package - Package files"
echo "  - POST /api/token/issue - Issue token"
echo "  - POST /api/unpack - Unpack files"
echo "  - GET  /api/files/:directory - List files"
echo "  - POST /api/files/read - Read file"
echo "  - POST /api/workflow/complete - Complete workflow"
echo
echo "🧪 Test the API:"
echo "  curl http://localhost:8080/health"
echo "  curl -X POST http://localhost:8080/api/workflow/complete"
echo
echo "📁 Local Directories:"
echo "  - Data:      $(pwd)/data/"
echo "  - Output:    $(pwd)/output/"
echo "  - Keys:      $(pwd)/keys/"
echo "  - Logs:      $(pwd)/logs/"
echo
echo "🛑 To stop services:"
echo "  $DOCKER_COMPOSE_CMD down"
echo
echo "🔧 To view logs:"
echo "  $DOCKER_COMPOSE_CMD logs -f web-demo"
echo
echo "📋 To rebuild and restart:"
echo "  $DOCKER_COMPOSE_CMD up --build -d"
echo

# Test the health endpoint
echo "🧪 Testing health endpoint..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Health check passed - service is running!"
else
    echo "❌ Health check failed - service may not be ready yet"
    echo "   Please wait a moment and try accessing http://localhost:8080"
fi

echo
echo "🎉 Demo is ready! Open http://localhost:8080 in your browser to start using the interface."
echo
echo "💡 Pro tip: Try the 'Complete Workflow' tab for a full demonstration!"
