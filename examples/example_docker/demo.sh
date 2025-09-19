#!/usr/bin/env bash
set -euo pipefail

# Demo script for secure_packager Docker integration
# This script demonstrates how to use Docker containers with secure_packager
# for decryption before starting the main application

echo "🐳 Secure Packager Docker Integration Demo"
echo "=========================================="
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
mkdir -p data keys demo_data logs

# Generate RSA keys if they don't exist
if [ ! -f "keys/customer_private.pem" ] || [ ! -f "keys/customer_public.pem" ]; then
    echo "🔑 Generating RSA key pairs..."
    openssl genrsa -out keys/customer_private.pem 2048 2>/dev/null
    openssl rsa -in keys/customer_private.pem -pubout -out keys/customer_public.pem 2>/dev/null
    echo "   Generated customer key pair"
fi

if [ ! -f "keys/vendor_private.pem" ] || [ ! -f "keys/vendor_public.pem" ]; then
    echo "🔑 Generating vendor key pairs..."
    openssl genrsa -out keys/vendor_private.pem 2048 2>/dev/null
    openssl rsa -in keys/vendor_private.pem -pubout -out keys/vendor_public.pem 2>/dev/null
    echo "   Generated vendor key pair"
fi

# Create sample data for encryption
echo "📄 Creating sample data..."
mkdir -p data
cat > data/sample.txt << 'EOF'
This is a sample file for demonstration.
It contains some text that will be encrypted.
The secure_packager will encrypt this file and
the Docker container will decrypt it before
the application starts processing it.
EOF

cat > data/config.json << 'EOF'
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "secure_db"
  },
  "api": {
    "version": "1.0",
    "endpoint": "/api/v1"
  }
}
EOF

echo "   Created sample files"

# Create demo data (unencrypted)
echo "📄 Creating demo data..."
mkdir -p demo_data
cat > demo_data/demo.txt << 'EOF'
This is demo data that doesn't require decryption.
It's used to show the application working without
encrypted files.
EOF

echo "   Created demo files"

# Encrypt the sample data
echo "🔐 Encrypting sample data..."
if command -v go &> /dev/null; then
    # Use local secure_packager if available
    if [ -f "../../cmd/packager/main.go" ]; then
        echo "   Using local secure_packager..."
        cd ../../
        go build -o /tmp/packager ./cmd/packager
        go build -o /tmp/issue-token ./cmd/issue-token
        cd examples/example_docker
        
        /tmp/packager -in data -out data -pub keys/customer_public.pem -zip=true -license -vendor-pub keys/vendor_public.pem
        /tmp/issue-token -priv keys/vendor_private.pem -expiry 2025-12-31 -company "Demo Co" -email "demo@example.com" -out keys/token.txt
        
        rm -f /tmp/packager /tmp/issue-token
    else
        echo "   Using Docker to encrypt data..."
        docker run --rm -v "$(pwd)/data:/in" -v "$(pwd)/keys:/keys" \
            stevef1uk/secure-packager:latest \
            packager -in /in -out /in -pub /keys/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem
        
        docker run --rm -v "$(pwd)/keys:/keys" \
            stevef1uk/secure-packager:latest \
            issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company "Demo Co" -email "demo@example.com" -out /keys/token.txt
    fi
else
    echo "   Using Docker to encrypt data..."
    docker run --rm -v "$(pwd)/data:/in" -v "$(pwd)/keys:/keys" \
        stevef1uk/secure-packager:latest \
        packager -in /in -out /in -pub /keys/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem
    
    docker run --rm -v "$(pwd)/keys:/keys" \
        stevef1uk/secure-packager:latest \
        issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company "Demo Co" -email "demo@example.com" -out /keys/token.txt
fi

echo "   Data encrypted successfully"

# Build the Docker image
echo "🔨 Building Docker image..."
$DOCKER_COMPOSE_CMD build

echo
echo "🚀 Starting containers..."
echo

# Start the encrypted version
echo "1️⃣ Starting encrypted file processor (port 8080)..."
$DOCKER_COMPOSE_CMD up -d file-processor-encrypted

