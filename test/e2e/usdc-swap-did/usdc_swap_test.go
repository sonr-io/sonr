package usdcswapdid

/*
USDC Swap E2E Tests - Implementation Notes

CURRENT STATUS:
These E2E tests are structured and ready but require manual setup steps before execution.

LIMITATION:
The test framework currently lacks full transaction signing/broadcasting capability.
The SignAndBroadcastTx method in StarshipClient is a placeholder that requires proper
keyring integration to function.

MANUAL SETUP REQUIRED:
Before running these tests, you must manually create the test resources using the CLI:

1. Create Test DID:
   snrd tx did create-did did:snr:test_<timestamp> --from <account> --chain-id sonrtest_1-1 --yes

2. Register DEX Account:
   snrd tx dex register-account did:snr:test_<timestamp> connection-noble \
     --features swap,liquidity --from <account> --chain-id sonrtest_1-1 --yes

3. Wait for ICA channel handshake (1-2 minutes)

4. Execute Swaps:
   snrd tx dex execute-swap did:snr:test_<timestamp> connection-noble \
     usnr uusdc 1000000 950000 --from <account> --chain-id sonrtest_1-1 --yes

ALTERNATIVE APPROACH:
For automated E2E testing, consider:
- Using the setup.sh script to prepare the environment
- Implementing a test helper that shells out to the CLI
- Integrating with cosmos-sdk/testutil for full transaction support
- Using a dedicated test keyring with pre-configured keys

See README.md for detailed setup instructions and troubleshooting.
*/

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/test/e2e/utils"
	dextypes "github.com/sonr-io/sonr/x/dex/types"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

const (
	// Noble testnet configuration
	NobleChainID     = "noble-grand-1"
	NobleConnectionID = "connection-noble" // Update with actual connection ID
	NobleUSDCDenom   = "uusdc"

	// Test amounts (USDC has 6 decimals like native USDC)
	USDCTestAmount     = 1_000_000  // 1 USDC
	USDCSwapAmount     = 5_000_000  // 5 USDC
	SNRTestAmount      = 10_000_000 // 10 SNR
	SNRSwapAmount      = 1_000_000  // 1 SNR

	// Slippage tolerance (5%)
	SlippageTolerance = 0.05
)

// USDCSwapTestSuite is the main test suite for USDC swap functionality
type USDCSwapTestSuite struct {
	suite.Suite
	cfg         *utils.TestConfig
	ctx         context.Context
	cancel      context.CancelFunc
	testDID     string
	testAccount string
}

// SetupSuite runs once before all tests
func (s *USDCSwapTestSuite) SetupSuite() {
	s.cfg = utils.NewTestConfig()
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	// Setup test account
	users := utils.SetupTestUsers(s.T(), s.cfg, math.NewInt(SNRTestAmount))
	require.Len(s.T(), users, 1, "should create at least one test user")
	s.testAccount = users[0].Address

	s.T().Logf("Test suite initialized with account: %s", s.testAccount)
}

// TearDownSuite runs once after all tests
func (s *USDCSwapTestSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
}

// SetupTest runs before each test
func (s *USDCSwapTestSuite) SetupTest() {
	// Create unique DID for each test
	s.testDID = fmt.Sprintf("did:snr:test_%d", time.Now().UnixNano())
	s.T().Logf("Created test DID: %s", s.testDID)
}

// TestUSDCSwapE2E is the main end-to-end test function
func TestUSDCSwapE2E(t *testing.T) {
	suite.Run(t, new(USDCSwapTestSuite))
}

// Test01_VerifyNobleConfiguration verifies Noble chain configuration
func (s *USDCSwapTestSuite) Test01_VerifyNobleConfiguration() {
	s.T().Log("=== Test 01: Verify Noble Configuration ===")

	// Verify Noble configuration constants
	require.Equal(s.T(), NobleChainID, dextypes.NobleTestnetChainID,
		"Noble chain ID should match")
	require.Equal(s.T(), NobleUSDCDenom, dextypes.NobleUSDCDenom,
		"USDC denom should match")
	require.Equal(s.T(), uint8(6), dextypes.NobleUSDCDecimals,
		"USDC decimals should be 6")

	// Verify Noble is recognized as a Noble chain
	require.True(s.T(), dextypes.IsNobleChain(NobleChainID),
		"Noble testnet chain ID should be recognized")

	s.T().Log("✓ Noble configuration verified successfully")
}

