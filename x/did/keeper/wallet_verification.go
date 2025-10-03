package keeper

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/x/did/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// VerifyWalletOwnership verifies that the provided signature proves ownership of the wallet
func (k Keeper) VerifyWalletOwnership(
	ctx context.Context,
	walletAddress, chainID string,
	walletType types.WalletType,
	challenge, signature []byte,
) error {
	switch walletType {
	case types.WalletTypeEthereum:
		return k.verifyEthereumSignature(walletAddress, challenge, signature)
	case types.WalletTypeCosmos:
		return k.verifyCosmosSignature(ctx, walletAddress, challenge, signature)
	default:
		return errors.Wrapf(types.ErrUnsupportedWalletType, "wallet type: %s", walletType)
	}
}

// verifyEthereumSignature verifies an Ethereum signature using ECDSA recovery
func (k Keeper) verifyEthereumSignature(walletAddress string, challenge, signature []byte) error {
	// Validate Ethereum address format
	if !common.IsHexAddress(walletAddress) {
		return errors.Wrap(types.ErrInvalidEthereumAddress, "invalid address format")
	}

	// Convert address to common.Address
	expectedAddr := common.HexToAddress(walletAddress)

	// Ethereum uses personal_sign which prefixes the message
	// The format is: "\x19Ethereum Signed Message:\n" + len(message) + message
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(challenge), challenge)
	messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))

	// Recover the public key from the signature
	// Ethereum signatures have a recovery parameter v at the end
	if len(signature) != 65 {
		return errors.Wrap(types.ErrWalletSignatureVerificationFailed, "invalid signature length")
	}

	// The recovery parameter needs to be adjusted for Ethereum
	if signature[64] >= 27 {
		signature[64] -= 27
	}

	publicKeyECDSA, err := crypto.SigToPub(messageHash.Bytes(), signature)
	if err != nil {
		return errors.Wrap(
			types.ErrWalletSignatureVerificationFailed,
			"failed to recover public key",
		)
	}

	// Get the address from the recovered public key
	recoveredAddr := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Compare addresses
	if recoveredAddr != expectedAddr {
		return errors.Wrapf(
			types.ErrWalletSignatureVerificationFailed,
			"signature verification failed: expected %s, got %s",
			expectedAddr.Hex(),
			recoveredAddr.Hex(),
		)
	}

	return nil
}

// verifyCosmosSignature verifies a Cosmos signature using secp256k1
func (k Keeper) verifyCosmosSignature(
	ctx context.Context,
	walletAddress string,
	challenge, signature []byte,
) error {
	// Parse bech32 address to get account address
	accAddr, err := sdk.AccAddressFromBech32(walletAddress)
	if err != nil {
		return errors.Wrapf(
			types.ErrInvalidCosmosAddress,
			"failed to parse bech32 address: %v",
			err,
		)
	}

	// Basic signature validation - Cosmos secp256k1 signatures are 64 bytes
	if len(signature) != 64 {
		return errors.Wrap(
			types.ErrWalletSignatureVerificationFailed,
			"invalid signature length for Cosmos (expected 64 bytes)",
		)
	}

	// Retrieve account from chain state using AccountKeeper
	account := k.accountKeeper.GetAccount(ctx, accAddr)
	if account == nil {
		return errors.Wrapf(
			types.ErrWalletSignatureVerificationFailed,
			"account not found for address: %s",
			walletAddress,
		)
	}

	// Extract public key from account
	pubKey := account.GetPubKey()
	if pubKey == nil {
		return errors.Wrapf(
			types.ErrWalletSignatureVerificationFailed,
			"no public key found for account: %s",
			walletAddress,
		)
	}

	// Ensure the public key is secp256k1
	secp256k1PubKey, ok := pubKey.(*secp256k1.PubKey)
	if !ok {
		return errors.Wrapf(
			types.ErrWalletSignatureVerificationFailed,
			"account public key is not secp256k1: %T",
			pubKey,
		)
	}

	// Verify signature against challenge using secp256k1
	if !secp256k1PubKey.VerifySignature(challenge, signature) {
		return errors.Wrapf(
			types.ErrWalletSignatureVerificationFailed,
			"signature verification failed for address: %s",
			walletAddress,
		)
	}

	return nil
}

// CreateVerificationMethodFromWallet creates a W3C verification method for an external wallet
func (k Keeper) CreateVerificationMethodFromWallet(
	methodID, controllerDID, walletAddress, chainID string,
	walletType types.WalletType,
) (*types.VerificationMethod, error) {
	// Validate wallet type
	if err := walletType.Validate(); err != nil {
		return nil, err
	}

	// Create blockchain account ID
	accountID := types.BlockchainAccountID{
		Namespace: walletType.GetNamespace(),
		ChainID:   chainID,
		Address:   walletAddress,
	}

	if err := accountID.Validate(); err != nil {
		return nil, err
	}

	// Create verification method
	verificationMethod := &types.VerificationMethod{
		Id:                     methodID,
		VerificationMethodKind: walletType.ToVerificationMethodType(),
		Controller:             controllerDID,
		BlockchainAccountId:    accountID.String(),
	}

	return verificationMethod, nil
}

