# Docker Containers

This document describes all Docker containers used in the Sonr testnet architecture.

## Overview

The testnet consists of 7 containers organized in a validator-sentry architecture:
- **3 Validators** (private, isolated networks)
- **3 Sentry Nodes** (public-facing, Cloudflare tunnel integration)
- **1 IPFS Node** (distributed storage)

All containers use the `unless-stopped` restart policy for resilience.

---

## Container Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Cloudflare Tunnel (DockFlare)              │
└────────────────────┬────────────────────────────────────┘
                     │
    ┌────────────────┴────────────────────┐
    │        net-public + cloudflare-net  │
    │                                     │
    │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────┐
    │  │ sentry-  │  │ sentry-  │  │ sentry-  │  │ IPFS │
    │  │  alice   │◄─►  bob    │◄─► carol   │  │      │
    │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └──────┘
    └───────┼─────────────┼─────────────┼─────────────────┘
            │             │             │
            │ private     │ private     │ private
            │             │             │
    ┌───────▼──────┐ ┌────▼──────┐ ┌────▼──────┐
    │  val-alice   │ │ val-bob   │ │ val-carol │
    │ (net-alice)  │ │ (net-bob) │ │(net-carol)│
    └──────────────┘ └───────────┘ └───────────┘
```

---

## Validator Containers

Validators are the core consensus nodes that produce blocks and maintain blockchain state. They are isolated on private networks for security.

### val-alice

**Purpose:** Alice's validator node
**Image:** `onsonr/snrd:latest`
**Network:** `net-alice` (private)
**Volume:** `./val-alice:/root/.sonr`

**Configuration:**
- **Chain ID:** `sonrtest_1-1`
- **Moniker:** `val-alice`
- **Pruning:** Disabled (`--pruning=nothing`)
- **Min Gas Prices:** `0usnr` (free for testnet)

**Security Features:**
- Runs on isolated private network
- Only connects to its own sentry (`sentry-alice`)
- No direct public access
- Protected by sentry's `private_peer_ids` configuration

**Data Directory:**
```
./val-alice/
├── config/
│   ├── genesis.json       # Genesis state
│   ├── config.toml        # Tendermint config
│   ├── app.toml          # Application config
│   └── priv_validator_key.json  # Validator signing key (CRITICAL!)
└── data/                  # Blockchain data
```

---

### val-bob

**Purpose:** Bob's validator node
**Image:** `onsonr/snrd:latest`
**Network:** `net-bob` (private)
**Volume:** `./val-bob:/root/.sonr`

**Configuration:**
- **Chain ID:** `sonrtest_1-1`
- **Moniker:** `val-bob`
- **Pruning:** Disabled
- **Min Gas Prices:** `0usnr`

**Security:** Same as val-alice, isolated on `net-bob`

---

### val-carol

**Purpose:** Carol's validator node
**Image:** `onsonr/snrd:latest`
**Network:** `net-carol` (private)
**Volume:** `./val-carol:/root/.sonr`

**Configuration:**
- **Chain ID:** `sonrtest_1-1`
- **Moniker:** `val-carol`
- **Pruning:** Disabled
- **Min Gas Prices:** `0usnr`

**Security:** Same as val-alice, isolated on `net-carol`

---

## Sentry Containers

Sentry nodes act as a protective layer between validators and the public internet. They handle all public RPC, REST, gRPC, and EVM requests while shielding validator identities.

### sentry-alice

**Purpose:** Alice's public-facing sentry node
**Image:** `onsonr/snrd:latest`
**Networks:**
- `net-alice` (connects to val-alice)
- `net-public` (connects to other sentries)
- `cloudflare-net` (for DockFlare tunnels)

**Volume:** `./sentry-alice:/root/.sonr`
**Depends On:** `val-alice`

**Configuration:**
- **Chain ID:** `sonrtest_1-1`
- **Moniker:** `sentry-alice`
- **Pruning:** Disabled
- **Min Gas Prices:** `0usnr`

**Exposed Services:**
- **RPC:** Port 26657 (Tendermint RPC)
- **REST:** Port 1317 (Cosmos REST API)
- **gRPC:** Port 9090 (gRPC endpoint)
- **EVM JSON-RPC:** Port 8545 (Ethereum-compatible)
- **EVM WebSocket:** Port 8546 (WebSocket subscriptions)

**JSON-RPC APIs Enabled:**
- `eth` - Ethereum JSON-RPC
- `txpool` - Transaction pool inspection
- `personal` - Account management
- `net` - Network info
- `debug` - Debugging APIs
- `web3` - Web3 utilities

**Cloudflare Tunnel Endpoints:**

| Endpoint | Hostname | Internal Service | Description |
|----------|----------|------------------|-------------|
| RPC | `alice-rpc.sonr.land` | `http://sentry-alice:26657` | Tendermint RPC for queries and transactions |
| REST | `alice-rest.sonr.land` | `http://sentry-alice:1317` | Cosmos REST API (LCD) |
| gRPC | `alice-grpc.sonr.land` | `http://sentry-alice:9090` | gRPC for efficient binary communication |
| EVM | `alice-evm.sonr.land` | `http://sentry-alice:8545` | EVM JSON-RPC (MetaMask compatible) |