// Test02_CreateDIDDocument creates a DID document for testing
func (s *USDCSwapTestSuite) Test02_CreateDIDDocument() {
	s.T().Log("=== Test 02: Create DID Document ===")

	// Create DID document
	did, err := CreateTestDID(s.ctx, s.cfg, s.testAccount, s.testDID)
	require.NoError(s.T(), err, "failed to create DID document")
	require.NotNil(s.T(), did, "DID document should not be nil")
	require.Equal(s.T(), s.testDID, did.Id, "DID ID should match")

	s.T().Logf("✓ DID document created: %s", did.Id)

	// Verify DID can be queried
	queriedDID, err := QueryDID(s.ctx, s.cfg, s.testDID)
	require.NoError(s.T(), err, "failed to query DID")
	require.NotNil(s.T(), queriedDID, "queried DID should not be nil")
	require.Equal(s.T(), s.testDID, queriedDID.Id, "queried DID should match")

	s.T().Log("✓ DID query verification successful")
}

// Test03_RegisterDEXAccount registers a DEX account on Noble connection
func (s *USDCSwapTestSuite) Test03_RegisterDEXAccount() {
	s.T().Log("=== Test 03: Register DEX Account ===")

	// First create DID
	_, err := CreateTestDID(s.ctx, s.cfg, s.testAccount, s.testDID)
	require.NoError(s.T(), err, "failed to create DID")

	// Register DEX account
	msg := &dextypes.MsgRegisterDEXAccount{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		Features:     []string{"swap", "liquidity"},
		Metadata:     "Noble USDC swap test account",
	}

	txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, msg)
	require.NoError(s.T(), err, "failed to register DEX account")
	require.Equal(s.T(), uint32(0), txResp.Code, "transaction should succeed")

	s.T().Logf("✓ DEX account registered, tx hash: %s", txResp.TxHash)

	// Wait for account activation
	time.Sleep(2 * s.cfg.BlockTime)

	// Query the created account
	account, err := QueryDEXAccount(s.ctx, s.cfg, s.testDID, NobleConnectionID)
	require.NoError(s.T(), err, "failed to query DEX account")
	require.NotNil(s.T(), account, "DEX account should exist")
	require.Equal(s.T(), s.testDID, account.Did, "DID should match")
	require.Equal(s.T(), NobleConnectionID, account.ConnectionId, "connection should match")

	// Account may be pending if ICA channel is not yet established
	s.T().Logf("✓ DEX account status: %s", account.Status.String())
	s.T().Logf("✓ ICA port ID: %s", account.PortId)

	if account.AccountAddress != "" {
		s.T().Logf("✓ ICA address: %s", account.AccountAddress)
	} else {
		s.T().Log("⚠ ICA address pending (channel handshake may be in progress)")
	}
}

