#!/bin/bash

# Setup IBC channels between Cosmos Hub testnet and Sonr chain
# for Identity DAO cross-chain communication

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Configuration
# Cosmos Hub testnet
HUB_CHAIN_ID="theta-testnet-001"
HUB_RPC="https://rpc.sentry-01.theta-testnet.polypore.xyz:443"
HUB_GRPC="grpc.sentry-01.theta-testnet.polypore.xyz:9090"
HUB_PREFIX="cosmos"
HUB_DENOM="uatom"
HUB_GAS_PRICES="0.025uatom"

# Sonr testnet
SONR_CHAIN_ID="sonrtest_1-1"
SONR_RPC="http://localhost:26657"
SONR_GRPC="localhost:9090"
SONR_PREFIX="sonr"
SONR_DENOM="usnr"
SONR_GAS_PRICES="0.025usnr"

# IBC Configuration
IBC_VERSION="ics20-1"
RELAYER_NAME="identity-dao-relayer"
RELAYER_HOME="${HOME}/.relayer"
PATH_NAME="hub-sonr"

# Contract ports (from deployment)
VOTING_PORT="wasm.cosmos1voting_contract_address"  # Will be replaced with actual address
PROPOSALS_PORT="wasm.cosmos1proposals_contract_address"

# Check if Hermes relayer is installed
check_hermes() {
    if ! command -v hermes &> /dev/null; then
        log_error "Hermes relayer not found. Installing..."
        
        # Install Hermes based on OS
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install hermes
        else
            # Linux installation
            curl -L https://github.com/informalsystems/hermes/releases/download/v1.8.0/hermes-v1.8.0-x86_64-unknown-linux-gnu.tar.gz | tar xz
            sudo mv hermes /usr/local/bin/
        fi
    fi
    
    log_info "Found Hermes: $(hermes version)"
}