**Peer Connections:**
- **Persistent Peer:** `val-alice` (private connection)
- **Seeds:** `sentry-bob`, `sentry-carol` (public P2P)
- **Private Peer IDs:** Marks `val-alice` as private to hide from network

**Security Features:**
- Shields validator from direct exposure
- Handles all public-facing traffic
- Rate limiting and DDoS protection via Cloudflare
- TLS encryption via Cloudflare tunnels

---

### sentry-bob

**Purpose:** Bob's public-facing sentry node
**Image:** `onsonr/snrd:latest`
**Networks:** `net-bob`, `net-public`, `cloudflare-net`
**Volume:** `./sentry-bob:/root/.sonr`
**Depends On:** `val-bob`

**Configuration:** Same as sentry-alice, with moniker `sentry-bob`

**Cloudflare Tunnel Endpoints:**
- `bob-rpc.sonr.land` → RPC (26657)
- `bob-rest.sonr.land` → REST (1317)
- `bob-grpc.sonr.land` → gRPC (9090)
- `bob-evm.sonr.land` → EVM JSON-RPC (8545)

**Peer Connections:**
- **Persistent Peer:** `val-bob`
- **Seeds:** `sentry-alice`, `sentry-carol`

---

### sentry-carol

**Purpose:** Carol's public-facing sentry node
**Image:** `onsonr/snrd:latest`
**Networks:** `net-carol`, `net-public`, `cloudflare-net`
**Volume:** `./sentry-carol:/root/.sonr`
**Depends On:** `val-carol`

**Configuration:** Same as sentry-alice, with moniker `sentry-carol`

**Cloudflare Tunnel Endpoints:**
- `carol-rpc.sonr.land` → RPC (26657)
- `carol-rest.sonr.land` → REST (1317)
- `carol-grpc.sonr.land` → gRPC (9090)
- `carol-evm.sonr.land` → EVM JSON-RPC (8545)

**Peer Connections:**
- **Persistent Peer:** `val-carol`
- **Seeds:** `sentry-alice`, `sentry-bob`

---

## IPFS Container

### ipfsctl

**Purpose:** Distributed file storage and content addressing
**Image:** `ipfs/kubo:latest` (official IPFS implementation)
**Networks:**
- `net-public` (connects to sentry nodes)
- `cloudflare-net` (for DockFlare tunnels)

**Volume:** `ipfs-data:/data/ipfs` (named volume)

**Configuration:**
- **IPFS Profile:** `server` (optimized for server deployment)

**Exposed Services:**
- **API:** Port 5001 (IPFS HTTP API)
- **Gateway:** Port 8080 (IPFS Gateway for content retrieval)
- **Swarm:** Port 4001 (P2P networking)

