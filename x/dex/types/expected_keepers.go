package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// ICAControllerKeeper defines the expected ICA controller keeper
type ICAControllerKeeper interface {
	// RegisterInterchainAccount registers an ICA account
	RegisterInterchainAccount(
		ctx sdk.Context,
		connectionID, owner, version string,
	) error

	// SendTx sends a transaction to the ICA host
	SendTx(
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		connectionID, portID string,
		packetData icatypes.InterchainAccountPacketData,
		timeoutTimestamp uint64,
	) (uint64, error)

	// GetActiveChannelID gets the active channel for an ICA
	GetActiveChannelID(ctx sdk.Context, connectionID, portID string) (string, bool)

	// GetInterchainAccountAddress gets the ICA address on the host chain
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
}

// ConnectionKeeper defines the expected connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

// ChannelKeeper defines the expected channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (uint64, error)
}

// PortKeeper defines the expected port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// ScopedKeeper defines the expected scoped keeper
type ScopedKeeper interface {
	GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool)
	AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
}

// DIDKeeper defines the expected DID keeper
type DIDKeeper interface {
	// GetDIDDocument retrieves a DID document
	GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error)
}

// UCANKeeper defines the expected UCAN keeper (placeholder)
type UCANKeeper interface {
	// ValidateCapability validates a UCAN token for a specific capability
	ValidateCapability(ctx sdk.Context, token string, resource string, ability string) error
}

// DWNKeeper defines the expected DWN keeper
type DWNKeeper interface {
	// Placeholder interface - will be implemented when DWN methods are available
}
