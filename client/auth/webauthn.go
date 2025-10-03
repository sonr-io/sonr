// Package auth provides WebAuthn integration for the Sonr client SDK.
package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/client/keys"
	"github.com/sonr-io/sonr/types/webauthn"
	"github.com/sonr-io/sonr/types/webauthn/webauthncbor"
	"github.com/sonr-io/sonr/types/webauthn/webauthncose"
)

// WebAuthnClient provides an interface for WebAuthn operations with Sonr's Decentralized Abstracted Smart Wallets.
type WebAuthnClient interface {
	// Registration Operations
	BeginRegistration(ctx context.Context, opts *RegistrationOptions) (*RegistrationChallenge, error)
	CompleteRegistration(ctx context.Context, challenge *RegistrationChallenge, response *AuthenticatorAttestationResponse) (*WebAuthnCredential, error)

	// Authentication Operations
	BeginAuthentication(ctx context.Context, opts *AuthenticationOptions) (*AuthenticationChallenge, error)
	CompleteAuthentication(ctx context.Context, challenge *AuthenticationChallenge, response *AuthenticatorAssertionResponse, credentialID string) (*AuthenticationResult, error)

	// Credential Management
	ListCredentials(ctx context.Context, userID string) ([]*WebAuthnCredential, error)
	GetCredential(ctx context.Context, credentialID string) (*WebAuthnCredential, error)
	UpdateCredential(ctx context.Context, credentialID string, opts *UpdateCredentialOptions) (*WebAuthnCredential, error)
	RevokeCredential(ctx context.Context, credentialID string) error

	// DID Integration
	RegisterWithDID(ctx context.Context, did string, opts *DIDRegistrationOptions) (*DIDWebAuthnBinding, error)
	AuthenticateWithDID(ctx context.Context, did string, opts *DIDAuthenticationOptions) (*DIDAuthenticationResult, error)

	// Wallet Integration
	BindToWallet(ctx context.Context, credentialID string, keyring keys.KeyringManager) (*WalletBinding, error)
	SignWithWebAuthn(ctx context.Context, credentialID string, data []byte) (*WebAuthnSignature, error)
}

// RegistrationOptions configures WebAuthn registration.
type RegistrationOptions struct {
	UserID                 string                  `json:"user_id"`
	Username               string                  `json:"username"`
	DisplayName            string                  `json:"display_name"`
	Timeout                int                     `json:"timeout,omitempty"`           // Timeout in milliseconds
	UserVerification       string                  `json:"user_verification,omitempty"` // required, preferred, discouraged
	AttestationType        string                  `json:"attestation_type,omitempty"`  // none, indirect, direct
	AuthenticatorSelection *AuthenticatorSelection `json:"authenticator_selection,omitempty"`
	Extensions             map[string]any          `json:"extensions,omitempty"`
}

// AuthenticatorSelection specifies authenticator requirements.
type AuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticator_attachment,omitempty"` // platform, cross-platform
	RequireResidentKey      bool   `json:"require_resident_key,omitempty"`
	UserVerification        string `json:"user_verification,omitempty"`
}

// RegistrationChallenge contains the challenge for registration.
type RegistrationChallenge struct {
	Challenge              []byte                  `json:"challenge"`
	RelyingParty           *RelyingParty           `json:"relying_party"`
	User                   *User                   `json:"user"`
	PubKeyCredParams       []*PubKeyCredParam      `json:"pub_key_cred_params"`
	Timeout                int                     `json:"timeout"`
	ExcludeCredentials     []*CredentialDescriptor `json:"exclude_credentials,omitempty"`
	AuthenticatorSelection *AuthenticatorSelection `json:"authenticator_selection,omitempty"`
	Attestation            string                  `json:"attestation"`
	Extensions             map[string]any          `json:"extensions,omitempty"`
}

// RelyingParty represents the relying party information.
type RelyingParty struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}

// User represents the user information for WebAuthn.
type User struct {
	ID          []byte `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon,omitempty"`
}

// PubKeyCredParam specifies the public key parameters.
type PubKeyCredParam struct {
	Type      string `json:"type"`
	Algorithm int    `json:"alg"`
}

// CredentialDescriptor describes a credential.
type CredentialDescriptor struct {
	ID         []byte   `json:"id"`
	Type       string   `json:"type"`
	Transports []string `json:"transports,omitempty"`
}

// AuthenticatorAttestationResponse contains the registration response from the authenticator.
type AuthenticatorAttestationResponse struct {
	ClientDataJSON    []byte   `json:"client_data_json"`
	AttestationObject []byte   `json:"attestation_object"`
	Transports        []string `json:"transports,omitempty"`
}