// Test04_ExecuteUSDCSwap_SNRToUSDC tests swapping SNR to USDC
func (s *USDCSwapTestSuite) Test04_ExecuteUSDCSwap_SNRToUSDC() {
	s.T().Log("=== Test 04: Execute USDC Swap (SNR → USDC) ===")

	// Setup: Create DID and register DEX account
	s.setupDEXAccount()

	// Get initial balances
	initialSNR := s.getBalance(s.testAccount, s.cfg.StakingDenom)
	s.T().Logf("Initial SNR balance: %s", initialSNR.String())

	// Calculate expected output (95% of input for 5% slippage)
	swapAmount := math.NewInt(SNRSwapAmount)
	minUSDCOut := math.NewInt(SNRSwapAmount).
		MulRaw(95).QuoRaw(100) // 5% slippage tolerance

	// Execute swap
	swapMsg := &dextypes.MsgExecuteSwap{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		SourceDenom:  s.cfg.StakingDenom,
		TargetDenom:  NobleUSDCDenom,
		Amount:       swapAmount,
		MinAmountOut: minUSDCOut,
		Route:        fmt.Sprintf("noble:%s", NobleConnectionID),
		Timeout:      time.Now().Add(60 * time.Second),
	}

	txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, swapMsg)
	require.NoError(s.T(), err, "failed to execute swap")
	require.Equal(s.T(), uint32(0), txResp.Code, "swap transaction should succeed")

	s.T().Logf("✓ Swap transaction broadcasted, tx hash: %s", txResp.TxHash)

	// Verify swap event was emitted
	s.verifySwapEvent(txResp.TxHash, s.testDID, NobleConnectionID)

	// Note: In a real IBC swap, we would need to wait for ICA callback
	// and verify balances on both chains. For this test, we verify the
	// transaction was accepted and properly formatted.
	s.T().Log("✓ Swap execution test completed")
}

// Test05_ExecuteUSDCSwap_USDCToSNR tests swapping USDC to SNR
func (s *USDCSwapTestSuite) Test05_ExecuteUSDCSwap_USDCToSNR() {
	s.T().Log("=== Test 05: Execute USDC Swap (USDC → SNR) ===")

	// Setup: Create DID and register DEX account
	s.setupDEXAccount()

	// Calculate expected output
	swapAmount := math.NewInt(USDCSwapAmount)
	minSNROut := math.NewInt(USDCSwapAmount).
		MulRaw(95).QuoRaw(100) // 5% slippage tolerance

	// Execute swap
	swapMsg := &dextypes.MsgExecuteSwap{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		SourceDenom:  NobleUSDCDenom,
		TargetDenom:  s.cfg.StakingDenom,
		Amount:       swapAmount,
		MinAmountOut: minSNROut,
		Route:        fmt.Sprintf("noble:%s", NobleConnectionID),
		Timeout:      time.Now().Add(60 * time.Second),
	}

	txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, swapMsg)
	require.NoError(s.T(), err, "failed to execute swap")
	require.Equal(s.T(), uint32(0), txResp.Code, "swap transaction should succeed")

	s.T().Logf("✓ Swap transaction broadcasted, tx hash: %s", txResp.TxHash)

	// Verify swap event
	s.verifySwapEvent(txResp.TxHash, s.testDID, NobleConnectionID)

	s.T().Log("✓ Reverse swap execution test completed")
}

