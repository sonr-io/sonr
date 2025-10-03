#!/usr/bin/make -f

# Version and build information
GIT_ROOT := $(shell git rev-parse --show-toplevel)
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null || echo "dev") | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

BUMP_LEVEL := PATCH

# Build configuration
BUILD_FLAGS := -trimpath

# Linker flags for version info
ldflags = -X main.Version=$(VERSION) \
		  -X main.Commit=$(COMMIT) \
		  -s -w

BUILD_FLAGS += -ldflags '$(ldflags)'
HWAY_OUT := $(GIT_ROOT)/build/hway

.PHONY: all build install clean test version help docker

all: build

build:
ifeq ($(OS),Windows_NT)
	$(error hway server not supported on Windows)
else
	@echo "Building Highway service..."
	@go build -mod=readonly $(BUILD_FLAGS) -o $(HWAY_OUT) .
	@echo "Binary built: ../../build/hway"
endif

install:
ifeq ($(OS),Windows_NT)
	$(error hway server not supported on Windows)
else
	@echo "Installing Highway service..."
	@go install -mod=readonly $(BUILD_FLAGS) .
	@echo "Binary installed to $(GOPATH)/bin/hway"
endif

tidy:
	@echo "Tidying Go module..."
	@go mod tidy

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(HWAY_OUT)
	@echo "Clean complete"

test:
	@echo "Running Highway service tests..."
	@go test -v -race ./...

release:
	@echo "Creating hway release..."
	@cd $(GIT_ROOT) && cz --config cmd/hway/.cz.toml --no-raise 6,21 bump --yes --increment PATCH

snapshot:
	@echo "Dry-Run Bumping Highway version..."
	@cd $(GIT_ROOT) && cz --config cmd/hway/.cz.toml bump --yes --no-verify --dry-run --increment PATCH
	@echo "Creating hway snapshots for all platforms..."
	@cd $(GIT_ROOT) && goreleaser release --snapshot --clean -f cmd/hway/.goreleaser.yml

# Docker build targets
docker:
	@echo "Building Highway Docker image..."
	@docker build -f Dockerfile -t onsonr/hway:latest -t ghcr.io/sonr-io/hway:latest ../..
	@echo "Docker image built and tagged:"
	@echo "  - onsonr/hway:latest"
	@echo "  - ghcr.io/sonr-io/hway:latest"

docker-local:
	@echo "Building Highway Docker image with local context..."
	@docker build -f Dockerfile -t onsonr/hway:local -t ghcr.io/sonr-io/hway:local ../..
	@echo "Docker image built and tagged:"
	@echo "  - onsonr/hway:local"
	@echo "  - ghcr.io/sonr-io/hway:local"

version:
	@echo "Highway Service"
	@echo "==============="
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"

help:
	@echo "Highway Service Makefile"
	@echo "========================"
	@echo ""
	@echo "Highway is an Asynq-based task processing service for vault operations,"
	@echo "using Redis-backed job queues and Proto.Actor framework for concurrency."
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build Highway binary (default)"
	@echo "  install   - Install Highway to GOPATH/bin"
	@echo "  docker    - Build Docker image (onsonr/hway:latest)"
	@echo "  clean     - Remove build artifacts"
	@echo "  test      - Run Highway tests"
	@echo "  version   - Display version information"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Requirements:"
	@echo "  - Redis server running on 127.0.0.1:6379"
	@echo "  - IPFS nodes for vault operations"
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build the Highway service"
	@echo "  make test     # Run tests"
	@echo "  make install  # Install to GOPATH/bin"
