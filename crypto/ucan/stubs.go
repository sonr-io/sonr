package ucan

import (
	"time"
)

// TokenBuilderInterface defines token building methods
type TokenBuilderInterface interface {
	CreateOriginToken(
		issuer string,
		capabilities []Attenuation,
		facts []Fact,
		start, expiry time.Time,
	) (*Token, error)
	CreateDelegatedToken(
		parentToken *Token,
		issuer string,
		capabilities []Attenuation,
		facts []Fact,
		start, expiry time.Time,
	) (*Token, error)
}

// TokenBuilder implements token builder functionality
type TokenBuilder struct {
	Capability Attenuation
}

// CreateOriginToken creates a new origin token
func (tb *TokenBuilder) CreateOriginToken(
	issuer string,
	capabilities []Attenuation,
	facts []Fact,
	start, expiry time.Time,
) (*Token, error) {
	return &Token{
		Raw:          "",
		Issuer:       issuer,
		Audience:     "",
		ExpiresAt:    expiry.Unix(),
		NotBefore:    start.Unix(),
		Attenuations: capabilities,
		Proofs:       []Proof{},
		Facts:        facts,
	}, nil
}

// CreateDelegatedToken creates a delegated token
func (tb *TokenBuilder) CreateDelegatedToken(
	parentToken *Token,
	issuer string,
	capabilities []Attenuation,
	facts []Fact,
	start, expiry time.Time,
) (*Token, error) {
	proofs := []Proof{}
	if parentToken.Raw != "" {
		proofs = append(proofs, Proof(parentToken.Raw))
	}

	return &Token{
		Raw:          "",
		Issuer:       issuer,
		Audience:     parentToken.Issuer,
		ExpiresAt:    expiry.Unix(),
		NotBefore:    start.Unix(),
		Attenuations: capabilities,
		Proofs:       proofs,
		Facts:        facts,
	}, nil
}

// Stub for DID validation
func isValidDID(did string) bool {
	// Basic DID validation stub
	return did != "" && len(did) > 5 && did[:4] == "did:"
}

// Stub for preparing delegation proofs
func prepareDelegationProofs(token *Token, capabilities []Attenuation) ([]Proof, error) {
	// Minimal stub implementation
	proofs := []Proof{}
	if token.Raw != "" {
		proofs = append(proofs, Proof(token.Raw))
	}
	return proofs, nil
}
