package handlers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// WebAuthnCredential represents a WebAuthn credential
type WebAuthnCredential struct {
	CredentialID      string    `json:"credential_id"`
	RawID             string    `json:"raw_id"`
	ClientDataJSON    string    `json:"client_data_json"`
	AttestationObject string    `json:"attestation_object"`
	Username          string    `json:"username"`
	Origin            string    `json:"origin"`
	PublicKey         []byte    `json:"public_key"`
	Algorithm         int32     `json:"algorithm"`
	CreatedAt         time.Time `json:"created_at"`
}

// WebAuthnRegistrationRequest represents a registration request
type WebAuthnRegistrationRequest struct {
	Username         string `json:"username"`
	Challenge        string `json:"challenge,omitempty"`
	AutoCreateVault  bool   `json:"auto_create_vault"`
	BroadcastToChain bool   `json:"broadcast_to_chain"`
}

// WebAuthnAuthenticationRequest represents an authentication request
type WebAuthnAuthenticationRequest struct {
	Username  string `json:"username"`
	Challenge string `json:"challenge,omitempty"`
}

// WebAuthnRegistrationResponse contains registration ceremony data
type WebAuthnRegistrationResponse struct {
	Challenge              string                         `json:"challenge"`
	RP                     WebAuthnRPEntity               `json:"rp"`
	User                   WebAuthnUserEntity             `json:"user"`
	PubKeyCredParams       []WebAuthnCredParam            `json:"pubKeyCredParams"`
	AuthenticatorSelection WebAuthnAuthenticatorSelection `json:"authenticatorSelection"`
	Timeout                int                            `json:"timeout"`
	Attestation            string                         `json:"attestation"`
}

// WebAuthnAuthenticationResponse contains authentication ceremony data
type WebAuthnAuthenticationResponse struct {
	Challenge        string                `json:"challenge"`
	Timeout          int                   `json:"timeout"`
	RPID             string                `json:"rpId"`
	AllowCredentials []WebAuthnAllowedCred `json:"allowCredentials"`
	UserVerification string                `json:"userVerification"`
}

// WebAuthnRPEntity represents relying party info
type WebAuthnRPEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WebAuthnUserEntity represents user info
type WebAuthnUserEntity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// WebAuthnCredParam represents credential parameters
type WebAuthnCredParam struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

// WebAuthnAuthenticatorSelection represents authenticator requirements
type WebAuthnAuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	UserVerification        string `json:"userVerification"`
	ResidentKey             string `json:"residentKey,omitempty"`
}

// WebAuthnAllowedCred represents allowed credentials for authentication
type WebAuthnAllowedCred struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// OIDCConfig represents OpenID Connect configuration
type OIDCConfig struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSEndpoint                      string   `json:"jwks_uri"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
}

// OIDCAuthorizationRequest represents an authorization request
type OIDCAuthorizationRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce,omitempty"`
	CodeChallenge       string `json:"code_challenge,omitempty"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
}

// OIDCTokenRequest represents a token request
type OIDCTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// OIDCTokenResponse represents a token response
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OIDCUserInfo represents user information
type OIDCUserInfo struct {
	Subject           string         `json:"sub"`
	Name              string         `json:"name,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	Email             string         `json:"email,omitempty"`
	EmailVerified     bool           `json:"email_verified,omitempty"`
	DID               string         `json:"did,omitempty"`
	VaultID           string         `json:"vault_id,omitempty"`
	UpdatedAt         int64          `json:"updated_at,omitempty"`
	Claims            map[string]any `json:"claims,omitempty"`
}

// JWKSet represents a JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use,omitempty"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg,omitempty"`
	N         string `json:"n,omitempty"`   // RSA modulus
	E         string `json:"e,omitempty"`   // RSA exponent
	X         string `json:"x,omitempty"`   // EC x coordinate
	Y         string `json:"y,omitempty"`   // EC y coordinate
	Curve     string `json:"crv,omitempty"` // EC curve
}

// SIOPRequest represents a Self-Issued OpenID Provider request
type SIOPRequest struct {
	ResponseType string           `json:"response_type"`
	ClientID     string           `json:"client_id"`
	RedirectURI  string           `json:"redirect_uri"`
	Scope        string           `json:"scope"`
	Nonce        string           `json:"nonce"`
	State        string           `json:"state,omitempty"`
	Claims       SIOPClaims       `json:"claims,omitempty"`
	Registration SIOPRegistration `json:"registration,omitempty"`
}

// SIOPClaims represents claims requested in SIOP
type SIOPClaims struct {
	IDToken map[string]ClaimRequest `json:"id_token,omitempty"`
	VPToken map[string]ClaimRequest `json:"vp_token,omitempty"`
}

// ClaimRequest represents a claim request
type ClaimRequest struct {
	Essential bool     `json:"essential,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
}

// SIOPRegistration represents client registration in SIOP
type SIOPRegistration struct {
	ClientName                  string         `json:"client_name,omitempty"`
	ClientPurpose               string         `json:"client_purpose,omitempty"`
	LogoURI                     string         `json:"logo_uri,omitempty"`
	SubjectSyntaxTypesSupported []string       `json:"subject_syntax_types_supported,omitempty"`
	VPFormats                   map[string]any `json:"vp_formats,omitempty"`
}

// SIOPResponse represents a Self-Issued OpenID Provider response
type SIOPResponse struct {
	IDToken string `json:"id_token"`
	VPToken string `json:"vp_token,omitempty"`
	State   string `json:"state,omitempty"`
}

// DIDAuthClaims represents DID-based authentication claims
type DIDAuthClaims struct {
	jwt.RegisteredClaims
	DID       string              `json:"did"`
	Challenge string              `json:"challenge,omitempty"`
	WebAuthn  *WebAuthnCredential `json:"webauthn,omitempty"`
	Extra     map[string]any      `json:"extra,omitempty"`
}

// AuthorizationCode represents an OAuth2 authorization code
type AuthorizationCode struct {
	Code                string
	ClientID            string
	RedirectURI         string
	UserDID             string
	Scope               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	Used                bool
}

// OIDCSession represents an active OIDC session
type OIDCSession struct {
	SessionID    string
	UserDID      string
	ClientID     string
	Scope        string
	Nonce        string
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// BroadcastRequest represents a blockchain broadcast request
type BroadcastRequest struct {
	Message     any    `json:"message"`
	Gasless     bool   `json:"gasless"`
	AutoSign    bool   `json:"auto_sign"`
	FromAddress string `json:"from_address,omitempty"`
}

// BroadcastResponse represents a blockchain broadcast response
type BroadcastResponse struct {
	TxHash  string `json:"tx_hash"`
	Height  int64  `json:"height,omitempty"`
	Code    uint32 `json:"code,omitempty"`
	RawLog  string `json:"raw_log,omitempty"`
	Success bool   `json:"success"`
}
