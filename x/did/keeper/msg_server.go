package keeper

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/errors"
	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/x/did/types"
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

// validateUCANPermission validates UCAN authorization for a DID operation
func (ms msgServer) validateUCANPermission(
	ctx context.Context,
	did string,
	controller string,
	operation types.DIDOperation,
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

	// Use controller-specific validation if controller is provided
	if controller != "" {
		return validator.ValidateControllerPermission(ctx, tokenString, did, controller, operation)
	}

	// Otherwise use general permission validation
	return validator.ValidatePermission(ctx, tokenString, did, operation)
}

// validateWebAuthnUCANPermission validates UCAN authorization for WebAuthn operations
func (ms msgServer) validateWebAuthnUCANPermission(
	ctx context.Context,
	did string,
	credentialID string,
	operation types.DIDOperation,
) error {
	// Try to extract UCAN token
	tokenString, hasToken := ms.extractUCANToken(ctx)
	if !hasToken {
		// No UCAN token present, proceed with normal validation
		return nil
	}

	// Validate WebAuthn delegation
	validator := ms.k.GetPermissionValidator()
	if validator == nil {
		return fmt.Errorf("UCAN permission validator not initialized")
	}

	return validator.ValidateWebAuthnDelegation(ctx, tokenString, did, credentialID, operation)
}

// checkGaslessSupport checks if the operation can be executed gaslessly via UCAN
func (ms msgServer) checkGaslessSupport(
	ctx context.Context,
	did string,
	operation types.DIDOperation,
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

	supportsGasless, gasLimit, err := validator.SupportsGaslessTransaction(
		ctx,
		tokenString,
		did,
		operation,
	)
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

// CreateDID implements types.MsgServer.
func (ms msgServer) CreateDID(
	ctx context.Context,
	msg *types.MsgCreateDID,
) (*types.MsgCreateDIDResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.DidDocument.Id, msg.Controller, types.DIDOpCreate); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Check for gasless execution support
	supportsGasless, gasLimit := ms.checkGaslessSupport(ctx, msg.DidDocument.Id, types.DIDOpCreate)
	if supportsGasless {
		// Log gasless execution (in production, this might set transaction fees to zero)
		ms.k.Logger().Info("Executing DID creation with gasless transaction",
			"did", msg.DidDocument.Id,
			"gas_limit", gasLimit)
	}

	// Check if DID already exists
	exists, err := ms.k.OrmDB.DIDDocumentTable().Has(ctx, msg.DidDocument.Id)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToCheckDIDExists, "%s: %v", msg.DidDocument.Id, err)
	}
	if exists {
		return nil, errors.Wrapf(types.ErrDIDAlreadyExists, "%s", msg.DidDocument.Id)
	}

	// Validate DID document
	if err := ms.validateDIDDocument(&msg.DidDocument); err != nil {
		return nil, err
	}

	// Set creation metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	blockTime := sdkCtx.BlockTime().Unix()

	didDocument := msg.DidDocument
	didDocument.CreatedAt = blockHeight
	didDocument.UpdatedAt = blockHeight
	didDocument.Deactivated = false
	didDocument.Version = 1

	// Convert to ORM type and insert
	ormDoc := didDocument.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Insert(ctx, ormDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToStoreDIDDocument, "%v", err)
	}

	// Create DID document metadata
	metadata := &types.DIDDocumentMetadata{
		Did:           didDocument.Id,
		Created:       blockTime,
		Updated:       blockTime,
		Deactivated:   0,
		VersionId:     "1",
		NextUpdate:    0,
		NextVersionId: "",
		EquivalentId:  []string{},
		CanonicalId:   didDocument.Id,
	}

	// Convert metadata to ORM type and insert
	ormMetadata := metadata.ToORM()
	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Insert(ctx, ormMetadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToStoreDIDMetadata, "%v", err)
	}

	// Auto-create vault for the new DID
	vaultID := fmt.Sprintf("%s-vault", didDocument.Id)
	keyID := fmt.Sprintf("%s-key-1", didDocument.Id)

	vaultResp, vaultErr := ms.k.CreateVaultForDID(
		ctx,
		didDocument.Id,
		msg.Controller,
		vaultID,
		keyID,
	)
	if vaultErr != nil {
		// Log warning but don't fail DID creation
		ms.k.Logger().With(
			"did", didDocument.Id,
			"vault_id", vaultID,
			"error", vaultErr,
		).Warn("Failed to auto-create vault for DID")
	}

	// Emit typed event
	event := &types.EventDIDCreated{
		Did:         didDocument.Id,
		Creator:     msg.Controller,
		PublicKeys:  extractPublicKeys(&didDocument),
		Services:    extractServiceIds(&didDocument),
		CreatedAt:   sdkCtx.BlockTime(),
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventDIDCreated")
	}

	// Build response with vault information if creation succeeded
	response := &types.MsgCreateDIDResponse{
		Did: didDocument.Id,
	}

	if vaultResp != nil {
		response.VaultId = vaultResp.VaultID
		// Convert string public key to bytes
		if vaultResp.VaultPublicKey != "" {
			pubKeyBytes, err := base64.StdEncoding.DecodeString(vaultResp.VaultPublicKey)
			if err == nil {
				response.VaultPublicKey = pubKeyBytes
			}
		}
		response.EnclaveId = vaultResp.EnclaveID
	}

	return response, nil
}

