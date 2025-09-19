# Multi-stage Dockerfile for Go web demo
FROM golang:1.24-alpine AS build

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    docker-cli

# Copy secure-packager tools from the official image
COPY --from=stevef1uk/secure-packager:latest /app/packager /usr/local/bin/packager
COPY --from=stevef1uk/secure-packager:latest /app/unpack /usr/local/bin/unpack
COPY --from=stevef1uk/secure-packager:latest /app/issue-token /usr/local/bin/issue-token

# Make tools executable
RUN chmod +x /usr/local/bin/packager /usr/local/bin/unpack /usr/local/bin/issue-token

# Create app user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary
COPY --from=build /app/main .

# Copy static files and templates
COPY --from=build /app/static ./static
COPY --from=build /app/templates ./templates

# Create necessary directories
RUN mkdir -p data output keys logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
