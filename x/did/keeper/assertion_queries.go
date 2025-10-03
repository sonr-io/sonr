package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/x/did/types"
)

// GetAssertionByControllerAndSubject retrieves an assertion by controller and subject
// This uses the unique index for optimal performance
func (k Keeper) GetAssertionByControllerAndSubject(
	ctx context.Context,
	controller string,
	subject string,
) (*apiv1.Assertion, error) {
	// Use the unique index on (controller, subject)
	return k.OrmDB.AssertionTable().GetByControllerSubject(ctx, controller, subject)
}

// GetAssertionsByController retrieves all assertions for a controller
func (k Keeper) GetAssertionsByController(
	ctx context.Context,
	controller string,
) ([]*apiv1.Assertion, error) {
	var assertions []*apiv1.Assertion

	// Use the index on controller
	indexKey := apiv1.AssertionControllerSubjectIndexKey{}.WithController(controller)
	iter, err := k.OrmDB.AssertionTable().List(ctx, indexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query assertions: %w", err)
	}
	defer iter.Close()

	for iter.Next() {
		assertion, err := iter.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get assertion value: %w", err)
		}
		assertions = append(assertions, assertion)
	}

	return assertions, nil
}

// HasAssertion checks if an assertion exists for a given DID
func (k Keeper) HasAssertion(ctx context.Context, did string) bool {
	_, err := k.OrmDB.AssertionTable().Get(ctx, did)
	return err == nil
}

// ValidateAssertionUniqueness validates that a controller+subject combination is unique
func (k Keeper) ValidateAssertionUniqueness(
	ctx context.Context,
	controller string,
	subject string,
) error {
	existing, err := k.GetAssertionByControllerAndSubject(ctx, controller, subject)
	if err == nil && existing != nil {
		return fmt.Errorf("assertion already exists for controller=%s, subject=%s", controller, subject)
	}

	return nil
}

// CreateAssertion creates a new assertion with uniqueness validation
func (k Keeper) CreateAssertion(
	ctx context.Context,
	did string,
	controller string,
	subject string,
	publicKeyBase64 string,
	didKind string,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate uniqueness
	if err := k.ValidateAssertionUniqueness(ctx, controller, subject); err != nil {
		return err
	}

	// Create assertion
	assertion := &apiv1.Assertion{
		Did:             did,
		Controller:      controller,
		Subject:         subject,
		PublicKeyBase64: publicKeyBase64,
		DidKind:         didKind,
		CreationBlock:   sdkCtx.BlockHeight(),
	}

	// Insert into ORM
	if err := k.OrmDB.AssertionTable().Insert(ctx, assertion); err != nil {
		return fmt.Errorf("failed to store assertion: %w", err)
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"assertion_created",
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("controller", controller),
			sdk.NewAttribute("subject", subject),
			sdk.NewAttribute("kind", didKind),
		),
	)

	return nil
}

// UpdateAssertion updates an existing assertion
func (k Keeper) UpdateAssertion(
	ctx context.Context,
	did string,
	publicKeyBase64 string,
) error {
	// Get existing assertion
	existing, err := k.OrmDB.AssertionTable().Get(ctx, did)
	if err != nil {
		return fmt.Errorf("assertion not found: %s", did)
	}

	// Update fields
	existing.PublicKeyBase64 = publicKeyBase64

	// Update in ORM
	if err := k.OrmDB.AssertionTable().Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update assertion: %w", err)
	}

	return nil
}

// DeleteAssertion removes an assertion
func (k Keeper) DeleteAssertion(ctx context.Context, did string) error {
	// Check if assertion exists
	existing, err := k.OrmDB.AssertionTable().Get(ctx, did)
	if err != nil {
		return fmt.Errorf("assertion not found: %s", did)
	}

	// Delete from ORM
	if err := k.OrmDB.AssertionTable().Delete(ctx, existing); err != nil {
		return fmt.Errorf("failed to delete assertion: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"assertion_deleted",
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("controller", existing.Controller),
			sdk.NewAttribute("subject", existing.Subject),
		),
	)

	return nil
}

// GetAssertionStats returns statistics about assertions
func (k Keeper) GetAssertionStats(ctx context.Context) (*types.AssertionStats, error) {
	stats := &types.AssertionStats{
		TotalAssertions:    0,
		EmailAssertions:    0,
		TelAssertions:      0,
		SonrAssertions:     0,
		WebAuthnAssertions: 0,
		OtherAssertions:    0,
	}

	// Iterate through all assertions
	iter, err := k.OrmDB.AssertionTable().List(ctx, apiv1.AssertionPrimaryKey{})
	if err != nil {
		return nil, fmt.Errorf("failed to list assertions: %w", err)
	}
	defer iter.Close()

	for iter.Next() {
		assertion, err := iter.Value()
		if err != nil {
			continue
		}

		stats.TotalAssertions++

		// Categorize by kind
		switch assertion.DidKind {
		case "email":
			stats.EmailAssertions++
		case "tel":
			stats.TelAssertions++
		case "sonr":
			stats.SonrAssertions++
		case "webauthn":
			stats.WebAuthnAssertions++
		default:
			stats.OtherAssertions++
		}
	}

	return stats, nil
}