// UpdateDID implements types.MsgServer.
func (ms msgServer) UpdateDID(
	ctx context.Context,
	msg *types.MsgUpdateDID,
) (*types.MsgUpdateDIDResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpUpdate); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	existingDoc := types.DIDDocumentFromORM(ormDoc)

	// Store copy of old document for comparison
	oldDoc := *existingDoc

	// Check if DID is deactivated
	if existingDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(existingDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Validate updated DID document
	if err := ms.validateDIDDocument(&msg.DidDocument); err != nil {
		return nil, err
	}

	// Ensure DID ID matches
	if msg.DidDocument.Id != msg.Did {
		return nil, errors.Wrapf(
			types.ErrDIDMismatch,
			"document ID %s does not match message DID %s",
			msg.DidDocument.Id,
			msg.Did,
		)
	}

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	blockTime := sdkCtx.BlockTime().Unix()

	updatedDoc := msg.DidDocument
	updatedDoc.CreatedAt = existingDoc.CreatedAt // Preserve creation time
	updatedDoc.UpdatedAt = blockHeight
	updatedDoc.Version = existingDoc.Version + 1
	updatedDoc.Deactivated = false

	// Convert to ORM type and update
	ormUpdatedDoc := updatedDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = blockTime
	metadata.VersionId = fmt.Sprintf("%d", updatedDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventDIDUpdated{
		Did:           msg.Did,
		Updater:       msg.Controller,
		FieldsUpdated: extractFieldsUpdated(&oldDoc, &updatedDoc),
		UpdatedAt:     sdkCtx.BlockTime(),
		BlockHeight:   uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventDIDUpdated")
	}

	return &types.MsgUpdateDIDResponse{}, nil
}

// DeactivateDID implements types.MsgServer.
func (ms msgServer) DeactivateDID(
	ctx context.Context,
	msg *types.MsgDeactivateDID,
) (*types.MsgDeactivateDIDResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpDeactivate); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	existingDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is already deactivated
	if existingDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDAlreadyDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(existingDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized to deactivate DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Update DID document
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	blockTime := sdkCtx.BlockTime().Unix()

	existingDoc.Deactivated = true
	existingDoc.UpdatedAt = blockHeight
	existingDoc.Version = existingDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := existingDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToDeactivateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = blockTime
	metadata.Deactivated = blockTime
	metadata.VersionId = fmt.Sprintf("%d", existingDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventDIDDeactivated{
		Did:           msg.Did,
		Deactivator:   msg.Controller,
		DeactivatedAt: sdkCtx.BlockTime(),
		BlockHeight:   uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventDIDDeactivated")
	}

	return &types.MsgDeactivateDIDResponse{}, nil
}

// AddVerificationMethod implements types.MsgServer.
func (ms msgServer) AddVerificationMethod(
	ctx context.Context,
	msg *types.MsgAddVerificationMethod,
) (*types.MsgAddVerificationMethodResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpAddVerificationMethod); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is deactivated
	if didDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(didDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Validate the new verification method
	if err := ms.validateVerificationMethod(&msg.VerificationMethod); err != nil {
		return nil, err
	}

	// Check if verification method ID already exists
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == msg.VerificationMethod.Id {
			return nil, errors.Wrapf(types.ErrVerificationMethodAlreadyExists, "%s", vm.Id)
		}
	}

	// Add the verification method
	didDoc.VerificationMethod = append(didDoc.VerificationMethod, &msg.VerificationMethod)

	// Add to specified relationships
	for _, relationship := range msg.Relationships {
		switch relationship {
		case "authentication":
			didDoc.Authentication = append(
				didDoc.Authentication,
				&types.VerificationMethodReference{
					VerificationMethodId: msg.VerificationMethod.Id,
				},
			)
		case "assertionMethod":
			didDoc.AssertionMethod = append(
				didDoc.AssertionMethod,
				&types.VerificationMethodReference{
					VerificationMethodId: msg.VerificationMethod.Id,
				},
			)
		case "keyAgreement":
			didDoc.KeyAgreement = append(didDoc.KeyAgreement, &types.VerificationMethodReference{
				VerificationMethodId: msg.VerificationMethod.Id,
			})
		case "capabilityInvocation":
			didDoc.CapabilityInvocation = append(
				didDoc.CapabilityInvocation,
				&types.VerificationMethodReference{
					VerificationMethodId: msg.VerificationMethod.Id,
				},
			)
		case "capabilityDelegation":
			didDoc.CapabilityDelegation = append(
				didDoc.CapabilityDelegation,
				&types.VerificationMethodReference{
					VerificationMethodId: msg.VerificationMethod.Id,
				},
			)
		}
	}

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	didDoc.UpdatedAt = blockHeight
	didDoc.Version = didDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := didDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = sdkCtx.BlockTime().Unix()
	metadata.VersionId = fmt.Sprintf("%d", didDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventVerificationMethodAdded{
		Did:         msg.Did,
		MethodId:    msg.VerificationMethod.Id,
		KeyType:     msg.VerificationMethod.VerificationMethodKind,
		PublicKey:   msg.VerificationMethod.PublicKeyMultibase,
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventVerificationMethodAdded")
	}

	return &types.MsgAddVerificationMethodResponse{}, nil
}

// RemoveVerificationMethod implements types.MsgServer.
func (ms msgServer) RemoveVerificationMethod(
	ctx context.Context,
	msg *types.MsgRemoveVerificationMethod,
) (*types.MsgRemoveVerificationMethodResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpRemoveVerificationMethod); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is deactivated
	if didDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(didDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Find and remove the verification method
	found := false
	var newVerificationMethods []*types.VerificationMethod
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id != msg.VerificationMethodId {
			newVerificationMethods = append(newVerificationMethods, vm)
		} else {
			found = true
		}
	}

	if !found {
		return nil, errors.Wrapf(
			types.ErrVerificationMethodNotFound,
			"%s",
			msg.VerificationMethodId,
		)
	}

	didDoc.VerificationMethod = newVerificationMethods

	// Remove from all relationships
	didDoc.Authentication = ms.removeVerificationMethodReference(
		didDoc.Authentication,
		msg.VerificationMethodId,
	)
	didDoc.AssertionMethod = ms.removeVerificationMethodReference(
		didDoc.AssertionMethod,
		msg.VerificationMethodId,
	)
	didDoc.KeyAgreement = ms.removeVerificationMethodReference(
		didDoc.KeyAgreement,
		msg.VerificationMethodId,
	)
	didDoc.CapabilityInvocation = ms.removeVerificationMethodReference(
		didDoc.CapabilityInvocation,
		msg.VerificationMethodId,
	)
	didDoc.CapabilityDelegation = ms.removeVerificationMethodReference(
		didDoc.CapabilityDelegation,
		msg.VerificationMethodId,
	)

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	didDoc.UpdatedAt = blockHeight
	didDoc.Version = didDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := didDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = sdkCtx.BlockTime().Unix()
	metadata.VersionId = fmt.Sprintf("%d", didDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventVerificationMethodRemoved{
		Did:         msg.Did,
		MethodId:    msg.VerificationMethodId,
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventVerificationMethodRemoved")
	}

	return &types.MsgRemoveVerificationMethodResponse{}, nil
}

// AddService implements types.MsgServer.
func (ms msgServer) AddService(
	ctx context.Context,
	msg *types.MsgAddService,
) (*types.MsgAddServiceResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpAddService); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is deactivated
	if didDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(didDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Validate the new service
	if err := ms.validateService(&msg.Service); err != nil {
		return nil, err
	}

	// Check if service ID already exists
	for _, svc := range didDoc.Service {
		if svc.Id == msg.Service.Id {
			return nil, errors.Wrapf(types.ErrServiceAlreadyExists, "%s", svc.Id)
		}
	}

	// Add the service
	didDoc.Service = append(didDoc.Service, &msg.Service)

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	didDoc.UpdatedAt = blockHeight
	didDoc.Version = didDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := didDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = sdkCtx.BlockTime().Unix()
	metadata.VersionId = fmt.Sprintf("%d", didDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventServiceAdded{
		Did:         msg.Did,
		ServiceId:   msg.Service.Id,
		Type:        msg.Service.ServiceKind,
		Endpoint:    msg.Service.SingleEndpoint,
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventServiceAdded")
	}

	return &types.MsgAddServiceResponse{}, nil
}

// RemoveService implements types.MsgServer.
func (ms msgServer) RemoveService(
	ctx context.Context,
	msg *types.MsgRemoveService,
) (*types.MsgRemoveServiceResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation
	if err := ms.validateUCANPermission(ctx, msg.Did, msg.Controller, types.DIDOpRemoveService); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is deactivated
	if didDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(didDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Find and remove the service
	found := false
	var newServices []*types.Service
	for _, svc := range didDoc.Service {
		if svc.Id != msg.ServiceId {
			newServices = append(newServices, svc)
		} else {
			found = true
		}
	}

	if !found {
		return nil, errors.Wrapf(types.ErrServiceNotFound, "%s", msg.ServiceId)
	}

	didDoc.Service = newServices

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	didDoc.UpdatedAt = blockHeight
	didDoc.Version = didDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := didDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = sdkCtx.BlockTime().Unix()
	metadata.VersionId = fmt.Sprintf("%d", didDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventServiceRemoved{
		Did:         msg.Did,
		ServiceId:   msg.ServiceId,
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventServiceRemoved")
	}

	return &types.MsgRemoveServiceResponse{}, nil
}

// IssueVerifiableCredential implements types.MsgServer.
func (ms msgServer) IssueVerifiableCredential(
	ctx context.Context,
	msg *types.MsgIssueVerifiableCredential,
) (*types.MsgIssueVerifiableCredentialResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// UCAN authorization validation for credential issuing
	if err := ms.validateUCANPermission(ctx, msg.Credential.Issuer, msg.Issuer, types.DIDOpIssueCredential); err != nil {
		return nil, errors.Wrapf(types.ErrUCANValidationFailed, "UCAN validation failed: %v", err)
	}

	// Validate issuer DID exists
	ormIssuerDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Credential.Issuer)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "issuer %s", msg.Credential.Issuer)
	}

	// Convert from ORM type
	issuerDoc := types.DIDDocumentFromORM(ormIssuerDoc)

	// Check if issuer DID is deactivated
	if issuerDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "issuer %s", msg.Credential.Issuer)
	}

	// Validate issuer authorization
	if !ms.isAuthorizedController(issuerDoc, msg.Issuer) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"address %s not authorized for issuer DID %s",
			msg.Issuer,
			msg.Credential.Issuer,
		)
	}

	// Check if credential ID already exists
	exists, err := ms.k.OrmDB.VerifiableCredentialTable().Has(ctx, msg.Credential.Id)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToCheckCredentialExists, "%v", err)
	}
	if exists {
		return nil, errors.Wrapf(types.ErrCredentialAlreadyExists, "%s", msg.Credential.Id)
	}

	// Set credential metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	credential := msg.Credential
	credential.IssuedAt = blockHeight
	credential.Revoked = false

	// Parse expiration date if provided
	if credential.ExpirationDate != "" {
		// For now, we'll set a default expiration block height
		// In a full implementation, this would parse the ISO date
		credential.ExpiresAt = blockHeight + 365*24*60*60/6 // ~1 year in blocks (6s blocks)
	} else {
		credential.ExpiresAt = 0 // No expiration
	}

	// Convert to ORM type and store the credential
	ormCredential := credential.ToORM()
	if err := ms.k.OrmDB.VerifiableCredentialTable().Insert(ctx, ormCredential); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToStoreCredential, "%v", err)
	}

	// Emit typed event
	credentialType := ""
	if len(credential.CredentialKinds) > 0 {
		credentialType = credential.CredentialKinds[0]
	}

	event := &types.EventCredentialIssued{
		CredentialId: credential.Id,
		Issuer:       credential.Issuer,
		Subject:      credential.Subject,
		Type:         credentialType,
		IssuedAt:     sdkCtx.BlockTime(),
		BlockHeight:  uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventCredentialIssued")
	}

	return &types.MsgIssueVerifiableCredentialResponse{
		CredentialId: credential.Id,
	}, nil
}

