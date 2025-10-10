#!/usr/bin/make -f

# Default target - show help when no target specified
.DEFAULT_GOAL := help

PACKAGE_NAME          := github.com/sonr-io/sonr
GORELEASER_VERSION    ?= latest
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
GOLANGCI_VERSION := v1.62.2
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
GIT_ROOT := $(shell git rev-parse --show-toplevel)

# Git configuration
HTTPS_GIT := github.com/sonr-io/sonr.git

export GO111MODULE = on

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags --always)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

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

# flags '-s -w' resolves an issue with xcode 16 and signing of go binaries
# ref: https://github.com/golang/go/issues/63997
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=sonr \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=snrd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -checklinkname=0 \
		  -s -w

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath


all: help

start: docker
	@gum log --level info "Starting all services..."
	@devbox services up

stop:
	@gum log --level info "Stopping all services..."
	@devbox services stop
	@$(MAKE) clean-docker

########################################
### Tools & dependencies
########################################
format: go-format
go-format:
	@gum log --level info "Formatting Go code with gofumpt and goimports..."
	@if command -v gofumpt > /dev/null; then \
		gofumpt -w .; \
	else \
		go run -mod=readonly mvdan.cc/gofumpt@latest -w .; \
	fi
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		go run -mod=readonly golang.org/x/tools/cmd/goimports@latest -w .; \
	fi
	@gum log --level info "✅ Code formatted"

lint: go-lint

go-lint:
	@gum log --level info "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=10m; \
	else \
		docker run --rm -v $$(pwd):/app -w /app \
			-v ~/.cache/golangci-lint:/root/.cache/golangci-lint \
			golangci/golangci-lint:$(GOLANGCI_VERSION) \
			golangci-lint run --timeout=10m; \
	fi

.PHONY: lint go-lint format

go-mod-cache: go.sum
	@gum log --level info "Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@gum log --level info "Ensure dependencies have not been modified"
	@go mod tidy
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz@latest
	@goviz -i ./cmd/snrd -d 2 | dot -Tpng -o .github/assets/dependency-graph.png


tidy:
	@go mod tidy
	@make -C client tidy

clean: tidy
	@gum log --level info "Cleaning build artifacts..."
	rm -rf snapcraft-local.yaml build/ dist/
	@$(MAKE) -C cmd/snrd clean

clean-docker:
	@gum log --level info "Removing all Docker volumes and networks..."
	@rm -rf .logs
	@docker compose down -v
	@docker network prune -f
	@docker volume prune -f

###############################################################################
###                              Build Targets                              ###
###############################################################################

install: go.sum
	@$(MAKE) -C cmd/snrd install LEDGER_ENABLED=$(LEDGER_ENABLED) WITH_CLEVELDB=$(WITH_CLEVELDB) LINK_STATICALLY=$(LINK_STATICALLY) BUILD_TAGS="$(BUILD_TAGS)"

build: go.sum
	@$(MAKE) -C cmd/snrd build LEDGER_ENABLED=$(LEDGER_ENABLED) WITH_CLEVELDB=$(WITH_CLEVELDB) LINK_STATICALLY=$(LINK_STATICALLY) BUILD_TAGS="$(BUILD_TAGS)"

build-snrd: build

build-client: go.sum
	@$(MAKE) -C client build
	@cd /tmp && go mod init test || true
	@cd /tmp && go get github.com/sonr-io/sonr/client@main || true
	@cd /tmp && gum log --level info "Client SDK import successful"

# Build all components in parallel
build-all: go.sum
	@gum log --level info "Building all components in parallel..."
	@$(MAKE) -j2 build build-client
	@gum log --level info "✅ All components built successfully"

.PHONY: install build build-client build-snrd build-all

########################################
### Docker & Services
########################################

docker:
	@gum log --level info "Building Docker images..."
	@bash scripts/containers.sh build-all

localnet: ## Cross-platform localnet (auto-detects best method for your system)
	@bash scripts/cross_platform_localnet.sh

dockernet:
	@gum log --level info "Starting network with Docker in detached mode..."
	@docker stop sonr-testnode 2>/dev/null || true
	@docker rm sonr-testnode 2>/dev/null || true
	@sleep 3
	@CHAIN_ID="sonrtest_1-1" BLOCK_TIME="1000ms" CLEAN=true FORCE_DOCKER=true DOCKER_DETACHED=true bash scripts/test_node.sh

