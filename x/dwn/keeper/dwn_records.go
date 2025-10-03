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

// RecordsWrite creates or updates a record in the DWN
func (k Keeper) RecordsWrite(
	ctx context.Context,
	msg *types.MsgRecordsWrite,
) (*types.MsgRecordsWriteResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate service registration for record operations
	if msg.Authorization != "" {
		// Try to extract service ID from authorization
		var serviceID string
		if len(msg.Authorization) > 8 && msg.Authorization[:8] == "service:" {
			serviceID = msg.Authorization[8:]
		}

		if err := k.ValidateServiceForProtocol(ctx, msg.Target, serviceID); err != nil {
			return nil, err
		}

		// Note: Service-based authorization handled by ValidateServiceForProtocol above
		// Legacy UCAN validation is now replaced by service module capabilities
	}

	// Validate record size against params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	if uint64(len(msg.Data)) > params.MaxRecordSize {
		return nil, errors.Wrapf(
			types.ErrRecordSizeExceeded,
			"record size %d exceeds max size %d",
			len(msg.Data),
			params.MaxRecordSize,
		)
	}

	// Determine if record should be encrypted
	shouldEncrypt, err := k.ShouldEncryptRecord(ctx, msg.Protocol, msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine encryption requirement")
	}

	var recordData []byte
	var encryptionMetadata *types.EncryptionMetadata
	var isEncrypted bool

	if shouldEncrypt && k.encryptionSubkeeper != nil {
		// Encrypt the record data using consensus-derived key
		encryptedData, errB := k.encryptionSubkeeper.EncryptWithConsensusKey(
			ctx,
			msg.Data,
			msg.Protocol,
		)
		if errB != nil {
			// Log error but fallback to unencrypted storage
			k.Logger().Error("Failed to encrypt record, storing unencrypted",
				"error", err,
				"protocol", msg.Protocol,
				"schema", msg.Schema,
			)
			recordData = msg.Data
			isEncrypted = false
		} else {
			recordData = encryptedData.Ciphertext
			encryptionMetadata = encryptedData.Metadata
			isEncrypted = true
			k.Logger().Info("Record encrypted successfully",
				"protocol", msg.Protocol,
				"data_size", len(msg.Data),
				"encrypted_size", len(recordData),
			)
		}
	} else {
		recordData = msg.Data
		isEncrypted = false
	}

	// Generate record ID from original content hash (not encrypted data)
	hasher := sha256.New()
	hasher.Write(msg.Data)
	hasher.Write([]byte(msg.Target))
	hasher.Write([]byte(msg.Descriptor_.MessageTimestamp))
	dataHash := hasher.Sum(nil)
	recordID := hex.EncodeToString(dataHash)

	// Calculate data CID (simplified - in production use proper IPLD CID)
	dataCID := "cid:" + hex.EncodeToString(dataHash[:16])

	// Check if record exists
	existingRecord, err := k.OrmDB.DWNRecordTable().Get(ctx, recordID)
	if err == nil && existingRecord != nil {
		// Update existing record
		existingRecord.Data = recordData // Use potentially encrypted data
		existingRecord.Descriptor_ = &apiv1.DWNMessageDescriptor{
			InterfaceName:    msg.Descriptor_.InterfaceName,
			Method:           msg.Descriptor_.Method,
			MessageTimestamp: msg.Descriptor_.MessageTimestamp,
			DataCid:          dataCID,
			DataSize:         int64(len(msg.Data)), // Original data size
			DataFormat:       msg.Descriptor_.DataFormat,
		}
		existingRecord.Authorization = msg.Authorization
		existingRecord.Protocol = msg.Protocol
		existingRecord.ProtocolPath = msg.ProtocolPath
		existingRecord.Schema = msg.Schema
		existingRecord.ParentId = msg.ParentId
		existingRecord.Published = msg.Published
		existingRecord.Encryption = msg.Encryption
		existingRecord.Attestation = msg.Attestation
		existingRecord.UpdatedAt = time.Now().Unix()
		existingRecord.IsEncrypted = isEncrypted
		if encryptionMetadata != nil {
			existingRecord.EncryptionMetadata = encryptionMetadata.ToAPIEncryptionMetadata()
		}

		if err := k.OrmDB.DWNRecordTable().Update(ctx, existingRecord); err != nil {
			return nil, errors.Wrap(err, "failed to update record")
		}

		k.Logger().Info("Updated DWN record",
			"record_id", recordID,
			"target", msg.Target,
			"encrypted", isEncrypted,
		)

		// Emit typed event
		event := &types.EventRecordWritten{
			RecordId:    recordID,
			Target:      msg.Target,
			Protocol:    msg.Protocol,
			Schema:      msg.Schema,
			DataCid:     dataCID,
			DataSize:    uint64(len(msg.Data)),
			Encrypted:   isEncrypted,
			BlockHeight: uint64(sdkCtx.BlockHeight()),
		}

		if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
			k.Logger().With("error", err).Error("Failed to emit EventRecordWritten")
		}
	} else {
		// Create new record
		record := &apiv1.DWNRecord{
			RecordId: recordID,
			Target:   msg.Target,
			Descriptor_: &apiv1.DWNMessageDescriptor{
				InterfaceName:    msg.Descriptor_.InterfaceName,
				Method:           msg.Descriptor_.Method,
				MessageTimestamp: msg.Descriptor_.MessageTimestamp,
				DataCid:          dataCID,
				DataSize:         int64(len(msg.Data)), // Original data size
				DataFormat:       msg.Descriptor_.DataFormat,
			},
			Authorization: msg.Authorization,
			Data:          recordData, // Use potentially encrypted data
			Protocol:      msg.Protocol,
			ProtocolPath:  msg.ProtocolPath,
			Schema:        msg.Schema,
			ParentId:      msg.ParentId,
			Published:     msg.Published,
			Attestation:   msg.Attestation,
			Encryption:    msg.Encryption,
			CreatedAt:     time.Now().Unix(),
			UpdatedAt:     time.Now().Unix(),
			CreatedHeight: sdkCtx.BlockHeight(),
			IsEncrypted:   isEncrypted,
		}

		if encryptionMetadata != nil {
			record.EncryptionMetadata = encryptionMetadata.ToAPIEncryptionMetadata()
		}

		if err := k.OrmDB.DWNRecordTable().Insert(ctx, record); err != nil {
			return nil, errors.Wrap(err, "failed to insert record")
		}

		k.Logger().Info("Created DWN record",
			"record_id", recordID,
			"target", msg.Target,
			"encrypted", isEncrypted,
		)

		// Emit typed event
		event := &types.EventRecordWritten{
			RecordId:    recordID,
			Target:      msg.Target,
			Protocol:    msg.Protocol,
			Schema:      msg.Schema,
			DataCid:     dataCID,
			DataSize:    uint64(len(msg.Data)),
			Encrypted:   isEncrypted,
			BlockHeight: uint64(sdkCtx.BlockHeight()),
		}

		if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
			k.Logger().With("error", err).Error("Failed to emit EventRecordWritten")
		}
	}

	return &types.MsgRecordsWriteResponse{
		RecordId: recordID,
		DataCid:  dataCID,
	}, nil
}

