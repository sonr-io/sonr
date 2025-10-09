package ante

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/crypto/keys"
	"github.com/sonr-io/crypto/ucan"
)

// UCANDecorator validates UCAN tokens in transactions
// This is a placeholder implementation that sets up the infrastructure
// for UCAN validation. In production, UCAN tokens would be passed
// in message fields or transaction extensions.
type UCANDecorator struct {
	verifier *ucan.Verifier
}

// NewUCANDecorator creates a new UCAN decorator
func NewUCANDecorator() UCANDecorator {
	// Create a basic DID resolver
	didResolver := &BasicDIDResolver{}
	verifier := ucan.NewVerifier(didResolver)

	return UCANDecorator{
		verifier: verifier,
	}
}

// AnteHandle validates UCAN tokens for transactions requiring authorization
func (ud UCANDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Skip validation in simulation mode
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Check if transaction has UCAN extension
	// This is where we would extract and validate UCAN tokens
	// For now, this is a placeholder that demonstrates the structure

	// Future implementation would:
	// 1. Extract UCAN token from transaction extensions or memo
	// 2. Validate the token using the verifier
	// 3. Check capabilities against message types
	// 4. Mark transaction as gasless if appropriate

	// Check if transaction qualifies for gasless execution
	if ud.isGaslessTransaction(ctx, tx) {
		// Mark context for gasless processing
		ctx = ctx.WithValue("gasless_ucan", true)
	}

	return next(ctx, tx, simulate)
}

// isGaslessTransaction checks if transaction qualifies for gasless execution
// This is a placeholder implementation
func (ud UCANDecorator) isGaslessTransaction(ctx sdk.Context, tx sdk.Tx) bool {
	// In production, this would check for UCAN tokens with gasless capabilities
	// For now, return false to maintain normal fee processing
	return false
}

// CheckTokenExpiration checks if UCAN token has expired
func (ud UCANDecorator) CheckTokenExpiration(ctx sdk.Context, token *ucan.Token) error {
	if token.ExpiresAt > 0 {
		currentTime := ctx.BlockTime().Unix()
		if currentTime > token.ExpiresAt {
			return fmt.Errorf("UCAN token has expired")
		}
	}

	// Check NotBefore
	if token.NotBefore > 0 {
		currentTime := ctx.BlockTime().Unix()
		if currentTime < token.NotBefore {
			return fmt.Errorf("UCAN token is not yet valid")
		}
	}

	return nil
}

// ValidateCapabilities validates UCAN capabilities against required permissions
func (ud UCANDecorator) ValidateCapabilities(token *ucan.Token, requiredCapabilities []string) error {
	// Check if token grants required capabilities
	for _, att := range token.Attenuations {
		if att.Capability.Grants(requiredCapabilities) {
			return nil
		}
	}

	return fmt.Errorf("UCAN token does not grant required capabilities")
}

// BasicDIDResolver implements ucan.DIDResolver for the ante handler
type BasicDIDResolver struct{}

// ResolveDIDKey resolves DID to public key for UCAN verification
func (r *BasicDIDResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	// This is a basic implementation that accepts all DIDs
	// In production, this would query the DID module
	return keys.Parse(did)
}

// ConditionalUCANDecorator wraps UCAN decorator to skip for certain transactions
type ConditionalUCANDecorator struct {
	decorator sdk.AnteDecorator
}

// NewConditionalUCANDecorator creates a conditional UCAN decorator
func NewConditionalUCANDecorator(decorator sdk.AnteDecorator) ConditionalUCANDecorator {
	return ConditionalUCANDecorator{decorator: decorator}
}

// AnteHandle conditionally applies UCAN validation
func (cud ConditionalUCANDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Skip UCAN validation if already marked as gasless WebAuthn
	if ctx.Value("bypass_ucan") != nil {
		return next(ctx, tx, simulate)
	}

	// Apply UCAN validation
	return cud.decorator.AnteHandle(ctx, tx, simulate, next)
}
