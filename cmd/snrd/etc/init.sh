#!/bin/bash
# Run this script to quickly install, setup, and run the current version of the network without docker.
#
# Examples:
# CHAIN_ID="localchain_9000-1" HOME_DIR="~/.sonr" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_node.sh
# CHAIN_ID="localchain_9000-2" HOME_DIR="~/.sonr" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 BLOCK_TIME="500ms" sh scripts/test_node.sh

set -eu

export KEY="acc0"
export KEY2="acc1"

export CHAIN_ID=${CHAIN_ID:-"sonrtest_1-1"}
export MONIKER="localvalidator"
export KEYALGO="eth_secp256k1"
export KEYRING=${KEYRING:-"test"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.sonr"}")
export BINARY=${BINARY:-snrd}
export DENOM=${DENOM:-usnr}

export CLEAN=${CLEAN:-"false"}
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}
export JSON_RPC=${JSON_RPC:-"8545"}
export JSON_RPC_WS=${JSON_RPC_WS:-"8546"}
export BLOCK_TIME=${BLOCK_TIME:-"5s"}

# Configurable Mnemomics
export SONR_MNEMONIC_1=${SONR_MNEMONIC_1:-"decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"}
export SONR_MNEMONIC_2=${SONR_MNEMONIC_2:-"wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"}

# Check if binary exists, if not use Docker (or force Docker if requested)
export FORCE_DOCKER=${FORCE_DOCKER:-"false"}
export SKIP_INSTALL=${SKIP_INSTALL:-"false"}
USE_DOCKER=false
if [[ "${FORCE_DOCKER}" == "true" ]] || [[ -z $(which "${BINARY}") ]]; then
	# Check if Docker is available and use it
	if command -v docker >/dev/null 2>&1; then
		if [[ "${FORCE_DOCKER}" == "true" ]]; then
			echo "Force Docker mode enabled, using Docker image onsonr/snrd:latest..."
		else
			echo "Binary ${BINARY} not found locally, checking for Docker image onsonr/snrd:latest..."
		fi
		if docker image inspect onsonr/snrd:latest >/dev/null 2>&1; then
			echo "Using Docker image onsonr/snrd:latest"
			USE_DOCKER=true
		else
			echo "Docker image onsonr/snrd:latest not found. Pulling image..."
			docker pull onsonr/snrd:latest || {
				echo "Failed to pull onsonr/snrd:latest. Please ensure Docker is running and you have internet access."
				exit 1
			}
			USE_DOCKER=true
		fi
	else
		echo "Binary ${BINARY} not found. Please either:"
		echo "  1. Install ${BINARY} with 'make install'"
		echo "  2. Install Docker to use the containerized version"
		exit 1
	fi
fi

# Final check if not using Docker
if [[ "${USE_DOCKER}" == "false" ]]; then
	command -v "${BINARY}" >/dev/null 2>&1 || {
		echo >&2 "${BINARY} command not found. Ensure this is setup / properly installed in your GOPATH (make install)."
		exit 1
	}