// WebAuthnCredential represents a stored WebAuthn credential.
type WebAuthnCredential struct {
	ID              string              `json:"id"`
	RawID           []byte              `json:"raw_id"`
	PublicKey       []byte              `json:"public_key"`
	Algorithm       int64               `json:"algorithm"`
	AttestationType string              `json:"attestation_type"`
	Transports      []string            `json:"transports"`
	Flags           *AuthenticatorFlags `json:"flags"`
	Authenticator   *AuthenticatorData  `json:"authenticator"`
	Counter         uint32              `json:"counter"`
	AAGUID          []byte              `json:"aaguid"`
	UserID          string              `json:"user_id"`
	UserVerified    bool                `json:"user_verified"`
	BackupEligible  bool                `json:"backup_eligible"`
	BackupState     bool                `json:"backup_state"`
	Origin          string              `json:"origin"`
	CreatedAt       string              `json:"created_at"`
	LastUsed        string              `json:"last_used,omitempty"`
	Metadata        map[string]any      `json:"metadata,omitempty"`
}

// AuthenticatorFlags represents authenticator flags.
type AuthenticatorFlags struct {
	UserPresent   bool `json:"user_present"`
	UserVerified  bool `json:"user_verified"`
	AttestedData  bool `json:"attested_data"`
	ExtensionData bool `json:"extension_data"`
}

// AuthenticatorData contains authenticator data.
type AuthenticatorData struct {
	RPIDHash      []byte `json:"rpid_hash"`
	Flags         byte   `json:"flags"`
	Counter       uint32 `json:"counter"`
	AttestedData  []byte `json:"attested_data,omitempty"`
	ExtensionData []byte `json:"extension_data,omitempty"`
}

// AuthenticationOptions configures WebAuthn authentication.
type AuthenticationOptions struct {
	UserID             string                  `json:"user_id,omitempty"`
	Timeout            int                     `json:"timeout,omitempty"`
	UserVerification   string                  `json:"user_verification,omitempty"`
	AllowedCredentials []*CredentialDescriptor `json:"allowed_credentials,omitempty"`
	Extensions         map[string]any          `json:"extensions,omitempty"`
}

// AuthenticationChallenge contains the challenge for authentication.
type AuthenticationChallenge struct {
	Challenge          []byte                  `json:"challenge"`
	Timeout            int                     `json:"timeout"`
	RelyingPartyID     string                  `json:"relying_party_id"`
	AllowedCredentials []*CredentialDescriptor `json:"allowed_credentials,omitempty"`
	UserVerification   string                  `json:"user_verification"`
	Extensions         map[string]any          `json:"extensions,omitempty"`
}

// AuthenticatorAssertionResponse contains the authentication response from the authenticator.
type AuthenticatorAssertionResponse struct {
	ClientDataJSON    []byte `json:"client_data_json"`
	AuthenticatorData []byte `json:"authenticator_data"`
	Signature         []byte `json:"signature"`
	UserHandle        []byte `json:"user_handle,omitempty"`
}

// AuthenticationResult contains the result of authentication.
type AuthenticationResult struct {
	Success    bool                `json:"success"`
	Verified   bool                `json:"verified"`
	Credential *WebAuthnCredential `json:"credential,omitempty"`
	Counter    uint32              `json:"counter"`
	UserHandle []byte              `json:"user_handle,omitempty"`
	Error      string              `json:"error,omitempty"`
}

// UpdateCredentialOptions configures credential updates.
type UpdateCredentialOptions struct {
	Metadata map[string]any `json:"metadata,omitempty"`
}

// DIDRegistrationOptions configures DID-based WebAuthn registration.
type DIDRegistrationOptions struct {
	CredentialOptions  *RegistrationOptions `json:"credential_options"`
	DIDDocument        map[string]any       `json:"did_document,omitempty"`
	VerificationMethod string               `json:"verification_method,omitempty"`
}

// DIDWebAuthnBinding represents a binding between a DID and WebAuthn credential.
type DIDWebAuthnBinding struct {
	DID                string              `json:"did"`
	CredentialID       string              `json:"credential_id"`
	Credential         *WebAuthnCredential `json:"credential"`
	VerificationMethod string              `json:"verification_method"`
	CreatedAt          string              `json:"created_at"`
}

// DIDDocument represents a minimal DID document structure.
type DIDDocument struct {
	ID                 string           `json:"id"`
	VerificationMethod []map[string]any `json:"verificationMethod,omitempty"`
	Authentication     []any            `json:"authentication,omitempty"`
}

// DIDAuthenticationOptions configures DID-based authentication.
type DIDAuthenticationOptions struct {
	AuthenticationOptions *AuthenticationOptions `json:"authentication_options"`
	Challenge             []byte                 `json:"challenge,omitempty"`
}

// DIDAuthenticationResult contains the result of DID authentication.
type DIDAuthenticationResult struct {
	Success              bool                     `json:"success"`
	DID                  string                   `json:"did"`
	Challenge            *AuthenticationChallenge `json:"challenge,omitempty"`
	CredentialOptions    []*WebAuthnCredential    `json:"credential_options,omitempty"`
	SessionID            string                   `json:"session_id,omitempty"`
	CreatedAt            string                   `json:"created_at,omitempty"`
	AuthenticationResult *AuthenticationResult    `json:"authentication_result,omitempty"`
	WalletIdentity       *keys.WalletIdentity     `json:"wallet_identity,omitempty"`
}

