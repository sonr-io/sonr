//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/extism/go-pdk"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonr-io/sonr/crypto/mpc"
)

const (
	KeyChainID     = "chain_id"
	KeyEnclave     = "enclave"
	KeyVaultConfig = "vault_config"
)

// GetChainID returns the chain ID to use for unlocking the enclave
func GetChainID() string {
	v := pdk.GetVar(KeyChainID)
	if v == nil {
		return "sonr-testnet-1"
	}
	return string(v)
}

// GetEnclaveData loads MPC enclave data from PDK environment
func GetEnclaveData() (*mpc.EnclaveData, error) {
	v := pdk.GetVar(KeyEnclave)
	if v == nil {
		return nil, fmt.Errorf("enclave data not provided in environment")
	}

	var data mpc.EnclaveData
	if err := json.Unmarshal(v, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enclave data: %w", err)
	}

	return &data, nil
}

// GetVaultConfig loads vault configuration from PDK environment
func GetVaultConfig() map[string]any {
	v := pdk.GetVar(KeyVaultConfig)
	if v == nil {
		return make(map[string]any)
	}

	var config map[string]any
	if err := json.Unmarshal(v, &config); err != nil {
		pdk.Log(pdk.LogWarn, fmt.Sprintf("Failed to parse vault config: %v", err))
		return make(map[string]any)
	}

	return config
}

// UCAN Token Request/Response types
type NewOriginTokenRequest struct {
	AudienceDID  string           `json:"audience_did"`
	Attenuations []map[string]any `json:"attenuations,omitempty"`
	Facts        []string         `json:"facts,omitempty"`
	NotBefore    int64            `json:"not_before,omitempty"`
	ExpiresAt    int64            `json:"expires_at,omitempty"`
}

type NewAttenuatedTokenRequest struct {
	ParentToken  string           `json:"parent_token"`
	AudienceDID  string           `json:"audience_did"`
	Attenuations []map[string]any `json:"attenuations,omitempty"`
	Facts        []string         `json:"facts,omitempty"`
	NotBefore    int64            `json:"not_before,omitempty"`
	ExpiresAt    int64            `json:"expires_at,omitempty"`
}

type UCANTokenResponse struct {
	Token   string `json:"token"`
	Issuer  string `json:"issuer"`
	Address string `json:"address"`
	Error   string `json:"error,omitempty"`
}

type SignDataRequest struct {
	Data []byte `json:"data"`
}

type SignDataResponse struct {
	Signature []byte `json:"signature"`
	Error     string `json:"error,omitempty"`
}

type VerifyDataRequest struct {
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
}

type VerifyDataResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type GetIssuerDIDResponse struct {
	IssuerDID string `json:"issuer_did"`
	Address   string `json:"address"`
	ChainCode string `json:"chain_code"`
	Error     string `json:"error,omitempty"`
}

var (
	enclave   mpc.Enclave
	issuerDID string
	address   string
)

func main() {
	// Initialize MPC enclave from PDK environment
	if err := initializeEnclave(); err != nil {
		pdk.SetError(fmt.Errorf("failed to initialize enclave: %w", err))
		return
	}
	pdk.Log(pdk.LogInfo, "Motor plugin initialized as MPC-based UCAN source")
}

// initializeEnclave initializes the MPC enclave from PDK environment
func initializeEnclave() error {
	// Load enclave data from PDK environment
	enclaveData, err := GetEnclaveData()
	if err != nil {
		return fmt.Errorf("failed to get enclave data: %w", err)
	}

	// Import MPC enclave from data
	enclave, err = mpc.ImportEnclave(mpc.WithEnclaveData(enclaveData))
	if err != nil {
		return fmt.Errorf("failed to import enclave: %w", err)
	}

	// Derive issuer DID and address from enclave public key
	pubKeyBytes := enclave.PubKeyBytes()
	issuerDID, address, err = deriveIssuerDIDFromBytes(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to derive issuer DID: %w", err)
	}
	return nil
}

