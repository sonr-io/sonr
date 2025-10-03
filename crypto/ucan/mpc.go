// Package ucan provides User-Controlled Authorization Networks (UCAN) implementation
// for decentralized authorization and capability delegation in the Sonr network.
// This package handles JWT-based tokens, cryptographic verification, and resource capabilities.
package ucan

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/mpc"
)

// MPCSigningMethod implements JWT signing using MPC enclaves
type MPCSigningMethod struct {
	Name    string
	enclave mpc.Enclave
}

// NewMPCSigningMethod creates a new MPC-based JWT signing method
func NewMPCSigningMethod(name string, enclave mpc.Enclave) *MPCSigningMethod {
	return &MPCSigningMethod{
		Name:    name,
		enclave: enclave,
	}
}

// Alg returns the signing method algorithm name
func (m *MPCSigningMethod) Alg() string {
	return m.Name
}

// Verify verifies a JWT signature using the MPC enclave
func (m *MPCSigningMethod) Verify(signingString string, signature []byte, key any) error {
	// signature is already decoded bytes
	sig := signature

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

// MPCTokenBuilder creates UCAN tokens using MPC signing
type MPCTokenBuilder struct {
	enclave       mpc.Enclave
	issuerDID     string
	address       string
	signingMethod *MPCSigningMethod
}

// NewMPCTokenBuilder creates a new MPC-based UCAN token builder
func NewMPCTokenBuilder(enclave mpc.Enclave) (*MPCTokenBuilder, error) {
	if !enclave.IsValid() {
		return nil, fmt.Errorf("invalid MPC enclave provided")
	}

	// Derive issuer DID and address from enclave public key
	pubKeyBytes := enclave.PubKeyBytes()
	issuerDID, address := deriveIssuerDIDFromBytes(pubKeyBytes)

	signingMethod := NewMPCSigningMethod("MPC256", enclave)

	return &MPCTokenBuilder{
		enclave:       enclave,
		issuerDID:     issuerDID,
		address:       address,
		signingMethod: signingMethod,
	}, nil
}

// GetIssuerDID returns the issuer DID derived from the enclave
func (b *MPCTokenBuilder) GetIssuerDID() string {
	return b.issuerDID
}

// GetAddress returns the address derived from the enclave
func (b *MPCTokenBuilder) GetAddress() string {
	return b.address
}

// CreateOriginToken creates a new origin UCAN token using MPC signing
func (b *MPCTokenBuilder) CreateOriginToken(
	audienceDID string,
	attenuations []Attenuation,
	facts []Fact,
	notBefore, expiresAt time.Time,
) (*Token, error) {
	return b.createToken(audienceDID, nil, attenuations, facts, notBefore, expiresAt)
}

// CreateDelegatedToken creates a delegated UCAN token using MPC signing
func (b *MPCTokenBuilder) CreateDelegatedToken(
	parent *Token,
	audienceDID string,
	attenuations []Attenuation,
	facts []Fact,
	notBefore, expiresAt time.Time,
) (*Token, error) {
	proofs, err := prepareDelegationProofs(parent, attenuations)
	if err != nil {
		return nil, err
	}

	return b.createToken(audienceDID, proofs, attenuations, facts, notBefore, expiresAt)
}

// createToken creates a UCAN token with MPC signing
func (b *MPCTokenBuilder) createToken(
	audienceDID string,
	proofs []Proof,
	attenuations []Attenuation,
	facts []Fact,
	notBefore, expiresAt time.Time,
) (*Token, error) {
	// Validate inputs
	if !isValidDID(audienceDID) {
		return nil, fmt.Errorf("invalid audience DID format: %s", audienceDID)
	}
	if len(attenuations) == 0 {
		return nil, fmt.Errorf("at least one attenuation is required")
	}

	// Create JWT token with MPC signing method
	token := jwt.New(b.signingMethod)

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

	// Convert attenuations to claim format
	attClaims := make([]map[string]any, len(attenuations))
	for i, att := range attenuations {
		attClaims[i] = map[string]any{
			"can":  att.Capability.GetActions(),
			"with": att.Resource.GetURI(),
		}
	}

	// Convert proofs to strings
	proofStrings := make([]string, len(proofs))
	for i, proof := range proofs {
		proofStrings[i] = string(proof)
	}

	// Convert facts to any slice
	factData := make([]any, len(facts))
	for i, fact := range facts {
		// Facts are stored as raw JSON, convert to any
		factData[i] = string(fact.Data)
	}

	// Set claims
	claims := jwt.MapClaims{
		"iss": b.issuerDID,
		"aud": audienceDID,
		"att": attClaims,
	}

	if nbfUnix > 0 {
		claims["nbf"] = nbfUnix
	}
	if expUnix > 0 {
		claims["exp"] = expUnix
	}
	if len(proofStrings) > 0 {
		claims["prf"] = proofStrings
	}
	if len(factData) > 0 {
		claims["fct"] = factData
	}

	token.Claims = claims

	// Sign the token using MPC enclave (key parameter is ignored for MPC signing)
	tokenString, err := token.SignedString(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token with MPC: %w", err)
	}

	return &Token{
		Raw:          tokenString,
		Issuer:       b.issuerDID,
		Audience:     audienceDID,
		ExpiresAt:    expUnix,
		NotBefore:    nbfUnix,
		Attenuations: attenuations,
		Proofs:       proofs,
		Facts:        facts,
	}, nil
}

// CreateVaultCapabilityToken creates a vault-specific UCAN token
func (b *MPCTokenBuilder) CreateVaultCapabilityToken(
	audienceDID string,
	vaultAddress string,
	enclaveDataCID string,
	actions []string,
	expiresAt time.Time,
) (*Token, error) {
	// Create vault-specific attenuation
	attenuation := CreateVaultAttenuation(actions, enclaveDataCID, vaultAddress)

	return b.CreateOriginToken(
		audienceDID,
		[]Attenuation{attenuation},
		nil,
		time.Time{}, // No not-before restriction
		expiresAt,
	)
}

// MPCDIDResolver resolves DIDs with special handling for MPC-derived DIDs
type MPCDIDResolver struct {
	enclave   mpc.Enclave
	issuerDID string
	fallback  DIDResolver
}

// NewMPCDIDResolver creates a new MPC DID resolver
func NewMPCDIDResolver(enclave mpc.Enclave, fallback DIDResolver) *MPCDIDResolver {
	pubKeyBytes := enclave.PubKeyBytes()
	issuerDID, _ := deriveIssuerDIDFromBytes(pubKeyBytes)

	return &MPCDIDResolver{
		enclave:   enclave,
		issuerDID: issuerDID,
		fallback:  fallback,
	}
}

// ResolveDIDKey resolves DID keys with MPC enclave support
func (r *MPCDIDResolver) ResolveDIDKey(ctx context.Context, didStr string) (keys.DID, error) {
	// Check if this is the MPC-derived DID
	if didStr == r.issuerDID {
		return r.createDIDFromEnclave()
	}

	// Fall back to standard DID resolution
	if r.fallback != nil {
		return r.fallback.ResolveDIDKey(ctx, didStr)
	}

	// Default fallback to string parsing
	return keys.Parse(didStr)
}

// createDIDFromEnclave creates a DID from the MPC enclave's public key
func (r *MPCDIDResolver) createDIDFromEnclave() (keys.DID, error) {
	// This would need to be implemented based on how MPC public keys
	// are converted to the keys.DID format
	// For now, parse from the derived DID string
	return keys.Parse(r.issuerDID)
}

// MPCVerifier provides UCAN verification with MPC support
type MPCVerifier struct {
	*Verifier
	enclave mpc.Enclave
}

// NewMPCVerifier creates a UCAN verifier with MPC support
func NewMPCVerifier(enclave mpc.Enclave) *MPCVerifier {
	resolver := NewMPCDIDResolver(enclave, StringDIDResolver{})
	verifier := NewVerifier(resolver)

	return &MPCVerifier{
		Verifier: verifier,
		enclave:  enclave,
	}
}

// VerifyMPCToken verifies a UCAN token that may be signed with MPC
func (v *MPCVerifier) VerifyMPCToken(ctx context.Context, tokenString string) (*Token, error) {
	// Try standard verification first
	token, err := v.VerifyToken(ctx, tokenString)
	if err == nil {
		return token, nil
	}

	// If standard verification fails, try MPC-specific verification
	return v.verifyWithMPC(ctx, tokenString)
}

// verifyWithMPC attempts to verify using MPC signing method
func (v *MPCVerifier) verifyWithMPC(_ context.Context, tokenString string) (*Token, error) {
	// Create MPC signing method for verification
	mpcMethod := NewMPCSigningMethod("MPC256", v.enclave)

	// Parse with MPC method
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Ensure the token uses MPC signing method
		if token.Method.Alg() != mpcMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		// For MPC verification, the key is not used
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("MPC token verification failed: %w", err)
	}

	// Extract and parse claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	ucanToken, err := v.parseUCANClaims(claims, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UCAN claims: %w", err)
	}

	return ucanToken, nil
}

