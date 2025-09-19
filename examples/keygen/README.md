# Key Generation Utilities

This directory contains Docker-based utilities for generating RSA key pairs without requiring OpenSSL to be installed locally.

## What's Included

- **Dockerfile.keygen**: Alpine Linux container with OpenSSL
- **docker-compose.yml**: Multi-container orchestration
- **Key generation script**: Automated RSA key pair generation

## Quick Start

### Generate Keys with Docker

```bash
# Build the keygen container
docker build -t secure-packager-keygen:latest -f Dockerfile.keygen .

# Generate customer keys (2048 bits)
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest customer 2048

# Generate vendor keys (2048 bits)
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest vendor 2048

# Generate both key pairs at once
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest both 2048
```

### Using Docker Compose

```bash
# Start keygen service
docker-compose --profile keygen up keygen

# Keys will be generated in ./keys/ directory
```

## Key Generation Options

### Command Line Usage

```bash
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest [command] [key_size]
```

**Commands:**
- `customer` - Generate customer key pair
- `vendor` - Generate vendor key pair  
- `both` - Generate both key pairs
- `help` - Show usage information

**Key Sizes:**
- `1024` - Minimum (not recommended for production)
- `2048` - Recommended default
- `3072` - High security
- `4096` - Maximum security

### Examples

```bash
# Generate 2048-bit customer keys
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest customer 2048

# Generate 4096-bit vendor keys
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest vendor 4096

# Generate both with 3072 bits
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest both 3072
```

## Generated Files

The container generates the following files in the output directory:

- `customer_private.pem` - Customer private key
- `customer_public.pem` - Customer public key
- `vendor_private.pem` - Vendor private key
- `vendor_public.pem` - Vendor public key

## Security Features

- **Container Isolation**: Keys generated in isolated container
- **No Local Dependencies**: No OpenSSL installation required
- **Secure Key Generation**: Uses OpenSSL's secure random number generator
- **Proper File Permissions**: Keys created with appropriate permissions

## Integration with Other Examples

This key generation utility is used by:

- **Go Web Demo**: Automatically generates keys for the web interface
- **Docker Integration**: Provides keys for containerized applications
- **Library Integration**: Can be used to generate keys for development

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure Docker has access to the output directory
2. **Container Not Found**: Build the container first with `docker build`
3. **Empty Output**: Check that the volume mount is correct

### Debug Mode

```bash
# Run container interactively for debugging
docker run -it --rm -v $(pwd)/keys:/output secure-packager-keygen:latest sh

# Check generated keys
ls -la /output/
```

## File Structure

```
keygen/
├── Dockerfile.keygen      # Key generation container
└── README.md             # This file
```

## License

Same as the main secure_packager project.
