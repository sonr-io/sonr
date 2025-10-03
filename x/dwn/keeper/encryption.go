// Package keeper provides consensus-based encryption for DWN records using VRF
package keeper

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/hkdf"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// EncryptionSubkeeper handles consensus-based encryption operations
type EncryptionSubkeeper struct {
	keeper       *Keeper
	logger       log.Logger
	vrfConsensus *VRFConsensus
}

// EncryptedData contains encrypted data with metadata
type EncryptedData struct {
	Ciphertext []byte
	Metadata   *types.EncryptionMetadata
}

// NewEncryptionSubkeeper creates a new encryption subkeeper
func NewEncryptionSubkeeper(k *Keeper) *EncryptionSubkeeper {
	es := &EncryptionSubkeeper{
		keeper: k,
		logger: k.logger.With("module", "encryption"),
	}
	es.vrfConsensus = NewVRFConsensus(k)
	return es
}

// DeriveConsensusKey creates shared encryption key using VRF consensus
func (es *EncryptionSubkeeper) DeriveConsensusKey(
	ctx context.Context,
	consensusInput []byte,
) ([]byte, error) {
	if len(consensusInput) == 0 {
		return nil, fmt.Errorf("consensus input cannot be empty")
	}

	// Use existing VRF computation from keeper
	vrfOutput, err := es.keeper.ComputeVRF(consensusInput)
	if err != nil {
		return nil, fmt.Errorf("VRF computation failed: %w", err)
	}

	// Derive AES-256 key using HKDF with SHA-256
	hkdfReader := hkdf.New(sha256.New, vrfOutput, nil, []byte("dwn-consensus-encryption"))

	// Generate 32 bytes for AES-256
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	es.logger.Debug("Consensus key derived",
		"input_len", len(consensusInput),
		"vrf_output_len", len(vrfOutput),
		"key_len", len(key),
	)

	return key, nil
}

// EncryptData encrypts data with consensus-derived key
func (es *EncryptionSubkeeper) EncryptData(
	ctx context.Context,
	plaintext []byte,
	consensusInput []byte,
	encryptionHeight int64,
) (*EncryptedData, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	// Derive consensus-based encryption key
	key, err := es.DeriveConsensusKey(ctx, consensusInput)
	if err != nil {
		return nil, fmt.Errorf("failed to derive consensus key: %w", err)
	}

	// Initialize AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, randErr := io.ReadFull(rand.Reader, nonce); randErr != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", randErr)
	}

	// Compute HMAC-SHA256 for data integrity verification before encryption
	hmacKey := es.deriveHMACKey(key, nonce)
	dataHMAC := es.computeHMAC(plaintext, hmacKey)

	// Encrypt data with authenticated encryption
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Extract authentication tag (last 16 bytes of GCM ciphertext)
	if len(ciphertext) < 16 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	authTag := ciphertext[len(ciphertext)-16:]
	actualCiphertext := ciphertext[:len(ciphertext)-16]

	// Create encryption metadata
	metadata := &types.EncryptionMetadata{
		Algorithm:        "AES-256-GCM",
		ConsensusInput:   consensusInput,
		Nonce:            nonce,
		AuthTag:          authTag,
		EncryptionHeight: encryptionHeight,
		ValidatorSet:     es.getValidatorSet(ctx),
		DataHmac:         dataHMAC,
	}

	es.logger.Info("Data encrypted successfully",
		"plaintext_len", len(plaintext),
		"ciphertext_len", len(actualCiphertext),
		"nonce_len", len(nonce),
		"auth_tag_len", len(authTag),
		"encryption_height", encryptionHeight,
	)

	return &EncryptedData{
		Ciphertext: actualCiphertext,
		Metadata:   metadata,
	}, nil
}

