#!/bin/bash
# Run this script to quickly install, setup, and run the current version of the network.
#
# Examples:
# CHAIN_ID="localchain-1" HOME_DIR="~/.simapp" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_ics_node.sh
# CHAIN_ID="localchain-2" HOME_DIR="~/.simapp" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 BLOCK_TIME="500ms" sh scripts/test_ics_node.sh

set -eu

export KEY="acc0"
export KEY2="acc1"

export CHAIN_ID=${CHAIN_ID:-"sonrtest_1-1"}
export MONIKER="localvalidator"
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"test"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.simapp"}")
export BINARY=${BINARY:-wasmd}
export DENOM=${DENOM:-token}

export CLEAN=${CLEAN:-"false"}
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}
export BLOCK_TIME=${BLOCK_TIME:-"5s"}

config_toml="${HOME_DIR}/config/config.toml"
client_toml="${HOME_DIR}/config/client.toml"
app_toml="${HOME_DIR}/config/app.toml"
genesis_json="${HOME_DIR}/config/genesis.json"

# if which binary does not exist, install it
if [[ -z $(which "${BINARY}") ]]; then
	make install

	if [[ -z $(which "${BINARY}") ]]; then
		echo "Ensure ${BINARY} is installed and in your PATH"
		exit 1
	fi
fi

command -v "${BINARY}" >/dev/null 2>&1 || {
	echo >&2 "${BINARY} command not found. Ensure this is setup / properly installed in your GOPATH (make install)."
	exit 1
}
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

set_config() {
	${BINARY} config set client chain-id "${CHAIN_ID}"
	${BINARY} config set client keyring-backend "${KEYRING}"
}
set_config

from_scratch() {
	# Fresh install on current branch
	make install

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
		echo "${mnemonic}" | ${BINARY} keys add "${key}" --home "${HOME_DIR}" --keyring-backend "${KEYRING}" --algo "${KEYALGO}" --recover
	}

	# cosmos1efd63aw40lxf3n4mhf7dzhjkr453axur6cpk92
	add_key "${KEY}" "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	# cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr
	add_key "${KEY2}" "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"

	${BINARY} init "${CHAIN_ID}" --chain-id "${CHAIN_ID}" --overwrite --default-denom "${DENOM}" --home "${HOME_DIR}"

	update_test_genesis() {
		cat "${HOME_DIR}"/config/genesis.json | jq "$1" >"${HOME_DIR}"/config/tmp_genesis.json && mv "${HOME_DIR}"/config/tmp_genesis.json "${HOME_DIR}"/config/genesis.json
	}

	# === CORE MODULES ===
	# block
	update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
	# crisis
	update_test_genesis $(printf '.app_state["crisis"]["constant_fee"]={"denom":"%s","amount":"1000"}' "${DENOM}")

	# === CUSTOM MODULES ===
	# tokenfactory
	update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[]'
	update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_gas_consume"]=100000'

	${BINARY} keys list --keyring-backend "${KEYRING}" --home "${HOME_DIR}"

	# Allocate genesis accounts
	${BINARY} genesis add-genesis-account "${KEY}" 10000000"${DENOM}",900test --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append
	${BINARY} genesis add-genesis-account "${KEY2}" 10000000"${DENOM}",800test --keyring-backend "${KEYRING}" --home "${HOME_DIR}" --append

	# ICS provider genesis hack
	HACK_DIR=icshack-1 && echo "${HACK_DIR}"
	rm -rf "${HACK_DIR}"
	cp -r "${HOME_DIR}" "${HACK_DIR}"

	${BINARY} add-consumer-section provider --home "${HACK_DIR}"
	ccvjson=$(jq '.app_state["ccvconsumer"]' "${HACK_DIR}"/config/genesis.json)
	echo "${ccvjson}"
	jq '.app_state["ccvconsumer"] = '"${ccvjson}" "${HACK_DIR}"/config/genesis.json >json.tmp && mv json.tmp "${genesis_json}"
	rm -rf "${HACK_DIR}"

	update_test_genesis $(printf '.app_state["ccvconsumer"]["params"]["unbonding_period"]="%s"' "240s")
}

# check if CLEAN is not set to false
if [[ ${CLEAN} != "false" ]]; then
	echo "Starting from a clean state"
	from_scratch
fi

# Opens the RPC endpoint to outside connections
sed -i -e 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'"${RPC}"'"/g' "${HOME_DIR}"/config/config.toml
sed -i -e 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' "${HOME_DIR}"/config/config.toml

# REST endpoint
sed -i -e 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:'"${REST}"'"/g' "${HOME_DIR}"/config/app.toml
sed -i -e 's/enable = false/enable = true/g' "${HOME_DIR}"/config/app.toml

# peer exchange
sed -i -e 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'"${PROFF}"'"/g' "${HOME_DIR}"/config/config.toml
sed -i -e 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'"${P2P}"'"/g' "${HOME_DIR}"/config/config.toml

# GRPC
sed -i -e 's/address = "localhost:9090"/address = "0.0.0.0:'"${GRPC}"'"/g' "${HOME_DIR}"/config/app.toml
sed -i -e 's/address = "localhost:9091"/address = "0.0.0.0:'"${GRPC_WEB}"'"/g' "${HOME_DIR}"/config/app.toml

# Rosetta Api
sed -i -e 's/address = ":8080"/address = "0.0.0.0:'"${ROSETTA}"'"/g' "${HOME_DIR}"/config/app.toml

# Faster blocks
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "'"${BLOCK_TIME}"'"/g' "${HOME_DIR}"/config/config.toml

# Start the daemon in the background
${BINARY} start --pruning=nothing --minimum-gas-prices=0"${DENOM}" --rpc.laddr="tcp://0.0.0.0:${RPC}" --home "${HOME_DIR}"
