# Environment Variables

This document describes all environment variables that can be configured for the Sonr testnet.

## Configuration File

All variables are defined in `.env` file in the repository root. Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

---

## Global Configuration

These variables are used during initialization and apply to all services.

### Chain Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAIN_ID` | `sonrtest_1-1` | Blockchain chain identifier |
| `DENOM` | `usnr` | Base denomination for the native token |
| `KEYRING` | `test` | Keyring backend (test, file, os) |
| `KEYALGO` | `eth_secp256k1` | Key algorithm for validator keys |

### Network Parameters

| Variable | Default | Description |
|----------|---------|-------------|
| `BLOCK_TIME` | `5s` | Target block time |
| `VOTING_PERIOD` | `30s` | Governance proposal voting period |
| `EXPEDITED_VOTING_PERIOD` | `15s` | Expedited proposal voting period |
| `MAX_GAS` | `100000000` | Maximum gas per block |
| `MIN_COMMISSION_RATE` | `0.050000000000000000` | Minimum validator commission rate (5%) |

### Validator Mnemonics

**⚠️ WARNING**: Change these for production! Default mnemonics are public.

| Variable | Description |
|----------|-------------|
| `ALICE_MNEMONIC` | 24-word mnemonic for Alice validator |
| `BOB_MNEMONIC` | 24-word mnemonic for Bob validator |
| `CAROL_MNEMONIC` | 24-word mnemonic for Carol validator |
| `FAUCET_MNEMONIC` | 24-word mnemonic for faucet account |

### Initial Balances

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_ALLOCATION` | `100000000000000000000000000usnr` | Base allocation per validator (100M SNR) |
| `STAKE_AMOUNT` | `30000000000000000000000usnr` | Initial stake per validator (30M SNR) |
| `FAUCET_ALLOCATION` | `250000000000000000000000000000usnr` | Faucet account balance (250M SNR) |

### Docker Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DOCKER_IMAGE` | `onsonr/snrd:latest` | Docker image for validator and sentry nodes |

---

## Container-Specific Variables

These variables are set in `docker-compose.yml` for each container.

### Validators (val-alice, val-bob, val-carol)

All validators share the same environment variables but with different values:

| Variable | Value | Description |
|----------|-------|-------------|
| `CHAIN_ID` | `sonrtest_1-1` | Chain identifier (same for all) |
| `MONIKER` | `val-{name}` | Validator node name (val-alice, val-bob, val-carol) |

**Command-line flags:**
- `--home /root/.sonr` - Node home directory
- `--pruning=nothing` - Disable state pruning
- `--minimum-gas-prices=0usnr` - Minimum gas price (0 for testnet)
- `--chain-id=sonrtest_1-1` - Chain identifier

**Networks:**
- Private network: `net-{name}` (net-alice, net-bob, net-carol)

**Volumes:**
- `./val-{name}:/root/.sonr` - Bind mount for node data

---

### Sentries (sentry-alice, sentry-bob, sentry-carol)

All sentries share the same environment variables but with different values:

| Variable | Value | Description |
|----------|-------|-------------|
| `CHAIN_ID` | `sonrtest_1-1` | Chain identifier (same for all) |
| `MONIKER` | `sentry-{name}` | Sentry node name (sentry-alice, sentry-bob, sentry-carol) |

**Command-line flags:**
- `--home /root/.sonr` - Node home directory
- `--pruning=nothing` - Disable state pruning
- `--minimum-gas-prices=0usnr` - Minimum gas price (0 for testnet)
- `--rpc.laddr=tcp://0.0.0.0:26657` - RPC listen address
- `--json-rpc.api=eth,txpool,personal,net,debug,web3` - Enabled JSON-RPC APIs
- `--json-rpc.address=0.0.0.0:8545` - EVM JSON-RPC address
- `--json-rpc.ws-address=0.0.0.0:8546` - EVM WebSocket address
- `--chain-id=sonrtest_1-1` - Chain identifier

**Networks:**
- Private network: `net-{name}` (connects to validator)
- `net-public` (connects to other sentries)
- `cloudflare-net` (for DockFlare tunnels)

