# Highway (hway) Migration Context

> **Repository Migration**: `sonr-io/sonr` → `sonr-io/hway`
> **Components Moved**: `cmd/hway/`, `bridge/`, `internal/migrations/`

## Overview

Highway is a high-performance, PostgreSQL-backed HTTP service that handles OAuth2/OIDC authentication, WebAuthn flows, and asynchronous vault operations for the Sonr blockchain ecosystem. It serves as the authentication and task processing layer between clients and the Sonr blockchain.

## Architecture

### Core Technology Stack

- **Go**: 1.24.4
- **Task Queue**: Asynq (Redis-backed distributed task queue)
- **Actor System**: Proto.Actor for concurrency management
- **Database**: PostgreSQL with database/sql
- **Web Framework**: Echo v4
- **Authentication**: OAuth2, OIDC, WebAuthn, SIOP
- **Cryptography**: UCAN tokens, JWT signing (RS256)

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Highway Service (hway)                    │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │   Bridge   │  │   Tasks    │  │  Handlers  │            │
│  │  (HTTP)    │──│  (Asynq)   │──│  (Auth)    │            │
│  └────────────┘  └────────────┘  └────────────┘            │
│        │                │                │                   │
│        ▼                ▼                ▼                   │
│  ┌─────────────────────────────────────────┐               │
│  │        Proto.Actor System                │               │
│  │    (Vault Actor Management)              │               │
│  └─────────────────────────────────────────┘               │
│                                                               │
└───────────┬───────────────────────┬─────────────────────────┘
            │                       │
            ▼                       ▼
      ┌──────────┐           ┌──────────┐
      │PostgreSQL│           │  Redis   │
      │ (State)  │           │ (Queue)  │
      └──────────┘           └──────────┘