// MPCTokenValidator provides comprehensive UCAN token validation with MPC support
type MPCTokenValidator struct {
	*MPCVerifier
	enclaveValidation bool
}

// NewMPCTokenValidator creates a comprehensive UCAN token validator with MPC support
func NewMPCTokenValidator(enclave mpc.Enclave, enableEnclaveValidation bool) *MPCTokenValidator {
	verifier := NewMPCVerifier(enclave)
	return &MPCTokenValidator{
		MPCVerifier:       verifier,
		enclaveValidation: enableEnclaveValidation,
	}
}

// ValidateTokenForVaultOperation performs comprehensive validation for vault operations
func (v *MPCTokenValidator) ValidateTokenForVaultOperation(
	ctx context.Context,
	tokenString string,
	enclaveDataCID string,
	requiredAction string,
	vaultAddress string,
) (*Token, error) {
	// Step 1: Verify token signature and structure
	token, err := v.VerifyMPCToken(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// Step 2: Validate vault-specific capability
	if err := ValidateVaultTokenCapability(token, enclaveDataCID, requiredAction); err != nil {
		return nil, fmt.Errorf("vault capability validation failed: %w", err)
	}

	// Step 3: Validate enclave data CID if enabled
	if v.enclaveValidation {
		if err := v.validateEnclaveDataCID(token, enclaveDataCID); err != nil {
			return nil, fmt.Errorf("enclave data validation failed: %w", err)
		}
	}

	// Step 4: Validate vault address if provided
	if vaultAddress != "" {
		if err := v.validateVaultAddress(token, vaultAddress); err != nil {
			return nil, fmt.Errorf("vault address validation failed: %w", err)
		}
	}

	// Step 5: Verify delegation chain if proofs exist
	if len(token.Proofs) > 0 {
		if err := v.VerifyDelegationChain(ctx, tokenString); err != nil {
			return nil, fmt.Errorf("delegation chain validation failed: %w", err)
		}
	}

	return token, nil
}

// ValidateTokenForResource validates token capabilities for a specific resource
func (v *MPCTokenValidator) ValidateTokenForResource(
	ctx context.Context,
	tokenString string,
	resourceURI string,
	requiredAbilities []string,
) (*Token, error) {
	token, err := v.VerifyCapability(ctx, tokenString, resourceURI, requiredAbilities)
	if err != nil {
		return nil, fmt.Errorf("capability verification failed: %w", err)
	}

	// Additional MPC-specific validation
	if v.enclaveValidation {
		if err := v.validateMPCIssuer(token); err != nil {
			return nil, fmt.Errorf("MPC issuer validation failed: %w", err)
		}
	}

	return token, nil
}

// validateEnclaveDataCID validates that the token contains the expected enclave data CID
func (v *MPCTokenValidator) validateEnclaveDataCID(token *Token, expectedCID string) error {
	tokenCID, err := GetEnclaveDataCID(token)
	if err != nil {
		return fmt.Errorf("failed to extract enclave data CID from token: %w", err)
	}

	if tokenCID != expectedCID {
		return fmt.Errorf("enclave data CID mismatch: token=%s, expected=%s", tokenCID, expectedCID)
	}

	return nil
}

// validateVaultAddress validates the vault address in token capabilities
func (v *MPCTokenValidator) validateVaultAddress(token *Token, expectedAddress string) error {
	for _, att := range token.Attenuations {
		if vaultCap, ok := att.Capability.(*VaultCapability); ok {
			if vaultCap.VaultAddress != "" && vaultCap.VaultAddress != expectedAddress {
				return fmt.Errorf("vault address mismatch: token=%s, expected=%s",
					vaultCap.VaultAddress, expectedAddress)
			}
		}
	}
	return nil
}

// validateMPCIssuer validates that the token issuer matches the MPC enclave
func (v *MPCTokenValidator) validateMPCIssuer(token *Token) error {
	expectedIssuer, _ := deriveIssuerDIDFromBytes(v.enclave.PubKeyBytes())

	if token.Issuer != expectedIssuer {
		return fmt.Errorf("token issuer does not match MPC enclave: token=%s, expected=%s",
			token.Issuer, expectedIssuer)
	}

	return nil
}

// createMPCVaultAttenuation creates MPC-specific vault attenuations
func createMPCVaultAttenuation(actions []string, enclaveDataCID, vaultAddress string) Attenuation {
	// Use the existing CreateVaultAttenuation function but add MPC-specific validation
	return CreateVaultAttenuation(actions, enclaveDataCID, vaultAddress)
}

// containsAdminAction checks if actions contain admin-level permissions
func containsAdminAction(actions []string) bool {
	adminActions := map[string]bool{
		"admin": true, "export": true, "import": true, "delete": true,
	}

	for _, action := range actions {
		if adminActions[action] {
			return true
		}
	}
	return false
}

// ValidateEnclaveDataIntegrity validates enclave data against IPFS CID
func ValidateEnclaveDataIntegrity(enclaveData *mpc.EnclaveData, expectedCID string) error {
	if enclaveData == nil {
		return fmt.Errorf("enclave data cannot be nil")
	}

	// Basic validation of enclave structure
	if len(enclaveData.PubBytes) == 0 {
		return fmt.Errorf("enclave public key bytes cannot be empty")
	}

	if enclaveData.PubHex == "" {
		return fmt.Errorf("enclave public key hex cannot be empty")
	}

	// Implement IPFS CID validation against enclave data hash
	// Serialize the enclave data for consistent hashing
	enclaveDataBytes, err := enclaveData.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal enclave data: %w", err)
	}

	// 1. Hash the enclave data using SHA-256
	hasher := sha256.New()
	hasher.Write(enclaveDataBytes)
	digest := hasher.Sum(nil)

	// 2. Create multihash with SHA-256 prefix
	mhash, err := multihash.EncodeName(digest, "sha2-256")
	if err != nil {
		return fmt.Errorf("failed to create multihash: %w", err)
	}

	// 3. Create CID and compare with expected
	parsedExpectedCID, err := cid.Parse(expectedCID)
	if err != nil {
		return fmt.Errorf("failed to parse expected CID: %w", err)
	}

	// Create CID v1 with dag-pb codec (IPFS default)
	calculatedCID := cid.NewCidV1(cid.DagProtobuf, mhash)

	// Compare CIDs
	if !parsedExpectedCID.Equals(calculatedCID) {
		return fmt.Errorf(
			"CID verification failed: expected %s, calculated %s",
			parsedExpectedCID.String(),
			calculatedCID.String(),
		)
	}

	return nil
}

