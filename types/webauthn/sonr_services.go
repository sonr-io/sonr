// Package webauthn provides Sonr-specific WebAuthn service binding functionality
// that integrates with the x/svc module for domain-verified WebAuthn credentials.
//
// This package contains service-related WebAuthn operations that enable binding
// WebAuthn credentials to verified domains for capability-based access control.
package webauthn

import (
	"fmt"
	"net/url"
	"strings"
)

// ServiceBinding represents a binding between a WebAuthn credential and a verified service domain.
type ServiceBinding struct {
	// CredentialID is the WebAuthn credential identifier
	CredentialID string

	// Domain is the verified domain this credential is bound to
	Domain string

	// ServiceID is the unique identifier for the registered service
	ServiceID string

	// Permissions are the specific permissions granted to this credential
	Permissions []string

	// Origin is the WebAuthn origin for this service binding
	Origin string

	// CreatedAt timestamp when the binding was created
	CreatedAt int64

	// ExpiresAt timestamp when the binding expires (optional, 0 = no expiry)
	ExpiresAt int64
}

// ValidateServiceBinding validates a WebAuthn credential for service binding.
// This ensures that the credential is legitimate and the service domain is verified.
func ValidateServiceBinding(
	credential WebAuthnCredential,
	domain string,
	permissions []string,
) error {
	if credential == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Validate credential structure first
	if err := ValidateStructure(credential); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	// Validate domain format
	if err := validateServiceDomain(domain); err != nil {
		return fmt.Errorf("invalid service domain: %w", err)
	}

	// Validate permissions
	if err := validateServicePermissions(permissions); err != nil {
		return fmt.Errorf("invalid service permissions: %w", err)
	}

	// Validate origin matches domain
	if credential.GetOrigin() != "" {
		expectedOrigin := fmt.Sprintf("https://%s", domain)
		if credential.GetOrigin() != expectedOrigin {
			return fmt.Errorf("credential origin %s does not match expected service origin %s",
				credential.GetOrigin(), expectedOrigin)
		}
	}

	return nil
}

// validateServiceDomain validates that a domain meets service binding requirements.
func validateServiceDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic domain validation - should not include protocol or path
	if strings.Contains(domain, "://") {
		return fmt.Errorf("domain should not include protocol (https://)")
	}

	if strings.Contains(domain, "/") {
		return fmt.Errorf("domain should not include path")
	}

	// Check domain format using URL parsing
	testURL := "https://" + domain
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return fmt.Errorf("invalid domain format: %w", err)
	}

	if parsedURL.Hostname() != domain {
		return fmt.Errorf("invalid domain format")
	}

	return nil
}

