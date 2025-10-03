package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sonr-io/sonr/x/dwn/client/plugin"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// GetWalletTxCommands returns wallet-specific transaction commands
func GetWalletTxCommands() *cobra.Command {
	walletTxCmd := &cobra.Command{
		Use:                        "wallet",
		Short:                      "Wallet transaction commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	walletTxCmd.AddCommand(
		GetCmdWalletExecute(),
		GetCmdWalletSponsor(),
		GetCmdWalletEVM(),
	)

	return walletTxCmd
}

// GetCmdWalletExecute creates a command to execute wallet transactions using UCAN tokens
func GetCmdWalletExecute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute [target-did] [permissions]",
		Short: "Execute wallet transaction using UCAN origin token",
		Long: `Execute a wallet transaction by creating and using a UCAN origin token.
The permissions should be provided as JSON array of capability attenuations.

Example:
  snrd tx dwn wallet execute did:sonr:target123 '[{"can":["sign"],"with":"vault://example"}]' --from alice
  
The command will:
1. Load the Motor plugin with enclave data
2. Create a UCAN origin token with specified permissions
3. Execute the transaction with proper authorization`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			targetDID := args[0]
			permissionsJSON := args[1]

			// Parse permissions
			var permissions []map[string]any
			if parseErr := json.Unmarshal([]byte(permissionsJSON), &permissions); parseErr != nil {
				return fmt.Errorf("failed to parse permissions JSON: %w", parseErr)
			}

			// Get enclave data from flags
			enclaveDataHex, err := cmd.Flags().GetString("enclave-data")
			if err != nil {
				return err
			}

			if enclaveDataHex == "" {
				return fmt.Errorf("enclave-data flag is required")
			}

			enclaveData, err := parseEnclaveData(enclaveDataHex)
			if err != nil {
				return fmt.Errorf("failed to parse enclave data: %w", err)
			}

			// Get optional expiration time
			expiresAt, err := cmd.Flags().GetInt64("expires-at")
			if err != nil {
				return err
			}

			// If no expiration provided, default to 1 hour
			if expiresAt == 0 {
				expiresAt = time.Now().Add(time.Hour).Unix()
			}

			// Create enclave configuration
			chainID := clientCtx.ChainID
			if chainID == "" {
				chainID = DefaultTestChainID
			}
			config := plugin.CreateEnclaveConfig(chainID, enclaveData)

			// Load plugin and create UCAN token
			ctx := context.Background()
			motorPlugin, err := plugin.LoadPluginWithManager(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to load Motor plugin: %w", err)
			}

			// Create UCAN origin token request
			tokenReq := &plugin.NewOriginTokenRequest{
				AudienceDID:  targetDID,
				Attenuations: permissions,
				ExpiresAt:    expiresAt,
			}

			// Create the UCAN token
			tokenResp, err := motorPlugin.NewOriginToken(tokenReq)
			if err != nil {
				return fmt.Errorf("failed to create UCAN token: %w", err)
			}

			if tokenResp.Error != "" {
				return fmt.Errorf("plugin error creating token: %s", tokenResp.Error)
			}

			// Create transaction message
			msg := &types.MsgRecordsWrite{
				Author:        clientCtx.GetFromAddress().String(),
				Target:        targetDID,
				Data:          []byte(fmt.Sprintf("UCAN Token Execution: %s", tokenResp.Token)),
				Authorization: tokenResp.Token,
			}

			// Validate message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)

	// Add wallet-specific flags
	cmd.Flags().String("enclave-data", "", "Hex-encoded enclave data for wallet operations")
	cmd.Flags().
		Int64("expires-at", 0, "UCAN token expiration timestamp (defaults to 1 hour from now)")

	// Mark required flags
	if err := cmd.MarkFlagRequired("enclave-data"); err != nil {
		panic(err)
	}

	return cmd
}

