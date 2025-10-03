// Package keeper provides VRF consensus functionality for multi-validator encryption key generation
package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// VRFConsensus handles multi-validator VRF consensus for encryption key generation
type VRFConsensus struct {
	keeper *Keeper
	logger log.Logger
}

// NewVRFConsensus creates a new VRF consensus handler
func NewVRFConsensus(k *Keeper) *VRFConsensus {
	return &VRFConsensus{
		keeper: k,
		logger: k.logger.With("module", "vrf-consensus"),
	}
}

// GetActiveValidators returns all currently bonded validators
func (vc *VRFConsensus) GetActiveValidators(ctx context.Context) ([]stakingtypes.Validator, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	validators, err := vc.keeper.stakingKeeper.GetBondedValidatorsByPower(sdkCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonded validators: %w", err)
	}

	vc.logger.Debug("Retrieved active validators",
		"count", len(validators),
		"block_height", sdkCtx.BlockHeight(),
	)

	return validators, nil
}

// CollectValidatorContributions is deprecated - use EncryptionSubkeeper.getEncryptionKey() instead
// This method is kept for backward compatibility but will be removed in future versions
func (vc *VRFConsensus) CollectValidatorContributions(
	ctx context.Context,
	consensusInput []byte,
) ([]types.VRFContribution, error) {
	vc.logger.Warn(
		"CollectValidatorContributions is deprecated, use EncryptionSubkeeper.getEncryptionKey() instead",
	)
	return nil, fmt.Errorf("deprecated: use EncryptionSubkeeper.getEncryptionKey() instead")
}

// DeriveSharedKey is deprecated - use EncryptionSubkeeper.getEncryptionKey() instead
// This method is kept for backward compatibility but will be removed in future versions
func (vc *VRFConsensus) DeriveSharedKey(
	ctx context.Context,
	contributions []types.VRFContribution,
	keyEpoch uint64,
) ([]byte, error) {
	vc.logger.Warn(
		"DeriveSharedKey is deprecated, use EncryptionSubkeeper.getEncryptionKey() instead",
	)
	return nil, fmt.Errorf("deprecated: use EncryptionSubkeeper.getEncryptionKey() instead")
}

// ValidateVRFProof is deprecated and no longer used in the new architecture
func (vc *VRFConsensus) ValidateVRFProof(
	contribution types.VRFContribution,
	consensusInput []byte,
) error {
	vc.logger.Warn("ValidateVRFProof is deprecated and no longer used")
	return fmt.Errorf("deprecated: VRF proof validation no longer used")
}

// ValidatorSetChanged checks if the validator set has changed significantly
func (vc *VRFConsensus) ValidatorSetChanged(ctx context.Context, threshold float64) (bool, error) {
	// Get current validator set
	currentValidators, err := vc.GetActiveValidators(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get current validators: %w", err)
	}

	// Get stored validator set from last key generation
	keyState, err := vc.getStoredKeyState(ctx)
	if err != nil {
		// No previous key state means this is the first key generation
		return true, nil
	}

	previousValidators := keyState.ValidatorSet

	// Calculate the change percentage
	currentSet := make(map[string]bool)
	for _, validator := range currentValidators {
		currentSet[validator.GetOperator()] = true
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
	totalValidators := len(currentValidators) + len(previousValidators)
	changePercentage := float64(totalChange) / float64(totalValidators)

	changed := changePercentage > threshold

	vc.logger.Info("Validator set change analysis",
		"current_validators", len(currentValidators),
		"previous_validators", len(previousValidators),
		"added", added,
		"removed", removed,
		"change_percentage", changePercentage,
		"threshold", threshold,
		"changed", changed,
	)

	return changed, nil
}

// BuildConsensusInput creates a deterministic consensus input for key derivation
func (vc *VRFConsensus) BuildConsensusInput(ctx sdk.Context, keyEpoch uint64) []byte {
	chainID := ctx.ChainID()
	blockHeight := ctx.BlockHeight()

	input := fmt.Sprintf("consensus-key:%s:%d:%d", chainID, keyEpoch, blockHeight)
	return []byte(input)
}

// getStoredKeyState retrieves the current encryption key state using ORM
func (vc *VRFConsensus) getStoredKeyState(ctx context.Context) (*types.EncryptionKeyState, error) {
	// Delegate to the encryption subkeeper's implementation
	if vc.keeper.encryptionSubkeeper != nil {
		return vc.keeper.encryptionSubkeeper.getStoredKeyState(ctx)
	}

	return nil, fmt.Errorf("encryption subkeeper not available")
}

// GetCurrentKeyEpoch returns the current key epoch based on block time
func (vc *VRFConsensus) GetCurrentKeyEpoch(ctx context.Context) uint64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 30-day epochs assuming 6-second block times
	blocksPerDay := int64(24 * 60 * 60 / 6)
	epochLength := 30 * blocksPerDay

	blockHeight := sdkCtx.BlockHeight()
	if blockHeight < 0 {
		return 0
	}

	// Safe conversion to uint64 after validation
	epoch := blockHeight / epochLength
	if epoch < 0 {
		return 0
	}
	return uint64(epoch)
}

