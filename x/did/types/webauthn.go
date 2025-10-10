// Package types provides x/did module types with WebAuthn validation
// All WebAuthn validation logic is now in validation.go in this package.
package types

// ValidateStructure validates a WebAuthn credential for gasless transaction processing.
// This method uses the local validation logic in this package.
func (c *WebAuthnCredential) ValidateStructure() error {
	return ValidateStructure(c)
}

// ValidateAttestation performs security validation of WebAuthn credential data.
// This method uses the local validation logic in this package.
func (c *WebAuthnCredential) ValidateAttestation(challenge, expectedOrigin string) error {
	return ValidateAttestation(c, challenge, expectedOrigin)
}

// ValidateForGaslessRegistration performs comprehensive validation for gasless WebAuthn registration.
// This method uses the local validation logic in this package.
func (c *WebAuthnCredential) ValidateForGaslessRegistration(
	challenge, expectedOrigin string,
) error {
	return ValidateForGaslessRegistration(c, challenge, expectedOrigin)
}
