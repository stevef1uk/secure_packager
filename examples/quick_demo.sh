#!/usr/bin/env bash
set -euo pipefail

# Quick end-to-end demo for secure_packager
# - Generates vendor and customer RSA keys
# - Creates a sample input file
# - Packages without license and with license
# - Issues a vendor-signed token
# - Unpacks both zips (license/non-license)

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$ROOT_DIR"
WORK_DIR="$ROOT_DIR/_demo"
INPUT_DIR="$WORK_DIR/input"
OUT_NO_LIC="$WORK_DIR/out_no_license"
OUT_LIC="$WORK_DIR/out_with_license"
DEC_NO_LIC="$WORK_DIR/decrypted_no_license"
DEC_LIC="$WORK_DIR/decrypted_with_license"

VENDOR_PRIV="$WORK_DIR/vendor_private.pem"
VENDOR_PUB="$WORK_DIR/vendor_public.pem"
CUSTOMER_PRIV="$WORK_DIR/customer_private.pem"
CUSTOMER_PUB="$WORK_DIR/customer_public.pem"
TOKEN_PATH="$WORK_DIR/token.txt"

echo "[1/7] Prepare demo workspace..."
rm -rf "$WORK_DIR"
mkdir -p "$INPUT_DIR" "$OUT_NO_LIC" "$OUT_LIC" "$DEC_NO_LIC" "$DEC_LIC"

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

echo "[4/7] Build tools..."
pushd "$BUILD_DIR" >/dev/null
go build ./cmd/packager
go build ./cmd/unpack
go build ./cmd/issue-token
popd >/dev/null

echo "[5/7] Package WITHOUT licensing..."
"$BUILD_DIR"/packager -in "$INPUT_DIR" -out "$OUT_NO_LIC" -pub "$CUSTOMER_PUB" -zip=true

echo "[6/7] Package WITH licensing (+ manifest + embedded vendor public key)..."
"$BUILD_DIR"/packager -in "$INPUT_DIR" -out "$OUT_LIC" -pub "$CUSTOMER_PUB" -zip=true -license -vendor-pub "$VENDOR_PUB"

echo "Issue a vendor-signed token (expiry far in future for demo)..."
"$BUILD_DIR"/issue-token -priv "$VENDOR_PRIV" -expiry 2099-12-31 -company "Demo Co" -email "demo@example.com" -out "$TOKEN_PATH"

echo "[7/7] Unpack both zips..."
echo "- Unpack NO LICENSE zip..."
"$BUILD_DIR"/unpack -zip "$OUT_NO_LIC/encrypted_files.zip" -priv "$CUSTOMER_PRIV" -out "$DEC_NO_LIC"

echo "- Unpack WITH LICENSE zip (auto-detect manifest, verify token)..."
"$BUILD_DIR"/unpack -zip "$OUT_LIC/encrypted_files.zip" -priv "$CUSTOMER_PRIV" -out "$DEC_LIC" \
  -license-token "$TOKEN_PATH"

echo "\nDemo completed. Outputs:"
echo "  No-license decrypted dir: $DEC_NO_LIC"
echo "  With-license decrypted dir: $DEC_LIC"
echo "  Token: $TOKEN_PATH"
echo "  Vendor pub: $VENDOR_PUB"
echo "  Customer pub: $CUSTOMER_PUB"


