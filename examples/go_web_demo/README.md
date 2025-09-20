# Go Web Demo for Secure Packager

This is a Go-based web interface alternative to Gradio for demonstrating the secure_packager functionality. It provides a modern, responsive web UI built with Go, Gin framework, and Bootstrap.

## Features

- **Modern Web UI**: Clean, responsive interface built with Bootstrap 5
- **Docker Integration**: All operations use Docker containers (no local OpenSSL required)
- **Interactive Workflow**: Step-by-step demonstration of secure packager functionality
- **File Browser**: View and inspect files at each stage
- **Complete Workflow**: Run the entire process with one click
- **Real-time Feedback**: Live status updates and error handling

## Architecture

- **Backend**: Go with Gin web framework
- **Frontend**: HTML5, CSS3, JavaScript (vanilla JS, no frameworks)
- **Styling**: Bootstrap 5 with custom CSS
- **Docker**: Multi-container setup for isolation and portability

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Running with Docker Compose

1. **Clone and navigate to the demo directory:**
   ```bash
   cd examples/go_web_demo
   ```

2. **Start the services:**
   ```bash
   docker-compose up --build
   ```

3. **Access the web interface:**
   Open your browser to `http://localhost:8081`

4. **Stop the services:**
   ```bash
   docker-compose down
   ```

### Running Locally (Development)

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Build the keygen container:**
   ```bash
   docker build -t secure-packager-keygen:latest -f ../gradio_demo/Dockerfile.keygen ../gradio_demo/
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

4. **Access the web interface:**
   Open your browser to `http://localhost:8080`

## Usage

### Web Interface Tabs

1. **🔑 Key Generation**
   - Generate RSA key pairs (customer and vendor)
   - Adjustable key size (1024-4096 bits)
   - Uses OpenSSL in Docker container

2. **📄 Create Files**
   - Create sample files for encryption
   - Customizable content
   - Generates both text and JSON files

3. **📦 Package Files**
   - Encrypt files with customer's public key
   - Optional licensing support
   - Creates encrypted ZIP archives

4. **🎫 License Token**
   - Issue vendor-signed license tokens
   - Configurable expiry and metadata
   - Required for licensed packages

5. **📤 Unpack Files**
   - Decrypt and verify files
   - Optional license verification
   - Extracts to decrypted directory

6. **📁 File Browser**
   - View files in different directories
   - Read file contents
   - Real-time file listing

7. **🚀 Complete Workflow**
   - Run entire process automatically
   - Step-by-step progress display
   - Comprehensive error handling

## Two-Person Secure File Sharing Workflow

This section demonstrates how Person A (sender) and Person B (recipient) can securely create and share encrypted files using the web interface, following proper security practices for key creation and sharing.

### Overview

The secure file sharing process involves:
1. **Person B** generates their own RSA key pair (public + private)
2. **Person B** shares their public key with Person A through a secure channel
3. **Person A** uses Person B's public key to encrypt files
4. **Person A** sends the encrypted files to Person B
5. **Person B** uses their private key to decrypt the files

### Step-by-Step Process

#### Phase 1: Person B (Recipient) - Key Generation

1. **Person B starts the web interface:**
   ```bash
   cd examples/go_web_demo
   docker-compose up --build
   ```
   Access: `http://localhost:8081`

2. **Generate RSA key pair:**
   - Navigate to the **🔑 Key Generation** tab
   - Select appropriate key size (recommended: 2048 or 4096 bits)
   - Click "Generate Keys"
   - **Important**: Save the private key securely - it will be needed for decryption

3. **Export the public key:**
   - In the **📁 File Browser** tab, navigate to the `keys` directory
   - Open `customer_public_key.pem` and copy its contents
   - This is the public key that will be shared with Person A

4. **Secure key storage:**
   - **Private key**: Store `customer_private_key.pem` in a secure location (password manager, encrypted storage)
   - **Public key**: Can be shared openly (this is safe to share)

#### Phase 2: Person A (Sender) - File Encryption

1. **Person A starts their own instance:**
   ```bash
   cd examples/go_web_demo
   docker-compose up --build
   ```
   Access: `http://localhost:8081`

2. **Prepare Person B's public key:**
   - In the **📁 File Browser** tab, navigate to the `keys` directory
   - Replace the contents of `customer_public_key.pem` with Person B's public key
   - **Important**: Only replace the public key, keep the private key as-is

3. **Create files to encrypt:**
   - Navigate to the **📄 Create Files** tab
   - Create the files you want to send to Person B
   - Customize content as needed

4. **Encrypt the files:**
   - Navigate to the **📦 Package Files** tab
   - Click "Package Files" to encrypt them with Person B's public key
   - The encrypted ZIP file will be created in the `output` directory

5. **Share the encrypted files:**
   - In the **📁 File Browser** tab, navigate to the `output` directory
   - Download `encrypted_files.zip`
   - Send this file to Person B through your preferred secure channel

#### Phase 3: Person B (Recipient) - File Decryption

1. **Person B receives the encrypted file:**
   - Place the `encrypted_files.zip` file in their `output` directory
   - Or use the web interface to upload it

2. **Decrypt the files:**
   - Navigate to the **📤 Unpack Files** tab
   - Click "Unpack Files" to decrypt using their private key
   - Decrypted files will appear in the `output/decrypted` directory

3. **Verify the files:**
   - Use the **📁 File Browser** tab to view the decrypted files
   - Ensure all files are intact and readable

### Security Best Practices