// DecryptData decrypts data using consensus-derived key
func (es *EncryptionSubkeeper) DecryptData(
	ctx context.Context,
	encryptedData *EncryptedData,
) ([]byte, error) {
	if encryptedData == nil || encryptedData.Metadata == nil {
		return nil, fmt.Errorf("encrypted data or metadata cannot be nil")
	}

	metadata := encryptedData.Metadata

	// Validate encryption metadata
	if metadata.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", metadata.Algorithm)
	}

	// Derive the same consensus key used for encryption
	key, err := es.DeriveConsensusKey(ctx, metadata.ConsensusInput)
	if err != nil {
		return nil, fmt.Errorf("failed to derive consensus key: %w", err)
	}

	// Initialize AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Validate nonce size
	if len(metadata.Nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf(
			"invalid nonce size: got %d, expected %d",
			len(metadata.Nonce),
			gcm.NonceSize(),
		)
	}

	// Reconstruct full ciphertext (data + auth tag)
	fullCiphertext := make([]byte, len(encryptedData.Ciphertext)+len(metadata.AuthTag))
	copy(fullCiphertext, encryptedData.Ciphertext)
	copy(fullCiphertext[len(encryptedData.Ciphertext):], metadata.AuthTag)

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, metadata.Nonce, fullCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (authentication failure): %w", err)
	}

	// Verify HMAC for data integrity
	if len(metadata.DataHmac) == 0 {
		return nil, fmt.Errorf("missing HMAC in encryption metadata")
	}

	hmacKey := es.deriveHMACKey(key, metadata.Nonce)
	hmacValid := es.verifyHMAC(plaintext, hmacKey, metadata.DataHmac)
	if !hmacValid {
		return nil, fmt.Errorf("HMAC verification failed: data integrity compromised")
	}

	// Log HMAC verification success
	es.logger.Debug("HMAC verification successful",
		"hmac_len", len(metadata.DataHmac),
		"plaintext_len", len(plaintext),
	)

	es.logger.Info("Data decrypted successfully",
		"ciphertext_len", len(encryptedData.Ciphertext),
		"plaintext_len", len(plaintext),
		"encryption_height", metadata.EncryptionHeight,
	)

	return plaintext, nil
}

// ValidateEncryptionMetadata validates encryption metadata for security
func (es *EncryptionSubkeeper) ValidateEncryptionMetadata(
	metadata *types.EncryptionMetadata,
) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	// Validate algorithm
	if metadata.Algorithm != "AES-256-GCM" {
		return fmt.Errorf("unsupported encryption algorithm: %s", metadata.Algorithm)
	}

	// Validate consensus input
	if len(metadata.ConsensusInput) == 0 {
		return fmt.Errorf("consensus input cannot be empty")
	}

	// Validate nonce
	if len(metadata.Nonce) != 12 { // GCM standard nonce size
		return fmt.Errorf("invalid nonce size: got %d, expected 12", len(metadata.Nonce))
	}

	// Validate auth tag
	if len(metadata.AuthTag) != 16 { // GCM auth tag size
		return fmt.Errorf("invalid auth tag size: got %d, expected 16", len(metadata.AuthTag))
	}

	// Validate encryption height
	if metadata.EncryptionHeight < 0 {
		return fmt.Errorf("encryption height cannot be negative")
	}

	return nil
}

