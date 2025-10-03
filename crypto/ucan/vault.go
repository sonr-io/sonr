// Package ucan provides User-Controlled Authorization Networks (UCAN) implementation
// for decentralized authorization and capability delegation in the Sonr network.
// This package handles JWT-based tokens, cryptographic verification, and resource capabilities.
package ucan

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// Constants for vault capability actions
const (
	VaultAdminAction = "vault/admin"
)

// VaultCapabilitySchema defines validation specifically for vault capabilities
var VaultCapabilitySchema = z.Struct(z.Shape{
	"can": z.String().Required().OneOf(
		[]string{
			VaultAdminAction,
			"vault/read",
			"vault/write",
			"vault/sign",
			"vault/export",
			"vault/import",
			"vault/delete",
		},
		z.Message("Invalid vault capability"),
	),
	"with": z.String().
		Required().
		TestFunc(ValidateIPFSCID, z.Message("Vault resource must be IPFS CID in format 'ipfs://CID'")),
	"actions": z.Slice(z.String().OneOf(
		[]string{"read", "write", "sign", "export", "import", "delete"},
		z.Message("Invalid vault action"),
	)).Optional(),
	"vault": z.String().Required().Min(1, z.Message("Vault address cannot be empty")),
	"cavs":  z.Slice(z.String()).Optional(), // Caveats as string array for vault capabilities
})

