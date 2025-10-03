package spec

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/mpc"
	"lukechampine.com/blake3"
)

// KeyshareSource provides MPC-based UCAN token creation and validation
type KeyshareSource interface {
	Address() string
	Issuer() string
	ChainCode() ([]byte, error)
	OriginToken() (*Token, error)
	SignData(data []byte) ([]byte, error)
	VerifyData(data []byte, sig []byte) (bool, error)
	Enclave() mpc.Enclave

	// UCAN token creation methods
	NewOriginToken(
		audienceDID string,
		att []Attenuation,
		fct []Fact,
		notBefore, expires time.Time,
	) (*Token, error)
	NewAttenuatedToken(
		parent *Token,
		audienceDID string,
		att []Attenuation,
		fct []Fact,
		nbf, exp time.Time,
	) (*Token, error)
}

// NewSource creates a new MPC-based keyshare source from an enclave
func NewSource(enclave mpc.Enclave) (KeyshareSource, error) {
	if !enclave.IsValid() {
		return nil, fmt.Errorf("invalid MPC enclave provided")
	}

	pubKeyBytes := enclave.PubKeyBytes()
	issuerDID, addr, err := getIssuerDIDFromBytes(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to derive issuer DID: %w", err)
	}

	return &mpcKeyshareSource{
		enclave:   enclave,
		issuerDID: issuerDID,
		addr:      addr,
	}, nil
}

// mpcKeyshareSource implements KeyshareSource using MPC enclave
type mpcKeyshareSource struct {
	enclave   mpc.Enclave
	issuerDID string
	addr      string
}

// Address returns the address derived from the enclave public key
func (k *mpcKeyshareSource) Address() string {
	return k.addr
}

// Issuer returns the DID of the issuer derived from the enclave public key
func (k *mpcKeyshareSource) Issuer() string {
	return k.issuerDID
}

// Enclave returns the underlying MPC enclave
func (k *mpcKeyshareSource) Enclave() mpc.Enclave {
	return k.enclave
}

// ChainCode derives a deterministic chain code from the enclave
func (k *mpcKeyshareSource) ChainCode() ([]byte, error) {
	// Sign the address to create a deterministic chain code
	sig, err := k.SignData([]byte(k.addr))
	if err != nil {
		return nil, fmt.Errorf("failed to sign address for chain code: %w", err)
	}

	// Hash the signature to create a 32-byte chain code
	hash := blake3.Sum256(sig)
	return hash[:32], nil
}

// OriginToken creates a default origin token with basic capabilities
func (k *mpcKeyshareSource) OriginToken() (*Token, error) {
	// Create basic capability for the MPC keyshare
	resource := &SimpleResource{
		Scheme: "mpc",
		Value:  k.addr,
		URI:    fmt.Sprintf("mpc://%s", k.addr),
	}

	capability := &SimpleCapability{Action: "sign"}

	attenuation := Attenuation{
		Capability: capability,
		Resource:   resource,
	}

	// Create token with no expiration for origin token
	zero := time.Time{}
	return k.NewOriginToken(k.issuerDID, []Attenuation{attenuation}, nil, zero, zero)
}

// SignData signs data using the MPC enclave
func (k *mpcKeyshareSource) SignData(data []byte) ([]byte, error) {
	if !k.enclave.IsValid() {
		return nil, fmt.Errorf("enclave is not valid")
	}

	return k.enclave.Sign(data)
}

// VerifyData verifies a signature using the MPC enclave
func (k *mpcKeyshareSource) VerifyData(data []byte, sig []byte) (bool, error) {
	if !k.enclave.IsValid() {
		return false, fmt.Errorf("enclave is not valid")
	}

	return k.enclave.Verify(data, sig)
}

// NewOriginToken creates a new UCAN origin token using MPC signing
func (k *mpcKeyshareSource) NewOriginToken(
	audienceDID string,
	att []Attenuation,
	fct []Fact,
	notBefore, expires time.Time,
) (*Token, error) {
	return k.newToken(audienceDID, nil, att, fct, notBefore, expires)
}