// getValidatorSet returns current validator set for consensus tracking
func (es *EncryptionSubkeeper) getValidatorSet(ctx context.Context) []string {
	// Handle case when stakingKeeper is nil (for testing scenarios)
	if es.keeper.stakingKeeper == nil {
		es.logger.Debug("stakingKeeper is nil, returning empty validator set")
		return []string{}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Query bonded validators ordered by power (descending)
	validators, err := es.keeper.stakingKeeper.GetBondedValidatorsByPower(sdkCtx)
	if err != nil {
		es.logger.Error("Failed to get bonded validators", "error", err)
		return []string{}
	}

	if len(validators) == 0 {
		es.logger.Debug("No bonded validators found")
		return []string{}
	}

	// Get module parameters for minimum validator threshold
	params, err := es.keeper.Params.Get(ctx)
	if err != nil {
		es.logger.Error("Failed to get module parameters", "error", err)
		// Continue with default threshold if params unavailable
		params.MinValidatorsForKeyGen = types.DefaultMinValidatorsForKeyGen
	}

	// Calculate minimum number of validators needed based on percentage
	totalValidators := len(validators)
	minValidators := (totalValidators * int(params.MinValidatorsForKeyGen)) / 100
	if minValidators < 1 {
		minValidators = 1 // Always include at least one validator
	}

	// Filter active validators (take up to minValidators count)
	activeValidators := validators
	if len(validators) > minValidators {
		activeValidators = validators[:minValidators]
	}

	// Extract validator operator addresses
	validatorAddresses := make([]string, 0, len(activeValidators))
	for _, validator := range activeValidators {
		// Get validator operator address (not consensus address)
		valAddr := validator.GetOperator()
		if valAddr != "" {
			validatorAddresses = append(validatorAddresses, valAddr)
		}
	}

	es.logger.Debug("Retrieved validator set for consensus tracking",
		"total_bonded", totalValidators,
		"min_threshold_pct", params.MinValidatorsForKeyGen,
		"min_validators", minValidators,
		"active_validators", len(validatorAddresses),
		"block_height", sdkCtx.BlockHeight(),
	)

	return validatorAddresses
}

// IsConsensusInputUnique checks if consensus input has been used before
func (es *EncryptionSubkeeper) IsConsensusInputUnique(
	ctx context.Context,
	consensusInput []byte,
) (bool, error) {
	if len(consensusInput) == 0 {
		return false, fmt.Errorf("consensus input cannot be empty")
	}

	// Unwrap SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Hash consensus input using SHA256 for efficient storage
	hasher := sha256.New()
	hasher.Write(consensusInput)
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	// Query database to check if hash exists in used consensus inputs
	// Using SaltStoreTable as it's the most appropriate table for storing hash data
	saltTable := es.keeper.OrmDB.SaltStoreTable()

	// Check if consensus input hash already exists
	exists, err := saltTable.Has(ctx, hashHex)
	if err != nil {
		es.logger.Error("Failed to check consensus input uniqueness", "error", err, "hash", hashHex)
		return false, fmt.Errorf("failed to query consensus input uniqueness: %w", err)
	}

	// If hash already exists, consensus input is not unique
	if exists {
		es.logger.Debug("Consensus input already used", "hash", hashHex[:16]+"...")
		return false, nil
	}

	// Store new consensus input hash for future uniqueness checks
	currentTime := sdkCtx.BlockTime().Unix()

	// Get module parameters for key rotation cycles to calculate expiration
	params, err := es.keeper.Params.Get(ctx)
	if err != nil {
		es.logger.Error("Failed to get module parameters", "error", err)
		// Continue with default expiration if params unavailable
		params.KeyRotationDays = types.DefaultKeyRotationDays
	}

	// Create salt store entry to track consensus input usage
	// Using original consensus input as salt for verification if needed
	saltEntry := &apiv1.SaltStore{
		RecordId:  hashHex,
		SaltValue: consensusInput,
		CreatedAt: currentTime,
	}

	// Insert the consensus input hash into storage
	if err := saltTable.Insert(ctx, saltEntry); err != nil {
		es.logger.Error(
			"Failed to store consensus input hash",
			"error",
			err,
			"hash",
			hashHex[:16]+"...",
		)
		return false, fmt.Errorf("failed to store consensus input hash: %w", err)
	}

	// Perform cleanup of old entries to prevent storage bloat
	// Calculate expiration based on key rotation cycles
	// Convert days to seconds (assuming key rotation cycles align with time-based expiration)
	expirationSeconds := int64(params.KeyRotationDays) * 24 * 60 * 60
	expirationTime := currentTime - expirationSeconds

	// Use async cleanup to avoid blocking the main operation
	go func() {
		// Create background context for cleanup operation
		cleanupCtx := context.Background()

		// Query and delete old entries using created_at index
		createdAtIndex := apiv1.SaltStoreCreatedAtIndexKey{}.WithCreatedAt(expirationTime)

		// Use DeleteRange to clean up expired entries efficiently
		// Delete all entries from beginning of time to expiration time
		err := saltTable.DeleteRange(cleanupCtx,
			apiv1.SaltStoreCreatedAtIndexKey{}.WithCreatedAt(0), // From beginning
			createdAtIndex, // To expiration time
		)
		if err != nil {
			es.logger.Error("Failed to cleanup expired consensus inputs", "error", err)
			return
		}

		es.logger.Debug("Cleaned up expired consensus inputs", "expiration_time", expirationTime)
	}()

	es.logger.Debug("Consensus input is unique and stored", "hash", hashHex[:16]+"...")
	return true, nil
}

// GetEncryptionStats returns encryption statistics for monitoring using ORM data
func (es *EncryptionSubkeeper) GetEncryptionStats(
	ctx context.Context,
) (*types.EncryptionStats, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Count total encrypted records by iterating through all records
	totalEncryptedRecords := uint64(0)
	lastEncryptionHeight := int64(0)

	// Query all records and count those with encryption metadata
	recordIter, err := es.keeper.OrmDB.DWNRecordTable().List(ctx, apiv1.DWNRecordPrimaryKey{})
	if err != nil {
		es.logger.Error("Failed to list records for encryption stats", "error", err)
	} else {
		defer recordIter.Close()
		for recordIter.Next() {
			record, iterErr := recordIter.Value()
			if iterErr != nil {
				continue
			}

			// Check if record has encryption metadata
			if record.EncryptionMetadata != nil {
				totalEncryptedRecords++
				if record.EncryptionMetadata.EncryptionHeight > lastEncryptionHeight {
					lastEncryptionHeight = record.EncryptionMetadata.EncryptionHeight
				}
			}
		}
	}

	// Get key state statistics
	keyState, keyStateErr := es.getStoredKeyState(ctx)
	totalKeyRotations := uint64(0)
	activeValidators := uint64(0)
	singleNodeMode := true

	if keyStateErr == nil {
		if keyState.KeyVersion > 0 {
			totalKeyRotations = keyState.KeyVersion
		}
		activeValidators = uint64(len(keyState.ValidatorSet))
		singleNodeMode = keyState.SingleNodeMode
	}

	// Count VRF contributions for additional metrics
	totalVrfContributions := uint64(0)
	vrfIter, vrfErr := es.keeper.OrmDB.VRFContributionTable().
		List(ctx, apiv1.VRFContributionPrimaryKey{})
	if vrfErr == nil {
		defer vrfIter.Close()
		for vrfIter.Next() {
			_, iterErr := vrfIter.Value()
			if iterErr == nil {
				totalVrfContributions++
			}
		}
	}

	// Safely convert uint64 to int64 to avoid overflow
	var totalEncryptedRecordsInt64 int64
	if totalEncryptedRecords > 9223372036854775807 { // Max int64
		totalEncryptedRecordsInt64 = 9223372036854775807
	} else {
		totalEncryptedRecordsInt64 = int64(totalEncryptedRecords)
	}

	stats := &types.EncryptionStats{
		TotalEncryptedRecords: totalEncryptedRecordsInt64,
		TotalDecryptionErrors: 0, // This would need to be tracked separately in production
		LastEncryptionHeight:  lastEncryptionHeight,
	}

	es.logger.Debug("Collected encryption statistics",
		"encrypted_records", totalEncryptedRecords,
		"last_encryption_height", lastEncryptionHeight,
		"key_rotations", totalKeyRotations,
		"active_validators", activeValidators,
		"vrf_contributions", totalVrfContributions,
		"single_node_mode", singleNodeMode,
		"current_block_height", sdkCtx.BlockHeight(),
	)

	return stats, nil
}

// EncryptWithConsensusKey encrypts data using the current consensus-derived encryption key
func (es *EncryptionSubkeeper) EncryptWithConsensusKey(
	ctx context.Context,
	plaintext []byte,
	protocol string,
) (*EncryptedData, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	// Create consensus input based on protocol
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consensusInput := fmt.Appendf(nil, "protocol:%s:height:%d", protocol, sdkCtx.BlockHeight())

	// Use the unified encryption key derivation
	encryptionKey, err := es.getEncryptionKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Perform AES-256-GCM encryption
	return es.encryptWithKey(ctx, plaintext, encryptionKey, consensusInput, sdkCtx.BlockHeight())
}

// encryptWithKey performs the actual AES-256-GCM encryption
func (es *EncryptionSubkeeper) encryptWithKey(
	ctx context.Context,
	plaintext []byte,
	key []byte,
	consensusInput []byte,
	blockHeight int64,
) (*EncryptedData, error) {
	// Initialize AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, randErr := io.ReadFull(rand.Reader, nonce); randErr != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", randErr)
	}

	// Encrypt data with authenticated encryption
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Extract authentication tag (last 16 bytes of GCM ciphertext)
	if len(ciphertext) < 16 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	authTag := ciphertext[len(ciphertext)-16:]
	actualCiphertext := ciphertext[:len(ciphertext)-16]

	// Compute HMAC-SHA256 for data integrity verification
	hmacKey := es.deriveHMACKey(key, nonce)
	dataHMAC := es.computeHMAC(plaintext, hmacKey)

	// Create encryption metadata
	metadata := &types.EncryptionMetadata{
		Algorithm:        "AES-256-GCM",
		ConsensusInput:   consensusInput,
		Nonce:            nonce,
		AuthTag:          authTag,
		EncryptionHeight: blockHeight,
		ValidatorSet:     es.getValidatorSet(ctx),
		KeyVersion:       es.calculateKeyEpoch(blockHeight), // Approximate 24h epochs
		SingleNodeMode:   es.isSingleNodeMode(ctx),
		DataHmac:         dataHMAC,
	}

	es.logger.Info("Data encrypted with consensus key",
		"plaintext_len", len(plaintext),
		"ciphertext_len", len(actualCiphertext),
		"single_node_mode", es.isSingleNodeMode(ctx),
	)

	return &EncryptedData{
		Ciphertext: actualCiphertext,
		Metadata:   metadata,
	}, nil
}

