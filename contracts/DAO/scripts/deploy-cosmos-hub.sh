#!/bin/bash

# Identity DAO Deployment Script for Cosmos Hub
# Deploys contracts to Cosmos Hub and establishes IBC channels to Sonr

set -e

# Configuration
COSMOS_HUB_CHAIN_ID="${COSMOS_HUB_CHAIN_ID:-cosmoshub-testnet}"
COSMOS_HUB_NODE="${COSMOS_HUB_NODE:-https://rpc.testnet.cosmos.network:443}"
SONR_CHAIN_ID="${SONR_CHAIN_ID:-sonrtest_1-1}"
SONR_NODE="${SONR_NODE:-http://localhost:26657}"
DEPLOYER="${DEPLOYER:-deployer}"
GAS_PRICES="${GAS_PRICES:-0.025uatom}"
CONTRACTS_DIR="../artifacts"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Identity DAO Deployment to Cosmos Hub${NC}"
echo -e "${BLUE}========================================${NC}"

# Function to store contract code
store_contract() {
    local wasm_file=$1
    local label=$2

    echo -e "${GREEN}Storing contract: ${label}${NC}"

    TX_HASH=$(gaiad tx wasm store "$wasm_file" \
        --from "$DEPLOYER" \
        --chain-id "$COSMOS_HUB_CHAIN_ID" \
        --node "$COSMOS_HUB_NODE" \
        --gas-prices "$GAS_PRICES" \
        --gas auto \
        --gas-adjustment 1.5 \
        --broadcast-mode sync \
        --yes \
        --output json | jq -r '.txhash')

    echo "Waiting for transaction..."
    sleep 6

    CODE_ID=$(gaiad query tx "$TX_HASH" \
        --node "$COSMOS_HUB_NODE" \
        --output json | jq -r '.logs[0].events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

    echo -e "${GREEN}✓ Stored ${label} with code ID: ${CODE_ID}${NC}"
    echo "$CODE_ID"
}

# Function to instantiate contract
instantiate_contract() {
    local code_id=$1
    local init_msg=$2
    local label=$3

    echo -e "${GREEN}Instantiating: ${label}${NC}"

    TX_HASH=$(gaiad tx wasm instantiate "$code_id" "$init_msg" \
        --from "$DEPLOYER" \
        --label "$label" \
        --chain-id "$COSMOS_HUB_CHAIN_ID" \
        --node "$COSMOS_HUB_NODE" \
        --gas-prices "$GAS_PRICES" \
        --gas auto \
        --gas-adjustment 1.5 \
        --admin "$DEPLOYER" \
        --broadcast-mode sync \
        --yes \
        --output json | jq -r '.txhash')

    echo "Waiting for transaction..."
    sleep 6

    CONTRACT_ADDR=$(gaiad query tx "$TX_HASH" \
        --node "$COSMOS_HUB_NODE" \
        --output json | jq -r '.logs[0].events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

    echo -e "${GREEN}✓ Instantiated ${label} at: ${CONTRACT_ADDR}${NC}"
    echo "$CONTRACT_ADDR"
}

# Check prerequisites
echo -e "${BLUE}Checking prerequisites...${NC}"

if ! command -v gaiad &>/dev/null; then
    echo -e "${RED}Error: gaiad not found. Please install Gaia (Cosmos Hub client)${NC}"
    exit 1
fi

if ! command -v hermes &>/dev/null; then
    echo -e "${RED}Error: hermes not found. Please install Hermes IBC relayer${NC}"
    exit 1
fi

# Build contracts if needed
if [ ! -d "$CONTRACTS_DIR" ]; then
    echo -e "${BLUE}Building contracts...${NC}"
    cd ..
    docker run --rm -v "$(pwd)":/code \
        --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
        --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
        cosmwasm/workspace-optimizer:0.13.0
    cd scripts
fi

# Store contract codes
echo -e "${BLUE}Storing contract codes on Cosmos Hub...${NC}"

CORE_CODE_ID=$(store_contract "$CONTRACTS_DIR/identity_dao_core.wasm" "Identity DAO Core")
VOTING_CODE_ID=$(store_contract "$CONTRACTS_DIR/did_voting.wasm" "DID-Based Voting")
PROPOSALS_CODE_ID=$(store_contract "$CONTRACTS_DIR/identity_proposals.wasm" "Identity Proposals")
PRE_PROPOSE_CODE_ID=$(store_contract "$CONTRACTS_DIR/pre_propose_identity.wasm" "Pre-Propose Identity")

# Instantiate contracts
echo -e "${BLUE}Instantiating contracts...${NC}"

# Core module
CORE_INIT='{
    "admin": "'$DEPLOYER'",
    "dao_name": "Sonr Identity DAO",
    "dao_uri": "https://sonr.io/dao",
    "voting_module": null,
    "proposal_modules": [],
    "wyoming_dao_info": {
        "entity_name": "Sonr Identity DAO LLC",
        "entity_type": "LLC",
        "registered_agent": "Wyoming Registered Agent LLC",
        "ein": "00-0000000"
    }
}'
CORE_ADDR=$(instantiate_contract "$CORE_CODE_ID" "$CORE_INIT" "identity-dao-core")

# Voting module
VOTING_INIT='{
    "dao_core": "'$CORE_ADDR'",
    "min_verification_level": 1,
    "use_reputation_weight": true
}'
VOTING_ADDR=$(instantiate_contract "$VOTING_CODE_ID" "$VOTING_INIT" "did-voting")

