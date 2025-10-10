# Motor (motr) Migration Context

> **Repository Migration**: `sonr-io/sonr` → `sonr-io/motr`
> **Components Moved**: `cmd/motr/`, `cmd/vault/`, `packages/`, `web/`
> **Note**: The `crypto/` library was moved to its own repository at `sonr-io/crypto`

## Overview

Motor (formerly "motr") is a multi-purpose WebAssembly service that provides:
1. **Worker**: WASM-based cryptographic vault operations (formerly "vault")
2. **Payment Gateway**: W3C Payment Handler API compliant payment processing
3. **TypeScript SDK**: Browser and Node.js client libraries for vault and payment operations
4. **Web Applications**: Authentication and dashboard web apps

The name "Motor" reflects its role as the execution engine powering secure operations in the Sonr ecosystem.

## Repository Structure

```
sonr-io/motr/
├── worker/              # WASM vault operations (Go → WASM)
│   ├── main.go         # Entrypoint with WASM exports
│   ├── vault/          # Vault operation implementations
│   └── mpc/            # Multi-party computation
│
├── server/             # HTTP server mode (Go)
│   ├── main.go         # HTTP/Payment Gateway server
│   ├── handlers/       # Payment & OIDC handlers
│   └── middleware/     # Security & rate limiting
│
├── packages/           # TypeScript SDK and libraries
│   ├── es/            # @motr/es - Core SDK
│   │   ├── client/    # Vault client
│   │   ├── worker/    # Service worker integration
│   │   ├── plugin/    # Plugin system
│   │   └── codec/     # Encoding/signing utilities
│   │
│   ├── sdk/           # @motr/sdk - High-level SDK
│   ├── ui/            # @motr/ui - UI components
│   └── com/           # @motr/com - Common utilities
│
└── web/               # Web applications
    ├── auth/          # Authentication app
    └── dash/          # Dashboard app
```

## Dependencies

Motor depends on the **Sonr Cryptography Library** (`github.com/sonr-io/crypto`), which was moved to its own repository. See MIGRATE_CRYPTO.md for details on the crypto library.

## Component 1: Worker (WASM Vault)

**Technology**: Go 1.24.4 → WebAssembly via TinyGo
**Runtime**: Extism (WebAssembly plugin host)
**Purpose**: Secure cryptographic operations in sandboxed environment

### Architecture

```
┌──────────────────────────────────────────────┐
│          Host Application                     │
│     (Browser, Node.js, or Highway)           │
└───────────────┬──────────────────────────────┘
                │
                ▼
        ┌───────────────┐
        │    Extism     │ ◄─── WASM Runtime
        │   Runtime     │
        └───────┬───────┘
                │
                ▼
┌───────────────────────────────────────────────┐
│          worker.wasm (Motor Worker)           │
├───────────────────────────────────────────────┤
│                                               │
│  ┌──────────────┐  ┌──────────────┐         │
│  │   Vault      │  │     MPC      │         │
│  │ Operations   │  │   Enclave    │         │
│  └──────────────┘  └──────────────┘         │
│                                               │
│  ┌──────────────────────────────────┐       │
│  │     Crypto Primitives             │       │
│  │  (Ed25519, ECDSA, BLS, etc.)     │       │
│  └──────────────────────────────────┘       │
│                                               │
└───────────────────────────────────────────────┘
```

### WASM Exports (Go Functions)

All exported functions follow the Extism PDK pattern:

#### Core Vault Operations

```go
//go:wasmexport generate
func generate() int32
// Creates new MPC enclave with key generation
// Input: GenerateRequest{id: string}
// Output: GenerateResponse{data: EnclaveData, public_key: []byte}

//go:wasmexport refresh
func refresh() int32
// Refreshes enclave cryptographic material
// Input: RefreshRequest{enclave: EnclaveData}
// Output: RefreshResponse{okay: bool, data: EnclaveData}

//go:wasmexport sign
func sign() int32
// Signs arbitrary message with vault key
// Input: SignRequest{message: []byte, enclave: EnclaveData}
// Output: SignResponse{signature: []byte}

//go:wasmexport verify
func verify() int32
// Verifies signature against public key
// Input: VerifyRequest{public_key: []byte, message: []byte, signature: []byte}
// Output: VerifyResponse{valid: bool}
```

#### Multi-Chain Transaction Signing

