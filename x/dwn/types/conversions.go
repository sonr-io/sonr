package types

import (
	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
)

// ConvertAPIRecordToType converts from API DWNRecord to types.DWNRecord
func ConvertAPIRecordToType(record *apiv1.DWNRecord) DWNRecord {
	typeRecord := DWNRecord{
		RecordId:            record.RecordId,
		Target:              record.Target,
		Authorization:       record.Authorization,
		Data:                record.Data,
		Protocol:            record.Protocol,
		ProtocolPath:        record.ProtocolPath,
		Schema:              record.Schema,
		ParentId:            record.ParentId,
		Published:           record.Published,
		Attestation:         record.Attestation,
		Encryption:          record.Encryption,
		KeyDerivationScheme: record.KeyDerivationScheme,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
		CreatedHeight:       record.CreatedHeight,
	}

	// Convert descriptor
	if record.Descriptor_ != nil {
		typeRecord.Descriptor_ = &DWNMessageDescriptor{
			InterfaceName:    record.Descriptor_.InterfaceName,
			Method:           record.Descriptor_.Method,
			MessageTimestamp: record.Descriptor_.MessageTimestamp,
			DataCid:          record.Descriptor_.DataCid,
			DataSize:         record.Descriptor_.DataSize,
			DataFormat:       record.Descriptor_.DataFormat,
		}
	}

	return typeRecord
}

// ConvertAPIProtocolToType converts from API DWNProtocol to types.DWNProtocol
func ConvertAPIProtocolToType(protocol *apiv1.DWNProtocol) DWNProtocol {
	return DWNProtocol{
		Target:        protocol.Target,
		ProtocolUri:   protocol.ProtocolUri,
		Definition:    protocol.Definition,
		Published:     protocol.Published,
		CreatedAt:     protocol.CreatedAt,
		CreatedHeight: protocol.CreatedHeight,
	}
}

// ConvertAPIPermissionToType converts from API DWNPermission to types.DWNPermission
func ConvertAPIPermissionToType(permission *apiv1.DWNPermission) DWNPermission {
	return DWNPermission{
		PermissionId:  permission.PermissionId,
		Grantor:       permission.Grantor,
		Grantee:       permission.Grantee,
		Target:        permission.Target,
		InterfaceName: permission.InterfaceName,
		Method:        permission.Method,
		Protocol:      permission.Protocol,
		RecordId:      permission.RecordId,
		Conditions:    permission.Conditions,
		ExpiresAt:     permission.ExpiresAt,
		CreatedAt:     permission.CreatedAt,
		Revoked:       permission.Revoked,
		CreatedHeight: permission.CreatedHeight,
	}
}

// ConvertAPIVaultToType converts from API VaultState to types.VaultState
func ConvertAPIVaultToType(vault *apiv1.VaultState) VaultState {
	typeVault := VaultState{
		VaultId:       vault.VaultId,
		Owner:         vault.Owner,
		PublicKey:     vault.PublicKey,
		CreatedAt:     vault.CreatedAt,
		LastRefreshed: vault.LastRefreshed,
		CreatedHeight: vault.CreatedHeight,
	}

	// Convert enclave data
	if vault.EnclaveData != nil {
		typeVault.EnclaveData = &EnclaveData{
			PrivateData: vault.EnclaveData.PrivateData,
			PublicKey:   vault.EnclaveData.PublicKey,
			EnclaveId:   vault.EnclaveData.EnclaveId,
			Version:     vault.EnclaveData.Version,
		}
	}

	return typeVault
}

// ConvertAPIEncryptionMetadataToType converts from API EncryptionMetadata to types.EncryptionMetadata
func ConvertAPIEncryptionMetadataToType(metadata *apiv1.EncryptionMetadata) *EncryptionMetadata {
	if metadata == nil {
		return nil
	}

	return &EncryptionMetadata{
		Algorithm:        metadata.Algorithm,
		ConsensusInput:   metadata.ConsensusInput,
		Nonce:            metadata.Nonce,
		AuthTag:          metadata.AuthTag,
		EncryptionHeight: metadata.EncryptionHeight,
		ValidatorSet:     metadata.ValidatorSet,
		KeyVersion:       metadata.KeyVersion,
		SingleNodeMode:   metadata.SingleNodeMode,
	}
}