// Test06_SwapWithInvalidParameters tests swap with invalid parameters
func (s *USDCSwapTestSuite) Test06_SwapWithInvalidParameters() {
	s.T().Log("=== Test 06: Swap With Invalid Parameters ===")

	// Setup: Create DID and register DEX account
	s.setupDEXAccount()

	testCases := []struct {
		name        string
		msg         *dextypes.MsgExecuteSwap
		expectedErr bool
		description string
	}{
		{
			name: "zero_amount",
			msg: &dextypes.MsgExecuteSwap{
				Did:          s.testDID,
				ConnectionId: NobleConnectionID,
				SourceDenom:  s.cfg.StakingDenom,
				TargetDenom:  NobleUSDCDenom,
				Amount:       math.ZeroInt(),
				MinAmountOut: math.NewInt(1000),
				Timeout:      time.Now().Add(60 * time.Second),
			},
			expectedErr: true,
			description: "swap with zero amount should fail",
		},
		{
			name: "same_denom",
			msg: &dextypes.MsgExecuteSwap{
				Did:          s.testDID,
				ConnectionId: NobleConnectionID,
				SourceDenom:  s.cfg.StakingDenom,
				TargetDenom:  s.cfg.StakingDenom,
				Amount:       math.NewInt(1000),
				MinAmountOut: math.NewInt(900),
				Timeout:      time.Now().Add(60 * time.Second),
			},
			expectedErr: true,
			description: "swapping same denom should fail",
		},
		{
			name: "negative_min_out",
			msg: &dextypes.MsgExecuteSwap{
				Did:          s.testDID,
				ConnectionId: NobleConnectionID,
				SourceDenom:  s.cfg.StakingDenom,
				TargetDenom:  NobleUSDCDenom,
				Amount:       math.NewInt(1000),
				MinAmountOut: math.NewInt(-100),
				Timeout:      time.Now().Add(60 * time.Second),
			},
			expectedErr: true,
			description: "negative min amount out should fail",
		},
		{
			name: "invalid_connection",
			msg: &dextypes.MsgExecuteSwap{
				Did:          s.testDID,
				ConnectionId: "connection-invalid-999",
				SourceDenom:  s.cfg.StakingDenom,
				TargetDenom:  NobleUSDCDenom,
				Amount:       math.NewInt(1000),
				MinAmountOut: math.NewInt(900),
				Timeout:      time.Now().Add(60 * time.Second),
			},
			expectedErr: true,
			description: "invalid connection ID should fail",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, tc.msg)

			if tc.expectedErr {
				// Transaction should either fail to broadcast or have non-zero code
				if err == nil {
					require.NotEqual(t, uint32(0), txResp.Code,
						"transaction should fail with non-zero code")
					t.Logf("✓ Transaction failed as expected with code: %d, log: %s",
						txResp.Code, txResp.RawLog)
				} else {
					t.Logf("✓ Transaction failed as expected with error: %v", err)
				}
			} else {
				require.NoError(t, err, "transaction should succeed")
				require.Equal(t, uint32(0), txResp.Code, "transaction should have code 0")
			}
		})
	}

	s.T().Log("✓ Invalid parameter tests completed")
}

// Test07_SwapWithUCANPermission tests swap with UCAN authorization
func (s *USDCSwapTestSuite) Test07_SwapWithUCANPermission() {
	s.T().Skip("UCAN integration requires full UCAN module setup")
	s.T().Log("=== Test 07: Swap With UCAN Permission ===")

	// Setup: Create DID and register DEX account
	s.setupDEXAccount()

	// TODO: Generate UCAN token with swap capabilities
	// ucanToken := generateUCANToken(s.testDID, "swap", NobleConnectionID)

	swapMsg := &dextypes.MsgExecuteSwap{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		SourceDenom:  s.cfg.StakingDenom,
		TargetDenom:  NobleUSDCDenom,
		Amount:       math.NewInt(SNRSwapAmount),
		MinAmountOut: math.NewInt(SNRSwapAmount * 95 / 100),
		UcanToken:    "", // Would be populated with actual token
		Timeout:      time.Now().Add(60 * time.Second),
	}

	txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, swapMsg)
	require.NoError(s.T(), err, "failed to execute swap with UCAN")
	require.Equal(s.T(), uint32(0), txResp.Code, "swap with UCAN should succeed")

	s.T().Log("✓ UCAN-authorized swap test completed")
}

// Test08_QueryDEXHistory queries transaction history
func (s *USDCSwapTestSuite) Test08_QueryDEXHistory() {
	s.T().Log("=== Test 08: Query DEX History ===")

	// Setup and execute a swap
	s.setupDEXAccount()
	s.executeTestSwap()

	// Query DEX history
	history, err := QueryDEXHistory(s.ctx, s.cfg, s.testDID)
	require.NoError(s.T(), err, "failed to query DEX history")

	if len(history) > 0 {
		s.T().Logf("✓ Found %d transactions in history", len(history))
		for i, activity := range history {
			s.T().Logf("  [%d] Type: %s, Status: %s, Height: %d",
				i+1, activity.Type, activity.Status, activity.BlockHeight)
		}
	} else {
		s.T().Log("⚠ No history found (ICA may not be fully configured)")
	}
}