```go
//go:wasmexport sign_cosmos_transaction
func signCosmosTransaction() int32
// Signs Cosmos SDK transaction
// Input: CosmosSignRequest{chain_id, account_number, sequence, tx_bytes}
// Output: SignedTransaction{signature, signed_doc}

//go:wasmexport sign_evm_transaction
func signEvmTransaction() int32
// Signs Ethereum/EVM transaction
// Input: EVMSignRequest{chain_id, nonce, tx_data, gas_limit}
// Output: SignedTransaction{v, r, s, raw_tx}

//go:wasmexport sign_message
func signMessage() int32
// Signs arbitrary message (EIP-191/EIP-712)
// Input: MessageSignRequest{message, encoding_type}
// Output: MessageSignature{signature, recovery_id}
```

#### Vault Import/Export (IPFS)

```go
//go:wasmexport export
func export() int32
// Exports encrypted vault to IPFS
// Input: ExportRequest{enclave: EnclaveData, password: []byte}
// Output: ExportResponse{cid: string, success: bool}

//go:wasmexport import
func import() int32
// Imports encrypted vault from IPFS
// Input: ImportRequest{cid: string, password: []byte}
// Output: ImportResponse{enclave: EnclaveData, success: bool}
```

#### WebAuthn Integration

```go
//go:wasmexport create_vault_enclave
func createVaultEnclave() int32
// Creates vault with WebAuthn configuration
// Input: VaultConfig{vault_id, webauthn_enabled, auto_lock_timeout}
// Output: VaultEnclave{enclave_data, webauthn_credentials}

//go:wasmexport unlock_vault
func unlockVault() int32
// Unlocks vault with WebAuthn or password
// Input: UnlockRequest{vault_id, auth_method, credentials}
// Output: UnlockResponse{success, session_token}

//go:wasmexport lock_vault
func lockVault() int32
// Locks vault and clears sensitive data
// Input: LockRequest{vault_id}
// Output: LockResponse{success}
```

#### Health & Monitoring

```go
//go:wasmexport get_vault_health
func getVaultHealth() int32
// Returns vault health status
// Output: EnclaveHealth{vault_id, status, last_activity, key_rotation_due}

//go:wasmexport get_version
func getVersion() int32
// Returns worker version and capabilities
// Output: VersionInfo{version, supported_chains, features}
```

### MPC Enclave Structure

```go
type EnclaveData struct {
    ID              string              `json:"id"`
    PublicKey       []byte              `json:"public_key"`
    PrivateKeyShare []byte              `json:"private_key_share"` // Encrypted
    Threshold       uint32              `json:"threshold"`
    Parties         uint32              `json:"parties"`
    ChainID         string              `json:"chain_id"`
    CreatedAt       int64               `json:"created_at"`
    LastRefresh     int64               `json:"last_refresh"`
    Metadata        map[string]string   `json:"metadata"`
}
```

**Note**: Motor relies on the Sonr Cryptography Library (`github.com/sonr-io/crypto`) for all cryptographic operations. The crypto library provides comprehensive primitives including Ed25519, ECDSA, BLS signatures, MPC, threshold cryptography, and more. See `MIGRATE_CRYPTO.md` for the complete crypto library documentation.

### Build Configuration

```bash
# Build WASM module with TinyGo
tinygo build -o worker.wasm -target wasi \
    -no-debug \
    -opt 2 \
    -scheduler none \
    ./worker/main.go

# Optimize with wasm-opt
wasm-opt -O3 -o worker.optimized.wasm worker.wasm

# Build with Extism toolchain
extism compile worker.wasm -o worker.plugin.wasm
```

### IPFS Integration

```go
// IPFS Configuration
const (
    IPFSStorageEndpoint   = "http://127.0.0.1:5001/api/v0/add"
    IPFSRetrievalEndpoint = "http://127.0.0.1:5001/api/v0/cat"
    IPFSGateway          = "https://ipfs.did.run/ipfs/"
)

// Export vault to IPFS
func ExportToIPFS(enclave *EnclaveData, password []byte) (cid string, error) {
    // 1. Serialize enclave data
    // 2. Encrypt with AES-256-GCM using password
    // 3. Upload to IPFS
    // 4. Return content ID (CID)
}
```

## Component 2: Server (Payment Gateway)

**Technology**: Go 1.24.4 HTTP Server
**Framework**: `go-wasm-http-server` (can run as service worker or HTTP)
**Purpose**: Payment processing and OIDC authorization

### Features

#### W3C Payment Handler API
- Payment request processing
- Card validation (Luhn algorithm, CVV, expiry)
- PCI DSS compliant tokenization
- Transaction signing with HMAC-SHA256
- AES-256-GCM encryption for card data
- Refund processing
- Audit logging

