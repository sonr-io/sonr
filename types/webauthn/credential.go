package webauthn

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/sonr-io/sonr/types/webauthn/metadata"
)

// TrustAnchor represents a trusted root certificate or public key for attestation validation
type TrustAnchor struct {
	// Format is the attestation format this anchor applies to (e.g., "packed", "tpm", "android-safetynet")
	Format string
	// AAGUID is the Authenticator Attestation GUID this anchor applies to (optional)
	AAGUID []byte
	// RootCertificate is the trusted root certificate (for X.509 chains)
	RootCertificate []byte
	// PublicKey is the trusted public key (for ECDAA or direct attestation)
	PublicKey []byte
	// Description provides human-readable information about this trust anchor
	Description string
}

// CredentialPolicy defines the policy for accepting credentials
type CredentialPolicy struct {
	// AllowSelfAttestation permits self-attestation (not recommended for high-security scenarios)
	AllowSelfAttestation bool
	// RequireAttestation requires attestation to be present and valid
	RequireAttestation bool
	// TrustAnchors is the list of acceptable trust anchors for attestation validation
	TrustAnchors []TrustAnchor
	// AllowedAAGUIDs restricts registration to specific authenticator models (if empty, all are allowed)
	AllowedAAGUIDs [][]byte
	// MinimumAuthenticatorLevel specifies the minimum authenticator certification level (1-3, 0 = no requirement)
	MinimumAuthenticatorLevel int
}

// DefaultPolicy returns a sensible default policy for most applications
func DefaultPolicy() *CredentialPolicy {
	return &CredentialPolicy{
		AllowSelfAttestation:      true,  // Allow self-attestation for broad compatibility
		RequireAttestation:        false, // Don't require attestation by default
		TrustAnchors:              []TrustAnchor{},
		AllowedAAGUIDs:            [][]byte{},
		MinimumAuthenticatorLevel: 0, // No certification requirement by default
	}
}

// HighSecurityPolicy returns a strict policy for high-security applications
func HighSecurityPolicy() *CredentialPolicy {
	return &CredentialPolicy{
		AllowSelfAttestation:      false,           // Reject self-attestation
		RequireAttestation:        true,            // Require valid attestation
		TrustAnchors:              []TrustAnchor{}, // Must be populated with specific trust anchors
		AllowedAAGUIDs:            [][]byte{},      // Should be populated with known secure authenticators
		MinimumAuthenticatorLevel: 2,               // Require at least Level 2 certification
	}
}

// Credential is the basic credential type from the Credential Management specification that is inherited by WebAuthn's
// PublicKeyCredential type.
//
// Specification: Credential Management §2.2. The Credential Interface (https://www.w3.org/TR/credential-management/#credential)
type Credential struct {
	// ID is The credential’s identifier. The requirements for the
	// identifier are distinct for each type of credential. It might
	// represent a username for username/password tuples, for example.
	ID string `json:"id"`
	// Type is the value of the object’s interface object's [[type]] slot,
	// which specifies the credential type represented by this object.
	// This should be type "public-key" for Webauthn credentials.
	Type string `json:"type"`
}

// ParsedCredential is the parsed PublicKeyCredential interface, inherits from Credential, and contains
// the attributes that are returned to the caller when a new credential is created, or a new assertion is requested.
type ParsedCredential struct {
	ID   string `cbor:"id"`
	Type string `cbor:"type"`
}

type PublicKeyCredential struct {
	Credential

	RawID                   URLEncodedBase64                      `json:"rawId"`
	ClientExtensionResults  AuthenticationExtensionsClientOutputs `json:"clientExtensionResults,omitempty"`
	AuthenticatorAttachment string                                `json:"authenticatorAttachment,omitempty"`
}

type ParsedPublicKeyCredential struct {
	ParsedCredential

	RawID                   []byte                                `json:"rawId"`
	ClientExtensionResults  AuthenticationExtensionsClientOutputs `json:"clientExtensionResults,omitempty"`
	AuthenticatorAttachment AuthenticatorAttachment               `json:"authenticatorAttachment,omitempty"`
}

type CredentialCreationResponse struct {
	PublicKeyCredential

	AttestationResponse AuthenticatorAttestationResponse `json:"response"`
}

// Implement WebAuthnCredential interface for CredentialCreationResponse
// This allows the protocol's credential types to work with Sonr's centralized validation

func (ccr *CredentialCreationResponse) GetCredentialId() string {
	return ccr.ID
}

func (ccr *CredentialCreationResponse) GetPublicKey() []byte {
	// URLEncodedBase64 is already []byte, so we can return it directly
	return []byte(ccr.AttestationResponse.PublicKey)
}

