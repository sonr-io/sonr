# Use build argument for base image to allow using pre-cached base
ARG BASE_IMAGE=golang:1.24.7-alpine3.22

FROM ${BASE_IMAGE} AS builder
SHELL ["/bin/sh", "-ecuxo", "pipefail"]

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    build-base \
    git \
    linux-headers \
    bash \
    binutils-gold \
    wget

WORKDIR /code

# Copy entire source code (including crypto module)
COPY . .

# Fix git ownership issue
RUN git config --global --add safe.directory /code

# Download WasmVM library
RUN --mount=type=cache,target=/tmp/wasmvm \
    set -eux; \
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | head -1 || echo ""); \
    if [ ! -z "${WASM_VERSION}" ]; then \
        WASMVM_REPO=$(echo $WASM_VERSION | awk '{print $1}'); \
        WASMVM_VERS=$(echo $WASM_VERSION | awk '{print $2}'); \
        WASMVM_FILE="libwasmvm_muslc.${ARCH}.a"; \
        if [ ! -f "/tmp/wasmvm/${WASMVM_FILE}" ]; then \
            wget -O "/tmp/wasmvm/${WASMVM_FILE}" "https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/${WASMVM_FILE}"; \
        fi; \
        cp "/tmp/wasmvm/${WASMVM_FILE}" /lib/libwasmvm_muslc.a; \
    fi

# Download Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Build binary with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev"); \
    COMMIT=$(git log -1 --format='%H' 2>/dev/null || echo "unknown"); \
    LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true \
    CGO_ENABLED=1 GOOS=linux \
    go build \
        -mod=readonly \
        -tags "netgo,ledger,muslc" \
        -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name=sonr \
                  -X github.com/cosmos/cosmos-sdk/version.AppName=snrd \
                  -X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
                  -X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} \
                  -X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger,muslc \
                  -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
        -buildvcs=false \
        -trimpath \
        -o /code/build/snrd \
        ./cmd/snrd; \
    file /code/build/snrd; \
    echo "Ensuring binary is statically linked ..."; \
    (file /code/build/snrd | grep "statically linked")

# --------------------------------------------------------
# Final minimal runtime image
FROM alpine:3.17

LABEL org.opencontainers.image.title="Sonr Daemon"
LABEL org.opencontainers.image.source="https://github.com/sonr-io/sonr"

# Copy binary from build stage
COPY --from=builder /code/build/snrd /usr/bin
COPY --from=builder /lib/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
COPY --from=builder /code/scripts/test_node.sh /usr/bin/testnet
RUN chmod +x /usr/bin/testnet

# Set up dependencies
ENV PACKAGES="curl make bash jq sed"

# Install minimum necessary dependencies
RUN apk add --no-cache $PACKAGES

WORKDIR /opt
