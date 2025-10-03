package keeper

import (
	"fmt"

	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/dex/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	portkeeper "github.com/cosmos/ibc-go/v8/modules/core/05-port/keeper"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

// Keeper defines the DEX module keeper
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.Codec
	schema       collections.Schema
	authority    string

	// IBC dependencies
	ics4Wrapper  porttypes.ICS4Wrapper
	PortKeeper   *portkeeper.Keeper
	ScopedKeeper capabilitykeeper.ScopedKeeper

	// External module dependencies
	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	icaControllerKeeper types.ICAControllerKeeper
	connectionKeeper    types.ConnectionKeeper
	channelKeeper       types.ChannelKeeper
	didKeeper           types.DIDKeeper
	dwnKeeper           types.DWNKeeper

	// UCAN functionality
	ucanVerifier        *ucan.Verifier
	permissionValidator *PermissionValidator

	// Collections for state management
	Params          collections.Item[types.Params]
	Accounts        collections.Map[string, types.InterchainDEXAccount]
	AccountSequence collections.Sequence
	DIDToAccounts   collections.Map[string, types.DIDAccounts] // DID -> account mappings
	DIDActivities   collections.Map[string, types.DEXActivity] // DID activity records
}

// SetDIDKeeper sets the DID keeper (called after initialization)
func (k *Keeper) SetDIDKeeper(didKeeper types.DIDKeeper) {
	k.didKeeper = didKeeper
}

// SetDWNKeeper sets the DWN keeper (called after initialization)
func (k *Keeper) SetDWNKeeper(dwnKeeper types.DWNKeeper) {
	k.dwnKeeper = dwnKeeper
}

// NewKeeper creates a new DEX Keeper instance
func NewKeeper(
	appCodec codec.Codec,
	storeService store.KVStoreService,
	ics4Wrapper porttypes.ICS4Wrapper,
	portKeeper *portkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	icaControllerKeeper types.ICAControllerKeeper,
	connectionKeeper types.ConnectionKeeper,
	channelKeeper types.ChannelKeeper,
	didKeeper types.DIDKeeper,
	dwnKeeper types.DWNKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          appCodec,
		storeService: storeService,
		authority:    authority,

		// IBC dependencies
		ics4Wrapper:  ics4Wrapper,
		PortKeeper:   portKeeper,
		ScopedKeeper: scopedKeeper,

		// External dependencies
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		icaControllerKeeper: icaControllerKeeper,
		connectionKeeper:    connectionKeeper,
		channelKeeper:       channelKeeper,
		didKeeper:           didKeeper,
		dwnKeeper:           dwnKeeper,

		// State collections
		Params: collections.NewItem(
			sb,
			collections.NewPrefix(0),
			"params",
			codec.CollValue[types.Params](appCodec),
		),
		Accounts: collections.NewMap(
			sb,
			collections.NewPrefix(1),
			"accounts",
			collections.StringKey,
			codec.CollValue[types.InterchainDEXAccount](appCodec),
		),
		AccountSequence: collections.NewSequence(
			sb,
			collections.NewPrefix(2),
			"account_sequence",
		),
		DIDToAccounts: collections.NewMap(
			sb,
			collections.NewPrefix(3),
			"did_accounts",
			collections.StringKey,
			codec.CollValue[types.DIDAccounts](appCodec),
		),
		DIDActivities: collections.NewMap(
			sb,
			collections.NewPrefix(4),
			"did_activities",
			collections.StringKey,
			codec.CollValue[types.DEXActivity](appCodec),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.schema = schema

	// Initialize UCAN verifier and permission validator
	if didKeeper != nil {
		didResolver := &DEXDIDResolver{keeper: k}
		k.ucanVerifier = ucan.NewVerifier(didResolver)
		k.permissionValidator = NewPermissionValidator(k)
	}

	return k
}

// WithICS4Wrapper sets the ICS4Wrapper
func (k *Keeper) WithICS4Wrapper(wrapper porttypes.ICS4Wrapper) {
	k.ics4Wrapper = wrapper
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ibcexported.ModuleName+"-"+types.ModuleName)
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetPermissionValidator returns the UCAN permission validator
func (k Keeper) GetPermissionValidator() *PermissionValidator {
	return k.permissionValidator
}

// GetAccountKey generates a unique key for DEX accounts
func GetAccountKey(did, connectionID string) string {
	return fmt.Sprintf("%s:%s", did, connectionID)
}

// GetPortID generates a unique port ID for a DEX account
func GetPortID(did, connectionID string) string {
	return fmt.Sprintf("dex-%s-%s", did, connectionID)
}
