FROM golang:1.24-alpine AS go-builder

SHELL ["/bin/sh", "-ecuxo", "pipefail"]

RUN apk add --no-cache ca-certificates build-base git

WORKDIR /code

COPY go.mod go.sum ./
RUN set -eux; \
  export ARCH=$(uname -m); \
  WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm || true); \
  if [ ! -z "${WASM_VERSION}" ]; then \
  WASMVM_REPO=$(echo $WASM_VERSION | awk '{print $1}');\
  WASMVM_VERS=$(echo $WASM_VERSION | awk '{print $2}');\
  wget -O /lib/libwasmvm_muslc.a https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm_muslc.$(uname -m).a;\
  fi; \
  go mod download;

# Copy over code
COPY . /code

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/sonrd
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/bin/snrd \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/bin/snrd | grep "statically linked")

# --------------------------------------------------------
FROM alpine:3.17

LABEL org.opencontainers.image.title="snrd"
LABEL org.opencontainers.image.authors="diDAO <hi@didao.xyz>"
LABEL org.opencontainers.image.source=https://github.com/sonr-io/snrd

COPY --from=go-builder /code/build/sonrd /usr/bin/sonrd

# Set up dependencies
ENV PACKAGES="curl make bash jq sed"

# Install minimum necessary dependencies
RUN apk add --no-cache $PACKAGES

WORKDIR /opt
