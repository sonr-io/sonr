// Package sonr provides the main client interface for interacting with the Sonr blockchain.
package sonr

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/client/keys"
	"github.com/sonr-io/sonr/client/modules/did"
	"github.com/sonr-io/sonr/client/modules/dwn"
	"github.com/sonr-io/sonr/client/modules/svc"
	"github.com/sonr-io/sonr/client/modules/ucan"
	"github.com/sonr-io/sonr/client/query"
	"github.com/sonr-io/sonr/client/tx"
)

// Client is the main interface for interacting with the Sonr blockchain.
// It provides access to query operations, transaction building, and module-specific functionality.
type Client interface {
	// Core functionality
	Query() query.QueryClient
	Transaction() tx.TxBuilder
	Keyring() keys.KeyringManager

	// Module clients
	DID() did.Client
	DWN() dwn.Client
	SVC() svc.Client
	UCAN() ucan.Client

	// Connection management
	Close() error
	Health(ctx context.Context) error
	Config() *config.ClientConfig
}

// client implements the Client interface.
type client struct {
	config *config.ClientConfig

	// gRPC connections
	grpcConn *grpc.ClientConn

	// Module clients
	queryClient query.QueryClient
	txBuilder   tx.TxBuilder
	keyring     keys.KeyringManager

	didClient  did.Client
	dwnClient  dwn.Client
	svcClient  svc.Client
	ucanClient ucan.Client
}

// ClientOption allows customization of the client during initialization.
type ClientOption func(*clientOptions)

type clientOptions struct {
	grpcDialOptions []grpc.DialOption
	keyringBackend  string
	keyringDir      string
}

// WithGRPCDialOptions allows setting custom gRPC dial options.
func WithGRPCDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcDialOptions = append(o.grpcDialOptions, opts...)
	}
}

// WithKeyringBackend sets the keyring backend (os, file, test, memory).
func WithKeyringBackend(backend string) ClientOption {
	return func(o *clientOptions) {
		o.keyringBackend = backend
	}
}

// WithKeyringDirectory sets the directory for file-based keyring.
func WithKeyringDirectory(dir string) ClientOption {
	return func(o *clientOptions) {
		o.keyringDir = dir
	}
}

// NewClient creates a new Sonr blockchain client with the given configuration.
func NewClient(cfg *config.ClientConfig, opts ...ClientOption) (Client, error) {
	if cfg == nil {
		return nil, errors.ErrMissingConfig
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.WrapError(err, errors.ErrInvalidConfig, "client configuration validation failed")
	}

	// Apply options
	options := &clientOptions{
		keyringBackend: cfg.KeyringBackend,
		keyringDir:     cfg.KeyringDir,
	}
	for _, opt := range opts {
		opt(options)
	}

	c := &client{
		config: cfg,
	}

	// Initialize gRPC connection
	if err := c.initGRPCConnection(options); err != nil {
		return nil, errors.WrapError(err, errors.ErrConnectionFailed, "failed to initialize gRPC connection")
	}

	// Initialize components
	if err := c.initComponents(options); err != nil {
		c.Close() // Clean up on error
		return nil, errors.WrapError(err, errors.ErrInvalidConfig, "failed to initialize client components")
	}

	return c, nil
}

// initGRPCConnection establishes the gRPC connection to the blockchain.
func (c *client) initGRPCConnection(opts *clientOptions) error {
	endpoint := c.config.Network.GRPC
	if endpoint == "" {
		return fmt.Errorf("gRPC endpoint not configured")
	}

	// Build dial options
	dialOpts := []grpc.DialOption{}

	// Configure TLS
	if c.config.Network.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	// Add custom dial options
	dialOpts = append(dialOpts, opts.grpcDialOptions...)

	// Create connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Network.RequestTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, dialOpts...)
	if err != nil {
		return errors.NewConnectionError(endpoint, err)
	}

	c.grpcConn = conn
	return nil
}

