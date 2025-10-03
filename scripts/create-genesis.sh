#!/bin/bash

set -eux

# generate_vrf_key generates a VRF keypair and stores it securely
generate_vrf_key() {
	local home_dir="$1"

	if [[ -z "${home_dir}" ]]; then
		echo "Error: HOME_DIR parameter is required" >&2
		return 1
	fi

	local genesis_file="${home_dir}/config/genesis.json"

	if [[ ! -f "${genesis_file}" ]]; then
		echo "Error: Genesis file not found at ${genesis_file}" >&2
		return 1
	fi

	local chain_id
	chain_id=$(jq -r '.chain_id' "${genesis_file}" 2>/dev/null)

	if [[ -z "${chain_id}" || "${chain_id}" == "null" ]]; then
		echo "Error: Failed to extract chain-id from genesis file" >&2
		return 1
	fi

	echo "Generating VRF keypair for network: ${chain_id}"

	local entropy_seed
	entropy_seed=$(echo -n "${chain_id}" | sha256sum | cut -d' ' -f1)

	local seed_part1="${entropy_seed}"
	local seed_part2
	seed_part2=$(echo -n "${entropy_seed}" | sha256sum | cut -d' ' -f1)

	local vrf_key_hex="${seed_part1}${seed_part2}"

	if [[ ${#vrf_key_hex} -ne 128 ]]; then
		echo "Error: Generated VRF key has incorrect size: ${#vrf_key_hex}" >&2
		return 1
	fi

	local vrf_key_path="${home_dir}/vrf_secret.key"
	mkdir -p "${home_dir}"

	echo -n "${vrf_key_hex}" | xxd -r -p > "${vrf_key_path}"
	chmod 0600 "${vrf_key_path}"

	local file_size
	file_size=$(wc -c < "${vrf_key_path}")

	if [[ ${file_size} -ne 64 ]]; then
		echo "Error: VRF key file has incorrect size: ${file_size} bytes" >&2
		rm -f "${vrf_key_path}"
		return 1
	fi

	echo "✓ VRF keypair generated for network: ${chain_id}"
	echo "✓ VRF secret key stored securely: ${vrf_key_path}"

	return 0
}

DENOM="${DENOM:=usnr}"
# Match init-testnet.sh allocation: 100000000000000000000000000snr = 100000000000000000000000000000000usnr
COINS="${COINS:=100000000000000000000000000000000$DENOM,100000000000000000000000000snr}"
CHAIN_ID="${CHAIN_ID:=sonrtest_1-1}"
CHAIN_BIN="${CHAIN_BIN:=snrd}"
CHAIN_DIR="${CHAIN_DIR:=$HOME/.sonr}"
KEYS_CONFIG="${KEYS_CONFIG:=configs/keys.json}"

FAUCET_ENABLED="${FAUCET_ENABLED:=true}"
NUM_VALIDATORS="${NUM_VALIDATORS:=1}"
NUM_RELAYERS="${NUM_RELAYERS:=0}"

# check if the binary has genesis subcommand or not, if not, set CHAIN_GENESIS_CMD to empty
CHAIN_GENESIS_CMD=$($CHAIN_BIN 2>&1 | grep -q "genesis-related subcommands" && echo "genesis" || echo "")

jq -r ".genesis[0].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN init "$CHAIN_ID" --chain-id "$CHAIN_ID" --default-denom "$DENOM" --recover

# Add genesis keys to the keyring and self delegate initial coins
echo "Adding key...." $(jq -r ".genesis[0].name" "$KEYS_CONFIG")
jq -r ".genesis[0].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add $(jq -r ".genesis[0].name" "$KEYS_CONFIG") --recover --keyring-backend="test"
$CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a $(jq -r .genesis[0].name "$KEYS_CONFIG") --keyring-backend="test") "$COINS" --keyring-backend="test"

# Add faucet key to the keyring and self delegate initial coins
echo "Adding key...." $(jq -r ".faucet[0].name" "$KEYS_CONFIG")
jq -r ".faucet[0].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add $(jq -r ".faucet[0].name" "$KEYS_CONFIG") --recover --keyring-backend="test"
$CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a $(jq -r .faucet[0].name "$KEYS_CONFIG") --keyring-backend="test") "$COINS" --keyring-backend="test"

# Add test keys to the keyring and self delegate initial coins
echo "Adding key...." $(jq -r ".keys[0].name" "$KEYS_CONFIG")
jq -r ".keys[0].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add $(jq -r ".keys[0].name" "$KEYS_CONFIG") --recover --keyring-backend="test"
$CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a $(jq -r .keys[0].name "$KEYS_CONFIG") --keyring-backend="test") "$COINS" --keyring-backend="test"