// MPCCapabilityBuilder helps build MPC-specific capabilities
type MPCCapabilityBuilder struct {
	enclave mpc.Enclave
	builder *MPCTokenBuilder
}

// NewMPCCapabilityBuilder creates a new MPC capability builder
func NewMPCCapabilityBuilder(enclave mpc.Enclave) (*MPCCapabilityBuilder, error) {
	builder, err := NewMPCTokenBuilder(enclave)
	if err != nil {
		return nil, fmt.Errorf("failed to create MPC token builder: %w", err)
	}

	return &MPCCapabilityBuilder{
		enclave: enclave,
		builder: builder,
	}, nil
}

// CreateVaultAdminCapability creates admin-level vault capabilities
func (b *MPCCapabilityBuilder) CreateVaultAdminCapability(
	vaultAddress, enclaveDataCID string,
) Attenuation {
	allActions := []string{"read", "write", "sign", "export", "import", "delete", "admin"}
	return CreateVaultAttenuation(allActions, enclaveDataCID, vaultAddress)
}

// CreateVaultReadOnlyCapability creates read-only vault capabilities
func (b *MPCCapabilityBuilder) CreateVaultReadOnlyCapability(
	vaultAddress, enclaveDataCID string,
) Attenuation {
	readActions := []string{"read"}
	return CreateVaultAttenuation(readActions, enclaveDataCID, vaultAddress)
}