// RecordsDelete deletes a record from the DWN
func (k Keeper) RecordsDelete(
	ctx context.Context,
	msg *types.MsgRecordsDelete,
) (*types.MsgRecordsDeleteResponse, error) {
	// Validate UCAN authorization if provided

	// Get the record
	record, err := k.OrmDB.DWNRecordTable().Get(ctx, msg.RecordId)
	if err != nil {
		return nil, errors.Wrapf(types.ErrRecordNotFound, "record %s not found", msg.RecordId)
	}

	// Verify ownership/permission
	if record.Target != msg.Target {
		return nil, errors.Wrapf(types.ErrRecordPermission, "target mismatch")
	}

	deletedCount := int32(1)

	// Handle pruning of child records if requested
	if msg.Prune && msg.RecordId != "" {
		// Find and delete all child records
		indexKey := apiv1.DWNRecordParentIdIndexKey{}.WithParentId(msg.RecordId)
		iter, err := k.OrmDB.DWNRecordTable().List(ctx, indexKey)
		if err == nil {
			defer iter.Close()
			for iter.Next() {
				childRecord, err := iter.Value()
				if err != nil {
					continue
				}
				if err := k.OrmDB.DWNRecordTable().Delete(ctx, childRecord); err == nil {
					deletedCount++
				}
			}
		}
	}

	// Delete the record
	if err := k.OrmDB.DWNRecordTable().Delete(ctx, record); err != nil {
		return nil, errors.Wrap(err, "failed to delete record")
	}

	k.Logger().Info("Deleted DWN record", "record_id", msg.RecordId, "pruned_count", deletedCount)

	// Emit typed event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventRecordDeleted{
		RecordId:    msg.RecordId,
		Target:      msg.Target,
		Deleter:     msg.Target, // The deleter is the target in this case
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		k.Logger().With("error", err).Error("Failed to emit EventRecordDeleted")
	}

	return &types.MsgRecordsDeleteResponse{
		Success:      true,
		DeletedCount: deletedCount,
	}, nil
}
