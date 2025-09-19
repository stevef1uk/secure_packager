# Secure Packager Docker Integration Example

This example demonstrates how to integrate `secure_packager` with a Go application using Docker containers and entrypoints. The approach follows the pattern used in production systems where decryption happens before the main application starts.

## Overview

This example shows how to:
- Use Docker containers to perform decryption before starting the main application
- Integrate `secure_packager` tools directly into the container build process
- Handle both encrypted and non-encrypted scenarios
- Provide a complete file processing API with secure data handling

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Container                         │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   Entrypoint    │    │        Main Application        │ │
│  │                 │    │                                 │ │
│  │ 1. Decrypt      │───▶│ 2. Process Files               │ │
│  │    Files        │    │    - Calculate Checksums       │ │
│  │ 2. Start App    │    │    - HTTP API                  │ │
│  │                 │    │    - Health Checks             │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Dockerfile
Multi-stage build that:
- Builds `secure_packager` tools from source
- Compiles the Go application
- Creates a minimal runtime image
- Sets up proper security (non-root user, minimal tools)

### 2. Entrypoint (`entrypoint/main.go`)
Go program that:
- Performs decryption using `secure_packager` tools
- Handles both licensed and non-licensed scenarios
- Starts the main application
- Provides health checks
- Manages the container lifecycle

### 3. Main Application (`app/main.go`)
File processing application that:
- Provides HTTP API for file operations
- Calculates checksums for files
- Lists and processes files in directories
- Works with decrypted data

### 4. Docker Compose
Orchestration setup with:
- Encrypted data scenario
- Demo mode (no encryption)
- Development mode with debugging

## Quick Start

### Prerequisites

- Docker and Docker Compose
- OpenSSL (for key generation)
- `jq` (optional, for JSON formatting)

### Running the Demo

```bash
# Run the complete demo
./demo.sh
```

This will:
1. Generate RSA key pairs
2. Create sample data
3. Encrypt the data using `secure_packager`
4. Build Docker images
5. Start containers
6. Test the API endpoints

### Manual Setup

```bash
# 1. Generate keys
openssl genrsa -out keys/customer_private.pem 2048
openssl rsa -in keys/customer_private.pem -pubout -out keys/customer_public.pem

# 2. Create sample data
mkdir -p data
echo "Sample data" > data/sample.txt

# 3. Encrypt data
docker run --rm -v $(pwd)/data:/in -v $(pwd)/keys:/keys \
  stevef1uk/secure-packager:latest \
  packager -in /in -out /in -pub /keys/customer_public.pem -zip=true

# 4. Build and run
docker-compose up --build
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Process Files
```bash
curl -X POST http://localhost:8080/api/process \
  -H "Content-Type: application/json" \
  -d '{"directory": "/app/decrypted", "algorithm": "sha256"}'
```

### List Files
```bash
curl "http://localhost:8080/api/files?directory=/app/decrypted"
```

## Configuration

### Environment Variables

- `APP_PORT`: Port for the application (default: 8080)
- `DECRYPT_OUTPUT_DIR`: Directory for decrypted files (default: /app/decrypted)
- `PRIVATE_KEY_PATH`: Path to private key (default: /app/keys/customer_private.pem)
- `TOKEN_FILE_PATH`: Path to license token (default: /app/keys/token.txt)
- `ENCRYPTED_ZIP_PATH`: Path to encrypted zip (default: /app/data/encrypted_files.zip)
- `DEBUG`: Enable debug mode (set to 1)

### Volume Mounts

- `/app/data`: Encrypted data (read-only)
- `/app/keys`: RSA keys and tokens (read-only)
- `/app/decrypted`: Decrypted files (read-write)
- `/app/logs`: Application logs (read-write)

## Docker Compose Services

### file-processor-encrypted
- Port: 8080
- Uses encrypted data
- Full security features

### file-processor-demo
- Port: 8081
- Uses unencrypted demo data
- Shows application without encryption

### file-processor-dev
- Port: 8082
- Debug mode enabled
- Additional logging

## Security Features

### Container Security
- Non-root user (UID 65532)
- Minimal base image
- Removed shell utilities
- Read-only volumes for sensitive data

### Encryption Security
- Envelope encryption (RSA + Fernet)
- No plaintext keys in container
- License verification
- Secure key handling

### Runtime Security
- Health checks
- Graceful error handling
- Secure file permissions
- Process isolation

## Production Considerations

### Key Management
- Store private keys securely (e.g., Kubernetes secrets)
- Use key rotation strategies
- Implement proper access controls

### Monitoring
- Health check endpoints
- Log aggregation
- Metrics collection
- Alerting on failures

### Scaling
- Horizontal scaling with load balancers
- Stateless application design
- Shared storage for decrypted data
- Container orchestration (Kubernetes)

### Security
- Network policies
- Pod security policies
- Regular security updates
- Vulnerability scanning

## Troubleshooting

### Common Issues

1. **Decryption Fails**
   - Check key file permissions
   - Verify key format (PEM)
   - Ensure token is valid

2. **Container Won't Start**
   - Check volume mounts
   - Verify file paths
   - Review container logs

3. **API Not Responding**
   - Check port mappings
   - Verify health checks
   - Review application logs

### Debug Mode

Enable debug mode for detailed logging:
```bash
docker run -e DEBUG=1 your-image
```

### Logs

View container logs:
```bash
docker-compose logs -f file-processor-encrypted
```

## Customization

### Adding New File Processors

Extend the `FileProcessor` struct in `app/main.go`:

```go
type CustomProcessor struct {
    *FileProcessor
    // Add custom fields
}

func (cp *CustomProcessor) ProcessFile(filePath string) (*FileInfo, error) {
    // Custom processing logic
    return cp.FileProcessor.ProcessFile(filePath)
}
```

### Supporting New Encryption Methods

Modify the entrypoint to support additional encryption methods:

```go
func runDecryption(config *Config) error {
    // Add support for different encryption methods
    switch config.EncryptionMethod {
    case "secure_packager":
        return runSecurePackagerDecryption(config)
    case "custom":
        return runCustomDecryption(config)
    }
}
```

### Adding New API Endpoints

Extend the web server in `app/main.go`:

```go
func (ws *WebServer) handleCustomEndpoint(w http.ResponseWriter, r *http.Request) {
    // Custom endpoint logic
}

// In Start() method:
http.HandleFunc("/api/custom", ws.handleCustomEndpoint)
```

## Integration Patterns

### Kubernetes Integration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: file-processor
spec:
  template:
    spec:
      containers:
      - name: file-processor
        image: your-registry/file-processor:latest
        env:
        - name: PRIVATE_KEY_PATH
          valueFrom:
            secretKeyRef:
              name: secure-keys
              key: private-key
        volumeMounts:
        - name: encrypted-data
          mountPath: /app/data
          readOnly: true
      volumes:
      - name: encrypted-data
        persistentVolumeClaim:
          claimName: encrypted-data-pvc
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Build and push
  run: |
    docker build -t ${{ secrets.REGISTRY }}/file-processor:${{ github.sha }} .
    docker push ${{ secrets.REGISTRY }}/file-processor:${{ github.sha }}
```

## License

This example is provided under the same license as the main secure_packager project.

## Contributing

To extend this example:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For questions or issues:
- Check the troubleshooting section
- Review the logs
- Open an issue on GitHub
- Contact the maintainers
