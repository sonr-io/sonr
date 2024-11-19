package session

import (
	"github.com/labstack/echo/v4"

	"github.com/onsonr/sonr/pkg/motr/config"
)

// HTTPContext is the context for DWN endpoints.
type HTTPContext struct {
	echo.Context

	role   PeerRole
	client *ClientConfig
	peer   *PeerSession
	user   *UserAgent
	vault  *VaultConfig
}

// loadHeaders loads the headers from the request.
func loadHTTPContext(c echo.Context) *HTTPContext {
	var err error
	cc := &HTTPContext{
		Context: c,
		role:    extractPeerRole(c),
		client:  extractConfigClient(c),
		peer:    extractPeerSession(c),
		user:    extractUserAgent(c),
	}

	if ok := cc.role.Is(RoleMotr); ok {
		cc.vault, err = extractConfigVault(c)
		if err != nil {
			c.Logger().Error(err)
		}
	}
	return cc
}

// ID returns the ksuid http cookie.
func (s *HTTPContext) ID() string {
	return s.peer.ID
}

func (s *HTTPContext) GetLoginParams(credentials []CredDescriptor) *LoginOptions {
	return &LoginOptions{
		Challenge:          s.peer.Challenge,
		Timeout:            10000,
		AllowedCredentials: credentials,
	}
}

func (s *HTTPContext) GetRegisterParams(subject string) *RegisterOptions {
	opts := baseRegisterOptions()
	opts.Challenge = s.peer.Challenge
	opts.User = buildUserEntity(subject)
	return opts
}

// Address returns the sonr address from the cookies.
func (s *HTTPContext) Address() string {
	return s.vault.Address
}

// IPFSGateway returns the IPFS gateway URL from the headers.
func (s *HTTPContext) IPFSGateway() string {
	return s.client.IPFSHost
}

// ChainID returns the chain ID from the headers.
func (s *HTTPContext) ChainID() string {
	return s.client.ChainID
}

// Schema returns the vault schema from the cookies.
func (s *HTTPContext) Schema() *config.Schema {
	return s.vault.Schema
}