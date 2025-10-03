package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// KeyRotationPolicy defines the policy for automated key rotation
type KeyRotationPolicy struct {
	// Time-based rotation
	RotationInterval time.Duration // How often to rotate keys
	NextRotationTime time.Time     // When the next rotation is scheduled

	// Usage-based rotation
	MaxUsageCount     uint64 // Maximum number of operations before rotation
	CurrentUsageCount uint64 // Current operation count

	// Event-based rotation triggers
	RotateOnCompromise      bool // Rotate immediately if compromise detected
	RotateOnValidatorChange bool // Rotate when validator set changes significantly

	// Configuration
	Enabled     bool          // Whether automated rotation is enabled
	GracePeriod time.Duration // Grace period before forcing rotation
}

// KeyRotationScheduler handles automated key rotation scheduling
type KeyRotationScheduler struct {
	keeper *Keeper
	logger log.Logger
	policy *KeyRotationPolicy
}

// NewKeyRotationScheduler creates a new key rotation scheduler
func NewKeyRotationScheduler(keeper *Keeper) *KeyRotationScheduler {
	return &KeyRotationScheduler{
		keeper: keeper,
		logger: keeper.Logger().With("module", "key-rotation-scheduler"),
		policy: &KeyRotationPolicy{
			RotationInterval:        30 * 24 * time.Hour, // 30 days default
			MaxUsageCount:           1000000,             // 1 million operations default
			RotateOnCompromise:      true,
			RotateOnValidatorChange: true,
			Enabled:                 true,
			GracePeriod:             24 * time.Hour,
		},
	}
}

// CheckRotationPolicy checks all rotation policies and triggers rotation if needed
func (krs *KeyRotationScheduler) CheckRotationPolicy(ctx context.Context) error {
	if !krs.policy.Enabled {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get current key state
	keyState, err := krs.keeper.encryptionSubkeeper.getStoredKeyState(ctx)
	if err != nil {
		// No key state means we need initial rotation
		krs.logger.Info("No key state found, triggering initial rotation")
		return krs.TriggerRotation(ctx, "initial_setup")
	}

	// Check time-based rotation
	if krs.shouldRotateByTime(keyState) {
		krs.logger.Info("Time-based rotation triggered",
			"last_rotation", keyState.LastRotation,
			"interval", krs.policy.RotationInterval,
		)
		return krs.TriggerRotation(ctx, "scheduled_time_based")
	}

	// Check usage-based rotation
	if krs.shouldRotateByUsage(keyState) {
		krs.logger.Info("Usage-based rotation triggered",
			"usage_count", keyState.UsageCount,
			"max_usage", keyState.MaxUsageCount,
		)
		return krs.TriggerRotation(ctx, "usage_limit_reached")
	}

	// Check validator set changes
	if krs.policy.RotateOnValidatorChange {
		// Use hasValidatorSetChanged with a threshold (0.33 = 33% change)
		if krs.keeper.encryptionSubkeeper.hasValidatorSetChanged(sdkCtx, 0.33) {
			krs.logger.Info("Validator set change triggered rotation")
			return krs.TriggerRotation(ctx, "validator_set_changed")
		}
	}

	return nil
}

// shouldRotateByTime checks if time-based rotation is due
func (krs *KeyRotationScheduler) shouldRotateByTime(keyState *types.EncryptionKeyState) bool {
	if keyState.RotationInterval <= 0 {
		// Use default interval if not set
		keyState.RotationInterval = int64(krs.policy.RotationInterval.Seconds())
	}

	lastRotation := time.Unix(keyState.LastRotation, 0)
	nextRotation := lastRotation.Add(time.Duration(keyState.RotationInterval) * time.Second)

	return time.Now().After(nextRotation)
}

// shouldRotateByUsage checks if usage-based rotation is due
func (krs *KeyRotationScheduler) shouldRotateByUsage(keyState *types.EncryptionKeyState) bool {
	maxUsage := keyState.MaxUsageCount
	if maxUsage == 0 {
		maxUsage = krs.policy.MaxUsageCount
	}

	return keyState.UsageCount >= maxUsage
}

// TriggerRotation triggers an immediate key rotation
func (krs *KeyRotationScheduler) TriggerRotation(ctx context.Context, reason string) error {
	krs.logger.Info("Triggering key rotation", "reason", reason)

	// Perform the rotation through the encryption subkeeper
	if err := krs.keeper.encryptionSubkeeper.InitiateKeyRotation(ctx, reason); err != nil {
		return fmt.Errorf("failed to initiate key rotation: %w", err)
	}

	// Update usage counter
	if err := krs.resetUsageCounter(ctx); err != nil {
		krs.logger.Error("Failed to reset usage counter after rotation", "error", err)
	}

	// Emit rotation event
	krs.emitRotationEvent(ctx, reason)

	return nil
}

// IncrementUsageCount increments the usage counter for the current key
func (krs *KeyRotationScheduler) IncrementUsageCount(ctx context.Context) error {
	keyState, err := krs.keeper.encryptionSubkeeper.getStoredKeyState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current key state: %w", err)
	}

	// Increment usage count
	keyState.UsageCount++

	// Convert to API type for ORM storage
	apiKeyState := &apiv1.EncryptionKeyState{
		KeyVersion:         keyState.KeyVersion,
		CurrentKey:         keyState.CurrentKey,
		ValidatorSet:       keyState.ValidatorSet,
		LastRotation:       keyState.LastRotation,
		NextRotation:       keyState.NextRotation,
		SingleNodeMode:     keyState.SingleNodeMode,
		UsageCount:         keyState.UsageCount,
		MaxUsageCount:      keyState.MaxUsageCount,
		RotationInterval:   keyState.RotationInterval,
		CreatedAt:          keyState.CreatedAt,
		PreviousKeyVersion: keyState.PreviousKeyVersion,
	}

	// Update the key state
	if err := krs.keeper.OrmDB.EncryptionKeyStateTable().Update(ctx, apiKeyState); err != nil {
		return fmt.Errorf("failed to update key state: %w", err)
	}

	// Check if rotation is needed after increment
	if krs.shouldRotateByUsage(keyState) {
		go func() {
			// Async rotation to not block the current operation
			if err := krs.TriggerRotation(ctx, "usage_limit_reached"); err != nil {
				krs.logger.Error("Failed to trigger usage-based rotation", "error", err)
			}
		}()
	}

	return nil
}

