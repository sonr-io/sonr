#!/bin/bash
# Identity DAO Deployment Script
# Deploys all Identity DAO contracts to Sonr testnet

set -e

# Configuration
CHAIN_ID="${CHAIN_ID:-sonrtest_1-1}"
NODE="${NODE:-http://localhost:26657}"
KEYRING="${KEYRING:-test}"
DEPLOYER="${DEPLOYER:-deployer}"
GAS_PRICES="${GAS_PRICES:-0.025usnr}"
GAS_ADJUSTMENT="${GAS_ADJUSTMENT:-1.5}"

# Contract paths
CONTRACTS_DIR="$(dirname "$0")/../artifacts"
SHARED_WASM="${CONTRACTS_DIR}/identity_dao_shared.wasm"
CORE_WASM="${CONTRACTS_DIR}/identity_dao_core.wasm"
VOTING_WASM="${CONTRACTS_DIR}/identity_dao_voting.wasm"
PROPOSALS_WASM="${CONTRACTS_DIR}/identity_dao_proposals.wasm"
PRE_PROPOSE_WASM="${CONTRACTS_DIR}/identity_dao_pre_propose.wasm"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v snrd &> /dev/null; then
        log_error "snrd is not installed. Please run 'make install'"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed. Please install jq"
        exit 1
    fi
    
    log_info "All dependencies satisfied"
}