```

## Component Breakdown

### 1. Bridge Module (`bridge/`)

**Purpose**: HTTP API layer providing authentication and authorization services

**Key Files**:
- `bridge.go` - Main bridge server initialization
- `config.go` - Configuration management
- `queue.go` - Asynq task queue setup and management

**Handlers** (`bridge/handlers/`):

#### Authentication & Authorization
- `auth.go` - General authentication handlers
- `oidc.go` - OpenID Connect provider implementation
  - Discovery endpoint (`.well-known/openid-configuration`)
  - Authorization endpoint with PKCE support
  - Token endpoint with JWT generation
  - UserInfo endpoint
  - JWKS endpoint for key rotation

- `siop.go` - Self-Issued OpenID Provider (SIOP) flows
  - DID-based authentication
  - Verifiable presentation handling

- `webauthn.go` - WebAuthn registration and authentication
  - Challenge generation
  - Credential verification
  - Device binding

#### OAuth2 Implementation
- `oauth2_provider.go` - Core OAuth2 provider
  - Authorization code flow
  - Client credentials flow
  - Refresh token flow
  - Token introspection
  - Token revocation

- `oauth2_register.go` - Dynamic client registration (RFC 7591)
- `oauth2_clients.go` - Client management and validation
- `oauth2_delegation.go` - Token delegation flows
- `oauth2_token_exchange.go` - Token exchange (RFC 8693)
- `oauth2_scopes.go` - Scope validation and management
- `oauth2_security.go` - Security utilities (PKCE, rate limiting)
- `oauth2_types.go` - OAuth2 type definitions

#### Vault Operations
- `vault.go` - Vault operation handlers
  - Generate vault enclaves
  - Sign with vault
  - Verify signatures
  - Import/Export to IPFS
  - Refresh vault state

#### Utility Handlers
- `broadcast.go` - Transaction broadcasting
- `health.go` - Health check endpoints
- `websocket.go` - WebSocket connection management
- `types.go` - Shared type definitions

### 2. Task Processing (`bridge/tasks/`)

**Purpose**: Asynchronous task definitions and processing

**Key Files**:
- `types.go` - Task type constants and definitions
- `generate.go` - Vault generation tasks
- `signing.go` - Signing operation tasks
- `attenuation.go` - UCAN token attenuation tasks

**Task Types**:
```go
const (
    TypeVaultGenerate  = "vault:generate"
    TypeVaultSign      = "vault:sign"
    TypeVaultRefresh   = "vault:refresh"
    TypeUCANAttenuation = "ucan:attenuation"
)
```

**Queue Configuration**:
```go
Queues: map[string]int{
    "critical": 6,  // High priority tasks
    "default":  3,  // Normal priority tasks
    "low":      1,  // Low priority tasks
}
```

### 3. Main Service (`cmd/hway/`)

**Purpose**: Service entry point and initialization

**Key Responsibilities**:
1. Initialize Asynq server with Redis connection
2. Configure worker pools and queue priorities
3. Register task handlers
4. Start HTTP server (Echo)
5. Setup signal handling for graceful shutdown

**Configuration**:
```go
const (
    RedisAddr        = "127.0.0.1:6379"
    PostgresAddr     = "127.0.0.1:5432"
    HTTPPort         = ":8090"
    WorkerConcurrency = 10
)
```

### 4. Database Migrations (`internal/migrations/`)

**Purpose**: PostgreSQL schema management

**Migration Files**:
- `001_accounts_table.sql` - User account storage
- `002_credentials_table.sql` - WebAuthn credential storage
- `003_profiles_table.sql` - User profile data
- `004_vaults_table.sql` - Vault state persistence
- `005_create_cosmos_registry.sql` - Cosmos chain registry
- `006_execute_cosmos_registry.sql` - Registry functions
- `007_webauthn_to_vc_func.sql` - WebAuthn to VC conversion
- `008_create_coinpaprika_market_data.sql` - Market data tables
- `009_webauthn_options_functions.sql` - WebAuthn helper functions
- `010_crypto_asset_symbol_linking.sql` - Asset metadata
- `011_common_functions.sql` - Shared SQL functions
- `012_crypto_coin_price_data.sql` - Price data storage
- `013_add_asset_quality_filters.sql` - Asset filtering
- `014_sessions_table.sql` - Session management

## Integration Points

### With Sonr Blockchain (`snrd`)
- **RPC/REST API**: Queries blockchain state via Cosmos SDK endpoints
- **Transaction Broadcasting**: Submits signed transactions to chain
- **DID Resolution**: Resolves DIDs from blockchain state
- **Vault State**: Stores vault metadata on-chain

### With Motor/Worker (WASM Plugin)
- **Task Execution**: Highway enqueues tasks, Motor executes via WASM
- **Vault Operations**: Motor provides cryptographic operations
- **Enclave Management**: Actor system manages WASM plugin lifecycle

### With Client Applications
- **OAuth2/OIDC**: Standard OAuth2 authorization flows
- **WebAuthn**: Browser-based passwordless authentication
- **WebSocket**: Real-time task status updates
- **SSE**: Server-Sent Events for progress tracking

### With External Services
- **IPFS**: Vault backup/restore operations
- **Redis**: Distributed task queue and caching
- **PostgreSQL**: Persistent state storage

## Key Features

### 1. OAuth2/OIDC Provider
- Full OAuth2 authorization server implementation
- OpenID Connect provider with ID tokens
- Dynamic client registration (RFC 7591)
- Token exchange (RFC 8693)
- PKCE support for public clients
- Refresh token rotation
- Token revocation and introspection

### 2. WebAuthn Support
- FIDO2/WebAuthn registration flows
- Authentication with platform authenticators
- Credential lifecycle management
- Challenge-response validation
- Attestation verification

### 3. Vault Task Processing
- Asynchronous cryptographic operations
- Priority-based queue management
- Actor-based concurrency model
- Retry logic with exponential backoff
- Task status tracking and notifications

### 4. UCAN Token Management
- UCAN token generation and signing
- Capability delegation and attenuation
- Token chain verification
- Integration with DID system

## Security Considerations

### Authentication & Authorization
- Multi-factor authentication support
- JWT token signing with RS256
- PKCE for authorization code flow
- Origin validation for WebAuthn
- Rate limiting on all endpoints

### Data Protection
- Password hashing with Argon2
- Encrypted vault data in PostgreSQL
- Secure token generation (crypto/rand)
- HTTPS-only in production
- CORS configuration

### Vault Security
- WASM sandbox isolation for cryptographic operations
- No private key exposure to server
- Encrypted backup to IPFS
- Session timeout and auto-lock

## Configuration

### Environment Variables
```bash
# Service Configuration
HIGHWAY_PORT=8090
LOG_LEVEL=info

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=hway

# Redis
REDIS_URL=redis://localhost:6379

# IPFS
IPFS_API_URL=http://localhost:5001

# OAuth2/OIDC
OIDC_ISSUER=http://localhost:8090
JWT_SIGNING_KEY_PATH=/path/to/private-key.pem
JWT_PUBLIC_KEY_PATH=/path/to/public-key.pem

