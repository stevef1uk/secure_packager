# Secure Packager Examples

This directory contains comprehensive examples demonstrating the secure_packager functionality in different scenarios and with different technologies.

## Available Examples

### 1. Go Web Demo (`go_web_demo/`)
**Modern Go-based web interface alternative to Gradio**

- **Technology**: Go, Gin framework, Bootstrap 5, vanilla JavaScript
- **Features**: Interactive web UI, Docker integration, real-time feedback
- **Best for**: Production deployments, Go developers, lightweight solutions
- **Size**: ~50MB Docker image
- **Performance**: High (native Go)

**Quick Start:**
```bash
cd examples/go_web_demo
./demo.sh
# Open http://localhost:8080
```

### 2. Key Generation Utilities (`keygen/`)
**Docker-based key generation utilities**

- **Technology**: Docker, OpenSSL
- **Features**: RSA key pair generation, containerized OpenSSL
- **Best for**: Key generation without local OpenSSL installation
- **Size**: ~20MB Docker image
- **Performance**: High (Alpine Linux)

**Quick Start:**
```bash
cd examples/ui
docker build -t secure-packager-keygen:latest -f Dockerfile.keygen .
docker run --rm -v $(pwd)/keys:/output secure-packager-keygen:latest both
```

### 3. Library Integration (`example/`)
**Go library integration example**

- **Technology**: Go, direct library usage
- **Features**: Programmatic access, checksum calculation, file processing
- **Best for**: Go applications, custom integrations, development
- **Size**: N/A (library)
- **Performance**: Highest (direct integration)

**Quick Start:**
```bash
cd examples/example
./demo.sh
```

### 4. Docker Integration (`example_docker/`)
**Containerized application with entrypoint decryption**

- **Technology**: Go, Docker, HTTP API
- **Features**: Production-ready, containerized deployment, API endpoints
- **Best for**: Microservices, Kubernetes, production deployments
- **Size**: ~100MB Docker image
- **Performance**: High (containerized Go)

**Quick Start:**
```bash
cd examples/example_docker
./demo.sh
```

## Technology Comparison

| Feature | Go Web Demo | Key Gen Utils | Library | Docker Integration |
|---------|-------------|---------------|---------|-------------------|
| **Language** | Go | Docker/OpenSSL | Go | Go |
| **UI Framework** | Custom HTML/JS | CLI | CLI | HTTP API |
| **Docker Required** | Yes | Yes | No | Yes |
| **Local Dependencies** | None | None | Go | None |
| **Customization** | Full | Limited | Full | Full |
| **Performance** | High | High | Highest | High |
| **Learning Curve** | Medium | Low | High | Medium |
| **Production Ready** | Yes | Yes | Yes | Yes |
| **File Size** | Small | Very Small | N/A | Medium |

## Use Case Recommendations

### Choose Go Web Demo when:
- You want a modern, responsive web interface
- You prefer Go over Python
- You need full control over the UI/UX
- You want lightweight, fast performance
- You're building production applications

### Choose Key Generation Utilities when:
- You need RSA key generation without local OpenSSL
- You want lightweight, containerized key generation
- You're building custom workflows
- You need to generate keys in CI/CD pipelines
- You want minimal dependencies

### Choose Library Integration when:
- You're building Go applications
- You need programmatic control
- You want the highest performance
- You're integrating with existing Go codebases
- You don't need a web interface

### Choose Docker Integration when:
- You're deploying to production
- You need containerized applications
- You're using microservices architecture
- You need HTTP API endpoints
- You're working with Kubernetes

## Quick Start Guide

### For Web Interfaces (Go or Gradio):

1. **Choose your preferred technology**
2. **Navigate to the example directory**
3. **Run the demo script**
4. **Open the web interface in your browser**
5. **Follow the interactive workflow**

### For Library Integration:

1. **Navigate to the example directory**
2. **Run the demo script**
3. **Examine the source code**
4. **Integrate into your application**

### For Docker Integration:

1. **Navigate to the example directory**
2. **Run the demo script**
3. **Test the API endpoints**
4. **Deploy to your container platform**

## Architecture Overview

All examples demonstrate the same core secure_packager workflow:

```
1. Key Generation (RSA key pairs)
   ↓
2. File Creation (sample data)
   ↓
3. File Packaging (encryption)
   ↓
4. License Token Issuance (optional)
   ↓
5. File Unpacking (decryption)
   ↓
6. Verification (file integrity)
```

## Security Features

All examples include:

- **Envelope Encryption**: RSA + Fernet encryption
- **No Plaintext Keys**: Fernet key never stored in plaintext
- **Key Isolation**: Only private key holder can decrypt
- **Optional Licensing**: Vendor-signed tokens for access control
- **File Integrity**: Checksum verification
- **Container Isolation**: All operations in isolated containers

## Development

### Adding New Examples

1. Create a new directory under `examples/`
2. Include a `README.md` with clear instructions
3. Provide a demo script for easy testing
4. Document the technology stack and use cases
5. Test thoroughly with different scenarios

### Contributing

1. Fork the repository
2. Create your example in a new directory
3. Follow the existing patterns and structure
4. Test with the provided demo scripts
5. Submit a pull request

## Troubleshooting

### Common Issues

1. **Docker not running**: Ensure Docker daemon is running
2. **Port conflicts**: Change ports in configuration files
3. **Permission issues**: Check Docker socket permissions
4. **Build failures**: Verify Dockerfile syntax and dependencies

### Getting Help

- Check the individual example README files
- Review the main project documentation
- Open an issue on GitHub
- Check the demo script outputs for error messages

## License

All examples are provided under the same license as the main secure_packager project.
