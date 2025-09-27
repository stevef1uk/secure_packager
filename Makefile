# Secure Packager Makefile

.PHONY: help redis-start redis-stop redis-clear redis-restart build clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  redis-start    - Start Redis container"
	@echo "  redis-stop     - Stop Redis container"
	@echo "  redis-clear    - Clear Redis cache"
	@echo "  redis-restart  - Restart Redis and clear cache"
	@echo "  build          - Build all Go binaries"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"

# Redis management
redis-start:
	@echo "Starting Redis container..."
	@docker run -d --name redis -p 6379:6379 -v redis-data:/data redis:latest
	@echo "✅ Redis started on port 6379"

redis-stop:
	@echo "Stopping Redis container..."
	@docker stop redis || true
	@docker rm redis || true
	@echo "✅ Redis stopped and removed"

redis-clear:
	@echo "Clearing Redis cache..."
	@docker exec redis redis-cli FLUSHALL || echo "Redis not running or no data to clear"
	@echo "✅ Redis cache cleared"

redis-restart: redis-stop redis-start
	@echo "✅ Redis restarted"

# Build commands
build:
	@echo "Building Go binaries..."
	@go build -o issue-token ./cmd/issue-token
	@go build -o packager ./cmd/packager
	@go build -o unpack ./cmd/unpack
	@echo "✅ All binaries built"

# Clean commands
clean:
	@echo "Cleaning build artifacts..."
	@rm -f issue-token packager unpack
	@echo "✅ Build artifacts cleaned"

# Test commands
test:
	@echo "Running tests..."
	@go test ./...
	@echo "✅ Tests completed"

