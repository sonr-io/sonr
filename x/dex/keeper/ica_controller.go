package keeper

import (
	"fmt"
	"time"

	"github.com/sonr-io/sonr/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// RegisterDEXAccount registers a new ICA account for DEX operations
func (k Keeper) RegisterDEXAccount(
	ctx sdk.Context,
	did string,
	connectionID string,
	features []string,
) (*types.InterchainDEXAccount, error) {
	// Validate inputs
	if did == "" {
		return nil, fmt.Errorf("DID cannot be empty")
	}
	if connectionID == "" {
		return nil, fmt.Errorf("connection ID cannot be empty")
	}

	// Validate DID exists by trying to get the document
	if _, err := k.didKeeper.GetDIDDocument(ctx, did); err != nil {
		return nil, fmt.Errorf("DID %s does not exist: %w", did, err)
	}

	// Check if account already exists
	accountKey := GetAccountKey(did, connectionID)
	existing, err := k.Accounts.Get(ctx, accountKey)
	if err == nil {
		// Return existing account regardless of status (idempotent)
		return &existing, nil
	}

	// Generate unique port ID
	portID := GetPortID(did, connectionID)

	// Register ICA account
	if err := k.icaControllerKeeper.RegisterInterchainAccount(
		ctx,
		connectionID,
		portID,
		"", // Use default version
	); err != nil {
		return nil, fmt.Errorf("failed to register ICA account: %w", err)
	}

	// Create DEX account record
	account := types.InterchainDEXAccount{
		Did:             did,
		ConnectionId:    connectionID,
		PortId:          portID,
		EnabledFeatures: features,
		Status:          types.ACCOUNT_STATUS_PENDING,
		CreatedAt:       ctx.BlockTime(),
	}

	// Store account
	if err := k.Accounts.Set(ctx, accountKey, account); err != nil {
		return nil, fmt.Errorf("failed to store DEX account: %w", err)
	}

	// Update DID mappings
	if err := k.addDIDMapping(ctx, did, connectionID); err != nil {
		return nil, fmt.Errorf("failed to update DID mappings: %w", err)
	}

	return &account, nil
}

// GetDEXAccount retrieves a DEX account by DID and connection
func (k Keeper) GetDEXAccount(
	ctx sdk.Context,
	did, connectionID string,
) (*types.InterchainDEXAccount, error) {
	accountKey := GetAccountKey(did, connectionID)
	account, err := k.Accounts.Get(ctx, accountKey)
	if err != nil {
		return nil, fmt.Errorf("DEX account not found: %w", err)
	}
	return &account, nil
}

// GetDEXAccountsByDID retrieves all DEX accounts for a DID
func (k Keeper) GetDEXAccountsByDID(
	ctx sdk.Context,
	did string,
) ([]types.InterchainDEXAccount, error) {
	didAccounts, err := k.DIDToAccounts.Get(ctx, did)
	if err != nil {
		return nil, nil // No accounts for this DID
	}

	var accounts []types.InterchainDEXAccount
	for _, connID := range didAccounts.Accounts {
		account, err := k.GetDEXAccount(ctx, did, connID)
		if err == nil {
			accounts = append(accounts, *account)
		}
	}
	return accounts, nil
}

// SendDEXTransaction sends a transaction through ICA
func (k Keeper) SendDEXTransaction(
	ctx sdk.Context,
	did string,
	connectionID string,
	msgs []sdk.Msg,
	memo string,
	timeoutDuration time.Duration,
) (uint64, error) {
	// Get DEX account
	account, err := k.GetDEXAccount(ctx, did, connectionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get DEX account: %w", err)
	}

	if account.Status != types.ACCOUNT_STATUS_ACTIVE {
		return 0, fmt.Errorf("DEX account is not active")
	}

	// Get ICA address
	icaAddress, found := k.icaControllerKeeper.GetInterchainAccountAddress(
		ctx,
		connectionID,
		account.PortId,
	)
	if !found {
		return 0, fmt.Errorf("ICA address not found")
	}

	// Get channel capability
	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, connectionID, account.PortId)
	if !found {
		return 0, fmt.Errorf("active channel not found")
	}

	chanCap, ok := k.ScopedKeeper.GetCapability(
		ctx,
		host.ChannelCapabilityPath(account.PortId, channelID),
	)
	if !ok {
		return 0, fmt.Errorf("channel capability not found")
	}

	// Encode messages
	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs, icatypes.EncodingProtobuf)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Create packet data
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: memo,
	}

	// Calculate timeout
	timeoutTimestamp := ctx.BlockTime().Add(timeoutDuration).UnixNano()

	// Send transaction
	sequence, err := k.icaControllerKeeper.SendTx(
		ctx,
		chanCap,
		connectionID,
		account.PortId,
		packetData,
		uint64(timeoutTimestamp),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send ICA transaction: %w", err)
	}

	// Log transaction
	k.Logger(ctx).Info("DEX transaction sent",
		"did", did,
		"connection", connectionID,
		"ica_address", icaAddress,
		"sequence", sequence,
	)

	return sequence, nil
}

// OnICAAccountCreated handles successful ICA account creation
func (k Keeper) OnICAAccountCreated(ctx sdk.Context, portID, address string) error {
	// Find account by port ID
	var account *types.InterchainDEXAccount
	k.Accounts.Walk(ctx, nil, func(key string, value types.InterchainDEXAccount) (bool, error) {
		if value.PortId == portID {
			account = &value
			return true, nil
		}
		return false, nil
	})

	if account == nil {
		return fmt.Errorf("DEX account not found for port %s", portID)
	}

	// Update account status and address
	account.Status = types.ACCOUNT_STATUS_ACTIVE
	account.AccountAddress = address
	account.HostChainId = k.getHostChainID(ctx, account.ConnectionId)

	// Store updated account
	accountKey := GetAccountKey(account.Did, account.ConnectionId)
	if err := k.Accounts.Set(ctx, accountKey, *account); err != nil {
		return fmt.Errorf("failed to update DEX account: %w", err)
	}

	return nil
}

// Helper functions

func (k Keeper) addDIDMapping(ctx sdk.Context, did, connectionID string) error {
	didAccounts, _ := k.DIDToAccounts.Get(ctx, did)

	// Check if already exists
	for _, conn := range didAccounts.Accounts {
		if conn == connectionID {
			return nil
		}
	}

	didAccounts.Accounts = append(didAccounts.Accounts, connectionID)
	return k.DIDToAccounts.Set(ctx, did, didAccounts)
}

func (k Keeper) getHostChainID(ctx sdk.Context, connectionID string) string {
	conn, found := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return ""
	}
	// Extract chain ID from connection counterparty
	// This is a simplified version - actual implementation may vary
	return conn.Counterparty.ClientId
}
