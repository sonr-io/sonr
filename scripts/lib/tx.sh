#!/bin/bash

# scripts/lib/tx.sh - Transaction utilities for Sonr scripts

set -euo pipefail

# Source environment and key helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/env.sh"
source "${SCRIPT_DIR}/keys.sh"

# Submit transaction with automatic gas and error handling
# Usage: submit_tx <from_key> <command...>
submit_tx() {
    local from_key="$1"
    shift

    ensure_key "$from_key"

    log_info "Submitting transaction from '$from_key': $*"

    local tx_result
    tx_result=$($CHAIN_BIN tx "$@" \
        --from "$from_key" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE_URL" \
        --keyring-backend "$KEYRING_BACKEND" \
        --gas "$GAS" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --output json \
        --yes 2>&1)

    # Check for transaction hash
    local tx_hash
    tx_hash=$(echo "$tx_result" | jq -r '.txhash // empty')

    if [[ -n "$tx_hash" && "$tx_hash" != "null" ]]; then
        log_success "Transaction submitted: $tx_hash"

        # Wait for confirmation
        sleep 5

        # Query transaction result
        local tx_query
        tx_query=$($CHAIN_BIN query tx "$tx_hash" \
            --node "$NODE_URL" \
            --output json 2>/dev/null)

        local tx_code
        tx_code=$(echo "$tx_query" | jq -r '.code // 0')

        if [[ "$tx_code" == "0" ]]; then
            log_success "Transaction confirmed successfully"
            echo "$tx_hash"
            return 0
        else
            local tx_log
            tx_log=$(echo "$tx_query" | jq -r '.raw_log // "Unknown error"')
            log_error "Transaction failed (code $tx_code): $tx_log"
            return 1
        fi
    else
        log_error "Transaction submission failed: $tx_result"
        return 1
    fi
}

# Query with error handling
# Usage: query_chain <command...>
query_chain() {
    log_info "Querying chain: $*"

    local result
    result=$($CHAIN_BIN query "$@" \
        --node "$NODE_URL" \
        --output json 2>/dev/null)

    if [[ -z "$result" || "$result" == "null" ]]; then
        log_error "Query failed or returned null"
        return 1
    fi

    echo "$result"
}

# Get validator address (first available)
# Usage: get_validator_address
get_validator_address() {
    log_info "Getting validator address..."

    local validators
    validators=$(query_chain staking validators)

    local validator_address
    validator_address=$(echo "$validators" | jq -r '.validators[0].operator_address // empty')

    if [[ -z "$validator_address" || "$validator_address" == "null" ]]; then
        log_error "No validators found"
        return 1
    fi

    log_success "Using validator: $validator_address"
    echo "$validator_address"
}

# Submit governance proposal
# Usage: submit_proposal <from_key> <proposal_file> [proposal_type]
submit_proposal() {
    local from_key="$1"
    local proposal_file="$2"
    local proposal_type="${3:-}"

    ensure_file "$proposal_file"
    ensure_key "$from_key"

    log_info "Submitting governance proposal from '$from_key'"

    local submit_cmd="submit-proposal"
    if [[ -n "$proposal_type" ]]; then
        submit_cmd="$submit_cmd $proposal_type"
    fi

    # Check if legacy proposal command is needed
    if ! $CHAIN_BIN tx gov submit-proposal --help 2>/dev/null | grep -q "submit-proposal"; then
        submit_cmd="submit-legacy-proposal"
    fi

    submit_tx "$from_key" gov "$submit_cmd" "$proposal_file"
}

# Vote on governance proposal
# Usage: vote_proposal <from_key> <proposal_id> <vote_option>
vote_proposal() {
    local from_key="$1"
    local proposal_id="$2"
    local vote_option="${3:-yes}"

    ensure_key "$from_key"

    log_info "Voting '$vote_option' on proposal $proposal_id from '$from_key'"

    submit_tx "$from_key" gov vote "$proposal_id" "$vote_option"
}

# Wait for proposal to pass
# Usage: wait_for_proposal <proposal_id> [max_tries] [interval]
wait_for_proposal() {
    local proposal_id="$1"
    local max_tries="${2:-3}"
    local interval="${3:-30}"

    log_info "Waiting for proposal $proposal_id to pass..."

    local tries=0
    while [[ $tries -lt $max_tries ]]; do
        local status
        status=$(query_chain gov proposal "$proposal_id" | jq -r '.status // "unknown"')

        case "$status" in
            "PROPOSAL_STATUS_PASSED")
                log_success "Proposal $proposal_id has passed"
                return 0
                ;;
            "PROPOSAL_STATUS_REJECTED")
                log_error "Proposal $proposal_id was rejected"
                return 1
                ;;
            "PROPOSAL_STATUS_FAILED")
                log_error "Proposal $proposal_id failed"
                return 1
                ;;
            *)
                log_info "Proposal status: $status ($((tries + 1))/$max_tries)"
                sleep "$interval"
                ((tries++))
                ;;
        esac
    done

    log_error "Proposal $proposal_id did not pass after $max_tries attempts"
    return 1
}

# Stake tokens to validator
# Usage: stake_tokens <from_key> <validator_address> <amount>
stake_tokens() {
    local from_key="$1"
    local validator_address="$2"
    local amount="$3"

    ensure_key "$from_key"

    log_info "Staking $amount from '$from_key' to $validator_address"

    submit_tx "$from_key" staking delegate "$validator_address" "$amount"
}

# Query account balance
# Usage: get_balance <address> [denom]
get_balance() {
    local address="$1"
    local denom="${2:-$DENOM}"

    log_info "Getting balance for $address (denom: $denom)"

    local balance
    balance=$(query_chain bank balances "$address" | jq -r ".balances[] | select(.denom == \"$denom\") | .amount // \"0\"")

    echo "$balance"
}

# Send tokens
# Usage: send_tokens <from_key> <to_address> <amount>
send_tokens() {
    local from_key="$1"
    local to_address="$2"
    local amount="$3"

    ensure_key "$from_key"

    log_info "Sending $amount from '$from_key' to $to_address"

    submit_tx "$from_key" bank send "$from_key" "$to_address" "$amount"
}

# Export functions
export -f submit_tx query_chain get_validator_address submit_proposal
export -f vote_proposal wait_for_proposal stake_tokens get_balance send_tokens