#### OIDC Authorization Server
- Full OpenID Connect provider
- Discovery endpoint
- Authorization with PKCE
- Token endpoint (JWT generation)
- UserInfo endpoint
- JWKS endpoint
- Refresh tokens

#### Security
- Rate limiting (100 req/min per client)
- Origin validation
- Security headers (CSP, X-Frame-Options)
- CORS configuration
- Secure token generation
- Card number masking

### API Endpoints

```
POST /api/payment/process        - Process payment
POST /api/payment/validate       - Validate payment method
POST /api/payment/refund         - Process refund

GET  /.well-known/openid-configuration
GET  /oauth2/authorize
POST /oauth2/token
GET  /oauth2/userinfo
GET  /oauth2/jwks
```

## Component 3: TypeScript SDK (`packages/`)

**Purpose**: Browser and Node.js integration for Motor services

### Package Structure

#### `@motr/es` (Core SDK)

**Client Module** (`client/`):
- `VaultClient`: Main vault operations client
- `VaultClientWithIPFS`: IPFS-enabled vault client
- RPC/REST API integration
- Transaction broadcasting

**Worker Module** (`worker/`):
- `MotorServiceWorkerManager`: Service worker lifecycle
- `MotorClient`: HTTP client for Motor API
- Payment gateway client
- OIDC client integration

**Plugin Module** (`plugin/`):
- Plugin loading and caching
- WASM module verification
- Enclave storage (IndexedDB, localStorage)
- IPFS pinning integration

**Codec Module** (`codec/`):
- Address encoding (Bech32, EIP-55)
- Key management
- Transaction signing
- Signature verification
- Message serialization

**Auth Module** (`auth/`):
- WebAuthn registration
- WebAuthn authentication
- Credential management
- Passkey integration

**Generated Protobufs** (`protobufs/`):
- Cosmos SDK types
- Sonr blockchain types
- IBC types
- CosmWasm types

#### `@motr/sdk` (High-Level SDK)
- Simplified API wrappers
- Common operation helpers
- Error handling utilities
- TypeScript type definitions

#### `@motr/ui` (UI Components)
- React components for vault operations
- WebAuthn UI flows
- Payment forms
- Dashboard widgets

#### `@motr/com` (Common Utilities)
- Validation helpers
- Formatting utilities
- Type definitions
- Constants

### Usage Examples

```typescript
// Initialize vault client
import { createVaultClient } from '@motr/es';

const client = await createVaultClient({
  rpcUrl: 'http://localhost:26657',
  restUrl: 'http://localhost:1317',
});

// Generate vault
const vault = await client.generate({ id: 'my-vault' });

// Sign message
const signature = await client.sign({
  message: new Uint8Array([1, 2, 3]),
  enclave: vault.data,
});

// Export to IPFS
const cid = await client.export({
  enclave: vault.data,
  password: new Uint8Array([/* password */]),
});
```

```typescript
// Service worker integration
import { registerMotorServiceWorker } from '@motr/es';

const registration = await registerMotorServiceWorker({
  workerUrl: '/worker.js',
  scope: '/motor',
});

// Use in browser
const plugin = await createMotorPlugin({
  auto_register_worker: true,
  prefer_service_worker: true,
});

await plugin.processPayment({
  amount: 100.00,
  currency: 'USD',
  method: 'card',
});
```

## Component 4: Web Applications

### Authentication App (`web/auth/`)

**Framework**: Next.js 14+
**Purpose**: WebAuthn registration and OIDC flows

**Features**:
- Passkey registration UI
- Login flows
- Session management
- OIDC client implementation
- WebAuthn ceremony handling

**Tech Stack**:
- Next.js (App Router)
- React 18
- TailwindCSS
- Fumadocs (documentation)
- Sonr UI components

### Dashboard App (`web/dash/`)

**Framework**: Next.js 14+
**Purpose**: Vault management and blockchain interaction

**Features**:
- Vault creation and management
- Transaction signing UI
- DID document viewer
- DWN record browser
- Token management
- Network switcher

**Tech Stack**:
- Next.js (App Router)
- React 18
- TailwindCSS
- @motr/sdk for blockchain interaction
- Charts and visualizations

### Shared Configuration

Both apps use:
- `@motr/ui` for shared components
- `@motr/sdk` for blockchain operations
- Environment-based configuration
- SSR/SSG optimization
- Edge runtime compatibility

## Build & Development

### Worker (WASM)
```bash
# Build worker WASM
make worker

# Or with TinyGo directly
tinygo build -o worker.wasm -target wasi ./worker/main.go
```

