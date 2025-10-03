package types

import (
	"encoding/json"
	"fmt"
)

const (
	// Default max record size: 10MB
	DefaultMaxRecordSize = 10 * 1024 * 1024
	// Default max protocols per DWN: 100
	DefaultMaxProtocolsPerDWN = 100
	// Default max permissions per DWN: 1000
	DefaultMaxPermissionsPerDWN = 1000
	// Default vault creation enabled
	DefaultVaultCreationEnabled = true
	// Default min vault refresh interval: 100 blocks
	DefaultMinVaultRefreshInterval = 100
	// Default encryption enabled
	DefaultEncryptionEnabled = true
	// Default key rotation days: 30
	DefaultKeyRotationDays = 30
	// Default min validators for key generation: 67% of active set
	DefaultMinValidatorsForKeyGen = 67
	// Default single node fallback disabled
	DefaultSingleNodeFallback = false
)

// DefaultEncryptedProtocols are protocols that require encryption by default
var DefaultEncryptedProtocols = []string{
	"vault.enclave/v1", // SECURITY: Enclave data must always be encrypted
	"medical.records/v1",
	"financial.data/v1",
	"private.messages/v1",
}

// DefaultEncryptedSchemas are schemas that require encryption by default
var DefaultEncryptedSchemas = []string{
	"https://schemas.sonr.io/medical/",
	"https://schemas.sonr.io/financial/",
	"https://schemas.sonr.io/personal/",
}

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		MaxRecordSize:           DefaultMaxRecordSize,
		MaxProtocolsPerDwn:      DefaultMaxProtocolsPerDWN,
		MaxPermissionsPerDwn:    DefaultMaxPermissionsPerDWN,
		VaultCreationEnabled:    DefaultVaultCreationEnabled,
		MinVaultRefreshInterval: DefaultMinVaultRefreshInterval,
		EncryptionEnabled:       DefaultEncryptionEnabled,
		KeyRotationDays:         DefaultKeyRotationDays,
		MinValidatorsForKeyGen:  DefaultMinValidatorsForKeyGen,
		EncryptedProtocols:      DefaultEncryptedProtocols,
		EncryptedSchemas:        DefaultEncryptedSchemas,
		SingleNodeFallback:      DefaultSingleNodeFallback,
	}
}

// String method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if err := validateMaxRecordSize(p.MaxRecordSize); err != nil {
		return err
	}
	if err := validateMaxProtocolsPerDWN(p.MaxProtocolsPerDwn); err != nil {
		return err
	}
	if err := validateMaxPermissionsPerDWN(p.MaxPermissionsPerDwn); err != nil {
		return err
	}
	if err := validateMinVaultRefreshInterval(p.MinVaultRefreshInterval); err != nil {
		return err
	}
	if err := validateKeyRotationDays(p.KeyRotationDays); err != nil {
		return err
	}
	if err := validateMinValidatorsForKeyGen(p.MinValidatorsForKeyGen); err != nil {
		return err
	}
	if err := validateEncryptedProtocols(p.EncryptedProtocols); err != nil {
		return err
	}
	if err := validateEncryptedSchemas(p.EncryptedSchemas); err != nil {
		return err
	}
	return nil
}

func validateMaxRecordSize(maxSize uint64) error {
	if maxSize == 0 {
		return fmt.Errorf("max record size must be positive")
	}
	if maxSize > 100*1024*1024 { // 100MB max
		return fmt.Errorf("max record size cannot exceed 100MB")
	}
	return nil
}

func validateMaxProtocolsPerDWN(max uint32) error {
	if max == 0 {
		return fmt.Errorf("max protocols per DWN must be positive")
	}
	if max > 10000 {
		return fmt.Errorf("max protocols per DWN cannot exceed 10000")
	}
	return nil
}

func validateMaxPermissionsPerDWN(max uint32) error {
	if max == 0 {
		return fmt.Errorf("max permissions per DWN must be positive")
	}
	if max > 100000 {
		return fmt.Errorf("max permissions per DWN cannot exceed 100000")
	}
	return nil
}

func validateMinVaultRefreshInterval(interval uint64) error {
	if interval == 0 {
		return fmt.Errorf("min vault refresh interval must be positive")
	}
	if interval > 1000000 {
		return fmt.Errorf("min vault refresh interval cannot exceed 1000000 blocks")
	}
	return nil
}

func validateKeyRotationDays(days uint32) error {
	if days == 0 {
		return fmt.Errorf("key rotation days must be positive")
	}
	if days > 365 {
		return fmt.Errorf("key rotation days cannot exceed 365")
	}
	return nil
}

func validateMinValidatorsForKeyGen(percentage uint32) error {
	if percentage == 0 {
		return fmt.Errorf("min validators percentage must be positive")
	}
	if percentage > 100 {
		return fmt.Errorf("min validators percentage cannot exceed 100")
	}
	return nil
}

func validateEncryptedProtocols(protocols []string) error {
	if len(protocols) > 1000 {
		return fmt.Errorf("cannot have more than 1000 encrypted protocols")
	}
	for _, protocol := range protocols {
		if protocol == "" {
			return fmt.Errorf("encrypted protocol cannot be empty")
		}
		if len(protocol) > 256 {
			return fmt.Errorf("encrypted protocol cannot exceed 256 characters")
		}
	}
	return nil
}

func validateEncryptedSchemas(schemas []string) error {
	if len(schemas) > 1000 {
		return fmt.Errorf("cannot have more than 1000 encrypted schemas")
	}
	for _, schema := range schemas {
		if schema == "" {
			return fmt.Errorf("encrypted schema cannot be empty")
		}
		if len(schema) > 512 {
			return fmt.Errorf("encrypted schema cannot exceed 512 characters")
		}
	}
	return nil
}