**Cloudflare Tunnel Endpoints:**

| Endpoint | Hostname | Internal Service | Description |
|----------|----------|------------------|-------------|
| API | `ipfs-api.sonr.land` | `http://ipfsctl:5001` | IPFS API for adding/pinning content |
| Gateway | `ipfs-gateway.sonr.land` | `http://ipfsctl:8080` | HTTP gateway for retrieving content |

**Features:**
- **Content-Addressed Storage:** Files identified by cryptographic hash (CID)
- **Distributed Network:** Connects to global IPFS network
- **Persistent Storage:** Data stored in named volume
- **HTTP API:** RESTful API for programmatic access
- **Gateway Access:** Retrieve content via HTTP using CID

**Common Operations:**

```bash
# Add file to IPFS
curl -X POST -F file=@myfile.txt https://ipfs-api.sonr.land/api/v0/add

# Retrieve file by CID
curl https://ipfs-gateway.sonr.land/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/readme

# Check version
curl -X POST https://ipfs-api.sonr.land/api/v0/version | jq
```

---

## Network Topology

### Private Networks

Each validator has its own isolated network:

| Network | Purpose | Containers |
|---------|---------|------------|
| `net-alice` | Alice validator isolation | `val-alice`, `sentry-alice` |
| `net-bob` | Bob validator isolation | `val-bob`, `sentry-bob` |
| `net-carol` | Carol validator isolation | `val-carol`, `sentry-carol` |

**Security Benefits:**
- Validators cannot directly communicate with each other
- Prevents Byzantine attacks at network layer
- Forces all communication through sentries

### Public Network

| Network | Purpose | Containers |
|---------|---------|------------|
| `net-public` | Sentry interconnection | `sentry-alice`, `sentry-bob`, `sentry-carol`, `ipfsctl` |

**Purpose:**
- Sentries discover and connect to each other
- IPFS accessible to all sentries
- Public P2P gossip network

### Cloudflare Network

| Network | Purpose | Containers |
|---------|---------|------------|
| `cloudflare-net` | DockFlare tunnel access | `sentry-alice`, `sentry-bob`, `sentry-carol`, `ipfsctl` |

**Type:** External (must be created before starting stack)
**Purpose:** Allows DockFlare to discover and tunnel traffic to containers

**Setup:**
```bash
docker network create cloudflare-net
```

---

## Volume Management

### Bind Mounts (Validators & Sentries)

```
./val-alice:/root/.sonr
./val-bob:/root/.sonr
./val-carol:/root/.sonr
./sentry-alice:/root/.sonr
./sentry-bob:/root/.sonr
./sentry-carol:/root/.sonr
```

**Advantages:**
- Direct file system access from host
- Easy backup and inspection
- No permission issues (files owned by host user)

**Backup:**
```bash
tar -czf testnet-backup.tar.gz val-* sentry-*
```

### Named Volume (IPFS)

```
ipfs-data:/data/ipfs
```

**Advantages:**
- Managed by Docker
- Better performance on some systems
- Automatic cleanup with `docker compose down -v`

**Backup:**
```bash
docker run --rm -v ipfs-data:/data -v $(pwd):/backup alpine tar czf /backup/ipfs-backup.tar.gz /data
```

---

## Container Management

### Start All Containers
```bash
make start
# or
docker compose up -d
```

### Stop All Containers
```bash
make stop
# or
docker compose down
```

### View Container Logs
```bash
# All containers
docker compose logs -f

# Specific container
docker compose logs -f sentry-alice

# Last 100 lines
docker compose logs --tail=100 val-alice
```

### Execute Commands in Containers
```bash
# Via bootstrap script
devbox run exec sentry-alice status

# Via docker compose
docker compose exec sentry-alice snrd status --home /root/.sonr

# Via docker
docker exec -it sentry-alice snrd keys list --keyring-backend test
```

