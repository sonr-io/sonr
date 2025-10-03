package utils

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/test/e2e/client"
)

// TestConfig holds common test configuration
type TestConfig struct {
	ChainID        string
	BaseURL        string
	FaucetURL      string
	StakingDenom   string
	NormalDenom    string
	Client         *client.StarshipClient
	FaucetClient   *FaucetClient
	DefaultTimeout time.Duration
	BlockTime      time.Duration
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		ChainID:        "sonrtest_1-1",
		BaseURL:        "http://localhost:1317",
		FaucetURL:      "http://localhost:8000",
		StakingDenom:   "usnr",
		NormalDenom:    "snr",
		DefaultTimeout: 30 * time.Second,
		BlockTime:      2 * time.Second,
		Client:         client.NewStarshipClient("http://localhost:1317"),
		FaucetClient:   NewFaucetClient("http://localhost:8000"),
	}
}

// AssertBalance asserts that an account has the expected balance
func AssertBalance(t *testing.T, cfg *TestConfig, address, denom string, expectedAmount math.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	balance, err := cfg.Client.GetBalance(ctx, address, denom)
	require.NoError(t, err, "failed to query balance")
	require.True(t, balance.Equal(expectedAmount),
		"expected balance %s, got %s", expectedAmount.String(), balance.String())
}

// AssertBalanceGreaterThan asserts that an account balance is greater than expected
func AssertBalanceGreaterThan(t *testing.T, cfg *TestConfig, address, denom string, minAmount math.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	balance, err := cfg.Client.GetBalance(ctx, address, denom)
	require.NoError(t, err, "failed to query balance")
	require.True(t, balance.GT(minAmount),
		"expected balance > %s, got %s", minAmount.String(), balance.String())
}

// AssertBalanceLessThan asserts that an account balance is less than expected
func AssertBalanceLessThan(t *testing.T, cfg *TestConfig, address, denom string, maxAmount math.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	balance, err := cfg.Client.GetBalance(ctx, address, denom)
	require.NoError(t, err, "failed to query balance")
	require.True(t, balance.LT(maxAmount),
		"expected balance < %s, got %s", maxAmount.String(), balance.String())
}

// AssertSupply asserts that a denomination has the expected total supply
func AssertSupply(t *testing.T, cfg *TestConfig, denom string, expectedSupply math.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	supply, err := cfg.Client.GetSupply(ctx, denom)
	require.NoError(t, err, "failed to query supply")
	require.True(t, supply.Equal(expectedSupply),
		"expected supply %s, got %s", expectedSupply.String(), supply.String())
}

// AssertTransferChannelExists asserts that an open transfer channel exists
func AssertTransferChannelExists(t *testing.T, cfg *TestConfig) string {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	channelID, err := cfg.Client.GetTransferChannel(ctx)
	require.NoError(t, err, "failed to find transfer channel")
	require.NotEmpty(t, channelID, "transfer channel ID should not be empty")

	return channelID
}

// WaitForBlocks waits for a specified number of blocks
func WaitForBlocks(ctx context.Context, cfg *TestConfig, blocks int) error {
	waitTime := time.Duration(blocks) * cfg.BlockTime
	select {
	case <-time.After(waitTime):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AssertNodeInfo asserts basic node information
func AssertNodeInfo(t *testing.T, cfg *TestConfig, expectedNetwork string) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	nodeInfo, err := cfg.Client.GetNodeInfo(ctx)
	require.NoError(t, err, "failed to query node info")
	require.Equal(t, expectedNetwork, nodeInfo.DefaultNodeInfo.Network,
		"unexpected network ID")
	require.NotEmpty(t, nodeInfo.ApplicationVersion.Version,
		"application version should not be empty")
}

// SetupTestUsers creates and funds test users
func SetupTestUsers(t *testing.T, cfg *TestConfig, fundAmount math.Int) []CreateTestUser {
	users := GetDefaultTestUsers(fundAmount, cfg.NormalDenom)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DefaultTimeout)
	defer cancel()

	err := cfg.FaucetClient.FundTestUsers(ctx, users)
	require.NoError(t, err, "failed to fund test users")

	// Wait for funding transactions to be included
	err = WaitForBlocks(ctx, cfg, 2)
	require.NoError(t, err, "failed to wait for blocks")

	// Verify funding
	for _, user := range users {
		AssertBalance(t, cfg, user.Address, user.Denom, user.Amount)
	}

	return users
}