// RevokeVerifiableCredential implements types.MsgServer.
func (ms msgServer) RevokeVerifiableCredential(
	ctx context.Context,
	msg *types.MsgRevokeVerifiableCredential,
) (*types.MsgRevokeVerifiableCredentialResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get the credential
	ormCredential, err := ms.k.OrmDB.VerifiableCredentialTable().Get(ctx, msg.CredentialId)
	if err != nil {
		return nil, errors.Wrapf(types.ErrCredentialNotFound, "%s", msg.CredentialId)
	}

	// Convert from ORM type
	credential := types.VerifiableCredentialFromORM(ormCredential)

	// Check if already revoked
	if credential.Revoked {
		return nil, errors.Wrapf(types.ErrCredentialAlreadyRevoked, "%s", msg.CredentialId)
	}

	// Validate that the revoker is the issuer
	ormIssuerDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, credential.Issuer)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "issuer %s", credential.Issuer)
	}

	// Convert from ORM type
	issuerDoc := types.DIDDocumentFromORM(ormIssuerDoc)

	// Validate issuer authorization
	if !ms.isAuthorizedController(issuerDoc, msg.Issuer) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"address %s not authorized for issuer DID %s",
			msg.Issuer,
			credential.Issuer,
		)
	}

	// Update credential status
	credential.Revoked = true

	// Convert to ORM type and update the credential
	ormUpdatedCredential := credential.ToORM()
	if err := ms.k.OrmDB.VerifiableCredentialTable().Update(ctx, ormUpdatedCredential); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateCredential, "%v", err)
	}

	// Emit typed event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventCredentialRevoked{
		CredentialId: msg.CredentialId,
		Revoker:      msg.Issuer,
		Reason:       msg.RevocationReason,
		RevokedAt:    sdkCtx.BlockTime(),
		BlockHeight:  uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventCredentialRevoked")
	}

	return &types.MsgRevokeVerifiableCredentialResponse{}, nil
}

