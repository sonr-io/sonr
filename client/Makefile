#!/usr/bin/make -f

DOCKER := $(shell which docker)
PROJECT_NAME = sonr-client-sdk

# golangci-lint Docker image version
GOLANGCI_VERSION := v2.3.1

###############################################################################
###                                  Build                                 ###
###############################################################################

build:
	@gum log --level info "Building Go client SDK..."
	@go build ./...

###############################################################################
###                                Formatting                              ###
###############################################################################

format fmt:
	@gum log --level info "Formatting Go code with golangci-lint Docker..."
	@$(DOCKER) run --rm -t -v $$(pwd):/app -w /app \
		--user $$(id -u):$$(id -g) \
		-v $$(go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
		-v $$(go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
		-v ~/.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run --fix --issues-exit-code=0

###############################################################################
###                                 Linting                                ###
###############################################################################

lint:
	@gum log --level info "Running golangci-lint Docker..."
	@$(DOCKER) run --rm -t -v $$(pwd):/app -w /app \
		--user $$(id -u):$$(id -g) \
		-v $$(go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
		-v $$(go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
		-v ~/.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run --timeout=10m

###############################################################################
###                                 Testing                                ###
###############################################################################

test:
	@gum log --level info "Running all tests..."
	@go test -mod=readonly -race -coverprofile=coverage.txt -covermode=atomic ./...

test-unit:
	@gum log --level info "Running unit tests..."
	@go test -mod=readonly -short -race ./...

test-integration:
	@gum log --level info "Running integration tests..."
	@go test -mod=readonly -race -tags=integration ./tests/integration/...

test-verbose:
	@gum log --level info "Running tests with verbose output..."
	@go test -mod=readonly -v -race ./...

test-cover:
	@gum log --level info "Running tests with coverage..."
	@go test -mod=readonly -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@gum log --level info "Coverage report generated: coverage.html"

benchmark:
	@gum log --level info "Running benchmarks..."
	@go test -mod=readonly -bench=. -benchmem ./...

###############################################################################
###                              Dependencies                              ###
###############################################################################

deps:
	@gum log --level info "Installing dependencies..."
	@go mod download

tidy:
	@gum log --level info "Tidying dependencies..."
	@go mod tidy

verify:
	@gum log --level info "Verifying dependencies..."
	@go mod verify

###############################################################################
###                                 Clean                                  ###
###############################################################################

clean:
	@gum log --level info "Cleaning build artifacts..."
	@rm -f coverage.txt coverage.html
	@go clean -cache

###############################################################################
###                                  Help                                  ###
###############################################################################

help:
	@gum log --level info "Available targets:"
	@gum log --level info "  build          - Build the Go client SDK"
	@gum log --level info "  format/fmt     - Format Go code using golangci-lint"
	@gum log --level info "  lint           - Run golangci-lint"
	@gum log --level info "  test           - Run all tests with coverage"
	@gum log --level info "  test-unit      - Run unit tests only"
	@gum log --level info "  test-integration - Run integration tests"
	@gum log --level info "  test-verbose   - Run tests with verbose output"
	@gum log --level info "  test-cover     - Run tests and generate HTML coverage report"
	@gum log --level info "  benchmark      - Run benchmarks"
	@gum log --level info "  deps           - Download dependencies"
	@gum log --level info "  tidy           - Tidy and verify dependencies"
	@gum log --level info "  verify         - Verify dependencies"
	@gum log --level info "  clean          - Clean build artifacts"
	@gum log --level info "  help           - Show this help message"

.DEFAULT_GOAL := help
.PHONY: build format fmt lint test test-unit test-integration test-verbose test-cover benchmark deps tidy verify clean help
