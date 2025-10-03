package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/sonr/x/dwn/types"
)

type msgServer struct {
	k Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (ms msgServer) UpdateParams(
	ctx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(
			types.ErrInvalidAuthorityFormat,
			"invalid authority; expected %s, got %s",
			ms.k.authority,
			msg.Authority,
		)
	}

	return &types.MsgUpdateParamsResponse{}, ms.k.Params.Set(ctx, msg.Params)
}

// RecordsWrite implements the RecordsWrite RPC method
func (ms msgServer) RecordsWrite(
	ctx context.Context,
	msg *types.MsgRecordsWrite,
) (*types.MsgRecordsWriteResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Validate UCAN permissions if authorization token is provided
	if msg.Authorization != "" {
		validator := ms.k.GetPermissionValidator()
		if err := validator.ValidatePermission(
			sdkCtx,
			msg.Authorization,
			msg.Target,
			types.RecordCreate, // Records write is create/update operation
		); err != nil {
			return nil, errors.Wrapf(
				types.ErrPermissionDenied,
				"UCAN validation failed for RecordsWrite: %v", err,
			)
		}
	}

	return ms.k.RecordsWrite(sdkCtx, msg)
}

// RecordsDelete implements the RecordsDelete RPC method
func (ms msgServer) RecordsDelete(
	ctx context.Context,
	msg *types.MsgRecordsDelete,
) (*types.MsgRecordsDeleteResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Validate UCAN permissions if authorization token is provided
	if msg.Authorization != "" {
		validator := ms.k.GetPermissionValidator()
		if err := validator.ValidatePermission(
			ctx,
			msg.Authorization,
			msg.Target,
			types.RecordDelete,
		); err != nil {
			return nil, errors.Wrapf(
				types.ErrPermissionDenied,
				"UCAN validation failed for RecordsDelete: %v", err,
			)
		}
	}

	return ms.k.RecordsDelete(ctx, msg)
}

// ProtocolsConfigure implements the ProtocolsConfigure RPC method
func (ms msgServer) ProtocolsConfigure(
	ctx context.Context,
	msg *types.MsgProtocolsConfigure,
) (*types.MsgProtocolsConfigureResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Validate UCAN permissions if authorization token is provided
	if msg.Authorization != "" {
		validator := ms.k.GetPermissionValidator()
		if err := validator.ValidateProtocolOperation(
			ctx,
			msg.Authorization,
			msg.Target,
			msg.ProtocolUri,
			types.ProtocolOpInstall,
		); err != nil {
			return nil, errors.Wrapf(
				types.ErrPermissionDenied,
				"UCAN validation failed for ProtocolsConfigure: %v", err,
			)
		}
	}

	return ms.k.ProtocolsConfigure(ctx, msg)
}

// PermissionsGrant implements the PermissionsGrant RPC method
func (ms msgServer) PermissionsGrant(
	ctx context.Context,
	msg *types.MsgPermissionsGrant,
) (*types.MsgPermissionsGrantResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Validate UCAN permissions if authorization token is provided
	if msg.Authorization != "" {
		validator := ms.k.GetPermissionValidator()
		if err := validator.ValidatePermission(
			ctx,
			msg.Authorization,
			msg.Target,
			types.PermissionGrant,
		); err != nil {
			return nil, errors.Wrapf(
				types.ErrPermissionDenied,
				"UCAN validation failed for PermissionsGrant: %v", err,
			)
		}
	}

	return ms.k.PermissionsGrant(ctx, msg)
}

// PermissionsRevoke implements the PermissionsRevoke RPC method
func (ms msgServer) PermissionsRevoke(
	ctx context.Context,
	msg *types.MsgPermissionsRevoke,
) (*types.MsgPermissionsRevokeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Validate UCAN permissions if authorization token is provided
	if msg.Authorization != "" {
		validator := ms.k.GetPermissionValidator()
		if err := validator.ValidatePermission(
			ctx,
			msg.Authorization,
			msg.Grantor, // PermissionsRevoke operates on grantor's DWN
			types.PermissionRevoke,
		); err != nil {
			return nil, errors.Wrapf(
				types.ErrPermissionDenied,
				"UCAN validation failed for PermissionsRevoke: %v", err,
			)
		}
	}

	return ms.k.PermissionsRevoke(ctx, msg)
}