// LinkExternalWallet implements types.MsgServer.
func (ms msgServer) LinkExternalWallet(
	ctx context.Context,
	msg *types.MsgLinkExternalWallet,
) (*types.MsgLinkExternalWalletResponse, error) {
	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get existing DID document
	ormDoc, err := ms.k.OrmDB.DIDDocumentTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", msg.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Check if DID is deactivated
	if didDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "%s", msg.Did)
	}

	// Validate controller authorization
	if !ms.isAuthorizedController(didDoc, msg.Controller) {
		return nil, errors.Wrapf(
			types.ErrUnauthorized,
			"controller %s not authorized for DID %s",
			msg.Controller,
			msg.Did,
		)
	}

	// Validate wallet type
	walletType := types.WalletType(msg.WalletType)
	if err := walletType.Validate(); err != nil {
		return nil, err
	}

	// Validate DWN vault controller requirement
	if err := ms.k.ValidateDWNVaultController(ctx, msg.Did); err != nil {
		return nil, errors.Wrap(types.ErrDWNVaultControllerRequired, err.Error())
	}

	// Check if wallet is already linked to any DID
	if err := ms.k.CheckWalletNotAlreadyLinked(ctx, msg.WalletAddress, msg.WalletChainId, walletType); err != nil {
		return nil, errors.Wrap(types.ErrWalletAlreadyLinked, err.Error())
	}

	// Verify wallet ownership through signature
	verifyErr := ms.k.VerifyWalletOwnership(
		ctx,
		msg.WalletAddress,
		msg.WalletChainId,
		walletType,
		msg.Challenge,
		msg.OwnershipProof,
	)
	if verifyErr != nil {
		return nil, verifyErr
	}

	// Check if verification method ID already exists
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == msg.VerificationMethodId {
			return nil, errors.Wrapf(types.ErrVerificationMethodAlreadyExists, "%s", vm.Id)
		}
	}

	// Create the verification method for the external wallet
	verificationMethod, err := ms.k.CreateVerificationMethodFromWallet(
		msg.VerificationMethodId,
		msg.Did,
		msg.WalletAddress,
		msg.WalletChainId,
		walletType,
	)
	if err != nil {
		return nil, errors.Wrapf(
			types.ErrInvalidVerificationMethod,
			"failed to create verification method: %v",
			err,
		)
	}

	// Add the verification method to the DID document
	didDoc.VerificationMethod = append(didDoc.VerificationMethod, verificationMethod)

	// Add to assertion method relationship (required for external wallets)
	didDoc.AssertionMethod = append(didDoc.AssertionMethod, &types.VerificationMethodReference{
		VerificationMethodId: msg.VerificationMethodId,
	})

	// Update metadata
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	didDoc.UpdatedAt = blockHeight
	didDoc.Version = didDoc.Version + 1

	// Convert to ORM type and update
	ormUpdatedDoc := didDoc.ToORM()
	if err := ms.k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdatedDoc); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDDocument, "%v", err)
	}

	// Update metadata
	metadata, err := ms.k.OrmDB.DIDDocumentMetadataTable().Get(ctx, msg.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrFailedToGetDIDMetadata, "%v", err)
	}

	metadata.Updated = sdkCtx.BlockTime().Unix()
	metadata.VersionId = fmt.Sprintf("%d", didDoc.Version)

	if err := ms.k.OrmDB.DIDDocumentMetadataTable().Update(ctx, metadata); err != nil {
		return nil, errors.Wrapf(types.ErrFailedToUpdateDIDMetadata, "%v", err)
	}

	// Emit typed event
	event := &types.EventExternalWalletLinked{
		Did:           msg.Did,
		WalletType:    msg.WalletType,
		WalletAddress: msg.WalletAddress,
		BlockHeight:   uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventExternalWalletLinked")
	}

	return &types.MsgLinkExternalWalletResponse{
		VerificationMethodId: msg.VerificationMethodId,
	}, nil
}

