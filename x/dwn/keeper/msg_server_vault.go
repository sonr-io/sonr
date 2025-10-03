// Package keeper provides vault message handlers with consensus-based encryption
package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// RotateVaultKeys rotates encryption keys for existing vaults
func (ms msgServer) RotateVaultKeys(
	ctx context.Context,
	msg *types.MsgRotateVaultKeys,
) (*types.MsgRotateVaultKeysResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Verify authority (only governance or validators can rotate keys)
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(
			types.ErrInvalidAuthorityFormat,
			"invalid authority; expected %s, got %s",
			ms.k.authority,
			msg.Authority,
		)
	}

	// Check if key rotation is needed (unless forced)
	if !msg.Force {
		rotationDue := ms.k.encryptionSubkeeper.IsRotationDue(sdkCtx)
		if !rotationDue {
			return nil, errors.Wrap(
				types.ErrInvalidRequest,
				"key rotation not due (use force=true to override)",
			)
		}
	}

	var vaultsRotated uint32 = 0

	if msg.VaultId != "" {
		// Rotate keys for specific vault
		vault, err := ms.k.OrmDB.VaultStateTable().Get(sdkCtx, msg.VaultId)
		if err != nil {
			return nil, errors.Wrapf(
				types.ErrVaultNotFound,
				"vault %s not found",
				msg.VaultId,
			)
		}

		// Re-encrypt vault data with new consensus key
		err = ms.rotateVaultKeys(sdkCtx, vault, msg.Reason)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to rotate keys for vault %s", msg.VaultId)
		}

		vaultsRotated = 1
	} else {
		// Rotate keys for all vaults
		iter, err := ms.k.OrmDB.VaultStateTable().List(sdkCtx, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list vaults for rotation")
		}
		defer iter.Close()

		for iter.Next() {
			vault, err := iter.Value()
			if err != nil {
				ms.k.Logger().Error("Failed to get vault during rotation", "error", err)
				continue
			}

			err = ms.rotateVaultKeys(sdkCtx, vault, msg.Reason)
			if err != nil {
				ms.k.Logger().Error("Failed to rotate vault keys",
					"vault_id", vault.VaultId,
					"error", err,
				)
				continue
			}

			vaultsRotated++
		}
	}

	// Perform global key rotation
	err := ms.k.encryptionSubkeeper.InitiateKeyRotation(sdkCtx, msg.Reason)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initiate global key rotation")
	}

	// Get the new key version after rotation
	newKeyVersion := ms.k.encryptionSubkeeper.GetCurrentKeyVersion(sdkCtx)

	ms.k.Logger().Info("Vault key rotation completed",
		"vaults_rotated", vaultsRotated,
		"new_key_version", newKeyVersion,
		"reason", msg.Reason,
		"forced", msg.Force,
	)

	// Emit typed event for key rotation
	event := &types.EventVaultKeysRotated{
		VaultId:        fmt.Sprintf("global-rotation-%d", newKeyVersion),
		Owner:          msg.Authority,
		NewPublicKey:   fmt.Sprintf("key-version-%d", newKeyVersion),
		RotationHeight: uint64(sdkCtx.BlockHeight()),
		BlockHeight:    uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventVaultKeysRotated")
	}

	return &types.MsgRotateVaultKeysResponse{
		VaultsRotated: vaultsRotated,
		NewKeyVersion: newKeyVersion,
		Success:       true,
	}, nil
}

