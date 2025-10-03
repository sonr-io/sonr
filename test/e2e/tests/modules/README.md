# Event Emission E2E Tests

This directory contains comprehensive End-to-End (E2E) tests for event emissions across Sonr blockchain modules, specifically focusing on the newly implemented typed Protobuf events for the DID and DWN modules.

## Test Coverage

### DID Module Events
- `EventDIDCreated` - Emitted when a new DID is created
- `EventDIDUpdated` - Emitted when a DID is updated  
- `EventDIDDeactivated` - Emitted when a DID is deactivated
- `EventVerificationMethodAdded` - Emitted when a verification method is added
- `EventVerificationMethodRemoved` - Emitted when a verification method is removed ⭐
- `EventServiceAdded` - Emitted when a service is added to a DID ⭐
- `EventServiceRemoved` - Emitted when a service is removed from a DID ⭐
- `EventWebAuthnRegistered` - Emitted when a WebAuthn credential is registered ⭐
- `EventExternalWalletLinked` - Emitted when an external wallet is linked ⭐

### DWN Module Events  
- `EventRecordWritten` - Emitted when a record is written to DWN
- `EventRecordDeleted` - Emitted when a record is deleted from DWN
- `EventProtocolConfigured` - Emitted when a protocol is configured ⭐
- `EventPermissionGranted` - Emitted when a permission is granted ⭐
- `EventPermissionRevoked` - Emitted when a permission is revoked ⭐
- `EventVaultCreated` - Emitted when a vault is created ⭐
- `EventVaultKeysRotated` - Emitted when vault keys are rotated ⭐

⭐ = Newly implemented events being tested

## Test Structure

### `events_test.go`

The main test file contains the following test suites:

#### `EventEmissionTestSuite`
Main test suite that validates:

1. **Real Transaction Event Emissions** (`TestDIDModuleEventEmissions`, `TestDWNModuleEventEmissions`)
   - Executes actual transactions that trigger events
   - Verifies events are emitted correctly with proper attributes
   - Tests each event type individually

2. **Event Persistence and Replay** (`TestEventPersistenceAndReplay`)  
   - Verifies events persist across multiple queries
   - Tests event queryability by attributes
   - Ensures event data consistency over time

3. **Event Querying** (`TestEventQuerying`)
   - Tests CometBFT query syntax patterns
   - Validates filtering by event type, creator, and custom attributes
   - Tests complex query conditions

4. **Multi-Event Transactions** (`TestMultiEventTransactions`)
   - Tests transactions that emit multiple events
   - Verifies correct event ordering
   - Validates block height consistency across events

5. **Event Subscription** (`TestEventSubscription`)
   - Tests WebSocket-based event subscription via CometBFT
   - Subscribes to new blocks, transactions, and custom events
   - Validates real-time event streaming

6. **Event Attribute Validation** (`TestEventAttributeValidation`)
   - Verifies all required attributes are present
   - Validates attribute values are correctly populated
   - Tests attribute consistency

## Client Extensions

### Enhanced StarshipClient (`client/chain.go`)

Extended the existing StarshipClient with comprehensive event querying capabilities:

- `QueryEventsByHeight(height)` - Query events by block height
- `QueryEventsByType(eventType, minHeight, maxHeight)` - Query by event type
- `QueryEventsByAttribute(key, value, minHeight, maxHeight)` - Query by attribute
- `SearchEvents(query, minHeight, maxHeight)` - General CometBFT query search
- `GetLatestBlockHeight()` - Get current block height
- `WaitForNextBlock()` - Wait for next block production
- `FilterEventsByType(events, eventType)` - Filter events by type
- `GetEventAttribute(event, key)` - Extract specific attribute values

### WebSocket Client (`client/websocket.go`)

New WebSocket client for real-time event subscription:

- `Connect()` - Establish WebSocket connection to CometBFT
- `Subscribe(query)` - Subscribe to events matching query
- `SubscribeToNewBlocks()` - Subscribe to new block events
- `SubscribeToTxEvents()` - Subscribe to transaction events  
- `SubscribeToDIDEvents()` - Subscribe to DID module events
- `SubscribeToDWNEvents()` - Subscribe to DWN module events
- `WaitForEvent(timeout, filter)` - Wait for specific events
- `WaitForEventByType(timeout, eventType)` - Wait for events by type
- `Unsubscribe()` - Unsubscribe from events

## Running the Tests

### Prerequisites

1. **Start the Sonr testnet:**
   ```bash
   make testnet  # or make start
   ```

2. **Ensure IPFS is running** (required for DWN tests):
   ```bash
   make ipfs-up
   ```

3. **Verify chain is running:**
   ```bash
   curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
   ```

### Run Event Tests

```bash
# Run all event tests
cd test/e2e
go test -v ./tests/modules/ -run TestEventEmission

# Run specific test suites  
go test -v ./tests/modules/ -run TestEventEmissionTestSuite/TestDIDModuleEventEmissions
go test -v ./tests/modules/ -run TestEventEmissionTestSuite/TestEventSubscription
go test -v ./tests/modules/ -run TestEventEmissionTestSuite/TestEventPersistenceAndReplay

# Run with detailed logging
go test -v ./tests/modules/ -run TestEventEmission -args -test.v
```

