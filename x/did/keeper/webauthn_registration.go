package keeper

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	webauthn "github.com/sonr-io/common/webauthn"
	"github.com/sonr-io/common/webauthn/webauthncbor"
	"github.com/sonr-io/sonr/x/did/types"
)

// WebAuthnRegistrationData represents the data from a WebAuthn registration ceremony
type WebAuthnRegistrationData struct {
	CredentialID      string
	RawID             string
	ClientDataJSON    string
	AttestationObject string
	Username          string
	PublicKey         []byte
	Algorithm         int32
	Origin            string
}

// ProcessWebAuthnRegistration processes a WebAuthn credential and creates a DID document
func (k Keeper) ProcessWebAuthnRegistration(
	ctx context.Context,
	regData *WebAuthnRegistrationData,
) (*types.DIDDocument, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate a new DID
	did := k.generateDID(regData.Username)

	// Create WebAuthn credential with full attestation data
	webAuthnCredential := &types.WebAuthnCredential{
		CredentialId:      regData.CredentialID,
		RawId:             regData.RawID,
		ClientDataJson:    regData.ClientDataJSON,
		AttestationObject: regData.AttestationObject,
		PublicKey:         regData.PublicKey,
		Algorithm:         regData.Algorithm,
		AttestationType:   "none", // For most platform authenticators
		Origin:            regData.Origin,
		CreatedAt:         sdkCtx.BlockTime().Unix(),
	}

	// Validate the WebAuthn credential using local types validation
	if err := types.ValidateStructure(webAuthnCredential); err != nil {
		return nil, fmt.Errorf("WebAuthn credential validation failed: %w", err)
	}

	// Check for credential uniqueness to prevent replay attacks
	if k.HasExistingCredential(sdkCtx, regData.CredentialID) {
		return nil, fmt.Errorf("WebAuthn credential already exists: %s", regData.CredentialID)
	}

	// Create verification method with WebAuthn credential
	verificationMethod := &types.VerificationMethod{
		Id:                     fmt.Sprintf("%s#webauthn-1", did),
		Controller:             did,
		VerificationMethodKind: "WebAuthnCredential2024",
		WebauthnCredential:     webAuthnCredential,
	}

	// Create verification method references
	authRef := &types.VerificationMethodReference{
		VerificationMethodId: verificationMethod.Id,
	}
	assertRef := &types.VerificationMethodReference{
		VerificationMethodId: verificationMethod.Id,
	}
	capInvRef := &types.VerificationMethodReference{
		VerificationMethodId: verificationMethod.Id,
	}

	// Create DID document
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "", // Will be set to the cosmos address later
		VerificationMethod: []*types.VerificationMethod{
			verificationMethod,
		},
		Authentication: []*types.VerificationMethodReference{
			authRef,
		},
		AssertionMethod: []*types.VerificationMethodReference{
			assertRef,
		},
		KeyAgreement: []*types.VerificationMethodReference{},
		CapabilityInvocation: []*types.VerificationMethodReference{
			capInvRef,
		},
		CapabilityDelegation: []*types.VerificationMethodReference{},
		Service:              []*types.Service{},
	}

	// Store the DID document
	if err := k.storeDIDDocument(ctx, didDoc); err != nil {
		return nil, fmt.Errorf("failed to store DID document: %w", err)
	}

	return didDoc, nil
}

// CreateWebAuthnChallenge creates a challenge for WebAuthn registration
func (k Keeper) CreateWebAuthnChallenge(ctx context.Context, username string) (string, error) {
	// Generate cryptographically secure challenge
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return "", fmt.Errorf("failed to generate random challenge: %w", err)
	}

	challenge := base64.URLEncoding.EncodeToString(challengeBytes)

	// Store challenge with expiration (in production, use proper session storage)
	// For now, we'll rely on the server-side session management

	return challenge, nil
}