func (ccr *CredentialCreationResponse) GetAlgorithm() int32 {
	return int32(ccr.AttestationResponse.PublicKeyAlgorithm)
}

func (ccr *CredentialCreationResponse) GetRawId() string {
	return string(ccr.RawID)
}

func (ccr *CredentialCreationResponse) GetClientDataJson() string {
	return string(ccr.AttestationResponse.ClientDataJSON)
}

func (ccr *CredentialCreationResponse) GetAttestationObject() string {
	return string(ccr.AttestationResponse.AttestationObject)
}

func (ccr *CredentialCreationResponse) GetOrigin() string {
	// Parse the origin from ClientDataJSON
	// The ClientDataJSON contains the origin field when parsed

	// Try to parse the credential to get the collected client data
	parsed, err := ccr.Parse()
	if err != nil {
		return ""
	}

	return parsed.Response.CollectedClientData.Origin
}

type ParsedCredentialCreationData struct {
	ParsedPublicKeyCredential

	Response ParsedAttestationResponse
	Raw      CredentialCreationResponse
}

// Implement WebAuthnCredential interface for ParsedCredentialCreationData
// This allows the parsed credential types to work with Sonr's centralized validation

func (pcc *ParsedCredentialCreationData) GetCredentialId() string {
	return pcc.ID
}

func (pcc *ParsedCredentialCreationData) GetPublicKey() []byte {
	// Get the public key from the parsed authenticator data
	if len(pcc.Response.AttestationObject.AuthData.AttData.CredentialPublicKey) > 0 {
		return pcc.Response.AttestationObject.AuthData.AttData.CredentialPublicKey
	}
	// Fallback to raw credential public key if available
	return []byte(pcc.Raw.AttestationResponse.PublicKey)
}

func (pcc *ParsedCredentialCreationData) GetAlgorithm() int32 {
	return int32(pcc.Raw.AttestationResponse.PublicKeyAlgorithm)
}

func (pcc *ParsedCredentialCreationData) GetRawId() string {
	return string(pcc.RawID)
}

func (pcc *ParsedCredentialCreationData) GetClientDataJson() string {
	return string(pcc.Raw.AttestationResponse.ClientDataJSON)
}

func (pcc *ParsedCredentialCreationData) GetAttestationObject() string {
	return string(pcc.Raw.AttestationResponse.AttestationObject)
}

func (pcc *ParsedCredentialCreationData) GetOrigin() string {
	return pcc.Response.CollectedClientData.Origin
}

// ParseCredentialCreationResponse is a non-agnostic function for parsing a registration response from the http library
// from stdlib. It handles some standard cleanup operations.
func ParseCredentialCreationResponse(request *http.Request) (*ParsedCredentialCreationData, error) {
	if request == nil || request.Body == nil {
		return nil, ErrBadRequest.WithDetails("No response given")
	}

	defer request.Body.Close()
	defer io.Copy(io.Discard, request.Body)

	return ParseCredentialCreationResponseBody(request.Body)
}

// ParseCredentialCreationResponseBody is an agnostic version of ParseCredentialCreationResponse. Implementers are
// therefore responsible for managing cleanup.
func ParseCredentialCreationResponseBody(
	body io.Reader,
) (pcc *ParsedCredentialCreationData, err error) {
	var ccr CredentialCreationResponse

	if err = decodeBody(body, &ccr); err != nil {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").
			WithInfo(err.Error()).
			WithError(err)
	}

	return ccr.Parse()
}

// ParseCredentialCreationResponseBytes is an alternative version of ParseCredentialCreationResponseBody that just takes
// a byte slice.
func ParseCredentialCreationResponseBytes(
	data []byte,
) (pcc *ParsedCredentialCreationData, err error) {
	var ccr CredentialCreationResponse

	if err = decodeBytes(data, &ccr); err != nil {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").
			WithInfo(err.Error()).
			WithError(err)
	}

	return ccr.Parse()
}