### Server (Payment Gateway)
```bash
# Build server
go build -o motr-server ./server/main.go

# Run server
./motr-server --port 8080
```

### TypeScript SDK
```bash
# Install dependencies
pnpm install

# Build all packages
pnpm -r build

# Build specific package
pnpm --filter @motr/es build

# Run tests
pnpm test

# Generate from protobufs
cd packages/es
pnpm gen:protobufs
```

### Web Applications
```bash
# Development
pnpm --filter @motr/auth dev
pnpm --filter @motr/dash dev

# Build
pnpm --filter @motr/auth build
pnpm --filter @motr/dash build

# Production
pnpm --filter @motr/auth start
```

## Testing Strategy

### Go (Worker)
```bash
# Unit tests
go test ./worker/...
go test ./server/...

# With race detection
go test -race ./...

# Coverage
go test -cover ./...
```

### TypeScript (SDK/Apps)
```bash
# Unit tests
pnpm test

# E2E tests
pnpm test:e2e

# Type checking
pnpm typecheck
```

### Integration Tests
```bash
# Requires Motor server running
INTEGRATION=true pnpm test
```

## Configuration

### Worker Environment Variables
```bash
# WASM runtime configuration (via Extism)
CHAIN_ID=sonr-testnet-1
PASSWORD=default-password
IPFS_GATEWAY=https://ipfs.did.run/ipfs/
```

### SDK Configuration
```typescript
interface MotorConfig {
  // RPC endpoints
  rpcUrl: string;
  restUrl: string;

  // IPFS
  ipfsGateways: string[];
  enableIPFSPersistence: boolean;

  // Service worker
  workerUrl: string;
  preferServiceWorker: boolean;

  // Security
  timeout: number;
  maxRetries: number;
}
```

### Web App Environment Variables
```bash
# Common
NODE_ENV=production
NEXT_PUBLIC_CHAIN_ID=sonr-testnet-1

# Auth App
NEXT_PUBLIC_AUTH_URL=http://localhost:3100
NEXT_PUBLIC_WEBAUTHN_RP_ID=localhost
NEXT_PUBLIC_WEBAUTHN_RP_NAME="Sonr Auth"

# Dashboard App
NEXT_PUBLIC_RPC_ENDPOINT=http://localhost:26657
NEXT_PUBLIC_REST_ENDPOINT=http://localhost:1317
NEXT_PUBLIC_IPFS_GATEWAY=https://ipfs.io/ipfs/
```

## Migration Checklist

When setting up the new `sonr-io/motr` repository:

### Worker
- [ ] Copy `cmd/vault/` → `worker/`
- [ ] Update import paths to use `github.com/sonr-io/crypto`
- [ ] Create `worker/go.mod` with crypto dependency
- [ ] Setup TinyGo build scripts
- [ ] Add WASM optimization pipeline
- [ ] Document export functions and types
- [ ] Create test suite for WASM functions

### Server
- [ ] Copy `cmd/motr/` → `server/`
- [ ] Update payment gateway handlers
- [ ] Configure OIDC provider
- [ ] Setup rate limiting
- [ ] Add monitoring/metrics
- [ ] Create Dockerfile
- [ ] Document API endpoints

### TypeScript SDK
- [ ] Copy `packages/` directory
- [ ] Update package names to `@motr/*`
- [ ] Setup pnpm workspace
- [ ] Configure build pipeline (tsup/rollup)
- [ ] Setup Biome for linting/formatting
- [ ] Generate types from protobuf
- [ ] Create comprehensive tests
- [ ] Setup Changesets for versioning
- [ ] Publish to npm registry

### Web Applications
- [ ] Copy `web/` directory
- [ ] Update dependencies to use `@motr/*`
- [ ] Configure environment variables
- [ ] Setup build and deployment
- [ ] Create Docker images
- [ ] Add E2E tests
- [ ] Document user flows

### General
- [ ] Create monorepo structure
- [ ] Setup CI/CD pipelines
- [ ] Configure release automation
- [ ] Create API documentation
- [ ] Write migration guide
- [ ] Setup npm organization (@motr)
- [ ] Configure security scanning
- [ ] Setup monitoring and logging

## Related Documentation

- TinyGo: https://tinygo.org/
- Extism: https://extism.org/
- W3C Payment Handler API: https://www.w3.org/TR/payment-handler/
- WebAuthn: https://www.w3.org/TR/webauthn/
- IPFS: https://docs.ipfs.tech/
- Next.js: https://nextjs.org/
- pnpm Workspaces: https://pnpm.io/workspaces
