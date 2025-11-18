# Noble Testnet Integration for Sonr DEX Module

## Overview

This document describes the integration of the Noble testnet with Sonr's x/dex module. Noble is a Cosmos appchain purpose-built for native asset issuance, particularly USDC, in the IBC ecosystem.

## What is Noble?

Noble is an application-specific blockchain built on the Cosmos SDK that serves as the canonical issuance hub for USDC and other real-world assets (RWAs) in the Cosmos ecosystem. Key features:

- **Native USDC Issuance**: Circle's official USDC is natively issued on Noble
- **IBC Connectivity**: Seamlessly transfers assets across 50+ IBC-enabled chains
- **Packet Forwarding**: Enables "1-click" asset transfers between chains
- **CCTP Integration**: Circle's Cross-Chain Transfer Protocol for bridging from EVM chains

## Noble Testnet Configuration

### Chain Details

- **Chain ID**: `noble-grand-1`
- **RPC Endpoint**: `https://noble-testnet-rpc.polkachu.com:443`
- **gRPC Endpoint**: `noble-testnet-grpc.polkachu.com:21590`
- **USDC Denomination**: `uusdc` (micro-USDC, 6 decimals)

### Default Configuration

The Noble testnet is included in the default DEX module parameters:

```go
AllowedConnections: []string{
    "noble-grand-1",  // Noble testnet - USDC hub
    "osmo-test-5",    // Osmosis testnet
}
```

## Integration Features

### 1. USDC Helper Functions

The integration includes helper functions for working with USDC:

```go
// Convert base units to USDC
usdcAmount := ConvertToUSDC(sdk.NewInt(1500000)) // Returns 1.5 USDC

// Convert USDC to base units
baseUnits := ConvertFromUSDC(sdk.MustNewDecFromStr("1.5")) // Returns 1500000

// Format for display
formatted := FormatUSDCAmount(sdk.NewInt(1500000)) // Returns "1.500000 USDC"

// Parse from string
amount, err := ParseUSDCAmount("1.5") // Returns 1500000 base units
```

### 2. Chain Configuration

```go
// Get Noble testnet configuration
config := GetNobleTestnetConfig()
// Returns: NobleChainConfig{
//     ChainID:      "noble-grand-1",
//     RPCEndpoint:  "https://noble-testnet-rpc.polkachu.com:443",
//     GRPCEndpoint: "noble-testnet-grpc.polkachu.com:21590",
//     USDCDenom:    "uusdc",
//     Decimals:     6,
// }

// Check if a chain is Noble
isNoble := IsNobleChain("noble-grand-1") // Returns true
```

### 3. Trading Pairs

Common USDC trading pairs are predefined:

```go
pairs := GetNobleUSDCPairs()
// Returns pairs like:
// - ATOM/USDC
// - OSMO/USDC
// - AKT/USDC
// - JUNO/USDC
// - STARS/USDC
```

### 4. Swap Parameters

Structured parameters for Noble swaps:

```go
swapParams := NobleSwapParams{
    InputDenom:  "uatom",
    OutputDenom: "uusdc",
    Amount:      sdk.NewInt(1000000), // 1 ATOM
    MinOutput:   sdk.NewInt(5000000), // Min 5 USDC
    Receiver:    "sonr1...",
}

if err := swapParams.Validate(); err != nil {
    // Handle validation error
}
```

### 5. Liquidity Parameters

Parameters for providing liquidity:

```go
liquidityParams := NobleLiquidityParams{
    PoolID:    "1",
    Token0:    "uusdc",
    Token1:    "uatom",
    Amount0:   sdk.NewInt(1000000), // 1 USDC
    Amount1:   sdk.NewInt(100000),  // 0.1 ATOM
    MinShares: sdk.NewInt(1),
}

if err := liquidityParams.Validate(); err != nil {
    // Handle validation error
}
```

## Usage Examples

### Registering a DEX Account for Noble

```bash
# Register an Interchain Account on Noble testnet
sonrd tx dex register-dex-account \
    did:sonr:user:abc123 \
    connection-0 \
    noble-grand-1 \
    --from mykey \
    --chain-id sonr_1-1
```

### Executing a Swap via Noble

```bash
# Swap ATOM for USDC on Noble
sonrd tx dex execute-swap \
    did:sonr:user:abc123 \
    connection-0 \
    uatom \
    uusdc \
    1000000 \
    5000000 \
    --from mykey \
    --chain-id sonr_1-1
```

### Querying USDC Balance

```bash
# Query USDC balance on Noble for a DEX account
sonrd query dex balance \
    did:sonr:user:abc123 \
    connection-0 \
    uusdc
```

## Module Parameters

The DEX module includes the following default parameters for Noble integration:

