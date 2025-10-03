package handlers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/did/types"
)

// BlockchainUCANSigner implements UCAN signing with blockchain keys
type BlockchainUCANSigner struct {
	didKeeper     DIDKeeperInterface
	issuerDID     string
	signingMethod jwt.SigningMethod
	privateKey    any
}

// DIDKeeperInterface defines the minimal interface needed from DID keeper
type DIDKeeperInterface interface {
	GetDIDDocument(ctx context.Context, did string) (*types.DIDDocument, error)
	GetVerificationMethod(
		ctx context.Context,
		did string,
		methodID string,
	) (*types.VerificationMethod, error)
}

// NewBlockchainUCANSigner creates a new blockchain-integrated UCAN signer
func NewBlockchainUCANSigner(
	didKeeper DIDKeeperInterface,
	issuerDID string,
) (*BlockchainUCANSigner, error) {
	return &BlockchainUCANSigner{
		didKeeper:     didKeeper,
		issuerDID:     issuerDID,
		signingMethod: jwt.SigningMethodES256, // Default to ES256
	}, nil
}

// Sign signs a UCAN token with blockchain keys
func (s *BlockchainUCANSigner) Sign(token *ucan.Token) (string, error) {
	// Build JWT claims from UCAN token
	claims := s.buildClaims(token)

	// Get signing key
	signingKey, method, err := s.getSigningKey(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get signing key: %w", err)
	}

	// Create JWT token
	jwtToken := jwt.NewWithClaims(method, claims)

	// Sign token
	signedToken, err := jwtToken.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// GetIssuerDID returns the issuer DID
func (s *BlockchainUCANSigner) GetIssuerDID() string {
	if s.issuerDID == "" {
		return "did:sonr:oauth-provider"
	}
	return s.issuerDID
}

// buildClaims builds JWT claims from UCAN token
func (s *BlockchainUCANSigner) buildClaims(token *ucan.Token) jwt.MapClaims {
	claims := jwt.MapClaims{
		"iss": token.Issuer,
		"aud": token.Audience,
		"exp": token.ExpiresAt,
		"iat": time.Now().Unix(),
		"nbf": token.NotBefore,
	}

	// Add attenuations
	if len(token.Attenuations) > 0 {
		attClaims := make([]map[string]any, len(token.Attenuations))
		for i, att := range token.Attenuations {
			attClaims[i] = s.serializeAttenuation(att)
		}
		claims["att"] = attClaims
	}

	// Add proofs if present
	if len(token.Proofs) > 0 {
		proofClaims := make([]string, len(token.Proofs))
		for i, proof := range token.Proofs {
			proofClaims[i] = string(proof)
		}
		claims["prf"] = proofClaims
	}

	// Add facts if present
	if len(token.Facts) > 0 {
		factClaims := make([]json.RawMessage, len(token.Facts))
		for i, fact := range token.Facts {
			factClaims[i] = fact.Data
		}
		claims["fct"] = factClaims
	}

	// Add UCAN version
	claims["ucv"] = "0.10.0"

	return claims
}

// serializeAttenuation serializes an attenuation for JWT claims
func (s *BlockchainUCANSigner) serializeAttenuation(att ucan.Attenuation) map[string]any {
	result := map[string]any{
		"with": att.Resource.GetURI(),
	}

	// Handle capability serialization
	actions := att.Capability.GetActions()
	if len(actions) == 1 {
		result["can"] = actions[0]
	} else {
		result["can"] = actions
	}

	// Add resource-specific metadata
	scheme := att.Resource.GetScheme()
	switch scheme {
	case "vault":
		result["type"] = "vault_operation"
	case "service", "svc":
		result["type"] = "service_operation"
	case "did":
		result["type"] = "identity_operation"
	case "dwn":
		result["type"] = "data_operation"
	case "dex", "pool":
		result["type"] = "trading_operation"
	}

	return result
}

// getSigningKey retrieves the signing key from blockchain
func (s *BlockchainUCANSigner) getSigningKey(ctx context.Context) (any, jwt.SigningMethod, error) {
	// If we have a cached private key, use it
	if s.privateKey != nil {
		return s.privateKey, s.signingMethod, nil
	}

	// For OAuth provider, use a service key
	if s.issuerDID == "did:sonr:oauth-provider" || s.issuerDID == "" {
		// Generate or retrieve service key
		privateKey, publicKey, err := s.generateServiceKey()
		if err != nil {
			return nil, nil, err
		}

		s.privateKey = privateKey
		s.signingMethod = jwt.SigningMethodEdDSA // Update to EdDSA for Ed25519 keys
		_ = publicKey                            // Store public key if needed

		return privateKey, jwt.SigningMethodEdDSA, nil
	}

	// For DID-based signing, retrieve from DID document
	if s.didKeeper != nil {
		didDoc, err := s.didKeeper.GetDIDDocument(ctx, s.issuerDID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get DID document: %w", err)
		}

		// Use the first verification method for signing
		if len(didDoc.VerificationMethod) > 0 {
			vm := didDoc.VerificationMethod[0]

			// Extract key based on verification method kind
			switch vm.VerificationMethodKind {
			case "Ed25519VerificationKey2020":
				// Extract Ed25519 key
				privateKey, err := s.extractEd25519Key(vm)
				if err != nil {
					return nil, nil, err
				}
				s.privateKey = privateKey
				s.signingMethod = jwt.SigningMethodEdDSA
				return privateKey, jwt.SigningMethodEdDSA, nil

			case "EcdsaSecp256k1VerificationKey2019":
				// Extract Secp256k1 key
				privateKey, err := s.extractSecp256k1Key(vm)
				if err != nil {
					return nil, nil, err
				}
				s.privateKey = privateKey
				s.signingMethod = jwt.SigningMethodES256
				return privateKey, jwt.SigningMethodES256, nil

			default:
				return nil, nil, fmt.Errorf(
					"unsupported verification method type: %s",
					vm.VerificationMethodKind,
				)
			}
		}
	}

	return nil, nil, fmt.Errorf("no signing key available")
}

// generateServiceKey generates a new service key pair
func (s *BlockchainUCANSigner) generateServiceKey() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// In production, this should retrieve from secure storage
	// For now, generate a new key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return privateKey, publicKey, nil
}

