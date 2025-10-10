#!/usr/bin/make -f

# Version and build information
GIT_ROOT := $(shell git rev-parse --show-toplevel)
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null || echo "dev") | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true

# Build tags configuration
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support)
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))

# Linker flags configuration
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=sonr \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=snrd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -checklinkname=0 \
		  -s -w

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath

# Build vault WASM module (required dependency)
VAULT_ROOT := $(GIT_ROOT)/cmd/vault
VAULT_OUT := $(GIT_ROOT)/x/dwn/client/plugin/vault.wasm
SNRD_OUT := $(GIT_ROOT)/build/snrd
SNRD_EXE_OUT := $(GIT_ROOT)/build/snrd.exe

.PHONY: all build install clean vault version help docker

all: build

build:
ifeq ($(OS),Windows_NT)
	GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(SNRD_EXE_OUT) .
else
	@echo "Building snrd binary..."
	@CGO_LDFLAGS="-lm" go build -mod=readonly $(BUILD_FLAGS) -o $(SNRD_OUT) .
	@echo "Binary built: ../../build/snrd"
endif

install:
	@$(MAKE) -C $(VAULT_ROOT) build
ifeq ($(OS),Windows_NT)
	@echo "Building snrd.exe for Windows..."
	GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(SNRD_EXE_OUT) .
else
	@echo "Installing snrd binary..."
	@CGO_LDFLAGS="-lm" go install -mod=readonly $(BUILD_FLAGS) .
	@echo "Binary installed to $(GOPATH)/bin/snrd"
endif

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(SNRD_OUT) $(SNRD_EXE_OUT)
	@$(MAKE) -C $(VAULT_ROOT) clean
	@echo "Clean complete"

version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build tags: $(build_tags_comma_sep)"

test:
	@echo "Running snrd tests..."
	@go test -v ./...

release:
	@echo "Creating snrd release..."
	@cd $(GIT_ROOT) && goreleaser release --clean

snapshot:
	@echo "Creating snrd snapshots for all platforms..."
	@cd $(GIT_ROOT) && goreleaser release --snapshot --clean

# Docker build targets
docker:
	@echo "Building snrd Docker image..."
	@docker build -f Dockerfile -t onsonr/snrd:latest -t ghcr.io/sonr-io/snrd:latest ../..
	@echo "Docker image built and tagged:"
	@echo "  - onsonr/snrd:latest"
	@echo "  - ghcr.io/sonr-io/snrd:latest"

docker-local:
	@echo "Building snrd Docker image with local context..."
	@docker build -f Dockerfile -t onsonr/snrd:local -t ghcr.io/sonr-io/snrd:local ../..
	@echo "Docker image built and tagged:"
	@echo "  - onsonr/snrd:local"
	@echo "  - ghcr.io/sonr-io/snrd:local"

link-wasmvm:
	@echo "Downloading WasmVM libraries..."
	@mkdir -p build
	@WASMVM_VERSION=$$(go list -m all | grep github.com/CosmWasm/wasmvm | head -1 | cut -d' ' -f2); \
	echo "WasmVM version: $$WASMVM_VERSION"; \
	echo "Downloading Linux AMD64 static library..."; \
	curl -L -o build/libwasmvm_muslc_amd64.a https://github.com/CosmWasm/wasmvm/releases/download/$$WASMVM_VERSION/libwasmvm_muslc.x86_64.a; \
	echo "Downloading Linux ARM64 static library..."; \
	curl -L -o build/libwasmvm_muslc_arm64.a https://github.com/CosmWasm/wasmvm/releases/download/$$WASMVM_VERSION/libwasmvm_muslc.aarch64.a; \
	echo "Downloading macOS AMD64 dynamic library..."; \
	curl -L -o build/libwasmvm.dylib.amd64 https://github.com/CosmWasm/wasmvm/releases/download/$$WASMVM_VERSION/libwasmvm.dylib; \
	echo "Downloading macOS ARM64 dynamic library..."; \
	curl -L -o build/libwasmvm.dylib.arm64 https://github.com/CosmWasm/wasmvm/releases/download/$$WASMVM_VERSION/libwasmvm.dylib; \
	echo "WasmVM libraries downloaded successfully"

help:
	@echo "snrd Build Makefile"
	@echo "==================="
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build snrd binary (with vault WASM)"
	@echo "  install   - Install snrd to GOPATH/bin"
	@echo "  vault     - Build vault WASM module only"
	@echo "  docker    - Build Docker image (onsonr/snrd:latest)"
	@echo "  clean     - Remove build artifacts"
	@echo "  version   - Display version information"
	@echo "  link-wasmvm - Download WasmVM libraries for goreleaser"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Build options:"
	@echo "  LEDGER_ENABLED=true|false  - Enable/disable ledger support (default: true)"
	@echo "  WITH_CLEVELDB=yes         - Build with CLevelDB support"
	@echo "  LINK_STATICALLY=true      - Build static binary"
	@echo "  BUILD_TAGS='tag1 tag2'    - Additional build tags"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make build LEDGER_ENABLED=false"
	@echo "  make install LINK_STATICALLY=true"
