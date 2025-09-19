# Quick Start Guide - Docker Integration

This guide shows how to quickly get started with the Docker-based secure_packager integration example.

## What This Example Demonstrates

- **Docker Entrypoint Pattern**: Decryption happens before the main application starts
- **Secure File Processing**: Files are decrypted and then processed by a Go application
- **Production-Ready**: Non-root user, minimal image, health checks
- **Multiple Scenarios**: Encrypted data, demo mode, development mode

## Prerequisites

- Docker and Docker Compose
- OpenSSL (for key generation)
- `jq` (optional, for JSON formatting)

## Quick Start (5 minutes)

### 1. Run the Demo
```bash
cd examples/example_docker
./demo.sh
```

This single command will:
- Generate RSA keys
- Create sample data
- Encrypt the data
- Build Docker images
- Start containers
- Test the API

### 2. Test the API
```bash
# Health check
curl http://localhost:8080/health

# Process files
curl -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"directory": "/app/decrypted", "algorithm": "sha256"}'
```

### 3. View Results
- **Encrypted processor**: http://localhost:8080
- **Demo processor**: http://localhost:8081
- **Logs**: `docker-compose logs -f`

## Manual Setup

If you prefer to set up manually:

### 1. Generate Keys
```bash
mkdir -p keys
openssl genrsa -out keys/customer_private.pem 2048
openssl rsa -in keys/customer_private.pem -pubout -out keys/customer_public.pem
```

### 2. Create Sample Data
```bash
mkdir -p data
echo "Sample data" > data/sample.txt
echo '{"config": "value"}' > data/config.json
```

### 3. Encrypt Data
```bash
docker run --rm -v $(pwd)/data:/in -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /in -pub /keys/customer_public.pem -zip=true
```

### 4. Build and Run
```bash
docker-compose up --build
```

## Understanding the Flow

1. **Container Starts**: Entrypoint program runs first
2. **Decryption**: Uses `secure_packager` tools to decrypt files
3. **Application Start**: Main Go application starts
4. **API Ready**: HTTP API becomes available
5. **File Processing**: Application processes decrypted files

## Key Files

- `Dockerfile`: Multi-stage build with secure_packager tools
- `entrypoint/main.go`: Decryption and application startup
- `app/main.go`: File processing HTTP API
- `docker-compose.yml`: Orchestration setup
- `demo.sh`: Automated demo script

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/process` | Process files in directory |
| GET | `/api/files` | List files in directory |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | 8080 | Application port |
| `DECRYPT_OUTPUT_DIR` | /app/decrypted | Decrypted files directory |
| `PRIVATE_KEY_PATH` | /app/keys/customer_private.pem | Private key path |
| `TOKEN_FILE_PATH` | /app/keys/token.txt | License token path |
| `ENCRYPTED_ZIP_PATH` | /app/data/encrypted_files.zip | Encrypted zip path |

## Docker Compose Services

- **file-processor-encrypted**: Port 8080, uses encrypted data
- **file-processor-demo**: Port 8081, uses demo data
- **file-processor-dev**: Port 8082, debug mode

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker-compose logs file-processor-encrypted

# Check if keys exist
ls -la keys/
```

### Decryption Fails
```bash
# Verify key format
openssl rsa -in keys/customer_private.pem -text -noout

# Check encrypted data
ls -la data/
```

### API Not Responding
```bash
# Check health
curl http://localhost:8080/health

# Check container status
docker-compose ps
```

## Next Steps

1. **Customize the Application**: Modify `app/main.go` for your use case
2. **Add New File Types**: Extend the file processing logic
3. **Integrate with Kubernetes**: Use the provided Kubernetes examples
4. **Add Monitoring**: Implement metrics and logging
5. **Security Hardening**: Review and enhance security measures

## Production Deployment

For production use:

1. **Use Secrets Management**: Store keys in Kubernetes secrets or similar
2. **Implement Monitoring**: Add Prometheus metrics and Grafana dashboards
3. **Set Up Logging**: Use structured logging with ELK stack
4. **Security Scanning**: Regular vulnerability scans
5. **Backup Strategy**: Secure backup of encrypted data

## Support

- Check the main README.md for detailed documentation
- Review the troubleshooting section
- Open an issue on GitHub for bugs or questions
