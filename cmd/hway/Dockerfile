# Use build argument for base image to allow using pre-cached base
ARG BASE_IMAGE=golang:1.24.7-alpine3.22

# --------------------------------------------------------
# Highway service build stage
FROM ${BASE_IMAGE} AS builder
SHELL ["/bin/sh", "-ecuxo", "pipefail"]

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    build-base \
    git \
    linux-headers \
    bash

WORKDIR /code

# Copy entire source code
COPY . .

# Fix git ownership issue
RUN git config --global --add safe.directory /code

# Download Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Build Highway binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev"); \
    COMMIT=$(git log -1 --format='%H' 2>/dev/null || echo "unknown"); \
    CGO_ENABLED=1 GOOS=linux \
    go build \
        -mod=readonly \
        -ldflags "-X main.Version=${VERSION} \
                  -X main.Commit=${COMMIT} \
                  -w -s" \
        -buildvcs=false \
        -trimpath \
        -o /code/build/hway \
        ./cmd/hway

# --------------------------------------------------------
# Highway runtime image
FROM alpine:3.17

LABEL org.opencontainers.image.title="Sonr Highway Service"
LABEL org.opencontainers.image.source="https://github.com/sonr-io/sonr"

# Copy binary from builder
COPY --from=builder /code/build/hway /usr/bin/hway

# Install runtime dependencies
RUN apk add --no-cache ca-certificates wget

# Create non-root user
RUN adduser -D -u 1000 highway

# Set working directory
WORKDIR /home/highway

# Switch to non-root user
USER highway

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8090/health || exit 1

# Expose Highway port
EXPOSE 8090

# Set default command
ENTRYPOINT ["/usr/bin/hway"]
