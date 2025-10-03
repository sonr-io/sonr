#!/usr/bin/env bash

set -e

GO_MOD_PACKAGE="github.com/sonr-io/sonr"

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find . -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in ${proto_dirs}; do
	for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
		# this regex checks if a proto file has its go_package set to github.com/strangelove-ventures/poa/...
		# gogo proto files SHOULD ONLY be generated if this is false
		# we don't want gogo proto to run for proto files which are natively built for google.golang.org/protobuf
		if grep -q "option go_package" "${file}" && grep -H -o -c "option go_package.*${GO_MOD_PACKAGE}/api" "${file}" | grep -q ':0$'; then
			buf generate --template buf.gen.gogo.yaml "${file}"
		fi
	done
done

echo "Generating pulsar proto code"
buf generate --template buf.gen.pulsar.yaml

cd ..

cp -r "${GO_MOD_PACKAGE}"/* ./
rm -rf github.com

# Copy files over for dep injection
rm -rf api && mkdir api
custom_modules=$(find . -name 'module' -type d -not -path "./proto/*" -not -path "./.cache/*" -not -path "./node_modules/*" -not -path "./packages/*")

# get the 1 up directory (so ./cosmos/mint/module becomes ./cosmos/mint)
# remove the relative path starter from base namespaces. so ./cosmos/mint becomes cosmos/mint
base_namespace=$(echo "${custom_modules}" | sed -e 's|/module||g' | sed -e 's|\./||g')

# echo "Base namespace: $base_namespace"
for module in ${base_namespace}; do
	echo " [+] Moving: ./${module} to ./api/${module}"

	mkdir -p api/"${module}"

	mv "${module}"/* ./api/"${module}"/

	# # incorrect reference to the module for coins
	find api/"${module}" -type f -name '*.go' -exec sed -i -e 's|types "github.com/cosmos/cosmos-sdk/types"|types "cosmossdk.io/api/cosmos/base/v1beta1"|g' {} \;
	find api/"${module}" -type f -name '*.go' -exec sed -i -e 's|types1 "github.com/cosmos/cosmos-sdk/x/bank/types"|types1 "cosmossdk.io/api/cosmos/bank/v1beta1"|g' {} \;

	rm -rf "${module}"
done
