#!/bin/bash

# scripts/update-config.sh - Update configuration files for Sonr node

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/config.sh"

# Initialize environment
init_env

# Ensure chain directory exists
ensure_chain_dir

log_info "Updating configuration files for Sonr node"

# Configure node with standard settings
configure_node "$CHAIN_DIR" \
    --rpc-port 26657 \
    --rest-port 1317 \
    --grpc-port 9090 \
    --grpc-web-port 9091 \
    --json-rpc-port 8545 8546 \
    --rosetta-port 8080 \
    --min-gas-prices "0${DENOM}" \
    --pruning default \
    --keyring-backend "$KEYRING_BACKEND" \
    --chain-id "$CHAIN_ID" \
    --output-format json

# Enable metrics if requested
if [[ "${METRICS:-false}" == "true" ]]; then
    enable_metrics "$CHAIN_DIR/config/config.toml" 3600
fi

# Set consensus timeouts if provided
if [[ -n "${TIMEOUT_PROPOSE:-}" ]]; then
    set_consensus_timeouts "$CHAIN_DIR/config/config.toml" \
        "$TIMEOUT_PROPOSE" \
        "${TIMEOUT_PREVOTE:-1s}" \
        "${TIMEOUT_PRECOMMIT:-1s}" \
        "${TIMEOUT_COMMIT:-5s}"
fi

log_success "Configuration update completed"
