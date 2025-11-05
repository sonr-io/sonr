# USDC Swap with DID E2E Tests

Comprehensive end-to-end tests for USDC swap functionality with DID-based Interchain Accounts (ICA) on Noble testnet.

## Overview

This test suite validates the complete USDC swap flow including:
- DID creation and management
- DEX account registration via ICA
- Cross-chain USDC swaps (SNR ↔ USDC)
- Noble testnet IBC integration
- Swap routing and slippage protection
- Event verification and transaction history

## Prerequisites

### Network Setup

1. **Sonr Testnet**: Running local testnet with DEX module enabled
2. **Noble Testnet Connection**: IBC connection established to Noble testnet (grand-1)
3. **IBC Relayer**: Active relayer between Sonr and Noble chains
4. **Test Tokens**:
   - SNR tokens for gas and swaps
   - USDC tokens on Noble testnet (optional for full E2E)

### Configuration

The tests use these default configurations:

```go
// Noble Testnet
NobleChainID     = "noble-grand-1"
NobleConnectionID = "connection-noble"  // Update with actual connection
NobleUSDCDenom   = "uusdc"

// Test Amounts
USDCTestAmount   = 1_000_000  // 1 USDC (6 decimals)
SNRSwapAmount    = 1_000_000  // 1 SNR (6 decimals)
SlippageTolerance = 0.05      // 5% slippage
```

## Test Suite Structure

### Test Cases

#### 01. Verify Noble Configuration
**Purpose**: Validates Noble chain configuration and constants
**Validates**:
- Noble chain ID recognition
- USDC denomination configuration
- Decimal places (6 for USDC)
- Chain type detection

#### 02. Create DID Document
**Purpose**: Creates a test DID for swap operations
**Operations**:
- Generate unique DID for test
- Create DID document on-chain
- Verify DID can be queried
- Validate DID structure

#### 03. Register DEX Account
**Purpose**: Registers an ICA account on Noble connection
**Operations**:
- Register DEX account with Noble connection
- Enable swap and liquidity features
- Wait for ICA channel handshake
- Verify account status and ICA address

**Expected States**:
- `ACCOUNT_STATUS_PENDING`: Initial state during handshake
- `ACCOUNT_STATUS_ACTIVE`: Account ready for operations
- `ACCOUNT_STATUS_FAILED`: Handshake failed

#### 04. Execute USDC Swap (SNR → USDC)
**Purpose**: Tests swapping SNR to USDC on Noble
**Flow**:
1. Setup DID and DEX account
2. Calculate minimum output with slippage
3. Execute swap via ICA
4. Verify transaction success
5. Check swap event emission

**Parameters**:
```go
Amount:       1_000_000 usnr (1 SNR)
MinAmountOut: 950_000 uusdc (0.95 USDC - 5% slippage)
Timeout:      60 seconds
```

#### 05. Execute USDC Swap (USDC → SNR)
**Purpose**: Tests reverse swap (USDC to SNR)
**Flow**: Same as Test 04 but with reversed denoms

#### 06. Swap With Invalid Parameters
**Purpose**: Tests error handling and validation
**Test Cases**:
- Zero amount swap
- Same source/target denom
- Negative minimum output
- Invalid connection ID

**Expected**: All cases should fail gracefully

#### 07. Swap With UCAN Permission
**Purpose**: Tests UCAN-authorized delegated swaps
**Status**: Currently skipped (requires full UCAN setup)
**Features**:
- UCAN token generation
- Permission validation
- Delegated execution

#### 08. Query DEX History
**Purpose**: Verifies transaction history tracking
**Validates**:
- History query endpoint
- Activity records
- Transaction metadata
- Status tracking

#### 09. Multiple Sequential Swaps
**Purpose**: Tests executing multiple swaps in sequence
**Operations**:
- Execute 3 swaps sequentially
- Verify each transaction succeeds
- Check for nonce/sequence handling
- Validate no race conditions

## Running the Tests

### Quick Start

```bash
# Navigate to test directory
cd test/e2e/usdc-swap-did

# Run all tests
go test -v

# Run specific test
go test -v -run TestUSDCSwapE2E/Test04_ExecuteUSDCSwap_SNRToUSDC

# Run with timeout
go test -v -timeout 10m
```

### Prerequisites Check

Before running tests, ensure:

```bash
# 1. Verify Noble connection exists
snrd query ibc connection end connection-noble

# 2. Check connection is open
# Expected output: state: STATE_OPEN

# 3. Verify ICA channel
snrd query ibc channel channels --node tcp://localhost:26657

# 4. Check DEX module is enabled
snrd query dex params
```

### Full E2E Setup

```bash
# 1. Start Sonr testnet
make testnet

# 2. Establish Noble connection (one-time setup)
# This requires:
# - Noble testnet access
# - IBC relayer configuration
# - Channel creation

# 3. Fund test accounts
# The tests will auto-fund via faucet, but you can pre-fund:
snrd tx bank send validator <test-address> 100000000usnr --yes

# 4. Run tests
cd test/e2e/usdc-swap-did && go test -v
```

## Test Utilities

### Helper Functions

#### DID Operations
```go
// Create test DID
did, err := CreateTestDID(ctx, cfg, account, didID)

// Query DID
did, err := QueryDID(ctx, cfg, didID)
```

#### DEX Operations
```go
// Query DEX account
account, err := QueryDEXAccount(ctx, cfg, didID, connectionID)

// Query transaction history
history, err := QueryDEXHistory(ctx, cfg, didID)

// Wait for account activation
active, err := WaitForDEXAccountActivation(ctx, cfg, didID, connectionID, timeout)
```

