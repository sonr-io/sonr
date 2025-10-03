#!/bin/bash
set -eu

# Load .env file if it exists
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
if [ -f "${REPO_ROOT}/.env" ]; then
    set -a
    source "${REPO_ROOT}/.env"
    set +a
fi

export CHAIN_ID=${CHAIN_ID:-"sonrtest_1-1"}
export DENOM=${DENOM:-"usnr"}
export KEYRING=${KEYRING:-"test"}
export KEYALGO="eth_secp256k1"

VALIDATORS=("alice" "bob" "carol")
VAL_HOMES=("./val-alice" "./val-bob" "./val-carol")
SENTRY_HOMES=("./sentry-alice" "./sentry-bob" "./sentry-carol")

# Get mnemonics from environment
declare -A MNEMONICS
MNEMONICS["alice"]="$ALICE_MNEMONIC"
MNEMONICS["bob"]="$BOB_MNEMONIC"
MNEMONICS["carol"]="$CAROL_MNEMONIC"
MNEMONICS["faucet"]="$FAUCET_MNEMONIC"

echo "🚀 Initializing Sonr Testnet with 3 validators..."
echo "Chain ID: $CHAIN_ID"
echo "Denom: $DENOM"
echo ""

# Check if snrd is available locally
if ! command -v snrd &> /dev/null; then
    echo "❌ snrd binary not found in PATH"
    echo "Please install snrd locally or add it to your PATH"
    exit 1
fi

echo "✅ Using local snrd: $(which snrd)"
echo ""

# Function to initialize a validator node
init_validator() {
    local name=$1
    local home_dir=$2
    local mnemonic=$3

    echo "📋 Initializing validator: $name"

    local abs_home="${REPO_ROOT}/${home_dir#./}"
    rm -rf "$abs_home" 2>/dev/null || true
    mkdir -p "$abs_home"

    echo | snrd --home "$abs_home" init "val-$name" --chain-id "$CHAIN_ID" --default-denom "$DENOM" --overwrite >/dev/null 2>&1

    snrd --home "$abs_home" keys delete "$name" --keyring-backend "$KEYRING" -y 2>/dev/null || true

    echo "$mnemonic" | snrd --home "$abs_home" keys add "$name" \
        --keyring-backend "$KEYRING" \
        --algo "$KEYALGO" \
        --recover

    update_config "$abs_home"
}

# Function to initialize a sentry node
init_sentry() {
    local name=$1
    local home_dir=$2

    echo "🛡️  Initializing sentry: $name"

    local abs_home="${REPO_ROOT}/${home_dir#./}"
    rm -rf "$abs_home" 2>/dev/null || true
    mkdir -p "$abs_home"

    echo | snrd --home "$abs_home" init "sentry-$name" --chain-id "$CHAIN_ID" --default-denom "$DENOM" --overwrite >/dev/null 2>&1

    update_config "$abs_home"
}

# Function to update node config
update_config() {
    local abs_home=$1

    sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/g' "${abs_home}/config/config.toml"
    sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["*"\]/g' "${abs_home}/config/config.toml"
    sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/g' "${abs_home}/config/app.toml"
    sed -i '/\[api\]/,/\[grpc\]/ s/enable = false/enable = true/' "${abs_home}/config/app.toml"
    sed -i 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' "${abs_home}/config/app.toml"
    sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/g' "${abs_home}/config/app.toml"
    sed -i 's/address = "localhost:9091"/address = "0.0.0.0:9091"/g' "${abs_home}/config/app.toml"
}

# Function to update genesis
update_genesis() {
    local genesis_file="$1"

    cat "$genesis_file" | \
    jq '.consensus_params.block.max_gas="100000000"' | \
    jq ".app_state.gov.params.min_deposit=[{\"denom\":\"$DENOM\",\"amount\":\"1000000\"}]" | \
    jq '.app_state.gov.params.voting_period="30s"' | \
    jq '.app_state.gov.params.expedited_voting_period="15s"' | \
    jq ".app_state.evm.params.evm_denom=\"$DENOM\"" | \
    jq '.app_state.evm.params.active_static_precompiles=["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805"]' | \
    jq '.app_state.erc20.params.native_precompiles=["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]' | \
    jq ".app_state.erc20.token_pairs=[{contract_owner:1,erc20_address:\"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE\",denom:\"$DENOM\",enabled:true}]" | \
    jq '.app_state.feemarket.params.no_base_fee=true' | \
    jq '.app_state.feemarket.params.base_fee="0.000000000000000000"' | \
    jq ".app_state.staking.params.bond_denom=\"$DENOM\"" | \
    jq '.app_state.staking.params.min_commission_rate="0.050000000000000000"' | \
    jq ".app_state.mint.params.mint_denom=\"$DENOM\"" | \
    jq ".app_state.crisis.constant_fee={\"denom\":\"$DENOM\",\"amount\":\"1000\"}" | \
    jq '.consensus.params.abci.vote_extensions_enable_height="1"' | \
    jq '.app_state.tokenfactory.params.denom_creation_fee=[]' | \
    jq '.app_state.tokenfactory.params.denom_creation_gas_consume=100000' \
    > "$genesis_file.tmp" && mv "$genesis_file.tmp" "$genesis_file"
}