# Build contracts
build_contracts() {
    log_info "Building Identity DAO contracts..."
    
    cd "$(dirname "$0")/.."
    
    # Build with optimizer
    docker run --rm -v "$(pwd)":/code \
        --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
        --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
        cosmwasm/workspace-optimizer:0.13.0
    
    # Move artifacts
    mkdir -p artifacts
    mv artifacts/*.wasm artifacts/ 2>/dev/null || true
    
    log_info "Contracts built successfully"
}

# Store contract code
store_contract() {
    local wasm_file=$1
    local contract_name=$2
    
    log_info "Storing $contract_name contract..."
    
    if [ ! -f "$wasm_file" ]; then
        log_error "Contract file not found: $wasm_file"
        exit 1
    fi
    
    local tx_result=$(snrd tx wasm store "$wasm_file" \
        --from "$DEPLOYER" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --gas-prices "$GAS_PRICES" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --keyring-backend "$KEYRING" \
        --output json \
        --yes)
    
    local tx_hash=$(echo "$tx_result" | jq -r .txhash)
    
    # Wait for transaction
    sleep 6
    
    # Get code ID from events
    local code_id=$(snrd query tx "$tx_hash" \
        --node "$NODE" \
        --output json | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
    
    if [ -z "$code_id" ]; then
        log_error "Failed to get code ID for $contract_name"
        exit 1
    fi
    
    log_info "$contract_name stored with code ID: $code_id"
    echo "$code_id"
}

# Instantiate contract
instantiate_contract() {
    local code_id=$1
    local init_msg=$2
    local label=$3
    local admin=${4:-$DEPLOYER}
    
    log_info "Instantiating contract: $label"
    
    local tx_result=$(snrd tx wasm instantiate "$code_id" "$init_msg" \
        --from "$DEPLOYER" \
        --label "$label" \
        --admin "$admin" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --gas-prices "$GAS_PRICES" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --keyring-backend "$KEYRING" \
        --output json \
        --yes)
    
    local tx_hash=$(echo "$tx_result" | jq -r .txhash)
    
    # Wait for transaction
    sleep 6
    
    # Get contract address from events
    local contract_addr=$(snrd query tx "$tx_hash" \
        --node "$NODE" \
        --output json | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
    
    if [ -z "$contract_addr" ]; then
        log_error "Failed to get contract address for $label"
        exit 1
    fi
    
    log_info "$label instantiated at: $contract_addr"
    echo "$contract_addr"
}

# Main deployment flow
main() {
    log_info "Starting Identity DAO deployment..."
    
    # Check dependencies
    check_dependencies
    
    # Build contracts if artifacts don't exist
    if [ ! -d "$CONTRACTS_DIR" ] || [ -z "$(ls -A $CONTRACTS_DIR/*.wasm 2>/dev/null)" ]; then
        build_contracts
    else
        log_info "Using existing contract artifacts"
    fi
    
    # Store contract codes
    log_info "Storing contract codes on chain..."
    
    CORE_CODE_ID=$(store_contract "$CORE_WASM" "Identity DAO Core")
    VOTING_CODE_ID=$(store_contract "$VOTING_WASM" "DID-Based Voting")
    PROPOSALS_CODE_ID=$(store_contract "$PROPOSALS_WASM" "Identity Proposals")
    PRE_PROPOSE_CODE_ID=$(store_contract "$PRE_PROPOSE_WASM" "Pre-Propose Identity")
    
    # Save code IDs
    cat > "${CONTRACTS_DIR}/code_ids.json" <<EOF
{
    "core": $CORE_CODE_ID,
    "voting": $VOTING_CODE_ID,
    "proposals": $PROPOSALS_CODE_ID,
    "pre_propose": $PRE_PROPOSE_CODE_ID
}
EOF
    
    log_info "Contract codes stored. Code IDs saved to code_ids.json"
    
    # Instantiate Core Module first
    log_info "Instantiating Identity DAO Core Module..."
    
    CORE_INIT_MSG=$(cat <<EOF
{
    "admin": "$DEPLOYER",
    "dao_name": "Sonr Identity DAO",
    "dao_uri": "https://sonr.io/dao",
    "voting_module": null,
    "proposal_modules": []
}
EOF
)
    
    CORE_ADDR=$(instantiate_contract "$CORE_CODE_ID" "$CORE_INIT_MSG" "identity-dao-core" "$DEPLOYER")
    
    # Instantiate Voting Module
    log_info "Instantiating DID-Based Voting Module..."
    
    VOTING_INIT_MSG=$(cat <<EOF
{
    "dao_address": "$CORE_ADDR",
    "min_verification_level": 1,
    "voting_period": 604800,
    "quorum_percentage": 20,
    "threshold_percentage": 51
}
EOF
)
    
    VOTING_ADDR=$(instantiate_contract "$VOTING_CODE_ID" "$VOTING_INIT_MSG" "did-voting" "$CORE_ADDR")
    
    # Instantiate Pre-Propose Module
    log_info "Instantiating Pre-Propose Identity Module..."
    
    PRE_PROPOSE_INIT_MSG=$(cat <<EOF
{
    "proposal_module": null,
    "min_verification_status": "Basic",
    "deposit_amount": "1000000",
    "deposit_denom": "usnr"
}
EOF
)
    
    PRE_PROPOSE_ADDR=$(instantiate_contract "$PRE_PROPOSE_CODE_ID" "$PRE_PROPOSE_INIT_MSG" "pre-propose-identity" "$CORE_ADDR")
    
    # Instantiate Proposals Module
    log_info "Instantiating Identity Proposals Module..."
    
    PROPOSALS_INIT_MSG=$(cat <<EOF
{
    "dao_address": "$CORE_ADDR",
    "voting_module": "$VOTING_ADDR",
    "pre_propose_module": "$PRE_PROPOSE_ADDR",
    "proposal_duration": 604800,
    "min_verification_level": 1
}
EOF
)
    
    PROPOSALS_ADDR=$(instantiate_contract "$PROPOSALS_CODE_ID" "$PROPOSALS_INIT_MSG" "identity-proposals" "$CORE_ADDR")
    
    # Update Core Module with voting and proposal modules
    log_info "Updating Core Module configuration..."
    
    UPDATE_MSG=$(cat <<EOF
{
    "update_config": {
        "voting_module": "$VOTING_ADDR",
        "proposal_modules": ["$PROPOSALS_ADDR"]
    }
}
EOF
)
    
    snrd tx wasm execute "$CORE_ADDR" "$UPDATE_MSG" \
        --from "$DEPLOYER" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --gas-prices "$GAS_PRICES" \
        --gas-adjustment "$GAS_ADJUSTMENT" \
        --keyring-backend "$KEYRING" \
        --yes
    
    # Save deployment addresses
    cat > "${CONTRACTS_DIR}/addresses.json" <<EOF
{
    "core": "$CORE_ADDR",
    "voting": "$VOTING_ADDR",
    "proposals": "$PROPOSALS_ADDR",
    "pre_propose": "$PRE_PROPOSE_ADDR",
    "deployer": "$DEPLOYER",
    "chain_id": "$CHAIN_ID",
    "deployment_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    log_info "Deployment complete! Contract addresses saved to addresses.json"
    log_info ""
    log_info "Contract Addresses:"
    log_info "  Core:        $CORE_ADDR"
    log_info "  Voting:      $VOTING_ADDR"
    log_info "  Proposals:   $PROPOSALS_ADDR"
    log_info "  Pre-Propose: $PRE_PROPOSE_ADDR"
}

# Run main function
main "$@"