//go:wasmexport new_origin_token
func newOriginToken() int32 {
	if !enclave.IsValid() {
		pdk.SetError(fmt.Errorf("enclave not initialized"))
		return 1
	}

	req := &NewOriginTokenRequest{}
	err := pdk.InputJSON(req)
	if err != nil {
		pdk.SetError(fmt.Errorf("failed to parse request: %w", err))
		return 1
	}

	// Convert timestamps
	var notBefore, expiresAt time.Time
	if req.NotBefore > 0 {
		notBefore = time.Unix(req.NotBefore, 0)
	}
	if req.ExpiresAt > 0 {
		expiresAt = time.Unix(req.ExpiresAt, 0)
	}

	// Create origin token using MPC signing
	tokenString, err := createUCANToken(
		req.AudienceDID,
		nil,
		req.Attenuations,
		req.Facts,
		notBefore,
		expiresAt,
	)
	if err != nil {
		resp := &UCANTokenResponse{Error: err.Error()}
		pdk.OutputJSON(resp)
		return 1
	}

	resp := &UCANTokenResponse{
		Token:   tokenString,
		Issuer:  issuerDID,
		Address: address,
	}
	pdk.OutputJSON(resp)
	return 0
}

//go:wasmexport new_attenuated_token
func newAttenuatedToken() int32 {
	if !enclave.IsValid() {
		pdk.SetError(fmt.Errorf("enclave not initialized"))
		return 1
	}

	req := &NewAttenuatedTokenRequest{}
	err := pdk.InputJSON(req)
	if err != nil {
		pdk.SetError(fmt.Errorf("failed to parse request: %w", err))
		return 1
	}

	// Convert timestamps
	var notBefore, expiresAt time.Time
	if req.NotBefore > 0 {
		notBefore = time.Unix(req.NotBefore, 0)
	}
	if req.ExpiresAt > 0 {
		expiresAt = time.Unix(req.ExpiresAt, 0)
	}

	// Create proofs from parent token
	proofs := []string{req.ParentToken}

	// Create attenuated token using MPC signing
	tokenString, err := createUCANToken(
		req.AudienceDID,
		proofs,
		req.Attenuations,
		req.Facts,
		notBefore,
		expiresAt,
	)
	if err != nil {
		resp := &UCANTokenResponse{Error: err.Error()}
		pdk.OutputJSON(resp)
		return 1
	}

	resp := &UCANTokenResponse{
		Token:   tokenString,
		Issuer:  issuerDID,
		Address: address,
	}
	pdk.OutputJSON(resp)
	return 0
}

//go:wasmexport sign_data
func signData() int32 {
	if !enclave.IsValid() {
		pdk.SetError(fmt.Errorf("enclave not initialized"))
		return 1
	}

	req := &SignDataRequest{}
	err := pdk.InputJSON(req)
	if err != nil {
		pdk.SetError(fmt.Errorf("failed to parse request: %w", err))
		return 1
	}

	// Sign data using MPC enclave
	signature, err := enclave.Sign(req.Data)
	if err != nil {
		resp := &SignDataResponse{Error: err.Error()}
		pdk.OutputJSON(resp)
		return 1
	}

	resp := &SignDataResponse{Signature: signature}
	pdk.OutputJSON(resp)
	return 0
}

//go:wasmexport verify_data
func verifyData() int32 {
	if !enclave.IsValid() {
		pdk.SetError(fmt.Errorf("enclave not initialized"))
		return 1
	}

	req := &VerifyDataRequest{}
	err := pdk.InputJSON(req)
	if err != nil {
		pdk.SetError(fmt.Errorf("failed to parse request: %w", err))
		return 1
	}

	// Verify data using MPC enclave
	valid, err := enclave.Verify(req.Data, req.Signature)
	if err != nil {
		resp := &VerifyDataResponse{Error: err.Error()}
		pdk.OutputJSON(resp)
		return 1
	}

	resp := &VerifyDataResponse{Valid: valid}
	pdk.OutputJSON(resp)
	return 0
}

//go:wasmexport get_issuer_did
func getIssuerDID() int32 {
	if !enclave.IsValid() {
		pdk.SetError(fmt.Errorf("enclave not initialized"))
		return 1
	}

	// Get chain code for deterministic derivation
	chainCode, err := getChainCode()
	if err != nil {
		resp := &GetIssuerDIDResponse{Error: err.Error()}
		pdk.OutputJSON(resp)
		return 1
	}

	resp := &GetIssuerDIDResponse{
		IssuerDID: issuerDID,
		Address:   address,
		ChainCode: fmt.Sprintf("%x", chainCode),
	}
	pdk.OutputJSON(resp)
	return 0
}

