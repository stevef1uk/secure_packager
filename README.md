## secure_packager
![secure_packager logo](https://raw.githubusercontent.com/stevef1uk/secure_packager/main/images/secure_packager.png)

### Why this exists

Modern teams need to ship valuable data/models to customers securely, without painful key management or custom builds on every machine. Plain zips leak keys, signatures don’t enforce access, and full-blown DRM is heavy and brittle. This project offers a pragmatic middle ground:

- **Problem**: You must distribute files inside containers or via zip, but only the intended recipient should be able to open them.
- **Solution**: Encrypt files with a symmetric key (Fernet), then wrap that key with the recipient’s RSA public key. Only their private key can unwrap and decrypt.
- **Optional licensing**: Add a vendor-signed token that’s verified at decrypt time for friendly messaging and basic enforcement (expiry, warnings, block within 24h).

### Key features

- **Confidentiality by default**: No plaintext Fernet key shipped.
- **Two modes**: with or without licensing enforcement (auto-detected from the zip).
- **Simple CLI or Docker**: Use locally or via container with volume mounts.
- **Portable**: Multi-arch container images (linux/amd64, linux/arm64).

Note: the recipient of the file should create the key pair and send the public key to be used to create the archive.


Envelope encryption utilities for distributing data/models:
- Packager: Fernet-encrypts files; wraps the Fernet key with customer's RSA public key (RSA-OAEP SHA-256)
- Unpack: Requires customer's RSA private key to unwrap key and decrypt files
- Issue-token: Generates vendor-signed license tokens for messaging/enforcement

### Architecture

![secure_packager architecture](https://raw.githubusercontent.com/stevef1uk/secure_packager/main/images/architecture.png)

### Quick demo (end-to-end)

Script: `secure_packager/examples/quick_demo.sh`

What it does:
- Generates vendor and customer RSA keys (OpenSSL)
- Builds tools
- Packages with and without licensing
- Issues a long-lived vendor-signed token
- Unpacks both zips; license flow is auto-enforced for the licensed one

Run:
```
./examples/quick_demo.sh
```

Outputs (under `secure_packager/tmp`):
- `out/encrypted_files.zip` and decrypted `dec/`
- `out_license/encrypted_files.zip` and decrypted `dec_license/`
- `keys/token.txt`, `keys/vendor_public.pem`, `keys/customer_private.pem`

### Quick demo with Docker CLI

Script: `secure_packager/examples/quick_demo_cli_docker.sh`

Same functionality as the regular demo but uses the released Docker container instead of building locally. Useful for testing without Go installation or for CI/CD environments.

What it does:
- Generates vendor and customer RSA keys (OpenSSL)
- Pulls the Docker image
- Packages with and without licensing using Docker CLI
- Issues a long-lived vendor-signed token using Docker CLI
- Unpacks both zips using Docker CLI; license flow is auto-enforced for the licensed one

Run:
```
./examples/quick_demo_cli_docker.sh
```

### Test Docker Integration Example

Script: `secure_packager/examples/test_docker_integration.sh`

Comprehensive test script for the Docker integration example that demonstrates containerized deployment with entrypoint decryption.

What it does:
- Runs the complete Docker integration demo
- Tests all API endpoints
- Verifies health checks
- Tests multiple checksum algorithms
- Inspects container logs and resources
- Shows decrypted files locally

Run:
```
./examples/test_docker_integration.sh
```

Outputs (under `secure_packager/tmp`):
- `out/encrypted_files.zip` and decrypted `dec/`
- `out_license/encrypted_files.zip` and decrypted `dec_license/`
- `keys/token.txt`, `keys/vendor_public.pem`, `keys/customer_private.pem`


### Build

```
cd secure_packager
go build ./cmd/packager
go build ./cmd/unpack
go build ./cmd/issue-token
```

### Docker (multi-arch)

Build multi-arch image (requires buildx):
```
cd secure_packager
docker buildx build --platform linux/amd64,linux/arm64 -t yourorg/secure-packager:latest --push .
```

Run examples (volume mount input/output):
```
# Packager (no licensing):
docker run --rm -v $(pwd)/input:/in -v $(pwd)/out:/out \
  yourorg/secure-packager:latest packager -in /in -out /out -pub /out/customer_public.pem -zip=true

# Packager (with licensing):
docker run --rm -v $(pwd)/input:/in -v $(pwd)/out:/out -v $(pwd)/keys:/keys \
  yourorg/secure-packager:latest packager -in /in -out /out -pub /out/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem

# Unpack (auto-detect licensing from zip):
docker run --rm -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  yourorg/secure-packager:latest unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt

# Issue token:
docker run --rm -v $(pwd)/keys:/keys yourorg/secure-packager:latest \
  issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company Acme -email ops@acme.com -out /keys/token.txt
```

### Docker quickstart (clean tmp workspace)

If you prefer to set everything up under a clean `tmp` directory and keep your project tree untouched:

1) Create a temporary workspace and directories
```
mkdir -p tmp && cd tmp
mkdir -p in out dec keys env
```

2) Create required keys
- Customer keypair (required in all modes; public key used by packager, private key used by unpacker)
```
openssl genrsa -out keys/customer_private.pem 2048
openssl rsa -in keys/customer_private.pem -pubout -out out/customer_public.pem
```
- Vendor keypair (only needed for licensing mode; signs license tokens)
```
openssl genrsa -out keys/vendor_private.pem 2048
openssl rsa -in keys/vendor_private.pem -pubout -out keys/vendor_public.pem
```