// Parse validates and parses the CredentialCreationResponse into a ParsedCredentialCreationData. This receiver
// is unlikely to be expressly guaranteed under the versioning policy. Users looking for this guarantee should see
// ParseCredentialCreationResponseBody instead, and this receiver should only be used if that function is inadequate
// for their use case.
func (ccr CredentialCreationResponse) Parse() (pcc *ParsedCredentialCreationData, err error) {
	if ccr.ID == "" {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").WithInfo("Missing ID")
	}

	testB64, err := base64.RawURLEncoding.DecodeString(ccr.ID)
	if err != nil || !(len(testB64) > 0) {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").
			WithInfo("ID not base64.RawURLEncoded")
	}

	if ccr.PublicKeyCredential.Credential.Type == "" {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").
			WithInfo("Missing type")
	}

	if ccr.PublicKeyCredential.Credential.Type != string(PublicKeyCredentialType) {
		return nil, ErrBadRequest.WithDetails("Parse error for Registration").
			WithInfo("Type not public-key")
	}

	response, err := ccr.AttestationResponse.Parse()
	if err != nil {
		return nil, ErrParsingData.WithDetails("Error parsing attestation response")
	}

	var attachment AuthenticatorAttachment

	switch ccr.AuthenticatorAttachment {
	case "platform":
		attachment = Platform
	case "cross-platform":
		attachment = CrossPlatform
	}

	return &ParsedCredentialCreationData{
		ParsedPublicKeyCredential{
			ParsedCredential{ccr.ID, ccr.Type}, ccr.RawID, ccr.ClientExtensionResults, attachment,
		},
		*response,
		ccr,
	}, nil
}

// Verify the Client and Attestation data.
//
// Specification: §7.1. Registering a New Credential (https://www.w3.org/TR/webauthn/#sctn-registering-a-new-credential)
func (pcc *ParsedCredentialCreationData) Verify(
	storedChallenge string,
	verifyUser bool,
	verifyUserPresence bool,
	relyingPartyID string,
	rpOrigins, rpTopOrigins []string,
	rpTopOriginsVerify TopOriginVerificationMode,
	mds metadata.Provider,
	credParams []CredentialParameter,
) (clientDataHash []byte, err error) {
	// Use default policy if none provided
	return pcc.VerifyWithPolicy(storedChallenge, verifyUser, verifyUserPresence, relyingPartyID,
		rpOrigins, rpTopOrigins, rpTopOriginsVerify, mds, credParams, DefaultPolicy())
}

// VerifyWithPolicy verifies the Client and Attestation data with a custom credential policy.
//
// Specification: §7.1. Registering a New Credential (https://www.w3.org/TR/webauthn/#sctn-registering-a-new-credential)
func (pcc *ParsedCredentialCreationData) VerifyWithPolicy(
	storedChallenge string,
	verifyUser bool,
	verifyUserPresence bool,
	relyingPartyID string,
	rpOrigins, rpTopOrigins []string,
	rpTopOriginsVerify TopOriginVerificationMode,
	mds metadata.Provider,
	credParams []CredentialParameter,
	policy *CredentialPolicy,
) (clientDataHash []byte, err error) {
	// Handles steps 3 through 6 - Verifying the Client Data against the Relying Party's stored data
	if err = pcc.Response.CollectedClientData.Verify(storedChallenge, CreateCeremony, rpOrigins, rpTopOrigins, rpTopOriginsVerify); err != nil {
		return nil, err
	}

	// Step 7. Compute the hash of response.clientDataJSON using SHA-256.
	sum := sha256.Sum256(pcc.Raw.AttestationResponse.ClientDataJSON)
	clientDataHash = sum[:]

	// Step 8. Perform CBOR decoding on the attestationObject field of the AuthenticatorAttestationResponse
	// structure to obtain the attestation statement format fmt, the authenticator data authData, and the
	// attestation statement attStmt.

	// We do the above step while parsing and decoding the CredentialCreationResponse
	// Handle steps 9 through 14 - This verifies the attestation object.
	if err = pcc.Response.AttestationObject.Verify(relyingPartyID, clientDataHash, verifyUser, verifyUserPresence, mds, credParams); err != nil {
		return clientDataHash, err
	}

	// Step 15. If validation is successful, obtain a list of acceptable trust anchors (attestation root
	// certificates or ECDAA-Issuer public keys) for that attestation type and attestation statement
	// format fmt, from a trusted source or from policy. For example, the FIDO Metadata Service provides
	// one way to obtain such information, using the AAGUID in the attestedCredentialData in authData.
	// [https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-metadata-service-v2.0-id-20180227.html]

	// Apply policy validation if provided
	if policy != nil {
		// Check if attestation is required
		if policy.RequireAttestation && pcc.Response.AttestationObject.Format == "none" {
			return clientDataHash, ErrAttestationFormat.WithDetails(
				"Attestation required by policy but not provided",
			)
		}

		// Check AAGUID restrictions
		if len(policy.AllowedAAGUIDs) > 0 {
			aaguid := pcc.Response.AttestationObject.AuthData.AttData.AAGUID
			allowed := false
			for _, allowedAAGUID := range policy.AllowedAAGUIDs {
				if string(aaguid) == string(allowedAAGUID) {
					allowed = true
					break
				}
			}
			if !allowed {
				return clientDataHash, ErrAttestationFormat.WithDetails(
					"Authenticator AAGUID not in allowed list",
				)
			}
		}
	}

	// Step 16. Assess the attestation trustworthiness using outputs of the verification procedure in step 14, as follows:
	// - If self attestation was used, check if self attestation is acceptable under Relying Party policy.
	// - If ECDAA was used, verify that the identifier of the ECDAA-Issuer public key used is included in
	//   the set of acceptable trust anchors obtained in step 15.
	// - Otherwise, use the X.509 certificates returned by the verification procedure to verify that the
	//   attestation public key correctly chains up to an acceptable root certificate.

	// Check self-attestation policy
	if policy != nil && !policy.AllowSelfAttestation {
		// Check if this is self-attestation (format is "packed" with no x5c chain)
		if pcc.Response.AttestationObject.Format == "packed" {
			// Self-attestation in packed format has no x5c certificate chain
			if _, hasX5C := pcc.Response.AttestationObject.AttStatement["x5c"]; !hasX5C {
				return clientDataHash, ErrAttestationFormat.WithDetails(
					"Self-attestation not allowed by policy",
				)
			}
		}
	}

	// Step 17. Check that the credentialId is not yet registered to any other user. If registration is
	// requested for a credential that is already registered to a different user, the Relying Party SHOULD
	// fail this registration ceremony, or it MAY decide to accept the registration, e.g. while deleting
	// the older registration.

	// Note: The Relying Party must check for duplicate credential IDs against their database.
	// This cannot be enforced at the library level.

	// Step 18 If the attestation statement attStmt verified successfully and is found to be trustworthy, then
	// register the new credential with the account that was denoted in the options.user passed to create(), by
	// associating it with the credentialId and credentialPublicKey in the attestedCredentialData in authData, as
	// appropriate for the Relying Party's system.

	// Step 19. If the attestation statement attStmt successfully verified but is not trustworthy per step 16 above,
	// the Relying Party SHOULD fail the registration ceremony.

	// Policy validation has been implemented above to handle trust assessment

	return clientDataHash, nil
}

