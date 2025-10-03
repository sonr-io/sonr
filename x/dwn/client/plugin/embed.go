package plugin

import (
	"bytes"
	"compress/zlib"
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/sonr-io/sonr/crypto/wasm"
)

// motrPluginBytes contains the embedded WebAssembly bytecode for the cryptographic enclave.
// This is embedded at compile time and loaded into the WASM runtime for secure operations.
//
//go:embed vault.wasm
var motrPluginBytes []byte

// motrPluginHash is the SHA256 hash of the embedded WASM module
// This will be computed at runtime for verification
var motrPluginHash string

// hashVerifier is the global hash verifier for WASM modules
var hashVerifier = wasm.NewHashVerifier()

// signatureVerifier is the global signature verifier for WASM modules
var signatureVerifier = wasm.NewSignatureVerifier()

// pluginSignatureManifest stores the signature manifest for the plugin
var pluginSignatureManifest *wasm.SignatureManifest

// init initializes the hash and signature verifiers
func init() {
	// Compute hash of embedded WASM module
	motrPluginHash = hashVerifier.ComputeHash(motrPluginBytes)
	// Add as trusted hash
	hashVerifier.AddTrustedHash("motr", motrPluginHash)

	// Initialize signature verification (signatures will be added via configuration)
	// In production, trusted keys would be loaded from secure configuration
	initializeTrustedSigningKeys()
}

// initializeTrustedSigningKeys loads trusted signing keys for verification
func initializeTrustedSigningKeys() {
	// These would typically come from secure configuration
	// For now, we'll prepare the infrastructure for key management
	// Production keys would be loaded from config files or environment
}

// MotrPluginRaw contains the raw WebAssembly bytecode for the cryptographic enclave.
func MotrPluginRaw() ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	defer w.Close()
	_, err := w.Write(motrPluginBytes)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GetManifest returns the WebAssembly manifest configuration for the MPC-based UCAN enclave plugin.
// This manifest specifies the WASM bytecode and configuration required to run
// the MPC-based UCAN token operations.
func GetManifest() extism.Manifest {
	return extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: motrPluginBytes,
			},
		},
		Config: map[string]string{},
	}
}

// GetManifestWithEnclave returns a WebAssembly manifest with MPC enclave configuration.
// This allows passing enclave data and vault configuration to the Motor plugin via PDK environment.
// DEPRECATED: Use GetManifestFromConfig for enhanced configuration support.
func GetManifestWithEnclave(
	chainID string,
	enclaveData []byte,
	vaultConfig map[string]any,
) extism.Manifest {
	// Prepare configuration for PDK environment variables
	config := map[string]string{
		"chain_id": chainID,
	}

	// Add enclave data as JSON-encoded environment variable
	if len(enclaveData) > 0 {
		config["enclave"] = string(enclaveData) // Motor plugin expects JSON-encoded enclave data
	}

	// Add vault configuration if provided
	if len(vaultConfig) > 0 {
		if configBytes, err := json.Marshal(vaultConfig); err == nil {
			config["vault_config"] = string(configBytes)
		}
	}

	return extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: motrPluginBytes,
			},
		},
		Config: config,
	}
}

// GetManifestFromConfig creates a WebAssembly manifest from an EnclaveConfig.
// This is the preferred method for creating manifests with comprehensive configuration.
func GetManifestFromConfig(config *EnclaveConfig) (extism.Manifest, error) {
	manifestConfig, err := config.ToManifestConfig()
	if err != nil {
		return extism.Manifest{}, fmt.Errorf(
			"failed to convert enclave config to manifest config: %w",
			err,
		)
	}

	return extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: motrPluginBytes,
			},
		},
		Config: manifestConfig,
	}, nil
}

// ValidateManifest validates that a manifest contains required configuration.
func ValidateManifest(manifest extism.Manifest) error {
	if len(manifest.Wasm) == 0 {
		return fmt.Errorf("manifest must contain WASM data")
	}

	// Check for required configuration keys
	requiredKeys := []string{"chain_id"}
	for _, key := range requiredKeys {
		if _, exists := manifest.Config[key]; !exists {
			return fmt.Errorf("manifest missing required config key: %s", key)
		}
	}

	// Validate enclave data if present
	if enclaveStr, exists := manifest.Config["enclave"]; exists && enclaveStr != "" {
		var enclaveData map[string]any
		if err := json.Unmarshal([]byte(enclaveStr), &enclaveData); err != nil {
			return fmt.Errorf("invalid enclave data in manifest: %w", err)
		}
	}

	return nil
}

// GetPluginConfig returns the configuration for the WebAssembly plugin runtime.
// It enables WASI (WebAssembly System Interface) for file system and system call access.
func GetPluginConfig() extism.PluginConfig {
	return extism.PluginConfig{
		EnableWasi: true,
	}
}

// VerifyPluginIntegrity verifies the integrity of the WASM plugin
func VerifyPluginIntegrity(wasmBytes []byte) error {
	return hashVerifier.VerifyHash("motr", wasmBytes)
}

// GetPluginHash returns the SHA256 hash of the embedded WASM module
func GetPluginHash() string {
	return motrPluginHash
}

// VerifyPluginSignature verifies the signature of the WASM plugin
func VerifyPluginSignature(wasmBytes []byte, signature []byte) error {
	return signatureVerifier.Verify(wasmBytes, signature)
}

// SetPluginSignatureManifest sets the signature manifest for the plugin
func SetPluginSignatureManifest(manifest *wasm.SignatureManifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest cannot be nil")
	}
	pluginSignatureManifest = manifest

	// Load trusted keys from manifest
	for _, key := range manifest.TrustedKeys {
		if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
			continue // Skip expired keys
		}

		publicKey, err := base64.StdEncoding.DecodeString(key.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}

		if err := signatureVerifier.AddTrustedKey(key.KeyID, ed25519.PublicKey(publicKey)); err != nil {
			return fmt.Errorf("failed to add trusted key: %w", err)
		}
	}

	return nil
}

// GetPluginSignatureManifest returns the current signature manifest
func GetPluginSignatureManifest() *wasm.SignatureManifest {
	return pluginSignatureManifest
}

// AddTrustedSigningKey adds a trusted public key for signature verification
func AddTrustedSigningKey(keyID string, publicKeyHex string) error {
	return signatureVerifier.AddTrustedKeyFromHex(keyID, publicKeyHex)
}