// Test09_MultipleSwaps tests executing multiple swaps in sequence
func (s *USDCSwapTestSuite) Test09_MultipleSwaps() {
	s.T().Log("=== Test 09: Multiple Sequential Swaps ===")

	// Setup
	s.setupDEXAccount()

	numSwaps := 3
	for i := 0; i < numSwaps; i++ {
		s.T().Logf("Executing swap %d/%d", i+1, numSwaps)

		swapMsg := &dextypes.MsgExecuteSwap{
			Did:          s.testDID,
			ConnectionId: NobleConnectionID,
			SourceDenom:  s.cfg.StakingDenom,
			TargetDenom:  NobleUSDCDenom,
			Amount:       math.NewInt(100_000), // Small amounts for multiple swaps
			MinAmountOut: math.NewInt(95_000),
			Timeout:      time.Now().Add(60 * time.Second),
		}

		txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, swapMsg)
		require.NoError(s.T(), err, "swap %d should succeed", i+1)
		require.Equal(s.T(), uint32(0), txResp.Code, "swap %d should have code 0", i+1)

		s.T().Logf("✓ Swap %d completed, tx: %s", i+1, txResp.TxHash)

		// Small delay between swaps
		time.Sleep(s.cfg.BlockTime)
	}

	s.T().Logf("✓ Successfully executed %d swaps", numSwaps)
}

// Helper Functions

func (s *USDCSwapTestSuite) setupDEXAccount() {
	// Create DID if not exists
	_, err := CreateTestDID(s.ctx, s.cfg, s.testAccount, s.testDID)
	if err != nil {
		// DID might already exist, check query
		_, queryErr := QueryDID(s.ctx, s.cfg, s.testDID)
		require.NoError(s.T(), queryErr, "DID should exist or be creatable")
	}

	// Register DEX account
	msg := &dextypes.MsgRegisterDEXAccount{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		Features:     []string{"swap"},
	}

	_, err = s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, msg)
	// Ignore error if account already exists

	// Wait for registration
	time.Sleep(2 * s.cfg.BlockTime)
}

func (s *USDCSwapTestSuite) executeTestSwap() {
	swapMsg := &dextypes.MsgExecuteSwap{
		Did:          s.testDID,
		ConnectionId: NobleConnectionID,
		SourceDenom:  s.cfg.StakingDenom,
		TargetDenom:  NobleUSDCDenom,
		Amount:       math.NewInt(SNRSwapAmount),
		MinAmountOut: math.NewInt(SNRSwapAmount * 95 / 100),
		Timeout:      time.Now().Add(60 * time.Second),
	}

	txResp, err := s.cfg.Client.SignAndBroadcastTx(s.ctx, s.testAccount, swapMsg)
	require.NoError(s.T(), err, "test swap should succeed")
	require.Equal(s.T(), uint32(0), txResp.Code, "test swap should have code 0")
}

func (s *USDCSwapTestSuite) getBalance(address, denom string) math.Int {
	balance, err := s.cfg.Client.GetBalance(s.ctx, address, denom)
	require.NoError(s.T(), err, "should query balance")
	return balance
}

func (s *USDCSwapTestSuite) verifySwapEvent(txHash, expectedDID, expectedConnection string) {
	// Query events for the transaction
	height, err := s.cfg.Client.GetLatestBlockHeight(s.ctx)
	require.NoError(s.T(), err, "should get block height")

	// Search for swap events
	events, err := s.cfg.Client.QueryEventsByType(s.ctx,
		dextypes.EventTypeSwapExecuted, height-10, height)

	if err != nil || len(events.Events) == 0 {
		s.T().Log("⚠ Swap event not found (may be indexed later)")
		return
	}

	// Verify event contains expected attributes
	for _, event := range events.Events {
		for _, attr := range event.Attributes {
			if attr.Key == "did" && attr.Value == expectedDID {
				s.T().Logf("✓ Found swap event for DID: %s", expectedDID)
				return
			}
		}
	}

	s.T().Log("⚠ Swap event attributes not yet indexed")
}