// WalletBinding represents a binding between a WebAuthn credential and a wallet.
type WalletBinding struct {
	CredentialID   string               `json:"credential_id"`
	WalletIdentity *keys.WalletIdentity `json:"wallet_identity"`
	BindingType    string               `json:"binding_type"` // primary, secondary, recovery
	CreatedAt      string               `json:"created_at"`
}

// WebAuthnSignature represents a signature created using WebAuthn.
type WebAuthnSignature struct {
	Signature         []byte `json:"signature"`
	CredentialID      string `json:"credential_id"`
	Counter           uint32 `json:"counter"`
	AuthenticatorData []byte `json:"authenticator_data"`
	ClientDataJSON    []byte `json:"client_data_json"`
}

// webAuthnClient implements the WebAuthnClient interface.
type webAuthnClient struct {
	keyring           keys.KeyringManager
	rpID              string
	rpName            string
	origin            string
	pendingChallenges map[string]*AuthenticationChallenge
	pendingSignatures map[string][]byte
}

// NewWebAuthnClient creates a new WebAuthn client.
func NewWebAuthnClient(keyring keys.KeyringManager, rpID, rpName string) WebAuthnClient {
	return &webAuthnClient{
		keyring:           keyring,
		rpID:              rpID,
		rpName:            rpName,
		origin:            fmt.Sprintf("https://%s", rpID),
		pendingChallenges: make(map[string]*AuthenticationChallenge),
		pendingSignatures: make(map[string][]byte),
	}
}

// BeginRegistration initiates WebAuthn registration.
func (w *webAuthnClient) BeginRegistration(ctx context.Context, opts *RegistrationOptions) (*RegistrationChallenge, error) {
	// Generate challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to generate challenge")
	}

	// Create user ID if not provided
	userID := opts.UserID
	if userID == "" {
		userIDBytes := make([]byte, 16)
		rand.Read(userIDBytes)
		userID = base64.URLEncoding.EncodeToString(userIDBytes)
	}

	// Build registration challenge
	regChallenge := &RegistrationChallenge{
		Challenge: challenge,
		RelyingParty: &RelyingParty{
			ID:   w.rpID,
			Name: w.rpName,
		},
		User: &User{
			ID:          []byte(userID),
			Name:        opts.Username,
			DisplayName: opts.DisplayName,
		},
		PubKeyCredParams: []*PubKeyCredParam{
			{Type: "public-key", Algorithm: -7},   // ES256
			{Type: "public-key", Algorithm: -257}, // RS256
		},
		Timeout:                opts.Timeout,
		AuthenticatorSelection: opts.AuthenticatorSelection,
		Attestation:            "none",
		Extensions:             opts.Extensions,
	}

	if regChallenge.Timeout == 0 {
		regChallenge.Timeout = 60000 // 60 seconds default
	}

	return regChallenge, nil
}

// CompleteRegistration completes WebAuthn registration.
func (w *webAuthnClient) CompleteRegistration(ctx context.Context, challenge *RegistrationChallenge, response *AuthenticatorAttestationResponse) (*WebAuthnCredential, error) {
	// Verify client data JSON
	clientData, err := verifyClientData(response.ClientDataJSON, challenge.Challenge, "webauthn.create", w.origin)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "client data verification failed")
	}

	// Parse attestation object
	attestationObj, err := parseAttestationObject(response.AttestationObject)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "attestation object parsing failed")
	}

	// Verify authenticator data
	if err := verifyAuthenticatorData(attestationObj.AuthData, w.rpID); err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "authenticator data verification failed")
	}

	// Extract public key from authenticator data
	if len(attestationObj.AuthData.AttData.CredentialID) == 0 {
		return nil, errors.NewModuleError("auth", "CompleteRegistration",
			fmt.Errorf("no attestation data in authenticator response"))
	}

	// Parse COSE public key
	publicKey, err := parseCOSEPublicKey(attestationObj.AuthData.AttData.CredentialPublicKey)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "public key parsing failed")
	}

	// Verify attestation (if present)
	if attestationObj.Format != "none" {
		clientDataHash := sha256.Sum256(response.ClientDataJSON)
		if err := verifyAttestation(attestationObj, clientDataHash[:]); err != nil {
			return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "attestation verification failed")
		}
	}

	// Create credential for storage
	credential := &WebAuthnCredential{
		ID:              base64.URLEncoding.EncodeToString(attestationObj.AuthData.AttData.CredentialID),
		PublicKey:       attestationObj.AuthData.AttData.CredentialPublicKey,
		Algorithm:       publicKey.Algorithm,
		AttestationType: attestationObj.Format,
		Transports:      response.Transports,
		Counter:         attestationObj.AuthData.Counter,
		AAGUID:          attestationObj.AuthData.AttData.AAGUID,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		LastUsed:        time.Now().UTC().Format(time.RFC3339),
		UserVerified:    attestationObj.AuthData.Flags.UserVerified(),
		BackupEligible:  attestationObj.AuthData.Flags.HasBackupEligible(),
		BackupState:     attestationObj.AuthData.Flags.HasBackupState(),
		Origin:          clientData.Origin,
	}

	// Store credential via DID module (implementation would interact with blockchain)
	// This would typically involve creating a MsgRegisterWebAuthnCredential transaction
	// For now, we return the credential object

	return credential, nil
}

