#!/bin/bash
# Run this script to quickly install, setup, and run the current version of the network without docker.
#
# Examples:
# CHAIN_ID="localchain_9000-1" HOME_DIR="~/.sonr" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_node.sh
# CHAIN_ID="localchain_9000-2" HOME_DIR="~/.sonr" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 BLOCK_TIME="500ms" sh scripts/test_node.sh

set -euo pipefail

# Source helper libraries
# Get the directory of this script reliably
SCRIPT_DIR="/usr/local/lib/sonr-scripts"
source "${SCRIPT_DIR}/env.sh"
source "${SCRIPT_DIR}/config.sh"
source "${SCRIPT_DIR}/keys.sh"
source "${SCRIPT_DIR}/genesis.sh"

# Initialize environment
init_env

# Set defaults for test node
export KEY="${KEY:-acc0}"
export KEY2="${KEY2:-acc1}"
export MONIKER="${MONIKER:-localvalidator}"
export KEYALGO="${KEYALGO:-eth_secp256k1}"

# Configurable ports
export RPC="${RPC:-26657}"
export REST="${REST:-1317}"
export PROFF="${PROFF:-6060}"
export P2P="${P2P:-26656}"
export GRPC="${GRPC:-9090}"
export GRPC_WEB="${GRPC_WEB:-9091}"
export ROSETTA="${ROSETTA:-8080}"
export JSON_RPC="${JSON_RPC:-8545}"
export JSON_RPC_WS="${JSON_RPC_WS:-8546}"
export BLOCK_TIME="${BLOCK_TIME:-5s}"

# Configurable mnemonics
export SONR_MNEMONIC_1="${SONR_MNEMONIC_1:-decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry}"
export SONR_MNEMONIC_2="${SONR_MNEMONIC_2:-wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise}"

# Docker and installation options
export FORCE_DOCKER="${FORCE_DOCKER:-false}"
export SKIP_INSTALL="${SKIP_INSTALL:-false}"
export CLEAN="${CLEAN:-false}"
export DOCKER_DETACHED="${DOCKER_DETACHED:-false}"

# Check if binary exists, if not use Docker (or force Docker if requested)
USE_DOCKER=false
if [[ "${FORCE_DOCKER}" == "true" ]] || ! command -v "$CHAIN_BIN" >/dev/null 2>&1; then
	# Check if Docker is available and use it
	if command -v docker >/dev/null 2>&1; then
		if [[ "${FORCE_DOCKER}" == "true" ]]; then
			log_info "Force Docker mode enabled, using Docker image onsonr/snrd:latest"
		else
			log_info "Binary $CHAIN_BIN not found locally, checking for Docker image onsonr/snrd:latest"
		fi
		if docker image inspect onsonr/snrd:latest >/dev/null 2>&1; then
			log_info "Using Docker image onsonr/snrd:latest"
			USE_DOCKER=true
		else
			log_info "Docker image onsonr/snrd:latest not found. Pulling image..."
			docker pull onsonr/snrd:latest || {
				log_error "Failed to pull onsonr/snrd:latest. Please ensure Docker is running and you have internet access."
				exit 1
			}
			USE_DOCKER=true
		fi
	else
		log_error "Binary $CHAIN_BIN not found. Please either:"
		log_error "  1. Install $CHAIN_BIN with 'make install'"
		log_error "  2. Install Docker to use the containerized version"
		exit 1
	fi
fi

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
		${CHAIN_BIN} "$@"
	fi
}

# Set client configuration
set_config() {
	run_binary config set client chain-id "${CHAIN_ID}"
	run_binary config set client keyring-backend "${KEYRING}"
}
set_config

