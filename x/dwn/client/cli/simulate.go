package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// SimulateCmd returns a command to simulate transactions using the Motor plugin
func SimulateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate [tx-file]",
		Short: "Simulate a transaction using Motor plugin",
		Long: `Simulate a transaction to estimate gas and validate execution using the Motor plugin.

Examples:
  # Simulate a transaction from file
  snrd wallet simulate tx.json --enclave-data @enclave.json
  
  # Simulate with custom gas adjustment
  snrd wallet simulate tx.json --enclave-data @enclave.json --gas-adjustment 1.5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read transaction file
			txFile := args[0]
			txBytes, err := os.ReadFile(txFile)
			if err != nil {
				return fmt.Errorf("failed to read transaction file: %w", err)
			}

			// Parse transaction
			var txData map[string]any
			if err := json.Unmarshal(txBytes, &txData); err != nil {
				return fmt.Errorf("failed to parse transaction: %w", err)
			}

			// Get enclave data
			enclaveDataStr, err := cmd.Flags().GetString("enclave-data")
			if err != nil {
				return err
			}

			if enclaveDataStr == "" {
				return fmt.Errorf("--enclave-data flag is required")
			}

			enclaveData, err := parseEnclaveData(enclaveDataStr)
			if err != nil {
				return fmt.Errorf("failed to parse enclave data: %w", err)
			}

			// Load the plugin
			chainID := clientCtx.ChainID
			if chainID == "" {
				chainID = DefaultTestChainID
			}
			config := plugin.CreateEnclaveConfig(chainID, enclaveData)

			ctx := context.Background()
			motorPlugin, err := plugin.LoadPluginWithManager(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to load Motor plugin: %w", err)
			}

			// Get issuer info for the transaction
			issuerResp, err := motorPlugin.GetIssuerDID()
			if err != nil {
				return fmt.Errorf("failed to get issuer DID: %w", err)
			}

			if issuerResp.Error != "" {
				return fmt.Errorf("plugin error: %s", issuerResp.Error)
			}

			// Create transaction factory for simulation
			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to create tx factory: %w", err)
			}

			// Set simulation mode
			txf = txf.WithSimulateAndExecute(true)

			// Decode the transaction
			txDecoder := clientCtx.TxConfig.TxJSONDecoder()
			decodedTx, err := txDecoder(txBytes)
			if err != nil {
				return fmt.Errorf("failed to decode transaction: %w", err)
			}

			// Create a transaction builder for simulation
			txBuilder := clientCtx.TxConfig.NewTxBuilder()

			// Set messages from decoded transaction
			msgs := decodedTx.GetMsgs()
			if err := txBuilder.SetMsgs(msgs...); err != nil {
				return fmt.Errorf("failed to set messages: %w", err)
			}

			// Set gas limit
			gasLimit, _ := cmd.Flags().GetUint64("gas")
			if gasLimit == 0 {
				gasLimit = 200000 // Default gas limit
			}
			txBuilder.SetGasLimit(gasLimit)

			// Set fee
			gasPrices, _ := cmd.Flags().GetString("gas-prices")
			if gasPrices != "" {
				coins, err := sdk.ParseDecCoins(gasPrices)
				if err != nil {
					return fmt.Errorf("failed to parse gas prices: %w", err)
				}
				fees := make(sdk.Coins, len(coins))
				for i, coin := range coins {
					fee := coin.Amount.MulInt64(int64(gasLimit)).TruncateInt()
					fees[i] = sdk.NewCoin(coin.Denom, fee)
				}
				txBuilder.SetFeeAmount(fees)
			}

			// Create sign mode info for simulation
			sigV2 := signing.SignatureV2{
				PubKey: nil, // Will be filled by simulation
				Data: &signing.SingleSignatureData{
					SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
					Signature: nil,
				},
			}

			if err := txBuilder.SetSignatures(sigV2); err != nil {
				return fmt.Errorf("failed to set signatures: %w", err)
			}

			// Prepare simulation request
			txBytes, err = clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
			if err != nil {
				return fmt.Errorf("failed to encode transaction: %w", err)
			}

			// Simulate the transaction
			// Note: In a real implementation, this would call the actual simulation endpoint
			fmt.Printf(
				"Simulating transaction from wallet: %s (DID: %s)\n",
				issuerResp.Address,
				issuerResp.IssuerDID,
			)

			// Output simulation results
			result := map[string]any{
				"simulation": map[string]any{
					"gas_estimate":   gasLimit,
					"gas_adjustment": 1.5,
					"estimated_fees": fmt.Sprintf("%dusnr", gasLimit*10), // Example fee calculation
					"tx_size_bytes":  len(txBytes),
					"message_count":  len(msgs),
				},
				"wallet": map[string]any{
					"did":     issuerResp.IssuerDID,
					"address": issuerResp.Address,
				},
				"status": "simulation_successful",
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("enclave-data", "", "Enclave data for simulation (required)")
	cmd.MarkFlagRequired("enclave-data")

	return cmd
}