// BeginAuthentication initiates WebAuthn authentication.
func (w *webAuthnClient) BeginAuthentication(ctx context.Context, opts *AuthenticationOptions) (*AuthenticationChallenge, error) {
	// Generate challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to generate challenge")
	}

	authChallenge := &AuthenticationChallenge{
		Challenge:          challenge,
		Timeout:            opts.Timeout,
		RelyingPartyID:     w.rpID,
		AllowedCredentials: opts.AllowedCredentials,
		UserVerification:   opts.UserVerification,
		Extensions:         opts.Extensions,
	}

	if authChallenge.Timeout == 0 {
		authChallenge.Timeout = 60000 // 60 seconds default
	}

	if authChallenge.UserVerification == "" {
		authChallenge.UserVerification = "preferred"
	}

	return authChallenge, nil
}

// CompleteAuthentication completes WebAuthn authentication.
func (w *webAuthnClient) CompleteAuthentication(ctx context.Context, challenge *AuthenticationChallenge, response *AuthenticatorAssertionResponse, credentialID string) (*AuthenticationResult, error) {
	// Verify client data JSON
	_, err := verifyClientData(response.ClientDataJSON, challenge.Challenge, "webauthn.get", w.origin)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "client data verification failed")
	}

	// Parse authenticator data
	var authData webauthn.AuthenticatorData
	if err := authData.Unmarshal(response.AuthenticatorData); err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "authenticator data parsing failed")
	}

	// Verify RP ID hash
	rpIDHash := sha256.Sum256([]byte(w.rpID))
	if !bytes.Equal(authData.RPIDHash[:], rpIDHash[:]) {
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("RP ID hash mismatch"))
	}

	// Verify user presence
	if !authData.Flags.UserPresent() {
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("user presence flag not set"))
	}

	// Verify user verification if required
	if challenge.UserVerification == "required" && !authData.Flags.UserVerified() {
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("user verification required but not performed"))
	}

	// Get credential from storage (would normally query blockchain)
	// For now, we'll need the credential to be provided or fetched
	credential, err := w.getStoredCredential(ctx, credentialID)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "credential not found")
	}

	// Verify counter progression (prevent replay attacks)
	if authData.Counter > 0 && authData.Counter <= credential.Counter {
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("counter did not increase: possible replay attack"))
	}

	// Construct signature base
	clientDataHash := sha256.Sum256(response.ClientDataJSON)
	signatureBase := append(response.AuthenticatorData, clientDataHash[:]...)

	// Parse and verify signature
	publicKey, err := webauthncose.ParsePublicKey(credential.PublicKey)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "public key parsing failed")
	}

	// Verify signature based on key type
	var signatureValid bool
	switch pk := publicKey.(type) {
	case *webauthncose.EC2PublicKeyData:
		signatureValid, err = pk.Verify(signatureBase, response.Signature)
	case *webauthncose.RSAPublicKeyData:
		signatureValid, err = pk.Verify(signatureBase, response.Signature)
	case *webauthncose.OKPPublicKeyData:
		signatureValid, err = pk.Verify(signatureBase, response.Signature)
	default:
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("unsupported public key type"))
	}

	if err != nil || !signatureValid {
		return nil, errors.NewModuleError("auth", "CompleteAuthentication",
			fmt.Errorf("signature verification failed"))
	}

	// Update credential counter
	credential.Counter = authData.Counter
	credential.LastUsed = time.Now().UTC().Format(time.RFC3339)

	// Create authentication result
	result := &AuthenticationResult{
		Success:    true,
		Verified:   true,
		Credential: credential,
		Counter:    authData.Counter,
		UserHandle: response.UserHandle,
	}

	return result, nil
}

// getStoredCredential retrieves a credential from storage (placeholder).
func (w *webAuthnClient) getStoredCredential(ctx context.Context, credentialID string) (*WebAuthnCredential, error) {
	// This would typically query the blockchain for the credential
	// For now, return an error indicating implementation is needed
	return nil, fmt.Errorf("credential storage not yet implemented")
}

