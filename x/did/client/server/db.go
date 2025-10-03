package server

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB initializes the SQLite database connection
func InitDB() error {
	// Create ~/.sonr directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	sonrDir := filepath.Join(homeDir, ".sonr")
	if mkdirErr := os.MkdirAll(sonrDir, 0o750); mkdirErr != nil {
		return fmt.Errorf("failed to create .sonr directory: %w", mkdirErr)
	}

	// Database file path
	dbPath := filepath.Join(sonrDir, "vault.db")

	// Open SQLite database with GORM
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// Disable GORM logging for cleaner CLI output
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&StoredWebAuthnCredential{},
		&UnsignedTransaction{},
		&AccountInfo{},
		&VaultInfo{},
		&SessionInfo{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// WebAuthnCredentialService provides database operations for WebAuthn credentials
type WebAuthnCredentialService struct{}

// NewWebAuthnCredentialService creates a new WebAuthn credential service
func NewWebAuthnCredentialService() *WebAuthnCredentialService {
	return &WebAuthnCredentialService{}
}

// Store saves a WebAuthn credential to the database
func (s *WebAuthnCredentialService) Store(credential *StoredWebAuthnCredential) error {
	return db.Create(credential).Error
}

// GetByCredentialID retrieves a credential by its ID
func (s *WebAuthnCredentialService) GetByCredentialID(
	credentialID string,
) (*StoredWebAuthnCredential, error) {
	var credential StoredWebAuthnCredential
	err := db.Where("credential_id = ?", credentialID).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByUsername retrieves all credentials for a username
func (s *WebAuthnCredentialService) GetByUsername(
	username string,
) ([]StoredWebAuthnCredential, error) {
	var credentials []StoredWebAuthnCredential
	err := db.Where("username = ?", username).Find(&credentials).Error
	return credentials, err
}

// UsernameExists checks if a username already has registered WebAuthn credentials
func (s *WebAuthnCredentialService) UsernameExists(username string) (bool, error) {
	var count int64
	err := db.Model(&StoredWebAuthnCredential{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AccountInfoService provides database operations for account information
type AccountInfoService struct{}

// NewAccountInfoService creates a new account info service
func NewAccountInfoService() *AccountInfoService {
	return &AccountInfoService{}
}

// Store saves account information to the database
func (s *AccountInfoService) Store(account *AccountInfo) error {
	return db.Create(account).Error
}

// GetByUsername retrieves account info by username
func (s *AccountInfoService) GetByUsername(username string) (*AccountInfo, error) {
	var account AccountInfo
	err := db.Where("username = ?", username).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateSequence updates the account sequence number
func (s *AccountInfoService) UpdateSequence(username string, sequence uint64) error {
	return db.Model(&AccountInfo{}).
		Where("username = ?", username).
		Update("sequence", sequence).
		Error
}

// VaultInfoService provides database operations for vault information
type VaultInfoService struct{}

// NewVaultInfoService creates a new vault info service
func NewVaultInfoService() *VaultInfoService {
	return &VaultInfoService{}
}

// Store saves vault information to the database
func (s *VaultInfoService) Store(vault *VaultInfo) error {
	return db.Create(vault).Error
}

// GetByVaultID retrieves vault info by vault ID
func (s *VaultInfoService) GetByVaultID(vaultID string) (*VaultInfo, error) {
	var vault VaultInfo
	err := db.Where("vault_id = ?", vaultID).First(&vault).Error
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

// GetByUsername retrieves all vaults for a username
func (s *VaultInfoService) GetByUsername(username string) ([]VaultInfo, error) {
	var vaults []VaultInfo
	err := db.Where("username = ?", username).Find(&vaults).Error
	return vaults, err
}

// UnsignedTransactionService provides database operations for unsigned transactions
type UnsignedTransactionService struct{}

// NewUnsignedTransactionService creates a new unsigned transaction service
func NewUnsignedTransactionService() *UnsignedTransactionService {
	return &UnsignedTransactionService{}
}

// Store saves an unsigned transaction to the database
func (s *UnsignedTransactionService) Store(tx *UnsignedTransaction) error {
	return db.Create(tx).Error
}

// GetByTxID retrieves a transaction by its ID
func (s *UnsignedTransactionService) GetByTxID(txID string) (*UnsignedTransaction, error) {
	var tx UnsignedTransaction
	err := db.Where("tx_id = ?", txID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetPendingByUsername retrieves all pending transactions for a username
func (s *UnsignedTransactionService) GetPendingByUsername(
	username string,
) ([]UnsignedTransaction, error) {
	var transactions []UnsignedTransaction
	err := db.Where("username = ? AND status = ?", username, "pending").Find(&transactions).Error
	return transactions, err
}

// UpdateStatus updates the transaction status
func (s *UnsignedTransactionService) UpdateStatus(txID, status string) error {
	return db.Model(&UnsignedTransaction{}).Where("tx_id = ?", txID).Update("status", status).Error
}

// SessionInfoService provides database operations for session information
type SessionInfoService struct{}

// NewSessionInfoService creates a new session info service
func NewSessionInfoService() *SessionInfoService {
	return &SessionInfoService{}
}

// Store saves session information to the database
func (s *SessionInfoService) Store(session *SessionInfo) error {
	return db.Create(session).Error
}

// GetBySessionID retrieves a session by its ID
func (s *SessionInfoService) GetBySessionID(sessionID string) (*SessionInfo, error) {
	var session SessionInfo
	err := db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateStatus updates the session status
func (s *SessionInfoService) UpdateStatus(sessionID, status string) error {
	return db.Model(&SessionInfo{}).
		Where("session_id = ?", sessionID).
		Update("status", status).
		Error
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionInfoService) CleanupExpiredSessions() error {
	return db.Where("expires_at < ?", fmt.Sprintf("%d", os.Getpid())).Delete(&SessionInfo{}).Error
}