# Start the demo version
echo "2️⃣ Starting demo file processor (port 8081)..."
$DOCKER_COMPOSE_CMD up -d file-processor-demo

# Wait for containers to be ready
echo
echo "⏳ Waiting for containers to be ready..."
sleep 10

# Test the encrypted version
echo
echo "🧪 Testing encrypted file processor..."
echo "Health check:"
curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health

echo
echo "Processing encrypted files:"
curl -s -X POST http://localhost:8080/api/process \
    -H "Content-Type: application/json" \
    -d '{"directory": "/app/decrypted", "algorithm": "sha256"}' | jq . 2>/dev/null || curl -s -X POST http://localhost:8080/api/process \
    -H "Content-Type: application/json" \
    -d '{"directory": "/app/decrypted", "algorithm": "sha256"}'

echo
echo "🧪 Testing demo file processor..."
echo "Health check:"
curl -s http://localhost:8081/health | jq . 2>/dev/null || curl -s http://localhost:8081/health

echo
echo "Processing demo files:"
curl -s -X POST http://localhost:8081/api/process \
    -H "Content-Type: application/json" \
    -d '{"directory": "/app/decrypted", "algorithm": "sha256"}' | jq . 2>/dev/null || curl -s -X POST http://localhost:8081/api/process \
    -H "Content-Type: application/json" \
    -d '{"directory": "/app/decrypted", "algorithm": "sha256"}'

echo
echo "✅ Demo completed successfully!"
echo
echo "📊 Results:"
echo "   Encrypted processor: http://localhost:8080"
echo "   Demo processor:      http://localhost:8081"
echo
echo "🔍 Available endpoints:"
echo "   GET  /health - Health check"
echo "   POST /api/process - Process files in directory"
echo "   GET  /api/files?directory=<path> - List files in directory"
echo
echo "📁 Local File System:"
echo "   Encrypted data: $(pwd)/data/"
echo "   Demo data:      $(pwd)/demo_data/"
echo "   Keys:           $(pwd)/keys/"
echo "   Logs:           $(pwd)/logs/"
echo
echo "🔍 Decrypted Files Location:"
echo "   The decrypted files are stored INSIDE the Docker containers at:"
echo "   - Encrypted processor: /app/decrypted/"
echo "   - Demo processor:      /app/decrypted/"
echo
echo "📋 To view decrypted files locally:"
echo "   # Copy decrypted files from encrypted container:"
echo "   docker cp secure-file-processor-encrypted:/app/decrypted ./decrypted_encrypted/"
echo "   # Copy decrypted files from demo container:"
echo "   docker cp secure-file-processor-demo:/app/decrypted ./decrypted_demo/"
echo
echo "📂 Current local directory contents:"
ls -la

echo
echo "📋 Copying decrypted files to local filesystem for inspection..."
echo "   Creating local decrypted directories..."

# Create local directories for decrypted files
mkdir -p decrypted_encrypted decrypted_demo

# Copy decrypted files from containers
echo "   Copying from encrypted container..."
docker cp secure-file-processor-encrypted:/app/decrypted/. ./decrypted_encrypted/ 2>/dev/null || echo "   (Container not running or files not available)"

echo "   Copying from demo container..."
docker cp secure-file-processor-demo:/app/decrypted/. ./decrypted_demo/ 2>/dev/null || echo "   (Container not running or files not available)"

echo
echo "📁 Decrypted files now available locally:"
echo "   Encrypted files: $(pwd)/decrypted_encrypted/"
echo "   Demo files:      $(pwd)/decrypted_demo/"
echo
echo "📋 Contents of decrypted files:"
echo "   Encrypted container files:"
ls -la decrypted_encrypted/ 2>/dev/null || echo "   (No files found)"
echo
echo "   Demo container files:"
ls -la decrypted_demo/ 2>/dev/null || echo "   (No files found)"

echo
echo "🛑 To stop containers:"
echo "   $DOCKER_COMPOSE_CMD down"
echo
echo "🔧 To view logs:"
echo "   $DOCKER_COMPOSE_CMD logs -f file-processor-encrypted"
echo "   $DOCKER_COMPOSE_CMD logs -f file-processor-demo"