# Proposals module
PROPOSALS_INIT='{
    "dao_core": "'$CORE_ADDR'",
    "voting_module": "'$VOTING_ADDR'",
    "min_voting_period": 86400,
    "max_voting_period": 604800,
    "pass_threshold": {"absolute_percentage": {"percentage": "0.5"}},
    "min_verification_level": 1
}'
PROPOSALS_ADDR=$(instantiate_contract "$PROPOSALS_CODE_ID" "$PROPOSALS_INIT" "identity-proposals")

# Pre-propose module
PRE_PROPOSE_INIT='{
    "dao_core": "'$CORE_ADDR'",
    "proposal_module": "'$PROPOSALS_ADDR'",
    "deposit_amount": "1000000",
    "deposit_denom": "uatom",
    "min_verification_level": 1,
    "admin_approval_required": false
}'
PRE_PROPOSE_ADDR=$(instantiate_contract "$PRE_PROPOSE_CODE_ID" "$PRE_PROPOSE_INIT" "pre-propose-identity")

# Update core configuration
echo -e "${BLUE}Updating core configuration...${NC}"

UPDATE_MSG='{
    "update_config": {
        "voting_module": "'$VOTING_ADDR'",
        "proposal_modules": ["'$PROPOSALS_ADDR'"]
    }
}'

gaiad tx wasm execute "$CORE_ADDR" "$UPDATE_MSG" \
    --from "$DEPLOYER" \
    --chain-id "$COSMOS_HUB_CHAIN_ID" \
    --node "$COSMOS_HUB_NODE" \
    --gas-prices "$GAS_PRICES" \
    --gas auto \
    --gas-adjustment 1.5 \
    --yes

echo -e "${GREEN}✓ Core configuration updated${NC}"

# Setup IBC channels
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Setting up IBC Channels${NC}"
echo -e "${BLUE}========================================${NC}"

# Create Hermes config if not exists
HERMES_CONFIG="$HOME/.hermes/config.toml"
if [ ! -f "$HERMES_CONFIG" ]; then
    echo -e "${BLUE}Creating Hermes configuration...${NC}"
    mkdir -p "$HOME/.hermes"
    cat >"$HERMES_CONFIG" <<EOF
[global]
log_level = 'info'

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

[[chains]]
id = '$COSMOS_HUB_CHAIN_ID'
type = 'CosmosSdk'
rpc_addr = '$COSMOS_HUB_NODE'
grpc_addr = 'http://localhost:9090'
websocket_addr = 'ws://localhost:26657/websocket'
rpc_timeout = '15s'
account_prefix = 'cosmos'
key_name = 'relayer-cosmos'
store_prefix = 'ibc'
gas_price = { price = 0.025, denom = 'uatom' }
gas_multiplier = 1.5
max_gas = 10000000
clock_drift = '15s'
trusting_period = '14days'
trust_threshold = { numerator = '2', denominator = '3' }

[[chains]]
id = '$SONR_CHAIN_ID'
type = 'CosmosSdk'
rpc_addr = '$SONR_NODE'
grpc_addr = 'http://localhost:9091'
websocket_addr = 'ws://localhost:26658/websocket'
rpc_timeout = '15s'
account_prefix = 'sonr'
key_name = 'relayer-sonr'
store_prefix = 'ibc'
gas_price = { price = 0.025, denom = 'usnr' }
gas_multiplier = 1.5
max_gas = 10000000
clock_drift = '15s'
trusting_period = '14days'
trust_threshold = { numerator = '2', denominator = '3' }
EOF
fi

