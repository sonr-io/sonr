package handlers

import "errors"

// Common errors for OAuth2 and UCAN handling
var (
	// Client errors
	ErrClientNotFound           = errors.New("client not found")
	ErrInvalidClientCredentials = errors.New("invalid client credentials")

	// Token errors
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token has expired")
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenRevoked  = errors.New("token has been revoked")

	// Authorization errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInsufficientScope = errors.New("insufficient scope")
	ErrInvalidScope      = errors.New("invalid scope")
	ErrScopeNotAllowed   = errors.New("scope not allowed for client")

	// UCAN errors
	ErrInvalidAttenuation  = errors.New("invalid attenuation")
	ErrInvalidDelegation   = errors.New("invalid delegation")
	ErrBrokenChain         = errors.New("broken delegation chain")
	ErrPrivilegeEscalation = errors.New("privilege escalation attempted")

	// OIDC errors
	ErrInvalidRedirectURI  = errors.New("invalid redirect URI")
	ErrInvalidResponseType = errors.New("invalid response type")
	ErrInvalidGrantType    = errors.New("invalid grant type")
)
