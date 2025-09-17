# Multi-arch container for secure_packager (packager, unpack, issue-token)
# Usage examples (buildx):
#   docker buildx build --platform linux/amd64,linux/arm64 -t yourorg/secure-packager:latest --push .
#   docker run --rm -v $(pwd)/input:/in -v $(pwd)/out:/out \
#     yourorg/secure-packager:latest packager -in /in -out /out -pub /out/customer_public.pem -zip=true

FROM golang:1.21-bookworm AS build
WORKDIR /src

# Enable reproducible, static builds
ENV CGO_ENABLED=0

# Build args provided by buildx
ARG TARGETOS
ARG TARGETARCH

# Cache go mod downloads
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy sources
COPY cmd/ ./cmd/

# Build the three commands for target platform
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/packager ./cmd/packager && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/unpack ./cmd/unpack && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /out/issue-token ./cmd/issue-token

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=build /out/packager /app/packager
COPY --from=build /out/unpack /app/unpack
COPY --from=build /out/issue-token /app/issue-token

# Simple dispatcher entrypoint
RUN printf '#!/bin/sh\nset -e\ncmd="$1"; shift || true\ncase "$cmd" in\n  packager) exec /app/packager "$@" ;;\n  unpack) exec /app/unpack "$@" ;;\n  issue-token) exec /app/issue-token "$@" ;;\n  ""|help|--help|-h) echo "Usage: secure-packager {packager|unpack|issue-token} [args...]"; exit 0 ;;\n  *) echo "Unknown command: $cmd"; exit 1 ;;\n esac\n' > /usr/local/bin/secure-packager && chmod +x /usr/local/bin/secure-packager

VOLUME ["/in", "/out", "/work", "/keys"]

ENTRYPOINT ["secure-packager"]
CMD ["--help"]


