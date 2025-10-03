// Package context provides the Sonr context system for managing node-specific state.
package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/sonr-io/sonr/crypto/vrf"
)

// SonrContext manages node-specific state and configuration
type SonrContext struct {
	logger log.Logger

	// VRF keypair for the node
	vrfPrivateKey vrf.PrivateKey
	vrfPublicKey  vrf.PublicKey

	// Client context for transaction operations
	clientCtx client.Context

	// Synchronization for thread-safe access
	mu sync.RWMutex

	// Initialization state
	initialized bool
}

// NewSonrContext creates a new SonrContext instance
func NewSonrContext(logger log.Logger) *SonrContext {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &SonrContext{
		logger:      logger.With("component", "sonr-context"),
		initialized: false,
	}
}

// SetClientContext sets the client context for transaction operations (thread-safe)
func (sc *SonrContext) SetClientContext(clientCtx client.Context) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.clientCtx = clientCtx
}

// GetClientContext returns the client context (thread-safe)
func (sc *SonrContext) GetClientContext() (client.Context, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if sc.clientCtx.Codec == nil {
		return client.Context{}, fmt.Errorf("client context not initialized")
	}

	return sc.clientCtx, nil
}

// Initialize loads the VRF keypair from storage
func (sc *SonrContext) Initialize() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.initialized {
		return nil
	}

	// Load VRF private key from storage
	// Use hardcoded default path to avoid import cycle
	defaultNodeHome := os.ExpandEnv("$HOME/.sonr")
	vrfKeyPath := filepath.Join(defaultNodeHome, "vrf_secret.key")

	// #nosec G304 - vrfKeyPath is constructed from trusted DefaultNodeHome constant
	vrfKeyData, err := os.ReadFile(vrfKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read VRF secret key from %s: %w\n"+
			"VRF keys are required for multi-validator encryption features.\n"+
			"To generate VRF keys:\n"+
			"  1. For new nodes: Run 'snrd init <moniker>' to initialize with VRF keys\n"+
			"  2. For existing nodes: VRF keys should have been generated during init\n"+
			"  3. If encryption is not needed, disable it in DWN module params",
			vrfKeyPath, err)
	}

	// Validate key size
	if len(vrfKeyData) != vrf.PrivateKeySize {
		return fmt.Errorf(
			"invalid VRF private key size: expected %d, got %d",
			vrf.PrivateKeySize,
			len(vrfKeyData),
		)
	}

	sc.vrfPrivateKey = vrf.PrivateKey(vrfKeyData)

	// Derive public key
	publicKey, ok := sc.vrfPrivateKey.Public()
	if !ok {
		return fmt.Errorf("failed to derive VRF public key from private key")
	}

	sc.vrfPublicKey = publicKey
	sc.initialized = true

	sc.logger.Info("SonrContext initialized successfully",
		"vrf_key_path", vrfKeyPath,
		"public_key_size", len(sc.vrfPublicKey),
	)

	return nil
}

// GetVRFPrivateKey returns the VRF private key (thread-safe)
func (sc *SonrContext) GetVRFPrivateKey() (vrf.PrivateKey, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.initialized {
		return nil, fmt.Errorf("SonrContext not initialized")
	}

	return sc.vrfPrivateKey, nil
}

// GetVRFPublicKey returns the VRF public key (thread-safe)
func (sc *SonrContext) GetVRFPublicKey() (vrf.PublicKey, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.initialized {
		return nil, fmt.Errorf("SonrContext not initialized")
	}

	return sc.vrfPublicKey, nil
}

// IsInitialized returns whether the context has been initialized (thread-safe)
func (sc *SonrContext) IsInitialized() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.initialized
}

// ComputeVRF generates VRF output for the given input using the loaded private key
func (sc *SonrContext) ComputeVRF(input []byte) ([]byte, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.initialized {
		return nil, fmt.Errorf("SonrContext not initialized")
	}

	if len(input) == 0 {
		return nil, fmt.Errorf("VRF input cannot be empty")
	}

	return sc.vrfPrivateKey.Compute(input), nil
}

// ProveVRF generates VRF output with proof for the given input
func (sc *SonrContext) ProveVRF(input []byte) (vrf []byte, proof []byte, err error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.initialized {
		return nil, nil, fmt.Errorf("SonrContext not initialized")
	}

	if len(input) == 0 {
		return nil, nil, fmt.Errorf("VRF input cannot be empty")
	}

	vrf, proof = sc.vrfPrivateKey.Prove(input)
	return vrf, proof, nil
}

// Global context instance (initialized by the node)
var globalSonrContext *SonrContext

// SetGlobalSonrContext sets the global SonrContext instance
func SetGlobalSonrContext(ctx *SonrContext) {
	globalSonrContext = ctx
}

// GetGlobalSonrContext returns the global SonrContext instance
func GetGlobalSonrContext() *SonrContext {
	return globalSonrContext
}