// InitiateKeyRotation starts the key rotation process
func (es *EncryptionSubkeeper) InitiateKeyRotation(ctx context.Context, reason string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentEpoch := es.calculateKeyEpoch(sdkCtx.BlockHeight()) // Approximate 24h epochs

	// Check if we've already rotated for this epoch
	keyState, err := es.getStoredKeyState(ctx)
	if err == nil && keyState.KeyVersion == currentEpoch {
		// Already rotated for this epoch, skip
		return nil
	}

	es.logger.Info("Initiating key rotation",
		"reason", reason,
		"block_height", sdkCtx.BlockHeight(),
		"key_epoch", currentEpoch,
	)

	// 1. Generate new encryption key using unified approach
	newKey, err := es.getEncryptionKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate new encryption key: %w", err)
	}

	// 2. Create new key state
	keyState = &types.EncryptionKeyState{
		CurrentKey:     newKey,
		KeyVersion:     currentEpoch,
		ValidatorSet:   es.getValidatorSet(ctx),
		Contributions:  []*types.VRFContribution{}, // Initialize empty slice
		LastRotation:   sdkCtx.BlockTime().Unix(),
		NextRotation:   calculateNextRotation(sdkCtx.BlockTime().Unix()),
		SingleNodeMode: es.isSingleNodeMode(ctx),
	}

	// 3. Store the new key state (placeholder implementation for now)
	err = es.storeKeyState(ctx, keyState)
	if err != nil {
		// For now, log the error but continue since storage is not implemented
		es.logger.Warn("Key state storage failed (not yet implemented)",
			"error", err,
			"key_version", currentEpoch,
		)
	}

	// 4. Emit key rotation event
	var oldKeyVersion uint64
	if prevKeyState, err := es.getStoredKeyState(ctx); err == nil {
		oldKeyVersion = prevKeyState.KeyVersion
	} else {
		oldKeyVersion = 0 // First rotation
	}

	event := &types.EventKeyRotation{
		OldKeyVersion:  oldKeyVersion,
		NewKeyVersion:  currentEpoch,
		Reason:         reason,
		BlockHeight:    uint64(sdkCtx.BlockHeight()),
		SingleNodeMode: es.isSingleNodeMode(ctx),
		ValidatorCount: uint32(len(es.getValidatorSet(ctx))),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		es.logger.Error("Failed to emit key rotation event",
			"error", err,
			"key_version", currentEpoch,
		)
	}

	es.logger.Info("Key rotation completed successfully",
		"old_key_version", oldKeyVersion,
		"new_key_version", currentEpoch,
		"single_node_mode", es.isSingleNodeMode(ctx),
		"reason", reason,
	)

	return nil
}

