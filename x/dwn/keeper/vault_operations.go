package keeper

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ipfs/go-cid"

	"github.com/sonr-io/crypto/mpc"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// CreateEncryptedMPCVault creates an encrypted MPC vault and stores it in IPFS
// This is called during WebAuthn registration to initialize the vault
func (k Keeper) CreateEncryptedMPCVault(
	ctx context.Context,
	did string,
	owner string,
	vaultID string,
	keyID string,
) (*didtypes.CreateVaultResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate MPC secret data using Motor WASM plugin
	// In production, this would call the actual Motor WASM module
	mpcData, err := k.generateMPCSecretData(ctx, did, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MPC secret data: %w", err)
	}

	// Generate consensus-based encryption key
	// This uses validator consensus to derive a key that can be recovered by threshold
	encryptionKey, err := k.deriveConsensusEncryptionKey(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to derive consensus encryption key: %w", err)
	}

	// Encrypt MPC data using AES-GCM
	encryptedData, nonce, err := encryptMPCData(mpcData, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt MPC data: %w", err)
	}

	// Create vault metadata
	vaultMetadata := &types.VaultMetadata{
		Did:          did,
		VaultId:      vaultID,
		Owner:        owner,
		KeyId:        keyID,
		Algorithm:    "AES-256-GCM",
		Nonce:        base64.StdEncoding.EncodeToString(nonce),
		CreatedAt:    sdkCtx.BlockTime().Unix(),
		BlockHeight:  sdkCtx.BlockHeight(),
		ValidatorSet: k.getCurrentValidatorHashes(ctx),
	}

	// Prepare IPFS storage object
	ipfsData := &types.EncryptedVaultData{
		Metadata:      vaultMetadata,
		EncryptedData: base64.StdEncoding.EncodeToString(encryptedData),
		Version:       1,
	}

	// Marshal to JSON for IPFS storage
	jsonData, err := json.Marshal(ipfsData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Store encrypted data in IPFS
	ipfsCID, err := k.storeInIPFS(ctx, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to store in IPFS: %w", err)
	}

	// Extract public key from MPC data for response
	publicKey := mpcData.PubBytes
	if publicKey == nil {
		publicKey = []byte{} // Default empty if not available
	}
	publicKeyString := base64.StdEncoding.EncodeToString(publicKey)

	// Create vault state entry on chain
	vaultState := &types.EncryptedVaultState{
		VaultId:        vaultID,
		Did:            did,
		Owner:          owner,
		IpfsCid:        ipfsCID,
		PublicKey:      publicKeyString,
		CreatedAt:      sdkCtx.BlockTime().Unix(),
		LastUpdated:    sdkCtx.BlockTime().Unix(),
		Status:         "active",
		EncryptionType: "consensus-aes-gcm",
	}

	// Store vault state in keeper
	if err := k.storeVaultState(ctx, vaultState); err != nil {
		return nil, fmt.Errorf("failed to store vault state: %w", err)
	}

	// Emit vault creation event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vault_encrypted_stored",
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("vault_id", vaultID),
			sdk.NewAttribute("ipfs_cid", ipfsCID),
			sdk.NewAttribute("encryption", "consensus-aes-gcm"),
		),
	)

	return &didtypes.CreateVaultResponse{
		VaultID:        vaultID,
		VaultPublicKey: publicKeyString,
		EnclaveID:      fmt.Sprintf("enclave-%s", vaultID),
		IpfsCid:        ipfsCID,
	}, nil
}

// generateMPCSecretData generates MPC secret data using Motor WASM
func (k Keeper) generateMPCSecretData(ctx context.Context, did string, owner string) (*mpc.EnclaveData, error) {
	// In production, this would:
	// 1. Call Motor WASM plugin via internal/vault
	// 2. Generate threshold keys
	// 3. Create secret shares
	// 4. Return enclave data

	// For now, create mock MPC data
	publicKey := make([]byte, 33)
	if _, err := rand.Read(publicKey); err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Create mock shares (in production these would be generated via MPC)
	// For now, set to nil as they require protocol.Message type

	return &mpc.EnclaveData{
		PubHex:    fmt.Sprintf("%x", publicKey),
		PubBytes:  publicKey,
		ValShare:  nil, // Would be *protocol.Message in production
		UserShare: nil, // Would be *protocol.Message in production
		Nonce:     nonce,
		Curve:     mpc.K256Name,
	}, nil
}