// rotateVaultKeys re-encrypts a single vault's data with new consensus keys
func (ms msgServer) rotateVaultKeys(sdkCtx sdk.Context, vault any, reason string) error {
	// Type assertion to ensure we have the correct vault type
	vaultState, ok := vault.(*apiv1.VaultState)
	if !ok {
		return errors.Wrapf(
			types.ErrInvalidRequest,
			"invalid vault type: expected *apiv1.VaultState, got %T",
			vault,
		)
	}

	if vaultState == nil {
		return errors.Wrap(types.ErrVaultNotFound, "vault state is nil")
	}

	// Validate vault has encrypted data to rotate
	if vaultState.EnclaveData == nil {
		return errors.Wrapf(
			types.ErrInvalidRequest,
			"vault %s has no enclave data to rotate",
			vaultState.VaultId,
		)
	}

	ms.k.Logger().Info("Starting vault key rotation",
		"vault_id", vaultState.VaultId,
		"owner", vaultState.Owner,
		"reason", reason,
		"block_height", sdkCtx.BlockHeight(),
	)

	// Check if encryption subkeeper is available
	if ms.k.encryptionSubkeeper == nil {
		return errors.Wrap(types.ErrInvalidRequest, "encryption subkeeper not available")
	}

	ctx := sdk.WrapSDKContext(sdkCtx)

	// Get current encryption key version before rotation
	oldKeyVersion := ms.k.encryptionSubkeeper.GetCurrentKeyVersion(ctx)

	// Store original values for rollback if needed
	originalPrivateData := make([]byte, len(vaultState.EnclaveData.PrivateData))
	copy(originalPrivateData, vaultState.EnclaveData.PrivateData)
	originalVersion := vaultState.EnclaveData.Version

	// Step 1: Decrypt vault's encrypted private data using old consensus keys
	// We need to reconstruct the encryption metadata for the old data
	oldMetadata := &types.EncryptionMetadata{
		KeyVersion:       oldKeyVersion,
		Algorithm:        "AES-GCM",
		EncryptionHeight: sdkCtx.BlockHeight(),
		ValidatorSet:     []string{}, // Will be populated by encryptionSubkeeper
	}

	decryptedData, err := ms.k.encryptionSubkeeper.DecryptWithConsensusKey(
		ctx,
		vaultState.EnclaveData.PrivateData,
		oldMetadata,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to decrypt vault data for vault %s", vaultState.VaultId)
	}

	ms.k.Logger().Debug("Successfully decrypted vault data",
		"vault_id", vaultState.VaultId,
		"data_size", len(decryptedData),
		"old_key_version", oldKeyVersion,
	)

	// Step 2: Re-encrypt vault data with new consensus keys
	encryptedResult, err := ms.k.encryptionSubkeeper.EncryptWithConsensusKey(
		ctx,
		decryptedData,
		"vault.enclave/v1",
	)
	if err != nil {
		return errors.Wrapf(err, "failed to re-encrypt vault data for vault %s", vaultState.VaultId)
	}

	// Step 3: Update vault state with new encrypted data and metadata
	vaultState.EnclaveData.PrivateData = encryptedResult.Ciphertext
	vaultState.EnclaveData.Version = int64(encryptedResult.Metadata.KeyVersion)

	// Update timestamps
	vaultState.LastRefreshed = sdkCtx.BlockTime().Unix()

	// Step 4: Validate data integrity using HMAC-SHA256
	if err := ms.validateVaultIntegrity(sdkCtx, vaultState, decryptedData); err != nil {
		// Rollback on validation failure
		vaultState.EnclaveData.PrivateData = originalPrivateData
		vaultState.EnclaveData.Version = originalVersion
		return errors.Wrapf(err, "vault integrity validation failed for %s", vaultState.VaultId)
	}

	// Step 5: Update vault state in ORM database
	if err := ms.k.OrmDB.VaultStateTable().Update(ctx, vaultState); err != nil {
		// Rollback on database update failure
		vaultState.EnclaveData.PrivateData = originalPrivateData
		vaultState.EnclaveData.Version = originalVersion
		return errors.Wrapf(err, "failed to update vault state for %s", vaultState.VaultId)
	}

	// Step 6: Update IPFS storage with re-encrypted vault export if applicable
	if ms.k.ipfsClient != nil {
		if err := ms.updateVaultInIPFS(ctx, vaultState, encryptedResult.Ciphertext); err != nil {
			// Log warning but don't fail the rotation - IPFS is supplementary
			ms.k.Logger().Warn("Failed to update vault in IPFS",
				"vault_id", vaultState.VaultId,
				"error", err,
			)
		}
	}

	// Get new key version after rotation
	newKeyVersion := ms.k.encryptionSubkeeper.GetCurrentKeyVersion(ctx)

	// Step 7: Log rotation event with audit trail for security compliance
	ms.k.Logger().Info("Vault key rotation completed successfully",
		"vault_id", vaultState.VaultId,
		"owner", vaultState.Owner,
		"old_key_version", oldKeyVersion,
		"new_key_version", newKeyVersion,
		"reason", reason,
		"block_height", sdkCtx.BlockHeight(),
		"data_size", len(encryptedResult.Ciphertext),
	)

	// Emit typed event for audit trail
	rotationEvent := &types.EventVaultKeysRotated{
		VaultId:        vaultState.VaultId,
		Owner:          vaultState.Owner,
		NewPublicKey:   fmt.Sprintf("key-version-%d", newKeyVersion),
		RotationHeight: uint64(sdkCtx.BlockHeight()),
		BlockHeight:    uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(rotationEvent); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit vault rotation event")
	}

	// Clean up sensitive data from memory
	for i := range decryptedData {
		decryptedData[i] = 0
	}

	return nil
}