// Helper functions for DID operations

// validateDIDDocument validates a W3C DID document structure
func (ms msgServer) validateDIDDocument(doc *types.DIDDocument) error {
	// Validate required fields
	if doc.Id == "" {
		return types.ErrMissingDIDDocumentID
	}

	// Validate DID syntax (basic check)
	if !ms.isValidDIDSyntax(doc.Id) {
		return types.ErrInvalidDIDSyntax
	}

	// Validate verification methods
	for _, vm := range doc.VerificationMethod {
		if err := ms.validateVerificationMethod(vm); err != nil {
			return errors.Wrapf(types.ErrInvalidVerificationMethod, "%s: %v", vm.Id, err)
		}
	}

	// Validate services
	for _, service := range doc.Service {
		if err := ms.validateService(service); err != nil {
			return errors.Wrapf(types.ErrInvalidService, "%s: %v", service.Id, err)
		}
	}

	return nil
}

// validateVerificationMethod validates a verification method structure
func (ms msgServer) validateVerificationMethod(vm *types.VerificationMethod) error {
	if vm.Id == "" {
		return types.ErrMissingVerificationMethodID
	}
	if vm.VerificationMethodKind == "" {
		return types.ErrMissingVerificationMethodKind
	}
	if vm.Controller == "" {
		return types.ErrMissingVerificationMethodController
	}

	// Check that at least one public key material is provided
	hasStandardKey := vm.PublicKeyJwk != "" || vm.PublicKeyMultibase != "" ||
		vm.PublicKeyBase58 != "" ||
		vm.PublicKeyBase64 != "" ||
		vm.PublicKeyPem != "" ||
		vm.PublicKeyHex != ""

	// Check for WebAuthn credential
	hasWebAuthnKey := vm.WebauthnCredential != nil && vm.WebauthnCredential.CredentialId != ""

	if !hasStandardKey && !hasWebAuthnKey {
		return types.ErrMissingVerificationMethodKey
	}

	// If WebAuthn credential is present, validate it
	if hasWebAuthnKey {
		if err := ms.validateWebAuthnCredentialId(vm.WebauthnCredential.CredentialId); err != nil {
			return err
		}
	}

	return nil
}

// validateWebAuthnCredentialId validates a WebAuthn credential ID
func (ms msgServer) validateWebAuthnCredentialId(credentialId string) error {
	if credentialId == "" {
		return errors.Wrap(types.ErrInvalidVerificationMethod, "WebAuthn credential ID is required")
	}

	// WebAuthn credential validation is now handled by types/webauthn package
	// Additional validation should use types/webauthn validation functions
	return nil
}

// validateService validates a service endpoint structure
func (ms msgServer) validateService(service *types.Service) error {
	if service.Id == "" {
		return types.ErrMissingServiceID
	}
	if service.ServiceKind == "" {
		return types.ErrMissingServiceKind
	}

	// Check that at least one endpoint is provided
	if service.SingleEndpoint == "" && service.MultipleEndpoints == nil &&
		service.ComplexEndpoint == nil {
		return types.ErrMissingServiceEndpoint
	}

	return nil
}

// isValidDIDSyntax validates DID syntax according to W3C DID Core specification
// ABNF: did = "did:" method-name ":" method-specific-id
func (ms msgServer) isValidDIDSyntax(did string) bool {
	// Minimum length check: "did:x:y" = 7 characters
	if len(did) < 7 {
		return false
	}

	// Must start with "did:"
	if !strings.HasPrefix(did, "did:") {
		return false
	}

	// Split into components
	parts := strings.SplitN(did[4:], ":", 2)
	if len(parts) != 2 {
		return false // Missing method-name or method-specific-id
	}

	methodName := parts[0]
	methodSpecificID := parts[1]

	// Validate method-name: 1*method-char
	// method-char = %x61-7A / DIGIT (lowercase letters a-z or digits 0-9)
	if !ms.isValidMethodName(methodName) {
		return false
	}

	// Validate method-specific-id: *( *idchar ":" ) 1*idchar
	if !ms.isValidMethodSpecificID(methodSpecificID) {
		return false
	}

	return true
}

