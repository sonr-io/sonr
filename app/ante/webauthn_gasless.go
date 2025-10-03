// Package ante provides ante handler implementations for transaction processing
// in the Sonr blockchain. It supports both Cosmos SDK and Ethereum transactions.
package ante

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// WebAuthnGaslessDecorator provides gasless transaction processing for WebAuthn registration.
// This decorator identifies WebAuthn credential registration messages and bypasses fee deduction,
// enabling users to create their first decentralized identity without requiring existing tokens.
//
// The decorator supports two modes:
// 1. Standard mode: Requires a controller address (for users with existing accounts)
// 2. Enhanced mode: Generates controller address from credential (for brand new users)
//
// Security considerations:
// - Only applies to MsgRegisterWebAuthnCredential messages
// - Validates WebAuthn credential authenticity before fee waiving
// - Prevents abuse through cryptographic WebAuthn requirements
// - Limited to one gasless transaction per unique credential
type WebAuthnGaslessDecorator struct {
	accountKeeper AccountKeeper
	didKeeper     WebAuthnKeeperInterface
	enhancedMode  bool // If true, allows address generation from credentials
}

// NewWebAuthnGaslessDecorator creates a new WebAuthn gasless transaction decorator.
// This decorator must be placed in the ante handler chain BEFORE the fee deduction decorator
// to effectively bypass fee requirements for qualifying WebAuthn transactions.
//
// Set enhancedMode to true to enable automatic address generation from credentials,
// allowing truly gasless onboarding without pre-existing accounts.
func NewWebAuthnGaslessDecorator(
	accountKeeper AccountKeeper,
	didKeeper WebAuthnKeeperInterface,
	enhancedMode bool,
) WebAuthnGaslessDecorator {
	return WebAuthnGaslessDecorator{
		accountKeeper: accountKeeper,
		didKeeper:     didKeeper,
		enhancedMode:  enhancedMode,
	}
}

// AnteHandle processes the transaction and determines if WebAuthn gasless processing applies.
// For qualifying WebAuthn registration transactions, it sets transaction fees to zero
// and validates the WebAuthn credential to prevent abuse.
//
// Gasless criteria:
// 1. Transaction contains exactly one MsgRegisterWebAuthnCredential
// 2. WebAuthn credential passes cryptographic validation
// 3. Credential ID has not been used before (prevents replay attacks)
// 4. Transaction sender account exists or can be created
//
// In enhanced mode, if no controller is provided, it generates one from the credential.
func (wgd WebAuthnGaslessDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Process during both CheckTx (simulation) and DeliverTx (execution)
	// This ensures gasless flags are set during mempool validation

	msgs := tx.GetMsgs()

	// Debug: Log entry into gasless decorator
	ctx.Logger().Debug("WebAuthnGaslessDecorator: processing transaction",
		"msg_count", len(msgs), "sim", sim)

	// Check if this is a single WebAuthn registration transaction
	if len(msgs) != 1 {
		// Multi-message transactions don't qualify for gasless processing
		return next(ctx, tx, sim)
	}

	msg, ok := msgs[0].(*didtypes.MsgRegisterWebAuthnCredential)
	if !ok {
		// Not a WebAuthn registration message
		return next(ctx, tx, sim)
	}

	// Check if WebAuthn bypass validation already occurred
	if bypassed, ok := ctx.Value("webauthn_bypass_validated").(bool); !ok || !bypassed {
		// Validate the WebAuthn credential to prevent abuse (if not already validated)
		if err := msg.WebauthnCredential.ValidateStructure(); err != nil {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInvalidRequest,
				"invalid WebAuthn credential for gasless transaction: %v", err,
			)
		}
	}

	// Prevent credential reuse (anti-replay protection)
	if wgd.didKeeper.HasExistingCredential(ctx, msg.WebauthnCredential.CredentialId) {
		return ctx, errorsmod.Wrapf(
			errortypes.ErrInvalidRequest,
			"WebAuthn credential already registered: %s", msg.WebauthnCredential.CredentialId,
		)
	}

	// Handle controller address
	var controllerAddr sdk.AccAddress

	// Enhanced mode: Generate address from credential if not provided
	if wgd.enhancedMode && msg.GetController() == "" {
		// Generate deterministic address from credential ID
		controllerAddr = GenerateAddressFromCredential(msg.WebauthnCredential.CredentialId)

		// Note: We can't modify the message directly in most cases,
		// but we can pass the generated address through context
		ctx = ctx.WithValue("generated_controller", controllerAddr.String())

		ctx.Logger().Info(
			"Generated controller address for gasless WebAuthn registration",
			"generated_address", controllerAddr.String(),
			"credential_id", msg.WebauthnCredential.CredentialId,
		)
	} else {
		// Standard mode or enhanced mode with provided controller
		controllerStr := msg.GetController()
		if controllerStr == "" {
			return ctx, errorsmod.Wrap(
				errortypes.ErrInvalidAddress,
				"controller address required for WebAuthn registration",
			)
		}

		controllerAddr, err = sdk.AccAddressFromBech32(controllerStr)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInvalidAddress,
				"invalid controller address: %v", err,
			)
		}
	}

	// Ensure the account exists (simulation-safe)
	account := wgd.accountKeeper.GetAccount(ctx, controllerAddr)
	if account == nil {
		// Create account if it doesn't exist (common for first-time WebAuthn users)
		// Only actually create during execution, not simulation
		if !sim {
			account = wgd.accountKeeper.NewAccountWithAddress(ctx, controllerAddr)
			wgd.accountKeeper.SetAccount(ctx, account)

			ctx.Logger().Info(
				"Created new account for gasless WebAuthn registration",
				"address", controllerAddr.String(),
			)
		} else {
			// During simulation (CheckTx), just create a temporary account for validation
			// We don't assign it back to account variable since it's only for validation
			wgd.accountKeeper.NewAccountWithAddress(ctx, controllerAddr)
		}
	}

	// Mark this transaction as gasless
	gaslessCtx := ctx.WithValue("webauthn_gasless", true)

	// In enhanced mode, also mark to skip signature verification
	if wgd.enhancedMode {
		gaslessCtx = gaslessCtx.WithValue("skip_sig_verification", true)
		gaslessCtx = gaslessCtx.WithValue("skip_pubkey_verification", true)
	}

	// Log the gasless transaction for monitoring and security purposes
	ctx.Logger().Info(
		"Processing gasless WebAuthn registration",
		"controller", controllerAddr.String(),
		"credential_id", msg.WebauthnCredential.CredentialId,
		"username", msg.Username,
		"auto_vault", msg.AutoCreateVault,
		"enhanced_mode", wgd.enhancedMode,
	)

	return next(gaslessCtx, tx, sim)
}