// GetAppID takes a AuthenticationExtensions object or nil. It then performs the following checks in order:
//
// 1. Check that the Session Data's AuthenticationExtensions has been provided and if it hasn't return an error.
// 2. Check that the AuthenticationExtensionsClientOutputs contains the extensions output and return an empty string if it doesn't.
// 3. Check that the Credential AttestationType is `fido-u2f` and return an empty string if it isn't.
// 4. Check that the AuthenticationExtensionsClientOutputs contains the appid key and if it doesn't return an empty string.
// 5. Check that the AuthenticationExtensionsClientOutputs appid is a bool and if it isn't return an error.
// 6. Check that the appid output is true and if it isn't return an empty string.
// 7. Check that the Session Data has an appid extension defined and if it doesn't return an error.
// 8. Check that the appid extension in Session Data is a string and if it isn't return an error.
// 9. Return the appid extension value from the Session data.
func (ppkc ParsedPublicKeyCredential) GetAppID(
	authExt AuthenticationExtensions,
	credentialAttestationType string,
) (appID string, err error) {
	var (
		value, clientValue any
		enableAppID, ok    bool
	)

	if authExt == nil {
		return "", nil
	}

	if ppkc.ClientExtensionResults == nil {
		return "", nil
	}

	// If the credential does not have the correct attestation type it is assumed to NOT be a fido-u2f credential.
	// https://www.w3.org/TR/webauthn/#sctn-fido-u2f-attestation
	if credentialAttestationType != CredentialTypeFIDOU2F {
		return "", nil
	}

	if clientValue, ok = ppkc.ClientExtensionResults[ExtensionAppID]; !ok {
		return "", nil
	}

	if enableAppID, ok = clientValue.(bool); !ok {
		return "", ErrBadRequest.WithDetails("Client Output appid did not have the expected type")
	}

	if !enableAppID {
		return "", nil
	}

	if value, ok = authExt[ExtensionAppID]; !ok {
		return "", ErrBadRequest.WithDetails(
			"Session Data does not have an appid but Client Output indicates it should be set",
		)
	}

	if appID, ok = value.(string); !ok {
		return "", ErrBadRequest.WithDetails("Session Data appid did not have the expected type")
	}

	return appID, nil
}

const (
	CredentialTypeFIDOU2F = "fido-u2f"
)