### Run Integration Tests (for comparison)
```bash
# Run the existing integration tests
cd ../..
go test -v ./test/ -run TestEventIntegration
```

## Configuration

### Default Test Configuration (`utils/utils.go`)

- **Chain ID**: `sonrtest_1-1`
- **Base URL**: `http://localhost:1317` (REST API)
- **WebSocket URL**: `ws://localhost:26657/websocket` (CometBFT WebSocket)
- **Staking Denom**: `usnr`
- **Normal Denom**: `snr`

### Pre-funded Test Accounts

The tests use pre-funded localnet accounts:
- `idx1fcqk3crpnyvyhtd4jepsnx5eat5ehc920epq29` (Account 0)
- `idx10n78mn09nx0f056wam35wkfvanf37kepuj28x4` (Account 1)  
- `idx1xygwjmmj8rq3rq3k4adqvhd55x5yqjc8ktcm7e` (Account 2)

## Implementation Status

### ✅ Completed Features

1. **Event Querying Infrastructure**
   - REST API event queries
   - Block height filtering
   - Attribute-based filtering
   - CometBFT query syntax support

2. **WebSocket Event Subscription**
   - Real-time event streaming
   - Custom query subscriptions
   - Event filtering and waiting

3. **Test Framework**
   - Comprehensive test structure
   - Mock transaction creation
   - Event validation helpers
   - Multi-event testing

### 🚧 In Progress / TODO

1. **Real Transaction Building**
   - Currently using mock transactions for testing
   - Need to implement actual DID/DWN message building and signing
   - Integration with existing transaction building utilities

2. **Complete Event Coverage**
   - Some event tests are marked as "Skip" pending real transaction implementation
   - Need to create actual transactions for each event type

3. **Chain Restart Testing**
   - Event persistence across chain restarts
   - Historical event replay validation

4. **Performance Testing**
   - Event query performance under load
   - WebSocket subscription scalability
   - Large event volume handling

## Key Testing Patterns

### Event Validation Pattern
```go
// 1. Execute transaction
txResp := suite.createTestTransaction(...)

// 2. Wait for inclusion
finalTx, err := suite.cfg.Client.WaitForTx(ctx, txResp.TxHash, 30*time.Second)

// 3. Filter and validate events
events := client.FilterEventsByType(finalTx.TxResponse.Events, "EventType")
require.NotEmpty(t, events, "should emit EventType")

// 4. Validate attributes
event := events[0]
value, found := client.GetEventAttribute(event, "key")
require.True(t, found, "attribute should be present")
require.Equal(t, expectedValue, value, "attribute value should match")
```

### WebSocket Subscription Pattern
```go
// 1. Connect to WebSocket
wsClient := client.NewWebSocketClient("ws://localhost:26657")
err := wsClient.Connect(ctx)

// 2. Subscribe to events
subscription, err := wsClient.Subscribe(ctx, "custom.query='value'")

// 3. Trigger event (execute transaction)
txResp := suite.executeTransaction(...)

// 4. Wait for event
event, err := subscription.WaitForEvent(ctx, 30*time.Second, filterFunc)
```

### Query Testing Pattern
```go
// 1. Record start height
startHeight, err := suite.cfg.Client.GetLatestBlockHeight(ctx)

// 2. Execute transactions
// ... create multiple transactions

// 3. Query events with filters
events, err := suite.cfg.Client.QueryEventsByType(ctx, "EventType", startHeight, 0)

// 4. Validate results
require.GreaterOrEqual(t, len(events.Events), expectedCount)
```

## Troubleshooting

### Common Issues

1. **WebSocket Connection Failed**
   - Ensure CometBFT is running on port 26657
   - Check WebSocket endpoint configuration
   - Verify network connectivity

2. **Event Not Found**
   - Verify transaction was actually executed
   - Check event type spelling and case sensitivity
   - Confirm transaction succeeded (code = 0)

3. **Query Timeout** 
   - Increase timeout values for slow networks
   - Check block production is active
   - Verify query syntax is correct

4. **Missing Events**
   - Ensure event emission is implemented in keeper
   - Verify protobuf event definitions match
   - Check transaction actually triggered the event

### Debug Commands

```bash
# Check chain status
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info

# Query latest block
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/blocks/latest

# Check WebSocket endpoint
curl -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:26657/websocket

# Query specific transaction
curl http://localhost:1317/cosmos/tx/v1beta1/txs/{TX_HASH}
```

## Future Enhancements

1. **Event Analytics Dashboard** - Real-time event monitoring and analytics
2. **Event Replay Service** - Historical event streaming service  
3. **Event Benchmarking** - Performance testing for high event volumes
4. **Cross-Chain Event Testing** - IBC event emission testing
5. **Event Schema Validation** - Automatic protobuf schema compliance testing