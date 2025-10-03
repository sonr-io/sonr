package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// UCANDelegator handles UCAN token generation for OAuth flows
type UCANDelegator struct {
	scopeMapper *ScopeMapper
	signer      UCANSigner
}

// UCANSigner interface for signing UCAN tokens
type UCANSigner interface {
	Sign(token *ucan.Token) (string, error)
	GetIssuerDID() string
}

// NewUCANDelegator creates a new UCAN delegator
func NewUCANDelegator(signer UCANSigner) *UCANDelegator {
	if signer == nil {
		// Use blockchain signer by default
		signer, _ = NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	}
	return &UCANDelegator{
		scopeMapper: NewScopeMapper(),
		signer:      signer,
	}
}

// CreateDelegation creates a UCAN token for user-to-client delegation
func (d *UCANDelegator) CreateDelegation(
	userDID string,
	clientID string,
	scopes []string,
	expiresAt time.Time,
) (*ucan.Token, error) {
	// Build resource context for the user
	resourceContext := d.buildResourceContext(userDID)

	// Map OAuth scopes to UCAN attenuations
	attenuations := d.scopeMapper.MapToUCAN(scopes, userDID, clientID, resourceContext)
	if len(attenuations) == 0 {
		return nil, fmt.Errorf("no valid attenuations for scopes: %v", scopes)
	}

	// Create UCAN token
	token := &ucan.Token{
		Issuer:       userDID,
		Audience:     clientID,
		ExpiresAt:    expiresAt.Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Facts: []ucan.Fact{
			{
				Data: d.createOAuthFact(scopes, "user_delegation"),
			},
		},
	}

	// Sign the token
	signedToken, err := d.signToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to sign UCAN token: %w", err)
	}

	token.Raw = signedToken
	return token, nil
}

// CreateServiceDelegation creates a UCAN token for service-to-service delegation
func (d *UCANDelegator) CreateServiceDelegation(
	clientID string,
	scopes []string,
	expiresAt time.Time,
) (*ucan.Token, error) {
	// Service delegations use the OAuth provider as issuer
	issuerDID := d.signer.GetIssuerDID()

	// Build resource context for service
	resourceContext := map[string]string{
		"service_id": clientID,
		"type":       "service",
	}

	// Map OAuth scopes to UCAN attenuations
	attenuations := d.scopeMapper.MapToUCAN(scopes, issuerDID, clientID, resourceContext)
	if len(attenuations) == 0 {
		return nil, fmt.Errorf("no valid attenuations for scopes: %v", scopes)
	}

	// Create UCAN token
	token := &ucan.Token{
		Issuer:       issuerDID,
		Audience:     clientID,
		ExpiresAt:    expiresAt.Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Facts: []ucan.Fact{
			{
				Data: d.createOAuthFact(scopes, "service_delegation"),
			},
		},
	}

	// Sign the token
	signedToken, err := d.signToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to sign UCAN token: %w", err)
	}

	token.Raw = signedToken
	return token, nil
}

// CreateDelegationChain creates a chain of UCAN tokens for complex delegations
func (d *UCANDelegator) CreateDelegationChain(
	userDID string,
	intermediaries []string,
	finalAudience string,
	scopes []string,
	expiresAt time.Time,
) ([]*ucan.Token, error) {
	chain := make([]*ucan.Token, 0, len(intermediaries)+1)

	currentIssuer := userDID
	proofs := []ucan.Proof{}

	// Create delegation for each intermediary
	for _, intermediary := range intermediaries {
		token, err := d.createIntermediateDelegation(
			currentIssuer,
			intermediary,
			scopes,
			expiresAt,
			proofs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create intermediate delegation: %w", err)
		}

		chain = append(chain, token)
		proofs = append(proofs, ucan.Proof(token.Raw))
		currentIssuer = intermediary
	}

	// Create final delegation
	finalToken, err := d.createFinalDelegation(
		currentIssuer,
		finalAudience,
		scopes,
		expiresAt,
		proofs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create final delegation: %w", err)
	}

	chain = append(chain, finalToken)
	return chain, nil
}

// RevokeDelegation revokes a UCAN delegation
func (d *UCANDelegator) RevokeDelegation(tokenID string) error {
	// TODO: Implement delegation revocation
	// This would typically involve:
	// 1. Adding the token to a revocation list
	// 2. Publishing revocation to a public ledger or database
	// 3. Notifying relevant parties
	return fmt.Errorf("delegation revocation not yet implemented")
}