**Volumes:**
- `./sentry-{name}:/root/.sonr` - Bind mount for node data

**DockFlare Labels (indexed):**

Each sentry exposes 4 endpoints via Cloudflare tunnels:

| Index | Hostname Pattern | Service | Description |
|-------|------------------|---------|-------------|
| `0` | `{name}-rpc.sonr.land` | `http://sentry-{name}:26657` | Tendermint RPC |
| `1` | `{name}-rest.sonr.land` | `http://sentry-{name}:1317` | Cosmos REST API |
| `2` | `{name}-grpc.sonr.land` | `http://sentry-{name}:9090` | gRPC endpoint |
| `3` | `{name}-evm.sonr.land` | `http://sentry-{name}:8545` | EVM JSON-RPC |

**Example for sentry-alice:**
```yaml
labels:
  - "dockflare.enable=true"
  - "dockflare.0.hostname=alice-rpc.sonr.land"
  - "dockflare.0.service=http://sentry-alice:26657"
  - "dockflare.1.hostname=alice-rest.sonr.land"
  - "dockflare.1.service=http://sentry-alice:1317"
  - "dockflare.2.hostname=alice-grpc.sonr.land"
  - "dockflare.2.service=http://sentry-alice:9090"
  - "dockflare.3.hostname=alice-evm.sonr.land"
  - "dockflare.3.service=http://sentry-alice:8545"
```

---

### IPFS (ipfsctl)

| Variable | Value | Description |
|----------|-------|-------------|
| `IPFS_PROFILE` | `server` | IPFS configuration profile for server mode |

**Networks:**
- `net-public` (connects to sentries)
- `cloudflare-net` (for DockFlare tunnels)

**Volumes:**
- `ipfs-data:/data/ipfs` - Named volume for IPFS data

**DockFlare Labels:**

| Index | Hostname | Service | Description |
|-------|----------|---------|-------------|
| `0` | `ipfs-api.sonr.land` | `http://ipfsctl:5001` | IPFS API endpoint |
| `1` | `ipfs-gateway.sonr.land` | `http://ipfsctl:8080` | IPFS Gateway |

---

## DockFlare Configuration

DockFlare requires the following environment variables (not in `.env`, set when running DockFlare container):

| Variable | Required | Description |
|----------|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | ✅ Yes | Cloudflare API token with Tunnel:Edit, DNS:Edit, Zone:Read permissions |
| `CLOUDFLARE_ACCOUNT_ID` | ⚠️ Recommended | Cloudflare account ID (found in dashboard) |
| `CLOUDFLARE_ZONE_ID` | ⚠️ Recommended | Cloudflare zone ID for domain (found in dashboard) |

**Example DockFlare Setup:**
```bash
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

---

## Runtime Overrides

You can override specific values at runtime by modifying `docker-compose.yml` or passing environment variables:

### Override Docker Image
```bash
export DOCKER_IMAGE=onsonr/snrd:v1.2.3
make start
```

### Override Chain ID
Edit `docker-compose.yml` and change all instances of:
```yaml
environment:
  - CHAIN_ID=your-custom-chain-id
```

And update command flags:
```yaml
command: >
  snrd start
  ...
  --chain-id=your-custom-chain-id
```

---

## Security Best Practices

1. **Never commit `.env` to version control** - It's already in `.gitignore`
2. **Generate new mnemonics for production**:
   ```bash
   snrd keys add test --keyring-backend test --output json | jq -r .mnemonic
   ```
3. **Use KMS for production validator keys** (e.g., tmkms)
4. **Rotate Cloudflare API tokens** regularly
5. **Use strong passwords** if switching from `test` keyring to `file` or `os`
6. **Backup mnemonics securely** - They cannot be recovered if lost

---

## Verification

After configuration, verify your setup:

```bash
# Check .env is loaded
make status

# Test endpoints
make test

# View container environment
docker inspect sentry-alice -f '{{json .Config.Env}}' | jq
```