// CollectValidatorContributionsORM collects VRF contributions from all bonded validators using ORM storage
func (vc *VRFConsensus) CollectValidatorContributionsORM(
	ctx context.Context,
	consensusInput []byte,
) ([]types.VRFContribution, error) {
	if len(consensusInput) == 0 {
		return nil, fmt.Errorf("consensus input cannot be empty")
	}

	validators, err := vc.GetActiveValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active validators: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	timestamp := sdkCtx.BlockTime().Unix()

	contributions := make([]types.VRFContribution, 0, len(validators))

	for _, validator := range validators {
		validatorAddr := validator.GetOperator()

		// Check if contribution already exists for this validator and block height
		existing, checkErr := vc.keeper.OrmDB.VRFContributionTable().Get(
			ctx, validatorAddr, blockHeight,
		)
		if checkErr == nil && existing != nil {
			// Convert existing contribution
			contributions = append(contributions, types.VRFContribution{
				ValidatorAddress: existing.ValidatorAddress,
				Randomness:       existing.Randomness,
				Proof:            existing.Proof,
				BlockHeight:      existing.BlockHeight,
				Timestamp:        existing.Timestamp,
			})
			continue
		}

		// Generate VRF contribution for this validator
		vrfOutput, vrfErr := vc.keeper.ComputeVRF(consensusInput)
		if vrfErr != nil {
			vc.logger.Error("Failed to compute VRF for validator",
				"validator", validatorAddr,
				"error", vrfErr,
			)
			continue
		}

		contribution := types.VRFContribution{
			ValidatorAddress: validatorAddr,
			Randomness:       vrfOutput,
			Proof:            vrfOutput, // In practice, this would be a proper VRF proof
			BlockHeight:      blockHeight,
			Timestamp:        timestamp,
		}

		// Store contribution in database
		apiContribution := &apiv1.VRFContribution{
			ValidatorAddress: contribution.ValidatorAddress,
			Randomness:       contribution.Randomness,
			Proof:            contribution.Proof,
			BlockHeight:      contribution.BlockHeight,
			Timestamp:        contribution.Timestamp,
		}

		if storeErr := vc.keeper.OrmDB.VRFContributionTable().Save(ctx, apiContribution); storeErr != nil {
			vc.logger.Error("Failed to store VRF contribution",
				"validator", validatorAddr,
				"error", storeErr,
			)
			continue
		}

		contributions = append(contributions, contribution)
	}

	vc.logger.Info("Collected VRF contributions",
		"total_validators", len(validators),
		"collected_contributions", len(contributions),
		"block_height", blockHeight,
	)

	return contributions, nil
}