// extractEd25519Key extracts Ed25519 private key from verification method
func (s *BlockchainUCANSigner) extractEd25519Key(
	vm *types.VerificationMethod,
) (ed25519.PrivateKey, error) {
	// In production, this would retrieve the private key from secure storage
	// based on the public key in the verification method

	// For now, return error as we don't have access to private keys
	return nil, fmt.Errorf("private key retrieval not implemented")
}

// extractSecp256k1Key extracts Secp256k1 private key from verification method
func (s *BlockchainUCANSigner) extractSecp256k1Key(
	vm *types.VerificationMethod,
) (*secp256k1.PrivKey, error) {
	// In production, this would retrieve the private key from secure storage
	// based on the public key in the verification method

	// For now, return error as we don't have access to private keys
	return nil, fmt.Errorf("private key retrieval not implemented")
}

// VerifySignature verifies a UCAN token signature
func (s *BlockchainUCANSigner) VerifySignature(tokenString string) (*ucan.Token, error) {
	// Parse token without verification first to get issuer
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("no issuer in token")
	}

	// Get public key for issuer
	publicKey, err := s.getPublicKey(context.Background(), issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Verify token with public key
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert JWT claims to UCAN token
	return s.claimsToUCAN(claims, tokenString)
}

// getPublicKey retrieves public key for a DID
func (s *BlockchainUCANSigner) getPublicKey(ctx context.Context, did string) (any, error) {
	if did == "did:sonr:oauth-provider" {
		// Return service public key
		// In production, this would be retrieved from configuration
		return []byte("oauth-provider-public-key"), nil
	}

	if s.didKeeper != nil {
		didDoc, err := s.didKeeper.GetDIDDocument(ctx, did)
		if err != nil {
			return nil, fmt.Errorf("failed to get DID document: %w", err)
		}

		if len(didDoc.VerificationMethod) > 0 {
			vm := didDoc.VerificationMethod[0]

			// Extract public key based on verification method kind
			switch vm.VerificationMethodKind {
			case "Ed25519VerificationKey2020":
				// Decode base64 public key
				publicKeyBytes, err := base64.StdEncoding.DecodeString(vm.PublicKeyMultibase)
				if err != nil {
					return nil, err
				}
				return ed25519.PublicKey(publicKeyBytes), nil

			case "EcdsaSecp256k1VerificationKey2019":
				// Decode and return Secp256k1 public key
				publicKeyBytes, err := base64.StdEncoding.DecodeString(vm.PublicKeyMultibase)
				if err != nil {
					return nil, err
				}
				return publicKeyBytes, nil

			default:
				return nil, fmt.Errorf(
					"unsupported verification method type: %s",
					vm.VerificationMethodKind,
				)
			}
		}
	}

	return nil, fmt.Errorf("no public key found for DID: %s", did)
}

