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
