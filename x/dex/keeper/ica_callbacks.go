package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/sonr-io/sonr/x/dex/types"
)

// OnChanOpenInit handles channel initialization for ICA
func (k Keeper) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	// Claim capability for the channel
	capability := k.PortKeeper.BindPort(ctx, portID)
	if err := k.ScopedKeeper.ClaimCapability(ctx, capability, channelCapabilityPath(portID, channelID)); err != nil {
		return fmt.Errorf("failed to claim capability: %w", err)
	}

	k.Logger(ctx).Info("ICA channel initialized",
		"port", portID,
		"channel", channelID,
		"connection", connectionHops[0],
	)

	return nil
}

// OnChanOpenAck handles channel acknowledgment for ICA
func (k Keeper) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// Parse counterparty version to get ICA address
	metadata, err := parseICAMetadata(counterpartyVersion)
	if err != nil {
		return fmt.Errorf("failed to parse ICA metadata: %w", err)
	}

	// Update DEX account with ICA address
	if err := k.OnICAAccountCreated(ctx, portID, metadata.Address); err != nil {
		return fmt.Errorf("failed to update DEX account: %w", err)
	}

	k.Logger(ctx).Info("ICA channel acknowledged",
		"port", portID,
		"channel", channelID,
		"ica_address", metadata.Address,
	)

	return nil
}

// OnAcknowledgementPacket handles ICA packet acknowledgments
func (k Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := k.cdc.Unmarshal(acknowledgement, &ack); err != nil {
		return fmt.Errorf("failed to unmarshal acknowledgement: %w", err)
	}

	// Log the acknowledgment
	k.Logger(ctx).Info("ICA packet acknowledged",
		"sequence", packet.Sequence,
		"source_port", packet.SourcePort,
		"source_channel", packet.SourceChannel,
		"success", ack.Success(),
	)

	// Emit event for successful/failed transaction
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeICAPacketAcknowledged,
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", packet.Sequence)),
			sdk.NewAttribute("source_port", packet.SourcePort),
			sdk.NewAttribute("source_channel", packet.SourceChannel),
			sdk.NewAttribute("success", fmt.Sprintf("%t", ack.Success())),
		),
	)

	return nil
}

// OnTimeoutPacket handles ICA packet timeouts
func (k Keeper) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	k.Logger(ctx).Error("ICA packet timed out",
		"sequence", packet.Sequence,
		"source_port", packet.SourcePort,
		"source_channel", packet.SourceChannel,
	)

	// Emit timeout event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeICAPacketTimeout,
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", packet.Sequence)),
			sdk.NewAttribute("source_port", packet.SourcePort),
			sdk.NewAttribute("source_channel", packet.SourceChannel),
		),
	)

	return nil
}

// Helper functions

func channelCapabilityPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", "ports", portID, "channels", channelID)
}

// ICAMetadata represents parsed ICA metadata from version string
type ICAMetadata struct {
	Address string
	Version string
}

// parseICAMetadata extracts ICA address from version metadata
func parseICAMetadata(version string) (*ICAMetadata, error) {
	// This is a simplified version - actual parsing depends on ICA version format
	// The version string typically contains JSON with the ICA address
	// For now, we'll return a placeholder
	return &ICAMetadata{
		Address: version, // In reality, this would be parsed from JSON
		Version: "ics27-1",
	}, nil
}
