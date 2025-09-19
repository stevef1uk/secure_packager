# Secure Packager Integration Example

This example demonstrates how to integrate the `secure_packager` package with a file processing application. It shows a complete workflow of processing files, encrypting them securely, and then decrypting and verifying them.

## Overview

The example consists of three main components:

1. **Checksum Calculator** (`checksum_demo.go`) - A standalone program that scans directories and calculates file checksums
2. **Secure Packager Library** (`secure_packager_lib.go`) - A library that wraps the secure_packager functionality
3. **Integration Example** (`integration_example.go`) - A complete demonstration showing how to integrate both components

## Features Demonstrated

- **File Processing**: Scan directories and calculate various checksums (MD5, SHA1, SHA256, SHA512)
- **Secure Encryption**: Use envelope encryption to protect files with RSA + Fernet
- **Licensing Support**: Optional licensing enforcement with vendor-signed tokens
- **Verification**: Decrypt files and verify their integrity using checksums

## Quick Start

### Prerequisites

- Go 1.21 or later
- The `github.com/fernet/fernet-go` dependency

### Running the Examples

1. **Checksum Calculator Demo**:
   ```bash
   cd examples/example
   go run checksum_demo.go -dir ./data -algo sha256
   ```

2. **Integration Example** (without licensing):
   ```bash
   cd examples/example
   go run integration_example.go -work ./demo_work
   ```

3. **Integration Example** (with licensing):
   ```bash
   cd examples/example
   go run integration_example.go -work ./demo_work -license
   ```

## File Structure

```
examples/example/
├── README.md                    # This file
├── go.mod                       # Go module definition
├── checksum_demo.go            # Standalone checksum calculator
├── secure_packager_lib.go      # Secure packager library
└── integration_example.go      # Complete integration demo
```

## How It Works

### 1. Checksum Calculator

The `ChecksumCalculator` provides methods to:
- Calculate various checksums (MD5, SHA1, SHA256, SHA512)
- Scan directories recursively
- Process files and display results in a formatted table

```go
calculator := NewChecksumCalculator("sha256")
err := calculator.ScanDirectoryAndChecksum("./data")
```

### 2. Secure Packager Library

The `SecurePackager` provides a clean API for:
- **Encryption**: Encrypt directories using envelope encryption
- **Decryption**: Decrypt zip archives and extract files
- **Key Management**: Handle RSA key pairs for encryption/decryption
- **Licensing**: Optional license verification and enforcement

```go
// Create packager
packager, err := NewSecurePackager("customer_public.pem")

// Encrypt directory
err = packager.EncryptDirectory("./data", "./encrypted", false)

// Decrypt zip
err = packager.DecryptZip("./encrypted/encrypted_files.zip", "./decrypted", "customer_private.pem", "")
```

### 3. Integration Example

The integration example demonstrates a complete workflow:

1. **Setup**: Generate RSA key pairs for customer and vendor
2. **Data Creation**: Create sample files for processing
3. **File Processing**: Calculate checksums of original files
4. **Encryption**: Encrypt files using secure_packager
5. **Decryption**: Decrypt files and verify checksums match

## Key Integration Points

### File Processing Pipeline

```go
// 1. Process files and calculate checksums
calculator := NewChecksumCalculator("sha256")
calculator.ScanDirectoryAndChecksum(dataDir)

// 2. Encrypt files securely
packager, _ := NewSecurePackager(customerPubPath)
packager.EncryptDirectory(dataDir, encryptedDir, false)

// 3. Decrypt and verify
packager.DecryptZip(zipPath, decryptedDir, customerPrivPath, "")
calculator.ScanDirectoryAndChecksum(decryptedDir)
```

### Error Handling

The library provides comprehensive error handling:
- File I/O errors
- Cryptographic errors
- Key management errors
- License verification errors

### Security Features

- **Envelope Encryption**: Files encrypted with Fernet, key wrapped with RSA
- **No Plaintext Keys**: Fernet key never stored in plaintext
- **Key Isolation**: Only the private key holder can decrypt
- **Optional Licensing**: Vendor-signed tokens for access control

## Customization

### Adding New File Processors

You can easily extend the example by adding new file processors:

```go
type FileProcessor interface {
    ProcessFile(filePath string) error
    GetResults() map[string]string
}

// Implement your processor
type MyProcessor struct {
    // your fields
}

func (p *MyProcessor) ProcessFile(filePath string) error {
    // your processing logic
    return nil
}
```

### Supporting New Checksum Algorithms

Add new algorithms to the `ChecksumCalculator`:

```go
func (cc *ChecksumCalculator) createHash() (hash.Hash, error) {
    switch cc.algorithm {
    case "blake2b":
        return blake2b.New512(nil)
    // ... existing cases
    }
}
```

### Custom License Verification

Implement custom license verification logic:

```go
func (sp *SecurePackager) verifyAndEnforceLicense(vendorPubPath, tokenPath string) error {
    // Your custom license verification logic
    return nil
}
```

## Use Cases

This integration pattern is useful for:

- **Data Distribution**: Securely distribute files to customers
- **Model Deployment**: Protect ML models and datasets
- **Document Management**: Encrypt sensitive documents
- **Backup Systems**: Secure backup and restore workflows
- **Compliance**: Meet data protection requirements

## Security Considerations

- Use strong RSA keys (2048+ bits)
- Store private keys securely
- Implement proper key rotation
- Monitor license usage
- Use secure communication channels

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure `go mod tidy` is run to download dependencies
2. **Key Errors**: Verify RSA key format and permissions
3. **File Permissions**: Check read/write permissions for directories
4. **License Errors**: Ensure token format matches expected structure

### Debug Mode

Enable verbose logging by setting environment variables:
```bash
export DEBUG=1
go run integration_example.go
```

## Contributing

To extend this example:

1. Fork the repository
2. Create a new branch for your feature
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## License

This example is provided under the same license as the main secure_packager project.
