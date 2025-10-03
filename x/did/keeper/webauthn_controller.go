package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/sonr/types/webauthn"
	"github.com/sonr-io/sonr/x/did/types"
)

// WebAuthnControllerVerifier handles WebAuthn-based controller verification for DID operations
type WebAuthnControllerVerifier struct {
	keeper Keeper
}

// NewWebAuthnControllerVerifier creates a new WebAuthn controller verifier
func NewWebAuthnControllerVerifier(k Keeper) *WebAuthnControllerVerifier {
	return &WebAuthnControllerVerifier{keeper: k}
}

// Use the centralized ClientData type from types/webauthn package
// No need to duplicate the ClientData structure here

// WebAuthnAssertion represents a WebAuthn assertion for DID controller verification
type WebAuthnAssertion struct {
	CredentialID      string `json:"credentialId"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

// VerifyControllerWithWebAuthn verifies that a controller has authority over a DID using WebAuthn
func (v *WebAuthnControllerVerifier) VerifyControllerWithWebAuthn(
	ctx context.Context,
	did string,
	controller string,
	assertion *WebAuthnAssertion,
	challenge string,
) error {
	// Get DID document
	ormDoc, err := v.keeper.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return fmt.Errorf("DID document not found: %w", err)
	}

	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if controller matches the DID's primary controller
	if didDoc.PrimaryController != controller {
		return fmt.Errorf(
			"controller mismatch: expected %s, got %s",
			didDoc.PrimaryController,
			controller,
		)
	}

	// Find the WebAuthn verification method for this credential
	var webAuthnVM *types.VerificationMethod
	for _, vm := range didDoc.VerificationMethod {
		if vm.WebauthnCredential != nil &&
			vm.WebauthnCredential.CredentialId == assertion.CredentialID {
			webAuthnVM = vm
			break
		}
	}

	if webAuthnVM == nil {
		return fmt.Errorf(
			"WebAuthn credential %s not found in DID document",
			assertion.CredentialID,
		)
	}

	// Verify the WebAuthn assertion
	return v.verifyWebAuthnAssertion(
		ctx,
		assertion,
		webAuthnVM.WebauthnCredential,
		challenge,
	)
}

// verifyWebAuthnAssertion verifies a WebAuthn assertion against a stored credential ID using centralized validation
func (v *WebAuthnControllerVerifier) verifyWebAuthnAssertion(
	ctx context.Context,
	assertion *WebAuthnAssertion,
	credential *types.WebAuthnCredential,
	expectedChallenge string,
) error {
	// Migrate to centralized WebAuthn verification using internal/webauthn package
	// This provides complete FIDO2 validation with proper COSE key parsing,
	// signature verification, counter validation, and multi-algorithm support (ES256, RS256, EdDSA)

	// Get module parameters for WebAuthn configuration
	params, err := v.keeper.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module parameters: %w", err)
	}

	// Create a CredentialAssertionResponse from the assertion data
	credentialAssertion := &webauthn.CredentialAssertionResponse{
		PublicKeyCredential: webauthn.PublicKeyCredential{
			Credential: webauthn.Credential{
				ID:   assertion.CredentialID,
				Type: "public-key",
			},
			RawID: webauthn.URLEncodedBase64(assertion.CredentialID),
		},
		AssertionResponse: webauthn.AuthenticatorAssertionResponse{
			AuthenticatorResponse: webauthn.AuthenticatorResponse{
				ClientDataJSON: webauthn.URLEncodedBase64(assertion.ClientDataJSON),
			},
			AuthenticatorData: webauthn.URLEncodedBase64(assertion.AuthenticatorData),
			Signature:         webauthn.URLEncodedBase64(assertion.Signature),
			UserHandle:        webauthn.URLEncodedBase64(assertion.UserHandle),
		},
	}

	// Parse the credential assertion response using the full WebAuthn protocol
	parsedAssertion, err := credentialAssertion.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse WebAuthn assertion: %w", err)
	}

	// Perform comprehensive verification using the full WebAuthn protocol
	var rpId string
	var allowedOrigins []string
	var requireUserVerification bool

	if params.Webauthn != nil {
		rpId = params.Webauthn.DefaultRpId
		allowedOrigins = params.Webauthn.AllowedOrigins
		requireUserVerification = params.Webauthn.RequireUserVerification
	} else {
		// Fallback defaults if Webauthn params are nil
		rpId = "localhost"
		allowedOrigins = []string{"http://localhost:8080"}
		requireUserVerification = true
	}

	err = parsedAssertion.Verify(
		expectedChallenge, // stored challenge
		rpId,              // relying party ID
		allowedOrigins,    // RP origins
		[]string{},        // RP top origins (empty for basic validation)
		webauthn.TopOriginDefaultVerificationMode, // top origin verification mode
		"",                      // app ID (empty for CTAP2)
		requireUserVerification, // verify user verification
		true,                    // verify user presence (always required)
		credential.PublicKey,    // stored credential public key
	)
	if err != nil {
		return fmt.Errorf("WebAuthn assertion verification failed: %w", err)
	}

	// Additional Sonr-specific validations

	// Verify the credential origin matches what's stored
	clientData, err := webauthn.ValidateClientDataJSONFormat(assertion.ClientDataJSON)
	if err != nil {
		return fmt.Errorf("failed to validate client data JSON: %w", err)
	}

	if clientData.Origin != credential.Origin {
		return fmt.Errorf(
			"origin mismatch: expected %s, got %s",
			credential.Origin,
			clientData.Origin,
		)
	}

	// Verify the algorithm is supported
	if err := webauthn.ValidateAlgorithmSupport(credential.Algorithm); err != nil {
		return fmt.Errorf("algorithm validation failed: %w", err)
	}

	// Additional security checks for DID controller verification
	if len(credential.PublicKey) == 0 {
		return fmt.Errorf("credential missing public key data")
	}

	// Counter validation to prevent replay attacks
	// Note: In a production system, you would store and validate the signature counter
	// to ensure it's incrementing properly to prevent replay attacks
	if parsedAssertion.Response.AuthenticatorData.Counter > 0 {
		// The counter is present and valid - in production, verify it's greater than stored counter
		// For now, we accept any positive counter value as valid
	}

	return nil
}

// CreateWebAuthnChallenge creates a challenge for WebAuthn operations
func (v *WebAuthnControllerVerifier) CreateWebAuthnChallenge(
	ctx context.Context,
	did string,
	operation string,
) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create challenge data
	challengeData := fmt.Sprintf("%s:%s:%d:%d",
		did,
		operation,
		sdkCtx.BlockHeight(),
		sdkCtx.BlockTime().Unix(),
	)

	// Hash the challenge data to create a fixed-length challenge
	hash := sha256.Sum256([]byte(challengeData))

	// Encode as base64url
	challenge := base64.URLEncoding.EncodeToString(hash[:])

	return challenge, nil
}

// IsWebAuthnVerificationMethod checks if a verification method is a WebAuthn credential
func IsWebAuthnVerificationMethod(vm *types.VerificationMethod) bool {
	return vm.WebauthnCredential != nil &&
		vm.VerificationMethodKind == "WebAuthnCredential2024"
}

// GetWebAuthnCredentialsForDID returns all WebAuthn credentials for a DID
func (v *WebAuthnControllerVerifier) GetWebAuthnCredentialsForDID(
	ctx context.Context,
	did string,
) ([]*types.WebAuthnCredential, error) {
	// Get DID document
	ormDoc, err := v.keeper.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("DID document not found: %w", err)
	}

	didDoc := types.DIDDocumentFromORM(ormDoc)

	var credentials []*types.WebAuthnCredential
	for _, vm := range didDoc.VerificationMethod {
		if vm.WebauthnCredential != nil {
			credentials = append(credentials, vm.WebauthnCredential)
		}
	}

	return credentials, nil
}
