# Go Web Demo for Secure Packager

This is a Go-based web interface alternative to Gradio for demonstrating the secure_packager functionality. It provides a modern, responsive web UI built with Go, Gin framework, and Bootstrap.

## Features

- **Modern Web UI**: Clean, responsive interface built with Bootstrap 5
- **Docker Integration**: All operations use Docker containers (no local OpenSSL required)
- **Interactive Workflow**: Step-by-step demonstration of secure packager functionality
- **File Browser**: View and inspect files at each stage
- **Complete Workflow**: Run the entire process with one click
- **Real-time Feedback**: Live status updates and error handling
- **Pre-generated Keys**: RSA key pairs generated automatically by demo script
- **File Upload/Download**: Upload files for encryption and download encrypted/decrypted files
- **License Management**: Full licensing workflow with token issuance and verification

## Architecture

- **Backend**: Go with Gin web framework
- **Frontend**: HTML5, CSS3, JavaScript (vanilla JS, no frameworks)
- **Styling**: Bootstrap 5 with custom CSS
- **Docker**: Multi-container setup for isolation and portability

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### Running with Demo Script (Recommended)

The easiest way to start the demo is using the provided `demo.sh` script, which handles all setup automatically:

1. **Navigate to the demo directory:**
   ```bash
   cd examples/go_web_demo
   ```

2. **Run the demo script:**
   ```bash
   ./demo.sh
   ```

   This script will:
   - Check Docker and Docker Compose availability
   - Clean up previous containers
   - Build the keygen container
   - Generate RSA key pairs (customer and vendor)
   - Build the web demo container
   - Start all services
   - Display access information

3. **Access the web interface:**
   Open your browser to `http://localhost:8080`

4. **Stop the services:**
   ```bash
   docker-compose down
   ```

### Manual Setup (Alternative)

If you prefer to run the setup manually:

1. **Generate keys first:**
   ```bash
   cd examples/go_web_demo
   ./generate_keys.sh
   ```

2. **Start the services:**
   ```bash
   docker-compose up --build
   ```

3. **Access the web interface:**
   Open your browser to `http://localhost:8080`

### Running Locally (Development)

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Build the keygen container:**
   ```bash
   docker build -t secure-packager-keygen:latest -f ../keygen/Dockerfile.keygen ../keygen/
   ```

3. **Generate keys:**
   ```bash
   ./generate_keys.sh
   ```

4. **Run the application:**
   ```bash
   go run main.go
   ```

5. **Access the web interface:**
   Open your browser to `http://localhost:8080`

## Demo Script (`demo.sh`)

The `demo.sh` script provides a complete automated setup for the Go Web Demo. It handles all the necessary steps to get the demo running:

### What the Script Does

1. **Environment Checks**: Verifies Docker and Docker Compose are available
2. **Cleanup**: Removes any existing containers to ensure a clean start
3. **Directory Setup**: Creates necessary directories (`data`, `output`, `keys`, `logs`)
4. **Key Generation**: Builds the keygen container and generates RSA key pairs
5. **Container Building**: Builds the web demo container
6. **Service Startup**: Starts all services with Docker Compose
7. **Health Check**: Tests the health endpoint to verify everything is working
8. **User Guidance**: Provides clear instructions and next steps

### Key Generation Process

The script automatically generates two RSA key pairs using OpenSSL in a Docker container:

- **Customer Keys**: `customer_private.pem` and `customer_public.pem` (2048 bits)
- **Vendor Keys**: `vendor_private.pem` and `vendor_public.pem` (2048 bits)

**Important Security Notes:**
- Keys are generated on the host system, not inside containers
- Private keys are stored securely in the `keys/` directory with proper permissions
- The customer private key is required for decryption - keep it secure
- The customer public key can be shared openly for encryption
- Vendor keys are used for license token signing and verification

**Key Management:**
- Keys are generated with 2048-bit RSA (recommended minimum)
- Proper file permissions are set automatically
- Keys persist between container restarts
- To regenerate keys, delete the `keys/` directory and run `./demo.sh` again

### Script Output

When you run `./demo.sh`, you'll see:
- Progress indicators for each step
- Service status information
- Access URL and port information
- Available API endpoints
- Local directory structure
- Commands for stopping, viewing logs, and rebuilding

### Troubleshooting the Script

If the script fails:
1. Ensure Docker is running: `docker ps`
2. Check Docker Compose: `docker-compose --version`
3. Verify port 8080 is available
4. Check the logs: `docker-compose logs -f web-demo`

## Usage

### Web Interface Tabs

1. **🔑 Key Generation**
   - View pre-generated RSA key pairs (customer and vendor)
   - Keys are generated by `demo.sh` script before starting the UI
   - Shows key information and regeneration instructions
   - **Note**: The UI displays existing keys and provides instructions for regeneration
   - Uses OpenSSL in Docker container for key generation

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
   Access: `http://localhost:8080`

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
   Access: `http://localhost:8080`

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

The web interface provides a comprehensive REST API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/keys/generate` | Generate RSA keys (shows instructions) |
| POST | `/api/files/create` | Create sample files |
| POST | `/api/files/upload` | Upload files to data directory |
| POST | `/api/package` | Package files with encryption |
| POST | `/api/token/issue` | Issue license token |
| POST | `/api/unpack` | Unpack files with decryption |
| POST | `/api/files/upload-unpack` | Upload and unpack encrypted files |
| GET | `/api/files/:directory` | List files in directory |
| POST | `/api/files/read` | Read file content |
| GET | `/api/files/download/:filename` | Download files |
| POST | `/api/files/clear-data` | Clear data directory |
| POST | `/api/files/clear-output` | Clear output directory |
| POST | `/api/files/clear-decrypted` | Clear decrypted directory |
| POST | `/api/workflow/complete` | Run complete workflow |

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

1. **Docker not running**: 
   - Ensure Docker daemon is running: `docker ps`
   - Check Docker Compose: `docker-compose --version`

2. **Permission denied**: 
   - Check Docker socket permissions
   - On Linux: `sudo usermod -aG docker $USER` and logout/login
   - On macOS: Ensure Docker Desktop is running

3. **Port conflicts**: 
   - Default port is 8080, change in `docker-compose.yml` if needed
   - Check if port is in use: `lsof -i :8080`

4. **Container build failures**: 
   - Check Dockerfile syntax and dependencies
   - Ensure Docker has enough resources allocated
   - Try rebuilding: `docker-compose build --no-cache`

5. **Keys not found error**: 
   - Always use `./demo.sh` instead of `docker-compose up --build` directly
   - Run `./generate_keys.sh` first if needed
   - Check that `keys/` directory exists and contains key files

6. **File permission issues**: 
   - On x86 machines, keys may be owned by root
   - Fix with: `sudo chown -R $(id -u):$(id -g) keys/`
   - The `generate_keys.sh` script handles this automatically

### Debug Mode

Enable debug logging by setting environment variable:
```bash
export GIN_MODE=debug
```

### Logs

View container logs:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f web-demo

# Key generation logs
docker-compose logs -f keygen
```

### Health Checks

Test the API endpoints:
```bash
# Health check
curl http://localhost:8080/health

# Test complete workflow
curl -X POST http://localhost:8080/api/workflow/complete

# List files
curl http://localhost:8080/api/files/data
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
