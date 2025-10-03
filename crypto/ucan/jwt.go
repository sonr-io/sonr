package ucan

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// StandardTemplate provides default authorization template
	StandardTemplate = NewCapabilityTemplate()

	// Revoked tokens tracking
	revokedTokens = make(map[string]bool)
)

func init() {
	// Setup standard templates with module-specific capabilities
	StandardTemplate.AddAllowedActions(
		"vault",
		[]string{"read", "write", "sign", "export", "import", "delete", "*"},
	)
	StandardTemplate.AddAllowedActions(
		"service",
		[]string{"read", "write", "register", "update", "delete"},
	)
	StandardTemplate.AddAllowedActions(
		"did",
		[]string{
			"create", "register", "update", "deactivate", "revoke",
			"add-verification-method", "remove-verification-method",
			"add-service", "remove-service", "issue-credential",
			"revoke-credential", "link-wallet", "register-webauthn", "*",
		},
	)
	StandardTemplate.AddAllowedActions(
		"dwn",
		[]string{
			"records-write", "records-delete", "protocols-configure",
			"permissions-grant", "permissions-revoke", "create", "read",
			"update", "delete", "*",
		},
	)
	StandardTemplate.AddAllowedActions(
		"dex",
		[]string{
			"register-account", "swap", "provide-liquidity", "remove-liquidity",
			"create-limit-order", "cancel-order", "*",
		},
	)
	StandardTemplate.AddAllowedActions(
		"pool",
		[]string{"swap", "provide-liquidity", "remove-liquidity", "*"},
	)
	StandardTemplate.AddAllowedActions(
		"svc",
		[]string{"register", "verify-domain", "delegate", "*"},
	)
}