// isValidMethodName validates DID method name according to W3C spec
// ABNF: method-name = 1*method-char
// method-char = %x61-7A / DIGIT
func (ms msgServer) isValidMethodName(methodName string) bool {
	if len(methodName) == 0 {
		return false
	}

	for _, ch := range methodName {
		// Must be lowercase letter (a-z) or digit (0-9)
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

// isValidMethodSpecificID validates method-specific-id according to W3C spec
// ABNF: method-specific-id = *( *idchar ":" ) 1*idchar
// idchar = ALPHA / DIGIT / "." / "-" / "_" / pct-encoded
func (ms msgServer) isValidMethodSpecificID(id string) bool {
	if len(id) == 0 {
		return false
	}

	// Split by ':' to validate each segment
	segments := strings.Split(id, ":")
	for _, segment := range segments {
		// Each segment must contain at least one idchar
		if len(segment) == 0 {
			// Empty segment is allowed except for the last one
			continue
		}
		if !ms.isValidIDSegment(segment) {
			return false
		}
	}

	// Ensure the last segment is not empty
	if len(segments) > 0 && len(segments[len(segments)-1]) == 0 {
		return false
	}

	return true
}

// isValidIDSegment validates a segment of method-specific-id
// idchar = ALPHA / DIGIT / "." / "-" / "_" / pct-encoded
func (ms msgServer) isValidIDSegment(segment string) bool {
	i := 0
	for i < len(segment) {
		ch := segment[i]

		// Check for percent-encoded characters (%HEXDIG HEXDIG)
		if ch == '%' {
			if i+2 >= len(segment) {
				return false // Not enough characters for percent-encoding
			}
			// Validate next two characters are hex digits
			if !isHexDigit(segment[i+1]) || !isHexDigit(segment[i+2]) {
				return false
			}
			i += 3
			continue
		}

		// Check for valid idchar: ALPHA / DIGIT / "." / "-" / "_"
		if !isValidIDChar(ch) {
			return false
		}
		i++
	}
	return true
}

// isValidIDChar checks if a character is valid for idchar (excluding percent-encoding)
func isValidIDChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || // Uppercase letters
		(ch >= 'a' && ch <= 'z') || // Lowercase letters
		(ch >= '0' && ch <= '9') || // Digits
		ch == '.' || ch == '-' || ch == '_' // Special characters
}

// isHexDigit checks if a character is a valid hexadecimal digit
func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'A' && ch <= 'F') ||
		(ch >= 'a' && ch <= 'f')
}

// isAuthorizedController validates controller authorization through verification methods
func (ms msgServer) isAuthorizedController(doc *types.DIDDocument, controller string) bool {
	// Check if controller is the primary controller
	if doc.PrimaryController == controller {
		return true
	}

	// Check if the controller owns any verification method with capability rights
	for _, vm := range doc.VerificationMethod {
		// Check if this verification method is controlled by the requesting controller
		if vm.Controller == controller {
			// Check if this verification method is referenced in capability relationships
			if ms.hasCapabilityRights(doc, vm.Id) {
				return true
			}
		}
	}

	// Check for delegation through controller DIDs
	if ms.hasControllerDelegation(doc, controller) {
		return true
	}

	return false
}

// hasCapabilityRights checks if a verification method has capability rights
func (ms msgServer) hasCapabilityRights(doc *types.DIDDocument, vmID string) bool {
	// Check capabilityInvocation - allows invoking capabilities on behalf of the DID
	for _, ref := range doc.CapabilityInvocation {
		if ref.VerificationMethodId == vmID {
			return true
		}
	}

	// Check capabilityDelegation - allows delegating capabilities to others
	for _, ref := range doc.CapabilityDelegation {
		if ref.VerificationMethodId == vmID {
			return true
		}
	}

	return false
}

// hasControllerDelegation checks if a controller has delegation through controller relationships
func (ms msgServer) hasControllerDelegation(doc *types.DIDDocument, controller string) bool {
	// Check if controller is referenced in verification methods as a delegated controller
	for _, vm := range doc.VerificationMethod {
		// A verification method can delegate control if:
		// 1. Its controller field references the requesting controller
		// 2. It has capability delegation rights
		if vm.Controller == controller {
			// Check if any capability delegation references point to this VM
			for _, ref := range doc.CapabilityDelegation {
				// Find the verification method that has delegation rights
				delegatingVM := ms.findVerificationMethod(doc, ref.VerificationMethodId)
				if delegatingVM != nil && delegatingVM.Controller == doc.Id {
					// The DID document owner has delegated capabilities
					return true
				}
			}
		}
	}

	return false
}

// findVerificationMethod finds a verification method by ID in the document
func (ms msgServer) findVerificationMethod(
	doc *types.DIDDocument,
	vmID string,
) *types.VerificationMethod {
	for _, vm := range doc.VerificationMethod {
		if vm.Id == vmID {
			return vm
		}
	}
	return nil
}

// removeVerificationMethodReference removes a verification method reference from a list
func (ms msgServer) removeVerificationMethodReference(
	refs []*types.VerificationMethodReference,
	vmId string,
) []*types.VerificationMethodReference {
	var newRefs []*types.VerificationMethodReference
	for _, ref := range refs {
		if ref.VerificationMethodId != vmId {
			newRefs = append(newRefs, ref)
		}
	}
	return newRefs
}

