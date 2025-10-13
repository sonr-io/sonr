#!/bin/bash

# scripts/lib/config.sh - Configuration file utilities for TOML and other formats

set -euo pipefail

# Source environment and JSON helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/env.sh"
source "${SCRIPT_DIR}/jq_patch.sh"

# Check if crudini is available for TOML operations
CRUDINI_AVAILABLE=false
if command -v crudini >/dev/null 2>&1; then
    CRUDINI_AVAILABLE=true
fi

# Set TOML value using crudini if available, fallback to sed
# Usage: set_toml_value <file> <section> <key> <value>
set_toml_value() {
    local file="$1"
    local section="$2"
    local key="$3"
    local value="$4"

    ensure_file "$file"

    if [[ "$CRUDINI_AVAILABLE" == "true" ]]; then
        if crudini --set "$file" "$section" "$key" "$value" 2>/dev/null; then
            log_success "Set $section.$key = $value in $file"
            return 0
        fi
    fi

    # Fallback to sed for simple cases
    log_warn "Using sed fallback for TOML update (install crudini for better support)"

    # Escape special characters in value
    local escaped_value
    escaped_value=$(printf '%s\n' "$value" | sed 's/[[\.*^$()+?{|]/\\&/g')

    # Try to find and replace the line
    if sed -i "s|^\s*$key\s*=.*|$key = \"$escaped_value\"|g" "$file"; then
        log_success "Set $key = $value in $file (using sed)"
        return 0
    fi

    log_error "Failed to set $section.$key in $file"
    return 1
}

# Enable RPC server
# Usage: enable_rpc <config_file> [port]
enable_rpc() {
    local config_file="$1"
    local port="${2:-26657}"

    ensure_file "$config_file"

    log_info "Enabling RPC server on port $port"

    # Update laddr
    set_toml_value "$config_file" "" "laddr" "tcp://0.0.0.0:$port"

    # Enable CORS
    set_toml_value "$config_file" "" "cors_allowed_origins" '["*"]'
}

# Enable REST API server
# Usage: enable_rest <app_config_file> [port]
enable_rest() {
    local app_config_file="$1"
    local port="${2:-1317}"

    ensure_file "$app_config_file"

    log_info "Enabling REST API server on port $port"

    # Update address
    set_toml_value "$app_config_file" "api" "address" "tcp://0.0.0.0:$port"

    # Enable API
    set_toml_value "$app_config_file" "api" "enable" "true"

    # Enable unsafe CORS
    set_toml_value "$app_config_file" "api" "enabled-unsafe-cors" "true"
}

# Enable gRPC server
# Usage: enable_grpc <app_config_file> [port]
enable_grpc() {
    local app_config_file="$1"
    local port="${2:-9090}"

    ensure_file "$app_config_file"

    log_info "Enabling gRPC server on port $port"

    # Update address
    set_toml_value "$app_config_file" "grpc" "address" "0.0.0.0:$port"
}

# Enable gRPC-Web server
# Usage: enable_grpc_web <app_config_file> [port]
enable_grpc_web() {
    local app_config_file="$1"
    local port="${2:-9091}"

    ensure_file "$app_config_file"

    log_info "Enabling gRPC-Web server on port $port"

    # Update address
    set_toml_value "$app_config_file" "grpc-web" "address" "0.0.0.0:$port"
}

# Enable JSON-RPC
# Usage: enable_json_rpc <app_config_file> [port] [ws_port]
enable_json_rpc() {
    local app_config_file="$1"
    local port="${2:-8545}"
    local ws_port="${3:-8546}"

    ensure_file "$app_config_file"

    log_info "Enabling JSON-RPC on port $port (WebSocket: $ws_port)"

    # Enable JSON-RPC
    set_toml_value "$app_config_file" "json-rpc" "enable" "true"

    # Set address
    set_toml_value "$app_config_file" "json-rpc" "address" "0.0.0.0:$port"

    # Set WebSocket address
    set_toml_value "$app_config_file" "json-rpc" "ws-address" "0.0.0.0:$ws_port"

    # Enable APIs
    set_toml_value "$app_config_file" "json-rpc" "api" "eth,txpool,personal,net,debug,web3"
}

# Enable Rosetta API
# Usage: enable_rosetta <app_config_file> [port]
enable_rosetta() {
    local app_config_file="$1"
    local port="${2:-8080}"

    ensure_file "$app_config_file"

    log_info "Enabling Rosetta API on port $port"

    # Update address
    set_toml_value "$app_config_file" "rosetta" "address" "0.0.0.0:$port"
}

