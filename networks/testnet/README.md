# Sonr Testnet

Production-ready 3-validator testnet with validator-sentry architecture and Cloudflare tunnel integration via DockFlare.

## Overview

This testnet provides:
- **3 Validators** with isolated private networks
- **3 Sentry Nodes** for public access and DDoS protection
- **IPFS Node** for distributed storage
- **Cloudflare Tunnels** for secure, zero-configuration public endpoints
- **14 Public Endpoints** via `*.sonr.land` domains

## Quick Start

### Prerequisites

1. **Docker & Docker Compose**
   ```bash
   docker --version
   docker compose version
   ```

2. **Devbox** (provides snrd binary via Nix)
   ```bash
   # Auto-install via Makefile
   make all

   # Or install manually
   curl -fsSL https://get.jetpack.io/devbox | bash
   ```

3. **DockFlare** (for Cloudflare tunnels)
   ```bash
   docker network create cloudflare-net
   docker run -d \
     --name dockflare \
     --restart unless-stopped \
     --network cloudflare-net \
     -v /var/run/docker.sock:/var/run/docker.sock \
     -e CLOUDFLARE_API_TOKEN=your_token_here \
     -e CLOUDFLARE_ACCOUNT_ID=your_account_id \
     -e CLOUDFLARE_ZONE_ID=your_zone_id \
     ghcr.io/sonr-io/dockflare:latest
   ```

   > Get credentials from: https://dash.cloudflare.com

### Setup

```bash
# 1. Clone repository
git clone https://github.com/sonr-io/testnet
cd testnet

# 2. Install devbox (if not already installed)
make all

# 3. Create environment configuration
make setup
# Optional: Edit .env to customize

# 4. Initialize testnet
make init

# 5. Start testnet
make start

# 6. Verify endpoints
make test
```

**That's it!** Your testnet is running with Cloudflare tunnels.

---

## Commands

All testnet operations are managed via `make`:

```bash
make all           # Check/install devbox
make setup         # Create .env from template
make init          # Initialize validators and sentries
make start         # Start testnet
make stop          # Stop testnet
make restart       # Restart testnet
make clean         # Clean all data (WARNING: destructive)
make status        # Show status and endpoints
make logs          # View logs
make test          # Run basic tests
make help          # Show available commands
```

> **Note:** All commands use devbox under the hood, which provides the `snrd` binary via Nix.

---

## Public Endpoints

All services are accessible via Cloudflare tunnels (no port conflicts):

### Sentry Endpoints

**Alice:**
- RPC: `https://alice-rpc.sonr.land`
- REST: `https://alice-rest.sonr.land`
- gRPC: `https://alice-grpc.sonr.land`
- EVM: `https://alice-evm.sonr.land`

**Bob:**
- RPC: `https://bob-rpc.sonr.land`
- REST: `https://bob-rest.sonr.land`
- gRPC: `https://bob-grpc.sonr.land`
- EVM: `https://bob-evm.sonr.land`

**Carol:**
- RPC: `https://carol-rpc.sonr.land`
- REST: `https://carol-rest.sonr.land`
- gRPC: `https://carol-grpc.sonr.land`
- EVM: `https://carol-evm.sonr.land`

### IPFS Endpoints

- API: `https://ipfs-api.sonr.land`
- Gateway: `https://ipfs-gateway.sonr.land`

### Example Usage

```bash
# Query blockchain status
curl https://alice-rpc.sonr.land/status | jq

# Query account balance
curl https://alice-rest.sonr.land/cosmos/bank/v1beta1/balances/idx16wx7ye3ce060tjvmmpu8lm0ak5xr7gm2vjyh4k | jq

# Send transaction
snrd tx bank send alice idx1... 1000000usnr \
  --node https://alice-rpc.sonr.land \
  --chain-id sonrtest_1-1 \
  --keyring-backend test \
  --yes

# IPFS operations
curl -X POST https://ipfs-api.sonr.land/api/v0/version | jq
curl https://ipfs-gateway.sonr.land/ipfs/<CID>
```

---

## Architecture

```
                    Cloudflare Tunnel (DockFlare)
                              │
                              ▼
       ┌──────────────────────────────────────────────┐
       │              net-public + cloudflare-net      │
       │                                               │
       │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────┐
       │  │sentry-alice  │◄─►sentry-bob    │◄─►sentry-carol  │  │   IPFS   │
       │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┘
       └─────────┼──────────────────┼──────────────────┼──────────────────────┘
                 │                  │                  │
                 │ Private          │ Private          │ Private
                 │                  │                  │
       ┌─────────▼────────┐  ┌──────▼────────┐  ┌─────▼─────────┐
       │   val-alice      │  │   val-bob     │  │  val-carol    │
       │  (net-alice)     │  │  (net-bob)    │  │ (net-carol)   │
       └──────────────────┘  └───────────────┘  └───────────────┘
```

### Key Features

- **Validator Isolation**: Each validator on private network
- **Sentry Protection**: Public traffic filtered through sentries
- **No Port Conflicts**: All services via Cloudflare tunnels
- **Auto-Discovery**: DockFlare automatically creates tunnels
- **TLS Encryption**: All traffic encrypted via Cloudflare

---

## Genesis Accounts

