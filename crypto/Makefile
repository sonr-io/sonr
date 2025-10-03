# Crypto Module Makefile
# Provides targets for testing all cryptographic packages individually

.PHONY: all test test-verbose clean help
.PHONY: test-accumulator test-bulletproof test-core test-daed test-dkg test-ecies
.PHONY: test-internal test-keys test-mpc test-ot test-paillier test-sharing
.PHONY: test-signatures test-subtle test-tecdsa test-ted25519 test-zkp
.PHONY: test-core-curves test-core-native test-dkg-frost test-dkg-gennaro test-dkg-gennaro2p
.PHONY: test-ot-base test-ot-extension test-sharing-v1 test-signatures-bbs test-signatures-bls
.PHONY: test-signatures-common test-signatures-schnorr test-tecdsa-dklsv1 test-ted25519-frost
.PHONY: test-ted25519-ted25519 test-zkp-schnorr
.PHONY: bench bench-verbose coverage coverage-html

# Default target
all: test

tidy:
	@go mod tidy
	@go mod download

# Test all packages
test: tidy
	@gum log --level info "Running tests for all crypto packages..."
	@go test ./... -v

# Test all packages with verbose output
test-verbose: tidy
	@gum log --level info "Running verbose tests for all crypto packages..."
	@go test ./... -v -count=1

# Help target
help:
	@gum log --level info "Crypto Module Makefile"
	@gum log --level info ""
	@gum log --level info "Available targets:"
	@gum log --level info "  all                    - Run all tests (default)"
	@gum log --level info "  test                   - Run all tests"
	@gum log --level info "  test-verbose           - Run all tests with verbose output"
	@gum log --level info ""
	@gum log --level info "Individual package tests:"
	@gum log --level info "  test-accumulator       - Test accumulator package"
	@gum log --level info "  test-bulletproof       - Test bulletproof package"
	@gum log --level info "  test-core              - Test core package"
	@gum log --level info "  test-daed              - Test daed package"
	@gum log --level info "  test-dkg               - Test dkg package"
	@gum log --level info "  test-ecies             - Test ecies package"
	@gum log --level info "  test-internal          - Test internal package"
	@gum log --level info "  test-keys              - Test keys package"
	@gum log --level info "  test-mpc               - Test mpc package"
	@gum log --level info "  test-ot                - Test ot package"
	@gum log --level info "  test-paillier          - Test paillier package"
	@gum log --level info "  test-sharing           - Test sharing package"
	@gum log --level info "  test-signatures        - Test signatures package"
	@gum log --level info "  test-subtle            - Test subtle package"
	@gum log --level info "  test-tecdsa            - Test tecdsa package"
	@gum log --level info "  test-ted25519          - Test ted25519 package"
	@gum log --level info "  test-zkp               - Test zkp package"
	@gum log --level info ""
	@gum log --level info "Subpackage tests:"
	@gum log --level info "  test-core-curves       - Test core/curves package"
	@gum log --level info "  test-core-native       - Test core/curves/native package"
	@gum log --level info "  test-dkg-frost         - Test dkg/frost package"
	@gum log --level info "  test-dkg-gennaro       - Test dkg/gennaro package"
	@gum log --level info "  test-dkg-gennaro2p     - Test dkg/gennaro2p package"
	@gum log --level info "  test-ot-base           - Test ot/base package"
	@gum log --level info "  test-ot-extension      - Test ot/extension package"
	@gum log --level info "  test-sharing-v1        - Test sharing/v1 package"
	@gum log --level info "  test-signatures-bbs    - Test signatures/bbs package"
	@gum log --level info "  test-signatures-bls    - Test signatures/bls package"
	@gum log --level info "  test-signatures-common - Test signatures/common package"
	@gum log --level info "  test-signatures-schnorr - Test signatures/schnorr package"
	@gum log --level info "  test-tecdsa-dklsv1     - Test tecdsa/dklsv1 package"
	@gum log --level info "  test-ted25519-frost    - Test ted25519/frost package"
	@gum log --level info "  test-ted25519-ted25519 - Test ted25519/ted25519 package"
	@gum log --level info "  test-zkp-schnorr       - Test zkp/schnorr package"
	@gum log --level info ""
	@gum log --level info "Performance and coverage:"
	@gum log --level info "  bench                  - Run benchmarks"
	@gum log --level info "  bench-verbose          - Run benchmarks with verbose output"
	@gum log --level info "  coverage               - Generate test coverage report"
	@gum log --level info "  coverage-html          - Generate HTML test coverage report"

# Top-level package tests
test-accumulator:
	@gum log --level info "Testing accumulator package..."
	@go test ./accumulator -v

test-bulletproof:
	@gum log --level info "Testing bulletproof package..."
	@go test ./bulletproof -v

test-core:
	@gum log --level info "Testing core package..."
	@go test ./core -v

