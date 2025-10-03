package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// ValidateConnection validates an IBC connection exists and is open
func (k Keeper) ValidateConnection(ctx sdk.Context, connectionID string) error {
	connection, found := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return fmt.Errorf("connection %s not found", connectionID)
	}

	if connection.State != connectiontypes.OPEN {
		return fmt.Errorf("connection %s is not open", connectionID)
	}

	return nil
}

// GetChannelCapability retrieves the channel capability
func (k Keeper) GetChannelCapability(ctx sdk.Context, portID, channelID string) (*capabilitytypes.Capability, error) {
	capability, ok := k.ScopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !ok {
		return nil, fmt.Errorf(
			"capability not found for port %s channel %s: %w",
			portID, channelID,
			channeltypes.ErrChannelCapabilityNotFound,
		)
	}
	return capability, nil
}

// GetChannel retrieves an IBC channel
func (k Keeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return k.channelKeeper.GetChannel(ctx, portID, channelID)
}

// GetNextSequenceSend returns the next sequence send for a channel
func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
}

// SendPacket sends an IBC packet
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	portID string,
	channelID string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	return k.channelKeeper.SendPacket(
		ctx,
		chanCap,
		portID,
		channelID,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
}

// BindPort binds a port and claims the capability
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	capability := k.PortKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, capability, host.PortPath(portID))
}

// ClaimCapability claims a capability
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.ScopedKeeper.ClaimCapability(ctx, cap, name)
}

// AuthenticateCapability authenticates a capability
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.ScopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// GetConnectionEnd retrieves an IBC connection
func (k Keeper) GetConnectionEnd(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool) {
	return k.connectionKeeper.GetConnection(ctx, connectionID)
}

// IsConnectionOpen checks if a connection is open
func (k Keeper) IsConnectionOpen(ctx sdk.Context, connectionID string) bool {
	connection, found := k.GetConnectionEnd(ctx, connectionID)
	return found && connection.State == connectiontypes.OPEN
}

// IsChannelOpen checks if a channel is open
func (k Keeper) IsChannelOpen(ctx sdk.Context, portID, channelID string) bool {
	channel, found := k.GetChannel(ctx, portID, channelID)
	return found && channel.State == channeltypes.OPEN
}