```yaml
params:
  enabled: true
  max_accounts_per_did: 5
  default_timeout_seconds: 600
  allowed_connections:
    - noble-grand-1  # Noble testnet
    - osmo-test-5    # Osmosis testnet
  min_swap_amount: "1000"
  max_daily_volume: "1000000000000"
  rate_limits:
    max_ops_per_block: 10
    max_ops_per_did_per_day: 100
    cooldown_blocks: 5
  fees:
    swap_fee_bps: 30      # 0.3%
    liquidity_fee_bps: 20 # 0.2%
    order_fee_bps: 10     # 0.1%
    fee_collector: ""
```

## IBC Connection Setup

To establish an IBC connection with Noble testnet:

### 1. Create IBC Client

```bash
# Create IBC client for Noble on Sonr chain
hermes create client \
    --host-chain sonr_1-1 \
    --reference-chain noble-grand-1
```

### 2. Create IBC Connection

```bash
# Create IBC connection
hermes create connection \
    --a-chain sonr_1-1 \
    --b-chain noble-grand-1
```

### 3. Create Transfer Channel

```bash
# Create transfer channel
hermes create channel \
    --a-chain sonr_1-1 \
    --a-connection connection-0 \
    --a-port transfer \
    --b-port transfer
```

### 4. Verify Connection

```bash
# Query IBC connections
sonrd query ibc connection connections

# Query IBC channels
sonrd query ibc channel channels
```

## Security Considerations

### 1. Connection Whitelisting

Only connections listed in `allowed_connections` parameter can be used:

```go
func ValidateNobleConnection(connectionID string, allowedConnections []string) error {
    for _, allowed := range allowedConnections {
        if connectionID == allowed {
            return nil
        }
    }
    return fmt.Errorf("connection %s not in allowed connections list", connectionID)
}
```

### 2. Rate Limiting

The module implements rate limiting to prevent abuse:

- **Per Block**: Maximum 10 operations per block
- **Per DID**: Maximum 100 operations per day
- **Cooldown**: 5 block cooldown between operations

### 3. Amount Validation

All amounts are validated before execution:

- Minimum swap amount: 1000 base units
- Maximum daily volume per DID: 1,000,000,000,000 base units
- Positive amount checks on all operations

### 4. UCAN Authorization

All DEX operations require valid UCAN tokens:

```go
// UCAN capabilities for DEX operations
capabilities := []string{
    "dex/swap",
    "dex/liquidity/provide",
    "dex/liquidity/remove",
    "dex/order/create",
    "dex/order/cancel",
}
```

## Integration Testing

### Unit Tests

```bash
# Run DEX module tests
cd x/dex
go test -v ./...
```

### E2E Tests

```bash
# Run end-to-end tests
cd test/e2e
go test -v -run TestDEXModuleOperations
```

### Manual Testing

1. **Start local testnet**:
   ```bash
   cd networks/testnet
   docker-compose up -d
   ```

2. **Register account**:
   ```bash
   sonrd tx dex register-dex-account did:sonr:test connection-0 noble-grand-1 --from test
   ```

3. **Execute swap**:
   ```bash
   sonrd tx dex execute-swap did:sonr:test connection-0 uatom uusdc 1000000 500000 --from test
   ```

## Roadmap

### Phase 1: Testnet Integration ✅
- [x] Add Noble testnet to allowed connections
- [x] Create USDC helper functions
- [x] Implement chain configuration
- [x] Add documentation

### Phase 2: Enhanced Features 🚧
- [ ] Implement actual ICA swap execution
- [ ] Add multi-hop routing via USDC
- [ ] Integrate with Osmosis pools
- [ ] Add limit order support

### Phase 3: Mainnet Deployment 📋
- [ ] Security audit
- [ ] Add Noble mainnet configuration
- [ ] Implement advanced slippage protection
- [ ] Add monitoring and alerting

## Resources

### Noble Documentation
- **Developer Hub**: https://www.noble.xyz/dev-hub
- **Docs**: https://docs.noble.xyz
- **GitHub**: https://github.com/noble-assets/noble

### Cosmos IBC
- **IBC Protocol**: https://ibc.cosmos.network
- **ICA Controller**: https://github.com/cosmos/ibc-go/tree/main/modules/apps/27-interchain-accounts

### Circle USDC
- **USDC on Noble**: https://www.circle.com/multi-chain-usdc/noble
- **CCTP**: https://www.circle.com/en/cross-chain-transfer-protocol

## Support

For questions or issues with the Noble integration:

1. Check the [DEX Module README](README.md)
2. Review the [Noble documentation](https://docs.noble.xyz)
3. Open an issue on GitHub
4. Join the Sonr Discord community

## License

This integration is part of the Sonr blockchain and follows the same license terms.
