# Sonr Starship Network Configurations

This directory contains Starship configurations for deploying Sonr blockchain networks in Kubernetes environments.

## Networks

### 🛠️ Devnet (`chains/devnet/`)

**Single-node development network for testing and development**

- **Purpose**: Local development and testing
- **Validators**: 1 node
- **Chain ID**: `sonr-1`
- **Features**: Basic faucet, explorer, minimal resources
- **Use case**: Development, debugging, feature testing

### 🌐 Testnet (`chains/testnet/`)

**Multi-node production-ready testnet**

- **Purpose**: Public testnet for user testing and staging
- **Validators**: 4 nodes
- **Chain ID**: `sonr-1`
- **Features**: Production faucet, explorer, ingress, SSL certificates
- **Use case**: User testing, staging, pre-production validation
- **Domain**: `*.sonr.land` with Cloudflare SSL

## Prerequisites

### Install Devbox

```bash
curl -fsSL https://get.jetify.com/devbox | bash
```

### Install Required Tools

Devbox will automatically install:

- `starship` - Kubernetes deployment tool
- `kubectl` - Kubernetes CLI
- `helm` - Kubernetes package manager

## Quick Start

### Start Devnet (Single Node)

```bash
# Start devnet (default)
make start

# Or explicitly
make start NETWORK=devnet
```

### Start Testnet (4 Nodes + Production Features)

```bash
make start NETWORK=testnet
```

### Stop Networks

```bash
# Stop current network
make stop

# Stop specific network
make stop NETWORK=testnet
```

### Restart Networks

```bash
# Restart current network
make restart

# Restart specific network
make restart NETWORK=testnet
```

## Network Configuration Details

### Devnet Configuration

- **Image**: `ghcr.io/sonr-io/snrd:latest`
- **Validators**: 1
- **Resources**: Minimal (for local development)
- **Faucet**: Basic CosmJS faucet
- **Explorer**: Ping Pub on port 8080

### Testnet Configuration

- **Image**: `ghcr.io/sonr-io/snrd:latest`
- **Validators**: 4 (production consensus)
- **Resources**: 2 CPU, 4GB RAM per validator
- **Faucet**: Production CosmJS faucet (5 concurrency, 1GB RAM)
- **Explorer**: Ping Pub on port 8080
- **Ingress**: Nginx with `*.sonr.land` domain and Cloudflare SSL
- **Security**: Pod security contexts for Kubernetes permissions

## Chain Parameters

Both networks use Sonr's custom configuration:

- **Base Denom**: `snr`
- **Micro Denom**: `usnr` (primary)
- **Account Allocation**: `100000000000000000000000000000000usnr` (massive amounts for custom DefaultPowerReduction)
- **Minimum Validator Stake**: Meets Sonr's custom `DefaultPowerReduction` of ~275 billion

## Port Mapping

When networks are running, the following ports are forwarded:

- **RPC**: `26657` - Tendermint RPC
- **REST**: `1317` - Cosmos REST API
- **gRPC**: `9090` - Cosmos gRPC
- **Faucet**: `8001` (devnet) / `8001` (testnet)
- **Explorer**: `8080` - Ping Pub interface

## Scripts

Both networks include custom initialization scripts:

- `scripts/create-genesis.sh` - Creates genesis with proper validator amounts
- `scripts/update-genesis.sh` - Updates genesis parameters for Sonr modules

## Development Workflow

### Local Development

```bash
# Start devnet for development
make start

# Access services
curl http://localhost:26657/status
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
open http://localhost:8080  # Explorer
```

### Testing with Testnet

```bash
# Start production-like testnet
make start NETWORK=testnet

# Test with multiple validators
# Access via same ports or ingress (if configured)
```

### Clean Up

```bash
# Stop network
make stop

# Or stop specific network
make stop NETWORK=testnet
```

## Troubleshooting

### Check Network Status

```bash
kubectl get pods
kubectl logs devnet-genesis-0  # or testnet-genesis-0
```

### Common Issues

1. **Permission Errors**: Both configs include `podSecurityContext` to handle volume permissions
2. **Validator Start Issues**: Large coin amounts ensure validators meet minimum delegation requirements
3. **Resource Constraints**: Testnet requires more resources than devnet

## Extending

To add a new network:

1. Create `chains/newnetwork/` directory
2. Add `config.yaml` with Starship configuration
3. Add `devbox.json` with required tools
4. Create custom scripts in `scripts/` if needed
5. Use `make start NETWORK=newnetwork`

## Documentation

- [Starship Documentation](https://docs.cosmology.zone/starship)
- [Cosmos SDK Documentation](https://docs.cosmos.network)
- [Sonr Documentation](https://sonr.dev)

