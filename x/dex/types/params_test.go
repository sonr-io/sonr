package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Verify basic fields
	require.True(t, params.Enabled)
	require.Equal(t, uint32(5), params.MaxAccountsPerDid)
	require.Equal(t, uint64(600), params.DefaultTimeoutSeconds)

	// Verify Noble testnet is in allowed connections
	require.Contains(t, params.AllowedConnections, "noble-grand-1")

	// Verify amounts
	require.Equal(t, "1000", params.MinSwapAmount)
	require.Equal(t, "1000000000000", params.MaxDailyVolume)

	// Verify rate limits
	require.Equal(t, uint32(10), params.RateLimits.MaxOpsPerBlock)
	require.Equal(t, uint32(100), params.RateLimits.MaxOpsPerDidPerDay)
	require.Equal(t, uint32(5), params.RateLimits.CooldownBlocks)

	// Verify fees
	require.Equal(t, uint32(30), params.Fees.SwapFeeBps)
	require.Equal(t, uint32(20), params.Fees.LiquidityFeeBps)
	require.Equal(t, uint32(10), params.Fees.OrderFeeBps)

	// Validate params
	err := params.Validate()
	require.NoError(t, err)
}

func TestParams_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		params    Params
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default params",
			params:    DefaultParams(),
			expectErr: false,
		},
		{
			name: "invalid - zero max accounts",
			params: Params{
				MaxAccountsPerDid:     0,
				DefaultTimeoutSeconds: 600,
				MinSwapAmount:         "1000",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "max_accounts_per_did must be positive",
		},
		{
			name: "invalid - max accounts too high",
			params: Params{
				MaxAccountsPerDid:     101,
				DefaultTimeoutSeconds: 600,
				MinSwapAmount:         "1000",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "max_accounts_per_did cannot exceed 100",
		},
		{
			name: "invalid - zero timeout",
			params: Params{
				MaxAccountsPerDid:     5,
				DefaultTimeoutSeconds: 0,
				MinSwapAmount:         "1000",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "default_timeout_seconds must be positive",
		},
		{
			name: "invalid - timeout too high",
			params: Params{
				MaxAccountsPerDid:     5,
				DefaultTimeoutSeconds: 3601,
				MinSwapAmount:         "1000",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "default_timeout_seconds cannot exceed 3600",
		},
		{
			name: "invalid - negative min swap amount",
			params: Params{
				MaxAccountsPerDid:     5,
				DefaultTimeoutSeconds: 600,
				MinSwapAmount:         "-1000",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "min_swap_amount cannot be negative",
		},
		{
			name: "invalid - invalid min swap amount format",
			params: Params{
				MaxAccountsPerDid:     5,
				DefaultTimeoutSeconds: 600,
				MinSwapAmount:         "invalid",
				MaxDailyVolume:        "1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "invalid min_swap_amount",
		},
		{
			name: "invalid - negative max daily volume",
			params: Params{
				MaxAccountsPerDid:     5,
				DefaultTimeoutSeconds: 600,
				MinSwapAmount:         "1000",
				MaxDailyVolume:        "-1000000",
				RateLimits:            DefaultParams().RateLimits,
				Fees:                  DefaultParams().Fees,
			},
			expectErr: true,
			errMsg:    "max_daily_volume cannot be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRateLimitParams_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		params    RateLimitParams
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default",
			params:    DefaultParams().RateLimits,
			expectErr: false,
		},
		{
			name: "invalid - zero max ops per block",
			params: RateLimitParams{
				MaxOpsPerBlock:     0,
				MaxOpsPerDidPerDay: 100,
				CooldownBlocks:     5,
			},
			expectErr: true,
			errMsg:    "max_ops_per_block must be positive",
		},
		{
			name: "invalid - max ops per block too high",
			params: RateLimitParams{
				MaxOpsPerBlock:     1001,
				MaxOpsPerDidPerDay: 100,
				CooldownBlocks:     5,
			},
			expectErr: true,
			errMsg:    "max_ops_per_block cannot exceed 1000",
		},
		{
			name: "invalid - zero max ops per did per day",
			params: RateLimitParams{
				MaxOpsPerBlock:     10,
				MaxOpsPerDidPerDay: 0,
				CooldownBlocks:     5,
			},
			expectErr: true,
			errMsg:    "max_ops_per_did_per_day must be positive",
		},
		{
			name: "valid - zero cooldown blocks",
			params: RateLimitParams{
				MaxOpsPerBlock:     10,
				MaxOpsPerDidPerDay: 100,
				CooldownBlocks:     0,
			},
			expectErr: false,
		},
		{
			name: "invalid - cooldown blocks too high",
			params: RateLimitParams{
				MaxOpsPerBlock:     10,
				MaxOpsPerDidPerDay: 100,
				CooldownBlocks:     1001,
			},
			expectErr: true,
			errMsg:    "cooldown_blocks cannot exceed 1000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFeeParams_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		params    FeeParams
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid default",
			params:    DefaultParams().Fees,
			expectErr: false,
		},
		{
			name: "valid - zero fees",
			params: FeeParams{
				SwapFeeBps:      0,
				LiquidityFeeBps: 0,
				OrderFeeBps:     0,
				FeeCollector:    "",
			},
			expectErr: false,
		},
		{
			name: "valid - max fees",
			params: FeeParams{
				SwapFeeBps:      10000,
				LiquidityFeeBps: 10000,
				OrderFeeBps:     10000,
				FeeCollector:    "",
			},
			expectErr: false,
		},
		{
			name: "invalid - swap fee too high",
			params: FeeParams{
				SwapFeeBps:      10001,
				LiquidityFeeBps: 20,
				OrderFeeBps:     10,
				FeeCollector:    "",
			},
			expectErr: true,
			errMsg:    "swap_fee_bps cannot exceed",
		},
		{
			name: "invalid - liquidity fee too high",
			params: FeeParams{
				SwapFeeBps:      30,
				LiquidityFeeBps: 10001,
				OrderFeeBps:     10,
				FeeCollector:    "",
			},
			expectErr: true,
			errMsg:    "liquidity_fee_bps cannot exceed",
		},
		{
			name: "invalid - order fee too high",
			params: FeeParams{
				SwapFeeBps:      30,
				LiquidityFeeBps: 20,
				OrderFeeBps:     10001,
				FeeCollector:    "",
			},
			expectErr: true,
			errMsg:    "order_fee_bps cannot exceed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
