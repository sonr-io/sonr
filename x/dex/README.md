# `x/dex`

The Decentralized Exchange (DEX) module provides cross-chain trading capabilities through IBC Interchain Accounts (ICA), enabling users to perform swaps, manage liquidity, and execute orders on remote DEX chains while maintaining custody through their Sonr DID. This module bridges the gap between self-sovereign identity and DeFi operations across the Cosmos ecosystem.

## Overview

The DEX module provides:

- **Cross-Chain Trading**: Execute swaps on remote DEX chains via ICA
- **Liquidity Management**: Add and remove liquidity from pools across chains
- **Order Management**: Create and manage limit orders on compatible DEXs
- **DID-Controlled Accounts**: All operations authorized through Sonr DIDs
- **UCAN Authorization**: Fine-grained permissions for trading operations
- **Multi-DEX Support**: Connect to multiple DEX chains simultaneously
- **Rate Limiting**: Protection against spam and excessive operations
- **Activity Tracking**: Complete history of all DEX operations

## Core Concepts

### Interchain Accounts (ICA)

The module leverages IBC's Interchain Accounts to create controlled accounts on remote DEX chains. Each account is linked to a Sonr DID and managed through ICA transactions.

### DID-Based Authorization

All DEX operations require authorization from a valid Sonr DID, ensuring that only authenticated users can perform trading operations.

### UCAN Tokens

User-Controlled Authorization Network (UCAN) tokens provide delegated authority for specific operations, enabling secure third-party integrations.

### Cross-Chain Liquidity

Users can provide liquidity to pools on any supported DEX chain while maintaining custody through their Sonr identity.

## State

### Interchain DEX Accounts

```protobuf
message InterchainDEXAccount {
  string did = 1;                                    // DID controller of this account
  string connection_id = 2;                         // IBC connection to the remote chain
  string host_chain_id = 3;                         // Remote chain ID (e.g., osmosis-1)
  string account_address = 4;                       // Account address on the remote chain
  string port_id = 5;                               // ICA port ID for this account
  google.protobuf.Timestamp created_at = 6;         // Account creation timestamp
  repeated string enabled_features = 7;             // Enabled features for this account
  AccountStatus status = 8;                         // Current account status
}
```

### Account Status

```protobuf
enum AccountStatus {
  ACCOUNT_STATUS_PENDING = 0;    // Account is pending creation
  ACCOUNT_STATUS_ACTIVE = 1;     // Account is active and ready
  ACCOUNT_STATUS_DISABLED = 2;   // Account is temporarily disabled
  ACCOUNT_STATUS_FAILED = 3;     // Account creation failed
}
```

### DEX Features

```protobuf
enum DEXFeatures {
  DEX_FEATURE_SWAP = 0;         // Basic swap functionality
  DEX_FEATURE_LIQUIDITY = 1;    // Liquidity provision
  DEX_FEATURE_ORDERS = 2;       // Limit orders
  DEX_FEATURE_STAKING = 3;      // Staking operations
  DEX_FEATURE_GOVERNANCE = 4;   // Governance participation
}
```

### Module Parameters

```protobuf
message Params {
  bool enabled = 1;                              // Enable/disable the module
  uint32 max_accounts_per_did = 2;              // Maximum accounts per DID
  uint64 default_timeout_seconds = 3;           // Default timeout for ICA operations
  repeated string allowed_connections = 4;       // Allowed DEX connections
  string min_swap_amount = 5;                   // Minimum swap amount
  string max_daily_volume = 6;                  // Maximum daily volume per DID
  RateLimitParams rate_limits = 7;              // Rate limit parameters
  FeeParams fees = 8;                            // Fee parameters
}
```

### Rate Limiting

```protobuf
message RateLimitParams {
  uint32 max_ops_per_block = 1;         // Maximum operations per block
  uint32 max_ops_per_did_per_day = 2;   // Maximum operations per DID per day
  uint32 cooldown_blocks = 3;           // Cooldown period between operations
}
```

### Fee Parameters

