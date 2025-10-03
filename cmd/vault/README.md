# Vault - WebAssembly Vault Plugin

Vault is a WebAssembly-based vault system for the Sonr blockchain that provides secure, isolated execution of cryptographic operations. Built using the Extism framework, Vault enables secure multi-party computation (MPC) and vault management within a sandboxed WebAssembly environment.

## Overview

Vault serves as a cryptographic vault system that:

- Provides secure enclave-based key generation and management
- Supports multi-chain transaction signing (Cosmos, EVM)
- Implements WebAuthn-based authentication
- Offers secure import/export functionality via IPFS
- Enables isolated execution through WebAssembly

## Architecture

### Core Components

- **MPC Enclave**: Multi-party computation system for secure key operations
- **Vault Management**: Create, unlock, and manage cryptographic vaults
- **IPFS Integration**: Secure backup and restore of encrypted vault data
- **WebAuthn Support**: Passwordless authentication for vault operations
- **Multi-Chain Support**: Transaction signing for different blockchain networks

### Build Configuration

Vault is built specifically for WebAssembly:

```go
//go:build js && wasm
// +build js,wasm
```

## API Reference

### Core Enclave Operations

#### `generate`

```go
//go:wasmexport generate
func generate() int32
```

Creates a new MPC enclave and returns the enclave data and public key.

**Input**: `GenerateRequest`

```json
{
  "id": "string"
}
```

**Output**: `GenerateResponse`

```json
{
  "data": "EnclaveData",
  "public_key": "[]byte"
}
```

#### `refresh`

```go
//go:wasmexport refresh
func refresh() int32
```

Refreshes an existing enclave with new cryptographic material.

**Input**: `RefreshRequest`

```json
{
  "enclave": "EnclaveData"
}
```

**Output**: `RefreshResponse`

```json
{
  "okay": "bool",
  "data": "EnclaveData"
}
```

#### `sign`

```go
//go:wasmexport sign
func sign() int32
```

Signs a message using the enclave's private key.

**Input**: `SignRequest`

```json
{
  "message": "[]byte",
  "enclave": "EnclaveData"
}
```

**Output**: `SignResponse`

```json
{
  "signature": "[]byte"
}
```

#### `verify`

```go
//go:wasmexport verify
func verify() int32
```

Verifies a signature against a message and public key.

**Input**: `VerifyRequest`

```json
{
  "public_key": "[]byte",
  "message": "[]byte",
  "signature": "[]byte"
}
```

**Output**: `VerifyResponse`

```json
{
  "valid": "bool"
}
```

### Vault Import/Export Operations

#### `export`

```go
//go:wasmexport export
func export() int32
```

Encrypts and exports vault data to IPFS, returning a Content ID (CID).

**Input**: `ExportRequest`

```json
{
  "enclave": "EnclaveData",
  "password": "[]byte"
}
```

**Output**: `ExportResponse`

```json
{
  "cid": "string",
  "success": "bool"
}
```

#### `import`

```go
//go:wasmexport import
func importVault() int32
```

Retrieves and decrypts vault data from IPFS using a CID and password.

**Input**: `ImportRequest`

```json
{
  "cid": "string",
  "password": "[]byte"
}
```

**Output**: `ImportResponse`

```json
{
  "enclave": "EnclaveData",
  "success": "bool"
}
```

### Advanced Vault Operations

#### `create_vault_enclave`

```go
//go:wasmexport create_vault_enclave
func createVaultEnclave() int32
```

Creates a new vault enclave with advanced configuration options.

**Input**: `EnclaveConfig`

```json
{
  "vault_id": "string",
  "key_derivation_method": "string",
  "encryption_algorithm": "string",
  "signing_algorithm": "string",
  "webauthn_enabled": "bool",
  "auto_lock_timeout": "int64",
  "key_rotation_interval": "int64",
  "supported_chains": ["string"],
  "max_concurrent_ops": "int",
  "memory_limit": "uint64"
}
```

#### `unlock_vault_enclave`

```go
//go:wasmexport unlock_vault_enclave
func unlockVaultEnclave() int32
```

Unlocks a vault enclave, optionally using WebAuthn authentication.

#### `lock_vault_enclave`

```go
//go:wasmexport lock_vault_enclave
func lockVaultEnclave() int32
```

Locks a vault enclave to prevent unauthorized access.

#### `rotate_vault_key`

```go
//go:wasmexport rotate_vault_key
func rotateVaultKey() int32
```

Rotates the cryptographic keys within a vault enclave.

### Multi-Chain Transaction Signing

#### `sign_cosmos_transaction`

```go
//go:wasmexport sign_cosmos_transaction
func signCosmosTransaction() int32
```

Signs transactions for Cosmos SDK-based blockchains.

#### `sign_evm_transaction`

```go
//go:wasmexport sign_evm_transaction
func signEvmTransaction() int32
```

Signs transactions for Ethereum Virtual Machine compatible chains.