test-daed:
	@gum log --level info "Testing daed package..."
	@go test ./daed -v

test-dkg:
	@gum log --level info "Testing dkg package..."
	@go test ./dkg/... -v

test-ecies:
	@gum log --level info "Testing ecies package..."
	@go test ./ecies -v

test-internal:
	@gum log --level info "Testing internal package..."
	@go test ./internal -v

test-keys:
	@gum log --level info "Testing keys package..."
	@go test ./keys -v

test-mpc:
	@gum log --level info "Testing mpc package..."
	@go test ./mpc -v

test-ot:
	@gum log --level info "Testing ot package..."
	@go test ./ot/... -v

test-paillier:
	@gum log --level info "Testing paillier package..."
	@go test ./paillier -v

test-sharing:
	@gum log --level info "Testing sharing package..."
	@go test ./sharing/... -v

test-signatures:
	@gum log --level info "Testing signatures package..."
	@go test ./signatures/... -v

test-subtle:
	@gum log --level info "Testing subtle package..."
	@go test ./subtle -v

test-tecdsa:
	@gum log --level info "Testing tecdsa package..."
	@go test ./tecdsa/... -v

test-ted25519:
	@gum log --level info "Testing ted25519 package..."
	@go test ./ted25519/... -v


test-zkp:
	@gum log --level info "Testing zkp package..."
	@go test ./zkp/... -v

# Subpackage tests
test-core-curves:
	@gum log --level info "Testing core/curves package..."
	@go test ./core/curves -v

test-core-native:
	@gum log --level info "Testing core/curves/native package..."
	@go test ./core/curves/native/... -v

test-dkg-frost:
	@gum log --level info "Testing dkg/frost package..."
	@go test ./dkg/frost -v

test-dkg-gennaro:
	@gum log --level info "Testing dkg/gennaro package..."
	@go test ./dkg/gennaro -v

test-dkg-gennaro2p:
	@gum log --level info "Testing dkg/gennaro2p package..."
	@go test ./dkg/gennaro2p -v

test-ot-base:
	@gum log --level info "Testing ot/base package..."
	@go test ./ot/base/... -v

test-ot-extension:
	@gum log --level info "Testing ot/extension package..."
	@go test ./ot/extension/... -v

test-sharing-v1:
	@gum log --level info "Testing sharing/v1 package..."
	@go test ./sharing/v1 -v

test-signatures-bbs:
	@gum log --level info "Testing signatures/bbs package..."
	@go test ./signatures/bbs -v

test-signatures-bls:
	@gum log --level info "Testing signatures/bls package..."
	@go test ./signatures/bls/... -v

test-signatures-common:
	@gum log --level info "Testing signatures/common package..."
	@go test ./signatures/common -v

test-signatures-schnorr:
	@gum log --level info "Testing signatures/schnorr package..."
	@go test ./signatures/schnorr/... -v

test-tecdsa-dklsv1:
	@gum log --level info "Testing tecdsa/dklsv1 package..."
	@go test ./tecdsa/dklsv1/... -v

test-ted25519-frost:
	@gum log --level info "Testing ted25519/frost package..."
	@go test ./ted25519/frost -v

test-ted25519-ted25519:
	@gum log --level info "Testing ted25519/ted25519 package..."
	@go test ./ted25519/ted25519 -v

test-zkp-schnorr:
	@gum log --level info "Testing zkp/schnorr package..."
	@go test ./zkp/schnorr -v

# Performance testing
bench:
	@gum log --level info "Running benchmarks for all packages..."
	@go test ./... -bench=. -run=^$

bench-verbose:
	@gum log --level info "Running verbose benchmarks for all packages..."
	@go test ./... -bench=. -benchmem -run=^$ -v

# Coverage testing
coverage:
	@gum log --level info "Generating test coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out

coverage-html:
	@gum log --level info "Generating HTML test coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@gum log --level info "Coverage report generated: coverage.html"

# Build verification
build:
	@gum log --level info "Building all packages to verify compilation..."
	@go build ./...

# Module maintenance
mod-tidy:
	@gum log --level info "Tidying go.mod..."
	@go mod tidy

mod-verify:
	@gum log --level info "Verifying go.mod..."
	@go mod verify

mod-download:
	@gum log --level info "Downloading dependencies..."
	@go mod download

# Security scanning (if staticcheck is available)
lint:
	@gum log --level info "Running static analysis..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		gum log --level warn "staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
		go vet ./...; \
	fi

# Format code
fmt:
	@gum log --level info "Formatting code..."
	@go fmt ./...

# Check for formatting issues
fmt-check:
	@gum log --level info "Checking code formatting..."
	@test -z "$$(go fmt ./...)"

# Run all quality checks
check: fmt-check lint test
	@gum log --level info "✅ All quality checks passed!"

# Development workflow
dev: clean mod-tidy fmt lint test
	@gum log --level info "✅ Development workflow completed successfully!"
