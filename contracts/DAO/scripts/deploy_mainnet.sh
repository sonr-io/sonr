#!/bin/bash

# Deploy Identity DAO Contracts to Cosmos Hub Mainnet
# PRODUCTION DEPLOYMENT - USE WITH CAUTION

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
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

log_critical() {
    echo -e "${MAGENTA}[CRITICAL]${NC} $1"
}

# Configuration
CHAIN_ID="cosmoshub-4"  # Cosmos Hub mainnet chain ID
NODE="https://cosmos-rpc.polkachu.com:443"
BACKUP_NODE="https://rpc-cosmoshub.blockapsis.com:443"
GAS_PRICES="0.025uatom"
GAS_AUTO="--gas auto --gas-adjustment 1.5"
KEYRING="--keyring-backend file"  # Use file backend for mainnet

# Contract paths
CONTRACTS_DIR="$(dirname "$0")/../artifacts"
CORE_WASM="${CONTRACTS_DIR}/identity_dao_core.wasm"
VOTING_WASM="${CONTRACTS_DIR}/identity_dao_voting.wasm"
PROPOSALS_WASM="${CONTRACTS_DIR}/identity_dao_proposals.wasm"
PRE_PROPOSE_WASM="${CONTRACTS_DIR}/identity_dao_pre_propose.wasm"

# Sonr mainnet configuration for IBC
SONR_CHAIN_ID="sonr-1"  # Sonr mainnet chain ID
SONR_NODE="https://rpc.sonr.io:443"
IBC_VERSION="ics20-1"

# Security checks
MAINNET_CONFIRMATION="I_UNDERSTAND_THIS_IS_MAINNET_DEPLOYMENT"
MULTISIG_THRESHOLD=3
REQUIRED_SIGNATURES=2

# Deployment configuration
MIN_BALANCE_ATOM=50  # Minimum ATOM balance required
PROPOSAL_DEPOSIT="10000000"  # 10 ATOM proposal deposit
VOTING_PERIOD=1209600  # 14 days in seconds
QUORUM="0.334"  # 33.4% quorum
THRESHOLD="0.5"  # 50% threshold

# Pre-deployment checks
pre_deployment_checks() {
    log_critical "=== MAINNET DEPLOYMENT PRE-CHECKS ==="
    
    # Confirm mainnet deployment
    echo -e "${RED}WARNING: You are about to deploy to Cosmos Hub MAINNET${NC}"
    echo -e "${RED}This is a production deployment that will use real ATOM tokens${NC}"
    echo ""
    read -p "Type '${MAINNET_CONFIRMATION}' to continue: " confirmation
    
    if [ "$confirmation" != "$MAINNET_CONFIRMATION" ]; then
        log_error "Mainnet deployment cancelled"
        exit 1
    fi
    
    log_info "Mainnet deployment confirmed"
}

# Check multisig setup
check_multisig() {
    log_info "Checking multisig configuration..."
    
    # Check if multisig account exists
    MULTISIG_NAME="identity-dao-multisig"
    
    if ! gaiad keys show "${MULTISIG_NAME}" ${KEYRING} &> /dev/null; then
        log_error "Multisig account not found. Please create it first:"
        echo "gaiad keys add ${MULTISIG_NAME} --multisig key1,key2,key3 --multisig-threshold ${REQUIRED_SIGNATURES}"
        exit 1
    fi
    
    MULTISIG_ADDR=$(gaiad keys show "${MULTISIG_NAME}" -a ${KEYRING})
    log_info "Multisig address: ${MULTISIG_ADDR}"
    
    # Check multisig balance
    BALANCE=$(gaiad query bank balances "${MULTISIG_ADDR}" \
        --node "${NODE}" \
        --output json | jq -r '.balances[] | select(.denom=="uatom") | .amount')
    
    BALANCE_ATOM=$((BALANCE / 1000000))
    
    if [ "$BALANCE_ATOM" -lt "$MIN_BALANCE_ATOM" ]; then
        log_error "Insufficient balance: ${BALANCE_ATOM} ATOM (required: ${MIN_BALANCE_ATOM} ATOM)"
        exit 1
    fi
    
    log_info "Multisig balance: ${BALANCE_ATOM} ATOM"
}