.PHONY: docker localnet dockernet

########################################
### Prepare Scripts - AI & Release Automation
########################################
# Smart component release detection and automation
release:
	@$(MAKE) -C cmd/snrd release

# Snapshot builds for development
snapshot:
	@gum log --level info "📦 Preparing component snapshot..."
	@$(MAKE) -C cmd/snrd snapshot

.PHONY: release snapshot

########################################
### Testing - Simplified
########################################

# Main test targets
test: test-unit
test-all: test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock test' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock test' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock test' ./...

test-e2e:
	@gum log --level info "Running basic e2e tests"
	@cd test/e2e && go test -race -v -run TestBasic ./tests/basic

test-e2e-all:
	@gum log --level info "Running all e2e tests"
	@cd test/e2e && go test -race -v ./tests/...

test-build-snrd: build
	@ls -la build/snrd
	@chmod +x build/snrd
	@./build/snrd version

test-tdd:
	go test -json ./... 2>&1 | tdd-guard-go -project-root ${GIT_ROOT}

test-app:
	@VERSION=$(VERSION) CGO_LDFLAGS="-lm" go test -C . -mod=readonly -tags='ledger test_ledger_mock test' github.com/sonr-io/sonr/app/... github.com/sonr-io/sonr/x/... github.com/sonr-io/common/... github.com/sonr-io/sonr/internal/...

test-devops:
	@echo "No devops tests"

test-client:
	@$(MAKE) -C client test

test-dwn-ci:
	@go test -mod=readonly -tags='ledger test_ledger_mock test' -run='!IPFS' ./x/dwn/...

test-internal:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock test' ./internal/...

# Module testing - Simplified
# MODULE=did|dwn|svc VARIANT=unit|race|cover|bench
test-module:
	@if [ -z "$(MODULE)" ]; then \
		gum log --level info "Testing all modules..."; \
		$(MAKE) test-module MODULE=did; \
		$(MAKE) test-module MODULE=dwn; \
		$(MAKE) test-module MODULE=svc; \
	else \
		if [ "$(VARIANT)" = "cover" ]; then \
			gum log --level info "Testing x/$(MODULE) with coverage..."; \
			go test -mod=readonly -timeout 30m -race -coverprofile=x/$(MODULE)/coverage.txt -covermode=atomic -tags='ledger test_ledger_mock test' ./x/$(MODULE)/...; \
		elif [ "$(VARIANT)" = "race" ]; then \
			gum log --level info "Testing x/$(MODULE) with race detector..."; \
			VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock test' ./x/$(MODULE)/...; \
		elif [ "$(VARIANT)" = "bench" ]; then \
			gum log --level info "Running x/$(MODULE) benchmarks..."; \
			go test -mod=readonly -bench=. ./x/$(MODULE)/...; \
		else \
			gum log --level info "Testing x/$(MODULE) module..."; \
			VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock test' ./x/$(MODULE)/...; \
		fi \
	fi
test-proto:
	@$(MAKE) -C proto lint
	@$(MAKE) -C proto check-breaking

test-benchmark:
	@go test -mod=readonly -bench=. ./...

.PHONY: test test-all test-unit test-race test-cover test-tdd test-module test-benchmark

###############################################################################
###                                Protobuf                                 ###
###############################################################################
climd-gen:
	@gum log --level info "Generating MD Docs from snrd CLI..."
	@sh ./scripts/cli-docgen.sh

proto-gen:
	@gum log --level info "Generating Go protobuf files..."
	@$(MAKE) -C proto gen
	@gum log --level info "Auto-formatting generated protobuf files..."
	@$(MAKE) format

