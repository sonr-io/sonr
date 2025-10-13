#!/bin/bash

# scripts/lib/env.sh - Environment defaults and utility functions for Sonr scripts

set -euo pipefail

# Default environment variables for Sonr
export DENOM="${DENOM:=usnr}"
export CHAIN_BIN="${CHAIN_BIN:=snrd}"
export CHAIN_DIR="${CHAIN_DIR:=$HOME/.sonr}"
export CHAIN_ID="${CHAIN_ID:=sonrtest_1-1}"
export KEYRING_BACKEND="${KEYRING_BACKEND:=test}"
export NODE_URL="${NODE_URL:=http://0.0.0.0:26657}"
export GAS="${GAS:=auto}"
export GAS_ADJUSTMENT="${GAS_ADJUSTMENT:=1.5}"

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

# Utility functions
require_cmd() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Required command '$cmd' not found. Please install it."
        exit 1
    fi
}

join_path() {
    local base="$1"
    local path="$2"
    echo "$base/$path" | sed 's#//#/#g'
}

# Check if running in Docker
is_docker() {
    [[ -f /.dockerenv ]] || [[ -n "${DOCKER_CONTAINER:-}" ]]
}

# Get absolute path
abs_path() {
    local path="$1"
    if [[ -d "$path" ]]; then
        cd "$path" && pwd
    else
        cd "$(dirname "$path")" && echo "$(pwd)/$(basename "$path")"
    fi
}

# Validate chain directory exists
ensure_chain_dir() {
    if [[ ! -d "$CHAIN_DIR" ]]; then
        log_error "Chain directory does not exist: $CHAIN_DIR"
        exit 1
    fi
}

# Validate binary exists
ensure_binary() {
    if ! command -v "$CHAIN_BIN" >/dev/null 2>&1; then
        log_error "Binary '$CHAIN_BIN' not found in PATH"
        exit 1
    fi
}

# Check if file exists
ensure_file() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        log_error "Required file not found: $file"
        exit 1
    fi
}

# Retry function for operations that might fail
retry() {
    local max_attempts="$1"
    local delay="$2"
    local cmd="$3"
    shift 3

    local attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        log_info "Attempt $attempt/$max_attempts: $cmd $*"
        if "$cmd" "$@"; then
            return 0
        fi

        if [[ $attempt -lt $max_attempts ]]; then
            log_warn "Command failed, retrying in ${delay}s..."
            sleep "$delay"
        fi
        ((attempt++))
    done

    log_error "Command failed after $max_attempts attempts: $cmd $*"
    return 1
}

# Initialize environment
init_env() {
    require_cmd jq
    ensure_binary

    # Set up cleanup trap
    trap cleanup EXIT

    log_info "Environment initialized for Sonr chain: $CHAIN_ID"
}

cleanup() {
    # Remove temporary files if they exist
    rm -f /tmp/genesis.json /tmp/config.toml /tmp/app.toml /tmp/client.toml
}

# Export functions for use in other scripts
export -f log_info log_warn log_error log_success
export -f require_cmd join_path is_docker abs_path
export -f ensure_chain_dir ensure_binary ensure_file
export -f retry init_env cleanup