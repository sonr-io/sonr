package ante

import (
	"context"

	addresscodec "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
)

// WebAuthnKeeperInterface defines the required methods from the DID keeper for WebAuthn gasless processing
type WebAuthnKeeperInterface interface {
	// HasExistingCredential checks if this credential ID already exists
	HasExistingCredential(ctx sdk.Context, credentialId string) bool
}

// BankKeeper defines the contract needed for supply related APIs.
// It provides methods for checking send permissions and transferring coins
// between accounts and modules.
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
}

// AccountKeeper defines the account management interface required by ante handlers.
// It provides methods for account creation, retrieval, modification, and
// sequence number management.
type AccountKeeper interface {
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, account sdk.AccountI)
	RemoveAccount(ctx context.Context, account sdk.AccountI)
	GetParams(ctx context.Context) (params authtypes.Params)
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
	AddressCodec() addresscodec.Codec
}

// HandlerOptions defines the list of module keepers and configurations required
// to run the ante handler decorators. It includes both standard Cosmos SDK
// keepers and EVM-specific components for processing different transaction types.
type HandlerOptions struct {
	Cdc                    codec.BinaryCodec
	AccountKeeper          AccountKeeper
	BankKeeper             BankKeeper
	FeegrantKeeper         ante.FeegrantKeeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	TxFeeChecker           ante.TxFeeChecker // safe to be nil

	MaxTxGasWanted     uint64
	FeeMarketKeeper    anteinterfaces.FeeMarketKeeper
	EvmKeeper          anteinterfaces.EVMKeeper
	ControlPanelKeeper anteinterfaces.ControlPanelKeeper

	IBCKeeper     *ibckeeper.Keeper
	CircuitKeeper *circuitkeeper.Keeper

	// WebAuthn gasless transaction support
	DidKeeper             WebAuthnKeeperInterface
	EnableEnhancedGasless bool // Enable enhanced gasless mode for true onboarding without pre-existing accounts

	// UCAN module keepers for permission validation
	DwnKeeper interface{} // Will be cast to proper type in decorator
	DexKeeper interface{} // Will be cast to proper type in decorator
	SvcKeeper interface{} // Will be cast to proper type in decorator
}

// Validate checks if all required keepers and handlers are properly initialized.
// It ensures that the HandlerOptions struct has all necessary components to
// process transactions without nil pointer errors.
func (options HandlerOptions) Validate() error {
	if options.Cdc == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "codec is required for AnteHandler")
	}
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SigGasConsumer == nil {
		return errorsmod.Wrap(
			errortypes.ErrLogic,
			"signature gas consumer is required for AnteHandler",
		)
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for AnteHandler")
	}
	if options.CircuitKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "circuit keeper is required for ante builder")
	}

	if options.TxFeeChecker == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "tx fee checker is required for AnteHandler")
	}
	if options.FeeMarketKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for AnteHandler")
	}
	if options.ControlPanelKeeper == nil {
		return errorsmod.Wrap(
			errortypes.ErrLogic,
			"control panel keeper is required for AnteHandler",
		)
	}

	return nil
}