```protobuf
message FeeParams {
  uint32 swap_fee_bps = 1;         // Platform fee for swaps (basis points)
  uint32 liquidity_fee_bps = 2;    // Platform fee for liquidity operations
  uint32 order_fee_bps = 3;        // Platform fee for orders
  string fee_collector = 4;        // Fee collector address
}
```

## Messages

### Account Management

#### MsgRegisterDEXAccount

Registers a new ICA account for DEX operations on a remote chain.

```protobuf
message MsgRegisterDEXAccount {
  string did = 1;                    // DID controller requesting the account
  string connection_id = 2;          // IBC connection to target chain
  repeated string features = 3;      // Requested features for this account
  string metadata = 4;                // Optional metadata
}
```

### Trading Operations

#### MsgExecuteSwap

Executes a token swap on a remote DEX chain.

```protobuf
message MsgExecuteSwap {
  string did = 1;                                  // DID initiating the swap
  string connection_id = 2;                        // IBC connection to DEX chain
  string source_denom = 3;                         // Token to swap from
  string target_denom = 4;                         // Token to swap to
  string amount = 5;                                // Amount to swap
  string min_amount_out = 6;                       // Minimum amount out (slippage protection)
  string route = 7;                                 // Optional specific route
  string ucan_token = 8;                           // UCAN authorization token
  google.protobuf.Timestamp timeout = 9;           // Timeout for the swap
}
```

### Liquidity Management

#### MsgProvideLiquidity

Adds liquidity to a pool on a remote DEX.

```protobuf
message MsgProvideLiquidity {
  string did = 1;                                  // DID providing liquidity
  string connection_id = 2;                        // IBC connection to DEX chain
  string pool_id = 3;                              // Pool ID to add liquidity to
  repeated cosmos.base.v1beta1.Coin assets = 4;   // Assets to provide
  string min_shares = 5;                           // Minimum shares to receive
  string ucan_token = 6;                           // UCAN authorization token
  google.protobuf.Timestamp timeout = 7;           // Timeout for the operation
}
```

#### MsgRemoveLiquidity

Removes liquidity from a pool on a remote DEX.

```protobuf
message MsgRemoveLiquidity {
  string did = 1;                                  // DID removing liquidity
  string connection_id = 2;                        // IBC connection to DEX chain
  string pool_id = 3;                              // Pool ID to remove liquidity from
  string shares = 4;                                // Amount of shares to remove
  repeated cosmos.base.v1beta1.Coin min_amounts = 5; // Minimum assets to receive
  string ucan_token = 6;                           // UCAN authorization token
  google.protobuf.Timestamp timeout = 7;           // Timeout for the operation
}
```

### Order Management

#### MsgCreateLimitOrder

Creates a limit order on a remote DEX.

```protobuf
message MsgCreateLimitOrder {
  string did = 1;                                  // DID creating the order
  string connection_id = 2;                        // IBC connection to DEX chain
  string sell_denom = 3;                           // Token to sell
  string buy_denom = 4;                            // Token to buy
  string amount = 5;                                // Amount to sell
  string price = 6;                                // Price per unit
  google.protobuf.Timestamp expiration = 7;        // Order expiration
  string ucan_token = 8;                           // UCAN authorization token
}
```

#### MsgCancelOrder

Cancels an existing order on a remote DEX.

```protobuf
message MsgCancelOrder {
  string did = 1;                    // DID canceling the order
  string connection_id = 2;          // IBC connection to DEX chain
  string order_id = 3;                // Order ID to cancel
  string ucan_token = 4;              // UCAN authorization token
}
```

## Queries

### Account Queries

- `Params`: Get module parameters
- `Account`: Query a specific DEX account by DID and connection
- `Accounts`: List all DEX accounts for a DID
- `Balance`: Query remote chain balance for an account

### Trading Queries

- `Pool`: Get pool information from a remote DEX
- `Orders`: Query orders for a DID on a specific connection
- `History`: Get transaction history for a DID

### Query Types

#### QueryAccountRequest/Response