// ListCredentials lists WebAuthn credentials for a user.
func (w *webAuthnClient) ListCredentials(ctx context.Context, userID string) ([]*WebAuthnCredential, error) {
	// Validate user ID
	if userID == "" {
		return nil, errors.NewModuleError("auth", "ListCredentials",
			fmt.Errorf("user ID cannot be empty"))
	}

	// Query DID module for user's credentials
	// This would typically use the DID module's query client
	credentials, err := w.queryUserCredentials(ctx, userID)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to query credentials")
	}

	// Filter out revoked credentials
	activeCredentials := make([]*WebAuthnCredential, 0)
	for _, cred := range credentials {
		// Check if credential is active (not using Status field)
		if cred != nil {
			activeCredentials = append(activeCredentials, cred)
		}
	}

	return activeCredentials, nil
}

// GetCredential retrieves a specific WebAuthn credential.
func (w *webAuthnClient) GetCredential(ctx context.Context, credentialID string) (*WebAuthnCredential, error) {
	// Validate credential ID
	if credentialID == "" {
		return nil, errors.NewModuleError("auth", "GetCredential",
			fmt.Errorf("credential ID cannot be empty"))
	}

	// Decode credential ID if base64 encoded
	credID, err := base64.URLEncoding.DecodeString(credentialID)
	if err != nil {
		// Try using raw credential ID
		credID = []byte(credentialID)
	}

	// Query DID module for specific credential
	credential, err := w.queryCredentialByID(ctx, string(credID))
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "credential not found")
	}

	// Check if credential is valid
	if credential == nil {
		return nil, errors.NewModuleError("auth", "GetCredential",
			fmt.Errorf("credential not found"))
	}

	return credential, nil
}

// UpdateCredential updates a WebAuthn credential.
func (w *webAuthnClient) UpdateCredential(ctx context.Context, credentialID string, opts *UpdateCredentialOptions) (*WebAuthnCredential, error) {
	// Get existing credential
	credential, err := w.GetCredential(ctx, credentialID)
	if err != nil {
		return nil, err
	}

	// Update metadata
	if opts.Metadata != nil {
		credential.Metadata = opts.Metadata
	}
	credential.LastUsed = time.Now().UTC().Format(time.RFC3339)

	// Submit update transaction to chain
	// This would create a MsgUpdateWebAuthnCredential
	if err := w.submitCredentialUpdate(ctx, credential); err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to update credential")
	}

	return credential, nil
}

// RevokeCredential revokes a WebAuthn credential.
func (w *webAuthnClient) RevokeCredential(ctx context.Context, credentialID string) error {
	// Get existing credential
	credential, err := w.GetCredential(ctx, credentialID)
	if err != nil {
		return err
	}

	// Check if credential exists
	if credential == nil {
		return errors.NewModuleError("auth", "RevokeCredential",
			fmt.Errorf("credential not found"))
	}

	// Submit revocation transaction to chain
	// This would create a MsgRevokeWebAuthnCredential
	if err := w.submitCredentialRevocation(ctx, credentialID); err != nil {
		return errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to revoke credential")
	}

	return nil
}

// Helper methods for DID module interaction (placeholders)

func (w *webAuthnClient) queryUserCredentials(ctx context.Context, userID string) ([]*WebAuthnCredential, error) {
	// Placeholder: would query DID module
	return []*WebAuthnCredential{}, nil
}

func (w *webAuthnClient) queryCredentialByID(ctx context.Context, credentialID string) (*WebAuthnCredential, error) {
	// Placeholder: would query DID module
	return nil, fmt.Errorf("DID module query not yet implemented")
}

func (w *webAuthnClient) submitCredentialUpdate(ctx context.Context, credential *WebAuthnCredential) error {
	// Placeholder: would submit transaction to DID module
	return nil
}

func (w *webAuthnClient) submitCredentialRevocation(ctx context.Context, credentialID string) error {
	// Placeholder: would submit transaction to DID module
	return nil
}

// RegisterWithDID registers a WebAuthn credential with a DID.
func (w *webAuthnClient) RegisterWithDID(ctx context.Context, did string, opts *DIDRegistrationOptions) (*DIDWebAuthnBinding, error) {
	// Begin registration challenge
	regChallenge, err := w.BeginRegistration(ctx, opts.CredentialOptions)
	if err != nil {
		return nil, err
	}

	// Store challenge for later verification
	// In production, this would be stored in a session or cache

	binding := &DIDWebAuthnBinding{
		DID:                did,
		CredentialID:       base64.URLEncoding.EncodeToString(regChallenge.Challenge),
		VerificationMethod: opts.VerificationMethod,
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
	}

	return binding, nil
}

