package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Test that nested params are not nil
	require.NotNil(t, params.Document)
	require.NotNil(t, params.Webauthn)

	// Test WebAuthn parameters
	require.Equal(t, int64(60), params.Webauthn.ChallengeTimeout)
	require.NotEmpty(t, params.Webauthn.AllowedOrigins)
	require.NotEmpty(t, params.Webauthn.SupportedAlgorithms)
	require.True(t, params.Webauthn.RequireUserVerification)
	require.Equal(t, int32(10), params.Webauthn.MaxCredentialsPerDid)
	require.Equal(t, "localhost", params.Webauthn.DefaultRpId)
	require.Equal(t, "Sonr Identity Platform", params.Webauthn.DefaultRpName)

	// Test Document parameters
	require.True(t, params.Document.AutoCreateVault)
	require.Equal(t, int32(20), params.Document.MaxVerificationMethods)
	require.Equal(t, int32(10), params.Document.MaxServiceEndpoints)
	require.Equal(t, int32(5), params.Document.MaxControllers)
	require.Equal(t, int64(65536), params.Document.DidDocumentMaxSize)
	require.Equal(t, int64(5), params.Document.DidResolutionTimeout)
	require.Equal(t, int64(2592000), params.Document.KeyRotationInterval)
	require.Equal(t, int64(31536000), params.Document.CredentialLifetime)
	require.NotEmpty(t, params.Document.SupportedAssertionMethods)
	require.NotEmpty(t, params.Document.SupportedAuthenticationMethods)

	// Validate that default params pass validation
	require.NoError(t, params.Validate())
}

func TestParamsValidation(t *testing.T) {
	testCases := []struct {
		name      string
		modifyFn  func(*Params)
		expectErr bool
	}{
		{
			name: "valid default params",
			modifyFn: func(p *Params) {
				// No modifications - should be valid
			},
			expectErr: false,
		},
		{
			name: "invalid webauthn challenge timeout - too low",
			modifyFn: func(p *Params) {
				p.Webauthn.ChallengeTimeout = 29
			},
			expectErr: true,
		},
		{
			name: "invalid webauthn challenge timeout - too high",
			modifyFn: func(p *Params) {
				p.Webauthn.ChallengeTimeout = 301
			},
			expectErr: true,
		},
		{
			name: "empty allowed origins",
			modifyFn: func(p *Params) {
				p.Webauthn.AllowedOrigins = []string{}
			},
			expectErr: true,
		},
		{
			name: "invalid origin",
			modifyFn: func(p *Params) {
				p.Webauthn.AllowedOrigins = []string{"invalid-origin"}
			},
			expectErr: true,
		},
		{
			name: "empty supported algorithms",
			modifyFn: func(p *Params) {
				p.Webauthn.SupportedAlgorithms = []string{}
			},
			expectErr: true,
		},
		{
			name: "invalid algorithm",
			modifyFn: func(p *Params) {
				p.Webauthn.SupportedAlgorithms = []string{"INVALID"}
			},
			expectErr: true,
		},
		{
			name: "invalid max credentials per DID - too low",
			modifyFn: func(p *Params) {
				p.Webauthn.MaxCredentialsPerDid = 0
			},
			expectErr: true,
		},
		{
			name: "invalid max credentials per DID - too high",
			modifyFn: func(p *Params) {
				p.Webauthn.MaxCredentialsPerDid = 101
			},
			expectErr: true,
		},
		{
			name: "invalid max verification methods - too low",
			modifyFn: func(p *Params) {
				p.Document.MaxVerificationMethods = 0
			},
			expectErr: true,
		},
		{
			name: "invalid max verification methods - too high",
			modifyFn: func(p *Params) {
				p.Document.MaxVerificationMethods = 51
			},
			expectErr: true,
		},
		{
			name: "invalid max service endpoints - too low",
			modifyFn: func(p *Params) {
				p.Document.MaxServiceEndpoints = -1
			},
			expectErr: true,
		},
		{
			name: "invalid max service endpoints - too high",
			modifyFn: func(p *Params) {
				p.Document.MaxServiceEndpoints = 21
			},
			expectErr: true,
		},
		{
			name: "invalid max controllers - too low",
			modifyFn: func(p *Params) {
				p.Document.MaxControllers = 0
			},
			expectErr: true,
		},
		{
			name: "invalid max controllers - too high",
			modifyFn: func(p *Params) {
				p.Document.MaxControllers = 11
			},
			expectErr: true,
		},
		{
			name: "invalid DID document max size - too small",
			modifyFn: func(p *Params) {
				p.Document.DidDocumentMaxSize = 1023
			},
			expectErr: true,
		},
		{
			name: "invalid DID document max size - too large",
			modifyFn: func(p *Params) {
				p.Document.DidDocumentMaxSize = 102401
			},
			expectErr: true,
		},
		{
			name: "invalid DID resolution timeout - too low",
			modifyFn: func(p *Params) {
				p.Document.DidResolutionTimeout = 0
			},
			expectErr: true,
		},
		{
			name: "invalid DID resolution timeout - too high",
			modifyFn: func(p *Params) {
				p.Document.DidResolutionTimeout = 31
			},
			expectErr: true,
		},
		{
			name: "invalid key rotation interval - too short",
			modifyFn: func(p *Params) {
				p.Document.KeyRotationInterval = 86399
			},
			expectErr: true,
		},
		{
			name: "invalid key rotation interval - too long",
			modifyFn: func(p *Params) {
				p.Document.KeyRotationInterval = 31536001
			},
			expectErr: true,
		},
		{
			name: "invalid credential lifetime - too short",
			modifyFn: func(p *Params) {
				p.Document.CredentialLifetime = 3599
			},
			expectErr: true,
		},
		{
			name: "invalid credential lifetime - too long",
			modifyFn: func(p *Params) {
				p.Document.CredentialLifetime = 315360001
			},
			expectErr: true,
		},
		{
			name: "empty supported assertion methods",
			modifyFn: func(p *Params) {
				p.Document.SupportedAssertionMethods = []string{}
			},
			expectErr: true,
		},
		{
			name: "empty supported authentication methods",
			modifyFn: func(p *Params) {
				p.Document.SupportedAuthenticationMethods = []string{}
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := DefaultParams()
			tc.modifyFn(&params)

			err := params.Validate()
			if tc.expectErr {
				require.Error(t, err, "Expected validation to fail but it passed")
			} else {
				require.NoError(t, err, "Expected validation to pass but it failed: %v", err)
			}
		})
	}
}