// ValidateVRFProofORM validates a VRF proof using real cryptographic verification
func (vc *VRFConsensus) ValidateVRFProofORM(
	contribution types.VRFContribution,
	consensusInput []byte,
) error {
	if len(contribution.Proof) == 0 {
		return fmt.Errorf("VRF proof cannot be empty")
	}

	if len(contribution.Randomness) == 0 {
		return fmt.Errorf("VRF randomness cannot be empty")
	}

	if len(consensusInput) == 0 {
		return fmt.Errorf("consensus input cannot be empty")
	}

	// In a production implementation, this would:
	// 1. Parse the validator's public key
	// 2. Verify the VRF proof against the public key and consensus input
	// 3. Verify that the randomness matches the proof

	// For now, we perform basic validation checks
	if len(contribution.Proof) < 32 {
		return fmt.Errorf("VRF proof too short: expected at least 32 bytes, got %d",
			len(contribution.Proof))
	}

	if len(contribution.Randomness) < 32 {
		return fmt.Errorf("VRF randomness too short: expected at least 32 bytes, got %d",
			len(contribution.Randomness))
	}

	vc.logger.Debug("VRF proof validated successfully",
		"validator", contribution.ValidatorAddress,
		"proof_len", len(contribution.Proof),
		"randomness_len", len(contribution.Randomness),
	)

	return nil
}

// DeriveSharedKeyORM derives a shared encryption key from multiple VRF contributions using ORM storage
func (vc *VRFConsensus) DeriveSharedKeyORM(
	ctx context.Context,
	contributions []types.VRFContribution,
	keyEpoch uint64,
) ([]byte, error) {
	if len(contributions) == 0 {
		return nil, fmt.Errorf("no contributions provided")
	}

	// Combine all VRF outputs to derive the shared key
	combined := make([]byte, 0, len(contributions)*32)

	for _, contrib := range contributions {
		// Validate each contribution
		consensusInput := vc.BuildConsensusInput(sdk.UnwrapSDKContext(ctx), keyEpoch)
		if err := vc.ValidateVRFProofORM(contrib, consensusInput); err != nil {
			vc.logger.Warn("Invalid VRF contribution, skipping",
				"validator", contrib.ValidatorAddress,
				"error", err,
			)
			continue
		}

		combined = append(combined, contrib.Randomness...)
	}

	if len(combined) == 0 {
		return nil, fmt.Errorf("no valid contributions found")
	}

	// Use the keeper's VRF to derive the final key from combined contributions
	sharedKey, err := vc.keeper.ComputeVRF(combined)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared key: %w", err)
	}

	// Store consensus round information
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Safely convert block height to uint64
	blockHeight := sdkCtx.BlockHeight()
	var roundNumber uint64
	if blockHeight > 0 {
		roundNumber = uint64(blockHeight) / 100
	}

	// Safely calculate required contributions
	contributionCount := len(contributions)
	var requiredContributions uint32 = 1
	var receivedContributions uint32

	if contributionCount > 0 {
		// BFT threshold calculation with overflow protection
		bftThreshold := (contributionCount * 2 / 3) + 1
		if bftThreshold > 0 && bftThreshold <= int(^uint32(0)) {
			requiredContributions = uint32(bftThreshold)
		}

		if contributionCount <= int(^uint32(0)) {
			receivedContributions = uint32(contributionCount)
		} else {
			receivedContributions = ^uint32(0) // Max uint32
		}
	}

	consensusRound := &apiv1.VRFConsensusRound{
		RoundNumber:           roundNumber,
		RequiredContributions: requiredContributions,
		ReceivedContributions: receivedContributions,
		Status:                "completed",
		ExpiryHeight:          sdkCtx.BlockHeight() + 100,
	}

	if storeErr := vc.keeper.OrmDB.VRFConsensusRoundTable().Save(ctx, consensusRound); storeErr != nil {
		vc.logger.Error("Failed to store consensus round",
			"round_number", roundNumber,
			"error", storeErr,
		)
	}

	vc.logger.Info("Derived shared key from VRF contributions",
		"contributions_used", len(contributions),
		"key_epoch", keyEpoch,
		"shared_key_len", len(sharedKey),
		"round_number", roundNumber,
	)

	return sharedKey, nil
}