# Add relayer keys
echo -e "${BLUE}Setting up relayer keys...${NC}"

# Export deployer key from both chains (you'll need to have these)
gaiad keys export "$DEPLOYER" 2>/dev/null | hermes keys add --chain "$COSMOS_HUB_CHAIN_ID" --key-file /dev/stdin || true
snrd keys export "$DEPLOYER" 2>/dev/null | hermes keys add --chain "$SONR_CHAIN_ID" --key-file /dev/stdin || true

# Create IBC connection
echo -e "${BLUE}Creating IBC connection...${NC}"

CONNECTION_RESULT=$(hermes create connection \
    --a-chain "$COSMOS_HUB_CHAIN_ID" \
    --b-chain "$SONR_CHAIN_ID")

CONNECTION_ID=$(echo "$CONNECTION_RESULT" | grep -oP 'connection-\d+' | head -1)

echo -e "${GREEN}✓ Created IBC connection: ${CONNECTION_ID}${NC}"

# Create channels for each contract
echo -e "${BLUE}Creating IBC channels for contracts...${NC}"

# Channel for voting module
VOTING_CHANNEL=$(hermes create channel \
    --a-chain "$COSMOS_HUB_CHAIN_ID" \
    --a-connection "$CONNECTION_ID" \
    --a-port "wasm.$VOTING_ADDR" \
    --b-port "did" \
    --order unordered \
    --version "identity-dao-1" | grep -oP 'channel-\d+' | head -1)

echo -e "${GREEN}✓ Created voting channel: ${VOTING_CHANNEL}${NC}"

# Channel for proposals module
PROPOSALS_CHANNEL=$(hermes create channel \
    --a-chain "$COSMOS_HUB_CHAIN_ID" \
    --a-connection "$CONNECTION_ID" \
    --a-port "wasm.$PROPOSALS_ADDR" \
    --b-port "dwn" \
    --order unordered \
    --version "identity-dao-1" | grep -oP 'channel-\d+' | head -1)

echo -e "${GREEN}✓ Created proposals channel: ${PROPOSALS_CHANNEL}${NC}"

# Start the relayer
echo -e "${BLUE}Starting IBC relayer...${NC}"

hermes start &
RELAYER_PID=$!

echo -e "${GREEN}✓ IBC relayer started with PID: ${RELAYER_PID}${NC}"

# Save deployment information
DEPLOYMENT_FILE="cosmos-hub-deployment.json"
cat >"$DEPLOYMENT_FILE" <<EOF
{
    "chain_id": "$COSMOS_HUB_CHAIN_ID",
    "deployment_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "contracts": {
        "core": {
            "code_id": $CORE_CODE_ID,
            "address": "$CORE_ADDR"
        },
        "voting": {
            "code_id": $VOTING_CODE_ID,
            "address": "$VOTING_ADDR",
            "ibc_channel": "$VOTING_CHANNEL"
        },
        "proposals": {
            "code_id": $PROPOSALS_CODE_ID,
            "address": "$PROPOSALS_ADDR",
            "ibc_channel": "$PROPOSALS_CHANNEL"
        },
        "pre_propose": {
            "code_id": $PRE_PROPOSE_CODE_ID,
            "address": "$PRE_PROPOSE_ADDR"
        }
    },
    "ibc": {
        "connection_id": "$CONNECTION_ID",
        "relayer_pid": $RELAYER_PID
    }
}
EOF

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Deployment Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Deployment information saved to: $DEPLOYMENT_FILE"
echo ""
echo "Contract Addresses:"
echo "  Core:        $CORE_ADDR"
echo "  Voting:      $VOTING_ADDR"
echo "  Proposals:   $PROPOSALS_ADDR"
echo "  Pre-Propose: $PRE_PROPOSE_ADDR"
echo ""
echo "IBC Channels:"
echo "  Voting:    $VOTING_CHANNEL"
echo "  Proposals: $PROPOSALS_CHANNEL"
echo ""
echo "Relayer PID: $RELAYER_PID"
echo ""
echo "To stop the relayer: kill $RELAYER_PID"
echo ""
echo "Next steps:"
echo "1. Verify IBC channels are active: hermes query channels --chain $COSMOS_HUB_CHAIN_ID"
echo "2. Test DID verification: gaiad tx wasm execute $VOTING_ADDR '{\"update_voter\":{\"did\":\"did:sonr:test\",\"address\":\"cosmos1...\"}}'"
echo "3. Create a test proposal through the pre-propose module"
