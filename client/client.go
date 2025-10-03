// Package client provides a high-level interface for interacting with the Sonr blockchain.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/sonr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SDK represents the main entry point for the Sonr Go Client SDK
type SDK struct {
	client   sonr.Client
	config   *Config
	grpcConn *grpc.ClientConn
}

// Config holds SDK configuration
type Config struct {
	// Network configuration
	Network *config.NetworkConfig

	// Connection timeout
	Timeout time.Duration

	// Enable debug logging
	Debug bool

	// Custom gRPC dial options
	GRPCOptions []grpc.DialOption
}

// DefaultConfig returns default SDK configuration for testnet
func DefaultConfig() *Config {
	clientCfg := config.TestnetConfig()
	return &Config{
		Network: &clientCfg.Network,
		Timeout: 30 * time.Second,
		Debug:   false,
		GRPCOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// LocalConfig returns SDK configuration for local development
func LocalConfig() *Config {
	clientCfg := config.LocalConfig()
	return &Config{
		Network: &clientCfg.Network,
		Timeout: 30 * time.Second,
		Debug:   true,
		GRPCOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// New creates a new SDK instance with the given configuration
func New(cfg *Config) (*SDK, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if cfg.Network == nil {
		return nil, fmt.Errorf("network configuration is required")
	}

	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	grpcConn, err := grpc.DialContext(
		ctx,
		cfg.Network.GRPC,
		cfg.GRPCOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC endpoint: %w", err)
	}

	// Create the main client
	clientCfg := &config.ClientConfig{
		Network:        *cfg.Network,
		KeyringBackend: "test", // Default to test for now
	}

	client, err := sonr.NewClient(clientCfg)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &SDK{
		client:   client,
		config:   cfg,
		grpcConn: grpcConn,
	}, nil
}

// NewWithNetwork creates a new SDK instance for a specific network
func NewWithNetwork(network string) (*SDK, error) {
	var cfg *Config

	switch network {
	case "testnet":
		cfg = DefaultConfig()
	case "local":
		cfg = LocalConfig()
	case "localapi":
		clientCfg := config.LocalAPIConfig()
		cfg = &Config{
			Network: &clientCfg.Network,
			Timeout: 30 * time.Second,
			Debug:   true,
			GRPCOptions: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
		}
	default:
		return nil, fmt.Errorf("unknown network: %s (supported: testnet, local, localapi)", network)
	}

	return New(cfg)
}

// Client returns the underlying Sonr client
func (s *SDK) Client() sonr.Client {
	return s.client
}

// Config returns the SDK configuration
func (s *SDK) Config() *Config {
	return s.config
}

// Close closes all connections
func (s *SDK) Close() error {
	if s.grpcConn != nil {
		return s.grpcConn.Close()
	}
	return nil
}

// WithTimeout returns a new SDK instance with updated timeout
func (s *SDK) WithTimeout(timeout time.Duration) *SDK {
	s.config.Timeout = timeout
	if s.client != nil {
		// Update client config timeout
		clientCfg := &config.ClientConfig{
			Network:        *s.config.Network,
			KeyringBackend: "test",
		}
		client, _ := sonr.NewClient(clientCfg)
		s.client = client
	}
	return s
}

// IsConnected checks if the SDK is connected to the network
func (s *SDK) IsConnected() bool {
	if s.grpcConn == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get node info to check connection
	_, err := s.client.Query().NodeInfo(ctx)
	return err == nil
}

// Version returns the SDK version
func Version() string {
	return "v0.1.0"
}
