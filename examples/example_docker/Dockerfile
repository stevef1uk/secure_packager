# Multi-stage Dockerfile for secure_packager integration example
# This demonstrates how to integrate secure_packager with a Go application
# using Docker entrypoints for decryption

# -----------------------------
# Build stage
# -----------------------------
FROM golang:1.21-bookworm AS build
WORKDIR /src

# -----------------------------
# Build secure_packager tools
# -----------------------------
RUN git clone --depth 1 https://github.com/stevef1uk/secure_packager /src/secure_packager
WORKDIR /src/secure_packager
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/packager ./cmd/packager
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/unpack ./cmd/unpack
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/issue-token ./cmd/issue-token

# -----------------------------
# Build Go application
# -----------------------------
WORKDIR /src/app
COPY app/go.mod ./
RUN go mod download
COPY app/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app .

# -----------------------------
# Build Go entrypoint
# -----------------------------
WORKDIR /src/entrypoint
COPY entrypoint/go.mod ./
RUN go mod download
COPY entrypoint/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /out/entrypoint main.go

# Strip binaries for smaller image (skip if strip fails)
RUN strip /out/packager /out/unpack /out/issue-token /out/app /out/entrypoint || true

# -----------------------------
# Directory creation stage
# -----------------------------
FROM debian:bookworm-slim AS dirs
RUN mkdir -p /app/keys /app/data /app/decrypted /app/logs /app/tmp
RUN adduser --disabled-password --uid 65532 appuser
RUN chown -R 65532:65532 /app

# -----------------------------
# Final minimal runtime image
# -----------------------------
FROM debian:bookworm-slim
WORKDIR /app

# Install minimal runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Remove common shell utilities for security
RUN rm -f /bin/ls /bin/cat /bin/sh /bin/bash /usr/bin/less /usr/bin/vim /usr/bin/git /usr/bin/find

# Copy directories
COPY --from=dirs --chown=65532:65532 /app /app

# Copy Go binaries
COPY --from=build /out/app /app/app
COPY --from=build /out/entrypoint /app/entrypoint
COPY --from=build /out/packager /app/packager
COPY --from=build /out/unpack /app/unpack
COPY --from=build /out/issue-token /app/issue-token

# Copy sample data and keys (these would be mounted in production)
COPY --chown=65532:65532 data/ /app/data/
COPY --chown=65532:65532 keys/ /app/keys/

# Run as non-root user
USER 65532:65532

# Environment variables
ENV APP_PORT=8080
ENV DECRYPT_OUTPUT_DIR=/app/decrypted
ENV PRIVATE_KEY_PATH=/app/keys/customer_private.pem
ENV TOKEN_FILE_PATH=/app/keys/token.txt
ENV ENCRYPTED_ZIP_PATH=/app/data/encrypted_files.zip

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/entrypoint --health-check || exit 1

# Entrypoint
ENTRYPOINT ["/app/entrypoint"]
