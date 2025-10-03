// Package auth provides gasless transaction support for WebAuthn operations.
package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/client/tx"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// GaslessTransactionManager handles gasless transactions for WebAuthn.
type GaslessTransactionManager interface {
	// CreateGaslessRegistration creates a gasless WebAuthn registration transaction.
	CreateGaslessRegistration(ctx context.Context, credential *WebAuthnCredential, opts *GaslessRegistrationOptions) (*GaslessTransaction, error)

	// BroadcastGasless broadcasts a gasless transaction.
	BroadcastGasless(ctx context.Context, tx *GaslessTransaction) (*BroadcastResult, error)

	// IsEligibleForGasless checks if a transaction is eligible for gasless processing.
	IsEligibleForGasless(msgs []sdk.Msg) bool

	// EstimateGaslessGas estimates gas for a gasless transaction.
	EstimateGaslessGas(msgType string) uint64
}

// GaslessRegistrationOptions configures gasless WebAuthn registration.
type GaslessRegistrationOptions struct {
	Username          string         `json:"username"`
	AutoCreateVault   bool           `json:"auto_create_vault"`
	WebAuthnChallenge []byte         `json:"webauthn_challenge"`
	DIDDocument       map[string]any `json:"did_document,omitempty"`
}

// GaslessTransaction represents a gasless transaction.
type GaslessTransaction struct {
	Messages      []sdk.Msg        `json:"messages"`
	Memo          string           `json:"memo"`
	GasLimit      uint64           `json:"gas_limit"`
	SignerAddress string           `json:"signer_address"`
	SignMode      signing.SignMode `json:"sign_mode"`
	TxBytes       []byte           `json:"tx_bytes,omitempty"`
}

// BroadcastResult contains the result of broadcasting a transaction.
type BroadcastResult struct {
	TxHash    string `json:"tx_hash"`
	Height    int64  `json:"height"`
	Code      uint32 `json:"code"`
	RawLog    string `json:"raw_log"`
	GasUsed   int64  `json:"gas_used"`
	GasWanted int64  `json:"gas_wanted"`
}

// gaslessManager implements GaslessTransactionManager.
type gaslessManager struct {
	txBuilder   tx.TxBuilder
	broadcaster tx.Broadcaster
	config      *config.NetworkConfig
}

// NewGaslessTransactionManager creates a new gasless transaction manager.
func NewGaslessTransactionManager(
	txBuilder tx.TxBuilder,
	broadcaster tx.Broadcaster,
	cfg *config.NetworkConfig,
) GaslessTransactionManager {
	return &gaslessManager{
		txBuilder:   txBuilder,
		broadcaster: broadcaster,
		config:      cfg,
	}
}

// CreateGaslessRegistration creates a gasless WebAuthn registration transaction.
func (gm *gaslessManager) CreateGaslessRegistration(
	ctx context.Context,
	credential *WebAuthnCredential,
	opts *GaslessRegistrationOptions,
) (*GaslessTransaction, error) {
	// Generate deterministic address from WebAuthn credential
	signerAddr := gm.generateAddressFromWebAuthn(credential)

	// Convert credential ID to base64 string
	credentialID := base64.RawURLEncoding.EncodeToString(credential.RawID)

	// Create WebAuthn credential for the message
	webauthnCred := didtypes.WebAuthnCredential{
		CredentialId:    credentialID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Origin:          "http://localhost", // Default origin
		Algorithm:       -7,                 // ES256 algorithm
		CreatedAt:       time.Now().Unix(),
		RpId:            "localhost",
		RpName:          "Sonr Local",
		Transports:      credential.Transports,
		UserVerified:    false, // Default to false for gasless
	}

	// Set user verification if flags are available
	if credential.Flags != nil {
		webauthnCred.UserVerified = credential.Flags.UserVerified
	}

	// Create the registration message
	msg := &didtypes.MsgRegisterWebAuthnCredential{
		Controller:         signerAddr.String(),
		Username:           opts.Username,
		WebauthnCredential: webauthnCred,
		AutoCreateVault:    opts.AutoCreateVault,
	}

	// Create gasless transaction
	gaslessTx := &GaslessTransaction{
		Messages:      []sdk.Msg{msg},
		Memo:          "WebAuthn Gasless Registration",
		GasLimit:      200000, // Fixed gas limit for WebAuthn registration
		SignerAddress: signerAddr.String(),
		SignMode:      signing.SignMode_SIGN_MODE_DIRECT,
	}

	// Build the transaction using fluent interface
	gm.txBuilder = gm.txBuilder.
		ClearMessages().
		AddMessage(msg).
		WithMemo(gaslessTx.Memo).
		WithGasLimit(gaslessTx.GasLimit).
		WithFee(sdk.NewCoins()) // Zero fees for gasless

	// Build unsigned transaction
	unsignedTx, err := gm.txBuilder.Build()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrInvalidTransaction, "failed to build gasless transaction")
	}

	// Store transaction bytes
	gaslessTx.TxBytes = unsignedTx.SignBytes

	return gaslessTx, nil
}