// resetUsageCounter resets the usage counter after rotation
func (krs *KeyRotationScheduler) resetUsageCounter(ctx context.Context) error {
	keyState, err := krs.keeper.encryptionSubkeeper.getStoredKeyState(ctx)
	if err != nil {
		return err
	}

	keyState.UsageCount = 0

	// Convert to API type for ORM storage
	apiKeyState := &apiv1.EncryptionKeyState{
		KeyVersion:         keyState.KeyVersion,
		CurrentKey:         keyState.CurrentKey,
		ValidatorSet:       keyState.ValidatorSet,
		LastRotation:       keyState.LastRotation,
		NextRotation:       keyState.NextRotation,
		SingleNodeMode:     keyState.SingleNodeMode,
		UsageCount:         keyState.UsageCount,
		MaxUsageCount:      keyState.MaxUsageCount,
		RotationInterval:   keyState.RotationInterval,
		CreatedAt:          keyState.CreatedAt,
		PreviousKeyVersion: keyState.PreviousKeyVersion,
	}

	return krs.keeper.OrmDB.EncryptionKeyStateTable().Update(ctx, apiKeyState)
}

// emitRotationEvent emits a key rotation event
func (krs *KeyRotationScheduler) emitRotationEvent(ctx context.Context, reason string) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"key_rotation",
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("timestamp", fmt.Sprintf("%d", time.Now().Unix())),
		),
	)
}

// SetRotationPolicy updates the rotation policy
func (krs *KeyRotationScheduler) SetRotationPolicy(policy *KeyRotationPolicy) {
	krs.policy = policy
	krs.logger.Info("Updated key rotation policy",
		"interval", policy.RotationInterval,
		"max_usage", policy.MaxUsageCount,
		"enabled", policy.Enabled,
	)
}

// GetRotationPolicy returns the current rotation policy
func (krs *KeyRotationScheduler) GetRotationPolicy() *KeyRotationPolicy {
	return krs.policy
}

// ScheduleNextRotation schedules the next rotation based on the policy
func (krs *KeyRotationScheduler) ScheduleNextRotation(ctx context.Context) error {
	keyState, err := krs.keeper.encryptionSubkeeper.getStoredKeyState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current key state: %w", err)
	}

	// Calculate next rotation time
	nextRotation := time.Now().Add(krs.policy.RotationInterval)
	keyState.NextRotation = nextRotation.Unix()

	// Convert to API type for ORM storage
	apiKeyState := &apiv1.EncryptionKeyState{
		KeyVersion:         keyState.KeyVersion,
		CurrentKey:         keyState.CurrentKey,
		ValidatorSet:       keyState.ValidatorSet,
		LastRotation:       keyState.LastRotation,
		NextRotation:       keyState.NextRotation,
		SingleNodeMode:     keyState.SingleNodeMode,
		UsageCount:         keyState.UsageCount,
		MaxUsageCount:      keyState.MaxUsageCount,
		RotationInterval:   keyState.RotationInterval,
		CreatedAt:          keyState.CreatedAt,
		PreviousKeyVersion: keyState.PreviousKeyVersion,
	}

	// Update the key state
	if err := krs.keeper.OrmDB.EncryptionKeyStateTable().Update(ctx, apiKeyState); err != nil {
		return fmt.Errorf("failed to update next rotation time: %w", err)
	}

	krs.logger.Info("Scheduled next rotation",
		"next_rotation", nextRotation,
		"interval", krs.policy.RotationInterval,
	)

	return nil
}

// HandleCompromiseEvent handles a potential key compromise event
func (krs *KeyRotationScheduler) HandleCompromiseEvent(ctx context.Context, details string) error {
	if !krs.policy.RotateOnCompromise {
		krs.logger.Warn(
			"Key compromise detected but rotation on compromise is disabled",
			"details",
			details,
		)
		return nil
	}

	krs.logger.Error("Key compromise detected, triggering emergency rotation", "details", details)
	return krs.TriggerRotation(ctx, fmt.Sprintf("compromise_detected: %s", details))
}

// BeginBlock runs rotation checks at the beginning of each block
func (krs *KeyRotationScheduler) BeginBlock(ctx context.Context) error {
	// Check rotation policy every block
	if err := krs.CheckRotationPolicy(ctx); err != nil {
		krs.logger.Error("Failed to check rotation policy", "error", err)
		// Don't fail the block, just log the error
		return nil
	}

	return nil
}

// EndBlock runs cleanup at the end of each block
func (krs *KeyRotationScheduler) EndBlock(ctx context.Context) error {
	// Any end-of-block cleanup can go here
	return nil
}