// NewAttenuatedToken creates a new attenuated UCAN token using MPC signing
func (k *mpcKeyshareSource) NewAttenuatedToken(
	parent *Token,
	audienceDID string,
	att []Attenuation,
	fct []Fact,
	nbf, exp time.Time,
) (*Token, error) {
	// Validate that new attenuations are more restrictive than parent
	if !isAttenuationSubset(att, parent.Attenuations) {
		return nil, fmt.Errorf("scope of ucan attenuations must be less than its parent")
	}

	// Add parent as proof
	proofs := []Proof{}
	if parent.Raw != "" {
		proofs = append(proofs, Proof(parent.Raw))
	}
	proofs = append(proofs, parent.Proofs...)

	return k.newToken(audienceDID, proofs, att, fct, nbf, exp)
}

// newToken creates a new UCAN token with MPC signing
func (k *mpcKeyshareSource) newToken(
	audienceDID string,
	proofs []Proof,
	att []Attenuation,
	fct []Fact,
	nbf, exp time.Time,
) (*Token, error) {
	// Validate audience DID
	if !isValidDID(audienceDID) {
		return nil, fmt.Errorf("invalid audience DID: %s", audienceDID)
	}

	// Create JWT with MPC signing method
	t := jwt.New(NewJWTSigningMethod("MPC256", k.enclave))

	// Set UCAN version header
	t.Header[UCANVersionKey] = UCANVersion

	var (
		nbfUnix int64
		expUnix int64
	)

	if !nbf.IsZero() {
		nbfUnix = nbf.Unix()
	}
	if !exp.IsZero() {
		expUnix = exp.Unix()
	}

	// Convert attenuations to claim format
	attClaims := make([]map[string]any, len(att))
	for i, a := range att {
		attClaims[i] = map[string]any{
			"can":  a.Capability.GetActions(),
			"with": a.Resource.GetURI(),
		}
	}

	// Convert proofs to strings
	proofStrings := make([]string, len(proofs))
	for i, proof := range proofs {
		proofStrings[i] = string(proof)
	}

	// Convert facts to any slice
	factData := make([]any, len(fct))
	for i, fact := range fct {
		factData[i] = string(fact.Data)
	}

	// Set claims
	claims := jwt.MapClaims{
		"iss": k.issuerDID,
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

	t.Claims = claims

	// Sign the token using MPC enclave
	tokenString, err := t.SignedString(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &Token{
		Raw:          tokenString,
		Issuer:       k.issuerDID,
		Audience:     audienceDID,
		ExpiresAt:    expUnix,
		NotBefore:    nbfUnix,
		Attenuations: att,
		Proofs:       proofs,
		Facts:        fct,
	}, nil
}

// isAttenuationSubset checks if child attenuations are a subset of parent attenuations
func isAttenuationSubset(child, parent []Attenuation) bool {
	for _, childAtt := range child {
		if !containsAttenuation(parent, childAtt) {
			return false
		}
	}
	return true
}

// containsAttenuation checks if the parent list contains an equivalent attenuation
func containsAttenuation(parent []Attenuation, att Attenuation) bool {
	for _, parentAtt := range parent {
		if parentAtt.Resource.Matches(att.Resource) &&
			parentAtt.Capability.Contains(att.Capability) {
			return true
		}
	}
	return false
}

// isValidDID validates DID format
func isValidDID(did string) bool {
	return did != "" && len(did) > 5 && strings.HasPrefix(did, "did:")
}

// getIssuerDIDFromBytes creates an issuer DID and address from public key bytes
func getIssuerDIDFromBytes(pubKeyBytes []byte) (string, string, error) {
	// Convert MPC public key bytes to libp2p crypto.PubKey
	pubKey, err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal secp256k1 key: %w", err)
	}

	// Create DID using the crypto/keys package
	did, err := keys.NewDID(pubKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create DID: %w", err)
	}

	didStr := did.String()

	// Generate address from DID (simplified implementation)
	address := fmt.Sprintf("addr_%x", pubKeyBytes[:8])

	return didStr, address, nil
}
