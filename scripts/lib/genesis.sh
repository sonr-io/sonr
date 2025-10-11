#!/bin/bash

# scripts/lib/genesis.sh - Genesis file utilities for Sonr scripts

set -euo pipefail

# Source environment and JSON helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/env.sh"
source "${SCRIPT_DIR}/jq_patch.sh"

# Generate VRF keypair for the network
# Usage: generate_vrf_key [chain_dir]
generate_vrf_key() {
    local chain_dir="${1:-$CHAIN_DIR}"

    ensure_chain_dir

    local genesis_file="$chain_dir/config/genesis.json"
    ensure_file "$genesis_file"

    local chain_id
    chain_id=$(get_json_value "$genesis_file" '.chain_id')

    if [[ -z "$chain_id" || "$chain_id" == "null" ]]; then
        log_error "Failed to extract chain-id from genesis file"
        return 1
    fi

    log_info "Generating VRF keypair for network: $chain_id"

    # Create deterministic entropy from chain-id using SHA256
    local entropy_seed
    entropy_seed=$(echo -n "$chain_id" | sha256sum | cut -d' ' -f1)

    # Generate 64 bytes of deterministic randomness
    local seed_part1="$entropy_seed"
    local seed_part2
    seed_part2=$(echo -n "$entropy_seed" | sha256sum | cut -d' ' -f1)

    # Combine to create 64 bytes of hex data
    local vrf_key_hex="${seed_part1}${seed_part2}"

    # Ensure we have exactly 128 hex characters (64 bytes)
    if [[ ${#vrf_key_hex} -ne 128 ]]; then
        log_error "Generated VRF key has incorrect size: ${#vrf_key_hex}"
        return 1
    fi

    # Path to store VRF secret key
    local vrf_key_path="$chain_dir/vrf_secret.key"

    # Ensure directory exists
    mkdir -p "$chain_dir"

    # Convert hex to binary and write to file
    echo -n "$vrf_key_hex" | xxd -r -p > "$vrf_key_path"

    # Set restrictive permissions (owner read/write only)
    chmod 0600 "$vrf_key_path"

    # Validate file was created with correct size (64 bytes)
    local file_size
    file_size=$(wc -c < "$vrf_key_path")

    if [[ $file_size -ne 64 ]]; then
        log_error "VRF key file has incorrect size: ${file_size} bytes"
        rm -f "$vrf_key_path"
        return 1
    fi

    log_success "VRF keypair generated for network: $chain_id"
    log_success "VRF secret key stored securely: $vrf_key_path"

    return 0
}

# Update genesis with Sonr-specific parameters
# Usage: update_genesis_params [genesis_file]
update_genesis_params() {
    local genesis_file="${1:-$CHAIN_DIR/config/genesis.json}"

    ensure_file "$genesis_file"

    log_info "Updating genesis parameters for Sonr"

    # Update stake denomination
    set_json_string "$genesis_file" 'app_state.staking.params.bond_denom' "$DENOM"

    # Update mint denomination
    set_json_string "$genesis_file" 'app_state.mint.params.mint_denom' "$DENOM"

    # Update crisis fee
    set_json_object "$genesis_file" 'app_state.crisis.constant_fee' "{\"denom\":\"$DENOM\",\"amount\":\"1000\"}"

    # Update minimum commission rate
    set_json_string "$genesis_file" 'app_state.staking.params.min_commission_rate' "0.050000000000000000"

    # Update block max gas
    set_json_string "$genesis_file" 'consensus.params.block.max_gas' "100000000000"

    # Update unbonding time
    set_json_string "$genesis_file" 'app_state.staking.params.unbonding_time' "300s"

    # Update downtime jail duration
    set_json_string "$genesis_file" 'app_state.slashing.params.downtime_jail_duration' "60s"

    # Update governance parameters for SDK v0.47+
    if json_path_exists "$genesis_file" '.app_state.gov.params'; then
        set_json_string "$genesis_file" 'app_state.gov.params.max_deposit_period' "30s"
        set_json_string "$genesis_file" 'app_state.gov.params.min_deposit[0].amount' "10"
        set_json_string "$genesis_file" 'app_state.gov.params.voting_period' "30s"
        set_json_string "$genesis_file" 'app_state.gov.params.quorum' "0.000000000000000000"
        set_json_string "$genesis_file" 'app_state.gov.params.threshold' "0.000000000000000000"
        set_json_string "$genesis_file" 'app_state.gov.params.veto_threshold' "0.000000000000000000"
    fi

    # Update EVM parameters if present
    if json_path_exists "$genesis_file" '.app_state.evm'; then
        set_json_string "$genesis_file" 'app_state.evm.params.evm_denom' "$DENOM"
        set_json_object "$genesis_file" 'app_state.evm.params.active_static_precompiles' '["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805"]'
    fi

    # Update ERC20 parameters if present
    if json_path_exists "$genesis_file" '.app_state.erc20'; then
        set_json_object "$genesis_file" 'app_state.erc20.params.native_precompiles' '["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]'
        set_json_object "$genesis_file" 'app_state.erc20.token_pairs' '[{"contract_owner":1,"erc20_address":"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE","denom":"'"$DENOM"'","enabled":true}]'
    fi

    # Update feemarket parameters if present
    if json_path_exists "$genesis_file" '.app_state.feemarket'; then
        set_json_bool "$genesis_file" 'app_state.feemarket.params.no_base_fee' true
        set_json_string "$genesis_file" 'app_state.feemarket.params.base_fee' "0.000000000000000000"
    fi

    # Update tokenfactory parameters if present
    if json_path_exists "$genesis_file" '.app_state.tokenfactory'; then
        set_json_object "$genesis_file" 'app_state.tokenfactory.params.denom_creation_fee' '[]'
        set_json_number "$genesis_file" 'app_state.tokenfactory.params.denom_creation_gas_consume' 100000
    fi

    # Update ABCI parameters if present
    if json_path_exists "$genesis_file" '.consensus.params.abci'; then
        set_json_string "$genesis_file" 'consensus.params.abci.vote_extensions_enable_height' "1"
    fi

    log_success "Genesis parameters updated for Sonr"
}

# Add constitution to governance if CONSTITUTION.md exists
# Usage: add_constitution [genesis_file] [constitution_file]
add_constitution() {
    local genesis_file="${1:-$CHAIN_DIR/config/genesis.json}"
    local constitution_file="${2:-CONSTITUTION.md}"

    ensure_file "$genesis_file"

    # Look for CONSTITUTION.md in the git root directory
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local git_root
    git_root="$(cd "${script_dir}/.." && pwd)"
    local constitution_path="$git_root/$constitution_file"

    if [[ ! -f "$constitution_path" ]]; then
        log_warn "Constitution file not found: $constitution_path"
        return 0
    fi

    log_info "Adding constitution from: $constitution_path"

    local constitution_content
    constitution_content=$(cat "$constitution_path" | jq -Rs .)

    set_json_object "$genesis_file" 'app_state.gov.constitution' "$constitution_content"

    log_success "Constitution added to governance"
}

# Validate genesis file
# Usage: validate_genesis [genesis_file]
validate_genesis() {
    local genesis_file="${1:-$CHAIN_DIR/config/genesis.json}"

    ensure_file "$genesis_file"

    log_info "Validating genesis file..."

    # Validate JSON
    validate_json "$genesis_file"

    # Check required fields
    local chain_id
    chain_id=$(get_json_value "$genesis_file" '.chain_id')
    if [[ -z "$chain_id" || "$chain_id" == "null" ]]; then
        log_error "Genesis file missing chain_id"
        return 1
    fi

    # Check app_state exists
    if ! json_path_exists "$genesis_file" '.app_state'; then
        log_error "Genesis file missing app_state"
        return 1
    fi

    log_success "Genesis file validation passed"
}

# Export functions
export -f generate_vrf_key update_genesis_params add_constitution validate_genesis