// VaultCapability implements Capability for vault-specific operations
// with support for admin permissions, actions, and enclave data management.
type VaultCapability struct {
	Action         string            `json:"can"`
	Actions        []string          `json:"actions,omitempty"`
	VaultAddress   string            `json:"vault,omitempty"`
	Caveats        []string          `json:"cavs,omitempty"`
	EnclaveDataCID string            `json:"enclave_data_cid,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// GetActions returns the actions this vault capability grants
func (c *VaultCapability) GetActions() []string {
	if c.Action == VaultAdminAction {
		// Admin capability grants all vault actions
		return []string{"read", "write", "sign", "export", "import", "delete", VaultAdminAction}
	}

	if len(c.Actions) > 0 {
		return c.Actions
	}

	// Extract action from the main capability string
	if strings.HasPrefix(c.Action, "vault/") {
		return []string{c.Action[6:]} // Remove "vault/" prefix
	}

	return []string{c.Action}
}

// Grants checks if this capability grants the required abilities
func (c *VaultCapability) Grants(abilities []string) bool {
	if c.Action == VaultAdminAction {
		// Admin capability grants everything
		return true
	}

	grantedActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		grantedActions[action] = true
		grantedActions["vault/"+action] = true // Support both formats
	}

	// Check each required ability
	for _, ability := range abilities {
		if !grantedActions[ability] {
			return false
		}
	}

	return true
}

// Contains checks if this capability contains another capability
func (c *VaultCapability) Contains(other Capability) bool {
	if c.Action == VaultAdminAction {
		// Admin contains all vault capabilities
		if otherVault, ok := other.(*VaultCapability); ok {
			return strings.HasPrefix(otherVault.Action, "vault/")
		}
		// Admin contains any action that starts with vault-related actions
		for _, action := range other.GetActions() {
			if strings.HasPrefix(action, "vault/") ||
				action == "read" || action == "write" || action == "sign" ||
				action == "export" || action == "import" || action == "delete" {
				return true
			}
		}
		return false
	}

	// Check if our actions contain all of the other capability's actions
	ourActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		ourActions[action] = true
		ourActions["vault/"+action] = true
	}

	for _, otherAction := range other.GetActions() {
		if !ourActions[otherAction] {
			return false
		}
	}

	return true
}

// String returns string representation
func (c *VaultCapability) String() string {
	return c.Action
}

// VaultResourceExt represents an extended IPFS-based vault resource (to avoid redeclaration)
type VaultResourceExt struct {
	SimpleResource
	VaultAddress   string `json:"vault_address"`
	EnclaveDataCID string `json:"enclave_data_cid"`
}

// ValidateIPFSCID validates IPFS CID format for vault resources
func ValidateIPFSCID(value *string, ctx z.Ctx) bool {
	if !strings.HasPrefix(*value, "ipfs://") {
		return false
	}
	cidStr := (*value)[7:] // Remove "ipfs://" prefix

	// Enhanced CID validation
	return validateCIDFormat(cidStr)
}

// validateCIDFormat performs comprehensive IPFS CID format validation
func validateCIDFormat(cidStr string) bool {
	if len(cidStr) == 0 {
		return false
	}

	// CIDv0: Base58-encoded SHA-256 multihash (starts with 'Qm' and is 46 characters)
	if strings.HasPrefix(cidStr, "Qm") && len(cidStr) == 46 {
		return isValidBase58(cidStr)
	}

	// CIDv1: Base32 or Base58 encoded (starts with 'b' for base32 or other prefixes)
	if len(cidStr) >= 59 {
		// CIDv1 in base32 typically starts with 'b' and is longer
		if strings.HasPrefix(cidStr, "b") {
			return isValidBase32(cidStr[1:]) // Remove 'b' prefix
		}
		// CIDv1 in base58 or other encodings
		return isValidBase58(cidStr)
	}

	return false
}

// isValidBase58 checks if string contains valid base58 characters
func isValidBase58(s string) bool {
	base58Chars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for _, char := range s {
		if !strings.Contains(base58Chars, string(char)) {
			return false
		}
	}
	return true
}

// isValidBase32 checks if string contains valid base32 characters
func isValidBase32(s string) bool {
	base32Chars := "abcdefghijklmnopqrstuvwxyz234567"
	for _, char := range s {
		if !strings.Contains(base32Chars, string(char)) {
			return false
		}
	}
	return true
}

// ValidateEnclaveDataCIDIntegrity validates enclave data against expected CID
func ValidateEnclaveDataCIDIntegrity(enclaveDataCID string, enclaveData []byte) error {
	if enclaveDataCID == "" {
		return fmt.Errorf("enclave data CID cannot be empty")
	}

	if len(enclaveData) == 0 {
		return fmt.Errorf("enclave data cannot be empty")
	}

	// Validate CID format first
	if !validateCIDFormat(enclaveDataCID) {
		return fmt.Errorf("invalid IPFS CID format: %s", enclaveDataCID)
	}

	// Implement actual CID verification by hashing enclave data
	// 1. Hash the enclave data using SHA-256
	hasher := sha256.New()
	hasher.Write(enclaveData)
	digest := hasher.Sum(nil)

	// 2. Create multihash with SHA-256 prefix
	mhash, err := multihash.EncodeName(digest, "sha2-256")
	if err != nil {
		return fmt.Errorf("failed to create multihash: %w", err)
	}

	// 3. Create CID and compare with expected
	expectedCID, err := cid.Parse(enclaveDataCID)
	if err != nil {
		return fmt.Errorf("failed to parse expected CID: %w", err)
	}

	// Create CID v1 with dag-pb codec (IPFS default)
	calculatedCID := cid.NewCidV1(cid.DagProtobuf, mhash)

	// Compare CIDs
	if !expectedCID.Equals(calculatedCID) {
		return fmt.Errorf(
			"CID verification failed: expected %s, calculated %s",
			expectedCID.String(),
			calculatedCID.String(),
		)
	}

	return nil
}

// ValidateVaultCapability validates vault-specific capabilities
func ValidateVaultCapability(att map[string]any) error {
	var validated struct {
		Can     string   `json:"can"`
		With    string   `json:"with"`
		Actions []string `json:"actions,omitempty"`
		Vault   string   `json:"vault"`
		Cavs    []string `json:"cavs,omitempty"`
	}

	errs := VaultCapabilitySchema.Parse(att, &validated)
	if errs != nil {
		return fmt.Errorf("vault capability validation failed: %v", errs)
	}

	return nil
}

// VaultAttenuationConstructor creates vault-specific attenuations with enhanced validation
func VaultAttenuationConstructor(m map[string]any) (Attenuation, error) {
	// First validate using vault-specific schema
	if err := ValidateVaultCapability(m); err != nil {
		return Attenuation{}, fmt.Errorf("vault attenuation validation failed: %w", err)
	}

	capStr, withStr, err := extractRequiredFields(m)
	if err != nil {
		return Attenuation{}, err
	}

	vaultCap := createVaultCapability(capStr, m)
	resource := createVaultResource(withStr, vaultCap.VaultAddress)

	// Set enclave data CID if using IPFS resource
	if vaultRes, ok := resource.(*VaultResource); ok {
		vaultCap.EnclaveDataCID = vaultRes.EnclaveDataCID
	}

	return Attenuation{
		Capability: vaultCap,
		Resource:   resource,
	}, nil
}

// extractRequiredFields extracts and validates required 'can' and 'with' fields
func extractRequiredFields(m map[string]any) (string, string, error) {
	capValue, exists := m["can"]
	if !exists {
		return "", "", fmt.Errorf("missing 'can' field in attenuation")
	}
	capStr, ok := capValue.(string)
	if !ok {
		return "", "", fmt.Errorf("'can' field must be a string")
	}

	withValue, exists := m["with"]
	if !exists {
		return "", "", fmt.Errorf("missing 'with' field in attenuation")
	}
	withStr, ok := withValue.(string)
	if !ok {
		return "", "", fmt.Errorf("'with' field must be a string")
	}

	return capStr, withStr, nil
}

// createVaultCapability creates and populates a VaultCapability from the input map
func createVaultCapability(action string, m map[string]any) *VaultCapability {
	vaultCap := &VaultCapability{Action: action}

	if actions, exists := m["actions"]; exists {
		vaultCap.Actions = extractStringSlice(actions)
	}

	if vault, exists := m["vault"]; exists {
		if vaultStr, ok := vault.(string); ok {
			vaultCap.VaultAddress = vaultStr
		}
	}

	if cavs, exists := m["cavs"]; exists {
		vaultCap.Caveats = extractStringSlice(cavs)
	}

	return vaultCap
}

// extractStringSlice safely extracts a string slice from an any
func extractStringSlice(value any) []string {
	if slice, ok := value.([]any); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

// createVaultResource creates appropriate Resource based on the URI scheme
func createVaultResource(withStr, vaultAddress string) Resource {
	parts := strings.SplitN(withStr, "://", 2)
	if len(parts) == 2 && parts[0] == "ipfs" {
		return &VaultResource{
			SimpleResource: SimpleResource{
				Scheme: "ipfs",
				Value:  parts[1],
				URI:    withStr,
			},
			VaultAddress:   vaultAddress,
			EnclaveDataCID: parts[1],
		}
	}

	return &SimpleResource{
		Scheme: "ipfs",
		Value:  withStr,
		URI:    withStr,
	}
}

// NewVaultAdminToken creates a new UCAN token with vault admin capabilities
func NewVaultAdminToken(
	builder TokenBuilderInterface,
	vaultOwnerDID string,
	vaultAddress string,
	enclaveDataCID string,
	exp time.Time,
) (*Token, error) {
	// Validate input parameters
	if !isValidDID(vaultOwnerDID) {
		return nil, fmt.Errorf("invalid vault owner DID: %s", vaultOwnerDID)
	}

	// Create vault admin attenuation with full permissions
	vaultResource := &VaultResource{
		SimpleResource: SimpleResource{
			Scheme: "ipfs",
			Value:  enclaveDataCID,
			URI:    fmt.Sprintf("ipfs://%s", enclaveDataCID),
		},
		VaultAddress:   vaultAddress,
		EnclaveDataCID: enclaveDataCID,
	}

	vaultCap := &VaultCapability{
		Action:         VaultAdminAction,
		Actions:        []string{"read", "write", "sign", "export", "import", "delete"},
		VaultAddress:   vaultAddress,
		EnclaveDataCID: enclaveDataCID,
	}

	// Validate the vault capability using vault-specific schema
	capMap := map[string]any{
		"can":     vaultCap.Action,
		"with":    vaultResource.URI,
		"actions": vaultCap.Actions,
		"vault":   vaultCap.VaultAddress,
	}
	if err := ValidateVaultCapability(capMap); err != nil {
		return nil, fmt.Errorf("invalid vault capability: %w", err)
	}

	attenuation := Attenuation{
		Capability: vaultCap,
		Resource:   vaultResource,
	}

	// Create token with vault admin capabilities
	return builder.CreateOriginToken(
		vaultOwnerDID,
		[]Attenuation{attenuation},
		nil,
		time.Now(),
		exp,
	)
}

// ValidateVaultTokenCapability validates a UCAN token for vault operations
func ValidateVaultTokenCapability(token *Token, enclaveDataCID, requiredAction string) error {
	expectedResource := fmt.Sprintf("ipfs://%s", enclaveDataCID)

	// Validate the required action parameter
	validActions := []string{"read", "write", "sign", "export", "import", "delete"}
	actionValid := slices.Contains(validActions, requiredAction)
	if !actionValid {
		return fmt.Errorf("invalid required action: %s", requiredAction)
	}

	// Check if token contains the required vault capability
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == expectedResource {
			// Check if this is a vault capability
			if vaultCap, ok := att.Capability.(*VaultCapability); ok {
				// Validate using vault-specific schema
				validationMap := map[string]any{
					"can":     vaultCap.Action,
					"with":    att.Resource.GetURI(),
					"actions": vaultCap.Actions,
					"vault":   vaultCap.VaultAddress,
				}

				if err := ValidateVaultCapability(validationMap); err != nil {
					continue // Skip invalid capabilities
				}

				// Check if capability grants the required action
				if vaultCap.Grants([]string{requiredAction}) {
					return nil
				}
			}
		}
	}

	return fmt.Errorf(
		"insufficient vault capability: required action '%s' for enclave '%s'",
		requiredAction,
		enclaveDataCID,
	)
}

// GetEnclaveDataCID extracts the enclave data CID from vault capabilities
func GetEnclaveDataCID(token *Token) (string, error) {
	for _, att := range token.Attenuations {
		resource := att.Resource.GetURI()
		if strings.HasPrefix(resource, "ipfs://") {
			return resource[7:], nil
		}
	}
	return "", fmt.Errorf("no enclave data CID found in token")
}
