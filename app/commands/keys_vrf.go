// Package commands contains utility functions for the snrd command
package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sonr-io/sonr/app"
	"github.com/sonr-io/crypto/vrf"
	"github.com/spf13/cobra"
)

// VRFKeysCmd returns the VRF key management command
func VRFKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrf",
		Short: "Manage VRF (Verifiable Random Function) keys",
		Long: `Manage VRF keys used for consensus-based encryption in multi-validator networks.

VRF keys are required for:
  - Multi-validator encryption key derivation
  - Consensus-based key rotation
  - Deterministic randomness in distributed systems

VRF keys are automatically generated during 'snrd init', but this command
allows you to inspect or regenerate them if needed.`,
	}

	cmd.AddCommand(
		vrfGenerateCmd(),
		vrfShowCmd(),
		vrfVerifyCmd(),
	)

	return cmd
}

// vrfGenerateCmd creates a command to generate new VRF keys
func vrfGenerateCmd() *cobra.Command {
	var chainID string
	var force bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate new VRF keypair",
		Long: `Generate a new VRF keypair for the node.

The keypair is generated deterministically from the chain-id to ensure
reproducibility. Keys are stored at ~/.sonr/vrf_secret.key with 0600 permissions.

WARNING: Regenerating VRF keys will invalidate any existing consensus-based
encryption keys. Only do this if you understand the implications.`,
		Example: `  # Generate VRF keys for a specific chain
  snrd keys vrf generate --chain-id sonrtest_1-1

  # Force regenerate (overwrite existing keys)
  snrd keys vrf generate --chain-id sonrtest_1-1 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get chain-id
			if chainID == "" {
				// Try to read from genesis file
				var err error
				chainID, err = getChainIDFromGenesis()
				if err != nil {
					return fmt.Errorf("chain-id not provided and could not be read from genesis: %w\n"+
						"Please provide --chain-id flag", err)
				}
			}

			vrfKeyPath := filepath.Join(app.DefaultNodeHome, "vrf_secret.key")

			// Check if keys already exist
			if _, err := os.Stat(vrfKeyPath); err == nil && !force {
				return fmt.Errorf("VRF keys already exist at %s\n"+
					"Use --force flag to overwrite existing keys\n"+
					"WARNING: Overwriting keys will invalidate existing consensus encryption", vrfKeyPath)
			}

			// Backup existing keys if they exist
			if _, err := os.Stat(vrfKeyPath); err == nil && force {
				backupPath := vrfKeyPath + ".backup"
				if err := os.Rename(vrfKeyPath, backupPath); err != nil {
					return fmt.Errorf("failed to backup existing VRF keys: %w", err)
				}
				fmt.Printf("Backed up existing keys to: %s\n", backupPath)
			}

			// Generate VRF keys
			fmt.Printf("Generating VRF keypair for chain-id: %s\n", chainID)

			entropySource, err := createChainIDEntropySource(chainID)
			if err != nil {
				return fmt.Errorf("failed to create entropy source: %w", err)
			}

			privateKey, err := vrf.GenerateKey(entropySource)
			if err != nil {
				return fmt.Errorf("failed to generate VRF keypair: %w", err)
			}

			// Validate keypair
			if len(privateKey) != vrf.PrivateKeySize {
				return fmt.Errorf("invalid VRF private key size: expected %d, got %d",
					vrf.PrivateKeySize, len(privateKey))
			}

			publicKey, ok := privateKey.Public()
			if !ok {
				return fmt.Errorf("failed to derive public key from VRF private key")
			}

			// Ensure directory exists
			if err := os.MkdirAll(app.DefaultNodeHome, 0o750); err != nil {
				return fmt.Errorf("failed to create node home directory: %w", err)
			}

			// Store VRF secret key with restrictive permissions
			if err := os.WriteFile(vrfKeyPath, privateKey, 0o600); err != nil {
				return fmt.Errorf("failed to save VRF secret key: %w", err)
			}

			fmt.Printf("✓ VRF keypair generated successfully\n")
			fmt.Printf("✓ Private key stored at: %s\n", vrfKeyPath)
			fmt.Printf("✓ Public key (hex): %s\n", hex.EncodeToString(publicKey))
			fmt.Printf("✓ Key size: %d bytes\n", len(privateKey))
			fmt.Printf("✓ Permissions: 0600 (owner read/write only)\n")

			return nil
		},
	}

	cmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID for deterministic key generation")
	cmd.Flags().BoolVar(&force, "force", false, "Force overwrite existing VRF keys (creates backup)")

	return cmd
}

// vrfShowCmd creates a command to display VRF key information
func vrfShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show VRF public key information",
		Long: `Display information about the node's VRF public key.

This command shows the public key without exposing the private key.
Useful for verifying that VRF keys are properly installed.`,
		Example: `  # Show VRF key information
  snrd keys vrf show`,
		RunE: func(cmd *cobra.Command, args []string) error {
			vrfKeyPath := filepath.Join(app.DefaultNodeHome, "vrf_secret.key")

			// Read VRF private key
			// #nosec G304 - vrfKeyPath is constructed from trusted DefaultNodeHome constant
			privateKeyData, err := os.ReadFile(vrfKeyPath)
			if err != nil {
				return fmt.Errorf("failed to read VRF keys from %s: %w\n"+
					"VRF keys may not be initialized. Run 'snrd init' or 'snrd keys vrf generate'",
					vrfKeyPath, err)
			}

			// Validate key size
			if len(privateKeyData) != vrf.PrivateKeySize {
				return fmt.Errorf("invalid VRF private key size: expected %d, got %d",
					vrf.PrivateKeySize, len(privateKeyData))
			}

			privateKey := vrf.PrivateKey(privateKeyData)
			publicKey, ok := privateKey.Public()
			if !ok {
				return fmt.Errorf("failed to derive public key from VRF private key")
			}

			// Get file info
			fileInfo, err := os.Stat(vrfKeyPath)
			if err != nil {
				return fmt.Errorf("failed to stat VRF key file: %w", err)
			}

			fmt.Println("VRF Key Information:")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("Key Path:        %s\n", vrfKeyPath)
			fmt.Printf("Public Key:      %s\n", hex.EncodeToString(publicKey))
			fmt.Printf("Key Size:        %d bytes\n", len(privateKeyData))
			fmt.Printf("Public Key Size: %d bytes\n", len(publicKey))
			fmt.Printf("Permissions:     %s\n", fileInfo.Mode().Perm())
			fmt.Printf("Modified:        %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

			// Verify permissions
			if fileInfo.Mode().Perm() != 0o600 {
				fmt.Println("\n⚠️  WARNING: VRF key file has incorrect permissions!")
				fmt.Printf("   Current: %s, Expected: 0600\n", fileInfo.Mode().Perm())
				fmt.Println("   Run: chmod 0600 " + vrfKeyPath)
			} else {
				fmt.Println("\n✓ VRF keys are properly configured")
			}

			return nil
		},
	}

	return cmd
}

// vrfVerifyCmd creates a command to verify VRF key functionality
func vrfVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify VRF key functionality",
		Long: `Verify that VRF keys are properly installed and functional.

This command performs a test VRF computation to ensure the keys work correctly.`,
		Example: `  # Verify VRF keys
  snrd keys vrf verify`,
		RunE: func(cmd *cobra.Command, args []string) error {
			vrfKeyPath := filepath.Join(app.DefaultNodeHome, "vrf_secret.key")

			// Read VRF private key
			// #nosec G304 - vrfKeyPath is constructed from trusted DefaultNodeHome constant
			privateKeyData, err := os.ReadFile(vrfKeyPath)
			if err != nil {
				return fmt.Errorf("failed to read VRF keys: %w", err)
			}

			if len(privateKeyData) != vrf.PrivateKeySize {
				return fmt.Errorf("invalid VRF private key size: expected %d, got %d",
					vrf.PrivateKeySize, len(privateKeyData))
			}

			privateKey := vrf.PrivateKey(privateKeyData)
			publicKey, ok := privateKey.Public()
			if !ok {
				return fmt.Errorf("failed to derive public key")
			}

			// Test VRF computation with proof
			testInput := []byte("test-input-for-verification")
			vrfOutput, proof := privateKey.Prove(testInput)

			// Verify the output
			if !publicKey.Verify(testInput, vrfOutput, proof) {
				return fmt.Errorf("VRF verification failed - keys may be corrupted")
			}

			// Compute output hash
			outputHash := sha256.Sum256(vrfOutput)

			fmt.Println("VRF Key Verification:")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("✓ VRF keys loaded successfully\n")
			fmt.Printf("✓ Public key derived successfully\n")
			fmt.Printf("✓ VRF proof generated successfully\n")
			fmt.Printf("✓ VRF verification successful\n")
			fmt.Printf("\nTest Output (hex): %s\n", hex.EncodeToString(vrfOutput))
			fmt.Printf("Proof (hex):       %s\n", hex.EncodeToString(proof))
			fmt.Printf("Output Hash:       %s\n", hex.EncodeToString(outputHash[:]))

			fmt.Println("\n✓ VRF keys are fully functional")

			return nil
		},
	}

	return cmd
}
