package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGetNobleTestnetConfig(t *testing.T) {
	config := GetNobleTestnetConfig()

	require.Equal(t, NobleTestnetChainID, config.ChainID)
	require.Equal(t, "https://noble-testnet-rpc.polkachu.com:443", config.RPCEndpoint)
	require.Equal(t, "noble-testnet-grpc.polkachu.com:21590", config.GRPCEndpoint)
	require.Equal(t, NobleUSDCDenom, config.USDCDenom)
	require.Equal(t, uint8(6), config.Decimals)
}

func TestGetNobleMainnetConfig(t *testing.T) {
	config := GetNobleMainnetConfig()

	require.Equal(t, NobleMainnetChainID, config.ChainID)
	require.Equal(t, "https://noble-rpc.polkachu.com:443", config.RPCEndpoint)
	require.Equal(t, "noble-grpc.polkachu.com:21590", config.GRPCEndpoint)
	require.Equal(t, NobleUSDCDenom, config.USDCDenom)
	require.Equal(t, uint8(6), config.Decimals)
}

func TestIsNobleChain(t *testing.T) {
	testCases := []struct {
		name     string
		chainID  string
		expected bool
	}{
		{
			name:     "noble testnet",
			chainID:  NobleTestnetChainID,
			expected: true,
		},
		{
			name:     "noble mainnet",
			chainID:  NobleMainnetChainID,
			expected: true,
		},
		{
			name:     "osmosis",
			chainID:  "osmosis-1",
			expected: false,
		},
		{
			name:     "cosmos hub",
			chainID:  "cosmoshub-4",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNobleChain(tc.chainID)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateNobleConnection(t *testing.T) {
	allowedConnections := []string{"noble-grand-1", "osmo-test-5"}

	testCases := []struct {
		name         string
		connectionID string
		expectErr    bool
	}{
		{
			name:         "valid noble connection",
			connectionID: "noble-grand-1",
			expectErr:    false,
		},
		{
			name:         "valid osmosis connection",
			connectionID: "osmo-test-5",
			expectErr:    false,
		},
		{
			name:         "invalid connection",
			connectionID: "unknown-chain",
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateNobleConnection(tc.connectionID, allowedConnections)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConvertToUSDC(t *testing.T) {
	testCases := []struct {
		name     string
		amount   sdk.Int
		expected string
	}{
		{
			name:     "1 USDC",
			amount:   sdk.NewInt(1000000),
			expected: "1.000000000000000000",
		},
		{
			name:     "1.5 USDC",
			amount:   sdk.NewInt(1500000),
			expected: "1.500000000000000000",
		},
		{
			name:     "0.1 USDC",
			amount:   sdk.NewInt(100000),
			expected: "0.100000000000000000",
		},
		{
			name:     "0 USDC",
			amount:   sdk.ZeroInt(),
			expected: "0.000000000000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertToUSDC(tc.amount)
			require.Equal(t, tc.expected, result.String())
		})
	}
}

func TestConvertFromUSDC(t *testing.T) {
	testCases := []struct {
		name     string
		usdc     sdk.Dec
		expected int64
	}{
		{
			name:     "1 USDC",
			usdc:     sdk.MustNewDecFromStr("1.0"),
			expected: 1000000,
		},
		{
			name:     "1.5 USDC",
			usdc:     sdk.MustNewDecFromStr("1.5"),
			expected: 1500000,
		},
		{
			name:     "0.1 USDC",
			usdc:     sdk.MustNewDecFromStr("0.1"),
			expected: 100000,
		},
		{
			name:     "0 USDC",
			usdc:     sdk.ZeroDec(),
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertFromUSDC(tc.usdc)
			require.Equal(t, tc.expected, result.Int64())
		})
	}
}

func TestFormatUSDCAmount(t *testing.T) {
	testCases := []struct {
		name     string
		amount   sdk.Int
		expected string
	}{
		{
			name:     "1 USDC",
			amount:   sdk.NewInt(1000000),
			expected: "1.000000000000000000 USDC",
		},
		{
			name:     "1.5 USDC",
			amount:   sdk.NewInt(1500000),
			expected: "1.500000000000000000 USDC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatUSDCAmount(tc.amount)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestParseUSDCAmount(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  int64
		expectErr bool
	}{
		{
			name:      "1 USDC",
			input:     "1.0",
			expected:  1000000,
			expectErr: false,
		},
		{
			name:      "1.5 USDC",
			input:     "1.5",
			expected:  1500000,
			expectErr: false,
		},
		{
			name:      "0 USDC",
			input:     "0",
			expected:  0,
			expectErr: false,
		},
		{
			name:      "invalid format",
			input:     "abc",
			expectErr: true,
		},
		{
			name:      "negative amount",
			input:     "-1.0",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseUSDCAmount(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result.Int64())
			}
		})
	}
}

func TestNobleSwapParams_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		params    NobleSwapParams
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid params",
			params: NobleSwapParams{
				InputDenom:  "uatom",
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.NewInt(500000),
				Receiver:    "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: false,
		},
		{
			name: "empty input denom",
			params: NobleSwapParams{
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.NewInt(500000),
				Receiver:    "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: true,
			errMsg:    "input denom cannot be empty",
		},
		{
			name: "empty output denom",
			params: NobleSwapParams{
				InputDenom: "uatom",
				Amount:     sdk.NewInt(1000000),
				MinOutput:  sdk.NewInt(500000),
				Receiver:   "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: true,
			errMsg:    "output denom cannot be empty",
		},
		{
			name: "same input and output denom",
			params: NobleSwapParams{
				InputDenom:  "uusdc",
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.NewInt(500000),
				Receiver:    "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: true,
			errMsg:    "input and output denoms must be different",
		},
		{
			name: "zero amount",
			params: NobleSwapParams{
				InputDenom:  "uatom",
				OutputDenom: "uusdc",
				Amount:      sdk.ZeroInt(),
				MinOutput:   sdk.NewInt(500000),
				Receiver:    "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: true,
			errMsg:    "amount must be positive",
		},
		{
			name: "zero min output",
			params: NobleSwapParams{
				InputDenom:  "uatom",
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.ZeroInt(),
				Receiver:    "cosmos1abc123def456ghi789jkl012mno345pqr678st",
			},
			expectErr: true,
			errMsg:    "min output must be positive",
		},
		{
			name: "empty receiver",
			params: NobleSwapParams{
				InputDenom:  "uatom",
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.NewInt(500000),
			},
			expectErr: true,
			errMsg:    "receiver cannot be empty",
		},
		{
			name: "invalid receiver address",
			params: NobleSwapParams{
				InputDenom:  "uatom",
				OutputDenom: "uusdc",
				Amount:      sdk.NewInt(1000000),
				MinOutput:   sdk.NewInt(500000),
				Receiver:    "invalid",
			},
			expectErr: true,
			errMsg:    "invalid receiver address",
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

func TestNobleLiquidityParams_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		params    NobleLiquidityParams
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid params",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token0:    "uusdc",
				Token1:    "uatom",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.NewInt(1),
			},
			expectErr: false,
		},
		{
			name: "empty pool ID",
			params: NobleLiquidityParams{
				Token0:    "uusdc",
				Token1:    "uatom",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.NewInt(1),
			},
			expectErr: true,
			errMsg:    "pool ID cannot be empty",
		},
		{
			name: "empty token0",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token1:    "uatom",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.NewInt(1),
			},
			expectErr: true,
			errMsg:    "token0 cannot be empty",
		},
		{
			name: "empty token1",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token0:    "uusdc",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.NewInt(1),
			},
			expectErr: true,
			errMsg:    "token1 cannot be empty",
		},
		{
			name: "zero amount0",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token0:    "uusdc",
				Token1:    "uatom",
				Amount0:   sdk.ZeroInt(),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.NewInt(1),
			},
			expectErr: true,
			errMsg:    "amount0 must be positive",
		},
		{
			name: "zero amount1",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token0:    "uusdc",
				Token1:    "uatom",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.ZeroInt(),
				MinShares: sdk.NewInt(1),
			},
			expectErr: true,
			errMsg:    "amount1 must be positive",
		},
		{
			name: "zero min shares",
			params: NobleLiquidityParams{
				PoolID:    "1",
				Token0:    "uusdc",
				Token1:    "uatom",
				Amount0:   sdk.NewInt(1000000),
				Amount1:   sdk.NewInt(500000),
				MinShares: sdk.ZeroInt(),
			},
			expectErr: true,
			errMsg:    "min shares must be positive",
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

func TestGetNobleUSDCPairs(t *testing.T) {
	pairs := GetNobleUSDCPairs()

	require.NotEmpty(t, pairs)
	require.GreaterOrEqual(t, len(pairs), 5)

	// Check that all pairs have USDC as quote
	for _, pair := range pairs {
		require.Equal(t, NobleUSDCDenom, pair.Quote)
		require.NotEmpty(t, pair.Base)
		require.NotEmpty(t, pair.Description)
	}
}

func TestTradingPair_String(t *testing.T) {
	testCases := []struct {
		name     string
		pair     TradingPair
		expected string
	}{
		{
			name: "with description",
			pair: TradingPair{
				Base:        "uatom",
				Quote:       "uusdc",
				Description: "ATOM/USDC",
			},
			expected: "ATOM/USDC",
		},
		{
			name: "without description",
			pair: TradingPair{
				Base:  "uatom",
				Quote: "uusdc",
			},
			expected: "uatom/uusdc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.pair.String()
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestTradingPair_Reverse(t *testing.T) {
	pair := TradingPair{
		Base:        "uatom",
		Quote:       "uusdc",
		Description: "ATOM/USDC",
	}

	reversed := pair.Reverse()

	require.Equal(t, "uusdc", reversed.Base)
	require.Equal(t, "uatom", reversed.Quote)
	require.Equal(t, "uusdc/uatom", reversed.Description)
}