// initComponents initializes all client components.
func (c *client) initComponents(opts *clientOptions) error {
	// Initialize keyring
	keyringBackend := opts.keyringBackend
	if keyringBackend == "" {
		keyringBackend = "test" // Default for safety
	}

	keyringManager, err := keys.NewKeyringManager(keyringBackend, opts.keyringDir, c.config.Network.ChainID)
	if err != nil {
		return fmt.Errorf("failed to initialize keyring: %w", err)
	}
	c.keyring = keyringManager

	// Initialize query client
	queryClient, err := query.NewQueryClient(c.grpcConn, &c.config.Network)
	if err != nil {
		return fmt.Errorf("failed to initialize query client: %w", err)
	}
	c.queryClient = queryClient

	// Initialize transaction builder
	txBuilder, err := tx.NewTxBuilder(&c.config.Network, c.grpcConn)
	if err != nil {
		return fmt.Errorf("failed to initialize transaction builder: %w", err)
	}
	c.txBuilder = txBuilder

	// Initialize module clients
	c.didClient = did.NewClient(c.grpcConn, &c.config.Network)
	c.dwnClient = dwn.NewClient(c.grpcConn, &c.config.Network)
	c.svcClient = svc.NewClient(c.grpcConn, &c.config.Network)
	c.ucanClient = ucan.NewClient(c.grpcConn, &c.config.Network)

	return nil
}

// Query returns the query client for read operations.
func (c *client) Query() query.QueryClient {
	return c.queryClient
}

// Transaction returns the transaction builder for write operations.
func (c *client) Transaction() tx.TxBuilder {
	return c.txBuilder
}

// Keyring returns the keyring manager for key operations.
func (c *client) Keyring() keys.KeyringManager {
	return c.keyring
}

// DID returns the DID module client.
func (c *client) DID() did.Client {
	return c.didClient
}

// DWN returns the DWN module client.
func (c *client) DWN() dwn.Client {
	return c.dwnClient
}

// SVC returns the SVC module client.
func (c *client) SVC() svc.Client {
	return c.svcClient
}

// UCAN returns the UCAN module client.
func (c *client) UCAN() ucan.Client {
	return c.ucanClient
}

// Health checks the health of the connection to the blockchain.
func (c *client) Health(ctx context.Context) error {
	if c.grpcConn == nil {
		return errors.ErrConnectionFailed
	}

	// Use a short timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get chain info as a health check
	_, err := c.queryClient.ChainInfo(healthCtx)
	if err != nil {
		return errors.WrapError(err, errors.ErrConnectionFailed, "health check failed")
	}

	return nil
}

// Config returns the client configuration.
func (c *client) Config() *config.ClientConfig {
	return c.config
}

// Close closes all connections and releases resources.
func (c *client) Close() error {
	var errs []error

	// Close gRPC connection
	if c.grpcConn != nil {
		if err := c.grpcConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close gRPC connection: %w", err))
		}
	}

	// Close keyring (if it supports closing)
	if c.keyring != nil {
		if closer, ok := c.keyring.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close keyring: %w", err))
			}
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("errors while closing client: %v", errs)
	}

	return nil
}

// NewTestClient creates a client configured for testing with sensible defaults.
func NewTestClient() (Client, error) {
	cfg := config.LocalConfig()
	cfg.KeyringBackend = "memory"

	return NewClient(cfg)
}

// ConnectToTestnet creates a client connected to the Sonr testnet.
func ConnectToTestnet(opts ...ClientOption) (Client, error) {
	cfg := config.TestnetConfig()
	return NewClient(cfg, opts...)
}

// ConnectToLocal creates a client connected to a local Sonr node.
func ConnectToLocal(opts ...ClientOption) (Client, error) {
	cfg := config.LocalConfig()
	return NewClient(cfg, opts...)
}

// ConnectToLocalAPI creates a client connected to a local Sonr API server using localhost.
func ConnectToLocalAPI(opts ...ClientOption) (Client, error) {
	cfg := config.LocalAPIConfig()
	return NewClient(cfg, opts...)
}

// ConnectWithConfig creates a client with custom configuration.
func ConnectWithConfig(cfg *config.ClientConfig, opts ...ClientOption) (Client, error) {
	return NewClient(cfg, opts...)
}
