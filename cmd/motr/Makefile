#!/usr/bin/make -f

# Output configuration - outputs to ES package for bundling
GIT_ROOT := $(shell git rev-parse --show-toplevel)
ES_PACKAGE_DIR := $(GIT_ROOT)/packages/es/src/worker
WASM_FILE := app.wasm
JS_FILE := wasm_exec.js
OUTPUT_PATH := $(ES_PACKAGE_DIR)/$(WASM_FILE)
JS_PATH := $(ES_PACKAGE_DIR)/$(JS_FILE)

# Build configuration for WASM
GOOS := js
GOARCH := wasm
CGO_ENABLED := 0

# Version information
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null || echo "dev") | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

# Go installation paths
GOROOT := $(shell go env GOROOT)
WASM_EXEC_SOURCE := $(GOROOT)/misc/wasm/wasm_exec.js

# Build flags
LDFLAGS := -s -w
BUILD_FLAGS := -ldflags="$(LDFLAGS)" -trimpath

.PHONY: all build clean test verify help version runtime tidy

all: build

build: clean-output runtime
	@echo "Building Motor WASM module for ES package..."
	@echo "Target: $(OUTPUT_PATH)"
	@mkdir -p $(ES_PACKAGE_DIR)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(OUTPUT_PATH) .
	@echo "✅ Motor WASM module built successfully"
	@echo "Output: $(OUTPUT_PATH)"
	@ls -lh $(OUTPUT_PATH) | awk '{print "Size: " $$5}'
	@$(MAKE) runtime

runtime:
	@echo "Copying WASM runtime..."
	@if [ -f "$(WASM_EXEC_SOURCE)" ]; then \
		cp "$(WASM_EXEC_SOURCE)" "$(JS_PATH)"; \
		echo "✅ WASM runtime copied to $(JS_PATH)"; \
	else \
		echo "⚠️  WASM runtime not found at $(WASM_EXEC_SOURCE)"; \
		echo "You may need to manually copy wasm_exec.js"; \
	fi

clean-output:
	@echo "Cleaning previous builds..."
	@rm -f $(OUTPUT_PATH)
	@rm -f $(JS_PATH)
	@mkdir -p $(ES_PACKAGE_DIR)

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(OUTPUT_PATH)
	@rm -f $(JS_PATH)
	@echo "✅ Clean complete"

release:
	@echo "Creating motr release..."
	@cd $(GIT_ROOT) && cz --config cmd/motr/.cz.toml --no-raise 6,21 bump --yes --increment PATCH

snapshot:
	@echo "Dry-Run Bumping Motor version..."
	@cd $(GIT_ROOT) && cz --config cmd/motr/.cz.toml bump --yes --no-verify --dry-run --increment PATCH
	@echo "Creating motr snapshots for all platforms..."
	@cd $(GIT_ROOT) && goreleaser release --snapshot --clean -f cmd/motr/.goreleaser.yml

tidy:
	@echo "Tidying Motor module..."
	@go mod tidy
	@echo "✅ Tidy complete"
	 
test:
	@echo "Running Motor tests..."
	@go test -v ./...

verify: build
	@echo "Verifying WASM module..."
	@if [ -f "$(OUTPUT_PATH)" ]; then \
		file "$(OUTPUT_PATH)"; \
		echo "✅ WASM file exists"; \
	else \
		echo "❌ WASM file not found"; \
		exit 1; \
	fi
	@if [ -f "$(JS_PATH)" ]; then \
		echo "✅ Runtime file exists"; \
	else \
		echo "⚠️  Runtime file not found"; \
	fi
	@if command -v wasm-validate >/dev/null 2>&1; then \
		if wasm-validate "$(OUTPUT_PATH)"; then \
			echo "✅ WASM module is valid"; \
		else \
			echo "❌ WASM module validation failed"; \
			exit 1; \
		fi \
	else \
		echo "⚠️  wasm-validate not available, skipping validation"; \
	fi
	@echo "✅ Verification complete"
	@echo ""
	@echo "The WASM module has been built in the ES package at:"
	@echo "  $(OUTPUT_PATH)"
	@echo ""
	@echo "The ES package will bundle and distribute this via jsDelivr"

version:
	@echo "Motor WASM Service Worker"
	@echo "========================="
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Target OS: $(GOOS)"
	@echo "Target Arch: $(GOARCH)"
	@echo "Output: $(OUTPUT_PATH)"

help:
	@echo "Motor WASM Module Makefile"
	@echo "=========================="
	@echo ""
	@echo "Motor provides WebAssembly-based DWN and Wallet operations"
	@echo "for the @sonr.io/es package to distribute via jsDelivr."
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build Motor WASM module (default)"
	@echo "  clean     - Remove all build artifacts"
	@echo "  test      - Run Motor tests"
	@echo "  tidy      - Tidy Go module dependencies"
	@echo "  verify    - Build and validate WASM module"
	@echo "  version   - Display version information"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Build components (called by build):"
	@echo "  runtime   - Copy WASM runtime (wasm_exec.js)"
	@echo ""
	@echo "Output location:"
	@echo "  ES Package: $(ES_PACKAGE_DIR)/"
	@echo "  WASM File:  $(OUTPUT_PATH)"
	@echo ""
	@echo "Integration:"
	@echo "  The WASM module is built directly into the ES package plugins"
	@echo "  directory for bundling and CDN distribution. The TypeScript"
	@echo "  client in @sonr.io/es/plugins/motor handles service worker management."
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build WASM module into ES package"
	@echo "  make verify   # Build and validate the module"
	@echo "  make clean    # Remove artifacts"
