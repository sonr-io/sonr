package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/ipfs"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// GetIPFSClient returns the IPFS client for external access
func (k Keeper) GetIPFSClient() (ipfs.IPFSClient, error) {
	if k.ipfsClient == nil {
		client, err := ipfs.GetClient()
		if err != nil {
			return nil, err
		}
		k.ipfsClient = client
	}
	return k.ipfsClient, nil
}

// AddEnclaveDataToIPFS adds MPC enclave data to IPFS with consensus-based encryption
func (k Keeper) AddEnclaveDataToIPFS(
	ctx context.Context,
	data *mpc.EnclaveData,
) (*apiv1.VaultState, error) {
	// Input validation
	pubKey := data.PubKeyBytes()
	did, err := keys.NewFromMPCPubKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create DID from public key: %w", err)
	}

	owner, err := bech32.ConvertAndEncode("idx", pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key to DID: %w", err)
	}

	k.logger.Info("AddMPCEnclaveData called",
		"did", did,
		"owner", owner,
	)

	// Get IPFS client (lazy initialization)
	ipfsClient, err := k.GetIPFSClient()
	if err != nil {
		return nil, fmt.Errorf(
			"IPFS client not available - vault creation requires IPFS client: %w", err)
	}

	// Marshal enclave data to bytes
	enclaveBytes, err := data.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enclave data: %w", err)
	}

	// SECURITY: EnclaveData MUST ALWAYS be encrypted - it contains sensitive cryptographic material
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Encrypt enclave data using consensus-based encryption (mandatory for enclave data)
	encryptedData, err := k.encryptionSubkeeper.EncryptWithConsensusKey(
		sdkCtx,
		enclaveBytes,
		"vault.enclave/v1", // Use vault protocol for enclave data
	)
	if err != nil {
		// CRITICAL: Never store enclave data unencrypted - fail the operation instead
		k.logger.Error("SECURITY: Failed to encrypt enclave data - operation aborted",
			"error", err,
			"did", did,
		)
		return nil, fmt.Errorf("mandatory encryption failed for sensitive enclave data: %w", err)
	}

	// Store encrypted data and metadata
	dataToStore := encryptedData.Ciphertext
	encryptionMetadata := encryptedData.Metadata

	k.logger.Info("Enclave data encrypted successfully",
		"did", did,
		"encrypted_size", len(dataToStore),
		"original_size", len(enclaveBytes),
		"key_version", encryptedData.Metadata.KeyVersion,
	)

	// Store the encrypted data to IPFS
	vaultCID, err := ipfsClient.Add(dataToStore)
	if err != nil {
		return nil, fmt.Errorf("failed to store vault data to IPFS: %w", err)
	}

	// Store the vault state in the database with encryption metadata
	vaultState := &apiv1.VaultState{
		VaultId:       vaultCID,
		Owner:         owner,
		PublicKey:     pubKey,
		CreatedAt:     time.Now().Unix(),
		LastRefreshed: time.Now().Unix(),
		CreatedHeight: sdkCtx.BlockHeight(), // Will be set by the block height in the message server
		EnclaveData: &apiv1.EnclaveData{
			PrivateData: dataToStore,
			PublicKey:   pubKey,
			EnclaveId:   vaultCID,
			Version:     1,
		},
	}

	// Store encryption metadata on-chain (always present for enclave data)
	apiMetadata := encryptionMetadata.ToAPIEncryptionMetadata()
	vaultState.EncryptionMetadata = apiMetadata

	k.logger.Debug("Stored encryption metadata with vault",
		"vault_id", vaultCID,
		"algorithm", apiMetadata.Algorithm,
		"key_version", apiMetadata.KeyVersion,
		"block_height", sdkCtx.BlockHeight(),
	)
	return vaultState, nil
}