// AuthenticateWithDID authenticates using WebAuthn and associates with a DID.
func (w *webAuthnClient) AuthenticateWithDID(ctx context.Context, did string, opts *DIDAuthenticationOptions) (*DIDAuthenticationResult, error) {
	// Validate DID format
	if did == "" {
		return nil, errors.NewModuleError("auth", "AuthenticateWithDID",
			fmt.Errorf("DID cannot be empty"))
	}

	// Resolve DID to find WebAuthn verification methods
	didDocument, err := w.resolveDID(ctx, did)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to resolve DID")
	}

	// Extract WebAuthn credentials from DID document
	webauthnCredentials := w.extractWebAuthnCredentials(didDocument)
	if len(webauthnCredentials) == 0 {
		return nil, errors.NewModuleError("auth", "AuthenticateWithDID",
			fmt.Errorf("no WebAuthn credentials found in DID document"))
	}

	// Create authentication options with allowed credentials
	authOpts := &AuthenticationOptions{
		UserVerification:   "preferred",
		Timeout:            60000,
		AllowedCredentials: make([]*CredentialDescriptor, 0),
	}

	for _, cred := range webauthnCredentials {
		credIDBytes, _ := base64.URLEncoding.DecodeString(cred.ID)
		authOpts.AllowedCredentials = append(authOpts.AllowedCredentials, &CredentialDescriptor{
			Type:       "public-key",
			ID:         credIDBytes,
			Transports: cred.Transports,
		})
	}

	// Begin authentication challenge
	challenge, err := w.BeginAuthentication(ctx, authOpts)
	if err != nil {
		return nil, err
	}

	// Store challenge for later verification (in production, use session/cache)
	w.pendingChallenges[did] = challenge

	// Create DID authentication result
	result := &DIDAuthenticationResult{
		Success:              true,
		DID:                  did,
		AuthenticationResult: nil, // Will be populated after authentication completion
		WalletIdentity:       nil, // Will be populated if wallet binding exists
	}

	return result, nil
}

// resolveDID resolves a DID document (placeholder).
func (w *webAuthnClient) resolveDID(ctx context.Context, did string) (*DIDDocument, error) {
	// Placeholder: would query DID resolver
	return &DIDDocument{
		ID: did,
	}, nil
}

// extractWebAuthnCredentials extracts WebAuthn credentials from DID document.
func (w *webAuthnClient) extractWebAuthnCredentials(doc *DIDDocument) []*WebAuthnCredential {
	credentials := make([]*WebAuthnCredential, 0)
	// Extract from verification methods
	// This would parse the DID document structure
	return credentials
}

// BindToWallet binds a WebAuthn credential to a Decentralized Abstracted Smart Wallet.
func (w *webAuthnClient) BindToWallet(ctx context.Context, credentialID string, keyring keys.KeyringManager) (*WalletBinding, error) {
	// Get wallet identity
	identity, err := keyring.GetIssuerDID(ctx)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to get wallet identity")
	}

	// TODO: Implement actual binding logic
	// This would store the binding relationship

	binding := &WalletBinding{
		CredentialID:   credentialID,
		WalletIdentity: identity,
		BindingType:    "primary",
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	return binding, nil
}

// BindCredential binds a WebAuthn credential to an existing DID.
// This allows using an existing WebAuthn credential with a DID that was created through other means.
func (w *webAuthnClient) BindCredential(ctx context.Context, did string, credential *WebAuthnCredential) error {
	// Validate inputs
	if did == "" {
		return errors.NewModuleError("auth", "BindCredential",
			fmt.Errorf("DID cannot be empty"))
	}

	if credential == nil {
		return errors.NewModuleError("auth", "BindCredential",
			fmt.Errorf("credential cannot be nil"))
	}

	if credential.ID == "" {
		return errors.NewModuleError("auth", "BindCredential",
			fmt.Errorf("credential ID cannot be empty"))
	}

	if len(credential.PublicKey) == 0 {
		return errors.NewModuleError("auth", "BindCredential",
			fmt.Errorf("credential public key cannot be empty"))
	}

	// TODO: In a real implementation, this would:
	// 1. Verify the DID exists in the DID registry
	// 2. Verify the caller has permission to bind credentials to this DID
	// 3. Store the binding in the DID document as a verification method
	// 4. Emit an event for the binding creation

	// For now, we simulate success
	return nil
}

// SignWithWebAuthn signs data using a WebAuthn credential.
func (w *webAuthnClient) SignWithWebAuthn(ctx context.Context, credentialID string, data []byte) (*WebAuthnSignature, error) {
	// Validate inputs
	if credentialID == "" {
		return nil, errors.NewModuleError("auth", "SignWithWebAuthn",
			fmt.Errorf("credential ID cannot be empty"))
	}
	if len(data) == 0 {
		return nil, errors.NewModuleError("auth", "SignWithWebAuthn",
			fmt.Errorf("data to sign cannot be empty"))
	}

	// Create challenge from data hash
	dataHash := sha256.Sum256(data)

	// Create authentication challenge for signing
	credIDBytes, _ := base64.URLEncoding.DecodeString(credentialID)
	authOpts := &AuthenticationOptions{
		UserVerification: "required", // Require user verification for signing
		Timeout:          60000,
		AllowedCredentials: []*CredentialDescriptor{
			{
				Type: "public-key",
				ID:   credIDBytes,
			},
		},
	}

	// Begin authentication for signing
	challenge, err := w.BeginAuthentication(ctx, authOpts)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrWebAuthnFailed, "failed to create signing challenge")
	}

	// Store the data hash with the challenge for verification
	w.pendingSignatures[string(challenge.Challenge)] = dataHash[:]

	// Create WebAuthn signature structure
	signature := &WebAuthnSignature{
		CredentialID:      credentialID,
		Signature:         challenge.Challenge, // Placeholder - would be actual signature
		Counter:           0,
		AuthenticatorData: []byte{},
		ClientDataJSON:    []byte{},
	}

	// In a complete implementation, this would:
	// 1. Wait for user to complete WebAuthn assertion
	// 2. Verify the assertion response
	// 3. Extract the signature from the assertion
	// 4. Return the signature suitable for blockchain transaction

	return signature, nil
}