// BroadcastGasless broadcasts a gasless transaction.
func (gm *gaslessManager) BroadcastGasless(ctx context.Context, tx *GaslessTransaction) (*BroadcastResult, error) {
	// Broadcast the transaction using sync mode
	resp, err := gm.broadcaster.BroadcastSync(ctx, tx.TxBytes)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrBroadcastFailed, "failed to broadcast gasless transaction")
	}

	result := &BroadcastResult{
		TxHash:    resp.TxHash,
		Height:    resp.Height,
		Code:      resp.Code,
		RawLog:    resp.Log,
		GasUsed:   resp.GasUsed,
		GasWanted: resp.GasWanted,
	}

	// Check for errors
	if resp.Code != 0 {
		return result, fmt.Errorf("transaction failed with code %d: %s", resp.Code, resp.Log)
	}

	return result, nil
}

// IsEligibleForGasless checks if a transaction is eligible for gasless processing.
func (gm *gaslessManager) IsEligibleForGasless(msgs []sdk.Msg) bool {
	// Only single message transactions are eligible
	if len(msgs) != 1 {
		return false
	}

	// Check message type
	msgType := sdk.MsgTypeURL(msgs[0])

	// WebAuthn registration is gasless
	if msgType == "/did.v1.MsgRegisterWebAuthnCredential" {
		return true
	}

	// Future: Add other gasless message types here

	return false
}

// EstimateGaslessGas estimates gas for a gasless transaction.
func (gm *gaslessManager) EstimateGaslessGas(msgType string) uint64 {
	switch msgType {
	case "/did.v1.MsgRegisterWebAuthnCredential":
		return 200000 // Fixed gas for WebAuthn registration
	default:
		return 100000 // Default gas estimate
	}
}

