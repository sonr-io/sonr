#!/bin/bash

# Build and optimize Identity DAO contracts for deployment
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Configuration
PROJECT_ROOT="$(dirname "$0")/.."
CONTRACTS_DIR="${PROJECT_ROOT}/contracts"
TARGET_DIR="${PROJECT_ROOT}/target/wasm32-unknown-unknown/release"
ARTIFACTS_DIR="${PROJECT_ROOT}/artifacts"

# Contract names
CONTRACTS=(
    "identity-dao-core"
    "identity-dao-voting"
    "identity-dao-proposals"
    "identity-dao-pre-propose"
)

# Check for Rust and wasm32 target
check_requirements() {
    log_info "Checking build requirements..."
    
    if ! command -v cargo &> /dev/null; then
        log_error "Rust/Cargo not found. Please install Rust."
        exit 1
    fi
    
    if ! rustup target list --installed | grep -q wasm32-unknown-unknown; then
        log_warning "wasm32-unknown-unknown target not installed. Installing..."
        rustup target add wasm32-unknown-unknown
    fi
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Docker is required for contract optimization."
        exit 1
    fi
    
    log_info "All requirements satisfied"
}

# Build contracts
build_contracts() {
    log_info "Building contracts..."
    
    cd "${PROJECT_ROOT}"
    
    # Clean previous builds
    cargo clean
    
    # Build all contracts in release mode
    RUSTFLAGS='-C link-arg=-s' cargo build --release --target wasm32-unknown-unknown
    
    if [ $? -eq 0 ]; then
        log_info "Contracts built successfully"
    else
        log_error "Failed to build contracts"
        exit 1
    fi
}

# Optimize contracts using CosmWasm optimizer
optimize_contracts() {
    log_info "Optimizing contracts for deployment..."
    
    # Create artifacts directory
    mkdir -p "${ARTIFACTS_DIR}"
    
    # Run optimizer in Docker
    docker run --rm -v "${PROJECT_ROOT}":/code \
        --mount type=volume,source="dao_contracts_cache",target=/target \
        --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
        cosmwasm/optimizer:0.16.0
    
    if [ $? -eq 0 ]; then
        log_info "Contract optimization complete"
        
        # Move optimized contracts to artifacts
        mv "${PROJECT_ROOT}"/artifacts/*.wasm "${ARTIFACTS_DIR}/" 2>/dev/null || true
    else
        log_error "Failed to optimize contracts"
        exit 1
    fi
}

# Generate schemas
generate_schemas() {
    log_info "Generating contract schemas..."
    
    cd "${PROJECT_ROOT}"
    
    for contract in "${CONTRACTS[@]}"; do
        contract_dir="${CONTRACTS_DIR}/${contract//-/_}"
        
        if [ -d "${contract_dir}" ]; then
            log_info "Generating schema for ${contract}..."
            cd "${contract_dir}"
            cargo schema
        fi
    done
    
    log_info "Schema generation complete"
}

# Verify contract sizes
verify_sizes() {
    log_info "Verifying contract sizes..."
    
    MAX_SIZE=$((600 * 1024)) # 600 KB max size for Cosmos chains
    
    for wasm_file in "${ARTIFACTS_DIR}"/*.wasm; do
        if [ -f "$wasm_file" ]; then
            size=$(stat -f%z "$wasm_file" 2>/dev/null || stat -c%s "$wasm_file" 2>/dev/null)
            size_kb=$((size / 1024))
            filename=$(basename "$wasm_file")
            
            if [ $size -gt $MAX_SIZE ]; then
                log_error "$filename is too large: ${size_kb}KB (max: 600KB)"
                exit 1
            else
                log_info "$filename: ${size_kb}KB ✓"
            fi
        fi
    done
}

# Generate checksums
generate_checksums() {
    log_info "Generating checksums..."
    
    cd "${ARTIFACTS_DIR}"
    
    if [ -f checksums.txt ]; then
        rm checksums.txt
    fi
    
    for wasm_file in *.wasm; do
        if [ -f "$wasm_file" ]; then
            if command -v sha256sum &> /dev/null; then
                sha256sum "$wasm_file" >> checksums.txt
            else
                shasum -a 256 "$wasm_file" >> checksums.txt
            fi
        fi
    done
    
    log_info "Checksums saved to artifacts/checksums.txt"
}

# Main build flow
main() {
    log_info "Starting Identity DAO contract build process..."
    
    # Check requirements
    check_requirements
    
    # Build contracts
    build_contracts
    
    # Optimize contracts
    optimize_contracts
    
    # Generate schemas
    generate_schemas
    
    # Verify sizes
    verify_sizes
    
    # Generate checksums
    generate_checksums
    
    log_info "✅ Build complete!"
    log_info ""
    log_info "=== Build Summary ==="
    log_info "Optimized contracts location: ${ARTIFACTS_DIR}"
    log_info "Contract schemas location: ${CONTRACTS_DIR}/*/schema"
    log_info ""
    log_info "Next steps:"
    log_info "1. Review contract sizes in ${ARTIFACTS_DIR}"
    log_info "2. Deploy contracts using: ./scripts/deploy_testnet.sh"
}

# Run main build
main "$@"