package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// BroadcastCmd returns a command to broadcast a signed transaction using the Motor plugin
func BroadcastCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast [signed-tx-file]",
		Short: "Broadcast a signed transaction using Motor plugin",
		Long: `Broadcast a signed transaction to the network using the Motor plugin.
The transaction should be provided as a file containing the signed transaction.

Example:
  snrd wallet broadcast signed_tx.json --enclave-data @enclave.json

This command supports broadcasting transactions that were previously signed using the Motor plugin.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read the signed transaction file
			txFile := args[0]
			txBytes, err := os.ReadFile(txFile)
			if err != nil {
				return fmt.Errorf("failed to read transaction file: %w", err)
			}

			// Get optional enclave data for verification
			enclaveDataStr, err := cmd.Flags().GetString("enclave-data")
			if err != nil {
				return err
			}

			if enclaveDataStr != "" {
				// Parse enclave data
				enclaveData, err := parseEnclaveData(enclaveDataStr)
				if err != nil {
					return fmt.Errorf("failed to parse enclave data: %w", err)
				}

				// Load the plugin for verification
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

				// Get issuer DID for verification
				resp, err := motorPlugin.GetIssuerDID()
				if err != nil {
					return fmt.Errorf("failed to get issuer DID: %w", err)
				}

				if resp.Error != "" {
					return fmt.Errorf("plugin error: %s", resp.Error)
				}

				fmt.Printf(
					"Broadcasting transaction from wallet: %s (DID: %s)\n",
					resp.Address,
					resp.IssuerDID,
				)
			}

			// Broadcast the transaction
			res, err := clientCtx.BroadcastTxSync(txBytes)
			if err != nil {
				return fmt.Errorf("failed to broadcast transaction: %w", err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("enclave-data", "", "Enclave data for wallet verification (hex or @file)")
	return cmd
}
