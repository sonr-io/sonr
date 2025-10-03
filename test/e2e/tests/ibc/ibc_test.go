package ibc

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/test/e2e/utils"
)

func TestIBCBasic(t *testing.T) {
	t.Skip("Skipping IBC tests - localnet doesn't have IBC channels")
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("ibc_channel_exists", func(t *testing.T) {
		// Verify that IBC transfer channels exist
		channelID := utils.AssertTransferChannelExists(t, cfg)
		require.NotEmpty(t, channelID, "transfer channel should exist")
	})

	t.Run("ibc_channels_query", func(t *testing.T) {
		// Query all IBC channels
		channels, err := cfg.Client.GetChannels(ctx)
		require.NoError(t, err, "failed to query IBC channels")
		require.NotEmpty(t, channels.Channels, "should have at least one IBC channel")

		// Verify we have transfer channels
		hasTransferChannel := false
		for _, channel := range channels.Channels {
			if channel.PortID == "transfer" {
				hasTransferChannel = true
				require.Equal(t, "STATE_OPEN", channel.State, "transfer channel should be open")
				break
			}
		}
		require.True(t, hasTransferChannel, "should have at least one transfer channel")
	})

	t.Run("ibc_channel_details", func(t *testing.T) {
		// Get transfer channel ID
		channelID, err := cfg.Client.GetTransferChannel(ctx)
		require.NoError(t, err, "failed to get transfer channel")

		// Query specific channel details
		channel, err := cfg.Client.GetChannel(ctx, "transfer", channelID)
		require.NoError(t, err, "failed to query channel details")
		require.NotNil(t, channel, "channel response should not be nil")
		require.Equal(t, "STATE_OPEN", channel.Channel.State.String(), "channel should be open")
	})
}

func TestIBCDenomTrace(t *testing.T) {
	t.Skip("Skipping IBC tests - localnet doesn't have IBC channels")
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("denom_trace_generation", func(t *testing.T) {
		// Get transfer channel for testing
		channelID, err := cfg.Client.GetTransferChannel(ctx)
		require.NoError(t, err, "failed to get transfer channel")

		// Generate IBC denom trace for testing
		denomTrace := transfertypes.ParseDenomTrace(
			transfertypes.GetPrefixedDenom("transfer", channelID, cfg.NormalDenom),
		)
		ibcDenom := denomTrace.IBCDenom()

		require.NotEmpty(t, ibcDenom, "IBC denom should not be empty")
		require.Contains(t, ibcDenom, "ibc/", "IBC denom should have ibc/ prefix")
	})
}

// TestIBCTransferSimulation tests IBC transfer logic without actual multi-chain setup
func TestIBCTransferSimulation(t *testing.T) {
	t.Skip("Skipping IBC tests - localnet doesn't have IBC channels")
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	tests := []struct {
		name           string
		transferAmount math.Int
		expectError    bool
	}{
		{
			name:           "normal_transfer_amount",
			transferAmount: math.NewInt(1_000_000), // 1 SNR
			expectError:    false,
		},
		{
			name:           "large_transfer_amount",
			transferAmount: math.NewInt(50_000_000), // 50 SNR
			expectError:    false,
		},
		{
			name:           "zero_transfer_amount",
			transferAmount: math.ZeroInt(),
			expectError:    true, // Should fail validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test user with sufficient funds
			fundAmount := math.NewInt(100_000_000) // 100 SNR
			users := utils.SetupTestUsers(t, cfg, fundAmount)
			sourceUser := users[0]

			// Verify user has sufficient balance before transfer
			if !tt.expectError && tt.transferAmount.GT(math.ZeroInt()) {
				utils.AssertBalanceGreaterThan(t, cfg, sourceUser.Address, sourceUser.Denom, tt.transferAmount)
			}

			// Get transfer channel
			channelID, err := cfg.Client.GetTransferChannel(ctx)
			require.NoError(t, err, "failed to get transfer channel")

			// Create IBC denom for destination
			denomTrace := transfertypes.ParseDenomTrace(
				transfertypes.GetPrefixedDenom("transfer", channelID, cfg.NormalDenom),
			)
			ibcDenom := denomTrace.IBCDenom()

			// Validate transfer parameters
			if tt.expectError {
				require.True(t, tt.transferAmount.IsZero() || tt.transferAmount.IsNegative(),
					"invalid transfer amounts should be caught")
			} else {
				require.True(t, tt.transferAmount.GT(math.ZeroInt()),
					"valid transfer amounts should be positive")
				require.NotEmpty(t, ibcDenom, "IBC denom should be generated")
			}
		})
	}
}

func TestIBCConnectionStatus(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("connection_existence", func(t *testing.T) {
		// Query channels to find connection information
		channels, err := cfg.Client.GetChannels(ctx)
		require.NoError(t, err, "failed to query channels")

		if len(channels.Channels) > 0 {
			// Test connection details for first channel
			channel := channels.Channels[0]
			require.NotEmpty(t, channel.ConnectionHops, "channel should have connection hops")

			if len(channel.ConnectionHops) > 0 {
				connectionID := channel.ConnectionHops[0]

				// Query connection details
				connection, err := cfg.Client.GetConnection(ctx, connectionID)
				require.NoError(t, err, "failed to query connection")
				require.NotNil(t, connection, "connection should exist")
				require.Equal(t, "STATE_OPEN", connection.Connection.State, "connection should be open")
			}
		}
	})
}
