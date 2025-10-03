package types

// VaultMetadata contains metadata about an encrypted vault
type VaultMetadata struct {
	Did          string   `json:"did"`
	VaultId      string   `json:"vault_id"`
	Owner        string   `json:"owner"`
	KeyId        string   `json:"key_id"`
	Algorithm    string   `json:"algorithm"`
	Nonce        string   `json:"nonce"`
	CreatedAt    int64    `json:"created_at"`
	BlockHeight  int64    `json:"block_height"`
	ValidatorSet []string `json:"validator_set"`
}

// EncryptedVaultData represents encrypted vault data stored in IPFS
type EncryptedVaultData struct {
	Metadata      *VaultMetadata `json:"metadata"`
	EncryptedData string         `json:"encrypted_data"`
	Version       int            `json:"version"`
}

// EncryptedVaultState represents the on-chain state of an encrypted vault
type EncryptedVaultState struct {
	VaultId        string `json:"vault_id"`
	Did            string `json:"did"`
	Owner          string `json:"owner"`
	IpfsCid        string `json:"ipfs_cid"`
	PublicKey      string `json:"public_key"`
	CreatedAt      int64  `json:"created_at"`
	LastUpdated    int64  `json:"last_updated"`
	Status         string `json:"status"`
	EncryptionType string `json:"encryption_type"`
}
