#!/bin/bash

ADDRESS="$1"
DENOM="$2"
FAUCET_URL="$3"
FAUCET_ENABLED="$4"

set -eux

function transfer_token() {
  status_code=$(curl --header "Content-Type: application/json" \
    --request POST --write-out %{http_code} --silent --output /dev/null \
    --data '{"denom":"'"$DENOM"'","address":"'"$ADDRESS"'"}' \
    $FAUCET_URL)
  echo $status_code
}

if [[ $FAUCET_ENABLED == "false" ]];
then
  echo "Faucet not enabled... skipping transfer token from faucet"
  exit 0
fi

echo "Try to send tokens, if failed, wait for 5 seconds and try again"
max_tries=5
while [[ max_tries -gt 0 ]]
do
  status_code=$(transfer_token)
  if [[ "$status_code" -eq 200 ]]; then
    echo "Successfully sent tokens"
    exit 0
  fi
  echo "Failed to send tokens. Sleeping for 2 secs. Tries left $max_tries"
  ((max_tries--))
  sleep 2
done