// IsRotationDue checks if a scheduled key rotation is due
func (es *EncryptionSubkeeper) IsRotationDue(ctx context.Context) bool {
	keyState, err := es.getStoredKeyState(ctx)
	if err != nil {
		// No stored key state means initial rotation is due
		es.logger.Info("No stored key state found, rotation due for initialization")
		return true
	}

	currentTime := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()

	// Check if scheduled rotation time has passed
	if currentTime >= keyState.NextRotation {
		es.logger.Info("Scheduled rotation time reached",
			"current_time", currentTime,
			"next_rotation", keyState.NextRotation,
		)
		return true
	}

	// Check if validator set has changed significantly (>10%)
	validatorSetChanged := es.hasValidatorSetChanged(ctx, 0.1)
	if validatorSetChanged {
		es.logger.Info("Validator set changed significantly, rotation due")
		return true
	}

	return false
}

// calculateNextRotation calculates the next scheduled rotation time (30 days from current)
func calculateNextRotation(currentTime int64) int64 {
	// 30 days in seconds
	thirtyDays := int64(30 * 24 * 60 * 60)
	return currentTime + thirtyDays
}

// storeKeyState stores the encryption key state to persistent storage using ORM
func (es *EncryptionSubkeeper) storeKeyState(
	ctx context.Context,
	keyState *types.EncryptionKeyState,
) error {
	// Convert to API type for ORM storage
	apiKeyState := &apiv1.EncryptionKeyState{
		KeyVersion:     keyState.KeyVersion,
		CurrentKey:     keyState.CurrentKey,
		ValidatorSet:   keyState.ValidatorSet,
		LastRotation:   keyState.LastRotation,
		NextRotation:   keyState.NextRotation,
		SingleNodeMode: keyState.SingleNodeMode,
		Contributions:  make([]*apiv1.VRFContribution, len(keyState.Contributions)),
	}

	// Convert VRF contributions
	for i, contrib := range keyState.Contributions {
		apiKeyState.Contributions[i] = &apiv1.VRFContribution{
			ValidatorAddress: contrib.ValidatorAddress,
			Randomness:       contrib.Randomness,
			Proof:            contrib.Proof,
			BlockHeight:      contrib.BlockHeight,
			Timestamp:        contrib.Timestamp,
		}
	}

	// Use Save method to insert or update
	if err := es.keeper.OrmDB.EncryptionKeyStateTable().Save(ctx, apiKeyState); err != nil {
		return fmt.Errorf("failed to save encryption key state: %w", err)
	}

	es.logger.Info("Key state stored successfully",
		"key_version", keyState.KeyVersion,
		"validator_count", len(keyState.ValidatorSet),
		"last_rotation", keyState.LastRotation,
		"next_rotation", keyState.NextRotation,
		"single_node_mode", keyState.SingleNodeMode,
		"block_height", sdk.UnwrapSDKContext(ctx).BlockHeight(),
	)

	return nil
}

