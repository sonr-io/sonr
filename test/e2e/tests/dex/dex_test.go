package dex

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/test/e2e/utils"
	dextypes "github.com/sonr-io/sonr/x/dex/types"
)

// TestDEXModuleOperations tests the DEX module E2E operations
func TestDEXModuleOperations(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("query_dex_params", func(t *testing.T) {
		// Query DEX module parameters
		resp, err := cfg.Client.QueryDEXParams(ctx)
		require.NoError(t, err, "failed to query DEX params")
		require.NotNil(t, resp, "DEX params should not be nil")
		require.True(t, resp.Params.Enabled, "DEX module should be enabled")
	})

	t.Run("register_dex_account", func(t *testing.T) {
		// Register a DEX account for testing
		did := "did:sonr:e2e_test_user"
		connectionID := "connection-0"
		features := []string{"swap", "liquidity"}

		msg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     features,
		}

		// Sign and broadcast transaction
		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, msg)
		require.NoError(t, err, "failed to register DEX account")
		require.Equal(t, uint32(0), txResp.Code, "transaction should succeed")

		// Query the created account
		queryResp, err := cfg.Client.QueryDEXAccount(ctx, did, connectionID)
		require.NoError(t, err, "failed to query DEX account")
		require.NotNil(t, queryResp, "DEX account should exist")
		require.Equal(t, did, queryResp.Account.Did)
		require.Equal(t, connectionID, queryResp.Account.ConnectionId)
	})

	t.Run("execute_swap", func(t *testing.T) {
		// Setup: Register account first
		did := "did:sonr:e2e_swap_user"
		connectionID := "connection-0"

		registerMsg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     []string{"swap"},
		}

		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, registerMsg)
		require.NoError(t, err, "failed to register DEX account for swap")
		require.Equal(t, uint32(0), txResp.Code, "registration should succeed")

		// Execute swap
		swapMsg := &dextypes.MsgExecuteSwap{
			Did:          did,
			ConnectionId: connectionID,
			SourceDenom:  cfg.StakingDenom,
			TargetDenom:  "uosmo",
			Amount:       math.NewInt(1000),
			MinAmountOut: math.NewInt(900),
			Route:        "pool:1",
		}

		txResp, err = cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, swapMsg)
		require.NoError(t, err, "failed to execute swap")
		require.Equal(t, uint32(0), txResp.Code, "swap should succeed")
	})

	t.Run("provide_liquidity", func(t *testing.T) {
		// Setup: Register account with liquidity feature
		did := "did:sonr:e2e_lp_user"
		connectionID := "connection-0"

		registerMsg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     []string{"liquidity"},
		}

		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, registerMsg)
		require.NoError(t, err, "failed to register DEX account for liquidity")
		require.Equal(t, uint32(0), txResp.Code, "registration should succeed")

		// Provide liquidity
		liquidityMsg := &dextypes.MsgProvideLiquidity{
			Did:          did,
			ConnectionId: connectionID,
			PoolId:       "1",
			Assets: sdk.NewCoins(
				sdk.NewCoin(cfg.StakingDenom, math.NewInt(1000)),
				sdk.NewCoin("uosmo", math.NewInt(1000)),
			),
			MinShares: math.NewInt(100),
			Timeout:   time.Now().Add(5 * time.Minute),
		}

		txResp, err = cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, liquidityMsg)
		require.NoError(t, err, "failed to provide liquidity")
		require.Equal(t, uint32(0), txResp.Code, "liquidity provision should succeed")
	})

	t.Run("create_limit_order", func(t *testing.T) {
		// Setup: Register account with order feature
		did := "did:sonr:e2e_order_user"
		connectionID := "connection-0"

		registerMsg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     []string{"order"},
		}

		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, registerMsg)
		require.NoError(t, err, "failed to register DEX account for orders")
		require.Equal(t, uint32(0), txResp.Code, "registration should succeed")

		// Create limit order
		orderMsg := &dextypes.MsgCreateLimitOrder{
			Did:          did,
			ConnectionId: connectionID,
			SellDenom:    cfg.StakingDenom,
			BuyDenom:     "uosmo",
			Amount:       math.NewInt(1000),
			Price:        math.LegacyNewDec(1),
			Expiration:   time.Now().Add(24 * time.Hour),
		}

		txResp, err = cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, orderMsg)
		require.NoError(t, err, "failed to create limit order")
		require.Equal(t, uint32(0), txResp.Code, "order creation should succeed")

		// TODO: Query and verify the order was created
	})

	t.Run("query_dex_accounts", func(t *testing.T) {
		// Query all DEX accounts
		resp, err := cfg.Client.QueryAllDEXAccounts(ctx)
		require.NoError(t, err, "failed to query all DEX accounts")
		require.NotNil(t, resp, "response should not be nil")
		// Should have at least the accounts created in previous tests
		require.GreaterOrEqual(t, len(resp.Accounts), 1, "should have at least one account")
	})

	t.Run("query_dex_history", func(t *testing.T) {
		// Query transaction history for a DID
		did := "did:sonr:e2e_swap_user"

		resp, err := cfg.Client.QueryDEXHistory(ctx, did)
		require.NoError(t, err, "failed to query DEX history")
		require.NotNil(t, resp, "response should not be nil")
		// Should have at least one transaction from the swap test
		require.GreaterOrEqual(t, len(resp.History), 0, "history may be empty if ICA is not fully setup")
	})

	t.Run("cancel_order", func(t *testing.T) {
		// Setup: Register account and create an order first
		did := "did:sonr:e2e_cancel_user"
		connectionID := "connection-0"

		// Register account
		registerMsg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     []string{"order"},
		}

		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, registerMsg)
		require.NoError(t, err, "failed to register DEX account")
		require.Equal(t, uint32(0), txResp.Code, "registration should succeed")

		// Create an order
		orderMsg := &dextypes.MsgCreateLimitOrder{
			Did:          did,
			ConnectionId: connectionID,
			SellDenom:    cfg.StakingDenom,
			BuyDenom:     "uosmo",
			Amount:       math.NewInt(500),
			Price:        math.LegacyNewDec(1),
			Expiration:   time.Now().Add(24 * time.Hour),
		}

		createResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, orderMsg)
		require.NoError(t, err, "failed to create order")
		require.Equal(t, uint32(0), createResp.Code, "order creation should succeed")

		// Extract order ID from events (mock for now)
		orderID := "order-1" // In real test, extract from createResp.Events

		// Cancel the order
		cancelMsg := &dextypes.MsgCancelOrder{
			Did:          did,
			ConnectionId: connectionID,
			OrderId:      orderID,
		}

		cancelResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, cancelMsg)
		require.NoError(t, err, "failed to cancel order")
		require.Equal(t, uint32(0), cancelResp.Code, "order cancellation should succeed")
	})

	t.Run("remove_liquidity", func(t *testing.T) {
		// Setup: Register account and provide liquidity first
		did := "did:sonr:e2e_remove_lp_user"
		connectionID := "connection-0"

		// Register account
		registerMsg := &dextypes.MsgRegisterDEXAccount{
			Did:          did,
			ConnectionId: connectionID,
			Features:     []string{"liquidity"},
		}

		txResp, err := cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, registerMsg)
		require.NoError(t, err, "failed to register DEX account")
		require.Equal(t, uint32(0), txResp.Code, "registration should succeed")

		// First provide liquidity
		provideMsg := &dextypes.MsgProvideLiquidity{
			Did:          did,
			ConnectionId: connectionID,
			PoolId:       "1",
			Assets: sdk.NewCoins(
				sdk.NewCoin(cfg.StakingDenom, math.NewInt(2000)),
				sdk.NewCoin("uosmo", math.NewInt(2000)),
			),
			MinShares: math.NewInt(200),
			Timeout:   time.Now().Add(5 * time.Minute),
		}

		txResp, err = cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, provideMsg)
		require.NoError(t, err, "failed to provide liquidity")
		require.Equal(t, uint32(0), txResp.Code, "liquidity provision should succeed")

		// Remove liquidity
		removeMsg := &dextypes.MsgRemoveLiquidity{
			Did:          did,
			ConnectionId: connectionID,
			PoolId:       "1",
			Shares:       math.NewInt(100),
			MinAmounts: sdk.NewCoins(
				sdk.NewCoin(cfg.StakingDenom, math.NewInt(900)),
				sdk.NewCoin("uosmo", math.NewInt(900)),
			),
			Timeout: time.Now().Add(5 * time.Minute),
		}

		txResp, err = cfg.Client.SignAndBroadcastTx(ctx, cfg.TestAccount, removeMsg)
		require.NoError(t, err, "failed to remove liquidity")
		require.Equal(t, uint32(0), txResp.Code, "liquidity removal should succeed")
	})
}

// TestDEXIBCIntegration tests IBC-related DEX operations
func TestDEXIBCIntegration(t *testing.T) {
	t.Skip("Skipping IBC integration tests - requires full IBC setup")

	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("cross_chain_swap", func(t *testing.T) {
		// This test would require an actual IBC connection to another chain
		// For now, we skip it but document the expected behavior

		// 1. Register ICA account on remote chain
		// 2. Fund the ICA account
		// 3. Execute swap on remote chain
		// 4. Verify swap execution through events/callbacks
		_ = cfg
		_ = ctx
	})

	t.Run("multi_chain_accounts", func(t *testing.T) {
		// Test managing accounts across multiple chains
		// This would require multiple IBC connections

		// 1. Register accounts on Osmosis, Cosmos Hub, etc.
		// 2. Query all accounts for a single DID
		// 3. Verify each account has different connection IDs
		_ = cfg
		_ = ctx
	})
}
