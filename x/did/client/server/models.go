package server

import (
	"time"
)

// StoredWebAuthnCredential represents a stored WebAuthn credential in database
type StoredWebAuthnCredential struct {
	ID                uint      `gorm:"primaryKey"`
	CredentialID      string    `gorm:"uniqueIndex;not null"`
	RawID             string    `gorm:"not null"`
	ClientDataJSON    string    `gorm:"type:text;not null"`
	AttestationObject string    `gorm:"type:text;not null"`
	Username          string    `gorm:"index;not null"`
	PublicKey         []byte    `gorm:"type:blob"`
	Algorithm         int32     `gorm:"not null"`
	Origin            string    `gorm:"not null"`
	RPID              string    `gorm:"not null"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

// UnsignedTransaction represents an unsigned transaction waiting to be signed
type UnsignedTransaction struct {
	ID          uint      `gorm:"primaryKey"`
	TxID        string    `gorm:"uniqueIndex;not null"`
	Username    string    `gorm:"index;not null"`
	TxData      []byte    `gorm:"type:blob;not null"` // Serialized transaction data
	TxType      string    `gorm:"not null"`           // e.g., "MsgRegisterWebAuthnCredential", "MsgCreateRecord"
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"not null;default:pending"` // pending, signed, broadcast, failed
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	ExpiresAt   *time.Time
}

// AccountInfo represents DWN wallet account information
type AccountInfo struct {
	ID               uint      `gorm:"primaryKey"`
	Username         string    `gorm:"uniqueIndex;not null"`
	Address          string    `gorm:"uniqueIndex;not null"`
	DID              string    `gorm:"uniqueIndex"`
	PublicKey        []byte    `gorm:"type:blob"`
	EncryptedPrivKey []byte    `gorm:"type:blob"` // Encrypted with user's WebAuthn credential
	KeyType          string    `gorm:"not null"`  // e.g., "secp256k1", "ed25519"
	ChainID          string    `gorm:"not null"`
	AccountNumber    uint64    `gorm:"not null"`
	Sequence         uint64    `gorm:"not null"`
	VaultID          string    `gorm:"index"`
	VaultPublicKey   []byte    `gorm:"type:blob"`
	EnclaveID        string    `gorm:"index"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// VaultInfo represents vault metadata and encryption keys
type VaultInfo struct {
	ID               uint      `gorm:"primaryKey"`
	VaultID          string    `gorm:"uniqueIndex;not null"`
	Username         string    `gorm:"index;not null"`
	EnclaveID        string    `gorm:"uniqueIndex;not null"`
	PublicKey        []byte    `gorm:"type:blob;not null"`
	EncryptedEnclave []byte    `gorm:"type:blob;not null"`      // MPC enclave data encrypted
	IPFSHash         string    `gorm:"index"`                   // IPFS hash for vault data
	Status           string    `gorm:"not null;default:active"` // active, rotated, deprecated
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// SessionInfo represents active WebAuthn sessions
type SessionInfo struct {
	ID          uint      `gorm:"primaryKey"`
	Username    string    `gorm:"index;not null"`
	SessionID   string    `gorm:"uniqueIndex;not null"`
	Challenge   string    `gorm:"not null"`
	SessionType string    `gorm:"not null"`                // registration, authentication
	Status      string    `gorm:"not null;default:active"` // active, completed, expired
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	ExpiresAt   time.Time `gorm:"not null"`
}
