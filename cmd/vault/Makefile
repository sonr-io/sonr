#!/usr/bin/make -f

# Output configuration - dual output for both plugin and ES package
GIT_ROOT := $(shell git rev-parse --show-toplevel)
PLUGIN_DIR := $(GIT_ROOT)/x/dwn/client/plugin
ES_PACKAGE_DIR := $(GIT_ROOT)/packages/es/src/plugin
OUTPUT_FILE := plugin.wasm
PLUGIN_PATH := $(PLUGIN_DIR)/vault.wasm
ES_PATH := $(ES_PACKAGE_DIR)/$(OUTPUT_FILE)

# Build configuration for WASM
GOOS := wasip1
GOARCH := wasm

# Version information
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null || echo "dev") | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

.PHONY: all build clean install test version help

all: build

build:
	@echo "Building vault WASM module..."
	@echo "Targets: Plugin and ES package"
	@mkdir -p $(PLUGIN_DIR)
	@mkdir -p $(ES_PACKAGE_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(PLUGIN_PATH) main.go
	@cp $(PLUGIN_PATH) $(ES_PATH)
	@echo "✅ Vault WASM module built successfully"
	@echo "Plugin output: $(PLUGIN_PATH)"
	@echo "ES package output: $(ES_PATH)"
	@ls -lh $(PLUGIN_PATH) | awk '{print "Plugin size: " $$5}'
	@ls -lh $(ES_PATH) | awk '{print "ES package size: " $$5}'

tidy:
	@echo "Tidying vault build artifacts..."
	@go mod tidy
	@echo "✅ Tidy complete"

test:
	@echo "Running vault tests..."
	@go test -v ./...
	@cd $(GIT_ROOT) && go test -C . -mod=readonly -v github.com/sonr-io/sonr/x/dwn/client/plugin/...

clean:
	@echo "Cleaning vault build artifacts..."
	@rm -f $(PLUGIN_PATH)
	@rm -f $(ES_PATH)
	@echo "✅ Clean complete"

release:
	@echo "Creating vault release..."
	@cd $(GIT_ROOT) && cz --config cmd/vault/.cz.toml --no-raise 6,21 bump --yes --increment PATCH

snapshot:
	@echo "Dry-Run Bumping Vault version..."
	@cd $(GIT_ROOT) && cz --config cmd/vault/.cz.toml bump --yes --dry-run --no-verify --increment PATCH
	@echo "Creating vault snapshots for all platforms..."
	@cd $(GIT_ROOT) && goreleaser release --snapshot --clean -f cmd/vault/.goreleaser.yml

version:
	@echo "Vault WASM Module"
	@echo "================="
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Target OS: $(GOOS)"
	@echo "Target Arch: $(GOARCH)"
	@echo "Plugin output: $(PLUGIN_PATH)"
	@echo "ES output: $(ES_PATH)"

verify: build
	@echo "Verifying WASM modules..."
	@if [ -f "$(PLUGIN_PATH)" ]; then \
		file "$(PLUGIN_PATH)"; \
		echo "✅ Plugin WASM file exists"; \
	else \
		echo "❌ Plugin WASM file not found"; \
		exit 1; \
	fi
	@if [ -f "$(ES_PATH)" ]; then \
		file "$(ES_PATH)"; \
		echo "✅ ES package WASM file exists"; \
	else \
		echo "❌ ES package WASM file not found"; \
		exit 1; \
	fi
	@echo "Plugin size: $$(du -h $(PLUGIN_PATH) | cut -f1)"
	@echo "ES size: $$(du -h $(ES_PATH) | cut -f1)"
	@echo "✅ Verification complete"

help:
	@echo "Vault WASM Module Makefile"
	@echo "=========================="
	@echo ""
	@echo "This Makefile builds the vault WebAssembly module for both the"
	@echo "Sonr blockchain plugin system and the @sonr.io/es package."
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build vault WASM module (default)"
	@echo "  clean     - Remove built WASM artifacts"
	@echo "  test      - Run vault tests"
	@echo "  tidy      - Tidy Go module dependencies"
	@echo "  version   - Display version information"
	@echo "  verify    - Build and verify the WASM modules"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Output locations:"
	@echo "  Plugin:     $(PLUGIN_PATH)"
	@echo "  ES Package: $(ES_PATH)"
	@echo ""
	@echo "Integration:"
	@echo "  The WASM module is built for two purposes:"
	@echo "  1. Plugin system in x/dwn/client/plugin"
	@echo "  2. ES package for browser distribution via jsDelivr"
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build WASM module to both locations"
	@echo "  make verify   # Build and verify both outputs"
	@echo "  make clean    # Remove all artifacts"
