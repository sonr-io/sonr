package basic

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/test/e2e/utils"
)

func TestBasicChain(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("node_info", func(t *testing.T) {
		utils.AssertNodeInfo(t, cfg, cfg.ChainID)
	})

	t.Run("validate_pre_funded_accounts", func(t *testing.T) {
		// Check pre-funded accounts from localnet
		acc0Addr := "idx1fcqk3crpnyvyhtd4jepsnx5eat5ehc920epq29"
		acc1Addr := "idx10n78mn09nx0f056wam35wkfvanf37kepuj28x4"

		// Verify acc0 has balance
		balance0, err := cfg.Client.GetBalance(ctx, acc0Addr, cfg.StakingDenom)
		require.NoError(t, err, "failed to query acc0 balance")
		require.True(t, balance0.GT(math.ZeroInt()), "acc0 should have balance")

		// Verify acc1 has balance
		balance1, err := cfg.Client.GetBalance(ctx, acc1Addr, cfg.StakingDenom)
		require.NoError(t, err, "failed to query acc1 balance")
		require.True(t, balance1.GT(math.ZeroInt()), "acc1 should have balance")
	})

	t.Run("bank_params", func(t *testing.T) {
		bankParams, err := cfg.Client.GetBankParams(ctx)
		require.NoError(t, err, "failed to query bank params")
		require.NotNil(t, bankParams, "bank params should not be nil")
	})

	t.Run("supply_queries", func(t *testing.T) {
		// Query total supply of test denom
		testSupply, err := cfg.Client.GetSupply(ctx, "test")
		require.NoError(t, err, "failed to query test supply")
		require.True(t, testSupply.GT(math.ZeroInt()), "test supply should be greater than zero")

		// Query total supply of staking denom
		stakingSupply, err := cfg.Client.GetSupply(ctx, cfg.StakingDenom)
		require.NoError(t, err, "failed to query staking supply")
		require.True(t, stakingSupply.GT(math.ZeroInt()), "staking supply should be greater than zero")
	})

	t.Run("balance_operations", func(t *testing.T) {
		// Use pre-funded account from localnet
		testAddr := "idx1fcqk3crpnyvyhtd4jepsnx5eat5ehc920epq29"

		// Get all balances
		balances, err := cfg.Client.GetAllBalances(ctx, testAddr)
		require.NoError(t, err, "failed to query all balances")
		require.NotEmpty(t, balances, "user should have at least one balance")

		// Check specific balance for staking denom
		balance, err := cfg.Client.GetBalance(ctx, testAddr, cfg.StakingDenom)
		require.NoError(t, err, "failed to query specific balance")
		require.True(t, balance.GT(math.ZeroInt()), "balance should be greater than zero")
	})
}

func TestFaucetOperations(t *testing.T) {
	t.Skip("Skipping faucet tests - localnet doesn't have a faucet")
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	tests := []struct {
		name        string
		fundAmount  math.Int
		expectError bool
	}{
		{
			name:        "normal_funding",
			fundAmount:  math.NewInt(1_000_000),
			expectError: false,
		},
		{
			name:        "large_funding",
			fundAmount:  math.NewInt(100_000_000),
			expectError: false,
		},
		{
			name:        "zero_funding",
			fundAmount:  math.ZeroInt(),
			expectError: false, // Faucet should handle this gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := utils.GetDefaultTestUsers(tt.fundAmount, cfg.NormalDenom)
			testUser := users[0]

			err := cfg.FaucetClient.FundTestUsers(ctx, []utils.CreateTestUser{testUser})
			if tt.expectError {
				require.Error(t, err, "expected funding to fail")
			} else {
				require.NoError(t, err, "funding should succeed")

				// Wait for transaction to be included
				err = utils.WaitForBlocks(ctx, cfg, 2)
				require.NoError(t, err, "failed to wait for blocks")

				// Verify balance if funding was expected to succeed
				if !tt.fundAmount.IsZero() {
					utils.AssertBalance(t, cfg, testUser.Address, testUser.Denom, testUser.Amount)
				}
			}
		})
	}
}

func TestChainConnectivity(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("rest_api_connectivity", func(t *testing.T) {
		// Test REST API connectivity by querying node info
		nodeInfo, err := cfg.Client.GetNodeInfo(ctx)
		require.NoError(t, err, "REST API should be accessible")
		require.Equal(t, cfg.ChainID, nodeInfo.DefaultNodeInfo.Network, "chain ID should match")
	})

	t.Run("faucet_connectivity", func(t *testing.T) {
		t.Skip("Skipping faucet connectivity test - localnet doesn't have a faucet")
	})
}