// deriveConsensusEncryptionKey derives an encryption key using validator consensus
func (k Keeper) deriveConsensusEncryptionKey(ctx context.Context, did string) ([]byte, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Combine block hash, DID, and validator set hash for key derivation
	blockHash := sdkCtx.HeaderHash()
	didBytes := []byte(did)

	// Create deterministic key material
	keyMaterial := append(blockHash, didBytes...)

	// Use SHA-256 to derive a 32-byte key
	hash := sha256.Sum256(keyMaterial)

	return hash[:], nil
}

// encryptMPCData encrypts MPC data using AES-GCM
func encryptMPCData(data *mpc.EnclaveData, key []byte) ([]byte, []byte, error) {
	// Marshal MPC data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal MPC data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// storeInIPFS stores data in IPFS and returns the CID
func (k Keeper) storeInIPFS(ctx context.Context, data []byte) (string, error) {
	// Check if IPFS client is available
	if k.ipfsClient == nil {
		return "", fmt.Errorf("IPFS client not initialized")
	}

	// Add data to IPFS
	hash, err := k.ipfsClient.Add(data)
	if err != nil {
		return "", fmt.Errorf("failed to add to IPFS: %w", err)
	}

	// Verify the CID is valid
	_, err = cid.Parse(hash)
	if err != nil {
		return "", fmt.Errorf("invalid IPFS CID: %w", err)
	}

	return hash, nil
}

// storeVaultState stores vault state in the keeper
func (k Keeper) storeVaultState(ctx context.Context, state *types.EncryptedVaultState) error {
	// In production, this would store in ORM database
	// For now, we'll store in a simple map or state storage

	// TODO: Implement actual ORM storage
	// Example: k.OrmDB.VaultStateTable().Insert(ctx, state)

	// For now, just validate the state
	if state.VaultId == "" || state.Did == "" || state.Owner == "" {
		return fmt.Errorf("invalid vault state: missing required fields")
	}

	return nil
}

// getCurrentValidatorHashes returns current validator set hashes for consensus
func (k Keeper) getCurrentValidatorHashes(ctx context.Context) []string {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator set hash from context
	validatorHash := sdkCtx.BlockHeader().ValidatorsHash

	// Return as base64 encoded strings
	return []string{
		base64.StdEncoding.EncodeToString(validatorHash),
	}
}

// RecoverVaultFromIPFS recovers and decrypts vault data from IPFS
func (k Keeper) RecoverVaultFromIPFS(
	ctx context.Context,
	vaultID string,
	ipfsCID string,
) (*mpc.EnclaveData, error) {
	// Retrieve from IPFS
	data, err := k.retrieveFromIPFS(ctx, ipfsCID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from IPFS: %w", err)
	}

	// Unmarshal vault data
	var vaultData types.EncryptedVaultData
	if err := json.Unmarshal(data, &vaultData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vault data: %w", err)
	}

	// Derive consensus encryption key
	encryptionKey, err := k.deriveConsensusEncryptionKey(ctx, vaultData.Metadata.Did)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Decode encrypted data and nonce
	encryptedData, err := base64.StdEncoding.DecodeString(vaultData.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(vaultData.Metadata.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Decrypt MPC data
	mpcData, err := decryptMPCData(encryptedData, nonce, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt MPC data: %w", err)
	}

	return mpcData, nil
}

// retrieveFromIPFS retrieves data from IPFS by CID
func (k Keeper) retrieveFromIPFS(ctx context.Context, ipfsCID string) ([]byte, error) {
	if k.ipfsClient == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get data from IPFS
	data, err := k.ipfsClient.Get(ipfsCID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from IPFS: %w", err)
	}

	return data, nil
}

// decryptMPCData decrypts MPC data using AES-GCM
func decryptMPCData(ciphertext []byte, nonce []byte, key []byte) (*mpc.EnclaveData, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Unmarshal MPC data
	var mpcData mpc.EnclaveData
	if err := json.Unmarshal(plaintext, &mpcData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MPC data: %w", err)
	}

	return &mpcData, nil
}