// RegisterWebAuthnCredential implements types.MsgServer.
func (ms msgServer) RegisterWebAuthnCredential(
	ctx context.Context,
	msg *types.MsgRegisterWebAuthnCredential,
) (*types.MsgRegisterWebAuthnCredentialResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate the controller address
	controllerAddr, err := sdk.AccAddressFromBech32(msg.Controller)
	if err != nil {
		return nil, types.ErrInvalidControllerAddress.Wrapf("invalid controller address: %s", err)
	}

	// Extract assertion information from the message
	// The username should contain the assertion value (email/tel)
	var assertionType, assertionValue string
	if strings.Contains(msg.Username, "@") {
		// Email assertion
		assertionType = "email"
		assertionValue = msg.Username
	} else if strings.HasPrefix(msg.Username, "+") || containsOnlyDigits(msg.Username) {
		// Phone assertion
		assertionType = "tel"
		assertionValue = msg.Username
	} else {
		// Default to username-based assertion
		assertionType = "sonr"
		assertionValue = msg.Username
	}

	// Hash the assertion value and create DID
	hashedValue := types.HashAssertionValue(assertionValue)
	assertionDID := fmt.Sprintf("did:%s:%s", assertionType, hashedValue)

	// Check if assertion already exists
	existingAssertion, _ := ms.k.OrmDB.AssertionTable().Get(ctx, assertionDID)
	if existingAssertion != nil {
		return nil, types.ErrDIDAlreadyExists.Wrapf("assertion already registered: %s", assertionDID)
	}

	// UCAN authorization validation for WebAuthn registration
	// For new registrations, we check if a UCAN token grants registration permission
	if msg.Username != "" {
		// Create a temporary DID pattern for authorization check
		didPattern := fmt.Sprintf("sonr:%s", msg.Username)
		if err := ms.validateUCANPermission(ctx, didPattern, msg.Controller, types.DIDOpRegisterWebAuthn); err != nil {
			// WebAuthn registration can proceed without UCAN for gasless onboarding
			ms.k.Logger().Debug("UCAN validation for WebAuthn registration", "error", err)
		}
	}

	// Check if we've exceeded the maximum number of credentials per DID
	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	if params.Webauthn != nil && params.Webauthn.MaxCredentialsPerDid > 0 {
		// For new registration, we can check existing credentials if a DID already exists
		// For now, we'll proceed with registration and handle limits in future iterations
	}

	// Create WebAuthn registration data from the message
	regData := &WebAuthnRegistrationData{
		CredentialID:      msg.WebauthnCredential.CredentialId,
		RawID:             msg.WebauthnCredential.RawId,
		ClientDataJSON:    msg.WebauthnCredential.ClientDataJson,
		AttestationObject: msg.WebauthnCredential.AttestationObject,
		Username:          msg.Username,
		PublicKey:         msg.WebauthnCredential.PublicKey,
		Algorithm:         msg.WebauthnCredential.Algorithm,
		Origin:            msg.WebauthnCredential.Origin,
	}

	// Process the WebAuthn registration using existing keeper logic
	didDoc, err := ms.k.ProcessWebAuthnRegistration(ctx, regData)
	if err != nil {
		return nil, err
	}

	// Update the DID document to set the controller
	didDoc.PrimaryController = controllerAddr.String()

	// Add assertion methods to the DID document
	// 1. Email/Tel assertion
	assertionMethod := &types.VerificationMethod{
		Id:                     fmt.Sprintf("%s#assertion-%s", didDoc.Id, assertionType),
		Controller:             didDoc.Id,
		VerificationMethodKind: "AssertionMethod2024",
		BlockchainAccountId:    assertionDID,
	}
	didDoc.VerificationMethod = append(didDoc.VerificationMethod, assertionMethod)
	didDoc.AssertionMethod = append(didDoc.AssertionMethod, &types.VerificationMethodReference{
		VerificationMethodId: assertionMethod.Id,
	})

	// 2. Sonr account assertion (controller address)
	sonrAssertion := &types.VerificationMethod{
		Id:                     fmt.Sprintf("%s#sonr-account", didDoc.Id),
		Controller:             didDoc.Id,
		VerificationMethodKind: "BlockchainAccountId2024",
		BlockchainAccountId:    fmt.Sprintf("sonr:%s", controllerAddr.String()),
	}
	didDoc.VerificationMethod = append(didDoc.VerificationMethod, sonrAssertion)
	didDoc.AssertionMethod = append(didDoc.AssertionMethod, &types.VerificationMethodReference{
		VerificationMethodId: sonrAssertion.Id,
	})

	// Store the updated DID document
	if err := ms.k.storeDIDDocument(ctx, didDoc); err != nil {
		return nil, types.ErrFailedToStoreDIDDocument.Wrapf(
			"failed to update DID controller: %s",
			err,
		)
	}

	// Create Controller entry in ORM
	controller := &apiv1.Controller{
		Did:             didDoc.Id,
		Address:         controllerAddr.String(),
		Subject:         assertionValue,
		PublicKeyBase64: base64.StdEncoding.EncodeToString(msg.WebauthnCredential.PublicKey),
		DidKind:         "webauthn",
		CreationBlock:   sdkCtx.BlockHeight(),
	}
	if err := ms.k.OrmDB.ControllerTable().Insert(ctx, controller); err != nil {
		ms.k.Logger().Error("Failed to store controller", "error", err)
	}

	// Create Assertion entry in ORM
	assertion := &apiv1.Assertion{
		Did:             assertionDID,
		Controller:      controllerAddr.String(),
		Subject:         assertionValue,
		PublicKeyBase64: base64.StdEncoding.EncodeToString(msg.WebauthnCredential.PublicKey),
		DidKind:         assertionType,
		CreationBlock:   sdkCtx.BlockHeight(),
	}
	if err := ms.k.OrmDB.AssertionTable().Insert(ctx, assertion); err != nil {
		ms.k.Logger().Error("Failed to store assertion", "error", err)
	}

	// Create Authentication entry in ORM
	authentication := &apiv1.Authentication{
		Did:             didDoc.Id,
		Controller:      controllerAddr.String(),
		Subject:         msg.WebauthnCredential.CredentialId,
		PublicKeyBase64: base64.StdEncoding.EncodeToString(msg.WebauthnCredential.PublicKey),
		DidKind:         "webauthn",
		CreationBlock:   sdkCtx.BlockHeight(),
	}
	if err := ms.k.OrmDB.AuthenticationTable().Insert(ctx, authentication); err != nil {
		ms.k.Logger().Error("Failed to store authentication", "error", err)
	}

	// Initialize UCAN delegation chain
	ucanChain, err := ms.k.InitializeUCANDelegationChain(
		ctx,
		didDoc.Id,
		controllerAddr.String(),
		msg.WebauthnCredential.CredentialId,
	)
	if err != nil {
		// Log error but don't fail the registration
		// UCAN can be initialized later if needed
		ms.k.Logger().Error(
			"Failed to initialize UCAN delegation chain",
			"did", didDoc.Id,
			"error", err,
		)
	}

	// Prepare response
	response := &types.MsgRegisterWebAuthnCredentialResponse{
		Did:                  didDoc.Id,
		VerificationMethodId: msg.VerificationMethodId,
	}

	// Add UCAN tokens to response if successfully initialized
	if ucanChain != nil {
		// Store tokens in response metadata (we may need to add these fields to the response proto)
		// For now, emit them as events
		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				"ucan_delegation_initialized",
				sdk.NewAttribute("did", didDoc.Id),
				sdk.NewAttribute("validator_issuer", ucanChain.ValidatorIssuer),
				sdk.NewAttribute("expires_at", fmt.Sprintf("%d", ucanChain.ExpiresAt)),
			),
		)
	}

	// Auto-create vault if requested
	if msg.AutoCreateVault {
		// Generate vault and key IDs if not provided
		vaultID := fmt.Sprintf("vault-%s", didDoc.Id)
		keyID := fmt.Sprintf("key-%s", didDoc.Id)

		// Create vault using the DWN keeper integration
		vaultResp, err := ms.k.CreateVaultForDID(
			ctx,
			didDoc.Id,
			controllerAddr.String(),
			vaultID,
			keyID,
		)
		if err != nil {
			// Log error but don't fail the registration
			// Vault can be created later if needed
			ms.k.Logger().Error(
				"Failed to create vault for DID",
				"did", didDoc.Id,
				"error", err,
			)
		} else if vaultResp != nil {
			// Update response with vault information
			response.VaultId = vaultResp.VaultID
			// Convert string public key to bytes
			if vaultResp.VaultPublicKey != "" {
				pubKeyBytes, err := base64.StdEncoding.DecodeString(vaultResp.VaultPublicKey)
				if err == nil {
					response.VaultPublicKey = pubKeyBytes
				}
			}
			response.EnclaveId = vaultResp.EnclaveID

			// Emit vault creation event
			sdkCtx.EventManager().EmitEvent(
				sdk.NewEvent(
					"vault_created",
					sdk.NewAttribute("did", didDoc.Id),
					sdk.NewAttribute("vault_id", vaultResp.VaultID),
					sdk.NewAttribute("enclave_id", vaultResp.EnclaveID),
				),
			)
		}
	}

	// Emit typed event
	event := &types.EventWebAuthnRegistered{
		Did:             didDoc.Id,
		CredentialId:    msg.WebauthnCredential.CredentialId,
		AttestationType: msg.WebauthnCredential.AttestationType,
		BlockHeight:     uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		ms.k.Logger().With("error", err).Error("Failed to emit EventWebAuthnRegistered")
	}

	return response, nil
}

