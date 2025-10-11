#!/bin/bash

# scripts/create-ics.sh - Create ICS proposal for Sonr network

set -euo pipefail

# Source helper libraries
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/keys.sh"
source "${SCRIPT_DIR}/lib/tx.sh"

# Initialize environment
init_env

# Set defaults for ICS setup
KEY_NAME="${KEY_NAME:-ics-setup}"
STAKE_AMOUNT="10000000${DENOM}"
MAX_RETRIES="${MAX_RETRIES:-3}"
RETRY_INTERVAL="${RETRY_INTERVAL:-30}"

# Validate required parameters
if [[ -z "${PROPOSAL_FILE:-}" ]]; then
    log_error "PROPOSAL_FILE environment variable is required"
    exit 1
fi

ensure_file "$PROPOSAL_FILE"

log_info "Setting up ICS proposal with key '$KEY_NAME'"

# Import key from keys config if available
if [[ -f "${KEYS_CONFIG:-}" ]]; then
    local key_mnemonic
    key_mnemonic=$(jq -r ".keys[0].mnemonic" "$KEYS_CONFIG" 2>/dev/null || echo "")

    if [[ -n "$key_mnemonic" && "$key_mnemonic" != "null" ]]; then
        import_mnemonic "$KEY_NAME" "$key_mnemonic"
    else
        log_warn "No valid mnemonic found in $KEYS_CONFIG, using existing key or create one manually"
    fi
else
    log_warn "KEYS_CONFIG not set, ensure key '$KEY_NAME' exists"
fi

# Ensure key exists
ensure_key "$KEY_NAME"

# Get validator address and stake tokens
local validator_address
validator_address=$(get_validator_address)

stake_tokens "$KEY_NAME" "$validator_address" "$STAKE_AMOUNT"

# Determine proposal command (legacy vs new)
local submit_cmd="submit-proposal"
if $CHAIN_BIN tx gov --help 2>/dev/null | grep -q "submit-legacy-proposal"; then
    submit_cmd="submit-legacy-proposal"
fi

log_info "Using proposal command: $submit_cmd"

# Submit proposal
local tx_hash
tx_hash=$(submit_proposal "$KEY_NAME" "$PROPOSAL_FILE" "consumer-addition")

# Extract proposal ID from transaction
local proposal_id
proposal_id=$(query_chain tx "$tx_hash" | jq -r '.logs[0].events[] | select(.type=="submit_proposal").attributes[] | select(.key=="proposal_id").value // empty')

if [[ -z "$proposal_id" || "$proposal_id" == "null" ]]; then
    log_error "Failed to extract proposal ID from transaction"
    exit 1
fi

log_info "Proposal ID: $proposal_id"

# Vote on proposal
vote_proposal "$KEY_NAME" "$proposal_id" "yes"

# Wait for proposal to pass
wait_for_proposal "$proposal_id" "$MAX_RETRIES" "$RETRY_INTERVAL"

log_success "ICS proposal setup completed successfully"