# Set consensus timeouts
# Usage: set_consensus_timeouts <config_file> [propose] [prevote] [precommit] [commit]
set_consensus_timeouts() {
    local config_file="$1"
    local timeout_propose="${2:-5s}"
    local timeout_prevote="${3:-1s}"
    local timeout_precommit="${4:-1s}"
    local timeout_commit="${5:-5s}"

    ensure_file "$config_file"

    log_info "Setting consensus timeouts: propose=$timeout_propose, prevote=$timeout_prevote, precommit=$timeout_precommit, commit=$timeout_commit"

    set_toml_value "$config_file" "consensus" "timeout_propose" "$timeout_propose"
    set_toml_value "$config_file" "consensus" "timeout_prevote" "$timeout_prevote"
    set_toml_value "$config_file" "consensus" "timeout_precommit" "$timeout_precommit"
    set_toml_value "$config_file" "consensus" "timeout_commit" "$timeout_commit"
}

# Set pruning strategy
# Usage: set_pruning <app_config_file> [strategy]
set_pruning() {
    local app_config_file="$1"
    local strategy="${2:-default}"

    ensure_file "$app_config_file"

    log_info "Setting pruning strategy to $strategy"

    set_toml_value "$app_config_file" "pruning" "strategy" "$strategy"
}

# Set minimum gas prices
# Usage: set_min_gas_prices <app_config_file> [price]
set_min_gas_prices() {
    local app_config_file="$1"
    local price="${2:-0${DENOM}}"

    ensure_file "$app_config_file"

    log_info "Setting minimum gas prices to $price"

    set_toml_value "$app_config_file" "minimum-gas-prices" "minimum-gas-prices" "$price"
}

# Enable metrics
# Usage: enable_metrics <config_file> [retention_time]
enable_metrics() {
    local config_file="$1"
    local retention_time="${2:-3600}"

    ensure_file "$config_file"

    log_info "Enabling metrics with retention time ${retention_time}s"

    # Enable prometheus in config.toml
    set_toml_value "$config_file" "instrumentation" "prometheus" "true"

    # Set retention time in app.toml
    if [[ -f "$config_file" ]]; then
        set_toml_value "$config_file" "telemetry" "prometheus-retention-time" "$retention_time"
    fi
}

# Set keyring backend
# Usage: set_keyring_backend <client_config_file> [backend]
set_keyring_backend() {
    local client_config_file="$1"
    local backend="${2:-test}"

    ensure_file "$client_config_file"

    log_info "Setting keyring backend to $backend"

    set_toml_value "$client_config_file" "keyring-backend" "keyring-backend" "$backend"
}

# Set chain ID in client config
# Usage: set_client_chain_id <client_config_file> [chain_id]
set_client_chain_id() {
    local client_config_file="$1"
    local chain_id="${2:-$CHAIN_ID}"

    ensure_file "$client_config_file"

    log_info "Setting client chain ID to $chain_id"

    set_toml_value "$client_config_file" "chain-id" "chain-id" "$chain_id"
}

# Set client output format
# Usage: set_client_output <client_config_file> [format]
set_client_output() {
    local client_config_file="$1"
    local format="${2:-json}"

    ensure_file "$client_config_file"

    log_info "Setting client output format to $format"

    set_toml_value "$client_config_file" "output" "output" "$format"
}

# Configure full node settings
# Usage: configure_node <config_dir> [options...]
configure_node() {
    local config_dir="$1"
    shift

    ensure_chain_dir

    local config_toml="$config_dir/config/config.toml"
    local app_toml="$config_dir/config/app.toml"
    local client_toml="$config_dir/config/client.toml"

    log_info "Configuring node in $config_dir"

    # Apply configurations based on arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --rpc-port)
                enable_rpc "$config_toml" "$2"
                shift 2
                ;;
            --rest-port)
                enable_rest "$app_toml" "$2"
                shift 2
                ;;
            --grpc-port)
                enable_grpc "$app_toml" "$2"
                shift 2
                ;;
            --grpc-web-port)
                enable_grpc_web "$app_toml" "$2"
                shift 2
                ;;
            --json-rpc-port)
                enable_json_rpc "$app_toml" "$2" "$3"
                shift 3
                ;;
            --rosetta-port)
                enable_rosetta "$app_toml" "$2"
                shift 2
                ;;
            --consensus-timeouts)
                set_consensus_timeouts "$config_toml" "$2" "$3" "$4" "$5"
                shift 5
                ;;
            --pruning)
                set_pruning "$app_toml" "$2"
                shift 2
                ;;
            --min-gas-prices)
                set_min_gas_prices "$app_toml" "$2"
                shift 2
                ;;
            --metrics)
                enable_metrics "$config_toml" "$2"
                shift 2
                ;;
            --keyring-backend)
                set_keyring_backend "$client_toml" "$2"
                shift 2
                ;;
            --chain-id)
                set_client_chain_id "$client_toml" "$2"
                shift 2
                ;;
            --output-format)
                set_client_output "$client_toml" "$2"
                shift 2
                ;;
            *)
                log_error "Unknown configuration option: $1"
                return 1
                ;;
        esac
    done

    log_success "Node configuration completed"
}

# Export functions
export -f set_toml_value enable_rpc enable_rest enable_grpc enable_grpc_web
export -f enable_json_rpc enable_rosetta set_consensus_timeouts set_pruning
export -f set_min_gas_prices enable_metrics set_keyring_backend set_client_chain_id
export -f set_client_output configure_node