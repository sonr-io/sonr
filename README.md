
[![Go Reference](https://pkg.go.dev/badge/github.com/sonr-io/sonr.svg)](https://pkg.go.dev/github.com/sonr-io/sonr)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/sonr-io/sonr)
[![Static Badge](https://img.shields.io/badge/homepage-sonr.io-blue?style=flat-square)](https://sonr.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonr-io/sonr)](https://goreportcard.com/report/github.com/sonr-io/sonr)

[![Sonr](.github/banner.png)](https://sonr.io)

> **Sonr is a blockchain ecosystem combining decentralized identity, secure data storage, and multi-chain interoperability. Built on Cosmos SDK v0.50.14, it provides users with self-sovereign identity through W3C DIDs, WebAuthn authentication, and personal data vaults—all without requiring cryptocurrency for onboarding.**

## 💡 Key Features

### 🔐 Gasless Onboarding

Create your first decentralized identity without owning cryptocurrency:

```bash
# Register with WebAuthn (no tokens required!)
snrd auth register --username alice

# Register with automatic vault creation
snrd auth register --username bob --auto-vault
```

### 🌐 Multi-Chain Support

- **Cosmos SDK**: Native integration with IBC ecosystem
- **EVM Compatibility**: Ethereum smart contract support
- **External Wallets**: MetaMask, Keplr, and more

### 🔑 Advanced Authentication

- **WebAuthn/Passkeys**: Biometric authentication
- **Hardware Security Keys**: YubiKey, Titan Key support
- **Multi-Signature**: Multiple verification methods per DID

### 📦 Decentralized Storage

- **IPFS Integration**: Distributed file storage
- **Encrypted Vaults**: Hardware-backed encryption
- **Protocol Schemas**: Structured data validation

## 📚 Technical Specifications

- **Cosmos SDK**: v0.50.14
- **CometBFT**: v0.38.17
- **IBC**: v8.7.0
- **Go**: 1.24.1 (toolchain 1.24.4)
- **WebAssembly**: CosmWasm v1.5.8
- **Task Queue**: Asynq (Redis-based)
- **Actor System**: Proto.Actor
- **Storage**: IPFS, LevelDB

## 🔒 Security

### Gasless Transaction Security

- Limited to WebAuthn registration only
- Full cryptographic validation required
- Credential uniqueness enforcement
- Anti-replay protection

### WebAuthn Security

- Origin validation
- Challenge-response authentication
- Device binding
- Attestation verification

### Multi-Algorithm Support

- Ed25519 (quantum-resistant)
- ECDSA (secp256k1, P-256)
- RSA (2048, 3072, 4096 bits)
- WebAuthn (ES256, RS256)

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/sonr-io/sonr
cd sonr

# Install the binary
make install

# Verify installation
snrd version
```

### Running a Local Node

```bash
# Start single-node testnet (quick iteration)
make localnet

# Start multi-node testnet with Starship
make testnet-start

# Stop testnet
make testnet-stop
```

### Building from Source

```bash
# Build blockchain node
make build

# Build Docker image
make docker

# Generate code from proto files
make proto-gen
```

### Local Development Network

```bash
# Standard localnet (auto-detects best method for your system)
make localnet     # Works on Arch Linux, Ubuntu, macOS, etc.

# Docker-based localnet (requires Docker)
make dockernet    # Runs in detached mode

# One-time setup for your system (optional)
./scripts/setup_localnet.sh  # Installs dependencies and configures environment
```

### Testing

```bash
# Run all tests
make test-all

# Module-specific tests
make test-did     # DID module tests
make test-dwn     # DWN module tests
make test-svc     # Service module tests

# E2E tests
make ictest-basic # Basic chain functionality
make ictest-ibc   # IBC transfers
make ictest-wasm  # CosmWasm integration

# Test with coverage
make test-cover
```

### Infrastructure

```bash
# IPFS for vault operations
make ipfs-up      # Start IPFS infrastructure
make ipfs-down    # Stop IPFS infrastructure
make ipfs-status  # Check IPFS connectivity
```

## 🏗️ Architecture

Sonr is a Cosmos SDK-based blockchain with integrated IPFS storage and three custom modules:

### Core Components

#### **Blockchain Node (`snrd`)**

The main blockchain daemon built with Cosmos SDK v0.50.14, providing:

- AutoCLI for command generation
- EVM compatibility via Evmos integration
- IBC for cross-chain communication
- CosmWasm smart contract support
- IPFS integration for decentralized storage

### Cross-Platform Support

The `localnet` target automatically detects and uses the best available method:
1. Checks for local binary (built with `make install`)
2. Falls back to Docker if available
3. Handles permission issues on systems like Arch Linux
4. Supports systemd service installation (see `etc/systemd/`)

## 📖 Module Documentation

### DID Module

W3C DID specification implementation with:

- **Gasless WebAuthn Registration**: Create DIDs without cryptocurrency
- **Multi-Algorithm Signatures**: Ed25519, ECDSA, RSA, WebAuthn
- **External Wallet Linking**: MetaMask, Keplr integration
- **Verifiable Credentials**: W3C-compliant credential issuance

```bash
# Create a DID
snrd tx did create-did did:sonr:alice '{"id":"did:sonr:alice",...}' --from alice

# Link external wallet
snrd tx did link-external-wallet did:sonr:alice \
  --wallet-address 0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42 \
  --wallet-type ethereum \
  --from alice

# Query DID
snrd query did resolve did:sonr:alice
```

[Full DID Module Documentation](x/did/README.md)

### DWN Module

Personal data stores with:

- **Structured Data Records**: Hierarchical data organization
- **Protocol-Based Interactions**: Enforceable data schemas
- **Secure Vaults**: Enclave-based key management
- **Multi-Chain Support**: Cosmos SDK and EVM transaction building

```bash
# Create a vault
snrd tx dwn create-vault --from alice

# Store a record
snrd tx dwn write-record '{"data":"...", "protocol":"example.com"}' --from alice

# Query records
snrd query dwn records --owner alice
```

[Full DWN Module Documentation](x/dwn/README.md)

### Service Module

Decentralized service registry featuring:

- **Domain Verification**: DNS-based ownership proof
- **Service Registration**: Verified service endpoints
- **Permission Management**: UCAN capability integration

```bash
# Verify domain ownership
snrd tx svc initiate-domain-verification example.com --from alice
snrd tx svc verify-domain example.com --from alice

# Register service
snrd tx svc register-service my-service example.com \
  --permissions "read,write" --from alice
```

[Full Service Module Documentation](x/svc/README.md)

## 🔧 Configuration

### Environment Variables

Environment variables can be configured via Docker Compose:

```bash
# Chain configuration
export CHAIN_ID="localchain_9000-1"
export BLOCK_TIME="1000ms"

# Network selection for Starship
export NETWORK="devnet"  # or "testnet"

# IPFS configuration
IPFS_API_URL=http://ipfs:5001
```

Environment variables can be set directly or via a `.env` file in the project root.

### Starship Configuration

Edit `starship.yml` to configure multi-node testnets:

```yaml
chains:
  - id: sonrtest_1-1
    name: custom
    numValidators: 3
    image: onsonr/snrd:latest
    # ... additional configuration
```

### Troubleshooting

**IPFS not accessible:**
```bash
# Verify IPFS is running
curl http://127.0.0.1:5001/api/v0/version

# Check IPFS status
make ipfs-status
```

**Port conflicts:**
- IPFS API: 5001
- IPFS Gateway: 8080
- Node gRPC: 9090
- Node REST API: 1317

Stop conflicting services or modify ports in configuration files.

## 🏗️ Project Structure

Sonr is a focused Cosmos SDK blockchain implementation:

```
sonr/
├── app/              # Application setup and module wiring
├── cmd/              # Binary entry points
│   └── snrd/        # Blockchain node daemon
├── x/               # Custom Cosmos SDK modules
│   ├── did/         # W3C DID implementation
│   ├── dwn/         # Decentralized Web Nodes
│   └── svc/         # Service management
├── types/           # Internal Go packages
├── proto/           # Protobuf definitions
├── scripts/         # Utility scripts
├── test/            # Integration tests
├── docs/            # Documentation
└── client/          # Client libraries and tooling
```

### Key Technologies

- **Cosmos SDK v0.50.14**: Blockchain framework
- **CometBFT v0.38.17**: Byzantine fault-tolerant consensus
- **IBC v8.7.0**: Inter-blockchain communication
- **CosmWasm v1.5.8**: Smart contract support
- **IPFS**: Decentralized storage integration

## 🤝 Community & Support

- [GitHub Discussions](https://github.com/sonr-io/sonr/discussions) - Community forum
- [GitHub Issues](https://github.com/sonr-io/sonr/issues) - Bug reports and feature requests
- [Twitter](https://sonr.io/twitter) - Latest updates
- [Documentation Wiki](https://github.com/sonr-io/sonr/wiki) - Detailed guides


## 📄 License

Copyright © 2024 Sonr, Inc.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

---

<p align="center">
  Built with ❤️ by the Sonr team
</p>
