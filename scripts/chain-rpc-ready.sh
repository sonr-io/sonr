#!/bin/bash
# chain-rpc-ready.sh - Check if a CometBFT or Tendermint RPC service is ready
# Usage: chain-rpc-ready.sh [RPC_URL]

set -euo pipefail

RPC_URL=${1:-"http://localhost:26657"}

echo 1>&2 "Checking if $RPC_URL is ready..."

# Check if the RPC URL is reachable,
json=$(curl -s --connect-timeout 2 "$RPC_URL/status")

# and the bootstrap block state has been validated,
if [ "$(echo "$json" | jq -r '.result.sync_info | (.earliest_block_height < .latest_block_height)')" != true ]; then
  echo 1>&2 "$RPC_URL is not ready: bootstrap block state has not been validated"
  exit 1
fi

# and the node is not catching up.
if [ "$(echo "$json" | jq -r .result.sync_info.catching_up)" != false ]; then
  echo 1>&2 "$RPC_URL is not ready: node is catching up"
  exit 1
fi

echo "$json" | jq -r .result
exit 0