// UCAN token creation and MPC signing implementation

// MPCSigningMethod implements JWT signing using MPC enclaves
type MPCSigningMethod struct {
	Name    string
	enclave mpc.Enclave
}

// Alg returns the signing method algorithm name
func (m *MPCSigningMethod) Alg() string {
	return m.Name
}

// Sign signs a JWT string using the MPC enclave
func (m *MPCSigningMethod) Sign(signingString string, key any) ([]byte, error) {
	// Hash the signing string
	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	digest := hasher.Sum(nil)

	// Use MPC enclave to sign the digest
	sig, err := m.enclave.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with MPC: %w", err)
	}

	return sig, nil
}

// Verify verifies a JWT signature using the MPC enclave
func (m *MPCSigningMethod) Verify(signingString string, sig []byte, key any) error {
	// Hash the signing string
	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	digest := hasher.Sum(nil)

	// Use MPC enclave to verify signature
	valid, err := m.enclave.Verify(digest, sig)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !valid {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// createUCANToken creates a UCAN token using MPC signing
func createUCANToken(
	audienceDID string,
	proofs []string,
	attenuations []map[string]any,
	facts []string,
	notBefore, expiresAt time.Time,
) (string, error) {
	// Validate audience DID
	if audienceDID == "" {
		return "", fmt.Errorf("audience DID is required")
	}

	// Create MPC signing method
	signingMethod := &MPCSigningMethod{
		Name:    "MPC256",
		enclave: enclave,
	}

	// Create JWT token
	token := jwt.New(signingMethod)

	// Set UCAN version in header
	token.Header["ucv"] = "0.9.0"

	// Prepare time claims
	var nbfUnix, expUnix int64
	if !notBefore.IsZero() {
		nbfUnix = notBefore.Unix()
	}
	if !expiresAt.IsZero() {
		expUnix = expiresAt.Unix()
	}

	// Set claims
	claims := jwt.MapClaims{
		"iss": issuerDID,
		"aud": audienceDID,
	}

	// Add attenuations if provided
	if len(attenuations) > 0 {
		claims["att"] = attenuations
	}

	// Add proofs if provided
	if len(proofs) > 0 {
		claims["prf"] = proofs
	}

	// Add facts if provided
	if len(facts) > 0 {
		claims["fct"] = facts
	}

	// Add time claims
	if nbfUnix > 0 {
		claims["nbf"] = nbfUnix
	}
	if expUnix > 0 {
		claims["exp"] = expUnix
	}

	token.Claims = claims

	// Sign the token using MPC enclave (key parameter is ignored for MPC signing)
	tokenString, err := token.SignedString(nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign token with MPC: %w", err)
	}

	return tokenString, nil
}

// deriveIssuerDIDFromBytes creates issuer DID and address from public key bytes
func deriveIssuerDIDFromBytes(pubKeyBytes []byte) (string, string, error) {
	if len(pubKeyBytes) == 0 {
		return "", "", fmt.Errorf("empty public key bytes")
	}

	// Generate address from public key (simplified implementation)
	address := fmt.Sprintf("sonr1%x", pubKeyBytes[:20])

	// Create DID from address (simplified implementation)
	issuerDID := fmt.Sprintf("did:sonr:%s", address)

	return issuerDID, address, nil
}

// getChainCode derives a deterministic chain code from the enclave
func getChainCode() ([]byte, error) {
	if !enclave.IsValid() {
		return nil, fmt.Errorf("enclave is not valid")
	}

	// Sign the address to create a deterministic chain code
	sig, err := enclave.Sign([]byte(address))
	if err != nil {
		return nil, fmt.Errorf("failed to sign address for chain code: %w", err)
	}

	// Hash the signature to create a 32-byte chain code
	hasher := sha256.New()
	hasher.Write(sig)
	hash := hasher.Sum(nil)

	return hash[:32], nil
}
