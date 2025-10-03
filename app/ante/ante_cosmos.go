package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmoscosmosante "github.com/cosmos/evm/ante/cosmos"
	evmante "github.com/cosmos/evm/ante/evm"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	circuitante "cosmossdk.io/x/circuit/ante"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
)

// NewCosmosAnteHandler creates the default ante handler for Cosmos SDK transactions.
// It sets up a chain of decorators that perform various checks and operations:
//   - Rejects Ethereum transactions in Cosmos context
//   - Enforces authz limitations
//   - Sets up transaction context
//   - Validates basic transaction properties
//   - Handles WebAuthn gasless transactions
//   - Handles gas consumption and fee deduction
//   - Performs signature verification
//   - Manages account sequences
//   - Handles IBC-specific checks
func NewCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	// Determine if we should use enhanced gasless mode
	// Enhanced mode allows address generation from credentials for true gasless onboarding
	enhancedGaslessMode := options.EnableEnhancedGasless

	// Build the decorator chain
	decorators := []sdk.AnteDecorator{
		// WebAuthn bypass - must be first to intercept WebAuthn transactions
		NewWebAuthnBypassDecorator(),
		evmoscosmosante.NewRejectMessagesDecorator(), // reject MsgEthereumTxs
		evmoscosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),

		ante.NewSetUpContextDecorator(),
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),

		// UCAN validation - must come before fee deduction for gasless support
		NewConditionalUCANDecorator(NewUCANDecorator()),
		evmoscosmosante.NewMinGasPriceDecorator(
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.ControlPanelKeeper,
		),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),

		// WebAuthn gasless transaction support - must come before fee deduction
		// Enhanced mode allows true gasless onboarding without pre-existing accounts
		NewWebAuthnGaslessDecorator(options.AccountKeeper, options.DidKeeper, enhancedGaslessMode),

		// Conditional fee deduction - skips fees for gasless WebAuthn and UCAN
		NewUCANGaslessDecorator(
			NewConditionalFeeDecorator(ante.NewDeductFeeDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeegrantKeeper,
				options.TxFeeChecker,
			)),
		),
	}

	// Add signature verification decorators
	// In enhanced gasless mode, we wrap these to be conditional
	if enhancedGaslessMode {
		// Conditional decorators that skip verification for gasless transactions
		decorators = append(
			decorators,
			NewConditionalPubKeyDecorator(ante.NewSetPubKeyDecorator(options.AccountKeeper)),
			NewConditionalSigCountDecorator(
				ante.NewValidateSigCountDecorator(options.AccountKeeper),
			),
			NewConditionalSigGasDecorator(
				ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
			),
			NewConditionalSignatureDecorator(
				ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
			),
		)
	} else {
		// Standard signature verification decorators
		decorators = append(decorators,
			ante.NewSetPubKeyDecorator(options.AccountKeeper),
			ante.NewValidateSigCountDecorator(options.AccountKeeper),
			ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
			ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		)
	}

	// Add remaining decorators
	decorators = append(decorators,
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	)

	return sdk.ChainAnteDecorators(decorators...)
}