// getStoredKeyState retrieves the current encryption key state using ORM
func (es *EncryptionSubkeeper) getStoredKeyState(
	ctx context.Context,
) (*types.EncryptionKeyState, error) {
	// Get the current key version
	currentKeyVersion := es.GetCurrentKeyVersion(ctx)

	// Try to retrieve the stored key state for current version
	apiKeyState, err := es.keeper.OrmDB.EncryptionKeyStateTable().Get(ctx, currentKeyVersion)
	if err != nil {
		// If not found, try the previous version as fallback
		if currentKeyVersion > 0 {
			apiKeyState, err = es.keeper.OrmDB.EncryptionKeyStateTable().
				Get(ctx, currentKeyVersion-1)
			if err != nil {
				return nil, fmt.Errorf("no stored key state found for versions %d or %d: %w",
					currentKeyVersion, currentKeyVersion-1, err)
			}
		} else {
			return nil, fmt.Errorf("no stored key state found for version %d: %w", currentKeyVersion, err)
		}
	}

	// Convert from API type back to types
	keyState := &types.EncryptionKeyState{
		KeyVersion:     apiKeyState.KeyVersion,
		CurrentKey:     apiKeyState.CurrentKey,
		ValidatorSet:   apiKeyState.ValidatorSet,
		LastRotation:   apiKeyState.LastRotation,
		NextRotation:   apiKeyState.NextRotation,
		SingleNodeMode: apiKeyState.SingleNodeMode,
		Contributions:  make([]*types.VRFContribution, len(apiKeyState.Contributions)),
	}

	// Convert VRF contributions
	for i, contrib := range apiKeyState.Contributions {
		keyState.Contributions[i] = &types.VRFContribution{
			ValidatorAddress: contrib.ValidatorAddress,
			Randomness:       contrib.Randomness,
			Proof:            contrib.Proof,
			BlockHeight:      contrib.BlockHeight,
			Timestamp:        contrib.Timestamp,
		}
	}

	es.logger.Debug("Retrieved stored key state",
		"key_version", keyState.KeyVersion,
		"validator_count", len(keyState.ValidatorSet),
		"single_node_mode", keyState.SingleNodeMode,
		"contributions_count", len(keyState.Contributions),
	)

	return keyState, nil
}

// CheckAndPerformRotation checks if rotation is due and performs it if needed
func (es *EncryptionSubkeeper) CheckAndPerformRotation(ctx context.Context) error {
	// Check if encryption is enabled before attempting rotation
	params, err := es.keeper.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get DWN params: %w", err)
	}

	if !params.EncryptionEnabled {
		// Encryption disabled - skip rotation silently
		return nil
	}

	if !es.IsRotationDue(ctx) {
		return nil
	}

	// Determine rotation reason
	reason := "scheduled_rotation"

	// Check if due to validator set change
	validatorSetChanged := es.hasValidatorSetChanged(ctx, 0.1)
	if validatorSetChanged {
		reason = "validator_set_change"
	}

	// Perform the rotation
	return es.InitiateKeyRotation(ctx, reason)
}