// claimsToUCAN converts JWT claims to UCAN token
func (s *BlockchainUCANSigner) claimsToUCAN(
	claims jwt.MapClaims,
	rawToken string,
) (*ucan.Token, error) {
	token := &ucan.Token{
		Raw: rawToken,
	}

	// Extract standard claims
	if iss, ok := claims["iss"].(string); ok {
		token.Issuer = iss
	}
	if aud, ok := claims["aud"].(string); ok {
		token.Audience = aud
	}
	if exp, ok := claims["exp"].(float64); ok {
		token.ExpiresAt = int64(exp)
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		token.NotBefore = int64(nbf)
	}

	// Extract attenuations
	if attClaims, ok := claims["att"].([]any); ok {
		attenuations := make([]ucan.Attenuation, 0, len(attClaims))
		for _, attItem := range attClaims {
			if attMap, ok := attItem.(map[string]any); ok {
				att, err := s.parseAttenuation(attMap)
				if err != nil {
					return nil, fmt.Errorf("failed to parse attenuation: %w", err)
				}
				attenuations = append(attenuations, att)
			}
		}
		token.Attenuations = attenuations
	}

	// Extract proofs
	if proofClaims, ok := claims["prf"].([]any); ok {
		proofs := make([]ucan.Proof, 0, len(proofClaims))
		for _, proofItem := range proofClaims {
			if proofStr, ok := proofItem.(string); ok {
				proofs = append(proofs, ucan.Proof(proofStr))
			}
		}
		token.Proofs = proofs
	}

	// Extract facts
	if factClaims, ok := claims["fct"].([]any); ok {
		facts := make([]ucan.Fact, 0, len(factClaims))
		for _, factItem := range factClaims {
			factData, _ := json.Marshal(factItem)
			facts = append(facts, ucan.Fact{
				Data: json.RawMessage(factData),
			})
		}
		token.Facts = facts
	}

	return token, nil
}

// parseAttenuation parses an attenuation from JWT claims
func (s *BlockchainUCANSigner) parseAttenuation(attMap map[string]any) (ucan.Attenuation, error) {
	// Extract resource URI
	resourceURI, ok := attMap["with"].(string)
	if !ok {
		return ucan.Attenuation{}, fmt.Errorf("missing resource URI")
	}

	// Parse resource
	resource := &SimpleResource{
		Scheme: "generic",
		Value:  resourceURI,
	}

	// Extract scheme from URI if possible
	if len(resourceURI) > 0 {
		for _, scheme := range []string{"vault:", "service:", "did:", "dwn:", "dex:", "pool:"} {
			if len(resourceURI) >= len(scheme) && resourceURI[:len(scheme)] == scheme {
				resource.Scheme = scheme[:len(scheme)-1]
				resource.Value = resourceURI[len(scheme):]
				break
			}
		}
	}

	// Extract capability
	var capability ucan.Capability
	switch can := attMap["can"].(type) {
	case string:
		capability = &ucan.SimpleCapability{Action: can}
	case []any:
		actions := make([]string, 0, len(can))
		for _, action := range can {
			if actionStr, ok := action.(string); ok {
				actions = append(actions, actionStr)
			}
		}
		capability = &ucan.MultiCapability{Actions: actions}
	default:
		return ucan.Attenuation{}, fmt.Errorf("invalid capability format")
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}, nil
}

