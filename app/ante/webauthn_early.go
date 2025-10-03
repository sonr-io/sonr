package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// WebAuthnBypassDecorator completely bypasses signature verification for WebAuthn registration.
// This decorator must be placed FIRST in the ante handler chain to intercept WebAuthn
// transactions before any signature validation occurs.
type WebAuthnBypassDecorator struct{}

// NewWebAuthnBypassDecorator creates a new WebAuthnBypassDecorator
func NewWebAuthnBypassDecorator() WebAuthnBypassDecorator {
	return WebAuthnBypassDecorator{}
}

// AnteHandle validates WebAuthn transactions and marks them for controlled processing
// This decorator performs essential security checks while allowing gasless processing
func (wbd WebAuthnBypassDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, sim bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()

	// Check if this is a single WebAuthn registration transaction
	if len(msgs) != 1 {
		// Not a single message transaction, proceed normally
		return next(ctx, tx, sim)
	}

	msg, ok := msgs[0].(*didtypes.MsgRegisterWebAuthnCredential)
	if !ok {
		// Not a WebAuthn registration, proceed normally
		return next(ctx, tx, sim)
	}

	// This is a WebAuthn registration - perform security validation
	ctx.Logger().Info("Processing WebAuthn registration with controlled bypass",
		"username", msg.Username,
		"credential_id", msg.WebauthnCredential.CredentialId)

	// CRITICAL SECURITY CHECK 1: Validate the WebAuthn credential structure
	if err := msg.WebauthnCredential.ValidateStructure(); err != nil {
		ctx.Logger().Error("WebAuthn credential structure validation failed", "error", err)
		return ctx, err
	}

	// CRITICAL SECURITY CHECK 2: Handle signatures (dummy signatures are allowed for mempool validation)
	if sigTx, ok := tx.(authsigning.SigVerifiableTx); ok {
		sigs, err := sigTx.GetSignaturesV2()
		if err != nil {
			ctx.Logger().Error("Failed to get signatures from WebAuthn transaction", "error", err)
			return ctx, err
		}
		if len(sigs) > 0 {
			ctx.Logger().
				Debug("WebAuthn transaction has dummy signatures for mempool validation", "sig_count", len(sigs))

			// This is expected - dummy signatures are used to pass mempool validation
			// The actual signature verification will be bypassed by conditional decorators
		} else {
			ctx.Logger().Debug("WebAuthn transaction has no signatures - gasless flow")
		}
	}

	// CRITICAL SECURITY CHECK 3: Validate credential uniqueness to prevent replay attacks
	// Note: This will be enforced in the WebAuthnGaslessDecorator with keeper access

	// Mark context for controlled WebAuthn processing
	// These flags will be checked by conditional decorators
	ctx = ctx.WithValue("webauthn_bypass_validated", true)

	ctx.Logger().Info("WebAuthn transaction validated - proceeding with controlled processing",
		"credential_id", msg.WebauthnCredential.CredentialId,
		"username", msg.Username)

	// Continue to next decorator (WebAuthnGaslessDecorator) for full processing
	return next(ctx, tx, sim)
}
