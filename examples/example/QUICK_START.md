# Quick Start Guide

This directory contains integration examples for the `secure_packager` package, demonstrating how to integrate secure file encryption with file processing applications.

## What's Included

### 1. Checksum Calculator (`checksum/`)
A standalone program that demonstrates file processing by calculating checksums for all files in a directory.

**Features:**
- Multiple hash algorithms (MD5, SHA1, SHA256, SHA512)
- Recursive directory scanning
- Formatted output with file names and checksums

**Usage:**
```bash
cd checksum
go run main.go -dir /path/to/files -algo sha256
```

### 2. Integration Example (`integration/`)
A complete demonstration showing how to integrate `secure_packager` with a file processing application.

**Features:**
- File processing with checksum calculation
- Secure encryption using envelope encryption (RSA + Fernet)
- Optional licensing support
- Complete workflow: encrypt → decrypt → verify

**Usage:**
```bash
cd integration
go run main.go -work ./demo_work
go run main.go -work ./demo_work -license  # With licensing
```

### 3. Demo Script (`demo.sh`)
A comprehensive demo script that runs both examples and shows the complete workflow.

**Usage:**
```bash
./demo.sh
```

## Key Integration Points

### File Processing Pipeline
1. **Process Files**: Calculate checksums of original files
2. **Encrypt Files**: Use `secure_packager` to encrypt files securely
3. **Decrypt Files**: Decrypt files and verify checksums match

### Security Features
- **Envelope Encryption**: Files encrypted with Fernet, key wrapped with RSA
- **No Plaintext Keys**: Fernet key never stored in plaintext
- **Key Isolation**: Only the private key holder can decrypt
- **Optional Licensing**: Vendor-signed tokens for access control

### Code Structure
```go
// 1. Create packager
packager, err := NewSecurePackager("customer_public.pem")

// 2. Encrypt directory
err = packager.EncryptDirectory("./data", "./encrypted", false)

// 3. Decrypt zip
err = packager.DecryptZip("./encrypted/encrypted_files.zip", "./decrypted", "customer_private.pem", "")
```

## Prerequisites

- Go 1.21 or later
- The `github.com/fernet/fernet-go` dependency (automatically downloaded)

## Running the Examples

### Option 1: Run the Complete Demo
```bash
./demo.sh
```

### Option 2: Run Individual Examples

**Checksum Calculator:**
```bash
cd checksum
go run main.go -dir ./sample_files -algo sha256
```

**Integration Example (No License):**
```bash
cd integration
go run main.go -work ./demo_work
```

**Integration Example (With License):**
```bash
cd integration
go run main.go -work ./demo_work -license
```

## Expected Output

The demo will show:
1. **File Processing**: Lists of files with their checksums
2. **Encryption**: Progress of file encryption
3. **Decryption**: Progress of file decryption
4. **Verification**: Checksums of decrypted files (should match originals)

## File Structure After Running

```
demo_work/
├── sample_files/           # Sample files for checksum demo
├── no_license/            # Integration demo without licensing
│   ├── data/              # Original files
│   ├── encrypted/         # Encrypted zip file
│   ├── decrypted/         # Decrypted files
│   └── keys/              # RSA key pairs
└── with_license/          # Integration demo with licensing
    ├── data/              # Original files
    ├── encrypted/         # Encrypted zip file with license manifest
    ├── decrypted/         # Decrypted files
    └── keys/              # RSA key pairs and license token
```

## Customization

### Adding New File Processors
Extend the `ChecksumCalculator` or create new processors that implement similar interfaces.

### Supporting New Algorithms
Add new hash algorithms to the `createHash()` method in the calculator.

### Custom License Verification
Implement custom license verification logic in the `verifyAndEnforceLicense()` method.

## Troubleshooting

- **Import Errors**: Run `go mod tidy` in each subdirectory
- **Permission Errors**: Ensure write permissions for the working directory
- **Key Errors**: The examples generate keys automatically, but ensure proper file permissions

## Next Steps

1. Study the code structure in both examples
2. Modify the file processing logic for your use case
3. Integrate the `SecurePackager` library into your application
4. Implement custom license verification if needed
5. Add error handling and logging as appropriate

For more detailed information, see the main [README.md](README.md) file.