3) Add files to encrypt
```
echo "demo secret" > in/demo.txt
```

4) Package without licensing
```
docker run --rm \
  -v $(pwd)/in:/in -v $(pwd)/out:/out \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true
```

5) Package with licensing (optional)
```
docker run --rm \
  -v $(pwd)/in:/in -v $(pwd)/out:/out -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /out -pub /out/customer_public.pem -zip=true -license -vendor-pub /keys/vendor_public.pem
```

6) Issue a token (licensing mode only)
```
docker run --rm -v $(pwd)/keys:/keys stevef1uk/secure-packager:latest \
  issue-token -priv /keys/vendor_private.pem -expiry 2025-12-31 -company Acme -email ops@acme.com -out /keys/token.txt
```

7) Unpack (auto-detects licensing from zip)
```
# no licensing
docker run --rm \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec

# with licensing
docker run --rm \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt
```

Tip: Simulate expiry by setting FAKE_NOW using an env file:
```
echo FAKE_NOW=2100-01-01 > env/.env
docker run --rm --env-file $(pwd)/env/.env \
  -v $(pwd)/out:/out -v $(pwd)/dec:/dec -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  unpack -zip /out/encrypted_files.zip -priv /keys/customer_private.pem -out /dec -license-token /keys/token.txt
```

### Modes

- Without licensing: default; zip contains encrypted files and `wrapped_key.bin` only
- With licensing: add manifest and vendor public key; unzip enforces license automatically

### Package (no licensing)

```
./packager -in ./input_dir -out ./out_dir -pub ./customer_public.pem -zip=true
```

### Package (license required)

```
./packager -in ./input_dir -out ./out_dir -pub ./customer_public.pem -zip=true \
  -license -vendor-pub ./vendor_public.pem
```

Outputs add:
- `manifest.json` with `{ "license_required": true, "vendor_public_key": "vendor_public.pem" }`
- `vendor_public.pem`

### Unpack (auto-detects licensing from zip)

```
# no licensing required in zip
./unpack -zip ./out_dir/encrypted_files.zip -priv ./customer_private.pem -out ./decrypted

# licensing required in zip (manifest present)
./unpack -zip ./out_dir/encrypted_files.zip -priv ./customer_private.pem -out ./decrypted \
  -license-token ./token.txt                # required
# -vendor-pub optional; defaults to vendor_public.pem inside zip when present
```

License token format (compatible with existing):
- `base64url(expiry:company:email:placeholder_key:signature_b64)`
- Signature: RSA-PSS over `expiry:company:email:placeholder_key`
- Behavior: prints Company/Email/Expiry, warns at <=7 days, blocks if expired or <=24h (supports `FAKE_NOW`)

### Issue a license token

```
./issue-token -priv ./vendor_private.pem -expiry 2025-12-31 -company "Acme" -email "ops@acme.com" -out ./token.txt
```

### Manual steps for first-time test

1) Generate keys (OpenSSL):
```
openssl genrsa -out vendor_private.pem 2048
openssl rsa -in vendor_private.pem -pubout -out vendor_public.pem
openssl genrsa -out customer_private.pem 2048
openssl rsa -in customer_private.pem -pubout -out customer_public.pem
```

2) Build tools:
```
cd secure_packager
go build ./cmd/packager && go build ./cmd/unpack && go build ./cmd/issue-token
```

3) Package without licensing:
```
./packager -in ./input_dir -out ./out_no_license -pub ./customer_public.pem -zip=true
./unpack -zip ./out_no_license/encrypted_files.zip -priv ./customer_private.pem -out ./dec_no_lic
```