### Inspect Container Configuration
```bash
# View environment variables
docker inspect sentry-alice -f '{{json .Config.Env}}' | jq

# View networks
docker inspect sentry-alice -f '{{json .NetworkSettings.Networks}}' | jq

# View mounts
docker inspect sentry-alice -f '{{json .Mounts}}' | jq
```

### Resource Usage
```bash
# Real-time stats
docker stats

# Container resource limits (if set)
docker inspect sentry-alice -f '{{json .HostConfig.Memory}}' | jq
```

---

## Health Checks

### Container Status
```bash
docker ps --filter "name=val-" --filter "name=sentry-" --filter "name=ipfs"
```

### Node Sync Status
```bash
# Check if nodes are syncing
for container in sentry-alice sentry-bob sentry-carol; do
  echo "=== $container ==="
  docker exec $container sh -c 'curl -s http://localhost:26657/status | jq -r ".result.sync_info.catching_up"'
done
```

### Block Height
```bash
# Current block height
docker exec sentry-alice sh -c 'curl -s http://localhost:26657/status | jq -r ".result.sync_info.latest_block_height"'
```

### IPFS Health
```bash
# Check IPFS daemon
curl -X POST https://ipfs-api.sonr.land/api/v0/version | jq

# Check connected peers
curl -X POST https://ipfs-api.sonr.land/api/v0/swarm/peers | jq
```

---

## Security Considerations

### Validator Security
1. **Never expose validators directly** - Always use sentries
2. **Backup `priv_validator_key.json`** - Cannot be recovered if lost
3. **Monitor validator uptime** - Downtime results in slashing
4. **Use KMS in production** - Hardware security modules for key management

### Sentry Security
1. **Rate limiting via Cloudflare** - Protects against DDoS
2. **TLS encryption** - All traffic encrypted via Cloudflare tunnels
3. **No exposed ports** - All access via tunnels only
4. **Regular updates** - Keep `onsonr/snrd` image updated

### IPFS Security
1. **Content verification** - All content verified by CID
2. **No private data** - IPFS is a public network
3. **Pin important content** - Prevent garbage collection
4. **Monitor storage** - IPFS can grow large over time

---

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker compose logs <container-name>

# Check resource usage
docker stats

# Verify networks exist
docker network ls | grep -E "net-|cloudflare"

# Recreate container
docker compose up -d --force-recreate <container-name>
```

### Permission Issues
```bash
# Fix ownership (if needed)
sudo chown -R $USER:$USER val-* sentry-*

# Check bind mount permissions
ls -la val-alice/
```

### Network Connectivity Issues
```bash
# Test connectivity between containers
docker exec sentry-alice ping -c 3 val-alice
docker exec sentry-alice ping -c 3 sentry-bob

# Check network configuration
docker network inspect net-public
```

### IPFS Issues
```bash
# Check IPFS daemon status
docker compose logs ipfsctl

# Restart IPFS
docker compose restart ipfsctl

# Clear IPFS cache (WARNING: destructive)
docker compose down
docker volume rm testnet_ipfs-data
docker compose up -d
```

---

## Performance Tuning

### Resource Limits (Optional)

Add to `docker-compose.yml`:

```yaml
services:
  val-alice:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
```

### Logging Configuration

Prevent log bloat:

```yaml
services:
  val-alice:
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"
```

---

## Maintenance

### Update Containers
```bash
# Pull latest images
docker compose pull

# Restart with new images
docker compose up -d
```

### Clean Up
```bash
# Remove stopped containers
docker compose down

# Remove all data (WARNING: destructive)
docker compose down -v

# Clean Docker system
docker system prune -a
```

### Backup Procedure
```bash
# Stop containers
make stop

# Backup data
tar -czf testnet-backup-$(date +%Y%m%d).tar.gz val-* sentry-* .env

# Backup IPFS volume
docker run --rm -v testnet_ipfs-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/ipfs-backup-$(date +%Y%m%d).tar.gz /data

# Restart containers
make start
```