#### Key Management
- **Never share private keys**: Only the public key should be shared
- **Use strong key sizes**: Minimum 2048 bits, preferably 4096 bits
- **Secure key storage**: Store private keys in encrypted password managers
- **Key rotation**: Consider generating new key pairs periodically

#### Communication Security
- **Verify public keys**: Confirm the public key belongs to the intended recipient
- **Use secure channels**: Share public keys through verified communication channels
- **Verify file integrity**: Use checksums or digital signatures when possible

#### File Handling
- **Secure transmission**: Use encrypted channels (HTTPS, encrypted email) for file sharing
- **Temporary storage**: Delete encrypted files after successful decryption
- **Access control**: Ensure only authorized parties have access to decrypted files

### Advanced Scenarios

#### With License Tokens
If Person A wants to add licensing restrictions:

1. **Person A generates vendor keys:**
   - Use the **🔑 Key Generation** tab to create vendor key pair
   - Keep vendor private key secure

2. **Create license token:**
   - Navigate to **🎫 License Token** tab
   - Issue a token with appropriate restrictions
   - Include this token with the encrypted files

3. **Person B verifies license:**
   - During decryption, the system will verify the license token
   - Files will only decrypt if the license is valid

#### Multiple Recipients
To send the same files to multiple people:

1. **Each recipient generates their own key pair**
2. **Person A encrypts files separately for each recipient**
3. **Share the appropriate encrypted file with each recipient**

### Troubleshooting Common Issues

#### "Invalid key format" error
- Ensure the public key is in PEM format
- Check that the key file contains the complete key (including headers)

#### "Decryption failed" error
- Verify you're using the correct private key
- Ensure the encrypted file wasn't corrupted during transmission
- Check that the file was encrypted with the matching public key

#### "License verification failed" error
- Verify the license token is valid and not expired
- Ensure the vendor public key is correct
- Check that the token was issued by the correct vendor

### Security Verification Checklist

Before sharing sensitive files, verify:

- [ ] Public key was received through a secure, verified channel
- [ ] Key size is appropriate (2048+ bits)
- [ ] Files are encrypted successfully
- [ ] Encrypted files are transmitted securely
- [ ] Recipient can successfully decrypt files
- [ ] Private keys are stored securely
- [ ] Temporary files are cleaned up

This workflow ensures that only the intended recipient can decrypt the files, even if the encrypted files or public keys are intercepted during transmission.

### API Endpoints

The web interface provides a REST API:

- `GET /health` - Health check
- `POST /api/keys/generate` - Generate RSA keys
- `POST /api/files/create` - Create sample files
- `POST /api/package` - Package files
- `POST /api/token/issue` - Issue license token
- `POST /api/unpack` - Unpack files
- `GET /api/files/:directory` - List files
- `POST /api/files/read` - Read file content
- `POST /api/workflow/complete` - Run complete workflow

## Docker Containers

### Key Generation Container
- **Image**: `secure-packager-keygen:latest`
- **Purpose**: OpenSSL key generation
- **Source**: `../gradio_demo/Dockerfile.keygen`

### Web Demo Container
- **Image**: Built from local Dockerfile
- **Purpose**: Go web application
- **Port**: 8080

### Secure Packager Container
- **Image**: `stevef1uk/secure-packager:latest`
- **Purpose**: Main secure packager tools
- **Usage**: Called by web application

## File Structure

```
go_web_demo/
├── main.go                 # Main Go application
├── go.mod                  # Go module file
├── Dockerfile              # Container definition
├── docker-compose.yml      # Multi-container setup
├── templates/
│   └── index.html         # HTML template
├── static/
│   ├── style.css          # Custom CSS
│   └── script.js          # JavaScript functionality
└── README.md              # This file
```

## Development

### Adding New Features

1. **Backend**: Add new API endpoints in `main.go`
2. **Frontend**: Update `templates/index.html` and `static/script.js`
3. **Styling**: Modify `static/style.css`

### Building for Production

```bash
# Build the web demo container
docker build -t secure-packager-web-demo:latest .

# Build the keygen container
docker build -t secure-packager-keygen:latest -f ../gradio_demo/Dockerfile.keygen ../gradio_demo/
```

### Testing

The application includes comprehensive error handling and user feedback. Test different scenarios:

- Invalid key sizes
- Missing files
- Docker container failures
- Network issues
- File permission problems

## Security Considerations

- **Container Isolation**: All operations run in isolated Docker containers
- **Non-root User**: Application runs as non-root user
- **Input Validation**: All user inputs are validated
- **Error Handling**: Sensitive information is not exposed in error messages

## Troubleshooting

### Common Issues

1. **Docker not running**: Ensure Docker daemon is running
2. **Permission denied**: Check Docker socket permissions
3. **Port conflicts**: Change port in docker-compose.yml
4. **Container build failures**: Check Dockerfile syntax and dependencies

### Debug Mode

Enable debug logging by setting environment variable:
```bash
export GIN_MODE=debug
```

### Logs

View container logs:
```bash
docker-compose logs -f web-demo
```

## Comparison with Gradio

| Feature | Go Web Demo | Gradio |
|---------|-------------|--------|
| **Language** | Go | Python |
| **Framework** | Gin | Gradio |
| **Frontend** | HTML/CSS/JS | Auto-generated |
| **Customization** | Full control | Limited |
| **Performance** | High | Medium |
| **Dependencies** | Minimal | Heavy |
| **Docker Size** | Small | Large |
| **Integration** | Native Go | External |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

Same as the main secure_packager project.