swagger-gen:
	@$(MAKE) -C proto swagger-gen
	@gum log --level info "Moving and renaming generated files..."
	@find docs/static/openapi -type f \( -name "query.swagger.yaml" -o -name "tx.swagger.yaml" \) | while read -r filepath; do \
		\
		parent_dir=$$(dirname "$$filepath"); \
		grandparent_dir=$$(dirname "$$parent_dir"); \
		module=$$(basename "$$grandparent_dir"); \
		filename=$$(basename "$$filepath"); \
		\
		new_filename="$$module.$$filename"; \
		destination="docs/static/openapi/$$new_filename"; \
		\
		gum log --level debug "Moving $$filepath to $$destination"; \
		mv "$$filepath" "$$destination"; \
	done
	@gum log --level info "Cleaning up empty source directories..."
	@find docs/static/openapi -mindepth 1 -maxdepth 1 -type d -exec rm -rf {} +
	@gum log --level info "✅ API documentation processing complete."

templ-gen:
	@docker run --rm -v `pwd`:/code -w=/code --user $(shell id -u):$(shell id -g) ghcr.io/a-h/templ:latest generate

.PHONY: proto-gen proto-swagger-gen swagger-gen proto-lint proto-check-breaking proto-publish

###############################################################################
###                           Network Operations                            ###
###############################################################################

# Starship network management
testnet: testnet-restart

testnet-restart: testnet-stop testnet-start
	@gum log --level info "✅ Starship network restarted"

testnet-start:
	@gum log --level info "Starting Starship network..."
	@if [ -z "$(NETWORK)" ]; then \
		NETWORK=devnet; \
	fi; \
	bash scripts/run.sh $$NETWORK

testnet-stop:
	@gum log --level info "Stopping Starship network..."
	@helm delete -n ci sonr-testnet 2>/dev/null || true
	@kubectl delete namespace ci --ignore-not-found=true 2>/dev/null || true
	@sleep 2

.PHONY: testnet testnet-restart testnet-start testnet-stop

###############################################################################
###                                   Help                                  ###
###############################################################################

help:
	@gum log --level info "Sonr Blockchain Makefile"
	@gum log --level info "========================"
	@gum log --level info ""
	@gum log --level info "🛠️  Build & Install:"
	@gum log --level info "  install             Install snrd binary"
	@gum log --level info "  build               Build snrd binary"
	@gum log --level info "  build-all           Build all components in parallel"
	@gum log --level info "  build-client        Build client SDK"
	@gum log --level info "  docker              Build Docker images"
	@gum log --level info ""
	@gum log --level info "📦 Release & Distribution:"
	@gum log --level info "  release             Create production release with GoReleaser"
	@gum log --level info "  snapshot            Create development snapshot builds"
	@gum log --level info ""
	@gum log --level info "🚀 Local Development:"
	@gum log --level info "  localnet            Start single-node testnet"
	@gum log --level info "  start               Start backend services"
	@gum log --level info "  stop                Stop backend services"
	@gum log --level info "  status              Check service health"
	@gum log --level info "  testnet             Manage Starship network (start/stop/restart)"
	@gum log --level info ""
	@gum log --level info "📦 Code Generation:"
	@gum log --level info "  proto-gen           Generate protobuf code"
	@gum log --level info "  swagger-gen         Generate OpenAPI docs"
	@gum log --level info ""
	@gum log --level info "🔧 Development Tools:"
	@gum log --level info "  format              Format code (Go + TypeScript)"
	@gum log --level info "  lint                Run all linters"
	@gum log --level info "  clean               Remove build artifacts"
	@gum log --level info ""
	@gum log --level info "🧪 Testing:"
	@gum log --level info "  benchmark           Run benchmarks"
	@gum log --level info "  test                Run unit tests"
	@gum log --level info "  test-all            Run all test variants"
	@gum log --level info "  test-cover          Generate coverage report"
	@gum log --level info "  test-e2e            Run e2e tests"
	@gum log --level info "  test-e2e-all        Run all e2e tests"
	@gum log --level info "  test-module         Test specific module (MODULE=did|dwn|svc)"
	@gum log --level info ""
	@gum log --level info "📚 Module Testing Examples:"
	@gum log --level info "  make test-module MODULE=did           # Test DID module"
	@gum log --level info "  make test-module MODULE=dwn VARIANT=cover  # DWN with coverage"
	@gum log --level info "  make test-module MODULE=svc VARIANT=race   # SVC with race detector"
	@gum log --level info "  make test-module MODULE=did VARIANT=bench  # DID benchmarks"
	@gum log --level info ""
	@gum log --level info "For more detailed options, see the Makefile source."

.PHONY: help release release-platform snapshot snapshot-platform
