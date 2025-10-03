// Package ucan provides User-Controlled Authorization Networks (UCAN) implementation
// for decentralized authorization and capability delegation in the Sonr network.
// This package handles JWT-based tokens, cryptographic verification, and resource capabilities.
package ucan

import (
	"context"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sonr-io/sonr/crypto/keys"
)

// Verifier provides UCAN token verification and validation functionality
type Verifier struct {
	didResolver DIDResolver
}

// DIDResolver resolves DID keys to public keys for signature verification
type DIDResolver interface {
	ResolveDIDKey(ctx context.Context, did string) (keys.DID, error)
}

// NewVerifier creates a new UCAN token verifier
func NewVerifier(didResolver DIDResolver) *Verifier {
	return &Verifier{
		didResolver: didResolver,
	}
}

// VerifyToken parses and verifies a UCAN JWT token
func (v *Verifier) VerifyToken(ctx context.Context, tokenString string) (*Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	// Parse the JWT token
	token, err := jwt.Parse(tokenString, v.keyFunc(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	// Parse UCAN-specific fields
	ucanToken, err := v.parseUCANClaims(claims, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UCAN claims: %w", err)
	}

	// Validate token structure
	if err := v.validateToken(ctx, ucanToken); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return ucanToken, nil
}

// VerifyCapability validates that a UCAN token grants specific capabilities
func (v *Verifier) VerifyCapability(
	ctx context.Context,
	tokenString string,
	resource string,
	abilities []string,
) (*Token, error) {
	token, err := v.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// Check if token grants required capabilities
	if err := v.checkCapabilities(token, resource, abilities); err != nil {
		return nil, fmt.Errorf("capability check failed: %w", err)
	}

	return token, nil
}

// VerifyDelegationChain validates the complete delegation chain of a UCAN token
func (v *Verifier) VerifyDelegationChain(ctx context.Context, tokenString string) error {
	token, err := v.VerifyToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("failed to verify root token: %w", err)
	}

	// Verify each proof in the delegation chain
	for i, proof := range token.Proofs {
		proofToken, err := v.VerifyToken(ctx, string(proof))
		if err != nil {
			return fmt.Errorf("failed to verify proof[%d] in delegation chain: %w", i, err)
		}

		// Validate delegation relationship
		if err := v.validateDelegation(token, proofToken); err != nil {
			return fmt.Errorf("invalid delegation at proof[%d]: %w", i, err)
		}
	}

	return nil
}

// keyFunc returns a function that resolves the signing key for JWT verification
func (v *Verifier) keyFunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		// Extract issuer from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid claims type")
		}

		issuer, ok := claims["iss"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid issuer claim")
		}

		// Resolve the issuer's DID to get public key
		did, err := v.didResolver.ResolveDIDKey(ctx, issuer)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve issuer DID: %w", err)
		}

		// Get verification key based on signing method
		switch token.Method {
		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
			return v.getRSAPublicKey(did)
		case jwt.SigningMethodEdDSA:
			return v.getEd25519PublicKey(did)
		default:
			return nil, fmt.Errorf("unsupported signing method: %v", token.Method)
		}
	}
}

// parseUCANClaims extracts UCAN-specific fields from JWT claims
func (v *Verifier) parseUCANClaims(claims jwt.MapClaims, raw string) (*Token, error) {
	issuer, audience := extractStandardClaims(claims)
	expiresAt, notBefore := extractTimeClaims(claims)

	attenuations, err := v.parseAttenuationsClaims(claims)
	if err != nil {
		return nil, err
	}

	proofs := parseProofsClaims(claims)
	facts := parseFactsClaims(claims)

	return &Token{
		Raw:          raw,
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    expiresAt,
		NotBefore:    notBefore,
		Attenuations: attenuations,
		Proofs:       proofs,
		Facts:        facts,
	}, nil
}

// extractStandardClaims extracts standard JWT claims (issuer and audience)
func extractStandardClaims(claims jwt.MapClaims) (string, string) {
	issuer, _ := claims["iss"].(string)
	audience, _ := claims["aud"].(string)
	return issuer, audience
}