```protobuf
message QueryAccountRequest {
  string did = 1;                    // DID of the account owner
  string connection_id = 2;          // IBC connection ID
}

message QueryAccountResponse {
  InterchainDEXAccount account = 1;  // The DEX account
}
```

#### QueryBalanceRequest/Response

```protobuf
message QueryBalanceRequest {
  string did = 1;                    // DID of the account owner
  string connection_id = 2;          // IBC connection ID
  string denom = 3;                  // Optional specific denom to query
}

message QueryBalanceResponse {
  repeated cosmos.base.v1beta1.Coin balances = 1;  // Balances on the remote chain
}
```

#### QueryHistoryRequest/Response

```protobuf
message QueryHistoryRequest {
  string did = 1;                                    // DID of the account owner
  string connection_id = 2;                          // Optional connection filter
  string operation_type = 3;                         // Optional operation type filter
  cosmos.base.query.v1beta1.PageRequest pagination = 4; // Pagination
}

message QueryHistoryResponse {
  repeated Transaction transactions = 1;             // Historical transactions
  cosmos.base.query.v1beta1.PageResponse pagination = 2; // Pagination response
}
```

## Activity Tracking

The module maintains comprehensive activity records for all DEX operations:

```protobuf
message DEXActivity {
  string type = 1;                                  // Activity type
  string did = 2;                                   // DID that performed the activity
  string connection_id = 3;                         // Connection where activity occurred
  string tx_hash = 4;                               // Transaction hash
  int64 block_height = 5;                           // Block height
  google.protobuf.Timestamp timestamp = 6;          // Activity timestamp
  string details = 7;                                // JSON-encoded details
  string status = 8;                                 // Activity status
  repeated cosmos.base.v1beta1.Coin amount = 9;    // Amount involved
  uint64 gas_used = 10;                             // Gas used for the activity
}
```

## Events

The DEX module emits comprehensive events for all operations, enabling efficient tracking and indexing of DEX activities.

### Trading Events

#### EventSwapExecuted
- **Emitted**: When a swap is successfully executed
- **Fields**:
  - `did`: DID of the trader
  - `connection_id`: IBC connection ID
  - `source`: Source token and amount
  - `target`: Target token and amount received
  - `tx_hash`: Transaction hash on remote chain
  - `sequence`: IBC packet sequence

#### EventLiquidityProvided
- **Emitted**: When liquidity is added to a pool
- **Fields**:
  - `did`: DID of the liquidity provider
  - `connection_id`: IBC connection ID
  - `pool_id`: Pool identifier
  - `assets`: Assets provided
  - `shares_received`: LP tokens received
  - `tx_hash`: Transaction hash on remote chain

#### EventLiquidityRemoved
- **Emitted**: When liquidity is removed from a pool
- **Fields**:
  - `did`: DID of the liquidity provider
  - `connection_id`: IBC connection ID
  - `pool_id`: Pool identifier
  - `shares_removed`: LP tokens burned
  - `assets`: Assets received
  - `tx_hash`: Transaction hash on remote chain

### Order Events

#### EventOrderCreated
- **Emitted**: When a limit order is created
- **Fields**:
  - `did`: DID of the trader
  - `connection_id`: IBC connection ID
  - `order_id`: Order identifier
  - `sell_denom`: Token to sell
  - `buy_denom`: Token to buy
  - `amount`: Order amount
  - `price`: Order price
  - `tx_hash`: Transaction hash on remote chain

#### EventOrderCancelled
- **Emitted**: When an order is cancelled
- **Fields**:
  - `did`: DID of the trader
  - `connection_id`: IBC connection ID
  - `order_id`: Cancelled order ID
  - `tx_hash`: Transaction hash on remote chain

#### EventOrderFilled
- **Emitted**: When an order is filled (partially or fully)
- **Fields**:
  - `did`: DID of the trader
  - `connection_id`: IBC connection ID
  - `order_id`: Filled order ID
  - `fill_amount`: Amount filled
  - `fill_price`: Fill price
  - `tx_hash`: Transaction hash on remote chain

### ICA Events

