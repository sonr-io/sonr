package did

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sonr-io/sonr/client/config"
)

// TestDIDClient tests DID module client functionality.
func TestDIDClient(t *testing.T) {
	// Create mock connection
	conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
	cfg := config.LocalNetwork()

	client := NewClient(conn, &cfg)
	require.NotNil(t, client)
}

// TestBasicDIDFunctionality tests that we can create a client
func TestBasicDIDFunctionality(t *testing.T) {
	// Just test that the package compiles and basic functions work
	conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
	cfg := config.LocalNetwork()

	client := NewClient(conn, &cfg)
	require.NotNil(t, client, "DID client should not be nil")

	// Test that the client implements the Client interface
	var _ Client = client
}