// extractTimeClaims extracts time-related claims (exp and nbf)
func extractTimeClaims(claims jwt.MapClaims) (int64, int64) {
	var expiresAt, notBefore int64

	if exp, ok := claims["exp"]; ok {
		if expFloat, ok := exp.(float64); ok {
			expiresAt = int64(expFloat)
		}
	}

	if nbf, ok := claims["nbf"]; ok {
		if nbfFloat, ok := nbf.(float64); ok {
			notBefore = int64(nbfFloat)
		}
	}

	return expiresAt, notBefore
}

// parseAttenuationsClaims parses the attenuations from claims
func (v *Verifier) parseAttenuationsClaims(claims jwt.MapClaims) ([]Attenuation, error) {
	attClaims, ok := claims["att"]
	if !ok {
		return nil, nil
	}

	attSlice, ok := attClaims.([]any)
	if !ok {
		return nil, nil
	}

	// Pre-allocate slice with known capacity
	attenuations := make([]Attenuation, 0, len(attSlice))

	for _, attItem := range attSlice {
		attMap, ok := attItem.(map[string]any)
		if !ok {
			continue
		}

		att, err := v.parseAttenuation(attMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse attenuation: %w", err)
		}
		attenuations = append(attenuations, att)
	}

	return attenuations, nil
}

// parseProofsClaims parses the proofs from claims
func parseProofsClaims(claims jwt.MapClaims) []Proof {
	var proofs []Proof

	prfClaims, ok := claims["prf"]
	if !ok {
		return proofs
	}

	prfSlice, ok := prfClaims.([]any)
	if !ok {
		return proofs
	}

	for _, prfItem := range prfSlice {
		if prfStr, ok := prfItem.(string); ok {
			proofs = append(proofs, Proof(prfStr))
		}
	}

	return proofs
}

// parseFactsClaims parses the facts from claims
func parseFactsClaims(claims jwt.MapClaims) []Fact {
	fctClaims, ok := claims["fct"]
	if !ok {
		return nil
	}

	fctSlice, ok := fctClaims.([]any)
	if !ok {
		return nil
	}

	// Pre-allocate slice with known capacity
	facts := make([]Fact, 0, len(fctSlice))

	for _, fctItem := range fctSlice {
		factData, _ := json.Marshal(fctItem)
		facts = append(facts, Fact{Data: factData})
	}

	return facts
}

// parseAttenuation converts a map to an Attenuation struct with enhanced module-specific support
func (v *Verifier) parseAttenuation(attMap map[string]any) (Attenuation, error) {
	// Extract capability
	canValue, ok := attMap["can"]
	if !ok {
		return Attenuation{}, fmt.Errorf("missing 'can' field in attenuation")
	}

	// Extract resource
	withValue, ok := attMap["with"]
	if !ok {
		return Attenuation{}, fmt.Errorf("missing 'with' field in attenuation")
	}

	withStr, ok := withValue.(string)
	if !ok {
		return Attenuation{}, fmt.Errorf("'with' field must be a string")
	}

	// Parse resource first to determine module type
	resource, err := v.parseResource(withStr)
	if err != nil {
		return Attenuation{}, fmt.Errorf("failed to parse resource: %w", err)
	}

	// Create module-specific capability based on resource scheme
	cap, err := v.createModuleSpecificCapability(resource.GetScheme(), canValue, attMap)
	if err != nil {
		return Attenuation{}, fmt.Errorf("failed to create capability: %w", err)
	}

	return Attenuation{
		Capability: cap,
		Resource:   resource,
	}, nil
}

// createModuleSpecificCapability creates appropriate capability type based on module
func (v *Verifier) createModuleSpecificCapability(scheme string, canValue any, attMap map[string]any) (Capability, error) {
	// Extract common fields
	caveats := extractStringSliceFromMap(attMap, "caveats")
	metadata := extractStringMapFromMap(attMap, "metadata")

	switch scheme {
	case "did":
		return v.createDIDCapability(canValue, caveats, metadata)
	case "dwn":
		return v.createDWNCapability(canValue, caveats, metadata)
	case "service", "svc":
		return v.createServiceCapability(canValue, caveats, metadata)
	case "dex", "pool":
		return v.createDEXCapability(canValue, caveats, metadata, attMap)
	case "ipfs", "vault":
		// Handle existing vault capabilities
		return v.createVaultCapabilityFromMap(canValue, attMap)
	default:
		// Fallback to simple/multi capability for unknown schemes
		return v.createGenericCapability(canValue)
	}
}