if [[ $FAUCET_ENABLED == "false" && $NUM_RELAYERS -gt "-1" ]]; then
  ## Add relayers keys and delegate tokens
  for i in $(seq 0 "$NUM_RELAYERS"); do
    # Add relayer key and delegate tokens
    RELAYER_KEY_NAME="$(jq -r ".relayers[$i].name" "$KEYS_CONFIG")"
    echo "Adding relayer key.... $RELAYER_KEY_NAME"
    jq -r ".relayers[$i].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add "$RELAYER_KEY_NAME" --recover --keyring-backend="test"
    $CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a "$RELAYER_KEY_NAME" --keyring-backend="test") "$COINS" --keyring-backend="test"
    # Add relayer-cli key and delegate tokens
    RELAYER_CLI_KEY_NAME="$(jq -r ".relayers_cli[$i].name" "$KEYS_CONFIG")"
    echo "Adding relayer-cli key.... $RELAYER_CLI_KEY_NAME"
    jq -r ".relayers_cli[$i].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add "$RELAYER_CLI_KEY_NAME" --recover --keyring-backend="test"
    $CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a "$RELAYER_CLI_KEY_NAME" --keyring-backend="test") "$COINS" --keyring-backend="test"
  done
fi

## if faucet not enabled then add validator and relayer with index as keys and into gentx
if [[ $FAUCET_ENABLED == "false" && $NUM_VALIDATORS -gt "1" ]]; then
  ## Add validators key and delegate tokens
  for i in $(seq 0 "$NUM_VALIDATORS"); do
    VAL_KEY_NAME="$(jq -r '.validators[0].name' "$KEYS_CONFIG")-$i"
    echo "Adding validator key.... $VAL_KEY_NAME"
    jq -r ".validators[0].mnemonic" "$KEYS_CONFIG" | $CHAIN_BIN keys add "$VAL_KEY_NAME" --index "$i" --recover --keyring-backend="test"
    $CHAIN_BIN "$CHAIN_GENESIS_CMD" add-genesis-account $($CHAIN_BIN keys show -a "$VAL_KEY_NAME" --keyring-backend="test") "$COINS" --keyring-backend="test"
  done
fi

echo "Creating gentx..."
COIN=$(echo "$COINS" | cut -d ',' -f1)
# Use full validator amount to meet minimum delegation requirement (274890886240)
# Match the working init-testnet.sh: 1000000000000000000000snr = 1000000000000000000000000000usnr
VALIDATOR_AMOUNT="1000000000000000000000000000$DENOM"
$CHAIN_BIN "$CHAIN_GENESIS_CMD" gentx $(jq -r ".genesis[0].name" "$KEYS_CONFIG") "$VALIDATOR_AMOUNT" --keyring-backend="test" --chain-id "$CHAIN_ID" --gas-prices="0$DENOM"

echo "Output of gentx"
cat "$CHAIN_DIR"/config/gentx/*.json | jq

echo "Running collect-gentxs"
$CHAIN_BIN "$CHAIN_GENESIS_CMD" collect-gentxs

ls "$CHAIN_DIR"/config

# Generate VRF keypair
echo ""
echo "Generating VRF keypair..."
if ! generate_vrf_key "${CHAIN_DIR}"; then
	echo "Warning: VRF key generation failed, but continuing..."
	echo "Note: Multi-validator encryption features may not work without VRF keys"
fi
