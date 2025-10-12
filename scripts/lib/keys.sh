#!/bin/bash

# scripts/lib/keys.sh - Key management utilities for Sonr scripts

set -euo pipefail

# Source environment helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/env.sh"

# Import mnemonic and add key to keyring
# Usage: import_mnemonic <key_name> <mnemonic> [algo]
import_mnemonic() {
    local key_name="$1"
    local mnemonic="$2"
    local algo="${3:-eth_secp256k1}"

    log_info "Importing key '$key_name'"

    if [[ -z "$mnemonic" ]]; then
        log_error "Mnemonic cannot be empty"
        return 1
    fi

    # Use Docker wrapper if needed
    if is_docker; then
        echo "$mnemonic" | $CHAIN_BIN keys add "$key_name" \
            --keyring-backend "$KEYRING_BACKEND" \
            --algo "$algo" \
            --recover
    else
        echo "$mnemonic" | $CHAIN_BIN keys add "$key_name" \
            --keyring-backend "$KEYRING_BACKEND" \
            --algo "$algo" \
            --home "$CHAIN_DIR" \
            --recover
    fi

    log_success "Key '$key_name' imported successfully"
}

# Ensure key exists in keyring
# Usage: ensure_key <key_name>
ensure_key() {
    local key_name="$1"

    if ! $CHAIN_BIN keys show "$key_name" \
        --keyring-backend "$KEYRING_BACKEND" \
        --home "$CHAIN_DIR" \
        --output json >/dev/null 2>&1; then

        log_error "Key '$key_name' not found in keyring"
        return 1
    fi

    log_info "Key '$key_name' exists in keyring"
}

# Get key address
# Usage: get_key_address <key_name>
get_key_address() {
    local key_name="$1"

    ensure_key "$key_name"

    local address
    address=$($CHAIN_BIN keys show "$key_name" \
        --keyring-backend "$KEYRING_BACKEND" \
        --home "$CHAIN_DIR" \
        --output json | jq -r '.address')

    if [[ -z "$address" || "$address" == "null" ]]; then
        log_error "Failed to get address for key '$key_name'"
        return 1
    fi

    echo "$address"
}

# Fund key with tokens from genesis
# Usage: fund_key <key_name> <amount>
fund_key() {
    local key_name="$1"
    local amount="$2"

    ensure_key "$key_name"

    local address
    address=$(get_key_address "$key_name")

    log_info "Funding key '$key_name' ($address) with $amount"

    if is_docker; then
        $CHAIN_BIN genesis add-genesis-account "$address" "$amount" \
            --keyring-backend "$KEYRING_BACKEND"
    else
        $CHAIN_BIN genesis add-genesis-account "$address" "$amount" \
            --keyring-backend "$KEYRING_BACKEND" \
            --home "$CHAIN_DIR"
    fi

    log_success "Funded key '$key_name' with $amount"
}

# Delegate tokens to validator
# Usage: delegate_to_validator <delegator_key> <validator_address> <amount>
delegate_to_validator() {
    local delegator_key="$1"
    local validator_address="$2"
    local amount="$3"

    ensure_key "$delegator_key"

    log_info "Delegating $amount from '$delegator_key' to validator $validator_address"

    $CHAIN_BIN tx staking delegate "$validator_address" "$amount" \
        --from "$delegator_key" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE_URL" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas "$GAS" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --yes

    log_success "Delegation completed"
}

# Create validator with key
# Usage: create_validator <key_name> <moniker> <amount> [options...]
create_validator() {
    local key_name="$1"
    local moniker="$2"
    local amount="$3"
    shift 3

    ensure_key "$key_name"

    local validator_address
    validator_address=$(get_key_address "$key_name")

    log_info "Creating validator '$moniker' with key '$key_name'"

    # Build validator JSON
    local validator_json
    validator_json=$(cat <<EOF
{
  "pubkey": "$($CHAIN_BIN tendermint show-validator)",
  "amount": "$amount",
  "moniker": "$moniker",
  "commission-rate": "0.1",
  "commission-max-rate": "0.2",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1000000"
}
EOF
    )

    echo "$validator_json" > /tmp/validator.json

    $CHAIN_BIN tx staking create-validator /tmp/validator.json \
        --from "$key_name" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE_URL" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas "$GAS" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --yes

    rm -f /tmp/validator.json

    log_success "Validator '$moniker' created successfully"
}

# Wait for node to sync
# Usage: wait_for_sync [max_tries]
wait_for_sync() {
    local max_tries="${1:-10}"

    log_info "Waiting for node to sync..."

    local tries=0
    while [[ $tries -lt $max_tries ]]; do
        if $CHAIN_BIN status \
            --node "$NODE_URL" \
            --output json 2>/dev/null | jq -e '.SyncInfo.catching_up == false' >/dev/null 2>&1; then

            log_success "Node is synced"
            return 0
        fi

        log_info "Still syncing... ($((tries + 1))/$max_tries)"
        sleep 30
        ((tries++))
    done

    log_error "Node failed to sync after $max_tries attempts"
    return 1
}

# List keys in keyring
# Usage: list_keys
list_keys() {
    $CHAIN_BIN keys list \
        --keyring-backend "$KEYRING_BACKEND" \
        --home "$CHAIN_DIR" \
        --output json | jq
}

# Export functions
export -f import_mnemonic ensure_key get_key_address fund_key
export -f delegate_to_validator create_validator wait_for_sync list_keys