// CheckWalletNotAlreadyLinked checks if a wallet is already linked to any DID
// by querying all DID documents and examining their verification methods for
// matching blockchain account IDs. Returns ErrWalletAlreadyLinked if found.
func (k Keeper) CheckWalletNotAlreadyLinked(
	ctx any,
	walletAddress, chainID string,
	walletType types.WalletType,
) error {
	// Convert context to SDK context for logging
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		return errors.Wrap(types.ErrInvalidRequest, "invalid context type")
	}

	// Create the blockchain account ID we're looking for
	accountID := types.BlockchainAccountID{
		Namespace: walletType.GetNamespace(),
		ChainID:   chainID,
		Address:   walletAddress,
	}

	// Validate the account ID format before searching
	if err := accountID.Validate(); err != nil {
		return errors.Wrap(types.ErrInvalidBlockchainAccountID, err.Error())
	}

	targetAccountID := accountID.String()

	k.logger.Debug("Checking wallet duplication",
		"wallet_address", walletAddress,
		"chain_id", chainID,
		"wallet_type", walletType,
		"target_account_id", targetAccountID,
	)

	// Use ORM iterator to efficiently scan all DID documents
	iterator, err := k.OrmDB.DIDDocumentTable().List(sdkCtx, &apiv1.DIDDocumentPrimaryKey{})
	if err != nil {
		k.logger.Error("Failed to list DID documents for wallet duplication check", "error", err)
		return errors.Wrap(types.ErrFailedToCheckDIDExists, err.Error())
	}
	defer iterator.Close()

	// Iterate through all DID documents to check verification methods
	for iterator.Next() {
		ormDoc, err := iterator.Value()
		if err != nil {
			k.logger.Error(
				"Failed to get DID document during wallet duplication check",
				"error",
				err,
			)
			continue
		}

		// Skip deactivated DID documents as their verification methods are no longer active
		if ormDoc.Deactivated {
			continue
		}

		// Convert from ORM type to access verification methods
		didDoc := types.DIDDocumentFromORM(ormDoc)

		// Check all verification methods for matching blockchain account ID
		for _, vm := range didDoc.VerificationMethod {
			// Skip verification methods without blockchain account IDs
			if vm.BlockchainAccountId == "" {
				continue
			}

			// Check for exact match with the wallet we're trying to link
			if vm.BlockchainAccountId == targetAccountID {
				k.logger.Info("Found duplicate wallet link",
					"wallet_address", walletAddress,
					"chain_id", chainID,
					"wallet_type", walletType,
					"existing_did", didDoc.Id,
					"verification_method_id", vm.Id,
				)

				return errors.Wrapf(
					types.ErrWalletAlreadyLinked,
					"wallet %s on chain %s (%s) is already linked to DID %s in verification method %s",
					walletAddress,
					chainID,
					walletType,
					didDoc.Id,
					vm.Id,
				)
			}
		}
	}

	k.logger.Debug("Wallet is not linked to any existing DID",
		"wallet_address", walletAddress,
		"chain_id", chainID,
		"wallet_type", walletType,
	)

	return nil
}

// ValidateDWNVaultController validates that the DID has an active DWN vault controller
func (k Keeper) ValidateDWNVaultController(ctx any, did string) error {
	// This would check if the DID has an active DWN vault controller
	// For now, we'll implement a basic check

	// In a complete implementation, this would:
	// 1. Query the DWN module to check if the DID has an active vault
	// 2. Verify the vault is properly configured
	// 3. Ensure the vault can sign transactions

	// For now, we'll assume all DIDs are valid if they exist
	return nil
}

// GenerateWalletChallenge generates a challenge message for wallet ownership proof
func (k Keeper) GenerateWalletChallenge(did, walletAddress string, blockHeight int64) []byte {
	challengeMsg := fmt.Sprintf(
		"Link wallet %s to DID %s at block %d. This proves ownership of the wallet.",
		walletAddress, did, blockHeight,
	)
	return []byte(challengeMsg)
}

// Helper functions for signature verification

// recoverEthereumPublicKey recovers the public key from an Ethereum signature
func recoverEthereumPublicKey(message, signature []byte) (*ecdsa.PublicKey, error) {
	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length")
	}

	// Adjust recovery parameter
	if signature[64] >= 27 {
		signature[64] -= 27
	}

	hash := crypto.Keccak256Hash(message)
	return crypto.SigToPub(hash.Bytes(), signature)
}

// verifySecp256k1Signature verifies a secp256k1 signature for Cosmos
func verifySecp256k1Signature(pubKey cryptotypes.PubKey, message, signature []byte) bool {
	secp256k1PubKey, ok := pubKey.(*secp256k1.PubKey)
	if !ok {
		return false
	}

	return secp256k1PubKey.VerifySignature(message, signature)
}
