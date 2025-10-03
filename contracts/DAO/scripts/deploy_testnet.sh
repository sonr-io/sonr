#!/bin/bash

# Deploy Identity DAO Contracts to Cosmos Hub Testnet
# This script handles the deployment of all DAO contracts and IBC setup

set -e

# Configuration
CHAIN_ID="theta-testnet-001"  # Cosmos Hub testnet chain ID
NODE="https://rpc.sentry-01.theta-testnet.polypore.xyz"
GAS_PRICES="0.025uatom"
GAS_AUTO="--gas auto --gas-adjustment 1.3"
KEYRING="--keyring-backend test"

# Contract paths
CONTRACTS_DIR="$(dirname "$0")/../target/wasm32-unknown-unknown/release"
CORE_WASM="${CONTRACTS_DIR}/identity_dao_core.wasm"
VOTING_WASM="${CONTRACTS_DIR}/identity_dao_voting.wasm"
PROPOSALS_WASM="${CONTRACTS_DIR}/identity_dao_proposals.wasm"
PRE_PROPOSE_WASM="${CONTRACTS_DIR}/identity_dao_pre_propose.wasm"

# Sonr chain configuration for IBC
SONR_CHAIN_ID="sonrtest_1-1"
SONR_NODE="http://localhost:26657"
IBC_VERSION="ics20-1"

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

# Check if gaiad is installed
check_gaiad() {
    if ! command -v gaiad &> /dev/null; then
        log_error "gaiad is not installed. Please install Cosmos Hub client."
        exit 1
    fi
    log_info "Found gaiad: $(gaiad version)"
}

# Check if contracts are built
check_contracts() {
    log_info "Checking for compiled contracts..."
    
    if [ ! -f "$CORE_WASM" ]; then
        log_error "Core contract not found at $CORE_WASM"
        log_info "Building contracts..."
        cd "$(dirname "$0")/.."
        cargo build --release --target wasm32-unknown-unknown
    fi
    
    log_info "All contracts found"
}

# Optimize contracts for deployment
optimize_contracts() {
    log_info "Optimizing contracts for deployment..."
    
    # Use CosmWasm optimizer
    docker run --rm -v "$(pwd)":/code \
        --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
        --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
        cosmwasm/optimizer:0.16.0
    
    log_info "Contract optimization complete"
}