// GetCmdWalletSponsor creates a command to sponsor wallets with UCAN delegation
func GetCmdWalletSponsor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponsor [wallet-address] [amount]",
		Short: "Sponsor a wallet with UCAN delegation token",
		Long: `Sponsor a wallet by creating a delegated UCAN token with spending permissions.
The amount should be specified in the base denomination (usnr for staking, snr for transfers).

Example:
  snrd tx dwn wallet sponsor sonr1abc123... 1000000usnr --from alice
  
This creates an attenuated UCAN token that allows the sponsored wallet to spend
up to the specified amount on behalf of the sponsor.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			walletAddress := args[0]
			amountStr := args[1]

			// Parse amount
			amount, err := strconv.ParseInt(amountStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse amount: %w", err)
			}

			// Get parent token and enclave data from flags
			parentToken, err := cmd.Flags().GetString("parent-token")
			if err != nil {
				return err
			}

			enclaveDataHex, err := cmd.Flags().GetString("enclave-data")
			if err != nil {
				return err
			}

			if enclaveDataHex == "" {
				return fmt.Errorf("enclave-data flag is required")
			}

			if parentToken == "" {
				return fmt.Errorf("parent-token flag is required for delegation")
			}

			enclaveData, err := parseEnclaveData(enclaveDataHex)
			if err != nil {
				return fmt.Errorf("failed to parse enclave data: %w", err)
			}

			// Get optional expiration time
			expiresAt, err := cmd.Flags().GetInt64("expires-at")
			if err != nil {
				return err
			}

			// If no expiration provided, default to 24 hours
			if expiresAt == 0 {
				expiresAt = time.Now().Add(24 * time.Hour).Unix()
			}

			// Create enclave configuration
			chainID := clientCtx.ChainID
			if chainID == "" {
				chainID = DefaultTestChainID
			}
			config := plugin.CreateEnclaveConfig(chainID, enclaveData)

			// Load plugin and create attenuated UCAN token
			ctx := context.Background()
			motorPlugin, err := plugin.LoadPluginWithManager(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to load Motor plugin: %w", err)
			}

			// Create attenuated UCAN token with spending limits
			attenuations := []map[string]any{
				{
					"can":  []string{"spend"},
					"with": walletAddress,
					"nb":   map[string]any{"max_amount": amount},
				},
			}

			tokenReq := &plugin.NewAttenuatedTokenRequest{
				ParentToken:  parentToken,
				AudienceDID:  walletAddress,
				Attenuations: attenuations,
				ExpiresAt:    expiresAt,
			}

			// Create the delegated UCAN token
			tokenResp, err := motorPlugin.NewAttenuatedToken(tokenReq)
			if err != nil {
				return fmt.Errorf("failed to create attenuated UCAN token: %w", err)
			}

			if tokenResp.Error != "" {
				return fmt.Errorf("plugin error creating token: %s", tokenResp.Error)
			}

			// Create sponsorship message
			sponsorshipData := map[string]any{
				"type":             "wallet_sponsorship",
				"sponsored_wallet": walletAddress,
				"max_amount":       amount,
				"sponsor":          clientCtx.GetFromAddress().String(),
				"ucan_token":       tokenResp.Token,
			}

			dataBytes, err := json.Marshal(sponsorshipData)
			if err != nil {
				return fmt.Errorf("failed to marshal sponsorship data: %w", err)
			}

			// Create transaction message
			msg := &types.MsgRecordsWrite{
				Author:        clientCtx.GetFromAddress().String(),
				Target:        walletAddress,
				Data:          dataBytes,
				Authorization: tokenResp.Token,
			}

			// Validate message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)

	// Add wallet-specific flags
	cmd.Flags().String("enclave-data", "", "Hex-encoded enclave data for wallet operations")
	cmd.Flags().String("parent-token", "", "Parent UCAN token to delegate from")
	cmd.Flags().
		Int64("expires-at", 0, "UCAN token expiration timestamp (defaults to 24 hours from now)")

	// Mark required flags
	if err := cmd.MarkFlagRequired("enclave-data"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("parent-token"); err != nil {
		panic(err)
	}

	return cmd
}

// GetCmdWalletEVM creates a command for EVM transaction execution
func GetCmdWalletEVM() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm [to-address] [data]",
		Short: "Execute EVM transaction using Motor plugin signing",
		Long: `Execute an EVM transaction using the Motor plugin for signing.
The transaction data should be provided as hex-encoded bytes.

Example:
  snrd tx dwn wallet evm 0x742d35Cc6e71cbC... 0xa9059cbb --from alice
  
This signs the EVM transaction using MPC-based signing in the Motor plugin
and submits it through the DWN module.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddress := args[0]
			evmData := args[1]

			// Get enclave data from flags
			enclaveDataHex, err := cmd.Flags().GetString("enclave-data")
			if err != nil {
				return err
			}

			if enclaveDataHex == "" {
				return fmt.Errorf("enclave-data flag is required")
			}

			enclaveData, err := parseEnclaveData(enclaveDataHex)
			if err != nil {
				return fmt.Errorf("failed to parse enclave data: %w", err)
			}

			// Get optional gas limit and gas price
			gasLimit, err := cmd.Flags().GetUint64("gas-limit")
			if err != nil {
				return err
			}

			gasPrice, err := cmd.Flags().GetString("gas-price")
			if err != nil {
				return err
			}

			// Create EVM transaction data
			evmTxData := map[string]any{
				"type":      "evm_transaction",
				"to":        toAddress,
				"data":      evmData,
				"gas_limit": gasLimit,
				"gas_price": gasPrice,
				"from":      clientCtx.GetFromAddress().String(),
			}

			dataBytes, err := json.Marshal(evmTxData)
			if err != nil {
				return fmt.Errorf("failed to marshal EVM transaction data: %w", err)
			}

			// Create enclave configuration
			chainID := clientCtx.ChainID
			if chainID == "" {
				chainID = DefaultTestChainID
			}
			config := plugin.CreateEnclaveConfig(chainID, enclaveData)

			// Load plugin and sign the transaction data
			ctx := context.Background()
			motorPlugin, err := plugin.LoadPluginWithManager(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to load Motor plugin: %w", err)
			}

			// Sign the EVM transaction data
			signReq := &plugin.SignDataRequest{
				Data: dataBytes,
			}

			signResp, err := motorPlugin.SignData(signReq)
			if err != nil {
				return fmt.Errorf("failed to sign EVM transaction: %w", err)
			}

			if signResp.Error != "" {
				return fmt.Errorf("plugin error signing data: %s", signResp.Error)
			}

			// Create DWN message with signed EVM transaction
			msg := &types.MsgRecordsWrite{
				Author:        clientCtx.GetFromAddress().String(),
				Target:        toAddress,
				Data:          dataBytes,
				Authorization: string(signResp.Signature),
			}

			// Validate message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	flags.AddTxFlagsToCmd(cmd)

	// Add EVM-specific flags
	cmd.Flags().String("enclave-data", "", "Hex-encoded enclave data for wallet operations")
	cmd.Flags().Uint64("gas-limit", 21000, "Gas limit for EVM transaction")
	cmd.Flags().String("gas-price", "1000000000", "Gas price in wei for EVM transaction")

	// Mark required flags
	if err := cmd.MarkFlagRequired("enclave-data"); err != nil {
		panic(err)
	}

	return cmd
}
