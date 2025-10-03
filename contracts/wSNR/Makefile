# Foundry Makefile for Sonr Smart Contracts

# Load environment variables
-include .env

# Default network
NETWORK ?= sonr

# Contract verification
VERIFIER ?= etherscan
VERIFIER_URL ?= https://api.etherscan.io/api

.PHONY: help
help: ## Display this help message
	@gum log --level info "Sonr Smart Contracts - Foundry Commands"
	@gum log --level info ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: install
install: ## Install Foundry and dependencies
	curl -L https://foundry.paradigm.xyz | bash
	foundryup
	forge install OpenZeppelin/openzeppelin-contracts@v5.0.0 --no-commit

.PHONY: build
build: ## Build contracts
	forge build

.PHONY: test
test: ## Run tests
	forge test -vvv

.PHONY: test-gas
test-gas: ## Run tests with gas reporting
	forge test -vvv --gas-report

.PHONY: coverage
coverage: ## Generate test coverage report
	forge coverage

.PHONY: format
format: ## Format Solidity code
	forge fmt

.PHONY: lint
lint: ## Lint Solidity code
	forge fmt --check

.PHONY: snapshot
snapshot: ## Create gas snapshot
	forge snapshot

.PHONY: clean
clean: ## Clean build artifacts
	forge clean

# Deployment commands
.PHONY: deploy-wsnr
deploy-wsnr: ## Deploy WSNR contract
	@gum log --level info "Deploying WSNR to $(NETWORK)..."
	forge script script/DeployWSNR.s.sol:DeployWSNR \
		--rpc-url $(NETWORK) \
		--broadcast \
		--verify \
		-vvvv

.PHONY: deploy-local
deploy-local: ## Deploy to local Sonr node
	@gum log --level info "Deploying to local Sonr node..."
	@cd .. && ./scripts/deploy-wsnr.sh sonr

.PHONY: deploy-testnet
deploy-testnet: ## Deploy to Sonr testnet
	@gum log --level info "Deploying to Sonr testnet..."
	@cd .. && ./scripts/deploy-wsnr.sh sonr-testnet

.PHONY: deploy
deploy: deploy-local ## Default deployment (alias for deploy-local)

# Interaction commands
.PHONY: console
console: ## Start Foundry console
	forge console --rpc-url $(NETWORK)

.PHONY: verify
verify: ## Verify contract on Etherscan
	forge verify-contract \
		--chain-id $(shell cast chain-id --rpc-url $(NETWORK)) \
		--etherscan-api-key $(ETHERSCAN_API_KEY) \
		--verifier $(VERIFIER) \
		$(CONTRACT_ADDRESS) \
		$(CONTRACT_NAME)

# Development helpers
.PHONY: anvil
anvil: ## Start local Anvil node
	anvil --chain-id 31337

.PHONY: cast-balance
cast-balance: ## Check ETH balance of an address
	@cast balance $(ADDRESS) --rpc-url $(NETWORK)

.PHONY: cast-send
cast-send: ## Send a transaction
	@cast send $(TO) --value $(VALUE) --rpc-url $(NETWORK) --private-key $(PRIVATE_KEY)