from_scratch() {
	# Fresh install on current branch (skip if using Docker or SKIP_INSTALL is true)
	if [[ "${USE_DOCKER}" == "false" ]] && [[ "${SKIP_INSTALL}" == "false" ]]; then
		log_info "Installing $CHAIN_BIN..."
		make install
	fi

	# Remove existing daemon files
	if [[ ${#HOME_DIR} -le 2 ]]; then
		log_error "HOME_DIR must be more than 2 characters long"
		return 1
	fi
	rm -rf "${HOME_DIR}"
	log_info "Removed existing chain directory: ${HOME_DIR}"

	# Reset configuration
	set_config

	# Add test keys
	log_info "Adding test keys..."
	import_mnemonic "${KEY}" "${SONR_MNEMONIC_1}" "$KEYALGO"
	import_mnemonic "${KEY2}" "${SONR_MNEMONIC_2}" "$KEYALGO"

	# Initialize chain
	log_info "Initializing chain with moniker: $MONIKER"
	if [[ "${USE_DOCKER}" == "true" ]]; then
		docker run --rm \
			-v "${HOME_DIR}:/root/.sonr" \
			--network host \
			onsonr/snrd:latest \
			snrd --home /root/.sonr init "${MONIKER}" --chain-id "${CHAIN_ID}" --default-denom "${DENOM}"
	else
		${CHAIN_BIN} init "${MONIKER}" --chain-id "${CHAIN_ID}" --default-denom "${DENOM}" --home "${HOME_DIR}"
	fi

	# Update genesis parameters
	log_info "Updating genesis parameters..."
	update_genesis_params

	# Add constitution if available
	add_constitution

	# Set up genesis accounts and transactions
	local BASE_GENESIS_ALLOCATIONS="100000000000000000000000000${DENOM},100000000test"

	log_info "Adding genesis accounts..."
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
		${CHAIN_BIN} genesis add-genesis-account "${KEY}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append
		${CHAIN_BIN} genesis add-genesis-account "${KEY2}" "${BASE_GENESIS_ALLOCATIONS}" --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append
		# Sign genesis transaction
		${CHAIN_BIN} genesis gentx "${KEY}" 1000000000000000000000"${DENOM}" --gas-prices 0"${DENOM}" --keyring-backend "${KEYRING}" --chain-id "${CHAIN_ID}" --home "${HOME_DIR}"
		${CHAIN_BIN} genesis collect-gentxs --home "${HOME_DIR}"
		${CHAIN_BIN} genesis validate-genesis --home "${HOME_DIR}"
	fi

	log_success "Genesis setup completed"
}

# Check if CLEAN is not set to false
if [[ ${CLEAN} != "false" ]]; then
	log_info "Starting from a clean state"
	from_scratch

	# Generate VRF keypair (must be done after genesis file is created)
	log_info "Generating VRF keypair..."
	if ! generate_vrf_key "${HOME_DIR}"; then
		log_warn "VRF key generation failed, but continuing..."
		log_warn "Note: Multi-validator encryption features may not work without VRF keys"
	fi
fi

log_info "Configuring node ports and settings..."

# Configure node with all the specified ports and settings
configure_node "$HOME_DIR" \
	--rpc-port "$RPC" \
	--rest-port "$REST" \
	--grpc-port "$GRPC" \
	--grpc-web-port "$GRPC_WEB" \
	--json-rpc-port "$JSON_RPC" \
	--rosetta-port "$ROSETTA" \
	--min-gas-prices "0${DENOM}" \
	--pruning nothing

# Set consensus timeouts
set_consensus_timeouts "$HOME_DIR/config/config.toml" "5s" "1s" "1s" "$BLOCK_TIME"

# Enable CORS for RPC
set_toml_value "$HOME_DIR/config/config.toml" "" "cors_allowed_origins" '["*"]'

# Set pprof address
set_toml_value "$HOME_DIR/config/config.toml" "" "pprof_laddr" "localhost:${PROFF}"

# Set P2P address
set_toml_value "$HOME_DIR/config/config.toml" "p2p" "laddr" "tcp://0.0.0.0:${P2P}"

log_info "Starting node..."

# Start the node (with or without Docker)
if [[ "${USE_DOCKER}" == "true" ]]; then
	log_info "Starting node using Docker..."

	# Check for detached mode via environment variable or prompt
	DETACHED_MODE=""
	if [[ "${DOCKER_DETACHED}" == "true" ]]; then
		DETACHED_MODE="-d"
		log_info "Running in detached mode. Use 'docker logs -f sonr-testnode' to view logs."
		log_info "Stop with: docker stop sonr-testnode"
	elif [ -t 0 ]; then
		log_info ""
		log_info "Would you like to run the node in detached mode (background)? [y/N]"
		read -r -n 1 DETACH_RESPONSE
		log_info ""
		if [[ "$DETACH_RESPONSE" =~ ^[Yy]$ ]]; then
			DETACHED_MODE="-d"
			log_info "Running in detached mode. Use 'docker logs -f sonr-testnode' to view logs."
			log_info "Stop with: docker stop sonr-testnode"
		else
			log_info "Running in foreground mode. Use Ctrl+C to stop."
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
		log_info ""
		log_success "Node started in background"
		log_info ""
		log_info "Useful commands:"
		log_info "  View logs:    docker logs -f sonr-testnode"
		log_info "  Stop node:    docker stop sonr-testnode"
		log_info "  Node status:  curl http://localhost:${RPC}/status | jq '.result.sync_info'"
		log_info ""
	fi
else
	log_info "Starting node locally..."
	${CHAIN_BIN} start --pruning=nothing --minimum-gas-prices=0"${DENOM}" --rpc.laddr="tcp://0.0.0.0:${RPC}" --home "${HOME_DIR}" --json-rpc.api=eth,txpool,personal,net,debug,web3 --json-rpc.address="0.0.0.0:${JSON_RPC}" --json-rpc.ws-address="0.0.0.0:${JSON_RPC_WS}" --chain-id="${CHAIN_ID}"
fi

log_success "Node startup completed"