# Initialize Hermes configuration
init_hermes_config() {
    log_info "Initializing Hermes configuration..."
    
    mkdir -p "${RELAYER_HOME}"
    
    cat > "${RELAYER_HOME}/config.toml" << EOF
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '127.0.0.1'
port = 3000

[telemetry]
enabled = false
host = '127.0.0.1'
port = 3001

# Cosmos Hub testnet configuration
[[chains]]
id = '${HUB_CHAIN_ID}'
type = 'CosmosSdk'
rpc_addr = '${HUB_RPC}'
grpc_addr = '${HUB_GRPC}'
event_source = { mode = 'push', url = '${HUB_RPC/https/wss}/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = '${HUB_PREFIX}'
key_name = 'hub-relayer'
store_prefix = 'ibc'
gas_price = { price = 0.025, denom = '${HUB_DENOM}' }
max_gas = 6000000
default_gas = 1000000
gas_multiplier = 1.2
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
memo_prefix = 'Identity DAO IBC'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[chains.packet_filter]
policy = 'allow'
list = [
    ['wasm*', '*'],  # Allow all CosmWasm IBC traffic
]

# Sonr testnet configuration
[[chains]]
id = '${SONR_CHAIN_ID}'
type = 'CosmosSdk'
rpc_addr = '${SONR_RPC}'
grpc_addr = '${SONR_GRPC}'
event_source = { mode = 'push', url = '${SONR_RPC/http/ws}/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = '${SONR_PREFIX}'
key_name = 'sonr-relayer'
store_prefix = 'ibc'
gas_price = { price = 0.025, denom = '${SONR_DENOM}' }
max_gas = 6000000
default_gas = 1000000
gas_multiplier = 1.2
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '30s'
memo_prefix = 'Identity DAO IBC'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[chains.packet_filter]
policy = 'allow'
list = [
    ['transfer', 'channel-*'],
    ['wasm*', '*'],
]
EOF

    log_info "Hermes config created at ${RELAYER_HOME}/config.toml"
}

# Add relayer keys
add_relayer_keys() {
    log_info "Adding relayer keys..."
    
    # Add Cosmos Hub key
    log_info "Adding Cosmos Hub relayer key..."
    hermes keys add \
        --chain "${HUB_CHAIN_ID}" \
        --mnemonic-file hub-relayer.mnemonic \
        --key-name hub-relayer
    
    # Add Sonr key
    log_info "Adding Sonr relayer key..."
    hermes keys add \
        --chain "${SONR_CHAIN_ID}" \
        --mnemonic-file sonr-relayer.mnemonic \
        --key-name sonr-relayer
}

# Create IBC clients
create_clients() {
    log_info "Creating IBC clients..."
    
    # Create client on Cosmos Hub for Sonr
    log_info "Creating Sonr client on Cosmos Hub..."
    hermes create client \
        --host-chain "${HUB_CHAIN_ID}" \
        --reference-chain "${SONR_CHAIN_ID}"
    
    # Create client on Sonr for Cosmos Hub
    log_info "Creating Cosmos Hub client on Sonr..."
    hermes create client \
        --host-chain "${SONR_CHAIN_ID}" \
        --reference-chain "${HUB_CHAIN_ID}"
}

# Create IBC connection
create_connection() {
    log_info "Creating IBC connection..."
    
    hermes create connection \
        --a-chain "${HUB_CHAIN_ID}" \
        --b-chain "${SONR_CHAIN_ID}"
}

# Create IBC channels for contracts
create_channels() {
    log_info "Creating IBC channels for Identity DAO contracts..."
    
    # Load deployed contract addresses
    if [ -f "deployment_ids.env" ]; then
        source deployment_ids.env
    else
        log_error "deployment_ids.env not found. Please deploy contracts first."
        exit 1
    fi
    
    # Channel for Voting contract to query x/did module
    log_info "Creating channel for Voting contract..."
    hermes create channel \
        --a-chain "${HUB_CHAIN_ID}" \
        --a-port "wasm.${VOTING_ADDR}" \
        --b-port "did" \
        --order unordered \
        --version "did-ibc-v1"
    
    # Channel for Proposals contract
    log_info "Creating channel for Proposals contract..."
    hermes create channel \
        --a-chain "${HUB_CHAIN_ID}" \
        --a-port "wasm.${PROPOSALS_ADDR}" \
        --b-port "dwn" \
        --order unordered \
        --version "dwn-ibc-v1"
    
    # Standard transfer channel for treasury operations
    log_info "Creating transfer channel..."
    hermes create channel \
        --a-chain "${HUB_CHAIN_ID}" \
        --a-port "transfer" \
        --b-port "transfer" \
        --order unordered \
        --version "${IBC_VERSION}"
}

# Start relayer
start_relayer() {
    log_info "Starting Hermes relayer..."
    
    # Start in background
    hermes start &
    RELAYER_PID=$!
    
    log_info "Relayer started with PID: ${RELAYER_PID}"
    echo "${RELAYER_PID}" > relayer.pid
    
    # Give it time to establish connections
    sleep 10
    
    # Check if relayer is running
    if ps -p ${RELAYER_PID} > /dev/null; then
        log_info "Relayer is running successfully"
    else
        log_error "Relayer failed to start"
        exit 1
    fi
}

# Query IBC channels
query_channels() {
    log_info "Querying established IBC channels..."
    
    # Query channels on Cosmos Hub
    log_info "Channels on Cosmos Hub:"
    hermes query channels --chain "${HUB_CHAIN_ID}"
    
    # Query channels on Sonr
    log_info "Channels on Sonr:"
    hermes query channels --chain "${SONR_CHAIN_ID}"
}

# Test IBC connectivity
test_ibc() {
    log_info "Testing IBC connectivity..."
    
    # Test transfer channel
    log_info "Testing transfer channel..."
    
    # Get channel IDs
    TRANSFER_CHANNEL=$(hermes query channels --chain "${HUB_CHAIN_ID}" | grep transfer | head -1 | awk '{print $1}')
    
    if [ -n "$TRANSFER_CHANNEL" ]; then
        log_info "Transfer channel: ${TRANSFER_CHANNEL}"
        
        # Send test transfer
        gaiad tx ibc-transfer transfer \
            transfer "${TRANSFER_CHANNEL}" \
            sonr1test_address \
            1000uatom \
            --from hub-relayer \
            --chain-id "${HUB_CHAIN_ID}" \
            --node "${HUB_RPC}" \
            --gas-prices "${HUB_GAS_PRICES}" \
            --packet-timeout-height 0-1000 \
            -y
        
        log_info "Test transfer sent"
    else
        log_warning "No transfer channel found"
    fi
}

# Monitor IBC packets
monitor_packets() {
    log_info "Monitoring IBC packet flow..."
    
    hermes query packet pending \
        --chain "${HUB_CHAIN_ID}" \
        --port transfer \
        --channel "${TRANSFER_CHANNEL}"
}

# Main IBC setup flow
main() {
    log_info "Starting IBC setup for Identity DAO..."
    
    # Check requirements
    check_hermes
    
    # Initialize configuration
    init_hermes_config
    
    # Check if keys exist or need to be created
    if [ ! -f "hub-relayer.mnemonic" ] || [ ! -f "sonr-relayer.mnemonic" ]; then
        log_error "Relayer mnemonics not found. Please create:"
        log_info "1. hub-relayer.mnemonic - Funded account on Cosmos Hub testnet"
        log_info "2. sonr-relayer.mnemonic - Funded account on Sonr testnet"
        exit 1
    fi
    
    # Add keys
    add_relayer_keys
    
    # Create IBC infrastructure
    create_clients
    create_connection
    create_channels
    
    # Start relayer
    start_relayer
    
    # Query established channels
    query_channels
    
    # Test connectivity
    test_ibc
    
    # Monitor packets
    monitor_packets
    
    log_info "✅ IBC setup complete!"
    log_info ""
    log_info "=== IBC Summary ==="
    log_info "Relayer PID: $(cat relayer.pid)"
    log_info "Config: ${RELAYER_HOME}/config.toml"
    log_info ""
    log_info "Next steps:"
    log_info "1. Monitor relayer logs: hermes start"
    log_info "2. Query channels: hermes query channels --chain ${HUB_CHAIN_ID}"
    log_info "3. Test cross-chain queries: ./scripts/test_did_query.sh"
}

# Run main IBC setup
main "$@"