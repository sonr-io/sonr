package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// PermissionsGrant grants permissions in the DWN
func (k Keeper) PermissionsGrant(
	ctx context.Context,
	msg *types.MsgPermissionsGrant,
) (*types.MsgPermissionsGrantResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check permission limits
	params, err := k.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	// Count existing permissions for this DWN
	permissionCount := 0
	indexKey := apiv1.DWNPermissionTargetInterfaceNameMethodIndexKey{}.WithTarget(msg.Target)
	iter, err := k.OrmDB.DWNPermissionTable().List(sdkCtx, indexKey)
	if err == nil {
		defer iter.Close()
		for iter.Next() {
			permission, err := iter.Value()
			if err != nil {
				continue
			}
			if !permission.Revoked {
				permissionCount++
			}
		}
	}

	if uint32(permissionCount) >= params.MaxPermissionsPerDwn {
		return nil, errors.Wrapf(
			types.ErrPermissionLimitReached,
			"permission limit %d reached for DWN %s",
			params.MaxPermissionsPerDwn,
			msg.Target,
		)
	}

	// Generate permission ID
	hasher := sha256.New()
	hasher.Write([]byte(msg.Grantor))
	hasher.Write([]byte(msg.Grantee))
	hasher.Write([]byte(msg.Target))
	hasher.Write([]byte(msg.InterfaceName))
	hasher.Write([]byte(msg.Method))
	hasher.Write([]byte(msg.Descriptor_.MessageTimestamp))
	permissionHash := hasher.Sum(nil)
	permissionID := hex.EncodeToString(permissionHash)

	// Create permission
	permission := &apiv1.DWNPermission{
		PermissionId:  permissionID,
		Grantor:       msg.Grantor,
		Grantee:       msg.Grantee,
		Target:        msg.Target,
		InterfaceName: msg.InterfaceName,
		Method:        msg.Method,
		Protocol:      msg.Protocol,
		RecordId:      msg.RecordId,
		Conditions:    msg.Conditions,
		ExpiresAt:     msg.ExpiresAt,
		CreatedAt:     time.Now().Unix(),
		Revoked:       false,
		CreatedHeight: sdkCtx.BlockHeight(),
	}

	if err := k.OrmDB.DWNPermissionTable().Insert(sdkCtx, permission); err != nil {
		return nil, errors.Wrap(err, "failed to insert permission")
	}

	k.Logger().Info("Granted DWN permission",
		"permission_id", permissionID,
		"grantor", msg.Grantor,
		"grantee", msg.Grantee,
		"target", msg.Target,
		"interface", msg.InterfaceName,
		"method", msg.Method,
	)

	// Emit typed event
	event := &types.EventPermissionGranted{
		PermissionId:  permissionID,
		Grantor:       msg.Grantor,
		Grantee:       msg.Grantee,
		InterfaceName: msg.InterfaceName,
		Method:        msg.Method,
		BlockHeight:   uint64(sdkCtx.BlockHeight()),
	}

	// Convert ExpiresAt from int64 to time.Time if it's set
	if msg.ExpiresAt > 0 {
		expiresAt := time.Unix(msg.ExpiresAt, 0)
		event.ExpiresAt = &expiresAt
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		k.Logger().With("error", err).Error("Failed to emit EventPermissionGranted")
	}

	return &types.MsgPermissionsGrantResponse{
		PermissionId: permissionID,
	}, nil
}

// PermissionsRevoke revokes permissions in the DWN
func (k Keeper) PermissionsRevoke(
	ctx context.Context,
	msg *types.MsgPermissionsRevoke,
) (*types.MsgPermissionsRevokeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Get the permission
	permission, err := k.OrmDB.DWNPermissionTable().Get(sdkCtx, msg.PermissionId)
	if err != nil {
		return nil, errors.Wrapf(
			types.ErrPermissionNotFound,
			"permission %s not found",
			msg.PermissionId,
		)
	}

	// Verify the grantor is revoking
	if permission.Grantor != msg.Grantor {
		return nil, errors.Wrapf(types.ErrPermissionDenied, "only grantor can revoke permission")
	}

	// Check if already revoked
	if permission.Revoked {
		return nil, errors.Wrapf(types.ErrPermissionAlreadyRevoked, "permission already revoked")
	}

	// Revoke the permission
	permission.Revoked = true

	if err := k.OrmDB.DWNPermissionTable().Update(sdkCtx, permission); err != nil {
		return nil, errors.Wrap(err, "failed to update permission")
	}

	k.Logger().Info("Revoked DWN permission",
		"permission_id", msg.PermissionId,
		"grantor", msg.Grantor,
	)

	// Emit typed event
	event := &types.EventPermissionRevoked{
		PermissionId: msg.PermissionId,
		Revoker:      msg.Grantor,
		BlockHeight:  uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		k.Logger().With("error", err).Error("Failed to emit EventPermissionRevoked")
	}

	return &types.MsgPermissionsRevokeResponse{
		Success: true,
	}, nil
}