// GenerateAddressFromCredential generates a deterministic address from a WebAuthn credential ID.
// This ensures the same credential always generates the same address, allowing for
// predictable account creation without requiring pre-existing blockchain state.
func GenerateAddressFromCredential(credentialID string) sdk.AccAddress {
	// Create a deterministic hash from the credential ID
	// Add a domain separator to prevent collisions with other address generation methods
	domainSeparator := "webauthn_gasless_v1"
	data := domainSeparator + credentialID

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))

	// Take the first 20 bytes for the address (Ethereum-compatible)
	return sdk.AccAddress(hash[:20])
}

// GenerateDIDFromCredential generates a deterministic DID from a WebAuthn credential.
// This creates a unique, reproducible DID for each WebAuthn credential.
func GenerateDIDFromCredential(credentialID string, username string) string {
	// Create a deterministic hash from credential ID and username
	data := credentialID + ":" + username
	hash := sha256.Sum256([]byte(data))

	// Create a DID with the sonr method
	// Format: did:sonr:<hex-encoded-hash-prefix>
	didSuffix := hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter DIDs
	return fmt.Sprintf("did:sonr:%s", didSuffix)
}

// ConditionalFeeDecorator wraps the standard fee deduction decorator to conditionally
// skip fee deduction for gasless WebAuthn transactions marked by WebAuthnGaslessDecorator.
// This is the simplest possible implementation that reuses all existing SDK infrastructure.
type ConditionalFeeDecorator struct {
	standardFeeDecorator sdk.AnteDecorator
}

// NewConditionalFeeDecorator creates a decorator that conditionally skips fee deduction
// for gasless transactions while maintaining all standard fee logic for other transactions.
func NewConditionalFeeDecorator(standardFeeDecorator sdk.AnteDecorator) ConditionalFeeDecorator {
	return ConditionalFeeDecorator{
		standardFeeDecorator: standardFeeDecorator,
	}
}

// AnteHandle processes transactions and conditionally skips fee deduction for gasless WebAuthn transactions.
func (cfd ConditionalFeeDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check if this transaction was marked as gasless by WebAuthnGaslessDecorator
	if gasless, ok := ctx.Value("webauthn_gasless").(bool); ok && gasless {
		// Skip fee deduction for gasless WebAuthn transactions
		ctx.Logger().Info("Waiving fees for gasless WebAuthn registration")
		return next(ctx, tx, sim)
	}

	// For all other transactions, use the standard fee deduction decorator
	return cfd.standardFeeDecorator.AnteHandle(ctx, tx, sim, next)
}

