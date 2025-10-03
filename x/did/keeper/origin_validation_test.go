package keeper_test

import (
	"github.com/sonr-io/sonr/x/did/keeper"
	"github.com/sonr-io/sonr/x/did/types"
)

// TestValidateServiceOrigin tests origin validation logic
func (suite *QueryServerTestSuite) TestValidateServiceOrigin() {
	// Initialize params with allowed origins
	params := types.DefaultParams()
	params.Webauthn.DefaultRpId = "sonr.io"
	params.Webauthn.AllowedOrigins = []string{
		"https://sonr.io",
		"https://app.sonr.io",
		"https://*.example.com",
	}
	err := suite.f.k.Params.Set(suite.f.ctx, params)
	suite.Require().NoError(err)

	querier := suite.f.queryServer.(keeper.Querier)

	testCases := []struct {
		name           string
		origin         string
		expErr         bool
		expErrContains string
	}{
		{
			name:   "success - exact match in allowed origins",
			origin: "https://sonr.io",
			expErr: false,
		},
		{
			name:   "success - subdomain exact match",
			origin: "https://app.sonr.io",
			expErr: false,
		},
		{
			name:   "success - wildcard subdomain match",
			origin: "https://app.example.com",
			expErr: false,
		},
		{
			name:   "success - wildcard match with multiple subdomains",
			origin: "https://deep.nested.example.com",
			expErr: false,
		},
		{
			name:   "success - wildcard matches base domain",
			origin: "https://example.com",
			expErr: false,
		},
		{
			name:   "success - localhost with http",
			origin: "http://localhost",
			expErr: false,
		},
		{
			name:   "success - localhost with https",
			origin: "https://localhost",
			expErr: false,
		},
		{
			name:   "success - 127.0.0.1 with http",
			origin: "http://127.0.0.1",
			expErr: false,
		},
		{
			name:   "success - localhost with port",
			origin: "http://localhost:3000",
			expErr: false,
		},
		{
			name:   "success - IPv6 localhost",
			origin: "http://[::1]",
			expErr: false,
		},
		{
			name:           "error - empty origin",
			origin:         "",
			expErr:         true,
			expErrContains: "origin cannot be empty",
		},
		{
			name:           "error - missing scheme",
			origin:         "sonr.io",
			expErr:         true,
			expErrContains: "origin must start with http:// or https://",
		},
		{
			name:           "error - invalid scheme",
			origin:         "ftp://sonr.io",
			expErr:         true,
			expErrContains: "origin must start with http:// or https://",
		},
		{
			name:           "error - http for non-localhost",
			origin:         "http://sonr.io",
			expErr:         true,
			expErrContains: "non-localhost origins must use HTTPS",
		},
		{
			name:           "error - unregistered origin",
			origin:         "https://malicious.com",
			expErr:         true,
			expErrContains: "not registered in x/svc module and not in allowed origins list",
		},
		{
			name:           "error - subdomain not matching wildcard",
			origin:         "https://app.different.com",
			expErr:         true,
			expErrContains: "not registered in x/svc module and not in allowed origins list",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := querier.ValidateServiceOrigin(suite.f.ctx, tc.origin)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrContains)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestIsLocalhostOrigin tests localhost detection
func (suite *QueryServerTestSuite) TestIsLocalhostOrigin() {
	querier := suite.f.queryServer.(keeper.Querier)

	testCases := []struct {
		domain      string
		isLocalhost bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"[::1]", true},
		{"sonr.io", false},
		{"app.localhost", false},
		{"127.0.0.2", false},
		{"[::2]", false},
	}

	for _, tc := range testCases {
		suite.Run(tc.domain, func() {
			result := querier.IsLocalhostOrigin(tc.domain)
			suite.Require().Equal(tc.isLocalhost, result)
		})
	}
}

// TestMatchesOrigin tests origin pattern matching
func (suite *QueryServerTestSuite) TestMatchesOrigin() {
	querier := suite.f.queryServer.(keeper.Querier)

	testCases := []struct {
		name          string
		fullOrigin    string
		domain        string
		allowedOrigin string
		matches       bool
	}{
		{
			name:          "exact match",
			fullOrigin:    "https://sonr.io",
			domain:        "sonr.io",
			allowedOrigin: "https://sonr.io",
			matches:       true,
		},
		{
			name:          "wildcard subdomain match",
			fullOrigin:    "https://app.example.com",
			domain:        "app.example.com",
			allowedOrigin: "https://*.example.com",
			matches:       true,
		},
		{
			name:          "wildcard base domain match",
			fullOrigin:    "https://example.com",
			domain:        "example.com",
			allowedOrigin: "https://*.example.com",
			matches:       true,
		},
		{
			name:          "wildcard deep subdomain match",
			fullOrigin:    "https://deep.nested.example.com",
			domain:        "deep.nested.example.com",
			allowedOrigin: "https://*.example.com",
			matches:       true,
		},
		{
			name:          "no match - different domain",
			fullOrigin:    "https://sonr.io",
			domain:        "sonr.io",
			allowedOrigin: "https://example.com",
			matches:       false,
		},
		{
			name:          "no match - different subdomain",
			fullOrigin:    "https://app.sonr.io",
			domain:        "app.sonr.io",
			allowedOrigin: "https://web.sonr.io",
			matches:       false,
		},
		{
			name:          "no match - wildcard different domain",
			fullOrigin:    "https://app.sonr.io",
			domain:        "app.sonr.io",
			allowedOrigin: "https://*.example.com",
			matches:       false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := querier.MatchesOrigin(tc.fullOrigin, tc.domain, tc.allowedOrigin)
			suite.Require().Equal(tc.matches, result)
		})
	}
}