4) Package with licensing and issue token:
```
./packager -in ./input_dir -out ./out_with_license -pub ./customer_public.pem -zip=true -license -vendor-pub ./vendor_public.pem
./issue-token -priv ./vendor_private.pem -expiry 2099-12-31 -company "Demo Co" -email "demo@example.com" -out ./token.txt
./unpack -zip ./out_with_license/encrypted_files.zip -priv ./customer_private.pem -out ./dec_with_lic -license-token ./token.txt
```

### Notes
- RSA key size >= 2048 recommended
- Only the private key holder can unwrap the Fernet key
- No Fernet key in plaintext is shipped

## Integration Examples

This project includes two comprehensive integration examples that demonstrate how to use `secure_packager` in real-world applications:

### 1. Library Integration Example (`examples/example/`)

**What it demonstrates:**
- How to integrate `secure_packager` as a Go library
- File processing with checksum calculation
- Complete workflow: encrypt → decrypt → verify
- Both licensing and non-licensing scenarios

**Key Components:**
- **Checksum Calculator** (`checksum/`): Standalone program for file checksumming
- **Integration Example** (`integration/`): Complete demo showing library integration
- **Demo Script** (`demo.sh`): Automated demonstration

**Quick Start:**
```bash
cd examples/example
./demo.sh
```

**Manual Usage:**
```bash
# Checksum calculator
cd examples/example/checksum
go run main.go -dir ./data -algo sha256

# Integration example
cd examples/example/integration
go run main.go -work ./demo_work
go run main.go -work ./demo_work -license
```

### 2. Docker Integration Example (`examples/example_docker/`)

**What it demonstrates:**
- How to use `secure_packager` with Docker containers
- Entrypoint pattern for decryption before application startup
- Production-ready containerized deployment
- HTTP API for file processing

**Key Components:**
- **Dockerfile**: Multi-stage build with security hardening
- **Entrypoint** (`entrypoint/`): Decryption and application startup
- **Main App** (`app/`): File processing HTTP API
- **Docker Compose**: Multiple service configurations
- **Demo Script** (`demo.sh`): Automated testing

**Quick Start:**
```bash
cd examples/example_docker
./demo.sh
```

**Manual Usage:**
```bash
# Build and run with Docker Compose
cd examples/example_docker
docker-compose up --build

# Test the API
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"directory": "/app/decrypted", "algorithm": "sha256"}'
```

### Integration Patterns

#### Library Integration Pattern
```go
// 1. Create packager
packager, err := NewSecurePackager("customer_public.pem")

// 2. Encrypt directory
err = packager.EncryptDirectory("./data", "./encrypted", false)

// 3. Decrypt zip
err = packager.DecryptZip("./encrypted/encrypted_files.zip", "./decrypted", "customer_private.pem", "")
```

#### Docker Integration Pattern
```dockerfile
# Multi-stage build
FROM golang:1.21-bookworm AS build
# Build secure_packager tools
# Build your application
# Build entrypoint

FROM debian:bookworm-slim
# Copy binaries
# Set entrypoint
ENTRYPOINT ["/app/entrypoint"]
```

### Use Cases

**Library Integration** is ideal for:
- Applications that need direct programmatic access
- Development and testing environments
- Custom file processing workflows
- Integration with existing Go applications

**Docker Integration** is ideal for:
- Production deployments
- Containerized applications
- Microservices architecture
- CI/CD pipelines
- Kubernetes deployments

### API Endpoints (Docker Example)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/process` | Process files in directory |
| GET | `/api/files` | List files in directory |

### Security Features

Both examples demonstrate:
- **Envelope Encryption**: RSA + Fernet encryption
- **No Plaintext Keys**: Fernet key never stored in plaintext
- **Key Isolation**: Only private key holder can decrypt
- **Optional Licensing**: Vendor-signed tokens for access control
- **File Integrity**: Checksum verification

### Getting Started

1. **Choose your integration approach**:
   - Library integration for Go applications
   - Docker integration for containerized deployments

2. **Run the examples**:
   - Follow the quick start guides above
   - Examine the code to understand the patterns

3. **Customize for your use case**:
   - Modify file processing logic
   - Add your own business logic
   - Implement custom license verification

4. **Deploy to production**:
   - Use the Docker example as a template
   - Implement proper key management
   - Add monitoring and logging

### Troubleshooting

