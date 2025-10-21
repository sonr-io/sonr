#!/usr/bin/make -f

# Docker configuration
DOCKER := $(shell which docker)
CURRENT_UID := $(shell id -u)
CURRENT_GID := $(shell id -g)

# Proto builder configuration
protoVer=0.13.2
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage="$(DOCKER)" run -e BUF_CACHE_DIR=/tmp/buf --rm -v "$(GIT_ROOT)":/workspace:rw --user ${CURRENT_UID}:${CURRENT_GID} --workdir /workspace $(protoImageName)

# Spawn configuration
spawnImage="$(DOCKER)" run --rm -v "$(GIT_ROOT)":/workspace:rw --user ${CURRENT_UID}:${CURRENT_GID} --workdir /workspace -e HOME=/workspace onsonr/spawn:latest

# Git configuration
HTTPS_GIT := github.com/sonr-io/sonr.git
GIT_ROOT := $(shell git rev-parse --show-toplevel)

###############################################################################
###                                Default Target                           ###
###############################################################################
.DEFAULT_GOAL := help

all: build gen openapi-gen push

proto: gen
generate: gen

###############################################################################
###                                Code Generation                          ###
###############################################################################
build:
	@gum log --level info "Building Protobuf files"
	@npx buf build

gen:
	@go install cosmossdk.io/orm/cmd/protoc-gen-go-cosmos-orm@v1.0.0-beta.3
	@gum log --level info "Generating Protobuf files"
	@$(protoImage) sh /workspace/scripts/protocgen.sh
	@gum log --level info "Generating stubs with spawn..."
	@$(spawnImage) spawn stub-gen
	@cd $(GIT_ROOT) && go mod tidy

###############################################################################
###                                Linting                                  ###
###############################################################################
lint:
	@$(protoImage) buf lint --error-format=json
test:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

###############################################################################
###                                Documentation                            ###
###############################################################################
openapi-gen:
	@gum log --level info "Generating OpenAPI documentation..."
	@$(protoImage) buf generate --template proto/buf.gen.openapi.yaml
	@gum log --level info "Combining OpenAPI files into unified specification..."
	@mkdir -p $(GIT_ROOT)/docs/static
	@npx swagger-combine $(GIT_ROOT)/client/docs/config.json \
		-o $(GIT_ROOT)/docs/static/openapi.yml \
		-f yaml \
		--continueOnConflictingPaths true \
		--includeDefinitions true
	@gum log --level info "Cleaning up temporary files..."
	@rm -rf $(GIT_ROOT)/tmp-openapi-gen
	@gum log --level info "✅ OpenAPI documentation generated successfully"
	@gum log --level info "✅ Unified API specification: docs/static/openapi.yml"

# Backwards compatibility alias
swagger-gen: openapi-gen

push:
	@npx buf push	

release:
	@gum log --level info "Releasing Protobuf files"
	@npx buf build
	@npx buf push

###############################################################################
###                                Help                                     ###
###############################################################################
help:
	@gum log --level info "Sonr Protocol Buffer Management"
	@gum log --level info ""
	@gum log --level info "Usage: make [target]"
	@gum log --level info ""
	@gum log --level info "Available targets:"
	@gum log --level info ""
	@gum log --level info "  all             Run lint, gen, and openapi-gen targets"
	@gum log --level info "  gen             Generate protobuf code and stubs"
	@gum log --level info "  lint            Lint protobuf files"
	@gum log --level info "  check-breaking  Check for breaking changes against main branch"
	@gum log --level info "  openapi-gen     Generate OpenAPI documentation"
	@gum log --level info "  help            Show this help message"
	@gum log --level info ""
	@gum log --level info "Environment:"
	@gum log --level info "  Proto builder: $(protoImageName)"
	@gum log --level info "  Spawn:         onsonr/spawn:latest"

.PHONY: all gen lint check-breaking openapi-gen swagger-gen help