// Helper functions for extracting event data

// extractPublicKeys extracts public key IDs from DID document
func extractPublicKeys(doc *types.DIDDocument) []string {
	keys := make([]string, 0, len(doc.VerificationMethod))
	for _, vm := range doc.VerificationMethod {
		keys = append(keys, vm.Id)
	}
	return keys
}

// extractServiceIds extracts service IDs from DID document
func extractServiceIds(doc *types.DIDDocument) []string {
	services := make([]string, 0, len(doc.Service))
	for _, svc := range doc.Service {
		services = append(services, svc.Id)
	}
	return services
}

// containsOnlyDigits checks if a string contains only digits
func containsOnlyDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// extractFieldsUpdated compares two DID documents and returns updated fields
func extractFieldsUpdated(oldDoc, newDoc *types.DIDDocument) []string {
	var fields []string

	// Check verification methods
	if len(oldDoc.VerificationMethod) != len(newDoc.VerificationMethod) {
		fields = append(fields, "verificationMethod")
	}

	// Check services
	if len(oldDoc.Service) != len(newDoc.Service) {
		fields = append(fields, "service")
	}

	// Check authentication
	if len(oldDoc.Authentication) != len(newDoc.Authentication) {
		fields = append(fields, "authentication")
	}

	// Check assertion method
	if len(oldDoc.AssertionMethod) != len(newDoc.AssertionMethod) {
		fields = append(fields, "assertionMethod")
	}

	// Check key agreement
	if len(oldDoc.KeyAgreement) != len(newDoc.KeyAgreement) {
		fields = append(fields, "keyAgreement")
	}

	// Check capability invocation
	if len(oldDoc.CapabilityInvocation) != len(newDoc.CapabilityInvocation) {
		fields = append(fields, "capabilityInvocation")
	}

	// Check capability delegation
	if len(oldDoc.CapabilityDelegation) != len(newDoc.CapabilityDelegation) {
		fields = append(fields, "capabilityDelegation")
	}

	return fields
}