// SetPrivateKey sets a private key for signing (for testing)
func (s *BlockchainUCANSigner) SetPrivateKey(privateKey any, method jwt.SigningMethod) {
	s.privateKey = privateKey
	s.signingMethod = method
}

// CreateDelegationToken creates a UCAN token for delegation
func (s *BlockchainUCANSigner) CreateDelegationToken(
	issuer, audience string,
	attenuations []ucan.Attenuation,
	proofs []ucan.Proof,
	expiresIn time.Duration,
) (string, error) {
	token := &ucan.Token{
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    time.Now().Add(expiresIn).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Proofs:       proofs,
	}

	return s.Sign(token)
}

// ValidateDelegationChain validates a chain of UCAN delegations
func (s *BlockchainUCANSigner) ValidateDelegationChain(tokens []string) error {
	if len(tokens) == 0 {
		return fmt.Errorf("empty delegation chain")
	}

	var previousToken *ucan.Token
	for i, tokenString := range tokens {
		// Verify current token
		token, err := s.VerifySignature(tokenString)
		if err != nil {
			return fmt.Errorf("failed to verify token %d: %w", i, err)
		}

		// Check expiration
		if time.Now().Unix() > token.ExpiresAt {
			return fmt.Errorf("token %d has expired", i)
		}

		// Check not before
		if token.NotBefore > 0 && time.Now().Unix() < token.NotBefore {
			return fmt.Errorf("token %d not yet valid", i)
		}

		// Validate delegation chain
		if previousToken != nil {
			// Check that previous token's audience matches current issuer
			if previousToken.Audience != token.Issuer {
				return fmt.Errorf("broken delegation chain at token %d", i)
			}

			// Check that capabilities are properly attenuated
			if !s.isProperlyAttenuated(previousToken.Attenuations, token.Attenuations) {
				return fmt.Errorf("improper attenuation at token %d", i)
			}
		}

		previousToken = token
	}

	return nil
}

// isProperlyAttenuated checks if child attenuations are properly attenuated from parent
func (s *BlockchainUCANSigner) isProperlyAttenuated(parent, child []ucan.Attenuation) bool {
	// Child cannot have more permissions than parent
	for _, childAtt := range child {
		found := false
		for _, parentAtt := range parent {
			// Check if child attenuation is covered by parent
			if s.attenuationCovers(parentAtt, childAtt) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// attenuationCovers checks if parent attenuation covers child
func (s *BlockchainUCANSigner) attenuationCovers(parent, child ucan.Attenuation) bool {
	// Check resource match
	if parent.Resource.GetScheme() != child.Resource.GetScheme() {
		// Wildcard scheme matches all
		if parent.Resource.GetScheme() != "*" {
			return false
		}
	}

	// Check capability coverage
	parentActions := parent.Capability.GetActions()
	childActions := child.Capability.GetActions()

	// If parent has wildcard, it covers everything
	for _, action := range parentActions {
		if action == "*" {
			return true
		}
	}

	// Check each child action is in parent
	for _, childAction := range childActions {
		found := false
		for _, parentAction := range parentActions {
			if parentAction == childAction {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// RefreshToken creates a new token from an existing one with updated expiration
func (s *BlockchainUCANSigner) RefreshToken(
	tokenString string,
	newExpiration time.Duration,
) (string, error) {
	// Verify existing token
	token, err := s.VerifySignature(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	// Create new token with same claims but new expiration
	newToken := &ucan.Token{
		Issuer:       token.Issuer,
		Audience:     token.Audience,
		ExpiresAt:    time.Now().Add(newExpiration).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: token.Attenuations,
		Proofs:       append(token.Proofs, ucan.Proof(tokenString)), // Add original as proof
		Facts:        token.Facts,
	}

	return s.Sign(newToken)
}