- **Library Integration**: Check Go module dependencies and file permissions
- **Docker Integration**: Verify container logs and volume mounts
- **Both**: Ensure RSA keys are properly formatted and accessible

For detailed documentation, see the README files in each example directory.

### 3. Go Web UI Demo (`examples/go_web_demo/`)

**What it demonstrates:**
- Complete web-based interface for `secure_packager`
- Interactive file management and encryption workflows
- Real-time file upload, packaging, and unpacking
- License token management with expiry controls
- File browser with download capabilities

**Key Features:**
- **Web Interface**: Modern, responsive UI built with Go and Bootstrap
- **File Management**: Upload, view, and download files
- **Complete Workflow**: Key generation → File creation → Packaging → Token issuance → Unpacking
- **Licensing Support**: Full licensing workflow with expiry management
- **File Browser**: View and download encrypted/decrypted files
- **Docker Integration**: Containerized deployment with pre-generated keys

**Key Components:**
- **Go Web Server** (`main.go`): Gin-based web application
- **HTML Templates** (`templates/`): Responsive web interface
- **Static Assets** (`static/`): CSS, JavaScript, and Bootstrap
- **Key Generation** (`keygen/`): OpenSSL-based key generation utilities
- **Docker Compose**: Multi-container orchestration

**Quick Start:**
```bash
cd examples/go_web_demo
./demo.sh
```

**Access the UI:**
- Open http://localhost:8081 in your browser
- All functionality is available through the web interface

**Manual Usage:**
```bash
# Generate keys first
cd examples/go_web_demo
./generate_keys.sh

# Build and run with Docker Compose
docker-compose up --build

# Access at http://localhost:8081
```

**Web UI Features:**
- **Key Generation Tab**: View pre-generated keys and regeneration instructions
- **File Management Tab**: Upload files, create sample files, clear data directory
- **Package Files Tab**: Encrypt files with/without licensing, clear output files
- **License Token Tab**: Issue vendor-signed tokens with custom expiry
- **Unpack Files Tab**: Decrypt files with licensing verification, clear decrypted files
- **Upload & Unpack Tab**: Upload encrypted packages and keys for direct unpacking
- **File Browser Tab**: View and download files from all directories

**API Endpoints:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/files/upload` | Upload files to data directory |
| POST | `/api/files/clear-data` | Clear data directory |
| POST | `/api/files/clear-output` | Clear output directory |
| POST | `/api/files/clear-decrypted` | Clear decrypted directory |
| GET | `/api/files/download/:filename` | Download files |
| POST | `/api/files/upload-unpack` | Upload and unpack encrypted files |
| POST | `/package` | Package files with encryption |
| POST | `/unpack` | Unpack encrypted files |
| POST | `/issue-token` | Issue license token |
| GET | `/api/files/:directory` | List files in directory |

**Use Cases:**
- **Demo and Testing**: Interactive demonstration of all features
- **File Management**: Upload and manage files through web interface
- **Workflow Testing**: Complete end-to-end encryption workflows
- **Key Management**: Generate and manage RSA key pairs
- **License Management**: Create and manage license tokens
- **File Sharing**: Upload encrypted packages for others to unpack

**Security Features:**
- **Pre-generated Keys**: Keys are generated on the host system, not in containers
- **File Isolation**: Separate directories for input, output, and decrypted files
- **Secure Downloads**: Files are served through secure endpoints
- **License Verification**: Full licensing workflow with expiry management
- **File Cleanup**: Clear functions for all directories

**Getting Started with Web UI:**
1. **Run the demo script**: `cd examples/go_web_demo && ./demo.sh`
   - This automatically generates RSA keys and starts the containers
   - **Note**: Don't use `docker-compose up --build` directly - it will fail without keys
2. **Access the UI**: Open http://localhost:8081
3. **Upload files**: Use the File Management tab to upload your files
4. **Package files**: Use the Package Files tab to encrypt with/without licensing
5. **Issue tokens**: Use the License Token tab to create license tokens
6. **Unpack files**: Use the Unpack Files tab to decrypt and verify licenses
7. **Browse files**: Use the File Browser tab to view and download files

**Troubleshooting:**
- **Port conflicts**: The UI runs on port 8081 by default
- **Key generation**: Always use `./demo.sh` instead of `docker-compose up --build` directly
- **Missing keys error**: If you get "keys not found" errors, run `./generate_keys.sh` first
- **File permissions**: Check Docker volume mounts and file permissions
- **Container logs**: Use `docker-compose logs` to debug issues

For detailed documentation, see the README files in each example directory.