// GenerateJWTToken creates a UCAN JWT token with given capability and expiration
func GenerateJWTToken(attenuation Attenuation, duration time.Duration) (string, error) {
	// Default expiration handling
	if duration == 0 {
		duration = 24 * time.Hour
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": "did:sonr:local", // Default issuer
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}

	// Add capability to claims - separate resource and capability
	capabilityBytes, err := json.Marshal(map[string]any{
		"can":  attenuation.Capability,
		"with": attenuation.Resource,
	})
	if err != nil {
		return "", fmt.Errorf("failed to serialize capability: %v", err)
	}
	claims["can"] = base64.URLEncoding.EncodeToString(capabilityBytes)

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Dummy secret for signing - in real-world, use proper key management
	tokenString, err := token.SignedString([]byte("sonr-ucan-secret"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

// VerifyJWTToken validates and parses a UCAN JWT token
func VerifyJWTToken(tokenString string) (*Token, error) {
	// Check if token is revoked
	if revokedTokens[tokenString] {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Parse token with custom claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Dummy secret verification - replace with proper key validation
		return []byte("sonr-ucan-secret"), nil
	}, jwt.WithLeeway(5*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Manual expiration check
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("no expiration time found")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("token has expired")
	}

	// Decode capability
	capabilityStr, ok := claims["can"].(string)
	if !ok {
		return nil, fmt.Errorf("no capability found in token")
	}

	capabilityBytes, err := base64.URLEncoding.DecodeString(capabilityStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode capability: %v", err)
	}

	// Parse capability and resource separately
	var capabilityMap map[string]any
	err = json.Unmarshal(capabilityBytes, &capabilityMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse capability: %v", err)
	}

	// Determine capability type
	var capability Capability
	var capData map[string]any
	switch v := capabilityMap["can"].(type) {
	case map[string]any:
		capData = v
	case string:
		// If it's a string, assume it's a simple action
		capability = &SimpleCapability{Action: v}
		capData = nil
	default:
		return nil, fmt.Errorf("invalid capability structure")
	}

	// Parse capability if needed
	if capData != nil {
		// Attempt to infer capability type
		if actions, ok := capData["actions"].([]any); ok {
			// MultiCapability
			stringActions := make([]string, len(actions))
			for i, action := range actions {
				if str, ok := action.(string); ok {
					stringActions[i] = str
				}
			}
			capability = &MultiCapability{Actions: stringActions}
		} else if action, ok := capData["action"].(string); ok {
			// SingleCapability
			capability = &SimpleCapability{Action: action}
		} else {
			return nil, fmt.Errorf("unable to parse capability type")
		}
	}

	// Parse resource
	var resourceData map[string]any
	switch resource := capabilityMap["with"].(type) {
	case map[string]any:
		resourceData = resource
	case string:
		// If it's a string, assume it's a simple URI
		resourceData = map[string]any{
			"Scheme": "generic",
			"Value":  resource,
			"URI":    resource,
		}
	default:
		return nil, fmt.Errorf("invalid resource structure")
	}

	// Create resource based on scheme
	scheme, _ := resourceData["Scheme"].(string)
	value, _ := resourceData["Value"].(string)
	uri, _ := resourceData["URI"].(string)

	resource := &SimpleResource{
		Scheme: scheme,
		Value:  value,
		URI:    uri,
	}

	// Validate attenuation
	attenuation := Attenuation{
		Capability: capability,
		Resource:   resource,
	}

	// Use standard template to validate
	err = StandardTemplate.ValidateAttenuation(attenuation)
	if err != nil {
		return nil, fmt.Errorf("capability validation failed: %v", err)
	}

	// Construct Token object
	parsedToken := &Token{
		Raw:          tokenString,
		Issuer:       claims["iss"].(string),
		ExpiresAt:    int64(exp),
		Attenuations: []Attenuation{attenuation},
	}

	return parsedToken, nil
}

// RevokeCapability adds a capability to the revocation list
func RevokeCapability(attenuation Attenuation) error {
	// Generate token to get its string representation
	token, err := GenerateJWTToken(attenuation, time.Hour)
	if err != nil {
		return err
	}

	// Add to revoked tokens
	revokedTokens[token] = true
	return nil
}

// NewCapability is a helper function to create a basic capability
func NewCapability(issuer, resource string, abilities []string) (Attenuation, error) {
	capability := &MultiCapability{Actions: abilities}
	resourceObj := &SimpleResource{
		Scheme: "generic",
		Value:  resource,
		URI:    resource,
	}

	return Attenuation{
		Capability: capability,
		Resource:   resourceObj,
	}, nil
}

// Enhanced JWT generation functions for module-specific capabilities

// GenerateModuleJWTToken creates a UCAN JWT token with module-specific capabilities
func GenerateModuleJWTToken(attenuations []Attenuation, issuer, audience string, duration time.Duration) (string, error) {
	if duration == 0 {
		duration = 24 * time.Hour
	}

	// Create JWT claims with enhanced structure
	claims := jwt.MapClaims{
		"iss": issuer,
		"aud": audience,
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	// Add attenuations to claims with module-specific serialization
	attClaims := make([]map[string]any, len(attenuations))
	for i, att := range attenuations {
		attMap, err := serializeModuleAttenuation(att)
		if err != nil {
			return "", fmt.Errorf("failed to serialize attenuation %d: %w", i, err)
		}
		attClaims[i] = attMap
	}
	claims["att"] = attClaims

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("sonr-ucan-secret"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// serializeModuleAttenuation serializes an attenuation based on its module type
func serializeModuleAttenuation(att Attenuation) (map[string]any, error) {
	attMap := map[string]any{
		"with": att.Resource.GetURI(),
	}

	scheme := att.Resource.GetScheme()
	switch scheme {
	case "did":
		return serializeDIDAttenuation(att, attMap)
	case "dwn":
		return serializeDWNAttenuation(att, attMap)
	case "dex", "pool":
		return serializeDEXAttenuation(att, attMap)
	case "service", "svc":
		return serializeServiceAttenuation(att, attMap)
	case "vault", "ipfs":
		return serializeVaultAttenuation(att, attMap)
	default:
		return serializeGenericAttenuation(att, attMap)
	}
}

// serializeDIDAttenuation serializes DID-specific attenuations
func serializeDIDAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	didCap, ok := att.Capability.(*DIDCapability)
	if !ok {
		return serializeGenericAttenuation(att, attMap)
	}

	if didCap.Action != "" {
		attMap["can"] = didCap.Action
	} else {
		attMap["can"] = didCap.Actions
	}

	if len(didCap.Caveats) > 0 {
		attMap["caveats"] = didCap.Caveats
	}
	if len(didCap.Metadata) > 0 {
		attMap["metadata"] = didCap.Metadata
	}

	return attMap, nil
}

// serializeDWNAttenuation serializes DWN-specific attenuations
func serializeDWNAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	dwnCap, ok := att.Capability.(*DWNCapability)
	if !ok {
		return serializeGenericAttenuation(att, attMap)
	}

	if dwnCap.Action != "" {
		attMap["can"] = dwnCap.Action
	} else {
		attMap["can"] = dwnCap.Actions
	}

	if len(dwnCap.Caveats) > 0 {
		attMap["caveats"] = dwnCap.Caveats
	}
	if len(dwnCap.Metadata) > 0 {
		attMap["metadata"] = dwnCap.Metadata
	}

	// Add DWN-specific fields
	if dwnRes, ok := att.Resource.(*DWNResource); ok {
		if dwnRes.RecordType != "" {
			attMap["record_type"] = dwnRes.RecordType
		}
		if dwnRes.Protocol != "" {
			attMap["protocol"] = dwnRes.Protocol
		}
		if dwnRes.Owner != "" {
			attMap["owner"] = dwnRes.Owner
		}
	}

	return attMap, nil
}

// serializeDEXAttenuation serializes DEX-specific attenuations
func serializeDEXAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	dexCap, ok := att.Capability.(*DEXCapability)
	if !ok {
		return serializeGenericAttenuation(att, attMap)
	}

	if dexCap.Action != "" {
		attMap["can"] = dexCap.Action
	} else {
		attMap["can"] = dexCap.Actions
	}

	if len(dexCap.Caveats) > 0 {
		attMap["caveats"] = dexCap.Caveats
	}
	if dexCap.MaxAmount != "" {
		attMap["max_amount"] = dexCap.MaxAmount
	}
	if len(dexCap.Metadata) > 0 {
		attMap["metadata"] = dexCap.Metadata
	}

	// Add DEX-specific fields
	if dexRes, ok := att.Resource.(*DEXResource); ok {
		if dexRes.PoolID != "" {
			attMap["pool_id"] = dexRes.PoolID
		}
		if dexRes.AssetPair != "" {
			attMap["asset_pair"] = dexRes.AssetPair
		}
		if dexRes.OrderID != "" {
			attMap["order_id"] = dexRes.OrderID
		}
	}

	return attMap, nil
}

// serializeServiceAttenuation serializes Service-specific attenuations
func serializeServiceAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	// Service capabilities still use MultiCapability
	multiCap, ok := att.Capability.(*MultiCapability)
	if !ok {
		return serializeGenericAttenuation(att, attMap)
	}

	attMap["can"] = multiCap.Actions

	// Add service-specific fields
	if svcRes, ok := att.Resource.(*ServiceResource); ok {
		if svcRes.ServiceID != "" {
			attMap["service_id"] = svcRes.ServiceID
		}
		if svcRes.Domain != "" {
			attMap["domain"] = svcRes.Domain
		}
		if len(svcRes.Metadata) > 0 {
			attMap["metadata"] = svcRes.Metadata
		}
	}

	return attMap, nil
}

// serializeVaultAttenuation serializes Vault-specific attenuations
func serializeVaultAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	vaultCap, ok := att.Capability.(*VaultCapability)
	if !ok {
		return serializeGenericAttenuation(att, attMap)
	}

	if vaultCap.Action != "" {
		attMap["can"] = vaultCap.Action
	} else {
		attMap["can"] = vaultCap.Actions
	}

	if vaultCap.VaultAddress != "" {
		attMap["vault"] = vaultCap.VaultAddress
	}
	if len(vaultCap.Caveats) > 0 {
		attMap["caveats"] = vaultCap.Caveats
	}
	if vaultCap.EnclaveDataCID != "" {
		attMap["enclave_data_cid"] = vaultCap.EnclaveDataCID
	}
	if len(vaultCap.Metadata) > 0 {
		attMap["metadata"] = vaultCap.Metadata
	}

	return attMap, nil
}

// serializeGenericAttenuation serializes generic attenuations
func serializeGenericAttenuation(att Attenuation, attMap map[string]any) (map[string]any, error) {
	actions := att.Capability.GetActions()
	if len(actions) == 1 {
		attMap["can"] = actions[0]
	} else {
		attMap["can"] = actions
	}
	return attMap, nil
}

// Enhanced verification with module-specific support

// VerifyModuleJWTToken validates and parses a UCAN JWT token with module-specific capabilities
func VerifyModuleJWTToken(tokenString string, expectedIssuer, expectedAudience string) (*Token, error) {
	// Check if token is revoked
	if revokedTokens[tokenString] {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Parse token with custom claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Dummy secret verification - replace with proper key validation
		return []byte("sonr-ucan-secret"), nil
	}, jwt.WithLeeway(5*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer and audience if provided
	if expectedIssuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != expectedIssuer {
			return nil, fmt.Errorf("invalid issuer: expected %s", expectedIssuer)
		}
	}
	if expectedAudience != "" {
		if aud, ok := claims["aud"].(string); !ok || aud != expectedAudience {
			return nil, fmt.Errorf("invalid audience: expected %s", expectedAudience)
		}
	}

	// Manual expiration check
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("no expiration time found")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("token has expired")
	}

	// Parse attenuations with module-specific support
	attenuations, err := parseEnhancedAttenuations(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attenuations: %w", err)
	}

	// Validate attenuations against templates
	for _, att := range attenuations {
		if err := StandardTemplate.ValidateAttenuation(att); err != nil {
			return nil, fmt.Errorf("capability validation failed: %w", err)
		}
	}

	// Construct Token object
	issuer, _ := claims["iss"].(string)
	audience, _ := claims["aud"].(string)
	nbf, _ := claims["nbf"].(float64)

	parsedToken := &Token{
		Raw:          tokenString,
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    int64(exp),
		NotBefore:    int64(nbf),
		Attenuations: attenuations,
	}

	return parsedToken, nil
}

// parseEnhancedAttenuations parses attenuations with module-specific capabilities
func parseEnhancedAttenuations(claims jwt.MapClaims) ([]Attenuation, error) {
	attClaims, ok := claims["att"]
	if !ok {
		return nil, fmt.Errorf("no attenuations found in token")
	}

	attSlice, ok := attClaims.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid attenuations format")
	}

	attenuations := make([]Attenuation, 0, len(attSlice))
	for i, attItem := range attSlice {
		attMap, ok := attItem.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid attenuation %d format", i)
		}

		att, err := parseEnhancedAttenuation(attMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse attenuation %d: %w", i, err)
		}
		attenuations = append(attenuations, att)
	}

	return attenuations, nil
}

// parseEnhancedAttenuation parses a single attenuation with module-specific support
func parseEnhancedAttenuation(attMap map[string]any) (Attenuation, error) {
	// Use the existing enhanced verifier logic
	verifier := &Verifier{} // Create temporary verifier for parsing
	return verifier.parseAttenuation(attMap)
}