// ConditionalSignatureDecorator wraps signature verification to skip it for gasless transactions
// This is used in enhanced mode where WebAuthn itself is the authentication mechanism.
type ConditionalSignatureDecorator struct {
	sigVerifyDecorator sdk.AnteDecorator
}

// NewConditionalSignatureDecorator creates a decorator that conditionally skips signature verification
func NewConditionalSignatureDecorator(
	sigVerifyDecorator sdk.AnteDecorator,
) ConditionalSignatureDecorator {
	return ConditionalSignatureDecorator{
		sigVerifyDecorator: sigVerifyDecorator,
	}
}

// AnteHandle conditionally skips signature verification for gasless transactions
func (csd ConditionalSignatureDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check if signature verification should be skipped
	if skip, ok := ctx.Value("skip_sig_verification").(bool); ok && skip {
		ctx.Logger().Debug("Skipping signature verification for gasless transaction")
		return next(ctx, tx, sim)
	}

	// For all other transactions, use the standard signature verification
	return csd.sigVerifyDecorator.AnteHandle(ctx, tx, sim, next)
}

// ConditionalPubKeyDecorator wraps pubkey setting to skip it for gasless transactions
// This is used in enhanced mode where accounts are created without pre-existing keys.
type ConditionalPubKeyDecorator struct {
	setPubKeyDecorator sdk.AnteDecorator
}

// NewConditionalPubKeyDecorator creates a decorator that conditionally skips pubkey setting
func NewConditionalPubKeyDecorator(
	setPubKeyDecorator sdk.AnteDecorator,
) ConditionalPubKeyDecorator {
	return ConditionalPubKeyDecorator{
		setPubKeyDecorator: setPubKeyDecorator,
	}
}

// AnteHandle conditionally skips pubkey setting for gasless transactions
func (cpd ConditionalPubKeyDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check if pubkey verification should be skipped
	if skip, ok := ctx.Value("skip_pubkey_verification").(bool); ok && skip {
		ctx.Logger().Debug("Skipping pubkey setting for gasless transaction")
		return next(ctx, tx, sim)
	}

	// For all other transactions, use the standard pubkey decorator
	return cpd.setPubKeyDecorator.AnteHandle(ctx, tx, sim, next)
}

// ConditionalSigCountDecorator wraps signature count validation to skip it for gasless transactions
// This is critical for gasless WebAuthn transactions which have no signatures to validate.
type ConditionalSigCountDecorator struct {
	sigCountDecorator sdk.AnteDecorator
}

// NewConditionalSigCountDecorator creates a decorator that conditionally skips signature count validation
func NewConditionalSigCountDecorator(
	sigCountDecorator sdk.AnteDecorator,
) ConditionalSigCountDecorator {
	return ConditionalSigCountDecorator{
		sigCountDecorator: sigCountDecorator,
	}
}

// AnteHandle conditionally skips signature count validation for gasless transactions
func (cscd ConditionalSigCountDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check if signature verification should be skipped (implies no signatures to count)
	if skip, ok := ctx.Value("skip_sig_verification").(bool); ok && skip {
		ctx.Logger().Debug("Skipping signature count validation for gasless transaction")
		return next(ctx, tx, sim)
	}

	// For all other transactions, use the standard signature count validator
	return cscd.sigCountDecorator.AnteHandle(ctx, tx, sim, next)
}

// ConditionalSigGasDecorator wraps signature gas consumption to skip it for gasless transactions
// This prevents "no signatures supplied" errors for gasless WebAuthn transactions.
type ConditionalSigGasDecorator struct {
	sigGasDecorator sdk.AnteDecorator
}

// NewConditionalSigGasDecorator creates a decorator that conditionally skips signature gas consumption
func NewConditionalSigGasDecorator(
	sigGasDecorator sdk.AnteDecorator,
) ConditionalSigGasDecorator {
	return ConditionalSigGasDecorator{
		sigGasDecorator: sigGasDecorator,
	}
}

// AnteHandle conditionally skips signature gas consumption for gasless transactions
func (csgd ConditionalSigGasDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check if signature verification should be skipped (implies no signature gas consumption needed)
	if skip, ok := ctx.Value("skip_sig_verification").(bool); ok && skip {
		ctx.Logger().Debug("Skipping signature gas consumption for gasless transaction")
		return next(ctx, tx, sim)
	}

	// For all other transactions, use the standard signature gas decorator
	return csgd.sigGasDecorator.AnteHandle(ctx, tx, sim, next)
}