// getEncryptionKey provides unified encryption key derivation with single-node fallback
func (es *EncryptionSubkeeper) getEncryptionKey(ctx context.Context) ([]byte, error) {
	validators, err := es.getActiveValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active validators: %w", err)
	}

	if len(validators) == 1 {
		// Single-node development mode - deterministic key generation
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		chainID := sdkCtx.ChainID()
		blockHeight := sdkCtx.BlockHeight()

		// 30-day epochs (assuming 6s blocks): 30 days * 24 hours * 60 minutes * 10 blocks/minute
		keyEpoch := blockHeight / (30 * 24 * 60 * 10)

		input := fmt.Sprintf("%s:%d", chainID, keyEpoch)

		es.logger.Info("Using single-node fallback for encryption key",
			"chain_id", chainID,
			"key_epoch", keyEpoch,
			"block_height", blockHeight,
		)

		return es.keeper.ComputeVRF([]byte(input))
	}

	// Multi-validator consensus mode
	es.logger.Info("Using multi-validator consensus for encryption key",
		"validator_count", len(validators),
	)

	return es.deriveConsensusKey(ctx, validators)
}

// deriveConsensusKey implements multi-validator consensus key derivation
func (es *EncryptionSubkeeper) deriveConsensusKey(
	ctx context.Context,
	validators []any,
) ([]byte, error) {
	// Create consensus input based on validators
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	keyEpoch := es.calculateKeyEpoch(sdkCtx.BlockHeight()) // 24h epochs

	consensusInput := fmt.Appendf(nil, "consensus:%s:%d:%d",
		sdkCtx.ChainID(), keyEpoch, len(validators))

	// Use VRF to derive consensus key
	return es.keeper.ComputeVRF(consensusInput)
}

// getActiveValidators returns the current active validators
func (es *EncryptionSubkeeper) getActiveValidators(ctx context.Context) ([]any, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validators, err := es.keeper.stakingKeeper.GetBondedValidatorsByPower(sdkCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonded validators: %w", err)
	}

	// Convert to interface slice
	result := make([]any, len(validators))
	for i, validator := range validators {
		result[i] = validator
	}

	return result, nil
}

// isSingleNodeMode checks if running in single-node mode
func (es *EncryptionSubkeeper) isSingleNodeMode(ctx context.Context) bool {
	validators, err := es.getActiveValidators(ctx)
	if err != nil {
		return true // Default to single-node if can't get validators
	}
	return len(validators) == 1
}

// hasValidatorSetChanged checks if validator set changed significantly (>threshold% change)
func (es *EncryptionSubkeeper) hasValidatorSetChanged(ctx context.Context, threshold float64) bool {
	// Get current validator set
	currentValidators, err := es.getActiveValidators(ctx)
	if err != nil {
		es.logger.Error("Failed to get current validators", "error", err)
		return false
	}

	// Get stored key state to compare with previous validator set
	keyState, err := es.getStoredKeyState(ctx)
	if err != nil {
		// No previous key state means this is the first key generation
		// Only trigger rotation if we have validators
		if len(currentValidators) > 0 {
			es.logger.Info("No previous validator set found, initial rotation needed")
			return true
		}
		// No validators and no previous state - no rotation needed
		return false
	}

	previousValidators := keyState.ValidatorSet

	// Build sets for comparison
	currentSet := make(map[string]bool)
	for _, validator := range currentValidators {
		currentSet[fmt.Sprintf("%v", validator)] = true
	}

	previousSet := make(map[string]bool)
	for _, validator := range previousValidators {
		previousSet[validator] = true
	}

	// Count added and removed validators
	added := 0
	for validator := range currentSet {
		if !previousSet[validator] {
			added++
		}
	}

	removed := 0
	for validator := range previousSet {
		if !currentSet[validator] {
			removed++
		}
	}

	totalChange := added + removed
	totalValidators := len(currentValidators)
	if len(previousValidators) > totalValidators {
		totalValidators = len(previousValidators)
	}

	// Avoid division by zero
	if totalValidators == 0 {
		return false
	}

	// If previous validators were empty and we now have validators, consider it a change
	if len(previousValidators) == 0 && len(currentValidators) > 0 {
		es.logger.Info("Validator set initialized",
			"current_validators", len(currentValidators),
			"previous_validators", 0,
		)
		return true
	}

	// If no actual changes in the validator addresses, no rotation needed
	if added == 0 && removed == 0 {
		return false
	}

	changePercentage := float64(totalChange) / float64(totalValidators)
	changed := changePercentage > threshold

	es.logger.Info("Validator set change analysis",
		"current_validators", len(currentValidators),
		"previous_validators", len(previousValidators),
		"added", added,
		"removed", removed,
		"change_percentage", changePercentage,
		"threshold", threshold,
		"changed", changed,
	)

	return changed
}