// validateServicePermissions validates the permissions granted to a service-bound credential.
func validateServicePermissions(permissions []string) error {
	if len(permissions) == 0 {
		return fmt.Errorf("at least one permission must be specified")
	}

	validPermissions := map[string]bool{
		"read":     true,
		"write":    true,
		"execute":  true,
		"admin":    true,
		"delegate": true,
	}

	for _, perm := range permissions {
		if !validPermissions[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	return nil
}

// GenerateServiceOrigin generates the expected WebAuthn origin for a service domain.
func GenerateServiceOrigin(domain string) (string, error) {
	if err := validateServiceDomain(domain); err != nil {
		return "", fmt.Errorf("invalid domain: %w", err)
	}

	return fmt.Sprintf("https://%s", domain), nil
}

// ValidateCredentialForDomain validates that a WebAuthn credential is valid for a specific domain.
// This includes checking the origin and ensuring the credential can be bound to the domain.
func ValidateCredentialForDomain(
	credential WebAuthnCredential,
	domain string,
	challengeToken string,
) error {
	if credential == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Generate expected origin for the domain
	expectedOrigin, err := GenerateServiceOrigin(domain)
	if err != nil {
		return fmt.Errorf("failed to generate origin: %w", err)
	}

	// Validate the credential with domain-specific requirements
	if err := ValidateForGaslessRegistration(credential, challengeToken, expectedOrigin); err != nil {
		return fmt.Errorf("credential validation failed for domain %s: %w", domain, err)
	}

	return nil
}

// CreateServiceBinding creates a new service binding for a WebAuthn credential.
func CreateServiceBinding(
	credentialID string,
	domain string,
	serviceID string,
	permissions []string,
	createdAt int64,
) (*ServiceBinding, error) {
	if credentialID == "" {
		return nil, fmt.Errorf("credential ID cannot be empty")
	}

	if err := validateServiceDomain(domain); err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}

	if serviceID == "" {
		return nil, fmt.Errorf("service ID cannot be empty")
	}

	if err := validateServicePermissions(permissions); err != nil {
		return nil, fmt.Errorf("invalid permissions: %w", err)
	}

	origin, err := GenerateServiceOrigin(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to generate origin: %w", err)
	}

	return &ServiceBinding{
		CredentialID: credentialID,
		Domain:       domain,
		ServiceID:    serviceID,
		Permissions:  permissions,
		Origin:       origin,
		CreatedAt:    createdAt,
		ExpiresAt:    0, // No expiration by default
	}, nil
}

// ValidateServiceBindingPermission checks if a service binding has a specific permission.
func ValidateServiceBindingPermission(binding *ServiceBinding, requiredPermission string) error {
	if binding == nil {
		return fmt.Errorf("service binding cannot be nil")
	}

	// Check if the binding has the required permission
	for _, perm := range binding.Permissions {
		if perm == requiredPermission || perm == "admin" {
			return nil // Permission granted
		}
	}

	return fmt.Errorf("permission denied: %s not granted to credential %s for domain %s",
		requiredPermission, binding.CredentialID, binding.Domain)
}

// IsServiceBindingExpired checks if a service binding has expired.
func IsServiceBindingExpired(binding *ServiceBinding, currentTime int64) bool {
	if binding == nil {
		return true
	}

	// If ExpiresAt is 0, the binding never expires
	if binding.ExpiresAt == 0 {
		return false
	}

	return currentTime >= binding.ExpiresAt
}

// ValidateServiceBindingAccess validates that a credential can access a service with specific permissions.
func ValidateServiceBindingAccess(
	binding *ServiceBinding,
	requiredPermission string,
	currentTime int64,
) error {
	if binding == nil {
		return fmt.Errorf("no service binding found")
	}

	// Check if binding is expired
	if IsServiceBindingExpired(binding, currentTime) {
		return fmt.Errorf("service binding expired at %d", binding.ExpiresAt)
	}

	// Check permission
	if err := ValidateServiceBindingPermission(binding, requiredPermission); err != nil {
		return err
	}

	return nil
}

// GenerateServiceBindingID generates a unique ID for a service binding.
func GenerateServiceBindingID(credentialID, domain, serviceID string) string {
	return fmt.Sprintf("%s:%s:%s", credentialID, domain, serviceID)
}

// ValidateDomainVerificationForBinding ensures that a domain is properly verified
// before allowing WebAuthn credential binding.
func ValidateDomainVerificationForBinding(
	domain string,
	verificationStatus string,
	verifiedAt int64,
	currentTime int64,
) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check verification status
	if verificationStatus != "DOMAIN_VERIFICATION_STATUS_VERIFIED" {
		return fmt.Errorf("domain %s is not verified (status: %s)", domain, verificationStatus)
	}

	// Ensure verification is not too old (e.g., within 30 days)
	const maxVerificationAge = 30 * 24 * 60 * 60 // 30 days in seconds
	if currentTime-verifiedAt > maxVerificationAge {
		return fmt.Errorf(
			"domain verification is too old (verified %d seconds ago)",
			currentTime-verifiedAt,
		)
	}

	return nil
}

// ExtractDomainFromOrigin extracts the domain from a WebAuthn origin.
func ExtractDomainFromOrigin(origin string) (string, error) {
	if origin == "" {
		return "", fmt.Errorf("origin cannot be empty")
	}

	parsedURL, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("invalid origin format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return "", fmt.Errorf("origin must use HTTPS")
	}

	return parsedURL.Hostname(), nil
}