# Upload contract to chain
upload_contract() {
    local wasm_file=$1
    local contract_name=$2
    
    log_info "Uploading ${contract_name} contract..."
    
    TX_HASH=$(gaiad tx wasm store "${wasm_file}" \
        --from deployer \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        --broadcast-mode sync \
        --output json \
        -y | jq -r '.txhash')
    
    log_info "Transaction submitted: ${TX_HASH}"
    sleep 6
    
    # Get code ID from transaction
    CODE_ID=$(gaiad query tx "${TX_HASH}" \
        --node "${NODE}" \
        --output json | jq -r '.logs[0].events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
    
    log_info "${contract_name} uploaded with code ID: ${CODE_ID}"
    echo "${CODE_ID}"
}

# Instantiate contract
instantiate_contract() {
    local code_id=$1
    local init_msg=$2
    local label=$3
    local admin=$4
    
    log_info "Instantiating ${label}..."
    
    TX_HASH=$(gaiad tx wasm instantiate "${code_id}" "${init_msg}" \
        --from deployer \
        --label "${label}" \
        --admin "${admin}" \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        --broadcast-mode sync \
        --output json \
        -y | jq -r '.txhash')
    
    log_info "Transaction submitted: ${TX_HASH}"
    sleep 6
    
    # Get contract address from transaction
    CONTRACT_ADDR=$(gaiad query tx "${TX_HASH}" \
        --node "${NODE}" \
        --output json | jq -r '.logs[0].events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
    
    log_info "${label} instantiated at: ${CONTRACT_ADDR}"
    echo "${CONTRACT_ADDR}"
}

# Setup IBC channel
setup_ibc_channel() {
    local contract_addr=$1
    local port=$2
    
    log_info "Setting up IBC channel for ${port}..."
    
    # Create client for Sonr chain
    gaiad tx ibc client create \
        --chain-id "${CHAIN_ID}" \
        --from deployer \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    sleep 6
    
    # Create connection
    gaiad tx ibc connection open-init \
        --chain-id "${CHAIN_ID}" \
        --from deployer \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    sleep 6
    
    # Create channel
    gaiad tx ibc channel open-init \
        --port "${port}" \
        --version "${IBC_VERSION}" \
        --chain-id "${CHAIN_ID}" \
        --from deployer \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    log_info "IBC channel setup initiated for ${port}"
}

# Main deployment flow
main() {
    log_info "Starting Identity DAO deployment to Cosmos Hub testnet..."
    
    # Prerequisites
    check_gaiad
    check_contracts
    
    # Get deployer address
    DEPLOYER_ADDR=$(gaiad keys show deployer -a ${KEYRING})
    log_info "Deployer address: ${DEPLOYER_ADDR}"
    
    # Check balance
    BALANCE=$(gaiad query bank balances "${DEPLOYER_ADDR}" \
        --node "${NODE}" \
        --output json | jq -r '.balances[] | select(.denom=="uatom") | .amount')
    
    if [ -z "$BALANCE" ] || [ "$BALANCE" -eq "0" ]; then
        log_error "Deployer has no ATOM tokens. Please fund the account."
        log_info "Visit https://discord.com/channels/669268347736686612/953641721746206780 for testnet faucet"
        exit 1
    fi
    
    log_info "Deployer balance: ${BALANCE} uatom"
    
    # Optimize contracts
    optimize_contracts
    
    # Upload contracts
    log_info "Uploading contracts to chain..."
    CORE_CODE_ID=$(upload_contract "${CONTRACTS_DIR}/identity_dao_core-optimized.wasm" "Identity DAO Core")
    VOTING_CODE_ID=$(upload_contract "${CONTRACTS_DIR}/identity_dao_voting-optimized.wasm" "DID Voting")
    PROPOSALS_CODE_ID=$(upload_contract "${CONTRACTS_DIR}/identity_dao_proposals-optimized.wasm" "Proposals")
    PRE_PROPOSE_CODE_ID=$(upload_contract "${CONTRACTS_DIR}/identity_dao_pre_propose-optimized.wasm" "Pre-Propose")
    
    # Save code IDs
    echo "CORE_CODE_ID=${CORE_CODE_ID}" > deployment_ids.env
    echo "VOTING_CODE_ID=${VOTING_CODE_ID}" >> deployment_ids.env
    echo "PROPOSALS_CODE_ID=${PROPOSALS_CODE_ID}" >> deployment_ids.env
    echo "PRE_PROPOSE_CODE_ID=${PRE_PROPOSE_CODE_ID}" >> deployment_ids.env
    
    # Instantiate Core contract
    CORE_INIT_MSG='{
        "name": "Sonr Identity DAO",
        "description": "Decentralized Identity Governance on Cosmos Hub",
        "voting_config": {
            "threshold": "0.51",
            "quorum": "0.1",
            "voting_period": 604800,
            "proposal_deposit": "1000000"
        },
        "admin": "'${DEPLOYER_ADDR}'",
        "enable_did_integration": true
    }'
    
    CORE_ADDR=$(instantiate_contract "${CORE_CODE_ID}" "${CORE_INIT_MSG}" "sonr-identity-dao-core" "${DEPLOYER_ADDR}")
    echo "CORE_ADDR=${CORE_ADDR}" >> deployment_ids.env
    
    # Instantiate Voting contract
    VOTING_INIT_MSG='{
        "dao_core": "'${CORE_ADDR}'",
        "min_verification_level": 1,
        "use_reputation_weight": true
    }'
    
    VOTING_ADDR=$(instantiate_contract "${VOTING_CODE_ID}" "${VOTING_INIT_MSG}" "sonr-did-voting" "${CORE_ADDR}")
    echo "VOTING_ADDR=${VOTING_ADDR}" >> deployment_ids.env
    
    # Instantiate Proposals contract
    PROPOSALS_INIT_MSG='{
        "dao_core": "'${CORE_ADDR}'",
        "voting_module": "'${VOTING_ADDR}'",
        "pre_propose_module": null,
        "proposal_deposit": "1000000",
        "max_voting_period": 604800
    }'
    
    PROPOSALS_ADDR=$(instantiate_contract "${PROPOSALS_CODE_ID}" "${PROPOSALS_INIT_MSG}" "sonr-proposals" "${CORE_ADDR}")
    echo "PROPOSALS_ADDR=${PROPOSALS_ADDR}" >> deployment_ids.env
    
    # Instantiate Pre-Propose contract
    PRE_PROPOSE_INIT_MSG='{
        "dao_core": "'${CORE_ADDR}'",
        "proposal_module": "'${PROPOSALS_ADDR}'",
        "require_verified_did": true,
        "min_reputation_score": 10,
        "deposit_amount": "1000000",
        "deposit_denom": "uatom"
    }'
    
    PRE_PROPOSE_ADDR=$(instantiate_contract "${PRE_PROPOSE_CODE_ID}" "${PRE_PROPOSE_INIT_MSG}" "sonr-pre-propose" "${CORE_ADDR}")
    echo "PRE_PROPOSE_ADDR=${PRE_PROPOSE_ADDR}" >> deployment_ids.env
    
    # Register modules with Core
    log_info "Registering modules with Core contract..."
    
    REGISTER_VOTING_MSG='{"register_module":{"module_type":"voting","module_address":"'${VOTING_ADDR}'"}}'
    gaiad tx wasm execute "${CORE_ADDR}" "${REGISTER_VOTING_MSG}" \
        --from deployer \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    sleep 6
    
    REGISTER_PROPOSALS_MSG='{"register_module":{"module_type":"proposal","module_address":"'${PROPOSALS_ADDR}'"}}'
    gaiad tx wasm execute "${CORE_ADDR}" "${REGISTER_PROPOSALS_MSG}" \
        --from deployer \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    sleep 6
    
    REGISTER_PRE_PROPOSE_MSG='{"register_module":{"module_type":"pre_propose","module_address":"'${PRE_PROPOSE_ADDR}'"}}'
    gaiad tx wasm execute "${CORE_ADDR}" "${REGISTER_PRE_PROPOSE_MSG}" \
        --from deployer \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    sleep 6
    
    # Setup IBC channels for Sonr integration
    log_info "Setting up IBC channels for Sonr integration..."
    setup_ibc_channel "${VOTING_ADDR}" "wasm.${VOTING_ADDR}"
    
    # Update Pre-Propose module in Proposals contract
    UPDATE_PRE_PROPOSE_MSG='{"update_pre_propose_module":{"module":"'${PRE_PROPOSE_ADDR}'"}}'
    gaiad tx wasm execute "${PROPOSALS_ADDR}" "${UPDATE_PRE_PROPOSE_MSG}" \
        --from deployer \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        -y
    
    log_info "✅ Deployment complete!"
    log_info "Deployment details saved to deployment_ids.env"
    
    # Display summary
    echo ""
    log_info "=== Deployment Summary ==="
    echo "Core Contract: ${CORE_ADDR}"
    echo "Voting Contract: ${VOTING_ADDR}"
    echo "Proposals Contract: ${PROPOSALS_ADDR}"
    echo "Pre-Propose Contract: ${PRE_PROPOSE_ADDR}"
    echo ""
    log_info "Next steps:"
    echo "1. Verify contract deployment: ./scripts/verify_deployment.sh"
    echo "2. Test IBC connectivity: ./scripts/test_ibc.sh"
    echo "3. Create first proposal: ./scripts/create_proposal.sh"
}

# Run main deployment
main "$@"