# Initialize all validators
for i in "${!VALIDATORS[@]}"; do
    name="${VALIDATORS[$i]}"
    init_validator "$name" "${VAL_HOMES[$i]}" "${MNEMONICS[$name]}"
done

# Initialize all sentries
for i in "${!VALIDATORS[@]}"; do
    init_sentry "${VALIDATORS[$i]}" "${SENTRY_HOMES[$i]}"
done

echo ""
echo "💰 Adding genesis accounts and creating genesis transactions..."

BASE_ALLOCATION="100000000000000000000000000${DENOM}"
STAKE_AMOUNT="30000000000000000000000${DENOM}"

# Use alice's validator for genesis creation
GENESIS_HOME="./val-alice"

# Add genesis accounts for all validators
GENESIS_ABS_HOME="${REPO_ROOT}/${GENESIS_HOME#./}"
for i in "${!VALIDATORS[@]}"; do
    name="${VALIDATORS[$i]}"
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    # Get address from each validator's keyring
    addr=$(snrd --home "$abs_home" keys show "$name" --keyring-backend "$KEYRING" -a)
    # Add to genesis using address
    snrd --home "$GENESIS_ABS_HOME" genesis add-genesis-account "$addr" "$BASE_ALLOCATION" --append
done

# Add faucet account to genesis
echo "💰 Creating faucet account..."
snrd --home "$GENESIS_ABS_HOME" keys delete faucet --keyring-backend "$KEYRING" -y 2>/dev/null || true
echo "${MNEMONICS["faucet"]}" | snrd --home "$GENESIS_ABS_HOME" keys add faucet \
    --keyring-backend "$KEYRING" \
    --algo "$KEYALGO" \
    --recover
FAUCET_ADDR=$(snrd --home "$GENESIS_ABS_HOME" keys show faucet --keyring-backend "$KEYRING" -a)
FAUCET_ALLOCATION="${FAUCET_ALLOCATION:-500000000000000000000000000${DENOM}}"
snrd --home "$GENESIS_ABS_HOME" genesis add-genesis-account "$FAUCET_ADDR" "$FAUCET_ALLOCATION" --append
echo "  Faucet address: $FAUCET_ADDR"
echo "  Faucet balance: $FAUCET_ALLOCATION"

# Distribute genesis with all accounts to all validators before creating gentx
for i in "${!VALIDATORS[@]}"; do
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    if [ "$abs_home" != "$GENESIS_ABS_HOME" ]; then
        cp "$GENESIS_ABS_HOME/config/genesis.json" "$abs_home/config/genesis.json"
    fi
done

