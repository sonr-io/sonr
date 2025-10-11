#!/bin/bash

# scripts/testnet-setup.sh - Initialize Sonr testnet nodes for Docker deployment

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="/usr/local/lib/sonr-scripts"
source "${SCRIPT_DIR}/env.sh"
source "${SCRIPT_DIR}/config.sh"
source "${SCRIPT_DIR}/keys.sh"
source "${SCRIPT_DIR}/genesis.sh"

# Initialize environment
init_env

# Set defaults for testnet
NODE_TYPE="${NODE_TYPE:-validator}"  # validator, sentry
MONIKER="${MONIKER:-validator}"
VALIDATOR_NAME="${VALIDATOR_NAME:-$MONIKER}"
CHAIN_ID="${CHAIN_ID:-sonrtest_1-1}"
HOME_DIR="${HOME_DIR:-/root/.sonr}"

log_info "Setting up Sonr testnet node: $NODE_TYPE ($VALIDATOR_NAME)"

# Ensure chain directory exists
ensure_chain_dir

# Update genesis parameters
log_info "Updating genesis parameters..."
update_genesis_params

# Add constitution if available
add_constitution

# Configure node
log_info "Configuring node..."
configure_node "$CHAIN_DIR" \
    --rpc-port 26657 \
    --rest-port 1317 \
    --grpc-port 9090 \
    --grpc-web-port 9091 \
    --json-rpc-port 8545 8546 \
    --rosetta-port 8080 \
    --min-gas-prices "0${DENOM}" \
    --pruning nothing \
    --keyring-backend "$KEYRING_BACKEND" \
    --chain-id "$CHAIN_ID" \
    --output-format json

# Set consensus timeouts for faster blocks
set_consensus_timeouts "$CHAIN_DIR/config/config.toml" "5s" "1s" "1s" "1s"

# Enable CORS for RPC
set_toml_value "$CHAIN_DIR/config/config.toml" "" "cors_allowed_origins" '["*"]'

# Set pprof address
set_toml_value "$CHAIN_DIR/config/config.toml" "" "pprof_laddr" "localhost:6060"

# Set P2P address
set_toml_value "$CHAIN_DIR/config/config.toml" "p2p" "laddr" "tcp://0.0.0.0:26656"

# Generate VRF keypair
log_info "Generating VRF keypair..."
if ! generate_vrf_key "$CHAIN_DIR"; then
    log_warn "VRF key generation failed, but continuing..."
    log_warn "Note: Multi-validator encryption features may not work without VRF keys"
fi

# Create validator if this is a validator node
if [[ "$NODE_TYPE" == "validator" ]]; then
    log_info "Creating validator..."

    # Wait for node to sync (in case it's connecting to other nodes)
    if [[ "${WAIT_FOR_SYNC:-false}" == "true" ]]; then
        wait_for_sync 10
    fi

    # Create validator (this would need proper key setup)
    log_info "Validator creation requires manual key setup and gentx creation"
    log_info "Run: snrd genesis gentx <validator-key> <amount> --chain-id $CHAIN_ID"
fi

log_success "Testnet setup completed for $NODE_TYPE node: $VALIDATOR_NAME"