#### `sign_message`

```go
//go:wasmexport sign_message
func signMessage() int32
```

Signs arbitrary messages using the vault's private key.

### Health and Monitoring

#### `get_vault_health`

```go
//go:wasmexport get_vault_health
func getVaultHealth() int32
```

Returns the health status of a vault enclave.

**Output**: `EnclaveHealth`

```json
{
  "vault_id": "string",
  "status": "string",
  "last_activity": "int64",
  "key_rotation_due": "bool",
  "attestation_valid": "bool"
}
```

## Configuration

### Environment Variables

Motor supports configuration through Extism variables:

- `chain_id`: Blockchain network identifier (default: "sonr-testnet-1")
- `password`: Default password for enclave operations (default: "password")
- `gateway`: IPFS gateway URL (default: "https://ipfs.did.run/ipfs/")

Access these via helper functions:

```go
func GetChainID() string
func GetPassword() []byte
func GetGateway() string
```

### IPFS Integration

Motor integrates with IPFS for secure vault backup and restore:

- **Storage Endpoint**: `http://127.0.0.1:5001/api/v0/add`
- **Retrieval Endpoint**: `http://127.0.0.1:5001/api/v0/cat`
- **Data Format**: Encrypted vault data stored as content-addressed objects
- **Security**: All vault data is encrypted before IPFS storage

## Security Features

### Enclave Isolation

- WebAssembly sandbox provides memory isolation
- Secure execution environment prevents side-channel attacks
- Attestation mechanisms ensure enclave integrity

### Authentication

- WebAuthn support for passwordless authentication
- Challenge-response authentication flows
- Automatic vault locking with configurable timeouts

### Key Management

- Multi-party computation for enhanced security
- Automatic key rotation with configurable intervals
- Secure key derivation and storage

### Data Protection

- AES encryption for sensitive data
- Password-based encryption for import/export
- Secure memory handling within WASM environment

## Usage Examples

### Basic Enclave Operations

```javascript
// Generate new enclave
const generateReq = { id: "my-vault" };
const result = call_wasm_function("generate", generateReq);

// Sign a message
const signReq = {
  message: new Uint8Array([1, 2, 3, 4]),
  enclave: result.data,
};
const signature = call_wasm_function("sign", signReq);
```

### Vault Management

```javascript
// Create vault with configuration
const config = {
  vault_id: "user-vault-001",
  webauthn_enabled: true,
  auto_lock_timeout: 300,
  supported_chains: ["cosmos", "ethereum"],
};
const vault = call_wasm_function("create_vault_enclave", config);

// Sign Cosmos transaction
const cosmosReq = {
  vault_id: "user-vault-001",
  chain_type: "cosmos",
  chain_id: "cosmoshub-4",
  message: transactionBytes,
};
const cosmosResult = call_wasm_function("sign_cosmos_transaction", cosmosReq);
```

### Import/Export Operations

```javascript
// Export vault to IPFS
const exportReq = {
  enclave: vaultData,
  password: new Uint8Array([
    /* password bytes */
  ]),
};
const exportResult = call_wasm_function("export", exportReq);
console.log("Vault exported to CID:", exportResult.cid);

// Import vault from IPFS
const importReq = {
  cid: "QmXxXxXxXxXxXxXxXxXxXxXxXxXxXxXxXxXxXxXxXx",
  password: new Uint8Array([
    /* password bytes */
  ]),
};
const importResult = call_wasm_function("import", importReq);
```

## Building and Deployment

### Prerequisites

- Go 1.24.4+
- Extism runtime
- IPFS node (for import/export functionality)

### Build Commands

```bash
# Build WebAssembly module
GOOS=js GOARCH=wasm go build -o motr.wasm main.go

# Build via Makefile
make motr
```

### Integration

Motor is designed to be integrated with:

- **Highway Service**: PostgreSQL-backed HTTP API
- **Sonr Blockchain**: Cosmos SDK-based blockchain node
- **IPFS Network**: Decentralized storage system
- **WebAuthn Infrastructure**: Passwordless authentication

## Error Handling

All functions return `int32` status codes:

- `0`: Success
- `1`: Error (details available via `pdk.SetError`)

Error information is logged using Extism's logging system:

```go
pdk.Log(pdk.LogError, "Error message")
pdk.Log(pdk.LogInfo, "Info message")
```

## Dependencies

### Core Dependencies

- `github.com/extism/go-pdk`: WebAssembly plugin development kit
- `github.com/sonr-io/sonr/crypto/mpc`: Multi-party computation library

### Cryptographic Libraries

- `filippo.io/edwards25519`: Edwards25519 elliptic curve
- `github.com/btcsuite/btcd/btcec/v2`: Bitcoin cryptography
- `github.com/consensys/gnark-crypto`: Zero-knowledge proof cryptography

## License

Motor is part of the Sonr blockchain project. See the project's main license for terms and conditions.