#### EventDEXAccountRegistered
- **Emitted**: When a new DEX account is registered
- **Fields**:
  - `did`: DID of the account owner
  - `connection_id`: IBC connection ID
  - `port_id`: Generated ICA port ID
  - `account_address`: Remote account address

#### EventICAPacketSent
- **Emitted**: When an ICA packet is sent
- **Fields**:
  - `did`: DID of the sender
  - `connection_id`: IBC connection ID
  - `packet_type`: Type of packet (swap, liquidity, order)
  - `sequence`: IBC packet sequence

#### EventICAPacketAcknowledged
- **Emitted**: When an ICA packet is acknowledged
- **Fields**:
  - `did`: DID of the sender
  - `connection_id`: IBC connection ID
  - `packet_type`: Type of packet
  - `sequence`: IBC packet sequence
  - `success`: Success status
  - `error`: Error message if failed

## CLI Examples

### Account Management

```bash
# Register a new DEX account on Osmosis
snrd tx dex register-account \
  --did did:sonr:alice \
  --connection connection-0 \
  --features swap,liquidity,orders \
  --from alice

# Query DEX account
snrd query dex account did:sonr:alice connection-0

# List all DEX accounts for a DID
snrd query dex accounts did:sonr:alice

# Check balance on remote chain
snrd query dex balance did:sonr:alice connection-0
```

### Trading Operations

```bash
# Execute a swap on Osmosis
snrd tx dex swap \
  --did did:sonr:alice \
  --connection connection-0 \
  --source-denom uosmo \
  --target-denom uatom \
  --amount 1000000 \
  --min-amount-out 950000 \
  --ucan-token "eyJ0eXAiOi..." \
  --from alice

# Provide liquidity to a pool
snrd tx dex provide-liquidity \
  --did did:sonr:alice \
  --connection connection-0 \
  --pool-id 1 \
  --assets 1000000uosmo,500000uatom \
  --min-shares 100000 \
  --ucan-token "eyJ0eXAiOi..." \
  --from alice

# Remove liquidity from a pool
snrd tx dex remove-liquidity \
  --did did:sonr:alice \
  --connection connection-0 \
  --pool-id 1 \
  --shares 100000 \
  --min-amounts 990000uosmo,495000uatom \
  --ucan-token "eyJ0eXAiOi..." \
  --from alice
```

### Order Management

```bash
# Create a limit order
snrd tx dex create-order \
  --did did:sonr:alice \
  --connection connection-0 \
  --sell-denom uosmo \
  --buy-denom uatom \
  --amount 1000000 \
  --price 1.2 \
  --expiration "2024-12-31T23:59:59Z" \
  --ucan-token "eyJ0eXAiOi..." \
  --from alice

# Cancel an order
snrd tx dex cancel-order \
  --did did:sonr:alice \
  --connection connection-0 \
  --order-id order-123 \
  --ucan-token "eyJ0eXAiOi..." \
  --from alice

# Query orders
snrd query dex orders did:sonr:alice connection-0 --status active
```

### Analytics and History

```bash
# Query transaction history
snrd query dex history did:sonr:alice \
  --connection connection-0 \
  --operation-type swap

# Get pool information
snrd query dex pool connection-0 pool-1

# Query module parameters
snrd query dex params
```

## Integration Guide

### For DApp Developers

1. **Account Setup**: Register ICA accounts for target DEX chains
2. **Permission Management**: Issue UCAN tokens for specific operations
3. **Execute Trades**: Use the module's messages to perform DEX operations
4. **Monitor Activity**: Subscribe to events for real-time updates
5. **Query State**: Use queries to display balances and history

### For DEX Integration

1. **IBC Connection**: Establish IBC connection to Sonr
2. **ICA Support**: Ensure ICA host module is enabled
3. **Message Handling**: Support standard Cosmos SDK messages
4. **Event Emission**: Emit appropriate events for tracking

### For Wallet Developers

1. **DID Integration**: Support Sonr DID authentication
2. **UCAN Generation**: Implement UCAN token creation
3. **Transaction Building**: Build DEX module transactions
4. **History Display**: Query and display DEX activity

