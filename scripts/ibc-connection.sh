#!/bin/bash

REGISTRY_URL="$1"
CHAIN_1="$2"
CHAIN_2="$3"

set -eux

function connection_id() {
  CONNECTION_ID=$(curl -s $REGISTRY_URL/ibc/$CHAIN_1/$CHAIN_2 | jq -r ".chain_1.connection_id")
  echo $CONNECTION_ID
}

echo "Try to get connection id, if failed, wait for 2 seconds and try again"
max_tries=20
while [[ max_tries -gt 0 ]]
do
  id=$(connection_id)
  if [[ -n "$id" ]]; then
    echo "Found connection id: $id"
    exit 0
  fi
  echo "Failed to get connection id. Sleeping for 2 secs. Tries left $max_tries"
  ((max_tries--))
  sleep 10
done
