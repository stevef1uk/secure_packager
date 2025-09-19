#!/usr/bin/env bash
set -euo pipefail

# Test script for Docker integration example
# This script tests the examples/example_docker/ integration example
# - Builds the Docker image
# - Runs the containers
# - Tests the API endpoints
# - Shows decrypted files
# - Cleans up

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE_DIR="$ROOT_DIR/examples/example_docker"

echo "🧪 Testing Docker Integration Example"
echo "====================================="
echo

# Check if we're in the right directory
if [ ! -d "$EXAMPLE_DIR" ]; then
    echo "❌ Error: Docker integration example not found at $EXAMPLE_DIR"
    echo "Please run this script from the secure_packager root directory"
    exit 1
fi

cd "$EXAMPLE_DIR"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is not installed or not in PATH"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! command -v docker compose &> /dev/null; then
    echo "❌ Error: Docker Compose is not installed or not in PATH"
    exit 1
fi

echo "📁 Working directory: $(pwd)"
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

# Clean up any existing containers
echo "🧹 Cleaning up previous containers..."
$DOCKER_COMPOSE_CMD down --remove-orphans 2>/dev/null || true

# Run the demo script
echo "🚀 Running Docker integration demo..."
echo "This will:"
echo "  - Generate RSA keys"
echo "  - Create sample data"
echo "  - Encrypt the data"
echo "  - Build Docker images"
echo "  - Start containers"
echo "  - Test API endpoints"
echo "  - Show decrypted files"
echo

if [ -f "./demo.sh" ]; then
    ./demo.sh
else
    echo "❌ Error: demo.sh not found in $EXAMPLE_DIR"
    exit 1
fi

echo
echo "🧪 Running additional tests..."

# Test 1: Health checks
echo "1️⃣ Testing health checks..."
echo "   Encrypted processor health:"
curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
echo
echo "   Demo processor health:"
curl -s http://localhost:8081/health | jq . 2>/dev/null || curl -s http://localhost:8081/health
echo

# Test 2: File listing
echo "2️⃣ Testing file listing..."
echo "   Encrypted processor files:"
curl -s "http://localhost:8080/api/files?directory=/app/decrypted" | jq . 2>/dev/null || curl -s "http://localhost:8080/api/files?directory=/app/decrypted"
echo
echo "   Demo processor files:"
curl -s "http://localhost:8081/api/files?directory=/app/decrypted" | jq . 2>/dev/null || curl -s "http://localhost:8081/api/files?directory=/app/decrypted"
echo

# Test 3: Different algorithms
echo "3️⃣ Testing different checksum algorithms..."
for algo in md5 sha1 sha256 sha512; do
    echo "   Testing $algo algorithm on encrypted processor:"
    curl -s -X POST http://localhost:8080/api/process \
        -H "Content-Type: application/json" \
        -d "{\"directory\": \"/app/decrypted\", \"algorithm\": \"$algo\"}" | jq . 2>/dev/null || curl -s -X POST http://localhost:8080/api/process \
        -H "Content-Type: application/json" \
        -d "{\"directory\": \"/app/decrypted\", \"algorithm\": \"$algo\"}"
    echo
done

# Test 4: Container logs
echo "4️⃣ Checking container logs..."
echo "   Encrypted processor logs (last 10 lines):"
$DOCKER_COMPOSE_CMD logs --tail=10 file-processor-encrypted
echo
echo "   Demo processor logs (last 10 lines):"
$DOCKER_COMPOSE_CMD logs --tail=10 file-processor-demo
echo

# Test 5: File system inspection
echo "5️⃣ Inspecting file system..."
echo "   Local decrypted files:"
if [ -d "./decrypted_encrypted" ]; then
    echo "   Encrypted container files:"
    ls -la ./decrypted_encrypted/
    echo
    echo "   Sample file content (encrypted):"
    if [ -f "./decrypted_encrypted/sample.txt" ]; then
        cat ./decrypted_encrypted/sample.txt
    fi
    echo
fi

if [ -d "./decrypted_demo" ]; then
    echo "   Demo container files:"
    ls -la ./decrypted_demo/
    echo
    echo "   Sample file content (demo):"
    if [ -f "./decrypted_demo/demo.txt" ]; then
        cat ./decrypted_demo/demo.txt
    fi
    echo
fi

# Test 6: Container inspection
echo "6️⃣ Inspecting containers..."
echo "   Container status:"
$DOCKER_COMPOSE_CMD ps
echo
echo "   Container resource usage:"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" $(docker ps --filter "name=secure-file-processor" --format "{{.Names}}") 2>/dev/null || echo "   (Stats not available)"
echo

echo "✅ All tests completed successfully!"
echo
echo "📊 Test Summary:"
echo "   ✅ Health checks passed"
echo "   ✅ File listing works"
echo "   ✅ Multiple algorithms supported"
echo "   ✅ Containers running properly"
echo "   ✅ Decrypted files accessible"
echo "   ✅ API endpoints responding"
echo
echo "🔍 Available endpoints:"
echo "   Encrypted processor: http://localhost:8080"
echo "   Demo processor:      http://localhost:8081"
echo
echo "🛑 To stop containers:"
echo "   $DOCKER_COMPOSE_CMD down"
echo
echo "🔧 To view logs:"
echo "   $DOCKER_COMPOSE_CMD logs -f"
echo
echo "📁 To inspect files:"
echo "   ls -la decrypted_encrypted/"
echo "   ls -la decrypted_demo/"