#### Swap Utilities
```go
// Validate parameters
err := ValidateSwapParameters(sourceDenom, targetDenom, amount, minOut)

// Calculate slippage
minOut := CalculateMinimumOutput(amount, slippagePct)

// Format amounts for display
formatted := FormatUSDCAmount(1000000) // "1.000000 USDC"
formatted = FormatSNRAmount(1000000)   // "1.000000 SNR"
```

### Assertions

The test suite uses `testify/suite` and `testify/require` for assertions:

```go
// Transaction success
require.NoError(t, err)
require.Equal(t, uint32(0), txResp.Code)

// Balance checks
balance := s.getBalance(address, denom)
require.True(t, balance.GT(minAmount))

// Event verification
s.verifySwapEvent(txHash, did, connectionID)
```

## Configuration

### Update Connection ID

Before running tests, update the Noble connection ID in `usdc_swap_test.go`:

```go
const (
    NobleConnectionID = "connection-0"  // Replace with actual connection ID
)
```

To find the connection ID:

```bash
# List all connections
snrd query ibc connection connections

# Look for connection to Noble (grand-1)
# Update NobleConnectionID constant with the correct value
```

### Custom Test Configuration

Create a custom config for your environment:

```go
cfg := &utils.TestConfig{
    ChainID:        "sonrtest_1-1",
    BaseURL:        "http://localhost:1317",
    NobleConnection: "connection-0",  // Your Noble connection
    // ... other settings
}
```

## Troubleshooting

### Common Issues

#### 1. Connection Not Found
```
Error: connection connection-noble not found
```
**Solution**: Update `NobleConnectionID` constant with actual connection ID

#### 2. ICA Account Pending
```
⚠ ICA address pending (channel handshake may be in progress)
```
**Solution**: Wait for ICA channel handshake to complete (can take 1-2 minutes)

#### 3. Swap Transaction Fails
```
Error: DEX account not active
```
**Solution**: Ensure ICA account is in `ACCOUNT_STATUS_ACTIVE` state

#### 4. Insufficient Funds
```
Error: insufficient funds
```
**Solution**: Fund test account with SNR tokens

### Debug Mode

Enable verbose logging:

```bash
# Run with verbose output
go test -v -run TestUSDCSwapE2E 2>&1 | tee test.log

# Check specific test output
grep "Test 04" test.log

# View all swap events
grep "Swap" test.log
```

### Verify IBC Setup

```bash
# Check IBC clients
snrd query ibc client states

# Check channels
snrd query ibc channel channels

# Check connections
snrd query ibc connection connections

# Monitor relayer
# (depends on your relayer setup)
```

## Integration with CI/CD

### GitHub Actions

```yaml
name: USDC Swap E2E Tests

on: [push, pull_request]

jobs:
  e2e-usdc-swap:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start Testnet
        run: make testnet

      - name: Wait for Network
        run: sleep 30

      - name: Run E2E Tests
        run: |
          cd test/e2e/usdc-swap-did
          go test -v -timeout 10m
```

## Expected Output

### Successful Test Run

```
=== RUN   TestUSDCSwapE2E
=== RUN   TestUSDCSwapE2E/Test01_VerifyNobleConfiguration
=== Test 01: Verify Noble Configuration ===
✓ Noble configuration verified successfully
=== RUN   TestUSDCSwapE2E/Test02_CreateDIDDocument
=== Test 02: Create DID Document ===
✓ DID document created: did:snr:test_1234567890
✓ DID query verification successful
=== RUN   TestUSDCSwapE2E/Test03_RegisterDEXAccount
=== Test 03: Register DEX Account ===
✓ DEX account registered, tx hash: ABC123...
✓ DEX account status: ACCOUNT_STATUS_ACTIVE
✓ ICA port ID: icacontroller-did:snr:test_1234567890-connection-0
✓ ICA address: noble1abc...xyz
=== RUN   TestUSDCSwapE2E/Test04_ExecuteUSDCSwap_SNRToUSDC
=== Test 04: Execute USDC Swap (SNR → USDC) ===
Initial SNR balance: 10000000
✓ Swap transaction broadcasted, tx hash: DEF456...
✓ Found swap event for DID: did:snr:test_1234567890
✓ Swap execution test completed
...
--- PASS: TestUSDCSwapE2E (45.23s)
PASS
ok      github.com/sonr-io/sonr/test/e2e/usdc-swap-did    45.234s
```

## Future Enhancements

### Planned Features
1. **Multi-hop Swaps**: Route through multiple pools/chains
2. **Liquidity Provision**: Add/remove liquidity tests
3. **Limit Orders**: Create and execute limit orders
4. **UCAN Integration**: Full delegation testing
5. **Performance Tests**: Load testing and benchmarks
6. **Failure Recovery**: Test ICA timeout handling

### Contributing

When adding new tests:
1. Follow existing naming convention (Test##_Description)
2. Add helper functions to `helpers.go`
3. Update this README with new test documentation
4. Ensure tests are idempotent
5. Add appropriate logging for debugging

## References

- [Noble Documentation](https://docs.noble.xyz)
- [IBC Protocol](https://ibc.cosmos.network)
- [Sonr DEX Module](../../../../x/dex/README.md)
- [DID Module](../../../../x/did/README.md)
- [E2E Testing Framework](../README.md)