// createDIDCapability creates a DID-specific capability
func (v *Verifier) createDIDCapability(canValue any, caveats []string, metadata map[string]string) (Capability, error) {
	switch canVal := canValue.(type) {
	case string:
		return &DIDCapability{
			Action:   canVal,
			Caveats:  caveats,
			Metadata: metadata,
		}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &DIDCapability{
			Actions:  actions,
			Caveats:  caveats,
			Metadata: metadata,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported DID capability type")
	}
}

// createDWNCapability creates a DWN-specific capability
func (v *Verifier) createDWNCapability(canValue any, caveats []string, metadata map[string]string) (Capability, error) {
	switch canVal := canValue.(type) {
	case string:
		return &DWNCapability{
			Action:   canVal,
			Caveats:  caveats,
			Metadata: metadata,
		}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &DWNCapability{
			Actions:  actions,
			Caveats:  caveats,
			Metadata: metadata,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported DWN capability type")
	}
}

// createServiceCapability creates a Service-specific capability
func (v *Verifier) createServiceCapability(canValue any, caveats []string, metadata map[string]string) (Capability, error) {
	// Service capabilities can still use MultiCapability for now
	switch canVal := canValue.(type) {
	case string:
		return &MultiCapability{Actions: []string{canVal}}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &MultiCapability{Actions: actions}, nil
	default:
		return nil, fmt.Errorf("unsupported Service capability type")
	}
}

// createDEXCapability creates a DEX-specific capability
func (v *Verifier) createDEXCapability(canValue any, caveats []string, metadata map[string]string, attMap map[string]any) (Capability, error) {
	maxAmount, _ := attMap["max_amount"].(string)

	switch canVal := canValue.(type) {
	case string:
		return &DEXCapability{
			Action:    canVal,
			Caveats:   caveats,
			MaxAmount: maxAmount,
			Metadata:  metadata,
		}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &DEXCapability{
			Actions:   actions,
			Caveats:   caveats,
			MaxAmount: maxAmount,
			Metadata:  metadata,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported DEX capability type")
	}
}

// createVaultCapabilityFromMap creates vault capability from existing logic
func (v *Verifier) createVaultCapabilityFromMap(canValue any, attMap map[string]any) (Capability, error) {
	// Use existing vault capability creation logic
	vaultAddress, _ := attMap["vault"].(string)
	caveats := extractStringSliceFromMap(attMap, "caveats")

	switch canVal := canValue.(type) {
	case string:
		return &VaultCapability{
			Action:       canVal,
			VaultAddress: vaultAddress,
			Caveats:      caveats,
		}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &VaultCapability{
			Actions:      actions,
			VaultAddress: vaultAddress,
			Caveats:      caveats,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported vault capability type")
	}
}

// createGenericCapability creates fallback capability for unknown schemes
func (v *Verifier) createGenericCapability(canValue any) (Capability, error) {
	switch canVal := canValue.(type) {
	case string:
		return &SimpleCapability{Action: canVal}, nil
	case []any:
		actions := extractStringSlice(canVal)
		return &MultiCapability{Actions: actions}, nil
	default:
		return nil, fmt.Errorf("unsupported capability type")
	}
}

// Helper functions for extracting data from maps
func extractStringSliceFromMap(m map[string]any, key string) []string {
	if value, exists := m[key]; exists {
		return extractStringSlice(value)
	}
	return nil
}

func extractStringMapFromMap(m map[string]any, key string) map[string]string {
	result := make(map[string]string)
	if value, exists := m[key]; exists {
		if mapValue, ok := value.(map[string]any); ok {
			for k, v := range mapValue {
				if strValue, ok := v.(string); ok {
					result[k] = strValue
				}
			}
		}
	}
	return result
}

// parseResource creates a Resource from a URI string
func (v *Verifier) parseResource(uri string) (Resource, error) {
	if uri == "" {
		return nil, fmt.Errorf("resource URI cannot be empty")
	}

	// Parse URI scheme and value - support both "scheme://value" and "scheme:value" formats
	var scheme, value string
	if strings.Contains(uri, "://") {
		parts := strings.SplitN(uri, "://", 2)
		if len(parts) == 2 {
			scheme = parts[0]
			value = parts[1]
		}
	} else if strings.Contains(uri, ":") {
		parts := strings.SplitN(uri, ":", 2)
		if len(parts) == 2 {
			scheme = parts[0]
			value = parts[1]
		}
	}

	if scheme == "" || value == "" {
		return nil, fmt.Errorf("invalid resource URI format: %s", uri)
	}

	return &SimpleResource{
		Scheme: scheme,
		Value:  value,
		URI:    uri,
	}, nil
}

// validateToken performs structural and temporal validation
func (v *Verifier) validateToken(_ context.Context, token *Token) error {
	// Check required fields
	if token.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if token.Audience == "" {
		return fmt.Errorf("audience is required")
	}
	if len(token.Attenuations) == 0 {
		return fmt.Errorf("at least one attenuation is required")
	}

	// Check temporal validity
	now := time.Now().Unix()

	if token.NotBefore > 0 && now < token.NotBefore {
		return fmt.Errorf("token is not yet valid (nbf: %d, now: %d)", token.NotBefore, now)
	}

	if token.ExpiresAt > 0 && now >= token.ExpiresAt {
		return fmt.Errorf("token has expired (exp: %d, now: %d)", token.ExpiresAt, now)
	}

	return nil
}

// checkCapabilities verifies that the token grants the required capabilities with enhanced module-specific validation
func (v *Verifier) checkCapabilities(token *Token, resource string, abilities []string) error {
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == resource {
			if att.Capability.Grants(abilities) {
				// Validate caveats for module-specific capabilities
				if err := v.validateCaveats(att.Capability, att.Resource); err != nil {
					return fmt.Errorf("caveat validation failed: %w", err)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("required capabilities not granted for resource %s", resource)
}

// validateCaveats validates constraints (caveats) for module-specific capabilities
func (v *Verifier) validateCaveats(cap Capability, resource Resource) error {
	scheme := resource.GetScheme()

	switch scheme {
	case "did":
		return v.validateDIDCaveats(cap, resource)
	case "dwn":
		return v.validateDWNCaveats(cap, resource)
	case "dex", "pool":
		return v.validateDEXCaveats(cap, resource)
	case "service", "svc":
		return v.validateServiceCaveats(cap, resource)
	case "vault", "ipfs":
		return v.validateVaultCaveats(cap, resource)
	default:
		return nil // No caveat validation for unknown schemes
	}
}

// validateDIDCaveats validates DID-specific constraints
func (v *Verifier) validateDIDCaveats(cap Capability, resource Resource) error {
	didCap, ok := cap.(*DIDCapability)
	if !ok {
		return nil // Not a DID capability
	}

	for _, caveat := range didCap.Caveats {
		switch caveat {
		case "owner":
			// Validate that the capability is for the owner's DID
			if err := v.validateOwnerCaveat(resource); err != nil {
				return fmt.Errorf("owner caveat validation failed: %w", err)
			}
		case "controller":
			// Validate controller permissions
			if err := v.validateControllerCaveat(resource); err != nil {
				return fmt.Errorf("controller caveat validation failed: %w", err)
			}
		}
	}
	return nil
}

// validateDWNCaveats validates DWN-specific constraints
func (v *Verifier) validateDWNCaveats(cap Capability, resource Resource) error {
	dwnCap, ok := cap.(*DWNCapability)
	if !ok {
		return nil // Not a DWN capability
	}

	for _, caveat := range dwnCap.Caveats {
		switch caveat {
		case "owner":
			// Validate record ownership
			if err := v.validateRecordOwnership(resource); err != nil {
				return fmt.Errorf("record ownership validation failed: %w", err)
			}
		case "protocol":
			// Validate protocol compliance
			if err := v.validateProtocolCaveat(resource); err != nil {
				return fmt.Errorf("protocol caveat validation failed: %w", err)
			}
		}
	}
	return nil
}

// validateDEXCaveats validates DEX-specific constraints
func (v *Verifier) validateDEXCaveats(cap Capability, resource Resource) error {
	dexCap, ok := cap.(*DEXCapability)
	if !ok {
		return nil // Not a DEX capability
	}

	for _, caveat := range dexCap.Caveats {
		switch caveat {
		case "max-amount":
			// Validate maximum swap amount
			if dexCap.MaxAmount != "" {
				if err := v.validateMaxAmountCaveat(dexCap.MaxAmount); err != nil {
					return fmt.Errorf("max amount caveat validation failed: %w", err)
				}
			}
		case "pool-member":
			// Validate pool membership
			if err := v.validatePoolMembershipCaveat(resource); err != nil {
				return fmt.Errorf("pool membership validation failed: %w", err)
			}
		}
	}
	return nil
}

// validateServiceCaveats validates Service-specific constraints
func (v *Verifier) validateServiceCaveats(cap Capability, resource Resource) error {
	// Service capabilities use MultiCapability for now
	// Add service-specific caveat validation if needed
	return nil
}

// validateVaultCaveats validates Vault-specific constraints
func (v *Verifier) validateVaultCaveats(cap Capability, resource Resource) error {
	vaultCap, ok := cap.(*VaultCapability)
	if !ok {
		return nil // Not a vault capability
	}

	for _, caveat := range vaultCap.Caveats {
		switch caveat {
		case "vault-owner":
			// Validate vault ownership
			if err := v.validateVaultOwnership(vaultCap.VaultAddress); err != nil {
				return fmt.Errorf("vault ownership validation failed: %w", err)
			}
		case "enclave-integrity":
			// Validate enclave data integrity
			if err := v.validateEnclaveIntegrity(vaultCap.EnclaveDataCID); err != nil {
				return fmt.Errorf("enclave integrity validation failed: %w", err)
			}
		}
	}
	return nil
}

// Caveat validation helper methods (placeholders for actual implementation)

// validateOwnerCaveat validates DID ownership constraint
func (v *Verifier) validateOwnerCaveat(resource Resource) error {
	// Placeholder: Implement actual DID ownership validation
	return nil
}

// validateControllerCaveat validates DID controller constraint
func (v *Verifier) validateControllerCaveat(resource Resource) error {
	// Placeholder: Implement actual controller validation
	return nil
}

// validateRecordOwnership validates DWN record ownership
func (v *Verifier) validateRecordOwnership(resource Resource) error {
	// Placeholder: Implement actual record ownership validation
	return nil
}

// validateProtocolCaveat validates DWN protocol constraint
func (v *Verifier) validateProtocolCaveat(resource Resource) error {
	// Placeholder: Implement actual protocol validation
	return nil
}

// validateMaxAmountCaveat validates DEX maximum amount constraint
func (v *Verifier) validateMaxAmountCaveat(maxAmount string) error {
	// Placeholder: Implement actual amount validation
	return nil
}

// validatePoolMembershipCaveat validates DEX pool membership
func (v *Verifier) validatePoolMembershipCaveat(resource Resource) error {
	// Placeholder: Implement actual pool membership validation
	return nil
}

// validateVaultOwnership validates vault ownership
func (v *Verifier) validateVaultOwnership(vaultAddress string) error {
	// Placeholder: Implement actual vault ownership validation
	return nil
}

// validateEnclaveIntegrity validates enclave data integrity
func (v *Verifier) validateEnclaveIntegrity(enclaveDataCID string) error {
	// Placeholder: Implement actual enclave integrity validation
	return nil
}

// validateDelegation checks that child token is properly attenuated from parent with enhanced module-specific validation
func (v *Verifier) validateDelegation(child, parent *Token) error {
	// Child's issuer must be parent's audience
	if child.Issuer != parent.Audience {
		return fmt.Errorf("delegation chain broken: child issuer must be parent audience")
	}

	// Child capabilities must be subset of parent with module-specific validation
	for _, childAtt := range child.Attenuations {
		if !v.isModuleCapabilitySubset(childAtt, parent.Attenuations) {
			return fmt.Errorf("child capability exceeds parent capabilities")
		}
	}

	// Child expiration must not exceed parent
	if parent.ExpiresAt > 0 && (child.ExpiresAt == 0 || child.ExpiresAt > parent.ExpiresAt) {
		return fmt.Errorf("child token expires after parent token")
	}

	// Validate cross-module delegation constraints
	if err := v.validateCrossModuleDelegation(child, parent); err != nil {
		return fmt.Errorf("cross-module delegation validation failed: %w", err)
	}

	return nil
}

// isModuleCapabilitySubset checks if a capability is a subset with module-specific logic
func (v *Verifier) isModuleCapabilitySubset(childAtt Attenuation, parentAtts []Attenuation) bool {
	for _, parentAtt := range parentAtts {
		if childAtt.Resource.GetURI() == parentAtt.Resource.GetURI() {
			if v.isModuleCapabilityContained(childAtt.Capability, parentAtt.Capability, childAtt.Resource.GetScheme()) {
				return true
			}
		}
	}
	return false
}

// isModuleCapabilityContained checks containment with module-specific logic
func (v *Verifier) isModuleCapabilityContained(child, parent Capability, scheme string) bool {
	// First check basic containment
	if parent.Contains(child) {
		// Additional module-specific containment validation
		switch scheme {
		case "did":
			return v.validateDIDContainment(child, parent)
		case "dwn":
			return v.validateDWNContainment(child, parent)
		case "dex", "pool":
			return v.validateDEXContainment(child, parent)
		case "vault", "ipfs":
			return v.validateVaultContainment(child, parent)
		default:
			return true // Basic containment is sufficient for unknown schemes
		}
	}
	return false
}

// validateCrossModuleDelegation validates constraints across different modules
func (v *Verifier) validateCrossModuleDelegation(child, parent *Token) error {
	childModules := v.extractModulesFromToken(child)
	parentModules := v.extractModulesFromToken(parent)

	// Check if child uses modules not present in parent
	for module := range childModules {
		if _, exists := parentModules[module]; !exists {
			return fmt.Errorf("child token uses module '%s' not delegated by parent", module)
		}
	}

	// Validate specific cross-module constraints
	return v.validateSpecificCrossModuleConstraints(child, parent)
}

// extractModulesFromToken extracts the modules used by a token
func (v *Verifier) extractModulesFromToken(token *Token) map[string]bool {
	modules := make(map[string]bool)
	for _, att := range token.Attenuations {
		scheme := att.Resource.GetScheme()
		modules[scheme] = true
	}
	return modules
}

// validateSpecificCrossModuleConstraints validates specific cross-module business logic
func (v *Verifier) validateSpecificCrossModuleConstraints(child, parent *Token) error {
	// Example: If DID operations require vault access, ensure both are present
	childHasDID := v.tokenHasModule(child, "did")
	childHasVault := v.tokenHasModule(child, "vault") || v.tokenHasModule(child, "ipfs")

	if childHasDID && !childHasVault {
		// Check if parent has vault capability that can be inherited
		parentHasVault := v.tokenHasModule(parent, "vault") || v.tokenHasModule(parent, "ipfs")
		if !parentHasVault {
			return fmt.Errorf("DID operations require vault access which is not available in delegation chain")
		}
	}

	// Add more cross-module constraints as needed
	return nil
}

// tokenHasModule checks if a token has capabilities for a specific module
func (v *Verifier) tokenHasModule(token *Token, module string) bool {
	for _, att := range token.Attenuations {
		if att.Resource.GetScheme() == module {
			return true
		}
	}
	return false
}

// Module-specific containment validation methods

// validateDIDContainment validates DID capability containment
func (v *Verifier) validateDIDContainment(child, parent Capability) bool {
	childDID, childOk := child.(*DIDCapability)
	parentDID, parentOk := parent.(*DIDCapability)

	if !childOk || !parentOk {
		return true // Not both DID capabilities, basic containment applies
	}

	// Validate that child caveats are more restrictive or equal
	return v.areCaveatsMoreRestrictive(childDID.Caveats, parentDID.Caveats)
}

// validateDWNContainment validates DWN capability containment
func (v *Verifier) validateDWNContainment(child, parent Capability) bool {
	childDWN, childOk := child.(*DWNCapability)
	parentDWN, parentOk := parent.(*DWNCapability)

	if !childOk || !parentOk {
		return true // Not both DWN capabilities, basic containment applies
	}

	// Validate that child caveats are more restrictive or equal
	return v.areCaveatsMoreRestrictive(childDWN.Caveats, parentDWN.Caveats)
}

// validateDEXContainment validates DEX capability containment
func (v *Verifier) validateDEXContainment(child, parent Capability) bool {
	childDEX, childOk := child.(*DEXCapability)
	parentDEX, parentOk := parent.(*DEXCapability)

	if !childOk || !parentOk {
		return true // Not both DEX capabilities, basic containment applies
	}

	// Validate max amount restriction
	if parentDEX.MaxAmount != "" && childDEX.MaxAmount != "" {
		// Child max amount should be less than or equal to parent
		if !v.isAmountLessOrEqual(childDEX.MaxAmount, parentDEX.MaxAmount) {
			return false
		}
	} else if parentDEX.MaxAmount != "" && childDEX.MaxAmount == "" {
		// Child must have max amount if parent does
		return false
	}

	// Validate that child caveats are more restrictive or equal
	return v.areCaveatsMoreRestrictive(childDEX.Caveats, parentDEX.Caveats)
}

// validateVaultContainment validates Vault capability containment
func (v *Verifier) validateVaultContainment(child, parent Capability) bool {
	childVault, childOk := child.(*VaultCapability)
	parentVault, parentOk := parent.(*VaultCapability)

	if !childOk || !parentOk {
		return true // Not both Vault capabilities, basic containment applies
	}

	// Vault address must match
	if childVault.VaultAddress != parentVault.VaultAddress {
		return false
	}

	// Validate that child caveats are more restrictive or equal
	return v.areCaveatsMoreRestrictive(childVault.Caveats, parentVault.Caveats)
}

// Helper methods for containment validation

// areCaveatsMoreRestrictive checks if child caveats are more restrictive than parent
func (v *Verifier) areCaveatsMoreRestrictive(childCaveats, parentCaveats []string) bool {
	parentCaveatSet := make(map[string]bool)
	for _, caveat := range parentCaveats {
		parentCaveatSet[caveat] = true
	}

	// All child caveats must be present in parent caveats (or child can have additional restrictions)
	for _, childCaveat := range childCaveats {
		if !parentCaveatSet[childCaveat] {
			// Child has additional restrictions, which is allowed
			continue
		}
	}

	return true
}

// isAmountLessOrEqual compares two amount strings (placeholder implementation)
func (v *Verifier) isAmountLessOrEqual(childAmount, parentAmount string) bool {
	// Placeholder: Implement actual amount comparison
	// This would parse the amounts and compare them numerically
	return true
}

// isCapabilitySubset checks if a capability is a subset of any parent capabilities
func (v *Verifier) isCapabilitySubset(childAtt Attenuation, parentAtts []Attenuation) bool {
	for _, parentAtt := range parentAtts {
		if childAtt.Resource.GetURI() == parentAtt.Resource.GetURI() {
			if parentAtt.Capability.Contains(childAtt.Capability) {
				return true
			}
		}
	}
	return false
}

// getRSAPublicKey extracts RSA public key from DID
func (v *Verifier) getRSAPublicKey(did keys.DID) (*rsa.PublicKey, error) {
	verifyKey, err := did.VerifyKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get verify key: %w", err)
	}

	rsaKey, ok := verifyKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("DID does not contain RSA public key")
	}

	return rsaKey, nil
}

// getEd25519PublicKey extracts Ed25519 public key from DID
func (v *Verifier) getEd25519PublicKey(did keys.DID) (ed25519.PublicKey, error) {
	pubKey := did.PublicKey()
	rawBytes, err := pubKey.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw public key: %w", err)
	}

	if pubKey.Type() != crypto.Ed25519 {
		return nil, fmt.Errorf("DID does not contain Ed25519 public key")
	}

	return ed25519.PublicKey(rawBytes), nil
}

// StringDIDResolver implements DIDResolver for did:key strings
type StringDIDResolver struct{}

// ResolveDIDKey extracts a public key from a did:key string
func (StringDIDResolver) ResolveDIDKey(ctx context.Context, didStr string) (keys.DID, error) {
	return keys.Parse(didStr)
}
