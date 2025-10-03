package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/errors"
	v1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/sonr/x/svc/types"
)

type msgServer struct {
	k Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

// UCAN validation helper functions

// extractUCANToken extracts UCAN token from transaction context
func (ms msgServer) extractUCANToken(ctx context.Context) (string, bool) {
	// In production, UCAN tokens would be extracted from:
	// 1. Transaction metadata set by ante handlers
	// 2. Message extension fields
	// 3. Transaction memo field
	// For now, we return false to proceed with normal validation
	return "", false
}

// validateUCANPermission validates UCAN authorization for a Service operation
func (ms msgServer) validateUCANPermission(
	ctx context.Context,
	serviceID string,
	operation types.ServiceOperation,
) error {
	// Try to extract UCAN token
	tokenString, hasToken := ms.extractUCANToken(ctx)
	if !hasToken {
		// No UCAN token present, proceed with normal validation
		return nil
	}

	// Validate UCAN token for the specific operation
	validator := ms.k.GetPermissionValidator()
	if validator == nil {
		return fmt.Errorf("UCAN permission validator not initialized")
	}

	// Use general permission validation
	return validator.ValidatePermission(ctx, tokenString, serviceID, operation)
}

// validateDomainBoundUCANPermission validates UCAN authorization for domain-bound operations
func (ms msgServer) validateDomainBoundUCANPermission(
	ctx context.Context,
	domain string,
	serviceID string,
	operation types.ServiceOperation,
) error {
	// Try to extract UCAN token
	tokenString, hasToken := ms.extractUCANToken(ctx)
	if !hasToken {
		// No UCAN token present, proceed with normal validation
		return nil
	}

	// Validate domain-bound UCAN token
	validator := ms.k.GetPermissionValidator()
	if validator == nil {
		return fmt.Errorf("UCAN permission validator not initialized")
	}

	return validator.ValidateDomainBoundPermission(ctx, tokenString, domain, serviceID, operation)
}

// checkGaslessSupport checks if the operation can be executed gaslessly via UCAN
func (ms msgServer) checkGaslessSupport(
	ctx context.Context,
	serviceID string,
	operation types.ServiceOperation,
) (bool, uint64) {
	// Try to extract UCAN token
	tokenString, hasToken := ms.extractUCANToken(ctx)
	if !hasToken {
		return false, 0
	}

	// Check gasless support
	validator := ms.k.GetPermissionValidator()
	if validator == nil {
		return false, 0
	}

	supportsGasless, gasLimit, err := validator.SupportsGaslessTransaction(ctx, tokenString, serviceID, operation)
	if err != nil {
		ms.k.Logger().Debug("Failed to check gasless support", "error", err)
		return false, 0
	}

	return supportsGasless, gasLimit
}

func (ms msgServer) UpdateParams(
	ctx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			ms.k.authority,
			msg.Authority,
		)
	}

	return nil, ms.k.Params.Set(ctx, msg.Params)
}

// InitiateDomainVerification implements types.MsgServer.
func (ms msgServer) InitiateDomainVerification(
	ctx context.Context,
	msg *types.MsgInitiateDomainVerification,
) (*types.MsgInitiateDomainVerificationResponse, error) {
	// UCAN authorization validation for domain verification
	if err := ms.validateDomainBoundUCANPermission(ctx, msg.Domain, "", types.ServiceOpInitiateDomainVerification); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidUCANDelegation, "UCAN validation failed: %v", err)
	}

	// Check for gasless execution support
	supportsGasless, gasLimit := ms.checkGaslessSupport(ctx, msg.Domain, types.ServiceOpInitiateDomainVerification)
	if supportsGasless {
		// Log gasless execution (in production, this might set transaction fees to zero)
		ms.k.Logger().Info("Executing domain verification with gasless transaction",
			"domain", msg.Domain,
			"gas_limit", gasLimit)
	}

	// Initiate domain verification using the keeper
	verification, err := ms.k.InitiateDomainVerification(ctx, msg.Domain, msg.Creator)
	if err != nil {
		return nil, err
	}

	// Generate DNS instructions for the user
	dnsInstructions := ms.k.GetDNSInstructions(msg.Domain, verification.VerificationToken)

	// Emit typed event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventDomainVerificationInitiated{
		Domain:         msg.Domain,
		VerificationId: msg.Domain, // Using domain as ID since it's unique
		Challenge:      verification.VerificationToken,
		Initiator:      msg.Creator,
		BlockHeight:    uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventDomainVerificationInitiated")
	}

	return &types.MsgInitiateDomainVerificationResponse{
		VerificationToken: verification.VerificationToken,
		DnsInstruction:    dnsInstructions,
	}, nil
}

// VerifyDomain implements types.MsgServer.
func (ms msgServer) VerifyDomain(
	ctx context.Context,
	msg *types.MsgVerifyDomain,
) (*types.MsgVerifyDomainResponse, error) {
	// UCAN authorization validation for domain verification
	if err := ms.validateDomainBoundUCANPermission(ctx, msg.Domain, "", types.ServiceOpVerifyDomain); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidUCANDelegation, "UCAN validation failed: %v", err)
	}

	// Verify domain ownership by checking DNS TXT records
	verification, err := ms.k.VerifyDomainOwnership(ctx, msg.Domain)
	if err != nil {
		return &types.MsgVerifyDomainResponse{
			Verified: false,
			Message:  err.Error(),
		}, nil // Return error message in response, not as gRPC error
	}

	// Check verification result
	verified := verification.Status == v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	message := "Domain verification successful"
	if !verified {
		message = "Domain verification failed - DNS TXT record not found or incorrect"
	} else {
		// Emit typed event for successful verification
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		event := &types.EventDomainVerified{
			Domain:         msg.Domain,
			VerificationId: msg.Domain, // Using domain as ID
			Verifier:       msg.Creator,
			VerifiedAt:     sdkCtx.BlockTime(),
			BlockHeight:    uint64(sdkCtx.BlockHeight()),
		}

		if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
			ms.k.Logger().With("error", err).Error("Failed to emit EventDomainVerified")
		}
	}

	return &types.MsgVerifyDomainResponse{
		Verified: verified,
		Message:  message,
	}, nil
}