// GetEnclaveDataFromIPFS retrieves MPC enclave data from IPFS with consensus-based encryption
func (k Keeper) GetEnclaveDataFromIPFS(
	ctx context.Context,
	cid string,
	encryptionMetadata *types.EncryptionMetadata,
) (*mpc.EnclaveData, error) {
	ipfsClient, err := k.GetIPFSClient()
	if err != nil {
		return nil, fmt.Errorf("IPFS client not available for operations: %w", err)
	}

	if cid == "" {
		return nil, fmt.Errorf("CID cannot be empty")
	}

	// Retrieve encrypted data from IPFS
	encryptedData, err := ipfsClient.Get(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from IPFS: %w", err)
	}

	if encryptionMetadata == nil {
		// SECURITY: EnclaveData must always have encryption metadata
		k.logger.Error("SECURITY: Missing encryption metadata for enclave data retrieval",
			"cid", cid,
		)
		return nil, fmt.Errorf(
			"missing encryption metadata - enclave data must always be encrypted",
		)
	}

	// Decrypt the data using the encryption subkeeper
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	decryptedData, err := k.encryptionSubkeeper.DecryptWithConsensusKey(
		sdkCtx,
		encryptedData,
		encryptionMetadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	k.logger.Debug("Successfully retrieved and decrypted data from IPFS",
		"cid", cid,
		"encrypted_size", len(encryptedData),
		"decrypted_size", len(decryptedData),
		"algorithm", encryptionMetadata.Algorithm,
	)
	data := &mpc.EnclaveData{}
	if err := data.Unmarshal(decryptedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted enclave data: %w", err)
	}
	return data, nil
}

// StoreEncryptedToIPFS stores encrypted data to IPFS with metadata tracking
func (k Keeper) StoreEncryptedToIPFS(
	ctx context.Context,
	data []byte,
	protocol string,
) (string, error) {
	ipfsClient, err := k.GetIPFSClient()
	if err != nil {
		return "", fmt.Errorf("IPFS client not available for operations: %w", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("cannot store empty data")
	}

	k.logger.Info("Storing encrypted data to IPFS",
		"data_size", len(data),
		"protocol", protocol,
	)

	// Store the encrypted data directly to IPFS
	// The data is assumed to already be encrypted by the caller
	cid, err := ipfsClient.Add(data)
	if err != nil {
		return "", fmt.Errorf("failed to store encrypted data to IPFS: %w", err)
	}

	k.logger.Debug("Successfully stored encrypted data to IPFS",
		"cid", cid,
		"protocol", protocol,
		"size", len(data),
	)

	return cid, nil
}

// RetrieveAndDecryptFromIPFS retrieves encrypted data from IPFS and decrypts it
func (k Keeper) RetrieveAndDecryptFromIPFS(
	ctx context.Context,
	cid string,
	encryptionMetadata *types.EncryptionMetadata,
) ([]byte, error) {
	ipfsClient, err := k.GetIPFSClient()
	if err != nil {
		return nil, fmt.Errorf("IPFS client not available for operations: %w", err)
	}

	if cid == "" {
		return nil, fmt.Errorf("CID cannot be empty")
	}

	// Retrieve encrypted data from IPFS
	encryptedData, err := ipfsClient.Get(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from IPFS: %w", err)
	}

	if encryptionMetadata == nil {
		// Data is unencrypted, return as-is
		k.logger.Debug("Retrieved unencrypted data from IPFS",
			"cid", cid,
			"size", len(encryptedData),
		)
		return encryptedData, nil
	}

	// Decrypt the data using the encryption subkeeper
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	decryptedData, err := k.encryptionSubkeeper.DecryptWithConsensusKey(
		sdkCtx,
		encryptedData,
		encryptionMetadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	k.logger.Debug("Successfully retrieved and decrypted data from IPFS",
		"cid", cid,
		"encrypted_size", len(encryptedData),
		"decrypted_size", len(decryptedData),
		"algorithm", encryptionMetadata.Algorithm,
	)
	return decryptedData, nil
}
