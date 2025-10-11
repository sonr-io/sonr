#!/bin/bash

# scripts/update-genesis.sh - Update genesis.json with Sonr-specific parameters

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/genesis.sh"

# Initialize environment
init_env

# Ensure chain directory exists
ensure_chain_dir

log_info "Updating genesis.json file with Sonr parameters"

# Update genesis parameters using helper functions
update_genesis_params

# Add constitution if available
add_constitution

# Validate genesis file
validate_genesis

# Show node ID
$CHAIN_BIN tendermint show-node-id

log_success "Genesis update completed"