// calculateKeyEpoch safely converts block height to key epoch with overflow protection
func (es *EncryptionSubkeeper) calculateKeyEpoch(blockHeight int64) uint64 {
	if blockHeight < 0 {
		return 0
	}
	// 24h epochs assuming 6s blocks: 14400 blocks per day
	epoch := blockHeight / 14400
	if epoch < 0 {
		return 0
	}
	return uint64(epoch)
}

// GetCurrentKeyVersion returns the current key version/epoch
func (es *EncryptionSubkeeper) GetCurrentKeyVersion(ctx context.Context) uint64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return es.calculateKeyEpoch(sdkCtx.BlockHeight())
}

// DecryptWithConsensusKey decrypts data using consensus-derived encryption keys
func (es *EncryptionSubkeeper) DecryptWithConsensusKey(
	ctx context.Context,
	ciphertext []byte,
	metadata *types.EncryptionMetadata,
) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext cannot be empty")
	}

	if metadata == nil {
		return nil, fmt.Errorf("encryption metadata cannot be nil")
	}

	// Verify algorithm support
	if metadata.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", metadata.Algorithm)
	}

	// Get the encryption key for decryption
	// For now, use current key - in production, would use key versioning
	encryptionKey, err := es.getEncryptionKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Perform AES-256-GCM decryption
	plaintext, err := es.decryptWithKey(ciphertext, encryptionKey, metadata)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	es.logger.Debug("Data decrypted successfully",
		"ciphertext_len", len(ciphertext),
		"plaintext_len", len(plaintext),
		"algorithm", metadata.Algorithm,
		"key_version", metadata.KeyVersion,
		"block_height", sdkCtx.BlockHeight(),
	)

	return plaintext, nil
}

// decryptWithKey performs the actual AES-256-GCM decryption
func (es *EncryptionSubkeeper) decryptWithKey(
	ciphertext []byte,
	key []byte,
	metadata *types.EncryptionMetadata,
) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Validate nonce size
	if len(metadata.Nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf(
			"invalid nonce size: expected %d, got %d",
			gcm.NonceSize(),
			len(metadata.Nonce),
		)
	}

	// Combine ciphertext and auth tag for GCM decryption
	combinedData := make([]byte, len(ciphertext)+len(metadata.AuthTag))
	copy(combinedData, ciphertext)
	copy(combinedData[len(ciphertext):], metadata.AuthTag)

	// Decrypt using GCM
	plaintext, err := gcm.Open(nil, metadata.Nonce, combinedData, metadata.ConsensusInput)
	if err != nil {
		return nil, fmt.Errorf("GCM decryption failed: %w", err)
	}

	// Verify HMAC for additional data integrity checking
	if len(metadata.DataHmac) > 0 {
		hmacKey := es.deriveHMACKey(key, metadata.Nonce)
		if !es.verifyHMAC(plaintext, hmacKey, metadata.DataHmac) {
			return nil, fmt.Errorf("HMAC verification failed: data integrity compromised")
		}
	}

	return plaintext, nil
}

// deriveHMACKey derives an HMAC key from the encryption key and nonce
func (es *EncryptionSubkeeper) deriveHMACKey(encryptionKey, nonce []byte) []byte {
	// Use HKDF to derive HMAC key from encryption key and nonce
	hkdfReader := hkdf.New(sha256.New, encryptionKey, nonce, []byte("dwn-hmac-key-derivation"))

	// Generate 32 bytes for HMAC-SHA256 key
	hmacKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, hmacKey); err != nil {
		es.logger.Error("Failed to derive HMAC key", "error", err)
		return nil
	}

	return hmacKey
}

// computeHMAC computes HMAC-SHA256 of data using the provided key
func (es *EncryptionSubkeeper) computeHMAC(data, key []byte) []byte {
	if len(key) == 0 {
		es.logger.Error("HMAC key is empty")
		return nil
	}

	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// verifyHMAC verifies HMAC-SHA256 of data against expected HMAC
func (es *EncryptionSubkeeper) verifyHMAC(data, key, expectedHMAC []byte) bool {
	if len(key) == 0 || len(expectedHMAC) == 0 {
		es.logger.Error("HMAC key or expected HMAC is empty")
		return false
	}

	computedHMAC := es.computeHMAC(data, key)
	if len(computedHMAC) == 0 {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal(expectedHMAC, computedHMAC)
}
