# Sonr Smart Contracts

This directory contains the smart contracts for the Sonr blockchain, including the WSNR (Wrapped SNR) ERC-20 token contract.

## Setup

### Prerequisites

1. Install Foundry:

```bash
curl -L https://foundry.paradigm.xyz | bash
foundryup
```

2. Install dependencies:

```bash
make install
```

3. Copy environment variables:

```bash
cp .env.example .env
```

4. Configure your `.env` file with appropriate values

## Development

### Build contracts:

```bash
make build
```

### Run tests:

```bash
make test
```

### Run tests with gas reporting:

```bash
make test-gas
```

### Generate coverage report:

```bash
make coverage
```

### Format code:

```bash
make format
```

### Deploy to local Sonr network:

```bash
make deploy-local
```

### Deploy to Sonr testnet:

```bash
make deploy-testnet
```

### Clean build artifacts:

```bash
make clean
```

### Start local Anvil node (for testing):

```bash
make anvil
```

## Project Structure

```
contracts/
├── src/           # Contract source files
├── test/          # Test files
├── script/        # Deployment scripts
├── lib/           # Dependencies (gitignored)
├── foundry.toml   # Foundry configuration
└── Makefile       # Build commands
```

## Testing on Sonr Native EVM

Start the local testnet with EVM enabled:

```bash
# From project root
docker-compose -f docker-compose.dev.yml up -d sonr-node
```

The EVM JSON-RPC endpoint will be available at:

- HTTP: http://localhost:8545
- WebSocket: ws://localhost:8546

## Foundry Commands

For a full list of available commands:

```bash
make help
```

## Contract Addresses

- WSNR: TBD (after deployment)
