package types

// AssertionStats contains statistics about assertions in the system
type AssertionStats struct {
	// TotalAssertions is the total number of assertions
	TotalAssertions int64 `json:"total_assertions"`

	// EmailAssertions is the number of email assertions
	EmailAssertions int64 `json:"email_assertions"`

	// TelAssertions is the number of telephone assertions
	TelAssertions int64 `json:"tel_assertions"`

	// SonrAssertions is the number of Sonr account assertions
	SonrAssertions int64 `json:"sonr_assertions"`

	// WebAuthnAssertions is the number of WebAuthn assertions
	WebAuthnAssertions int64 `json:"webauthn_assertions"`

	// OtherAssertions is the number of other assertion types
	OtherAssertions int64 `json:"other_assertions"`
}
