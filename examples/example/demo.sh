#!/usr/bin/env bash
set -euo pipefail

# Demo script for secure_packager integration examples
# This script demonstrates both the checksum calculator and integration example

echo "🚀 Secure Packager Integration Examples Demo"
echo "==========================================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go 1.21 or later and try again"
    exit 1
fi

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK_DIR="$SCRIPT_DIR/demo_work"

echo "📁 Working directory: $WORK_DIR"
echo

# Clean up previous demo if it exists
if [ -d "$WORK_DIR" ]; then
    echo "🧹 Cleaning up previous demo..."
    rm -rf "$WORK_DIR"
fi

# Create work directory
mkdir -p "$WORK_DIR"

echo "1️⃣ Running Checksum Calculator Demo"
echo "===================================="
cd "$SCRIPT_DIR/checksum"

# Create some sample files for the checksum demo
mkdir -p "$WORK_DIR/sample_files"
echo "Hello, World!" > "$WORK_DIR/sample_files/hello.txt"
echo "This is a test file" > "$WORK_DIR/sample_files/test.txt"
echo "Binary data: $(head -c 1000 /dev/urandom | base64)" > "$WORK_DIR/sample_files/binary.dat"

echo "Running checksum calculator on sample files..."
go run main.go -dir "$WORK_DIR/sample_files" -algo sha256

echo
echo "2️⃣ Running Integration Example (No License)"
echo "==========================================="
cd "$SCRIPT_DIR/integration"

echo "Running integration example without licensing..."
go run main.go -work "$WORK_DIR/no_license"

echo
echo "3️⃣ Running Integration Example (With License)"
echo "============================================="
echo "Running integration example with licensing..."
go run main.go -work "$WORK_DIR/with_license" -license

echo
echo "✅ Demo completed successfully!"
echo
echo "📊 Results:"
echo "   Checksum demo files: $WORK_DIR/sample_files"
echo "   No-license demo: $WORK_DIR/no_license"
echo "   With-license demo: $WORK_DIR/with_license"
echo
echo "🔍 You can inspect the generated files and directories to see the results."
echo "   - 'data' directories contain the original files"
echo "   - 'encrypted' directories contain the encrypted zip files"
echo "   - 'decrypted' directories contain the decrypted files"
echo "   - 'keys' directories contain the RSA key pairs"
