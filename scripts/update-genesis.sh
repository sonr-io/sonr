#!/bin/bash

DENOM="${DENOM:=usnr}"
CHAIN_BIN="${CHAIN_BIN:=snrd}"
CHAIN_DIR="${CHAIN_DIR:=$HOME/.sonr}"

set -eux

ls "$CHAIN_DIR"/config

echo "Update genesis.json file with updated local params"
sed -i -e "s/\"stake\"/\"$DENOM\"/g" "$CHAIN_DIR"/config/genesis.json
sed -i "s/\"time_iota_ms\": \".*\"/\"time_iota_ms\": \"$TIME_IOTA_MS\"/" "$CHAIN_DIR"/config/genesis.json

echo "Update max gas param"
jq -r '.consensus.params.block.max_gas |= "100000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json

echo "Update staking unbonding time and slashing jail time"
jq -r '.app_state.staking.params.unbonding_time |= "300s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
jq -r '.app_state.slashing.params.downtime_jail_duration |= "60s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json

# overrides for older sdk versions, before 0.47
function gov_overrides_sdk_v46() {
  jq -r '.app_state.gov.deposit_params.max_deposit_period |= "30s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.deposit_params.min_deposit[0].amount |= "10"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.voting_params.voting_period |= "30s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.tally_params.quorum |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.tally_params.threshold |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.tally_params.veto_threshold |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
}

# overrides for newer sdk versions, post 0.47
function gov_overrides_sdk_v47() {
  jq -r '.app_state.gov.params.max_deposit_period |= "30s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.params.min_deposit[0].amount |= "10"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.params.voting_period |= "30s"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.params.quorum |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.params.threshold |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.gov.params.veto_threshold |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
}

# EVM and feemarket configuration
if [ "$(jq -r '.app_state.evm' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r ".app_state.evm.params.evm_denom |= \"$DENOM\"" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.evm.params.active_static_precompiles |= ["0x0000000000000000000000000000000000000100","0x0000000000000000000000000000000000000400","0x0000000000000000000000000000000000000800","0x0000000000000000000000000000000000000801","0x0000000000000000000000000000000000000802","0x0000000000000000000000000000000000000803","0x0000000000000000000000000000000000000804","0x0000000000000000000000000000000000000805"]' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

if [ "$(jq -r '.app_state.erc20' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r '.app_state.erc20.params.native_precompiles |= ["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r ".app_state.erc20.token_pairs |= [{\"contract_owner\":1,\"erc20_address\":\"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE\",\"denom\":\"$DENOM\",\"enabled\":true}]" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

if [ "$(jq -r '.app_state.feemarket.params' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r '.app_state.feemarket.params.no_base_fee |= true' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.feemarket.params.base_fee |= "0.000000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

# Staking and mint configuration
if [ "$(jq -r '.app_state.staking' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r ".app_state.staking.params.bond_denom |= \"$DENOM\"" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.staking.params.min_commission_rate |= "0.050000000000000000"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

if [ "$(jq -r '.app_state.mint' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r ".app_state.mint.params.mint_denom |= \"$DENOM\"" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

if [ "$(jq -r '.app_state.crisis' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r ".app_state.crisis.constant_fee |= {\"denom\":\"$DENOM\",\"amount\":\"1000\"}" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

# Token factory configuration
if [ "$(jq -r '.app_state.tokenfactory' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r '.app_state.tokenfactory.params.denom_creation_fee |= []' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
  jq -r '.app_state.tokenfactory.params.denom_creation_gas_consume |= 100000' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

# ABCI configuration
if [ "$(jq -r '.consensus.params.abci' "$CHAIN_DIR"/config/genesis.json)" != "null" ]; then
  jq -r '.consensus.params.abci.vote_extensions_enable_height |= "1"' "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

# Add CONSTITUTION.md to governance if it exists
# Look for CONSTITUTION.md in the git root directory (parent of scripts directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CONSTITUTION_FILE="${GIT_ROOT}/CONSTITUTION.md"

if [ -f "$CONSTITUTION_FILE" ]; then
  echo "Adding CONSTITUTION.md to governance module from: $CONSTITUTION_FILE"
  CONSTITUTION_CONTENT=$(cat "$CONSTITUTION_FILE" | jq -Rs .)
  jq -r ".app_state.gov.constitution = $CONSTITUTION_CONTENT" "$CHAIN_DIR"/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json "$CHAIN_DIR"/config/genesis.json
fi

if [ "$(jq -r '.app_state.gov.params' "$CHAIN_DIR"/config/genesis.json)" == "null" ]; then
  gov_overrides_sdk_v46
else
  gov_overrides_sdk_v47
fi

$CHAIN_BIN tendermint show-node-id