| Name | Address | Balance | Purpose |
|------|---------|---------|---------|
| Alice | `idx140fehngcrxvhdt84x729p3f0qmkmea8n570lrg` | 100M SNR | Validator |
| Bob | `idx1r6yue0vuyj9m7xw78npspt9drq2tmtvgcrf7sr` | 100M SNR | Validator |
| Carol | `idx1pe9mc2q72u94sn2gg52ramrt26x5efw6kslflg` | 100M SNR | Validator |
| Faucet | `idx16wx7ye3ce060tjvmmpu8lm0ak5xr7gm2vjyh4k` | 250M SNR | Faucet |

**Total Genesis Supply:** 550M SNR (300M staked + 250M faucet)

---

## Configuration

Configuration is managed via `.env` file:

```bash
cp .env.example .env
```

Quick setup:
```bash
make setup  # Creates .env from .env.example
```

**Key Variables:**
- `CHAIN_ID=sonrtest_1-1` - Chain identifier
- `DENOM=usnr` - Native token denomination
- `BLOCK_TIME=5s` - Target block time
- `VOTING_PERIOD=30s` - Governance voting period
- `ALICE_MNEMONIC`, `BOB_MNEMONIC`, `CAROL_MNEMONIC` - Validator keys
- `FAUCET_MNEMONIC` - Faucet account key
- `DOCKER_IMAGE=onsonr/snrd:latest` - Container image

**⚠️ WARNING:** Default mnemonics in `.env.example` are public. Generate new ones for production!

For complete configuration details, see [docs/Environment.md](docs/Environment.md)

---

## Documentation

### Detailed Guides

- **[Docker.md](docs/Docker.md)** - Complete container documentation
  - All 7 containers explained
  - Network topology and security
  - Volume management
  - Health checks and troubleshooting
  - Performance tuning
  - Backup procedures

- **[Cloudflare.md](docs/Cloudflare.md)** - DockFlare integration guide
  - Complete domain mapping (14 endpoints)
  - DockFlare setup and configuration
  - DNS management
  - Testing all endpoint types
  - Access policies
  - Troubleshooting tunnels

- **[Environment.md](docs/Environment.md)** - Environment variable reference
  - Global configuration variables
  - Container-specific variables
  - DockFlare environment variables
  - Runtime overrides
  - Security best practices

- **[Architecture.md](docs/Architecture.md)** - Architecture deep dive
- **[CLAUDE.md](CLAUDE.md)** - Quick reference for AI assistants

---

## Troubleshooting

### Initialization Takes 1-2 Minutes

This is normal. The script initializes 6 nodes, creates genesis, and configures peer connections.

**Expected output:**
- 📋 Initializing validators
- 🛡️ Initializing sentries
- 💰 Adding genesis accounts
- 🔗 Setting up peer connections
- ✅ Completion with validator addresses

**If init fails:**
```bash
make clean
make init
```

### Validators Not Syncing

```bash
# Check logs
docker compose logs val-alice

# Verify peer connections
docker exec val-alice snrd tendermint show-node-id --home /root/.sonr
```

### Cloudflare Tunnels Not Working

Check DockFlare logs:
```bash
docker logs dockflare
```

Common issues:
- Missing `cloudflare-net` network
- Invalid API token
- Container not on `cloudflare-net`

See [docs/Cloudflare.md#troubleshooting](docs/Cloudflare.md#troubleshooting) for detailed help.

### IPFS Port Conflict

If IPFS fails with "port 8080 already allocated":
1. Stop the conflicting service
2. Or modify `docker-compose.yml` to use different port

---

## Production Deployment

**Before deploying to production:**

1. ✅ **Generate new mnemonics** - Never use defaults
   ```bash
   snrd keys add test --keyring-backend test --output json | jq -r .mnemonic
   ```

2. ✅ **Use KMS** for validator key management (e.g., tmkms)

3. ✅ **Configure firewall rules** to restrict validator access

4. ✅ **Enable monitoring** (Prometheus, Grafana)

5. ✅ **Set up backups** of validator keys and state
   ```bash
   tar -czf backup.tar.gz val-* sentry-* .env
   ```

6. ✅ **Rotate Cloudflare API tokens** regularly (every 90 days)

7. ✅ **Enable rate limiting** in Cloudflare dashboard

8. ✅ **Use persistent volumes** instead of bind mounts (optional)

---

## Development

### View Logs

```bash
# All services
make logs

# Specific service
docker compose logs -f sentry-alice

# Search logs
docker compose logs val-alice | grep -i error
```

### Execute Commands

```bash
# Via docker compose (recommended)
docker compose exec sentry-alice snrd query bank total

# Direct docker exec
docker exec -it sentry-alice snrd keys list --keyring-backend test
```

### Clean Restart

```bash
make clean   # Remove all data (destructive!)
make init    # Reinitialize
make start   # Start fresh
```

---

## Testing

```bash
# Run basic tests
make test
```

### Manual Testing

```bash
# RPC
curl https://alice-rpc.sonr.land/status | jq

# REST
curl https://alice-rest.sonr.land/cosmos/base/tendermint/v1beta1/node_info | jq

# EVM JSON-RPC
curl -X POST https://alice-evm.sonr.land \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' | jq

# IPFS
curl -X POST https://ipfs-api.sonr.land/api/v0/version | jq
```

---

## Support

- **Documentation**: See [docs/](docs/) directory
- **Issues**: https://github.com/sonr-io/testnet/issues
- **Discord**: https://discord.gg/sonr

---

## License

Apache 2.0