# Create gentx for each validator
for i in "${!VALIDATORS[@]}"; do
    echo "Creating gentx for ${VALIDATORS[$i]}..."
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    snrd --home "$abs_home" genesis gentx "${VALIDATORS[$i]}" "$STAKE_AMOUNT" \
        --keyring-backend "$KEYRING" \
        --chain-id "$CHAIN_ID" \
        --gas-prices "0${DENOM}"

    # Copy gentx to genesis home (skip if same directory)
    if [ "$abs_home" != "$GENESIS_ABS_HOME" ]; then
        cp "${abs_home}/config/gentx"/* "$GENESIS_ABS_HOME/config/gentx/"
    fi
done

# Collect gentxs
echo "Collecting genesis transactions..."
snrd --home "$GENESIS_ABS_HOME" genesis collect-gentxs

# Update genesis parameters
echo "Updating genesis parameters..."
update_genesis "$GENESIS_ABS_HOME/config/genesis.json"

# Validate genesis
echo "Validating genesis..."
snrd --home "$GENESIS_ABS_HOME" genesis validate-genesis

# Distribute genesis to all nodes
echo ""
echo "📤 Distributing genesis to all nodes..."
for home in "${VAL_HOMES[@]}" "${SENTRY_HOMES[@]}"; do
    if [ "$home" != "$GENESIS_HOME" ]; then
        abs_home="${REPO_ROOT}/${home#./}"
        cp "$GENESIS_ABS_HOME/config/genesis.json" "$abs_home/config/genesis.json"
    fi
done

echo ""
echo "🔗 Setting up peer connections..."

# Get validator node IDs
declare -A VAL_IDS
for i in "${!VALIDATORS[@]}"; do
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    VAL_IDS[${VALIDATORS[$i]}]=$(snrd --home "$abs_home" tendermint show-node-id | tr -d '\r\n')
    echo "  val-${VALIDATORS[$i]}: ${VAL_IDS[${VALIDATORS[$i]}]}"
done

# Get sentry node IDs
declare -A SENTRY_IDS
for i in "${!VALIDATORS[@]}"; do
    abs_home="${REPO_ROOT}/${SENTRY_HOMES[$i]#./}"
    SENTRY_IDS[${VALIDATORS[$i]}]=$(snrd --home "$abs_home" tendermint show-node-id | tr -d '\r\n')
    echo "  sentry-${VALIDATORS[$i]}: ${SENTRY_IDS[${VALIDATORS[$i]}]}"
done

# Configure validators to connect to their sentries only
for i in "${!VALIDATORS[@]}"; do
    name="${VALIDATORS[$i]}"
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    sed -i "s/persistent_peers = \"\"/persistent_peers = \"${SENTRY_IDS[$name]}@sentry-$name:26656\"/g" "${abs_home}/config/config.toml"
done

# Configure sentries to connect to their validators and other sentries
for i in "${!VALIDATORS[@]}"; do
    name="${VALIDATORS[$i]}"
    abs_home="${REPO_ROOT}/${SENTRY_HOMES[$i]#./}"

    # Build seeds list (all other sentries)
    seeds=""
    for j in "${!VALIDATORS[@]}"; do
        other_name="${VALIDATORS[$j]}"
        if [ "$name" != "$other_name" ]; then
            if [ -n "$seeds" ]; then
                seeds="${seeds},"
            fi
            seeds="${seeds}${SENTRY_IDS[$other_name]}@sentry-$other_name:26656"
        fi
    done

    # Set persistent peer to own validator and seeds to other sentries
    sed -i "s/persistent_peers = \"\"/persistent_peers = \"${VAL_IDS[$name]}@val-$name:26656\"/g" "${abs_home}/config/config.toml"
    sed -i "s/seeds = \"\"/seeds = \"$seeds\"/g" "${abs_home}/config/config.toml"
    sed -i "s/private_peer_ids = \"\"/private_peer_ids = \"${VAL_IDS[$name]}\"/g" "${abs_home}/config/config.toml"
done

echo ""
echo "✅ Testnet initialization complete!"
echo ""
echo "🎯 Validator Addresses:"
for i in "${!VALIDATORS[@]}"; do
    abs_home="${REPO_ROOT}/${VAL_HOMES[$i]#./}"
    addr=$(snrd --home "$abs_home" keys show "${VALIDATORS[$i]}" --keyring-backend "$KEYRING" -a | tr -d '\r\n')
    echo "  ${VALIDATORS[$i]}: $addr"
done

echo ""
echo "💰 Faucet Address:"
echo "  faucet: $FAUCET_ADDR"

echo ""
echo "📡 Service Endpoints (via Cloudflare Tunnel):"
echo "  Alice RPC:       https://alice-rpc.sonr.land"
echo "  Alice REST:      https://alice-rest.sonr.land"
echo "  Alice gRPC:      https://alice-grpc.sonr.land"
echo "  Alice EVM:       https://alice-evm.sonr.land"
echo "  Bob RPC:         https://bob-rpc.sonr.land"
echo "  Bob REST:        https://bob-rest.sonr.land"
echo "  Bob gRPC:        https://bob-grpc.sonr.land"
echo "  Bob EVM:         https://bob-evm.sonr.land"
echo "  Carol RPC:       https://carol-rpc.sonr.land"
echo "  Carol REST:      https://carol-rest.sonr.land"
echo "  Carol gRPC:      https://carol-grpc.sonr.land"
echo "  Carol EVM:       https://carol-evm.sonr.land"
echo "  IPFS API:        https://ipfs-api.sonr.land"
echo "  IPFS Gateway:    https://ipfs-gateway.sonr.land"
echo ""
echo "🚀 Start testnet with: docker compose up -d"
echo "📊 View logs with:     docker compose logs -f"
echo "🛑 Stop testnet with:  docker compose down"