# Security
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3100
SESSION_SECRET=change-me-in-production
```

### Asynq Configuration
```go
asynq.Config{
    Concurrency: 10,
    Queues: map[string]int{
        "critical": 6,
        "default":  3,
        "low":      1,
    },
    StrictPriority: false,
    ErrorHandler: asynq.ErrorHandlerFunc(handleError),
    Logger: slog.Default(),
}
```

## API Endpoints

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Refresh access token

### OAuth2/OIDC
- `GET /.well-known/openid-configuration` - OIDC discovery
- `GET /oauth2/authorize` - Authorization endpoint
- `POST /oauth2/token` - Token endpoint
- `GET /oauth2/userinfo` - User info endpoint
- `GET /oauth2/jwks` - JSON Web Key Set
- `POST /oauth2/register` - Dynamic client registration
- `POST /oauth2/revoke` - Token revocation
- `POST /oauth2/introspect` - Token introspection

### WebAuthn
- `POST /webauthn/register/begin` - Start registration
- `POST /webauthn/register/finish` - Complete registration
- `POST /webauthn/login/begin` - Start authentication
- `POST /webauthn/login/finish` - Complete authentication

### Vault Operations
- `POST /vault/generate` - Generate new vault
- `POST /vault/sign` - Sign with vault
- `POST /vault/verify` - Verify signature
- `POST /vault/refresh` - Refresh vault state
- `POST /vault/export` - Export to IPFS
- `POST /vault/import` - Import from IPFS

### WebSocket
- `WS /ws/tasks/{task_id}` - Task status updates

### Health & Monitoring
- `GET /health` - Health check
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

## Testing Strategy

### Unit Tests
```bash
go test ./bridge/...
go test ./bridge/handlers/...
go test ./bridge/tasks/...
```

### Integration Tests
```bash
# Requires PostgreSQL and Redis
INTEGRATION=true go test ./...
```

### E2E Tests
```bash
# Requires full stack (PostgreSQL, Redis, IPFS)
E2E=true go test ./e2e/...
```

## Dependencies

### Required Services
- PostgreSQL 14+
- Redis 7+
- IPFS node (for vault operations)

### Go Modules (Key Dependencies)
- `github.com/hibiken/asynq` - Distributed task queue
- `github.com/labstack/echo/v4` - HTTP framework
- `github.com/asynkron/protoactor-go` - Actor system
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/go-webauthn/webauthn` - WebAuthn library
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `github.com/redis/go-redis/v9` - Redis client

## Build & Deployment

### Build Commands
```bash
# Build binary
go build -o hway ./cmd/hway

# Build with specific tags
go build -tags production -o hway ./cmd/hway

# Build Docker image
docker build -t sonr-hway:latest .
```

### Docker Deployment
```yaml
services:
  hway:
    image: onsonr/hway:latest
    environment:
      POSTGRES_HOST: postgres
      REDIS_URL: redis://redis:6379
    depends_on:
      - postgres
      - redis
    ports:
      - "8090:8090"
```

## Migration Checklist

When setting up the new `sonr-io/hway` repository:

- [ ] Copy `cmd/hway/` directory
- [ ] Copy `bridge/` directory (all handlers and tasks)
- [ ] Copy `internal/migrations/` for database schema
- [ ] Update import paths from `github.com/sonr-io/sonr` to new repo
- [ ] Create standalone `go.mod` with required dependencies
- [ ] Setup CI/CD for independent releases
- [ ] Create Dockerfile for containerized deployment
- [ ] Document environment variables and configuration
- [ ] Add PostgreSQL and Redis setup instructions
- [ ] Create integration test suite with testcontainers
- [ ] Setup database migration tooling (e.g., golang-migrate)
- [ ] Configure monitoring and observability (Prometheus/Grafana)
- [ ] Document OAuth2 client registration process
- [ ] Create API documentation (OpenAPI/Swagger)

## Related Documentation

- OAuth2 RFC 6749: https://tools.ietf.org/html/rfc6749
- OpenID Connect Core: https://openid.net/specs/openid-connect-core-1_0.html
- Dynamic Client Registration RFC 7591: https://tools.ietf.org/html/rfc7591
- WebAuthn Spec: https://www.w3.org/TR/webauthn/
- Asynq Documentation: https://github.com/hibiken/asynq
- Proto.Actor: https://proto.actor/
