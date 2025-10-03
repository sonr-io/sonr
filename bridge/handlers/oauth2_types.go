package handlers

import (
	"encoding/json"
	"time"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// OAuth2Client represents a registered OAuth2 client application
type OAuth2Client struct {
	ClientID        string            `json:"client_id"`
	ClientSecret    string            `json:"client_secret,omitempty"`
	ClientType      string            `json:"client_type"` // "confidential" or "public"
	RedirectURIs    []string          `json:"redirect_uris"`
	AllowedScopes   []string          `json:"allowed_scopes"`
	AllowedGrants   []string          `json:"allowed_grants"`
	TokenLifetime   time.Duration     `json:"token_lifetime"`
	RequirePKCE     bool              `json:"require_pkce"`
	TrustedClient   bool              `json:"trusted_client"`
	RequiresConsent bool              `json:"requires_consent"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// OAuth2AuthorizationRequest represents an OAuth2 authorization request
type OAuth2AuthorizationRequest struct {
	ResponseType        string `json:"response_type"         form:"response_type"         query:"response_type"`
	ClientID            string `json:"client_id"             form:"client_id"             query:"client_id"`
	RedirectURI         string `json:"redirect_uri"          form:"redirect_uri"          query:"redirect_uri"`
	Scope               string `json:"scope"                 form:"scope"                 query:"scope"`
	State               string `json:"state"                 form:"state"                 query:"state"`
	CodeChallenge       string `json:"code_challenge"        form:"code_challenge"        query:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method" form:"code_challenge_method" query:"code_challenge_method"`
	Nonce               string `json:"nonce"                 form:"nonce"                 query:"nonce"`
	Prompt              string `json:"prompt"                form:"prompt"                query:"prompt"`
	MaxAge              int    `json:"max_age"               form:"max_age"               query:"max_age"`
	LoginHint           string `json:"login_hint"            form:"login_hint"            query:"login_hint"`
}

// OAuth2TokenRequest represents an OAuth2 token exchange request
type OAuth2TokenRequest struct {
	GrantType           string `json:"grant_type"            form:"grant_type"`
	Code                string `json:"code"                  form:"code"`
	RedirectURI         string `json:"redirect_uri"          form:"redirect_uri"`
	ClientID            string `json:"client_id"             form:"client_id"`
	ClientSecret        string `json:"client_secret"         form:"client_secret"`
	CodeVerifier        string `json:"code_verifier"         form:"code_verifier"`
	RefreshToken        string `json:"refresh_token"         form:"refresh_token"`
	Scope               string `json:"scope"                 form:"scope"`
	Username            string `json:"username"              form:"username"`
	Password            string `json:"password"              form:"password"`
	Assertion           string `json:"assertion"             form:"assertion"`
	ClientAssertion     string `json:"client_assertion"      form:"client_assertion"`
	ClientAssertionType string `json:"client_assertion_type" form:"client_assertion_type"`
}

// OAuth2TokenResponse represents an OAuth2 token response
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	UCANToken    string `json:"ucan_token,omitempty"` // UCAN delegation token
}

// OAuth2ErrorResponse represents an OAuth2 error response
type OAuth2ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	State            string `json:"state,omitempty"`
}

// OAuth2AuthorizationCode represents an authorization code with UCAN context
type OAuth2AuthorizationCode struct {
	Code                string           `json:"code"`
	ClientID            string           `json:"client_id"`
	UserDID             string           `json:"user_did"`
	RedirectURI         string           `json:"redirect_uri"`
	Scopes              []string         `json:"scopes"`
	State               string           `json:"state"`
	Nonce               string           `json:"nonce"`
	CodeChallenge       string           `json:"code_challenge"`
	CodeChallengeMethod string           `json:"code_challenge_method"`
	ExpiresAt           time.Time        `json:"expires_at"`
	Used                bool             `json:"used"`
	UCANContext         *UCANAuthContext `json:"ucan_context"`
}

// UCANAuthContext holds UCAN-related data for authorization
type UCANAuthContext struct {
	VaultAddress   string          `json:"vault_address"`
	EnclaveDataCID string          `json:"enclave_data_cid"`
	DIDDocument    json.RawMessage `json:"did_document"`
	Capabilities   []string        `json:"capabilities"`
}

// OAuth2AccessToken represents an access token with embedded UCAN
type OAuth2AccessToken struct {
	Token     string      `json:"token"`
	UserDID   string      `json:"user_did"`
	ClientID  string      `json:"client_id"`
	Scopes    []string    `json:"scopes"`
	ExpiresAt time.Time   `json:"expires_at"`
	IssuedAt  time.Time   `json:"issued_at"`
	UCANToken *ucan.Token `json:"ucan_token"`
	SessionID string      `json:"session_id"`
	TokenType string      `json:"token_type"`
}

