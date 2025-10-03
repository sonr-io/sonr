package keeper

import (
	"context"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// ProtocolsConfigure configures a protocol in the DWN
func (k Keeper) ProtocolsConfigure(
	ctx context.Context,
	msg *types.MsgProtocolsConfigure,
) (*types.MsgProtocolsConfigureResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate service registration for protocol operations
	// For now, we'll extract serviceID from authorization field if it contains service information
	// In a future version, this could be a dedicated field in the message
	if msg.Authorization != "" {
		// Try to extract service ID from authorization (e.g., "service:serviceID" format)
		// This is a simple implementation - in production, you might parse JWT tokens or other formats
		var serviceID string
		if len(msg.Authorization) > 8 && msg.Authorization[:8] == "service:" {
			serviceID = msg.Authorization[8:]
		}

		if err := k.ValidateServiceForProtocol(sdkCtx, msg.Target, serviceID); err != nil {
			return nil, err
		}
	}

	// Check protocol limits
	params, err := k.Params.Get(sdkCtx)
	if err != nil {
		return nil, err
	}

	// Count existing protocols for this DWN
	protocolCount := 0
	indexKey := apiv1.DWNProtocolTargetProtocolUriIndexKey{}.WithTarget(msg.Target)
	iter, err := k.OrmDB.DWNProtocolTable().List(sdkCtx, indexKey)
	if err == nil {
		defer iter.Close()
		for iter.Next() {
			protocolCount++
		}
	}

	// Check if we're updating or creating new
	existingProtocol, err := k.OrmDB.DWNProtocolTable().Get(sdkCtx, msg.Target, msg.ProtocolUri)
	if err == nil && existingProtocol != nil {
		// Update existing protocol
		existingProtocol.Definition = msg.Definition
		existingProtocol.Published = msg.Published

		if err := k.OrmDB.DWNProtocolTable().Update(sdkCtx, existingProtocol); err != nil {
			return nil, errors.Wrap(err, "failed to update protocol")
		}

		k.Logger().
			Info("Updated DWN protocol", "target", msg.Target, "protocol_uri", msg.ProtocolUri)
	} else {
		// Check limit for new protocol
		if uint32(protocolCount) >= params.MaxProtocolsPerDwn {
			return nil, errors.Wrapf(types.ErrProtocolLimitReached, "protocol limit %d reached for DWN %s", params.MaxProtocolsPerDwn, msg.Target)
		}

		// Create new protocol
		protocol := &apiv1.DWNProtocol{
			Target:        msg.Target,
			ProtocolUri:   msg.ProtocolUri,
			Definition:    msg.Definition,
			Published:     msg.Published,
			CreatedAt:     time.Now().Unix(),
			CreatedHeight: sdkCtx.BlockHeight(),
		}

		if err := k.OrmDB.DWNProtocolTable().Insert(sdkCtx, protocol); err != nil {
			return nil, errors.Wrap(err, "failed to insert protocol")
		}

		k.Logger().Info("Created DWN protocol", "target", msg.Target, "protocol_uri", msg.ProtocolUri)
	}

	// Emit typed event
	event := &types.EventProtocolConfigured{
		Target:      msg.Target,
		ProtocolUri: msg.ProtocolUri,
		Published:   msg.Published,
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		k.Logger().With("error", err).Error("Failed to emit EventProtocolConfigured")
	}

	return &types.MsgProtocolsConfigureResponse{
		ProtocolUri: msg.ProtocolUri,
		Success:     true,
	}, nil
}
