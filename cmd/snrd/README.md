# snrd - Sonr Blockchain Daemon

`snrd` is the command-line interface and daemon for the Sonr blockchain network. It provides a comprehensive set of tools for running nodes, interacting with the network, managing identities, and performing blockchain operations.

## Overview

Sonr is a Cosmos SDK-based blockchain that combines decentralized identity (DID), WebAuthn authentication, and IPFS storage capabilities. The `snrd` daemon serves as the primary interface for:

- Running blockchain nodes (validators, full nodes)
- Managing decentralized identities and credentials
- Performing wallet operations and transaction signing
- Interacting with IPFS for decentralized storage
- Querying blockchain state and submitting transactions

## Installation

### Prerequisites

- Go 1.24+
- Docker (for IPFS operations)
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/sonr-io/sonr.git
cd sonr

# Install the binary
make install

# Verify installation
snrd version
```

## Architecture

The `snrd` binary is built on the Cosmos SDK v0.50.13 and includes:

- **Cosmos SDK modules**: Standard blockchain functionality (auth, bank, staking, etc.)
- **Custom modules**:
  - `x/did`: W3C DID specification with WebAuthn support
  - `x/dwn`: Decentralized Web Node for data storage
  - `x/svc`: Service management and domain verification
  - `x/ucan`: User-Controlled Authorization Networks
- **EVM compatibility**: Ethereum Virtual Machine support via Evmos
- **IPFS integration**: Decentralized storage for large data objects

### Network Configuration

- **Address prefix**: `idx` (Bech32 encoded addresses)
- **Coin type**: 60 (BIP44 - Ethereum compatible)
- **Default node home**: `~/.snrd`
- **Minimum gas prices**: `0stake`

## Usage

### Node Operations

#### Initialize a New Node

```bash
# Initialize node configuration
snrd init <moniker> --chain-id <chain-id>

# Example for testnet
snrd init my-node --chain-id sonr-testnet-1
```

#### Start the Node

```bash
# Start the blockchain node
snrd start

# Start with custom configuration
snrd start --home ~/.sonr --p2p.laddr tcp://0.0.0.0:26656
```

#### Node Management

```bash
# Check node status
snrd status

# View node information
snrd query block

# Export application state
snrd export

# Reset node data
snrd reset
```

### Identity Management

#### Authentication Commands

```bash
# Register a new WebAuthn credential
snrd auth register --username <username>

# Login with existing credentials
snrd auth login
```

The authentication system uses WebAuthn for passwordless, cryptographically secure user authentication. The registration process opens a browser window for biometric or security key authentication.

### Wallet Operations

#### Transaction Management

```bash
# Sign a transaction
snrd wallet sign <tx-file>

# Verify a signature
snrd wallet verify <signature> <message>

# Simulate transaction execution
snrd wallet simulate <tx-file>

# Broadcast a signed transaction
snrd wallet broadcast <signed-tx-file>
```

### IPFS Integration

#### IPFS Node Management

```bash
# Start IPFS containers
snrd ipfs start

# View IPFS logs with interactive interface
snrd ipfs logs

# Stop IPFS containers
snrd ipfs stop
```

The IPFS integration provides decentralized storage capabilities for large data objects referenced by blockchain transactions.

### Querying the Network

#### Basic Queries

```bash
# Query account information
snrd query auth account <address>

# Query account balance
snrd query bank balances <address>

# Query validator information
snrd query staking validator <validator-address>

# Query transaction by hash
snrd query tx <hash>
```

#### Module-Specific Queries

```bash
# Query DID documents
snrd query did did-document <did>

# Query WebAuthn credentials
snrd query did webauthn-credentials <address>

# Query service records
snrd query svc service <service-id>

# Query UCAN capabilities
snrd query ucan capability <capability-id>
```

### Transaction Commands

#### Standard Transactions

```bash
# Send tokens
snrd tx bank send <from-address> <to-address> <amount>

# Delegate to validator
snrd tx staking delegate <validator-address> <amount>

