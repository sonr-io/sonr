#!/bin/bash

set -euxo pipefail

mkdir -p /tmp/chains $UPGRADE_DIR

echo "Fetching code from tag"
mkdir -p /tmp/chains/$CHAIN_NAME
cd /tmp/chains/$CHAIN_NAME

if [[ $CODE_TAG =~ ^[0-9a-fA-F]{40}$ ]]; then
  echo "Trying to fetch code from commit hash"
  curl -LO $CODE_REPO/archive/$CODE_TAG.zip
  unzip $CODE_TAG.zip
  code_dir=${CODE_REPO##*/}-${CODE_TAG}
elif [[ $CODE_TAG = v* ]]; then
  echo "Trying to fetch code from tag with 'v' prefix"
  curl -LO $CODE_REPO/archive/refs/tags/$CODE_TAG.zip
  unzip $CODE_TAG.zip
  code_dir=${CODE_REPO##*/}-${CODE_TAG#"v"}
else
  echo "Trying to fetch code from tag or branch"
  if curl -fsLO $CODE_REPO/archive/refs/tags/$CODE_TAG.zip; then
    unzip $CODE_TAG.zip
    code_dir=${CODE_REPO##*/}-$CODE_TAG
  elif curl -fsLO $CODE_REPO/archive/refs/heads/$CODE_TAG.zip; then
    unzip $(echo $CODE_TAG | rev | cut -d "/" -f 1 | rev).zip
    code_dir=${CODE_REPO##*/}-${CODE_TAG/\//-}
  else
    echo "Tag or branch '$CODE_TAG' not found"
    exit 1
  fi
fi

echo "Fetch wasmvm if needed"
cd /tmp/chains/$CHAIN_NAME/$code_dir
WASM_VERSION=$(cat go.mod | grep -oe "github.com/CosmWasm/wasmvm v[0-9.]*" | cut -d ' ' -f 2)
if [[ WASM_VERSION != "" ]]; then
  mkdir -p /tmp/chains/libwasmvm_muslc
  cd /tmp/chains/libwasmvm_muslc
  curl -LO https://github.com/CosmWasm/wasmvm/releases/download/$WASM_VERSION/libwasmvm_muslc.x86_64.a
  cp libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a
fi

echo "Build chain binary"
cd /tmp/chains/$CHAIN_NAME/$code_dir
CGO_ENABLED=1 BUILD_TAGS="muslc linkstatic" LINK_STATICALLY=true LEDGER_ENABLED=false make install

echo "Copy created binary to the upgrade directories"
if [[ $UPGRADE_NAME == "genesis" ]]; then
  mkdir -p $UPGRADE_DIR/genesis/bin
  cp $GOBIN/$CHAIN_BIN $UPGRADE_DIR/genesis/bin
else
  mkdir -p $UPGRADE_DIR/upgrades/$UPGRADE_NAME/bin
  cp $GOBIN/$CHAIN_BIN $UPGRADE_DIR/upgrades/$UPGRADE_NAME/bin
fi

echo "Cleanup"
rm -rf /tmp/chains/$CHAIN_NAME
