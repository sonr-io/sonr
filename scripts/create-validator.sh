#!/bin/bash

# scripts/create-validator.sh - Create a validator for Sonr network

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/keys.sh"
source "${SCRIPT_DIR}/lib/tx.sh"

# Initialize environment
init_env

# Ensure chain directory exists
ensure_chain_dir

# Set defaults for validator creation
VAL_NAME="${VAL_NAME:-$CHAIN_ID-validator}"
VALIDATOR_AMOUNT="5000000000${DENOM}"

log_info "Creating validator '$VAL_NAME' for Sonr network"

# Wait for node to sync
wait_for_sync 10

# Get Cosmos SDK version to determine validator creation format
set +e
cosmos_sdk_version=$($CHAIN_BIN version --long | sed -n 's/cosmos_sdk_version: \(.*\)/\1/p')
set -e

log_info "Cosmos SDK version: $cosmos_sdk_version"

# Create validator based on SDK version
if [[ "$cosmos_sdk_version" > "v0.50.0" ]]; then
    log_info "Using Cosmos SDK v0.50+ validator creation format"

    # Create validator JSON for v0.50+
    local validator_json
    validator_json=$(cat <<EOF
{
  "pubkey": "$($CHAIN_BIN tendermint show-validator $NODE_ARGS)",
  "amount": "$VALIDATOR_AMOUNT",
  "moniker": "$VAL_NAME",
  "commission-rate": "0.1",
  "commission-max-rate": "0.2",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1000000"
}
EOF
    )

    echo "$validator_json" > /tmp/validator.json

    # Submit validator creation transaction
    submit_tx "$VAL_NAME" staking create-validator /tmp/validator.json

    rm -f /tmp/validator.json
else
    log_info "Using legacy validator creation format"

    # Check if min-self-delegation parameter is supported
    local args=""
    if $CHAIN_BIN tx staking create-validator --help 2>/dev/null | grep -q "min-self-delegation"; then
        args="--min-self-delegation=1000000"
    fi

    # Submit validator creation transaction with legacy format
    submit_tx "$VAL_NAME" staking create-validator \
        --pubkey="$($CHAIN_BIN tendermint show-validator $NODE_ARGS)" \
        --moniker "$VAL_NAME" \
        --amount "$VALIDATOR_AMOUNT" \
        --commission-rate="0.10" \
        --commission-max-rate="0.20" \
        --commission-max-change-rate="0.01" \
        $args
fi

log_success "Validator '$VAL_NAME' created successfully"
