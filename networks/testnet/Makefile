# Sonr Testnet Makefile
.PHONY: help all init start stop restart clean status logs test setup

# Default target
.DEFAULT_GOAL := help

# Help target
help:
	@echo "Sonr Testnet Management (via Devbox)"
	@echo ""
	@echo "Setup:"
	@echo "  make all           - Check/install devbox"
	@echo ""
	@echo "Testnet Operations:"
	@echo "  make setup         - Create .env from template"
	@echo "  make init          - Initialize validators and sentries"
	@echo "  make start         - Start testnet"
	@echo "  make stop          - Stop testnet"
	@echo "  make restart       - Restart testnet"
	@echo "  make clean         - Clean all data"
	@echo "  make status        - Show status"
	@echo "  make logs          - View logs"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run basic tests"

# Check/install devbox
all:
	@command -v devbox >/dev/null 2>&1 || (echo "Installing devbox..." && curl -fsSL https://get.jetpack.io/devbox | bash)
	@echo "✅ devbox is ready"

# Setup .env file
setup:
	@devbox run setup

# Initialize validators
init:
	@devbox run init

# Start testnet
start:
	@devbox run start

# Stop testnet
stop:
	@devbox run stop

# Restart testnet
restart:
	@devbox run restart

# Clean all data
clean:
	@devbox run clean

# Show status
status:
	@devbox run status

# View logs
logs:
	@devbox run logs

# Run tests
test:
	@devbox run test