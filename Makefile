#!/usr/bin/make -f

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app

# Fetch from env
VERSION ?= $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT ?= $(shell git log -1 --format='%H')
OS ?= $(shell uname -s)
ROOT ?= $(shell git rev-parse --show-toplevel)

# for dockerized protobuf tools
DOCKER := $(shell which docker)
HTTPS_GIT := github.com/sonr-io/snrd.git

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=sonr \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=snrd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath

# The below include contains the tools and runsim targets.
all: install lint test

build: go.sum
ifeq ($(OS),Windows_NT)
	$(error wasmd server not supported. Use "make build-windows-client" for client)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o bin/snrd .
endif

build-windows-client: go.sum
	GOOS=windows GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o bin/snrd.exe .

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o bin/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o bin/contract_tests ./cmd/contract_tests
endif

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) .

########################################
### Tools & dependencies
########################################

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz@latest
	@goviz -i . -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf .aider*
	rm -rf static
	rm -rf .out
	rm -rf hway.db
	rm -rf snapcraft-local.yaml bin/
	rm -rf build

distclean: clean
	rm -rf vendor/

########################################
### Testing
########################################

test: test-unit
test-all: test-race test-cover test-system

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -mod=readonly -bench=. ./...

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestFullAppSimulation

test-sim-deterministic: runsim
	@echo "Running application deterministic simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 1 1 TestAppStateDeterminism

test-system: install
	$(MAKE) -C tests/system/ test

###############################################################################
###                                Linting                                  ###
###############################################################################

format-tools:
	go install mvdan.cc/gofumpt@v0.4.0
	go install github.com/client9/misspell/cmd/misspell@v0.3.4
	go install github.com/daixiang0/gci@v0.11.2

lint: format-tools
	golangci-lint run --tests=false
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "*_test.go" | xargs gofumpt -d

format: format-tools
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofumpt -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gci write --skip-generated -s standard -s default -s "prefix(cosmossdk.io)" -s "prefix(github.com/cosmos/cosmos-sdk)" -s "prefix(github.com/CosmWasm/wasmd)" --custom-order

mod-tidy:
	go mod tidy

.PHONY: format-tools lint format mod-tidy


###############################################################################
###                                Protobuf                                 ###
###############################################################################
protoVer=0.15.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-gen:
	@echo "Generating Protobuf files"
	@go install cosmossdk.io/orm/cmd/protoc-gen-go-cosmos-orm@latest
	@$(protoImage) sh ./scripts/protocgen.sh
	spawn stub-gen

proto-format:
	@echo "Formatting Protobuf files"
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=master

.PHONY: all install install-debug \
	go-mod-cache draw-deps clean build format \
	test test-all test-build test-cover test-unit test-race \
	test-sim-import-export build-windows-client \
	test-system

## --- Testnet Utilities ---
get-localic:
	@echo "Installing local-interchain"
	git clone --branch v8.7.0 https://github.com/strangelove-ventures/interchaintest.git interchaintest-downloader
	cd interchaintest-downloader/local-interchain && make install
	@echo ✅ local-interchain installed $(shell which local-ic)

is-localic-installed:
ifeq (,$(shell which local-ic))
	make get-localic
endif

get-heighliner:
	git clone https://github.com/strangelove-ventures/heighliner.git
	cd heighliner && go install

local-image:
ifeq (,$(shell which heighliner))
	echo 'heighliner' binary not found. Consider running `make get-heighliner`
else
	heighliner build -c snrd --local -f chains.yaml
endif

.PHONY: get-heighliner local-image is-localic-installed

###############################################################################
###                                     e2e                                 ###
###############################################################################

ictest-basic:
	@echo "Running basic interchain tests"
	@cd interchaintest && go test -race -v -run TestBasicChain .

ictest-ibc:
	@echo "Running IBC interchain tests"
	@cd interchaintest && go test -race -v -run TestIBC .

ictest-wasm:
	@echo "Running cosmwasm interchain tests"
	@cd interchaintest && go test -race -v -run TestCosmWasmIntegration .

ictest-packetforward:
	@echo "Running packet forward middleware interchain tests"
	@cd interchaintest && go test -race -v -run TestPacketForwardMiddleware .

ictest-poa:
	@echo "Running proof of authority interchain tests"
	@cd interchaintest && go test -race -v -run TestPOA .

ictest-tokenfactory:
	@echo "Running token factory interchain tests"
	@cd interchaintest && go test -race -v -run TestTokenFactory .

###############################################################################
###                                    testnet                              ###
###############################################################################

setup-ipfs:
	./scripts/ipfs_config.sh

setup-testnet: mod-tidy is-localic-installed install local-image set-testnet-configs setup-testnet-keys

# Run this before testnet keys are added
# chainid-1 is used in the testnet.json
set-testnet-configs:
	snrd config set client chain-id sonr-testnet-1
	snrd config set client keyring-backend test
	snrd config set client output text

# import keys from testnet.json into test keyring
setup-testnet-keys:
	-`echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | snrd keys add acc0 --recover`
	-`echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | snrd keys add acc1 --recover`

# default testnet is with IBC
testnet: setup-testnet
	spawn local-ic start ibc-testnet

testnet-basic: setup-testnet
	spawn local-ic start testnet

sh-testnet: mod-tidy
	CHAIN_ID="sonr-testnet-1" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_node.sh

.PHONY: setup-testnet set-testnet-configs testnet testnet-basic sh-testnet dop-testnet

###############################################################################
###                                    extra utils                          ###
###############################################################################
status:
	@gh run ls -L 3
	@gum format -- "# Sonr ($OS-$VERSION)" "- ($(COMMIT)) $ROOT" "- $(RELEASE_DATE)"
	@sleep 3

push-docker:
	@docker build -t ghcr.io/onsonr/sonr:$(VERSION) .
	@docker tag ghcr.io/onsonr/sonr:$(VERSION) ghcr.io/onsonr/sonr:latest
	@docker push ghcr.io/onsonr/sonr:$(VERSION)
	@docker push ghcr.io/onsonr/sonr:latest

release:
	@devbox run cz:bump

release-dry:
	@devbox run release:dry

deploy-deps:
	@echo "Installing deploy dependencies"
	npm install -g @starship-ci/cli
	starship install

up:
	@echo "Starting deployment"
	starship start --config .github/deploy/config.yml

down:
	@echo "Stopping deployment"
	starship stop --config .github/deploy/config.yml

###############################################################################
###                                     help                                ###
###############################################################################

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@echo "  install             : Install the binary"
	@echo "  local-image         : Install the docker image"
	@echo "  proto-gen           : Generate code from proto files"
	@echo "  testnet             : Local devnet with IBC"
	@echo "  sh-testnet          : Shell local devnet"
	@echo "  ictest-basic        : Basic end-to-end test"
	@echo "  ictest-ibc          : IBC end-to-end test"

.PHONY: help