# Verify contracts
verify_contracts() {
    log_info "Verifying contract checksums..."
    
    if [ ! -f "${CONTRACTS_DIR}/checksums.txt" ]; then
        log_error "Checksums file not found"
        exit 1
    fi
    
    # Verify each contract
    cd "${CONTRACTS_DIR}"
    
    if command -v sha256sum &> /dev/null; then
        sha256sum -c checksums.txt
    else
        shasum -a 256 -c checksums.txt
    fi
    
    if [ $? -eq 0 ]; then
        log_info "All contract checksums verified"
    else
        log_error "Contract checksum verification failed"
        exit 1
    fi
    
    cd - > /dev/null
}

# Backup current state
create_backup() {
    log_info "Creating deployment backup..."
    
    BACKUP_DIR="backups/mainnet_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "${BACKUP_DIR}"
    
    # Copy contracts
    cp -r "${CONTRACTS_DIR}" "${BACKUP_DIR}/"
    
    # Save deployment configuration
    cat > "${BACKUP_DIR}/deployment_config.json" << EOF
{
    "chain_id": "${CHAIN_ID}",
    "node": "${NODE}",
    "multisig_addr": "${MULTISIG_ADDR}",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "contracts": {
        "core": "${CORE_WASM}",
        "voting": "${VOTING_WASM}",
        "proposals": "${PROPOSALS_WASM}",
        "pre_propose": "${PRE_PROPOSE_WASM}"
    },
    "config": {
        "proposal_deposit": "${PROPOSAL_DEPOSIT}",
        "voting_period": ${VOTING_PERIOD},
        "quorum": "${QUORUM}",
        "threshold": "${THRESHOLD}"
    }
}
EOF
    
    log_info "Backup created at ${BACKUP_DIR}"
}

# Generate multisig transaction
generate_multisig_tx() {
    local msg=$1
    local output_file=$2
    local description=$3
    
    log_info "Generating multisig transaction: ${description}"
    
    # Generate unsigned transaction
    gaiad tx wasm $msg \
        --from "${MULTISIG_NAME}" \
        --chain-id "${CHAIN_ID}" \
        --node "${NODE}" \
        --gas-prices "${GAS_PRICES}" \
        ${GAS_AUTO} \
        ${KEYRING} \
        --generate-only > "${output_file}"
    
    log_info "Unsigned transaction saved to ${output_file}"
}

# Upload contract with multisig
upload_contract_multisig() {
    local wasm_file=$1
    local contract_name=$2
    
    log_info "Preparing ${contract_name} upload..."
    
    # Generate store transaction
    TX_FILE="tx_store_${contract_name}.json"
    
    generate_multisig_tx \
        "store ${wasm_file}" \
        "${TX_FILE}" \
        "Store ${contract_name}"
    
    log_info "${contract_name} upload transaction prepared"
    log_warning "Requires multisig signatures before broadcasting"
    
    echo "${TX_FILE}"
}

# Main deployment flow
main() {
    log_critical "Starting Identity DAO MAINNET deployment..."
    
    # Pre-deployment checks
    pre_deployment_checks
    
    # Check multisig setup
    check_multisig
    
    # Verify contracts
    verify_contracts
    
    # Create backup
    create_backup
    
    # Prepare upload transactions
    log_info "Preparing contract upload transactions..."
    
    CORE_TX=$(upload_contract_multisig "${CORE_WASM}" "Identity DAO Core")
    VOTING_TX=$(upload_contract_multisig "${VOTING_WASM}" "DID Voting")
    PROPOSALS_TX=$(upload_contract_multisig "${PROPOSALS_WASM}" "Proposals")
    PRE_PROPOSE_TX=$(upload_contract_multisig "${PRE_PROPOSE_WASM}" "Pre-Propose")
    
    # Generate instantiation messages
    log_info "Generating instantiation messages..."
    
    CORE_INIT_MSG='{
        "name": "Sonr Identity DAO",
        "description": "Decentralized Identity Governance on Cosmos Hub",
        "voting_config": {
            "threshold": "'${THRESHOLD}'",
            "quorum": "'${QUORUM}'",
            "voting_period": '${VOTING_PERIOD}',
            "proposal_deposit": "'${PROPOSAL_DEPOSIT}'"
        },
        "admin": "'${MULTISIG_ADDR}'",
        "enable_did_integration": true
    }'
    
    echo "$CORE_INIT_MSG" > init_core.json
    
    # Save deployment instructions
    cat > MAINNET_DEPLOYMENT_INSTRUCTIONS.md << EOF
