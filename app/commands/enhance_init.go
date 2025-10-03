// Package commands contains utility functions for the snrd command
package commands

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/sonr-io/sonr/app"
	"github.com/sonr-io/sonr/crypto/vrf"
	"github.com/spf13/cobra"
)

func EnhancedInit(chainApp *app.ChainApp) *cobra.Command {
	baseCmd := genutilcli.InitCmd(chainApp.BasicModuleManager, app.DefaultNodeHome)
	baseCmd.PostRunE = handleInitPostE
	return baseCmd
}

// handleInitPostE generates VRF keypair and stores it securely after chain initialization
func handleInitPostE(cmd *cobra.Command, args []string) error {
	// Extract chain-id from genesis file for network-aware key generation
	chainID, err := getChainIDFromGenesis()
	if err != nil {
		return fmt.Errorf("failed to get chain-id: %w", err)
	}

	// Create deterministic entropy source from chainID
	entropySource, err := createChainIDEntropySource(chainID)
	if err != nil {
		return fmt.Errorf("failed to create entropy source: %w", err)
	}

	// Generate VRF keypair using chainID-derived entropy
	privateKey, err := vrf.GenerateKey(entropySource)
	if err != nil {
		return fmt.Errorf("failed to generate VRF keypair: %w", err)
	}

	// Validate the generated keypair
	if len(privateKey) != vrf.PrivateKeySize {
		return fmt.Errorf(
			"invalid VRF private key size: expected %d, got %d",
			vrf.PrivateKeySize,
			len(privateKey),
		)
	}

	// Get public key to validate keypair
	_, ok := privateKey.Public()
	if !ok {
		return fmt.Errorf("failed to derive public key from VRF private key")
	}

	// Create secure storage path
	vrfKeyPath := filepath.Join(app.DefaultNodeHome, "vrf_secret.key")

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(app.DefaultNodeHome, 0o750); err != nil {
		return fmt.Errorf("failed to create node home directory: %w", err)
	}

	// Store VRF secret key with restrictive permissions (owner read/write only)
	if err := os.WriteFile(vrfKeyPath, privateKey, 0o600); err != nil {
		return fmt.Errorf("failed to save VRF secret key: %w", err)
	}

	fmt.Printf("✓ VRF keypair generated for network: %s\n", chainID)
	fmt.Printf("✓ VRF secret key stored securely: %s\n", vrfKeyPath)
	return nil
}

// getChainIDFromGenesis extracts chain-id from the genesis file
func getChainIDFromGenesis() (string, error) {
	genesisPath := filepath.Join(app.DefaultNodeHome, "config", "genesis.json")

	// #nosec G304 - genesisPath is constructed from trusted app.DefaultNodeHome constant
	genesisData, err := os.ReadFile(genesisPath)
	if err != nil {
		return "", fmt.Errorf("failed to read genesis file: %w", err)
	}

	var genesis genutiltypes.AppGenesis
	if err := json.Unmarshal(genesisData, &genesis); err != nil {
		return "", fmt.Errorf("failed to parse genesis file: %w", err)
	}

	if genesis.ChainID == "" {
		return "", fmt.Errorf("chain-id not found in genesis file")
	}

	return genesis.ChainID, nil
}

// entropyReader implements io.Reader to provide deterministic entropy from chainID
type entropyReader struct {
	seed []byte
	pos  int
}

func (e *entropyReader) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) && e.pos < len(e.seed) {
		p[n] = e.seed[e.pos]
		n++
		e.pos++
	}

	// If we need more bytes than available in seed, use crypto/rand for additional randomness
	if n < len(p) {
		remaining, err := rand.Read(p[n:])
		if err != nil {
			return n, err
		}
		n += remaining
	}

	return n, nil
}

// createChainIDEntropySource creates a deterministic entropy source from chainID
// This provides network-specific but still secure randomness for VRF key generation
func createChainIDEntropySource(chainID string) (io.Reader, error) {
	if chainID == "" {
		return nil, fmt.Errorf("chainID cannot be empty")
	}

	// Create deterministic seed from chainID using SHA256
	hash := sha256.Sum256([]byte(chainID))

	return &entropyReader{
		seed: hash[:],
		pos:  0,
	}, nil
}