// CreateVaultSigningCapability creates signing-specific vault capabilities
func (b *MPCCapabilityBuilder) CreateVaultSigningCapability(
	vaultAddress, enclaveDataCID string,
) Attenuation {
	signActions := []string{"read", "sign"}
	return CreateVaultAttenuation(signActions, enclaveDataCID, vaultAddress)
}

// CreateCustomCapability creates a custom capability with specified actions
func (b *MPCCapabilityBuilder) CreateCustomCapability(
	actions []string,
	vaultAddress, enclaveDataCID string,
) Attenuation {
	return CreateVaultAttenuation(actions, enclaveDataCID, vaultAddress)
}

// Utility functions

// deriveIssuerDIDFromBytes creates issuer DID and address from public key bytes
// Enhanced version using the crypto/keys package
func deriveIssuerDIDFromBytes(pubKeyBytes []byte) (string, string) {
	// Use the enhanced NewFromMPCPubKey method from crypto/keys
	did, err := keys.NewFromMPCPubKey(pubKeyBytes)
	if err != nil {
		// Fallback to simplified implementation
		address := fmt.Sprintf("addr_%x", pubKeyBytes[:8])
		issuerDID := fmt.Sprintf("did:sonr:%s", address)
		return issuerDID, address
	}

	// Use the proper DID generation and address derivation
	didStr := did.String()
	address, err := did.Address()
	if err != nil {
		// Fallback to simplified address
		address = fmt.Sprintf("addr_%x", pubKeyBytes[:8])
	}

	return didStr, address
}