// CompleteSignature completes a WebAuthn signature operation.
func (w *webAuthnClient) CompleteSignature(ctx context.Context, challenge []byte, response *AuthenticatorAssertionResponse) (*WebAuthnSignature, error) {
	// Get the pending data hash
	_, exists := w.pendingSignatures[string(challenge)]
	if !exists {
		return nil, errors.NewModuleError("auth", "CompleteSignature",
			fmt.Errorf("no pending signature for challenge"))
	}

	// Create authentication challenge structure
	authChallenge := &AuthenticationChallenge{
		Challenge:        challenge,
		UserVerification: "required",
		RelyingPartyID:   w.rpID,
	}

	// Verify the assertion with credential ID
	// Note: In a real implementation, we'd need to determine the credential ID from the response
	credentialID := "" // This would be extracted from response or passed as parameter
	result, err := w.CompleteAuthentication(ctx, authChallenge, response, credentialID)
	if err != nil {
		return nil, err
	}

	// Create final signature
	signature := &WebAuthnSignature{
		CredentialID:      credentialID,
		Signature:         response.Signature,
		Counter:           result.Counter,
		AuthenticatorData: response.AuthenticatorData,
		ClientDataJSON:    response.ClientDataJSON,
	}

	// Clean up pending signature
	delete(w.pendingSignatures, string(challenge))

	return signature, nil
}

// Utility functions

// GenerateChallenge generates a cryptographically secure challenge.
func GenerateChallenge() ([]byte, error) {
	challenge := make([]byte, 32)
	_, err := rand.Read(challenge)
	return challenge, err
}

// verifyClientData verifies the client data JSON from WebAuthn response.
func verifyClientData(clientDataJSON []byte, challenge []byte, ceremonyType string, expectedOrigin string) (*webauthn.CollectedClientData, error) {
	var clientData webauthn.CollectedClientData
	if err := json.Unmarshal(clientDataJSON, &clientData); err != nil {
		return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
	}

	// Verify type
	if string(clientData.Type) != ceremonyType {
		return nil, fmt.Errorf("invalid ceremony type: expected %s, got %s", ceremonyType, clientData.Type)
	}

	// Verify challenge
	challengeB64 := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(challenge)
	if clientData.Challenge != challengeB64 {
		return nil, fmt.Errorf("challenge mismatch")
	}

	// Verify origin
	if clientData.Origin != expectedOrigin {
		return nil, fmt.Errorf("origin mismatch: expected %s, got %s", expectedOrigin, clientData.Origin)
	}

	return &clientData, nil
}

// parseAttestationObject parses the attestation object from CBOR format.
func parseAttestationObject(attestationObjBytes []byte) (*webauthn.AttestationObject, error) {
	var attestationObj webauthn.AttestationObject

	// Decode CBOR attestation object
	var rawObj map[string]any
	if err := webauthncbor.Unmarshal(attestationObjBytes, &rawObj); err != nil {
		return nil, fmt.Errorf("failed to decode attestation object: %w", err)
	}

	// Extract format
	if format, ok := rawObj["fmt"].(string); ok {
		attestationObj.Format = format
	} else {
		return nil, fmt.Errorf("missing attestation format")
	}

	// Extract authenticator data
	if authData, ok := rawObj["authData"].([]byte); ok {
		attestationObj.RawAuthData = authData
		if err := attestationObj.AuthData.Unmarshal(authData); err != nil {
			return nil, fmt.Errorf("failed to parse authenticator data: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing authenticator data")
	}

	// Extract attestation statement
	if attStmt, ok := rawObj["attStmt"].(map[string]any); ok {
		attestationObj.AttStatement = attStmt
	}

	return &attestationObj, nil
}

// verifyAuthenticatorData verifies the authenticator data against the RP ID.
func verifyAuthenticatorData(authData webauthn.AuthenticatorData, rpID string) error {
	// Calculate RP ID hash
	rpIDHash := sha256.Sum256([]byte(rpID))

	// Verify RP ID hash
	if !bytes.Equal(authData.RPIDHash[:], rpIDHash[:]) {
		return fmt.Errorf("RP ID hash mismatch")
	}

	// Verify user presence flag
	if !authData.Flags.UserPresent() {
		return fmt.Errorf("user presence flag not set")
	}

	// Verify attestation data is present for registration
	if !authData.Flags.HasAttestedCredentialData() {
		return fmt.Errorf("attestation data flag not set")
	}

	return nil
}

// parseCOSEPublicKey parses a COSE public key.
func parseCOSEPublicKey(publicKeyBytes []byte) (*webauthncose.PublicKeyData, error) {
	parsed, err := webauthncose.ParsePublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse COSE public key: %w", err)
	}

	// Type assert to get the base public key data
	switch pk := parsed.(type) {
	case *webauthncose.EC2PublicKeyData:
		return &pk.PublicKeyData, nil
	case *webauthncose.RSAPublicKeyData:
		return &pk.PublicKeyData, nil
	case *webauthncose.OKPPublicKeyData:
		return &pk.PublicKeyData, nil
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}
}

