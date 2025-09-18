#!/usr/bin/env bash
set -euo pipefail

# Quick end-to-end demo for secure_packager using Docker
# - Generates vendor and customer RSA keys
# - Creates a sample input file
# - Packages without license and with license using Docker
# - Issues a vendor-signed token using Docker
# - Unpacks both zips (license/non-license) using Docker

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
WORK_DIR="$ROOT_DIR/tmp"
INPUT_DIR="$WORK_DIR/in"
OUT_NO_LIC="$WORK_DIR/out"
OUT_LIC="$WORK_DIR/out_license"
DEC_NO_LIC="$WORK_DIR/dec"
DEC_LIC="$WORK_DIR/dec_license"
KEYS_DIR="$WORK_DIR/keys"

VENDOR_PRIV="$KEYS_DIR/vendor_private.pem"
VENDOR_PUB="$KEYS_DIR/vendor_public.pem"
CUSTOMER_PRIV="$KEYS_DIR/customer_private.pem"
CUSTOMER_PUB="$OUT_NO_LIC/customer_public.pem"
TOKEN_PATH="$KEYS_DIR/token.txt"

DOCKER_IMAGE="stevef1uk/secure-packager:latest"

echo "[1/7] Prepare demo workspace..."
# Clean only the directories we'll use, preserve existing tmp structure
rm -rf "$OUT_LIC" "$DEC_LIC"
mkdir -p "$INPUT_DIR" "$OUT_NO_LIC" "$OUT_LIC" "$DEC_NO_LIC" "$DEC_LIC" "$KEYS_DIR"

echo "[2/7] Generate RSA keys (vendor + customer) using openssl..."
# Vendor keys (for token signing)
openssl genrsa -out "$VENDOR_PRIV" 2048 1>/dev/null 2>&1
openssl rsa -in "$VENDOR_PRIV" -pubout -out "$VENDOR_PUB" 1>/dev/null 2>&1

# Customer keys (for key unwrapping)
openssl genrsa -out "$CUSTOMER_PRIV" 2048 1>/dev/null 2>&1
openssl rsa -in "$CUSTOMER_PRIV" -pubout -out "$CUSTOMER_PUB" 1>/dev/null 2>&1

echo "[3/7] Create sample input files..."
echo "hello secure world" > "$INPUT_DIR/hello.txt"
dd if=/dev/urandom of="$INPUT_DIR/random.bin" bs=1k count=32 2>/dev/null

echo "[4/7] Pull Docker image..."
docker pull "$DOCKER_IMAGE"

echo "[5/7] Package WITHOUT licensing using Docker..."
docker run --rm \
  -v "$INPUT_DIR:/in" -v "$OUT_NO_LIC:/out" \
  "$DOCKER_IMAGE" \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true

echo "[6/7] Package WITH licensing using Docker (+ manifest + embedded vendor public key)..."
# Copy customer public key to the licensing output directory
cp "$CUSTOMER_PUB" "$OUT_LIC/customer_public.pem"
docker run --rm \
  -v "$INPUT_DIR:/in" -v "$OUT_LIC:/out" -v "$KEYS_DIR:/keys" \
  "$DOCKER_IMAGE" \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem

echo "Issue a vendor-signed token using Docker (expiry far in future for demo)..."
docker run --rm -v "$KEYS_DIR:/keys" \
  "$DOCKER_IMAGE" \
  issue-token -priv /keys/vendor_private.pem -expiry 2099-12-31 -company "Demo Co" -email "demo@example.com" -out /keys/token.txt

echo "[7/7] Unpack both zips using Docker..."
echo "- Unpack NO LICENSE zip..."
docker run --rm \
  -v "$OUT_NO_LIC:/out" -v "$DEC_NO_LIC:/dec" -v "$KEYS_DIR:/keys" \
  "$DOCKER_IMAGE" \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec

echo "- Unpack WITH LICENSE zip (auto-detect manifest, verify token)..."
docker run --rm \
  -v "$OUT_LIC:/out" -v "$DEC_LIC:/dec" -v "$KEYS_DIR:/keys" \
  "$DOCKER_IMAGE" \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt

echo "\nDemo completed. Outputs:"
echo "  No-license decrypted dir: $DEC_NO_LIC"
echo "  With-license decrypted dir: $DEC_LIC"
echo "  Token: $TOKEN_PATH"
echo "  Vendor pub: $VENDOR_PUB"
echo "  Customer pub: $CUSTOMER_PUB"
echo "\nAll outputs are in the tmp/ directory:"
echo "  ls -la tmp/"
