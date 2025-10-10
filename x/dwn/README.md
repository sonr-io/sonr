# `x/dwn`

The `x/dwn` module is the foundational engine of the Sonr ecosystem. It provides a comprehensive **Decentralized Web Node (DWN)** implementation that serves as the backbone for user-controlled data storage, protocol management, and secure vault operations. The module enables users to maintain sovereign control over their data while participating in a decentralized ecosystem.

## Overview

The DWN module implements the [Decentralized Web Node specification](https://identity.foundation/decentralized-web-node/spec/), providing:

- **Personal Data Stores**: User-controlled storage for structured data records
- **Protocol-Based Interactions**: Define and enforce data schemas and interaction patterns
- **Granular Permissions**: Fine-grained access control using capability-based authorization
- **Secure Vaults**: Enclave-based key management and transaction signing via WebAssembly
- **Multi-Chain Transaction Building**: Support for both Cosmos SDK and EVM transaction construction
- **Enhanced Address Derivation**: BIP44 HD wallet address derivation for multiple blockchain networks

## Module Structure

The DWN module is organized as follows:

```
x/dwn/
├── client/           # Client implementations
│   └── wasm/        # WebAssembly motor client
│       └── main.go  # Motor enclave implementation
├── keeper/          # Business logic and state management
│   ├── keeper.go    # Main keeper implementation
│   ├── vault_keeper.go      # VaultKeeper implementation
│   ├── wallet_derivation.go # Multi-chain address derivation
│   ├── dwn_records.go       # DWN record management
│   ├── dwn_protocols.go     # Protocol management
│   ├── dwn_permissions.go   # Permission management
│   ├── msg_server.go        # Message server implementation
│   └── query_server.go      # Query server implementation
├── types/           # Protobuf-generated types and interfaces
│   ├── vault_types.go       # Vault-specific types and requests
│   ├── vault_spawn.go       # Vault spawning functionality
│   ├── vault_plugin.go      # WebAssembly plugin integration
│   └── ipfs/        # IPFS integration types
├── vaults/          # Vault storage (motr.wasm output)
└── Makefile         # Module-specific build targets
```

## Core Concepts

### Decentralized Web Nodes (DWNs)

A DWN is a personal data store that enables individuals to manage their data independently of centralized providers. Each user's DWN serves as their agent in the decentralized web, storing data, managing permissions, and executing protocols on their behalf.

### Records

Records are the fundamental unit of data storage in a DWN. Each record can:

- Store arbitrary data with optional encryption
- Be organized hierarchically using parent-child relationships
- Conform to specific protocols and schemas
- Be published for public access or kept private

### Protocols

Protocols define structured ways for applications to interact with DWN data. They specify:

- Data schemas for validation
- Permission models
- Interaction patterns between different parties

### Permissions

The DWN uses a capability-based permission system where:

- Permissions are granted as signed tokens (JWTs)
- Access can be scoped to specific interfaces, methods, protocols, or records
- Permissions can be delegated and revoked

### Vaults

Vaults provide secure, enclave-based key management enabling:

- Hardware-backed key generation and storage
- Secure transaction signing without exposing private keys
- Multi-party computation capabilities
- WebAssembly-based secure execution environment
- Multi-chain transaction building for Cosmos SDK and EVM networks
- BIP44 HD wallet address derivation with configurable coin types

#### VaultKeeper Implementation

The VaultKeeper provides a comprehensive interface for managing cryptographic vaults within the DWN module. It implements the following core functionality:

- **Vault Creation**: Creates new vaults with enclave-based key generation using WebAssembly plugins
- **State Management**: Manages vault states including ownership, public keys, and enclave data
- **Secure Operations**: Provides signing and transaction broadcasting capabilities through secure enclaves
- **Refresh Mechanisms**: Handles vault state refresh with configurable intervals for security
- **Verification**: Cryptographic signature verification for vault operations
- **Multi-Chain Support**: Transaction building for both Cosmos SDK and EVM networks using pkg/txns
- **Address Derivation**: BIP44 HD wallet address derivation for multiple blockchain networks

The VaultKeeper integrates with the Motor plugin system (`motr.wasm`) to provide isolated, secure execution environments for cryptographic operations.

##### VaultKeeper Interface Methods

The VaultKeeper interface provides the following methods for vault management:

**Core Operations:**

- `CreateVault(ctx, msg)`: Creates a new vault with enclave-based key generation
- `RefreshVault(ctx, msg)`: Refreshes the enclave state with configurable intervals
- `SignWithVault(ctx, msg)`: Signs messages using the vault's secure enclave
- `BroadcastTx(ctx, msg)`: Broadcasts transactions through the vault's enclave
- `VerifySignature(ctx, publicKey, message, signature)`: Verifies cryptographic signatures

**Query Operations:**

- `GetVaultState(ctx, vaultID)`: Retrieves vault state by ID
- `ListVaultsByOwner(ctx, owner)`: Lists all vaults owned by a specific address

**Actor System Integration:**

- `SpawnVault(opts...)`: Creates vault actors with configuration options
- `SpawnSimpleVault()`: Creates vaults without database persistence (testing)
- `SpawnSimpleVaultNamed(name)`: Creates named vaults for testing scenarios

All methods return appropriate response types and handle error conditions including ownership verification, parameter validation, and enclave communication failures.

**Transaction Building and Address Derivation:**

The VaultKeeper integrates with the `pkg/txns` package to provide enhanced multi-chain transaction capabilities:

- `BuildCosmosTransaction(params)`: Builds unsigned Cosmos SDK transactions with proper fee estimation
- `BuildEVMTransaction(params)`: Builds unsigned Ethereum/EVM transactions with gas estimation
- `CreateVaultSigner(vaultID)`: Creates MPC-based signers for secure transaction signing
- `EstimateTransactionFee(txType, params)`: Estimates transaction fees for both Cosmos and EVM networks
- `DeriveWalletAddresses(did, salt)`: Derives both Cosmos and EVM addresses using BIP44 HD wallet derivation

**Address Derivation Features:**

- **Multi-Chain Support**: Generates addresses for both Cosmos SDK (Bech32) and EVM (0x) formats
- **Deterministic Derivation**: Uses DID and salt for reproducible address generation
- **Configurable Coin Types**: Supports different coin types for various blockchain networks
- **BIP44 Compliance**: Follows BIP44 hierarchical deterministic wallet standard
- **Secure Generation**: Address derivation uses cryptographically secure methods

## State

The module maintains the following state:

### DWN Records

```protobuf
message DWNRecord {
  string record_id = 1;        // Unique identifier
  string target = 2;           // Target DWN (DID)
  DWNMessageDescriptor descriptor = 3;
  string authorization = 4;    // JWT/signature
  bytes data = 5;             // Record data
  string protocol = 6;        // Optional protocol URI
  string protocol_path = 7;   // Optional protocol path
  string schema = 8;          // Optional schema URI
  string parent_id = 9;       // Optional parent record
  bool published = 10;        // Public visibility flag
}
```

### DWN Protocols

```protobuf
message DWNProtocol {
  string protocol_uri = 1;     // Unique protocol identifier
  string target = 2;           // Target DWN (DID)
  DWNMessageDescriptor descriptor = 3;
  string authorization = 4;    // JWT/signature
  bytes definition = 5;        // Protocol definition (JSON)
  bool published = 6;          // Public visibility flag
}
```

### DWN Permissions

```protobuf
message DWNPermission {
  string permission_id = 1;    // Unique identifier
  string grantor = 2;          // Permission grantor (DID)
  string grantee = 3;          // Permission recipient (DID)
  string target = 4;           // Target DWN (DID)
  DWNMessageDescriptor descriptor = 5;
  string authorization = 6;    // JWT/signature
  // Permission scope fields...
  bool revoked = 7;           // Revocation status
}
```

### Vault States

```protobuf
message VaultState {
  string vault_id = 1;         // Unique vault identifier
  string owner = 2;            // Vault owner address
  string public_key = 3;       // Vault public key
  string enclave_report = 4;   // Attestation report
  uint64 created_at = 5;       // Creation timestamp
  uint64 last_heartbeat = 6;   // Last activity timestamp
}
```

## Messages

### Records Management

#### MsgRecordsWrite

Creates or updates a record in the DWN.

```protobuf
message MsgRecordsWrite {
  string author = 1;           // Message author (DID or address)
  string target = 2;           // Target DWN (DID)
  DWNMessageDescriptor descriptor = 3;
  bytes data = 4;              // Record data
  string authorization = 5;    // Optional JWT/signature
}
```

#### MsgRecordsDelete

Deletes a record from the DWN.

```protobuf
message MsgRecordsDelete {
  string author = 1;
  string target = 2;
  string record_id = 3;
  DWNMessageDescriptor descriptor = 4;
  string authorization = 5;
  bool prune = 6;              // Delete all descendants
}
```

### Protocol Management

#### MsgProtocolsConfigure

Configures a protocol in the DWN.

```protobuf
message MsgProtocolsConfigure {
  string author = 1;
  string target = 2;
  string protocol_uri = 3;
  bytes definition = 4;        // Protocol definition (JSON)
  DWNMessageDescriptor descriptor = 5;
  string authorization = 6;
  bool published = 7;
}
```

### Permission Management

#### MsgPermissionsGrant

Grants permissions in the DWN.

```protobuf
message MsgPermissionsGrant {
  string grantor = 1;
  string target = 2;
  string grantee = 3;
  DWNMessageDescriptor descriptor = 4;
  string authorization = 5;
}
```

#### MsgPermissionsRevoke

Revokes permissions in the DWN.

```protobuf
message MsgPermissionsRevoke {
  string grantor = 1;
  string permission_id = 2;
  DWNMessageDescriptor descriptor = 3;
  string authorization = 4;
}
```

### Vault Operations

#### MsgCreateVault

Creates a new vault with enclave-based key generation.

```protobuf
message MsgCreateVault {
  string owner = 1;
  string vault_id = 2;
  string key_id = 3;
}
```

#### MsgRefreshVault

Refreshes the enclave state of a vault.

```protobuf
message MsgRefreshVault {
  string owner = 1;
  string vault_id = 2;
}
```

#### MsgSignWithVault

Signs a message using the vault's enclave.

```protobuf
message MsgSignWithVault {
  string owner = 1;
  string vault_id = 2;
  bytes message = 3;
}
```

#### MsgBroadcastTx

Broadcasts a transaction using the vault's enclave.

```protobuf
message MsgBroadcastTx {
  string owner = 1;
  string vault_id = 2;
  bytes tx_bytes = 3;
}
```

## Queries

### Records Queries

- `Records`: List records with filters (protocol, schema, parent, published status)
- `Record`: Get a specific record by ID

### Protocol Queries

- `Protocols`: List protocols with optional published filter
- `Protocol`: Get a specific protocol by URI

### Permission Queries

- `Permissions`: List permissions with filters (grantor, grantee, interface, method)

### Vault Queries

- `Vault`: Get a specific vault by ID
- `Vaults`: List vaults by owner

### Utility Queries

- `VerifySignature`: Verify a cryptographic signature
- `Params`: Get module parameters

### Wallet Derivation Queries

- `WalletDerivation`: Derive Cosmos and EVM addresses from DID and salt using BIP44 HD wallet derivation
- `WalletStatus`: Get wallet initialization status and balance information for a given address

## CLI Examples

### Records Operations

```bash
# Write a new record
snrd tx dwn records-write did:example:123 '{"interface_name":"Records","method":"Write"}' '{"name":"Alice","age":30}' \
  --protocol example.com/profile/v1 \
  --published \
  --from alice

# Query records
snrd query dwn records did:example:123 --protocol example.com/profile/v1

# Delete a record
snrd tx dwn records-delete did:example:123 record-123 '{"interface_name":"Records","method":"Delete"}' \
  --from alice
```

### Protocol Configuration

```bash
# Configure a new protocol
snrd tx dwn protocols-configure did:example:123 example.com/social/v1 \
  '{"types":{"post":{"schema":"https://example.com/schemas/post.json"}}}' \
  '{"interface_name":"Protocols","method":"Configure"}' \
  --published \
  --from alice

# Query protocols
snrd query dwn protocols did:example:123 --published-only
```

### Permission Management

```bash
# Grant permissions
snrd tx dwn permissions-grant did:example:123 did:example:456 \
  '{"interface_name":"Permissions","method":"Grant"}' \
  --from alice

# Query permissions
snrd query dwn permissions did:example:123 --grantee did:example:456
```

### Vault Operations

The VaultKeeper provides several operations for managing cryptographic vaults:

```bash
# Create a vault with enclave-based key generation
snrd tx dwn create-vault my-vault key-1 --from alice

# Refresh vault enclave state (requires minimum interval)
snrd tx dwn refresh-vault my-vault --from alice

# Sign a message with vault's secure enclave
snrd tx dwn sign-with-vault my-vault "48656c6c6f20576f726c64" --from alice

# Broadcast a transaction using vault's enclave
snrd tx dwn broadcast-tx my-vault "transaction-bytes" --from alice

# Query specific vault state
snrd query dwn vault my-vault

# Query all vaults owned by an address
snrd query dwn vaults sonr1... --owner-only

# Verify a signature against a public key
snrd query dwn verify-signature "public-key-hex" "message-hex" "signature-hex"
```

### Wallet Derivation Operations

The DWN module provides enhanced address derivation capabilities using BIP44 HD wallet standards:

```bash
# Derive multi-chain addresses from DID and salt
snrd query dwn wallet-derivation "did:example:alice" "my-salt-123"

# Check wallet status and balances for an address
snrd query dwn wallet-status "idx1abcd..."

# Example response for wallet derivation:
# {
#   "cosmos_address": "idx1abcdef...",
#   "evm_address": "0x1234abcd...",
#   "derivation_path": "m/44'/60'/0'/0/0",
#   "did": "did:example:alice",
#   "salt": "my-salt-123"
# }
```

#### VaultKeeper Features

- **Secure Key Generation**: Uses WebAssembly enclaves for tamper-resistant key generation
- **Ownership Verification**: Ensures only vault owners can perform operations
- **Refresh Intervals**: Configurable minimum intervals between vault refreshes for security
- **Signature Operations**: Secure message signing without exposing private keys
- **Transaction Broadcasting**: Direct transaction submission through vault enclaves
- **State Persistence**: Vault states are stored in the blockchain state with ORM integration
- **Multi-Chain Transaction Building**: Integration with pkg/txns for Cosmos SDK and EVM transaction construction
- **Enhanced Address Derivation**: BIP44 HD wallet derivation with support for multiple blockchain networks
- **Fee Estimation**: Automatic fee calculation for both Cosmos and EVM transaction types
- **MPC Integration**: Multi-party computation capabilities for secure distributed key management

## Integration Guide

### For Application Developers

1. **Define Your Protocol**: Create a protocol definition that describes your data structures and permissions
2. **Configure Protocol**: Register your protocol with target DWNs
3. **Request Permissions**: Use the SVC module to request necessary permissions from users
4. **Store Data**: Write records that conform to your protocol
5. **Query Data**: Read records based on granted permissions

### For Wallet Developers

1. **VaultKeeper Integration**:
   - Use `CreateVault` for secure key generation in WebAssembly enclaves
   - Implement `RefreshVault` calls based on configured refresh intervals
   - Use `SignWithVault` for transaction signing without exposing private keys
   - Leverage `BroadcastTx` for direct transaction submission through vaults
   - Utilize multi-chain transaction building for both Cosmos SDK and EVM networks
   - Implement address derivation for multi-chain wallet support

2. **Vault Management UI**:
   - Display vault states and ownership information
   - Show vault public keys and enclave data
   - Provide refresh status and last refresh timestamps
   - Enable vault-based message signing interfaces
   - Show derived addresses for multiple blockchain networks
   - Display transaction building capabilities and fee estimations

3. **Permission Dashboard**: Build UI for users to manage granted permissions
4. **Data Browser**: Create interfaces for users to view and manage their DWN records
5. **Protocol Registry**: Show installed protocols and their data

## Events

The DWN module emits comprehensive typed events for all state-changing operations. These events provide a detailed audit trail and enable efficient tracking of DWN-related activities.

### Event Types

#### 1. EventRecordWritten
- **Emitted**: When a record is written to DWN
- **Fields**:
  - `record_id`: Unique record identifier
  - `target`: Target DID
  - `protocol`: Protocol URI defining record structure
  - `schema`: Schema URI for record validation
  - `data_cid`: Content Identifier for stored data
  - `data_size`: Size of record data in bytes
  - `encrypted`: Whether data is encrypted
  - `block_height`: Block number of record creation

#### 2. EventRecordDeleted
- **Emitted**: When a record is deleted from DWN
- **Fields**:
  - `record_id`: Unique record identifier
  - `target`: Target DID
  - `deleter`: Address performing deletion
  - `block_height`: Block number of deletion

#### 3. EventProtocolConfigured
- **Emitted**: When a protocol is configured in a DWN
- **Fields**:
  - `target`: Target DID
  - `protocol_uri`: Unique protocol identifier
  - `published`: Public visibility flag
  - `block_height`: Block number of configuration

#### 4. EventPermissionGranted
- **Emitted**: When a permission is granted
- **Fields**:
  - `permission_id`: Unique permission identifier
  - `grantor`: DID granting permission
  - `grantee`: DID receiving permission
  - `interface_name`: Targeted interface
  - `method`: Specific method being permitted
  - `expires_at`: Expiration timestamp
  - `block_height`: Block number of permission grant

#### 5. EventPermissionRevoked
- **Emitted**: When a permission is revoked
- **Fields**:
  - `permission_id`: Unique permission identifier
  - `revoker`: DID revoking the permission
  - `block_height`: Block number of revocation

#### 6. EventVaultCreated
- **Emitted**: When a new vault is created
- **Fields**:
  - `vault_id`: Unique vault identifier
  - `owner`: Vault owner address
  - `public_key`: Vault's public key
  - `block_height`: Block number of vault creation

#### 7. EventVaultKeysRotated
- **Emitted**: When vault keys are rotated
- **Fields**:
  - `vault_id`: Unique vault identifier
  - `owner`: Vault owner address
  - `new_public_key`: New public key
  - `rotation_height`: Block number of key rotation
  - `block_height`: Block number of rotation event

### Event Indexing and Querying

Events can be queried and filtered using CometBFT WebSocket or standard blockchain explorers. Example queries:

```bash
# Query all record write events
tm.event='Tx' AND dwn.v1.EventRecordWritten.record_id EXISTS

# Query events by target DID
dwn.v1.EventRecordWritten.target='did:sonr:example'

# Query protocol configuration events
tm.event='Tx' AND dwn.v1.EventProtocolConfigured.published=true
```

### Best Practices for Event Consumers

1. **Indexing**: Configure comprehensive CometBFT event indexing
2. **Performance**: Use efficient, targeted event queries
3. **Replay Handling**: Implement robust event replay mechanisms
4. **Error Resilience**: Handle missing or out-of-order events gracefully

## Security Considerations

1. **Authorization**: All write operations require proper authorization (JWT/signature)
2. **Encryption**: Sensitive data should be encrypted before storage
3. **Vault Security**:
   - Vaults use WebAssembly enclaves for tamper-resistant key protection
   - Private keys never leave the secure enclave environment
   - VaultKeeper enforces ownership verification for all operations
   - Refresh intervals prevent stale enclave states
4. **Permission Scoping**: Grant minimal required permissions
5. **Revocation**: Regularly review and revoke unused permissions
6. **VaultKeeper Security**:
   - Enclave attestation ensures execution environment integrity
   - Key generation uses cryptographically secure random sources
   - Signature operations are isolated within the enclave
   - Vault states are immutably stored in blockchain state

## Building and Testing

### Building the Motor Client

The DWN module includes a WebAssembly-based motor client for secure vault operations:

```bash
# Build the motor WASM client
make -C x/dwn motr

# Or from the root directory
make motr
```

The motor client will be built to `x/dwn/vaults/motr.wasm`.

### Running Tests

```bash
# Run unit tests
make -C x/dwn test

# Run tests with race detection
make -C x/dwn test-race

# Generate coverage report
make -C x/dwn test-cover

# Run benchmarks
make -C x/dwn benchmark
```

### Cleaning Build Artifacts

```bash
# Clean motor build artifacts
make -C x/dwn clean
```

## Dependencies

### Core Package Dependencies

The DWN module integrates with several core packages to provide enhanced functionality:

- **pkg/txns**: Multi-chain transaction building for Cosmos SDK and EVM networks
  - Transaction encoding/decoding in Protobuf, Amino, and RLP formats
  - Fee estimation and gas calculation for both transaction types
  - Enhanced address derivation using BIP44 HD wallet standards
  - Support for MPC-based signing with secure enclaves

- **pkg/coins**: Token and coin management for multi-chain operations
  - Standardized coin handling across different blockchain networks
  - Integration with transaction builders for proper fee calculation
  - Support for multiple denomination formats and conversions

- **github.com/sonr-io/crypto**: Enhanced cryptographic operations
  - Address derivation from public keys, entropy, and MPC enclaves
  - Multi-party computation support for distributed key management
  - Secure key generation and cryptographic primitives

### Architecture Integration

The DWN module's VaultKeeper leverages these packages to provide:

1. **Unified Transaction Interface**: Single API for building transactions across multiple blockchain networks
2. **Secure Key Management**: Integration with MPC enclaves for tamper-resistant key operations
3. **Multi-Chain Address Derivation**: Deterministic address generation for both Cosmos and EVM networks
4. **Enhanced Fee Management**: Automatic fee estimation and optimization for different transaction types

## Future Enhancements

- **Replication**: Multi-node data replication for availability
- **Sync Protocol**: Efficient data synchronization between nodes
- **Advanced Queries**: GraphQL-like query capabilities
- **Compression**: Automatic data compression for efficiency
- **IPFS Integration**: Content-addressed storage backend
- **Cross-Chain Interoperability**: Enhanced cross-chain transaction capabilities via pkg/txns
- **Advanced MPC Features**: Extended multi-party computation capabilities for complex cryptographic operations