// ValidateDelegation validates a UCAN token delegation
func (d *UCANDelegator) ValidateDelegation(token *ucan.Token, requiredScopes []string) error {
	// Check expiration
	if time.Now().Unix() > token.ExpiresAt {
		return fmt.Errorf("token expired")
	}

	// Check not before
	if token.NotBefore > 0 && time.Now().Unix() < token.NotBefore {
		return fmt.Errorf("token not yet valid")
	}

	// Verify the token has required capabilities
	for _, scope := range requiredScopes {
		if !d.tokenGrantsScope(token, scope) {
			return fmt.Errorf("token does not grant required scope: %s", scope)
		}
	}

	// TODO: Verify signature
	// TODO: Verify delegation chain if proofs exist

	return nil
}

// Private helper methods

func (d *UCANDelegator) buildResourceContext(userDID string) map[string]string {
	// TODO: Fetch actual resource information for the user
	return map[string]string{
		"vault_address": fmt.Sprintf("vault:%s", userDID),
		"enclave_cid":   fmt.Sprintf("cid:%s", userDID),
		"dwn_id":        fmt.Sprintf("dwn:%s", userDID),
		"service_id":    fmt.Sprintf("service:%s", userDID),
	}
}

func (d *UCANDelegator) createOAuthFact(scopes []string, delegationType string) json.RawMessage {
	fact := map[string]any{
		"oauth_scopes":    scopes,
		"delegation_type": delegationType,
		"issued_at":       time.Now().Unix(),
		"oauth_version":   "2.0",
	}

	data, _ := json.Marshal(fact)
	return json.RawMessage(data)
}

func (d *UCANDelegator) signToken(token *ucan.Token) (string, error) {
	if d.signer == nil {
		return "", fmt.Errorf("no signer configured")
	}

	return d.signer.Sign(token)
}

func (d *UCANDelegator) createIntermediateDelegation(
	issuer, audience string,
	scopes []string,
	expiresAt time.Time,
	proofs []ucan.Proof,
) (*ucan.Token, error) {
	return d.createDelegationWithType(
		issuer,
		audience,
		scopes,
		expiresAt,
		proofs,
		"intermediate_delegation",
	)
}

func (d *UCANDelegator) createFinalDelegation(
	issuer, audience string,
	scopes []string,
	expiresAt time.Time,
	proofs []ucan.Proof,
) (*ucan.Token, error) {
	return d.createDelegationWithType(
		issuer,
		audience,
		scopes,
		expiresAt,
		proofs,
		"final_delegation",
	)
}

// createDelegationWithType creates a delegation token with a specific type
func (d *UCANDelegator) createDelegationWithType(
	issuer, audience string,
	scopes []string,
	expiresAt time.Time,
	proofs []ucan.Proof,
	delegationType string,
) (*ucan.Token, error) {
	resourceContext := d.buildResourceContext(issuer)
	attenuations := d.scopeMapper.MapToUCAN(scopes, issuer, audience, resourceContext)

	token := &ucan.Token{
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    expiresAt.Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Proofs:       proofs,
		Facts: []ucan.Fact{
			{
				Data: d.createOAuthFact(scopes, delegationType),
			},
		},
	}

	signedToken, err := d.signToken(token)
	if err != nil {
		return nil, err
	}

	token.Raw = signedToken
	return token, nil
}

func (d *UCANDelegator) tokenGrantsScope(token *ucan.Token, scope string) bool {
	// Get the scope definition
	scopeDef, exists := d.scopeMapper.GetScope(scope)
	if !exists {
		return false
	}

	// Check if any attenuation grants the required capabilities
	for _, attenuation := range token.Attenuations {
		if attenuation.Capability.Grants(scopeDef.UCANActions) {
			// Also check resource type matches
			if d.resourceMatches(attenuation.Resource, scopeDef.ResourceType) {
				return true
			}
		}
	}

	return false
}

func (d *UCANDelegator) resourceMatches(resource ucan.Resource, resourceType string) bool {
	// Simple scheme matching for now
	return resource.GetScheme() == resourceType || resource.GetScheme() == "*"
}