# Submit governance proposal
snrd tx gov submit-proposal <proposal-file>
```

#### Custom Module Transactions

```bash
# Register a DID document
snrd tx did register-did <did-document-file>

# Register WebAuthn credential
snrd tx did register-webauthn-credential <credential-file>

# Create service record
snrd tx svc create-service <service-config>

# Issue UCAN capability
snrd tx ucan issue-capability <capability-config>
```

## Configuration

### Node Configuration

Node configuration is stored in `~/.snrd/config/`:

- `config.toml`: Node and P2P configuration
- `app.toml`: Application-specific settings
- `client.toml`: Client configuration
- `genesis.json`: Genesis state

#### Key Configuration Options

```toml
# config.toml
[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = ""
max_num_inbound_peers = 40
max_num_outbound_peers = 10

[consensus]
timeout_commit = "5s"
timeout_propose = "3s"

# app.toml
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"

[json-rpc]
enable = true
address = "0.0.0.0:8545"
```

### Environment Variables

```bash
# Override default home directory
export SNRD_HOME=/path/to/custom/home

# Set custom chain ID
export SNRD_CHAIN_ID=custom-chain-1

# Configure logging level
export SNRD_LOG_LEVEL=info
```

## Development

### Building

```bash
# Build binary
make build

# Build with race detection
make build-race

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 make build
```

### Testing

```bash
# Run unit tests
make test

# Run tests with coverage
make test-cover

# Run integration tests
make test-integration
```

### Code Generation

```bash
# Generate protobuf code
make proto-gen

# Generate swagger documentation
make swagger-gen

# Format code
make format

# Run linter
make lint
```

## Troubleshooting

### Common Issues

#### Node Won't Start

```bash
# Check configuration
snrd validate-genesis

# Reset corrupted data
snrd unsafe-reset-all

# Check logs
snrd start --log_level debug
```

#### Connection Issues

```bash
# Test network connectivity
snrd status

# Check peer connections
snrd query tendermint-validator-set
```

#### IPFS Issues

```bash
# Check Docker status
docker ps

# Reset IPFS containers
snrd ipfs stop
docker system prune
snrd ipfs start
```

### Debugging Commands

```bash
# Enable debug logging
snrd start --log_level debug --log_format json

# Profile performance
snrd start --cpu-profile cpu.prof

# Memory profiling
snrd start --mem-profile mem.prof
```

## Network Endpoints

### Testnet

- **RPC**: `https://testnet-rpc.sonr.network`
- **REST**: `https://testnet-api.sonr.network`
- **gRPC**: `testnet-grpc.sonr.network:443`
- **Chain ID**: `sonr-testnet-1`

### Mainnet

- **RPC**: `https://rpc.sonr.network`
- **REST**: `https://api.sonr.network`
- **gRPC**: `grpc.sonr.network:443`
- **Chain ID**: `sonr-mainnet-1`

## API Documentation

- **REST API**: Available at `/swagger/` endpoint when API server is running
- **gRPC**: Protocol buffer definitions in `/proto` directory
- **GraphQL**: Available at `/graphql` endpoint (if enabled)

## Security Considerations

### Key Management

- Private keys are stored in the OS keyring by default
- Hardware wallet support available via Ledger integration
- WebAuthn provides passwordless authentication with biometric security

### Network Security

- TLS encryption for all API endpoints
- P2P encryption via Tendermint
- Signature verification for all transactions

### Best Practices

- Regular backups of validator keys and node data
- Use hardware security modules (HSMs) for validator keys
- Enable firewall rules for production deployments
- Monitor node health and connectivity

## Support

- **Documentation**: [https://docs.sonr.network](https://docs.sonr.network)
- **GitHub Issues**: [https://github.com/sonr-io/sonr/issues](https://github.com/sonr-io/sonr/issues)
- **Discord**: [https://discord.gg/sonr](https://discord.gg/sonr)
- **Telegram**: [https://t.me/sonrnetwork](https://t.me/sonrnetwork)

## License

Licensed under the Apache License 2.0. See [LICENSE](../../LICENSE) for details.