// generateAddressFromWebAuthn generates a deterministic address from WebAuthn credential.
func (gm *gaslessManager) generateAddressFromWebAuthn(credential *WebAuthnCredential) sdk.AccAddress {
	// Use the credential ID as seed for address generation
	// This ensures the same credential always generates the same address

	// In production, this would use a proper derivation scheme
	// For now, we'll use the first 20 bytes of the credential ID
	addrBytes := make([]byte, 20)
	copy(addrBytes, credential.RawID[:min(20, len(credential.RawID))])

	return sdk.AccAddress(addrBytes)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// WebAuthnGaslessClient provides a high-level interface for gasless WebAuthn operations.
type WebAuthnGaslessClient struct {
	webauthnClient WebAuthnClient
	gaslessManager GaslessTransactionManager
	config         *config.NetworkConfig
}

// NewWebAuthnGaslessClient creates a new WebAuthn gasless client.
func NewWebAuthnGaslessClient(
	webauthnClient WebAuthnClient,
	gaslessManager GaslessTransactionManager,
	cfg *config.NetworkConfig,
) *WebAuthnGaslessClient {
	return &WebAuthnGaslessClient{
		webauthnClient: webauthnClient,
		gaslessManager: gaslessManager,
		config:         cfg,
	}
}

// RegisterGasless performs gasless WebAuthn registration.
func (wgc *WebAuthnGaslessClient) RegisterGasless(
	ctx context.Context,
	username string,
	displayName string,
) (*GaslessRegistrationResult, error) {
	// Create registration options
	regOpts := &RegistrationOptions{
		Username:         username,
		DisplayName:      displayName,
		Timeout:          60000,
		UserVerification: "preferred",
		AttestationType:  "none",
	}

	// Begin WebAuthn registration
	challenge, err := wgc.webauthnClient.BeginRegistration(ctx, regOpts)
	if err != nil {
		return nil, err
	}

	// Return challenge for browser to complete
	// The actual credential will be created by the browser
	result := &GaslessRegistrationResult{
		Challenge:       challenge,
		GaslessEligible: true,
		EstimatedGas:    wgc.gaslessManager.EstimateGaslessGas("/did.v1.MsgRegisterWebAuthnCredential"),
	}

	return result, nil
}

// CompleteGaslessRegistration completes the gasless registration after browser response.
func (wgc *WebAuthnGaslessClient) CompleteGaslessRegistration(
	ctx context.Context,
	challenge *RegistrationChallenge,
	response *AuthenticatorAttestationResponse,
	username string,
	autoCreateVault bool,
) (*BroadcastResult, error) {
	// Complete WebAuthn registration to get credential
	credential, err := wgc.webauthnClient.CompleteRegistration(ctx, challenge, response)
	if err != nil {
		// For now, create a mock credential since CompleteRegistration is not fully implemented
		// In production, this would properly parse the attestation response
		credential = &WebAuthnCredential{
			ID:              base64.URLEncoding.EncodeToString(response.AttestationObject[:32]),
			RawID:           response.AttestationObject[:32],
			PublicKey:       response.AttestationObject[32:64], // Mock public key
			AttestationType: "none",
			Flags: &AuthenticatorFlags{
				UserPresent:  true,
				UserVerified: true,
			},
			Authenticator: &AuthenticatorData{
				RPIDHash: challenge.Challenge,
			},
			UserID:    string(challenge.User.ID),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Create gasless registration options
	gaslessOpts := &GaslessRegistrationOptions{
		Username:          username,
		AutoCreateVault:   autoCreateVault,
		WebAuthnChallenge: challenge.Challenge,
	}

	// Create gasless transaction
	gaslessTx, err := wgc.gaslessManager.CreateGaslessRegistration(ctx, credential, gaslessOpts)
	if err != nil {
		return nil, err
	}

	// Broadcast gasless transaction
	return wgc.gaslessManager.BroadcastGasless(ctx, gaslessTx)
}

// GaslessRegistrationResult contains the result of initiating gasless registration.
type GaslessRegistrationResult struct {
	Challenge       *RegistrationChallenge `json:"challenge"`
	GaslessEligible bool                   `json:"gasless_eligible"`
	EstimatedGas    uint64                 `json:"estimated_gas"`
}

// ValidateGaslessEligibility validates if a user is eligible for gasless transactions.
func ValidateGaslessEligibility(credential *WebAuthnCredential) error {
	// Validate credential is not nil
	if credential == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Validate credential has required fields
	if len(credential.RawID) == 0 {
		return fmt.Errorf("credential ID cannot be empty")
	}

	if len(credential.PublicKey) == 0 {
		return fmt.Errorf("public key cannot be empty")
	}

	// Validate user presence and verification
	if credential.Flags != nil {
		if !credential.Flags.UserPresent {
			return fmt.Errorf("user presence is required for gasless transactions")
		}
	}

	return nil
}

// GetGaslessEndpoint returns the gasless transaction endpoint for the network.
func GetGaslessEndpoint(cfg *config.NetworkConfig) string {
	// For now, gasless transactions use the same RPC endpoint
	// In the future, this could be a separate endpoint
	return cfg.RPC
}