// OAuth2RefreshToken represents a refresh token
type OAuth2RefreshToken struct {
	Token         string    `json:"token"`
	AccessToken   string    `json:"access_token"`
	ClientID      string    `json:"client_id"`
	UserDID       string    `json:"user_did"`
	Scopes        []string  `json:"scopes"`
	ExpiresAt     time.Time `json:"expires_at"`
	IssuedAt      time.Time `json:"issued_at"`
	RotationCount int       `json:"rotation_count"`
}

// OAuth2Session extends OIDCSession with OAuth2-specific fields
type OAuth2Session struct {
	SessionID     string      `json:"session_id"`
	UserDID       string      `json:"user_did"`
	ClientID      string      `json:"client_id"`
	Scopes        []string    `json:"scopes"`
	AccessToken   string      `json:"access_token"`
	RefreshToken  string      `json:"refresh_token"`
	IDToken       string      `json:"id_token,omitempty"`
	UCANToken     *ucan.Token `json:"ucan_token"`
	ConsentGiven  bool        `json:"consent_given"`
	ConsentScopes []string    `json:"consent_scopes"`
	ExpiresAt     time.Time   `json:"expires_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// OAuth2ConsentRequest represents a consent request for a client
type OAuth2ConsentRequest struct {
	UserDID         string    `json:"user_did"`
	ClientID        string    `json:"client_id"`
	RequestedScopes []string  `json:"requested_scopes"`
	ConsentID       string    `json:"consent_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// OAuth2ConsentResponse represents a user's consent response
type OAuth2ConsentResponse struct {
	ConsentID      string   `json:"consent_id"`
	Approved       bool     `json:"approved"`
	ApprovedScopes []string `json:"approved_scopes"`
	RememberChoice bool     `json:"remember_choice"`
}

// OAuth2ScopeDefinition defines an OAuth scope and its UCAN mapping
type OAuth2ScopeDefinition struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	UCANActions  []string       `json:"ucan_actions"`
	ResourceType string         `json:"resource_type"`
	RequiresAuth bool           `json:"requires_auth"`
	Sensitive    bool           `json:"sensitive"`
	ParentScope  string         `json:"parent_scope,omitempty"`
	ChildScopes  []string       `json:"child_scopes,omitempty"`
	Metadata     map[string]any `json:"metadata"`
}

// OAuth2ClientRegistrationRequest represents a dynamic client registration request
type OAuth2ClientRegistrationRequest struct {
	ClientName              string   `json:"client_name"`
	ClientType              string   `json:"client_type"` // "confidential" or "public"
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	Scopes                  string   `json:"scopes,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	TOSUri                  string   `json:"tos_uri,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
}

// OAuth2ClientRegistrationResponse represents a successful client registration
type OAuth2ClientRegistrationResponse struct {
	ClientID                string `json:"client_id"`
	ClientSecret            string `json:"client_secret,omitempty"`
	ClientIDIssuedAt        int64  `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt   int64  `json:"client_secret_expires_at,omitempty"`
	RegistrationAccessToken string `json:"registration_access_token,omitempty"`
	RegistrationClientURI   string `json:"registration_client_uri,omitempty"`
}

// OAuth2IntrospectionRequest represents a token introspection request
type OAuth2IntrospectionRequest struct {
	Token         string `json:"token"           form:"token"`
	TokenTypeHint string `json:"token_type_hint" form:"token_type_hint"`
	ClientID      string `json:"client_id"       form:"client_id"`
	ClientSecret  string `json:"client_secret"   form:"client_secret"`
}

// OAuth2IntrospectionResponse represents a token introspection response
type OAuth2IntrospectionResponse struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
	Audience  []string `json:"aud,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	JWTID     string   `json:"jti,omitempty"`
	UCANToken string   `json:"ucan_token,omitempty"`
}

// OAuth2RevocationRequest represents a token revocation request
type OAuth2RevocationRequest struct {
	Token         string `json:"token"           form:"token"`
	TokenTypeHint string `json:"token_type_hint" form:"token_type_hint"`
	ClientID      string `json:"client_id"       form:"client_id"`
	ClientSecret  string `json:"client_secret"   form:"client_secret"`
}

// OAuth2Config extends OIDCConfig with OAuth2-specific endpoints
type OAuth2Config struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint"`
	JWKSEndpoint                               string   `json:"jwks_uri"`
	RegistrationEndpoint                       string   `json:"registration_endpoint"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint"`
	RevocationEndpoint                         string   `json:"revocation_endpoint"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
	ClaimsSupported                            []string `json:"claims_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"`
	OpPolicyURI                                string   `json:"op_policy_uri,omitempty"`
	OpTosURI                                   string   `json:"op_tos_uri,omitempty"`
	UCANSupported                              bool     `json:"ucan_supported"` // Custom field for UCAN support
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	Valid     bool        `json:"valid"`
	UserDID   string      `json:"user_did"`
	ClientID  string      `json:"client_id"`
	Scopes    []string    `json:"scopes"`
	UCANToken *ucan.Token `json:"ucan_token,omitempty"`
	ExpiresAt time.Time   `json:"expires_at"`
	Error     string      `json:"error,omitempty"`
}