// TestExtractDomainFromOrigin tests domain extraction
func (suite *QueryServerTestSuite) TestExtractDomainFromOrigin() {
	testCases := []struct {
		origin         string
		expectedDomain string
	}{
		{"https://sonr.io", "sonr.io"},
		{"http://sonr.io", "sonr.io"},
		{"https://app.sonr.io", "app.sonr.io"},
		{"https://sonr.io:443", "sonr.io"},
		{"http://localhost:3000", "localhost"},
		{"https://sonr.io/path", "sonr.io"},
		{"https://sonr.io:8080/path?query=1", "sonr.io"},
		{"https://[::1]", "[::1]"},
		{"https://[::1]:8080", "[::1]"},
	}

	for _, tc := range testCases {
		suite.Run(tc.origin, func() {
			result := keeper.ExtractDomainFromOrigin(tc.origin)
			suite.Require().Equal(tc.expectedDomain, result)
		})
	}
}

// TestValidateServiceOriginWithEmptyParams tests validation when no allowed origins configured
func (suite *QueryServerTestSuite) TestValidateServiceOriginWithEmptyParams() {
	// Initialize params with empty allowed origins
	params := types.DefaultParams()
	params.Webauthn.DefaultRpId = "sonr.io"
	params.Webauthn.AllowedOrigins = []string{}
	err := suite.f.k.Params.Set(suite.f.ctx, params)
	suite.Require().NoError(err)

	querier := suite.f.queryServer.(keeper.Querier)

	testCases := []struct {
		name           string
		origin         string
		expErr         bool
		expErrContains string
	}{
		{
			name:   "success - localhost still allowed",
			origin: "http://localhost",
			expErr: false,
		},
		{
			name:           "error - non-localhost requires config",
			origin:         "https://sonr.io",
			expErr:         true,
			expErrContains: "not registered in x/svc and no allowed origins configured",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := querier.ValidateServiceOrigin(suite.f.ctx, tc.origin)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrContains)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestValidateServiceOriginWithNilWebAuthnParams tests validation when webauthn params are nil
func (suite *QueryServerTestSuite) TestValidateServiceOriginWithNilWebAuthnParams() {
	// Initialize params with nil webauthn
	params := types.DefaultParams()
	params.Webauthn = nil
	err := suite.f.k.Params.Set(suite.f.ctx, params)
	suite.Require().NoError(err)

	querier := suite.f.queryServer.(keeper.Querier)

	err = querier.ValidateServiceOrigin(suite.f.ctx, "https://sonr.io")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not registered in x/svc and no allowed origins configured")
}
