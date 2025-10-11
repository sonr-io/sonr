#!/bin/bash

# scripts/create-genesis.sh - Create genesis file for Sonr network

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/keys.sh"
source "${SCRIPT_DIR}/lib/genesis.sh"

# Initialize environment
init_env

# Set defaults for genesis creation
FAUCET_ENABLED="${FAUCET_ENABLED:=true}"
NUM_VALIDATORS="${NUM_VALIDATORS:=1}"
NUM_RELAYERS="${NUM_RELAYERS:=0}"

# Match init-testnet.sh allocation: 100000000000000000000000000snr = 100000000000000000000000000000000usnr
BASE_COINS="${BASE_COINS:=100000000000000000000000000000000${DENOM},100000000000000000000000000snr}"
VALIDATOR_AMOUNT="1000000000000000000000000000${DENOM}"

# Check if binary has genesis subcommand
CHAIN_GENESIS_CMD=""
if $CHAIN_BIN 2>&1 | grep -q "genesis-related subcommands"; then
    CHAIN_GENESIS_CMD="genesis"
fi

log_info "Creating genesis for Sonr network: $CHAIN_ID"

# Initialize chain
local genesis_mnemonic
genesis_mnemonic=$(jq -r ".genesis[0].mnemonic" "$KEYS_CONFIG")
echo "$genesis_mnemonic" | $CHAIN_BIN init "$CHAIN_ID" --chain-id "$CHAIN_ID" --default-denom "$DENOM" --recover

# Add genesis accounts
log_info "Adding genesis accounts..."

# Genesis key
local genesis_key_name
genesis_key_name=$(jq -r ".genesis[0].name" "$KEYS_CONFIG")
import_mnemonic "$genesis_key_name" "$genesis_mnemonic"
fund_key "$genesis_key_name" "$BASE_COINS"

# Faucet key
local faucet_key_name
faucet_key_name=$(jq -r ".faucet[0].name" "$KEYS_CONFIG")
local faucet_mnemonic
faucet_mnemonic=$(jq -r ".faucet[0].mnemonic" "$KEYS_CONFIG")
import_mnemonic "$faucet_key_name" "$faucet_mnemonic"
fund_key "$faucet_key_name" "$BASE_COINS"

# Test key
local test_key_name
test_key_name=$(jq -r ".keys[0].name" "$KEYS_CONFIG")
local test_mnemonic
test_mnemonic=$(jq -r ".keys[0].mnemonic" "$KEYS_CONFIG")
import_mnemonic "$test_key_name" "$test_mnemonic"
fund_key "$test_key_name" "$BASE_COINS"

# Add relayer keys if faucet is disabled
if [[ "$FAUCET_ENABLED" == "false" && "$NUM_RELAYERS" -gt 0 ]]; then
    for i in $(seq 0 "$NUM_RELAYERS"); do
        local relayer_key_name
        relayer_key_name=$(jq -r ".relayers[$i].name" "$KEYS_CONFIG")
        local relayer_mnemonic
        relayer_mnemonic=$(jq -r ".relayers[$i].mnemonic" "$KEYS_CONFIG")
        import_mnemonic "$relayer_key_name" "$relayer_mnemonic"
        fund_key "$relayer_key_name" "$BASE_COINS"

        local relayer_cli_key_name
        relayer_cli_key_name=$(jq -r ".relayers_cli[$i].name" "$KEYS_CONFIG")
        local relayer_cli_mnemonic
        relayer_cli_mnemonic=$(jq -r ".relayers_cli[$i].mnemonic" "$KEYS_CONFIG")
        import_mnemonic "$relayer_cli_key_name" "$relayer_cli_mnemonic"
        fund_key "$relayer_cli_key_name" "$BASE_COINS"
    done
fi

# Add additional validator keys if needed
if [[ "$FAUCET_ENABLED" == "false" && "$NUM_VALIDATORS" -gt 1 ]]; then
    for i in $(seq 1 "$NUM_VALIDATORS"); do
        local val_key_name="${genesis_key_name}-${i}"
        local val_mnemonic
        val_mnemonic=$(jq -r ".validators[0].mnemonic" "$KEYS_CONFIG")
        import_mnemonic "$val_key_name" "$val_mnemonic"
        fund_key "$val_key_name" "$BASE_COINS"
    done
fi

# Create genesis transaction
log_info "Creating genesis transaction..."
$CHAIN_BIN "$CHAIN_GENESIS_CMD" gentx "$genesis_key_name" "$VALIDATOR_AMOUNT" \
    --keyring-backend "$KEYRING_BACKEND" \
    --chain-id "$CHAIN_ID" \
    --gas-prices "0${DENOM}"

log_info "Genesis transaction output:"
cat "$CHAIN_DIR/config/gentx"/*.json | jq

# Collect genesis transactions
log_info "Collecting genesis transactions..."
$CHAIN_BIN "$CHAIN_GENESIS_CMD" collect-gentxs

# Generate VRF keypair
log_info "Generating VRF keypair..."
if ! generate_vrf_key "$CHAIN_DIR"; then
    log_warn "VRF key generation failed, but continuing..."
    log_warn "Note: Multi-validator encryption features may not work without VRF keys"
fi

log_success "Genesis creation completed"
