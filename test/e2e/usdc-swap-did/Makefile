.PHONY: test test-verbose test-quick test-single clean help

# Default target
all: test

# Run all E2E tests
test:
	@echo "Running USDC Swap E2E tests..."
	go test -v -timeout 10m

# Run tests with verbose output and detailed logs
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -timeout 10m 2>&1 | tee test.log
	@echo "Test log saved to test.log"

# Run quick tests (skip slow integration tests)
test-quick:
	@echo "Running quick tests..."
	go test -v -short -timeout 5m

# Run a single test
# Usage: make test-single TEST=Test04_ExecuteUSDCSwap_SNRToUSDC
test-single:
	@echo "Running single test: $(TEST)"
	go test -v -run "TestUSDCSwapE2E/$(TEST)" -timeout 5m

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -timeout 10m
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test categories
test-config:
	@echo "Running configuration tests..."
	go test -v -run "Test01" -timeout 2m

test-did:
	@echo "Running DID tests..."
	go test -v -run "Test02" -timeout 3m

test-dex:
	@echo "Running DEX account tests..."
	go test -v -run "Test03" -timeout 3m

test-swaps:
	@echo "Running swap execution tests..."
	go test -v -run "Test04|Test05|Test09" -timeout 10m

test-validation:
	@echo "Running validation tests..."
	go test -v -run "Test06" -timeout 3m

test-query:
	@echo "Running query tests..."
	go test -v -run "Test08" -timeout 2m

# Verify prerequisites
check-prereqs:
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@echo "✓ Go installed: $$(go version)"
	@command -v snrd >/dev/null 2>&1 || { echo "snrd is not installed"; exit 1; }
	@echo "✓ snrd installed"
	@curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info >/dev/null 2>&1 || \
		{ echo "✗ Sonr testnet not running at localhost:1317"; exit 1; }
	@echo "✓ Sonr testnet running"
	@echo "All prerequisites met!"

# Setup test environment
setup:
	@echo "Setting up test environment..."
	@echo "1. Checking network status..."
	@curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info | grep -q network && \
		echo "✓ Network is running" || { echo "✗ Network not accessible"; exit 1; }
	@echo "2. Checking faucet..."
	@curl -s http://localhost:8000/health >/dev/null 2>&1 && \
		echo "✓ Faucet is running" || echo "⚠ Faucet not accessible (tests may fail)"
	@echo "Setup complete!"

# Verify IBC connections
check-ibc:
	@echo "Checking IBC connections..."
	@snrd query ibc connection connections --node tcp://localhost:26657 --output json | \
		jq -r '.connections[] | "\(.id): \(.state)"'
	@echo ""
	@echo "Checking IBC channels..."
	@snrd query ibc channel channels --node tcp://localhost:26657 --output json | \
		jq -r '.channels[] | "\(.port_id)/\(.channel_id): \(.state)"'

# Query DEX module status
check-dex:
	@echo "Checking DEX module..."
	@snrd query dex params --node tcp://localhost:26657 --output json | \
		jq -r '.params | "Enabled: \(.enabled), Max Accounts: \(.max_accounts_per_did)"'

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f test.log coverage.out coverage.html
	@echo "✓ Cleaned"

# Watch mode - run tests on file changes
# Requires: entr (brew install entr or apt-get install entr)
watch:
	@echo "Watching for file changes..."
	@ls *.go | entr -c make test-verbose

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -timeout 10m

# Generate test data
generate-test-data:
	@echo "Generating test DIDs and accounts..."
	@for i in {1..5}; do \
		did="did:snr:test_preset_$$i"; \
		echo "$$did"; \
	done

# Help target
help:
	@echo "Available targets:"
	@echo "  make test              - Run all E2E tests"
	@echo "  make test-verbose      - Run with verbose output"
	@echo "  make test-quick        - Run quick tests only"
	@echo "  make test-single TEST=<name> - Run a single test"
	@echo "  make test-coverage     - Run with coverage report"
	@echo "  make test-config       - Run configuration tests"
	@echo "  make test-did          - Run DID tests"
	@echo "  make test-dex          - Run DEX account tests"
	@echo "  make test-swaps        - Run swap execution tests"
	@echo "  make test-validation   - Run validation tests"
	@echo "  make test-query        - Run query tests"
	@echo "  make check-prereqs     - Verify prerequisites"
	@echo "  make setup             - Setup test environment"
	@echo "  make check-ibc         - Check IBC connections"
	@echo "  make check-dex         - Check DEX module status"
	@echo "  make clean             - Clean test artifacts"
	@echo "  make watch             - Watch mode (requires entr)"
	@echo "  make bench             - Run benchmarks"
	@echo "  make help              - Show this help"