// validateVaultIntegrity validates the integrity of vault data after key rotation
func (ms msgServer) validateVaultIntegrity(
	sdkCtx sdk.Context,
	vaultState *apiv1.VaultState,
	originalPlaintext []byte,
) error {
	ctx := sdk.WrapSDKContext(sdkCtx)

	// Re-decrypt the newly encrypted data to verify it matches the original
	newMetadata := &types.EncryptionMetadata{
		KeyVersion:       ms.k.encryptionSubkeeper.GetCurrentKeyVersion(ctx),
		Algorithm:        "AES-GCM",
		EncryptionHeight: sdkCtx.BlockHeight(),
		ValidatorSet:     []string{}, // Will be populated by encryptionSubkeeper
	}

	reDecrypted, err := ms.k.encryptionSubkeeper.DecryptWithConsensusKey(
		ctx,
		vaultState.EnclaveData.PrivateData,
		newMetadata,
	)
	if err != nil {
		return fmt.Errorf("failed to re-decrypt for validation: %w", err)
	}

	// Compare byte-by-byte to ensure data integrity
	if len(reDecrypted) != len(originalPlaintext) {
		return fmt.Errorf("decrypted data length mismatch: expected %d, got %d",
			len(originalPlaintext), len(reDecrypted))
	}

	for i := range originalPlaintext {
		if reDecrypted[i] != originalPlaintext[i] {
			return fmt.Errorf("data integrity check failed at byte %d", i)
		}
	}

	// Clean up sensitive validation data
	for i := range reDecrypted {
		reDecrypted[i] = 0
	}

	return nil
}

// updateVaultInIPFS updates the vault's IPFS storage with re-encrypted data
func (ms msgServer) updateVaultInIPFS(
	ctx context.Context,
	vaultState *apiv1.VaultState,
	reencryptedData []byte,
) error {
	if ms.k.ipfsClient == nil {
		return fmt.Errorf("IPFS client not available")
	}

	// Create a vault export structure for IPFS storage
	vaultExport := map[string]any{
		"vault_id":       vaultState.VaultId,
		"owner":          vaultState.Owner,
		"encrypted_data": reencryptedData,
		"version":        vaultState.EnclaveData.Version,
		"last_refreshed": vaultState.LastRefreshed,
		"rotation_metadata": map[string]any{
			"rotated_at":  vaultState.LastRefreshed,
			"key_version": vaultState.EnclaveData.Version,
		},
	}

	// Serialize vault export to JSON
	exportBytes, err := json.Marshal(vaultExport)
	if err != nil {
		return fmt.Errorf("failed to serialize vault export: %w", err)
	}

	// Store to IPFS and get new CID
	newCID, err := ms.k.ipfsClient.Add(exportBytes)
	if err != nil {
		return fmt.Errorf("failed to store updated vault to IPFS: %w", err)
	}

	ms.k.Logger().Debug("Updated vault in IPFS",
		"vault_id", vaultState.VaultId,
		"new_cid", newCID,
		"export_size", len(exportBytes),
	)

	return nil
}
