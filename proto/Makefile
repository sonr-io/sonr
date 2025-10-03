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

all: build gen swagger-gen push

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
swagger-gen:
	@gum log --level info "Generating OpenAPI documentation..."
	@buf generate --template buf.gen.openapi.yaml
	@gum log --level info "✅ OpenAPI documentation generated in docs/static/openapi/"
	@gum log --level info "✅ Merged API specification: docs/static/openapi/sonr.swagger.yaml"

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
	@gum log --level info "  all             Run lint, gen, and swagger-gen targets"
	@gum log --level info "  gen             Generate protobuf code and stubs"
	@gum log --level info "  lint            Lint protobuf files"
	@gum log --level info "  check-breaking  Check for breaking changes against main branch"
	@gum log --level info "  swagger-gen     Generate OpenAPI documentation"
	@gum log --level info "  help            Show this help message"
	@gum log --level info ""
	@gum log --level info "Environment:"
	@gum log --level info "  Proto builder: $(protoImageName)"
	@gum log --level info "  Spawn:         onsonr/spawn:latest"

.PHONY: all gen lint check-breaking swagger-gen help