// RegisterService implements types.MsgServer.
func (ms msgServer) RegisterService(
	ctx context.Context,
	msg *types.MsgRegisterService,
) (*types.MsgRegisterServiceResponse, error) {
	// UCAN authorization validation for service registration
	if err := ms.validateDomainBoundUCANPermission(ctx, msg.Domain, msg.ServiceId, types.ServiceOpRegister); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidUCANDelegation, "UCAN validation failed: %v", err)
	}

	// Check for gasless execution support
	supportsGasless, gasLimit := ms.checkGaslessSupport(ctx, msg.ServiceId, types.ServiceOpRegister)
	if supportsGasless {
		// Log gasless execution (in production, this might set transaction fees to zero)
		ms.k.Logger().Info("Executing service registration with gasless transaction",
			"service_id", msg.ServiceId,
			"domain", msg.Domain,
			"gas_limit", gasLimit)
	}

	// 1. Verify domain ownership
	if !ms.k.IsVerifiedDomain(ctx, msg.Domain) {
		return nil, errors.Wrapf(
			types.ErrDomainNotVerified,
			"domain %s is not verified",
			msg.Domain,
		)
	}

	// 2. Validate service owner DID
	if err := ms.k.ValidateServiceOwnerDID(ctx, msg.Creator); err != nil {
		return nil, errors.Wrapf(types.ErrInvalidOwnerDID, "owner DID validation failed: %v", err)
	}

	// 3. Validate service ID format
	if msg.ServiceId == "" {
		return nil, errors.Wrap(types.ErrInvalidServiceID, "service ID cannot be empty")
	}

	// 4. Check if service ID already exists
	existing, err := ms.k.OrmDB.ServiceTable().Get(ctx, msg.ServiceId)
	if err == nil && existing != nil {
		return nil, errors.Wrapf(
			types.ErrServiceAlreadyExists,
			"service with ID %s already exists",
			msg.ServiceId,
		)
	}

	// 5. Check if domain is already bound to another service
	existingByDomain, err := ms.k.OrmDB.ServiceTable().GetByDomain(ctx, msg.Domain)
	if err == nil && existingByDomain != nil {
		return nil, errors.Wrapf(
			types.ErrDomainAlreadyBound,
			"domain %s is already bound to service %s",
			msg.Domain,
			existingByDomain.Id,
		)
	}

	// 6. Validate requested permissions
	if err := ms.k.ValidateServicePermissions(ctx, msg.RequestedPermissions); err != nil {
		return nil, errors.Wrap(types.ErrInvalidPermissions, err.Error())
	}

	// 7. Validate UCAN delegation chain if provided
	if msg.UcanDelegationChain != "" {
		err := ms.k.ValidateUCANDelegationChain(
			ctx,
			msg.UcanDelegationChain,
		)
		if err != nil {
			return nil, errors.Wrap(types.ErrInvalidUCANDelegation, err.Error())
		}

		// Additionally validate the UCAN token grants the required permissions for the domain
		resource := fmt.Sprintf("service://%s", msg.Domain)
		_, err = ms.k.ValidateUCANToken(
			ctx,
			msg.UcanDelegationChain,
			resource,
			msg.RequestedPermissions,
		)
		if err != nil {
			return nil, errors.Wrap(
				types.ErrInvalidUCANDelegation,
				fmt.Sprintf("UCAN token validation failed: %v", err),
			)
		}
	}

	// 8. Create root capability for the service
	rootCapabilityCID, err := ms.k.CreateServiceRootCapability(ctx, msg)
	if err != nil {
		return nil, errors.Wrap(types.ErrFailedToCreateCapability, err.Error())
	}

	// 9. Create and save the service
	now := time.Now().Unix()
	service := &v1.Service{
		Id:                msg.ServiceId,
		Domain:            msg.Domain,
		Owner:             msg.Creator,
		RootCapabilityCid: rootCapabilityCID,
		Permissions:       msg.RequestedPermissions,
		Status:            v1.ServiceStatus_SERVICE_STATUS_ACTIVE,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = ms.k.OrmDB.ServiceTable().Insert(ctx, service)
	if err != nil {
		return nil, errors.Wrap(types.ErrFailedToSaveService, err.Error())
	}

	// Emit typed event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventServiceRegistered{
		ServiceId:   msg.ServiceId,
		Domain:      msg.Domain,
		Owner:       msg.Creator,
		Endpoints:   []string{}, // Can be populated if endpoints are provided
		Metadata:    "",         // Can be populated with service metadata if needed
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventServiceRegistered")
	}

	return &types.MsgRegisterServiceResponse{
		RootCapabilityCid: rootCapabilityCID,
		ServiceId:         msg.ServiceId,
	}, nil
}