// VerifyWebAuthnRegistration verifies a WebAuthn registration response
func (k Keeper) VerifyWebAuthnRegistration(
	ctx context.Context,
	regData *WebAuthnRegistrationData,
	challenge string,
) error {
	// Decode and verify client data
	clientDataBytes, err := base64.URLEncoding.DecodeString(regData.ClientDataJSON)
	if err != nil {
		return fmt.Errorf("failed to decode client data JSON: %w", err)
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}

	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	// Verify type
	if clientData.Type != "webauthn.create" {
		return fmt.Errorf("invalid client data type: %s", clientData.Type)
	}

	// Verify challenge
	if clientData.Challenge != challenge {
		return fmt.Errorf("challenge mismatch")
	}

	// Verify origin (should be localhost for CLI usage)
	if clientData.Origin != "http://localhost" &&
		!k.isValidLocalhost(clientData.Origin) {
		return fmt.Errorf("invalid origin: %s", clientData.Origin)
	}

	// Parse attestation object and extract public key using CBOR
	publicKey, algorithm, err := k.extractPublicKeyFromAttestation(regData.AttestationObject)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	// Update registration data with extracted information
	regData.PublicKey = publicKey
	regData.Algorithm = algorithm
	regData.Origin = clientData.Origin

	return nil
}

// generateDID generates a new DID identifier
func (k Keeper) generateDID(username string) string {
	// For now, generate a simple DID based on username and timestamp
	// In production, this should be more sophisticated
	return fmt.Sprintf("did:sonr:%s-%d", username, time.Now().Unix())
}

// storeDIDDocument stores a DID document in the state
func (k Keeper) storeDIDDocument(ctx context.Context, didDoc *types.DIDDocument) error {
	// Convert to ORM format and store
	ormDoc := didDoc.ToORM()

	// Store in the ORM database
	if err := k.OrmDB.DIDDocumentTable().Insert(ctx, ormDoc); err != nil {
		return fmt.Errorf("failed to insert DID document: %w", err)
	}

	return nil
}

// isValidLocalhost checks if the origin is a valid localhost URL
func (k Keeper) isValidLocalhost(origin string) bool {
	validOrigins := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
		"http://localhost:8084",
		"http://localhost:8085",
		"http://localhost:8086",
		"http://localhost:8087",
		"http://localhost:8088",
		"http://localhost:8089",
	}

	return slices.Contains(validOrigins, origin)
}

// extractPublicKeyFromAttestation extracts the public key from WebAuthn attestation object
// Now leverages the full WebAuthn protocol implementation for proper CBOR parsing
func (k Keeper) extractPublicKeyFromAttestation(attestationObject string) ([]byte, int32, error) {
	// Use the local types validation to extract public key
	if err := types.ValidateAttestationObjectFormat(attestationObject); err != nil {
		return nil, 0, fmt.Errorf("invalid attestation object format: %w", err)
	}

	// Decode the attestation object using the full WebAuthn protocol
	attestationBytes, err := base64.RawURLEncoding.DecodeString(attestationObject)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode attestation object: %w", err)
	}

	// Parse the attestation object using CBOR
	var attestationObj webauthn.AttestationObject
	if err := webauthncbor.Unmarshal(attestationBytes, &attestationObj); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal attestation object: %w", err)
	}

	// Unmarshal the authenticator data
	if err := attestationObj.AuthData.Unmarshal(attestationObj.RawAuthData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal authenticator data: %w", err)
	}

	// Extract the attested credential data
	if !attestationObj.AuthData.Flags.HasAttestedCredentialData() {
		return nil, 0, fmt.Errorf("attestation object missing attested credential data")
	}

	publicKey := attestationObj.AuthData.AttData.CredentialPublicKey
	if len(publicKey) == 0 {
		return nil, 0, fmt.Errorf("no public key found in attested credential data")
	}

	// For now, assume ES256 algorithm. In the future, this could be extracted
	// from the COSE key format in the public key bytes
	algorithm := int32(-7) // ES256

	return publicKey, algorithm, nil
}

// GetWebAuthnCredentialsByDID retrieves all WebAuthn credentials for a DID
func (k Keeper) GetWebAuthnCredentialsByDID(
	ctx context.Context,
	did string,
) ([]*types.WebAuthnCredential, error) {
	// Get DID document
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		if err == collections.ErrNotFound {
			return nil, fmt.Errorf("DID document not found: %s", did)
		}
		return nil, fmt.Errorf("failed to get DID document: %w", err)
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

// ValidateWebAuthnCredential validates a WebAuthn credential exists and is valid
func (k Keeper) ValidateWebAuthnCredential(ctx context.Context, did, credentialID string) error {
	credentials, err := k.GetWebAuthnCredentialsByDID(ctx, did)
	if err != nil {
		return err
	}

	for _, cred := range credentials {
		if cred.CredentialId == credentialID {
			// Credential found and valid
			return nil
		}
	}

	return fmt.Errorf("WebAuthn credential %s not found for DID %s", credentialID, did)
}