# Cosmos Hub Mainnet Deployment Instructions

## Prerequisites
- Multisig address: ${MULTISIG_ADDR}
- Required signatures: ${REQUIRED_SIGNATURES} of ${MULTISIG_THRESHOLD}
- Chain ID: ${CHAIN_ID}

## Step 1: Sign Upload Transactions

Each signer must sign the upload transactions:

\`\`\`bash
# Sign Core contract upload
gaiad tx sign ${CORE_TX} --from <signer_key> --chain-id ${CHAIN_ID} ${KEYRING} > signed_core_<signer>.json

# Sign Voting contract upload
gaiad tx sign ${VOTING_TX} --from <signer_key> --chain-id ${CHAIN_ID} ${KEYRING} > signed_voting_<signer>.json

# Sign Proposals contract upload
gaiad tx sign ${PROPOSALS_TX} --from <signer_key> --chain-id ${CHAIN_ID} ${KEYRING} > signed_proposals_<signer>.json

# Sign Pre-Propose contract upload
gaiad tx sign ${PRE_PROPOSE_TX} --from <signer_key> --chain-id ${CHAIN_ID} ${KEYRING} > signed_pre_propose_<signer>.json
\`\`\`

## Step 2: Combine Signatures

\`\`\`bash
# Combine Core signatures
gaiad tx multisign ${CORE_TX} ${MULTISIG_NAME} signed_core_*.json --chain-id ${CHAIN_ID} ${KEYRING} > tx_core_signed.json

# Combine Voting signatures
gaiad tx multisign ${VOTING_TX} ${MULTISIG_NAME} signed_voting_*.json --chain-id ${CHAIN_ID} ${KEYRING} > tx_voting_signed.json

# Combine Proposals signatures
gaiad tx multisign ${PROPOSALS_TX} ${MULTISIG_NAME} signed_proposals_*.json --chain-id ${CHAIN_ID} ${KEYRING} > tx_proposals_signed.json

# Combine Pre-Propose signatures
gaiad tx multisign ${PRE_PROPOSE_TX} ${MULTISIG_NAME} signed_pre_propose_*.json --chain-id ${CHAIN_ID} ${KEYRING} > tx_pre_propose_signed.json
\`\`\`

## Step 3: Broadcast Transactions

\`\`\`bash
# Broadcast uploads (one at a time)
gaiad tx broadcast tx_core_signed.json --node ${NODE}
# Wait for confirmation and note CODE_ID

gaiad tx broadcast tx_voting_signed.json --node ${NODE}
# Wait for confirmation and note CODE_ID

gaiad tx broadcast tx_proposals_signed.json --node ${NODE}
# Wait for confirmation and note CODE_ID

gaiad tx broadcast tx_pre_propose_signed.json --node ${NODE}
# Wait for confirmation and note CODE_ID
\`\`\`

## Step 4: Instantiate Contracts

After obtaining code IDs, instantiate each contract following the same multisig process.

## Step 5: Verify Deployment

Run verification script:
\`\`\`bash
./scripts/verify_deployment.sh
\`\`\`

## Security Checklist

- [ ] All signers have verified contract checksums
- [ ] Multisig threshold is correctly configured
- [ ] Admin keys are securely stored
- [ ] Backup of deployment configuration created
- [ ] IBC channels will be established post-deployment
- [ ] Emergency procedures documented

## Emergency Contacts

- Technical Lead: [Contact]
- Security Team: [Contact]
- Multisig Signers: [List]

---
Generated: $(date)
EOF
    
    log_info "✅ Mainnet deployment preparation complete!"
    log_info ""
    log_critical "=== IMPORTANT NEXT STEPS ==="
    echo "1. Review MAINNET_DEPLOYMENT_INSTRUCTIONS.md"
    echo "2. Coordinate with multisig signers"
    echo "3. Execute deployment following the instructions"
    echo "4. Verify deployment using verify_deployment.sh"
    echo "5. Establish IBC channels to Sonr mainnet"
    echo ""
    log_warning "All unsigned transactions saved to current directory"
    log_warning "DO NOT share private keys or signed transactions insecurely"
}

# Run main deployment
main "$@"