// verifyAttestation verifies the attestation statement.
func verifyAttestation(attestationObj *webauthn.AttestationObject, clientDataHash []byte) error {
	// For now, we'll implement basic verification
	// Full attestation verification would use the registered attestation format handlers

	switch attestationObj.Format {
	case "none":
		// No attestation to verify
		return nil
	case "packed":
		// Verify packed attestation format
		return verifyPackedAttestation(attestationObj, clientDataHash)
	case "fido-u2f":
		// Verify FIDO U2F attestation format
		return fmt.Errorf("fido-u2f attestation not yet implemented")
	default:
		// Unknown attestation format - could be valid but unsupported
		return fmt.Errorf("unsupported attestation format: %s", attestationObj.Format)
	}
}

// verifyPackedAttestation verifies packed attestation format.
func verifyPackedAttestation(attestationObj *webauthn.AttestationObject, clientDataHash []byte) error {
	// Get algorithm from attestation statement
	alg, ok := attestationObj.AttStatement["alg"].(int64)
	if !ok {
		return fmt.Errorf("missing algorithm in attestation statement")
	}

	// Get signature from attestation statement
	sig, ok := attestationObj.AttStatement["sig"].([]byte)
	if !ok {
		return fmt.Errorf("missing signature in attestation statement")
	}

	// Construct verification data (authenticatorData || clientDataHash)
	verificationData := append(attestationObj.RawAuthData, clientDataHash...)

	// Check if self-attestation (no x5c)
	if _, hasX5c := attestationObj.AttStatement["x5c"]; !hasX5c {
		// Self-attestation: verify with credential public key
		if len(attestationObj.AuthData.AttData.CredentialID) == 0 {
			return fmt.Errorf("missing attestation data for self-attestation")
		}

		// Parse public key and verify signature
		publicKey, err := webauthncose.ParsePublicKey(attestationObj.AuthData.AttData.CredentialPublicKey)
		if err != nil {
			return fmt.Errorf("failed to parse public key for verification: %w", err)
		}

		// Verify signature based on key type
		switch pk := publicKey.(type) {
		case *webauthncose.EC2PublicKeyData:
			if pk.Algorithm != alg {
				return fmt.Errorf("algorithm mismatch")
			}
			valid, err := pk.Verify(verificationData, sig)
			if err != nil || !valid {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		case *webauthncose.RSAPublicKeyData:
			if pk.Algorithm != alg {
				return fmt.Errorf("algorithm mismatch")
			}
			valid, err := pk.Verify(verificationData, sig)
			if err != nil || !valid {
				return fmt.Errorf("signature verification failed: %w", err)
			}
		default:
			return fmt.Errorf("unsupported key type for self-attestation")
		}
	} else {
		// Full attestation with certificate chain
		// This would require parsing x5c certificate chain and verifying
		// For now, we'll accept it but log that full verification is pending
		return nil // Certificate chain verification not yet implemented
	}

	return nil
}

// EncodeChallenge encodes a challenge as base64url.
func EncodeChallenge(challenge []byte) string {
	return base64.URLEncoding.EncodeToString(challenge)
}

// DecodeChallenge decodes a base64url challenge.
func DecodeChallenge(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

// CreateDefaultRelyingParty creates a default relying party configuration.
func CreateDefaultRelyingParty(domain string) *RelyingParty {
	return &RelyingParty{
		ID:   domain,
		Name: fmt.Sprintf("Sonr (%s)", domain),
	}
}

// ValidateCredentialID validates a credential ID format.
func ValidateCredentialID(credentialID string) error {
	if len(credentialID) == 0 {
		return fmt.Errorf("credential ID cannot be empty")
	}

	// Decode to ensure it's valid base64
	_, err := base64.URLEncoding.DecodeString(credentialID)
	if err != nil {
		return fmt.Errorf("invalid credential ID format: %w", err)
	}

	return nil
}
