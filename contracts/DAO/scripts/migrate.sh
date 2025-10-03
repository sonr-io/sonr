#!/bin/bash
# Identity DAO Migration Script
# Migrates Identity DAO contracts to new versions

set -e

# Configuration
CHAIN_ID="${CHAIN_ID:-sonrtest_1-1}"
NODE="${NODE:-http://localhost:26657}"
KEYRING="${KEYRING:-test}"
ADMIN="${ADMIN:-deployer}"
GAS_PRICES="${GAS_PRICES:-0.025usnr}"
GAS_ADJUSTMENT="${GAS_ADJUSTMENT:-1.5}"

# Contract paths
CONTRACTS_DIR="$(dirname "$0")/../artifacts"
ADDRESSES_FILE="${CONTRACTS_DIR}/addresses.json"

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
    
    if [ ! -f "$ADDRESSES_FILE" ]; then
        log_error "Contract addresses file not found. Please deploy first."
        exit 1
    fi
    
    log_info "All dependencies satisfied"
}

# Store new contract code
store_new_code() {
    local wasm_file=$1
    local contract_name=$2
    
    log_info "Storing new $contract_name contract code..."
    
    if [ ! -f "$wasm_file" ]; then
        log_error "Contract file not found: $wasm_file"
        exit 1
    fi
    
    local tx_result=$(snrd tx wasm store "$wasm_file" \
        --from "$ADMIN" \
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
    
    log_info "$contract_name new code stored with ID: $code_id"
    echo "$code_id"
}

# Migrate contract
migrate_contract() {
    local contract_addr=$1
    local new_code_id=$2
    local migrate_msg=$3
    local contract_name=$4
    
    log_info "Migrating $contract_name contract..."
    
    local tx_result=$(snrd tx wasm migrate "$contract_addr" "$new_code_id" "$migrate_msg" \
        --from "$ADMIN" \
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
    
    # Check migration success
    local tx_query=$(snrd query tx "$tx_hash" \
        --node "$NODE" \
        --output json)
    
    local code=$(echo "$tx_query" | jq -r .code)
    
    if [ "$code" != "0" ]; then
        log_error "Migration failed for $contract_name"
        echo "$tx_query" | jq -r .raw_log
        exit 1
    fi
    
    log_info "$contract_name migrated successfully"
}

# Main migration flow
main() {
    local MODULE=$1
    
    log_info "Starting Identity DAO migration..."
    
    # Check dependencies
    check_dependencies
    
    # Load contract addresses
    CORE_ADDR=$(jq -r .core "$ADDRESSES_FILE")
    VOTING_ADDR=$(jq -r .voting "$ADDRESSES_FILE")
    PROPOSALS_ADDR=$(jq -r .proposals "$ADDRESSES_FILE")
    PRE_PROPOSE_ADDR=$(jq -r .pre_propose "$ADDRESSES_FILE")
    
    log_info "Loaded contract addresses from deployment"
    
    # Migrate specific module or all
    case "$MODULE" in
        core)
            log_info "Migrating Core Module..."
            NEW_CODE_ID=$(store_new_code "${CONTRACTS_DIR}/identity_dao_core.wasm" "Core")
            MIGRATE_MSG='{"update_version":{}}'
            migrate_contract "$CORE_ADDR" "$NEW_CODE_ID" "$MIGRATE_MSG" "Core"
            ;;
        voting)
            log_info "Migrating Voting Module..."
            NEW_CODE_ID=$(store_new_code "${CONTRACTS_DIR}/identity_dao_voting.wasm" "Voting")
            MIGRATE_MSG='{"update_version":{}}'
            migrate_contract "$VOTING_ADDR" "$NEW_CODE_ID" "$MIGRATE_MSG" "Voting"
            ;;
        proposals)
            log_info "Migrating Proposals Module..."
            NEW_CODE_ID=$(store_new_code "${CONTRACTS_DIR}/identity_dao_proposals.wasm" "Proposals")
            MIGRATE_MSG='{"update_version":{}}'
            migrate_contract "$PROPOSALS_ADDR" "$NEW_CODE_ID" "$MIGRATE_MSG" "Proposals"
            ;;
        pre-propose)
            log_info "Migrating Pre-Propose Module..."
            NEW_CODE_ID=$(store_new_code "${CONTRACTS_DIR}/identity_dao_pre_propose.wasm" "Pre-Propose")
            MIGRATE_MSG='{"update_version":{}}'
            migrate_contract "$PRE_PROPOSE_ADDR" "$NEW_CODE_ID" "$MIGRATE_MSG" "Pre-Propose"
            ;;
        all)
            log_info "Migrating all modules..."
            
            # Store all new codes first
            CORE_NEW_CODE=$(store_new_code "${CONTRACTS_DIR}/identity_dao_core.wasm" "Core")
            VOTING_NEW_CODE=$(store_new_code "${CONTRACTS_DIR}/identity_dao_voting.wasm" "Voting")
            PROPOSALS_NEW_CODE=$(store_new_code "${CONTRACTS_DIR}/identity_dao_proposals.wasm" "Proposals")
            PRE_PROPOSE_NEW_CODE=$(store_new_code "${CONTRACTS_DIR}/identity_dao_pre_propose.wasm" "Pre-Propose")
            
            # Migrate in order
            MIGRATE_MSG='{"update_version":{}}'
            migrate_contract "$CORE_ADDR" "$CORE_NEW_CODE" "$MIGRATE_MSG" "Core"
            migrate_contract "$VOTING_ADDR" "$VOTING_NEW_CODE" "$MIGRATE_MSG" "Voting"
            migrate_contract "$PRE_PROPOSE_ADDR" "$PRE_PROPOSE_NEW_CODE" "$MIGRATE_MSG" "Pre-Propose"
            migrate_contract "$PROPOSALS_ADDR" "$PROPOSALS_NEW_CODE" "$MIGRATE_MSG" "Proposals"
            ;;
        *)
            log_error "Invalid module: $MODULE"
            log_info "Usage: $0 [core|voting|proposals|pre-propose|all]"
            exit 1
            ;;
    esac
    
    # Save migration info
    cat > "${CONTRACTS_DIR}/migration_$(date +%Y%m%d_%H%M%S).json" <<EOF
{
    "module": "$MODULE",
    "migrated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "admin": "$ADMIN",
    "chain_id": "$CHAIN_ID"
}
EOF
    
    log_info "Migration complete!"
}

# Run main function
main "$@"