## Technical Architecture

### ICA Message Flow

1. **Message Creation**: User creates DEX operation message
2. **DID Verification**: Module verifies DID ownership
3. **UCAN Validation**: Validates authorization token
4. **ICA Packet**: Constructs ICA packet for remote chain
5. **IBC Relay**: Packet sent via IBC to target chain
6. **Remote Execution**: Operation executed on DEX chain
7. **Acknowledgment**: Result returned via IBC
8. **Event Emission**: Events emitted for tracking

### Rate Limiting System

The module implements multi-layer rate limiting:

```go
// Per-block rate limiting
if opsThisBlock >= params.RateLimits.MaxOpsPerBlock {
    return errorsmod.Wrap(ErrRateLimited, "max operations per block exceeded")
}

// Per-DID daily rate limiting
if opsToday >= params.RateLimits.MaxOpsPerDidPerDay {
    return errorsmod.Wrap(ErrRateLimited, "daily operation limit exceeded")
}

// Cooldown period enforcement
if blocksSinceLastOp < params.RateLimits.CooldownBlocks {
    return errorsmod.Wrap(ErrCooldown, "operation cooldown period active")
}
```

### Fee Collection

Platform fees are collected on successful operations:

```go
// Calculate platform fee
fee := amount.Mul(params.Fees.SwapFeeBps).Quo(10000)

// Transfer fee to collector
err := bankKeeper.SendCoins(ctx, userAddr, feeCollector, fee)
```

## Security Considerations

1. **DID Authentication**: All operations require valid DID signatures
2. **UCAN Authorization**: Fine-grained permissions prevent unauthorized operations
3. **Rate Limiting**: Protects against spam and DoS attacks
4. **Slippage Protection**: Minimum output amounts prevent sandwich attacks
5. **Timeout Enforcement**: Operations expire to prevent stale execution
6. **Connection Whitelisting**: Only approved IBC connections allowed
7. **Volume Limits**: Daily volume caps prevent excessive exposure
8. **ICA Security**: Leverages IBC's security guarantees
9. **Event Auditing**: Comprehensive event trail for all operations
10. **Fee Mechanisms**: Platform fees discourage spam

## Performance Optimization

### Batch Operations

The module supports batching for improved efficiency:
- Multiple swaps in single ICA packet
- Bulk order creation/cancellation
- Aggregated liquidity operations

### Caching Strategy

- Account data cached for quick lookups
- Pool information cached with TTL
- Order book snapshots for fast queries

### Query Optimization

- Indexed by DID for fast account lookups
- Pagination for large result sets
- Filtered queries for specific operations

## Supported DEX Chains

### Currently Supported

- **Osmosis**: Full swap, liquidity, and order support
- **Crescent**: Swap and liquidity operations
- **Neutron**: Astroport DEX integration

### Planned Support

- **Kujira**: FIN orderbook integration
- **Injective**: Derivatives and spot trading
- **Sei**: High-frequency trading support

## Building and Testing

### Running Tests

```bash
# Run unit tests
make test-dex

# Run integration tests with IBC
make test-dex-ibc

# Run benchmark tests
make benchmark-dex

# Test with specific DEX chain
make test-dex-osmosis
```

### Local Development

```bash
# Start local chain with ICA enabled
make localnet-dex

# Deploy test DEX contracts
make deploy-test-dex

# Run E2E test suite
make e2e-test-dex
```

## Future Enhancements

- **Advanced Order Types**: Stop-loss, trailing stops, OCO orders
- **Cross-Chain Arbitrage**: Automated arbitrage between DEXs
- **Portfolio Management**: Automated rebalancing strategies
- **Yield Farming**: Integration with liquidity mining programs
- **Derivatives Trading**: Support for perpetuals and options
- **MEV Protection**: Private mempool submission for sensitive trades
- **Analytics Dashboard**: Real-time trading metrics and P&L tracking
- **Social Trading**: Copy trading and strategy sharing
- **DeFi Aggregation**: Route optimization across multiple DEXs
- **Governance Integration**: Participate in DEX governance