fi
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# generate_vrf_key generates a VRF keypair and stores it securely
# Mirrors the Go implementation in app/commands/enhance_init.go
generate_vrf_key() {
	local home_dir="$1"
	local use_docker="${2:-false}"

	# Validate parameters
	if [[ -z "${home_dir}" ]]; then
		echo "Error: HOME_DIR parameter is required" >&2
		return 1
	fi

	# Path to genesis file
	local genesis_file="${home_dir}/config/genesis.json"

	# Check if genesis file exists
	if [[ ! -f "${genesis_file}" ]]; then
		echo "Error: Genesis file not found at ${genesis_file}" >&2
		return 1
	fi

	# Extract chain-id from genesis file
	local chain_id
	chain_id=$(jq -r '.chain_id' "${genesis_file}" 2>/dev/null)

	if [[ -z "${chain_id}" || "${chain_id}" == "null" ]]; then
		echo "Error: Failed to extract chain-id from genesis file" >&2
		return 1
	fi

	echo "Generating VRF keypair for network: ${chain_id}"

	# Create deterministic entropy from chain-id using SHA256
	local entropy_seed
	entropy_seed=$(echo -n "${chain_id}" | sha256sum | cut -d' ' -f1)

	# Generate 64 bytes of deterministic randomness
	local seed_part1="${entropy_seed}"
	local seed_part2
	seed_part2=$(echo -n "${entropy_seed}" | sha256sum | cut -d' ' -f1)

	# Combine to create 64 bytes of hex data
	local vrf_key_hex="${seed_part1}${seed_part2}"

	# Ensure we have exactly 128 hex characters (64 bytes)
	if [[ ${#vrf_key_hex} -ne 128 ]]; then
		echo "Error: Generated VRF key has incorrect size: ${#vrf_key_hex}" >&2
		return 1
	fi

	# Path to store VRF secret key
	local vrf_key_path="${home_dir}/vrf_secret.key"

	# Ensure directory exists
	mkdir -p "${home_dir}"

	# Convert hex to binary and write to file
	echo -n "${vrf_key_hex}" | xxd -r -p > "${vrf_key_path}"

	# Set restrictive permissions (owner read/write only)
	chmod 0600 "${vrf_key_path}"

	# Validate file was created with correct size (64 bytes)
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

# Create wrapper function for binary execution
run_binary() {
	if [[ "${USE_DOCKER}" == "true" ]]; then
		# Ensure the directory exists on the host
		mkdir -p "${HOME_DIR}"
		# Determine if we're in a TTY
		DOCKER_TTY_FLAG=""
		if [ -t 0 ]; then
			DOCKER_TTY_FLAG="-it"
		fi
		# Mount home directory to container's /root/.sonr
		docker run --rm ${DOCKER_TTY_FLAG} \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr "$@"
	else
		${BINARY} "$@"
	fi
}

set_config() {
	run_binary config set client chain-id "${CHAIN_ID}"
	run_binary config set client keyring-backend "${KEYRING}"
}
set_config

from_scratch() {
	# Fresh install on current branch (skip if using Docker or SKIP_INSTALL is true)
	if [[ "${USE_DOCKER}" == "false" ]] && [[ "${SKIP_INSTALL}" == "false" ]]; then
		make install
	fi

	# remove existing daemon files.
	if [[ ${#HOME_DIR} -le 2 ]]; then
		echo "HOME_DIR must be more than 2 characters long"
		return
	fi
	rm -rf "${HOME_DIR}" && echo "Removed ${HOME_DIR}"

	# reset values if not set already after whipe
	set_config

	add_key() {
		key=$1
		mnemonic=$2
		if [[ "${USE_DOCKER}" == "true" ]]; then
			# For Docker, we need to pass the mnemonic differently
			mkdir -p "${HOME_DIR}"
			echo "${mnemonic}" | docker run --rm -i \
				-v "${HOME_DIR}:/root/.sonr" \
				--network host \
				onsonr/snrd:latest \
				snrd --home /root/.sonr keys add "${key}" --keyring-backend "${KEYRING}" --algo "${KEYALGO}" --recover
		else
			echo "${mnemonic}" | ${BINARY} keys add "${key}" --keyring-backend "${KEYRING}" --algo "${KEYALGO}" --home "${HOME_DIR}" --recover
		fi
	}

	# idx140fehngcrxvhdt84x729p3f0qmkmea8n570lrg
	add_key "${KEY}" "${SONR_MNEMONIC_1}"

	# idx1r6yue0vuyj9m7xw78npspt9drq2tmtvgcrf7sr
	add_key "${KEY2}" "${SONR_MNEMONIC_2}"

	if [[ "${USE_DOCKER}" == "true" ]]; then
		# For Docker init, we need to handle it specially
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr init "${MONIKER}" --chain-id "${CHAIN_ID}" --default-denom "${DENOM}"
	else
		${BINARY} init "${MONIKER}" --chain-id "${CHAIN_ID}" --default-denom "${DENOM}" --home "${HOME_DIR}"
	fi

	update_test_genesis() {
		cat "${HOME_DIR}"/config/genesis.json | jq "$1" >"${HOME_DIR}"/config/tmp_genesis.json && mv "${HOME_DIR}"/config/tmp_genesis.json "${HOME_DIR}"/config/genesis.json
	}

	# === CORE MODULES ===

	# Block
	update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'

	# Gov
	update_test_genesis $(printf '.app_state["gov"]["params"]["min_deposit"]=[{"denom":"%s","amount":"1000000"}]' "${DENOM}")
	update_test_genesis '.app_state["gov"]["params"]["voting_period"]="30s"'
	update_test_genesis '.app_state["gov"]["params"]["expedited_voting_period"]="15s"'

	# Add CONSTITUTION.md to governance if it exists
	if [ -f "CONSTITUTION.md" ]; then
		CONSTITUTION_CONTENT=$(cat CONSTITUTION.md | jq -Rs .)
		update_test_genesis ".app_state[\"gov\"][\"constitution\"]=$CONSTITUTION_CONTENT"
	fi

	update_test_genesis $(printf '.app_state["evm"]["params"]["evm_denom"]="%s"' "${DENOM}")
	update_test_genesis '.app_state["evm"]["params"]["active_static_precompiles"]=["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805"]'
	update_test_genesis '.app_state["erc20"]["params"]["native_precompiles"]=["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]' # https://eips.ethereum.org/EIPS/eip-7528
	update_test_genesis $(printf '.app_state["erc20"]["token_pairs"]=[{contract_owner:1,erc20_address:"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE",denom:"%s",enabled:true}]' "${DENOM}")
	update_test_genesis '.app_state["feemarket"]["params"]["no_base_fee"]=true'
	update_test_genesis '.app_state["feemarket"]["params"]["base_fee"]="0.000000000000000000"'

	# staking
	update_test_genesis $(printf '.app_state["staking"]["params"]["bond_denom"]="%s"' "${DENOM}")
	update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.050000000000000000"'

	# mint
	update_test_genesis $(printf '.app_state["mint"]["params"]["mint_denom"]="%s"' "${DENOM}")

	# crisis
	update_test_genesis $(printf '.app_state["crisis"]["constant_fee"]={"denom":"%s","amount":"1000"}' "${DENOM}")

	## abci
	update_test_genesis '.consensus["params"]["abci"]["vote_extensions_enable_height"]="1"'

	# === CUSTOM MODULES ===
	# tokenfactory
	update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[]'
	update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_gas_consume"]=100000'

	BASE_GENESIS_ALLOCATIONS="100000000000000000000000000${DENOM},100000000test"

	# Allocate genesis accounts
	if [[ "${USE_DOCKER}" == "true" ]]; then
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr genesis add-genesis-account "${KEY}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --append
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr genesis add-genesis-account "${KEY2}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --append
		# Sign genesis transaction
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr genesis gentx "${KEY}" 1000000000000000000000"${DENOM}" --gas-prices 0"${DENOM}" --keyring-backend "${KEYRING}" --chain-id "${CHAIN_ID}"
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr genesis collect-gentxs
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr genesis validate-genesis
	else
		${BINARY} genesis add-genesis-account "${KEY}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append
		${BINARY} genesis add-genesis-account "${KEY2}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append
		# Sign genesis transaction
		${BINARY} genesis gentx "${KEY}" 1000000000000000000000"${DENOM}" --gas-prices 0"${DENOM}" --keyring-backend "${KEYRING}" --chain-id "${CHAIN_ID}" --home "${HOME_DIR}"
		${BINARY} genesis collect-gentxs --home "${HOME_DIR}"
		${BINARY} genesis validate-genesis --home "${HOME_DIR}"
	fi
	err=$?
	if [[ ${err} -ne 0 ]]; then
		echo "Failed to validate genesis"
		return
	fi
}

# check if CLEAN is not set to false
if [[ ${CLEAN} != "false" ]]; then
	echo "Starting from a clean state"
	from_scratch

	# Generate VRF keypair (must be done after genesis file is created)
	echo ""
	echo "Generating VRF keypair..."
	if ! generate_vrf_key "${HOME_DIR}" "${USE_DOCKER}"; then
		echo "Warning: VRF key generation failed, but continuing..."
		echo "Note: Multi-validator encryption features may not work without VRF keys"
	fi
fi

echo "Starting node..."

# Opens the RPC endpoint to outside connections
sed -i -e 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:'"${RPC}"'"/g' "${HOME_DIR}"/config/config.toml
sed -i -e 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["*"\]/g' "${HOME_DIR}"/config/config.toml

# REST endpoint
sed -i -e 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:'"${REST}"'"/g' "${HOME_DIR}"/config/app.toml
sed -i -e 's/enable = false/enable = true/g' "${HOME_DIR}"/config/app.toml
sed -i -e 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' "${HOME_DIR}"/config/app.toml

# peer exchange
sed -i -e 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'"${PROFF}"'"/g' "${HOME_DIR}"/config/config.toml
sed -i -e 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'"${P2P}"'"/g' "${HOME_DIR}"/config/config.toml

# GRPC
sed -i -e 's/address = "localhost:9090"/address = "0.0.0.0:'"${GRPC}"'"/g' "${HOME_DIR}"/config/app.toml
sed -i -e 's/address = "localhost:9091"/address = "0.0.0.0:'"${GRPC_WEB}"'"/g' "${HOME_DIR}"/config/app.toml

# Rosetta Api
sed -i -e 's/address = ":8080"/address = "0.0.0.0:'"${ROSETTA}"'"/g' "${HOME_DIR}"/config/app.toml

# JSON-RPC
sed -i -e '/\[json-rpc\]/,/^\[/ s/enable = false/enable = true/' "${HOME_DIR}"/config/app.toml
sed -i -e '/\[json-rpc\]/,/^\[/ s/address = "127.0.0.1:8545"/address = "0.0.0.0:'"${JSON_RPC}"'"/' "${HOME_DIR}"/config/app.toml
sed -i -e '/\[json-rpc\]/,/^\[/ s/ws-address = "127.0.0.1:8546"/ws-address = "0.0.0.0:'"${JSON_RPC_WS}"'"/' "${HOME_DIR}"/config/app.toml

# Faster blocks
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "'"${BLOCK_TIME}"'"/g' "${HOME_DIR}"/config/config.toml

# Start the node (with or without Docker)
if [[ "${USE_DOCKER}" == "true" ]]; then
	echo "Starting node using Docker..."

	# Check for detached mode via environment variable or prompt
	DETACHED_MODE=""
	if [[ "${DOCKER_DETACHED}" == "true" ]]; then
		DETACHED_MODE="-d"
		echo "Running in detached mode. Use 'docker logs -f sonr-testnode' to view logs."
		echo "Stop with: docker stop sonr-testnode"
	elif [ -t 0 ]; then
		echo ""
		echo "Would you like to run the node in detached mode (background)? [y/N]"
		read -r -n 1 DETACH_RESPONSE
		echo ""
		if [[ "$DETACH_RESPONSE" =~ ^[Yy]$ ]]; then
			DETACHED_MODE="-d"
			echo "Running in detached mode. Use 'docker logs -f sonr-testnode' to view logs."
			echo "Stop with: docker stop sonr-testnode"
		else
			echo "Running in foreground mode. Use Ctrl+C to stop."
		fi
	fi

	# Determine if we're in a TTY (only for non-detached mode)
	DOCKER_TTY_FLAG=""
	if [ -t 0 ] && [ -z "$DETACHED_MODE" ]; then
		DOCKER_TTY_FLAG="-it"
	fi

	docker run --rm ${DETACHED_MODE} ${DOCKER_TTY_FLAG} \
		-v "${HOME_DIR}:/root/.sonr" \
		--network host \
		--name sonr-testnode \
		onsonr/snrd:latest \
		snrd start --pruning=nothing --minimum-gas-prices=0"${DENOM}" --rpc.laddr="tcp://0.0.0.0:${RPC}" --home /root/.sonr --json-rpc.api=eth,txpool,personal,net,debug,web3 --json-rpc.address="0.0.0.0:${JSON_RPC}" --json-rpc.ws-address="0.0.0.0:${JSON_RPC_WS}" --chain-id="${CHAIN_ID}"

	# If running detached, show status
	if [ -n "$DETACHED_MODE" ]; then
		echo ""
		echo "✅ Node started in background"
		echo ""
		echo "Useful commands:"
		echo "  View logs:    docker logs -f sonr-testnode"
		echo "  Stop node:    docker stop sonr-testnode"
		echo "  Node status:  curl http://localhost:${RPC}/status | jq '.result.sync_info'"
		echo ""
	fi
else
	${BINARY} start --pruning=nothing --minimum-gas-prices=0"${DENOM}" --rpc.laddr="tcp://0.0.0.0:${RPC}" --home "${HOME_DIR}" --json-rpc.api=eth,txpool,personal,net,debug,web3 --json-rpc.address="0.0.0.0:${JSON_RPC}" --json-rpc.ws-address="0.0.0.0:${JSON_RPC_WS}" --chain-id="${CHAIN_ID}"
fi
