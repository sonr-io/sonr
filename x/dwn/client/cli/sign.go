package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// SignCmd returns a command to sign data or transactions using the Motor plugin
func SignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [message-or-file]",
		Short: "Sign a message or transaction using Motor plugin",
		Long: `Sign a message or transaction using the Motor plugin's MPC-based signing.

Examples:
  # Sign a text message
  snrd wallet sign "Hello World" --enclave-data @enclave.json
  
  # Sign a transaction file
  snrd wallet sign @unsigned_tx.json --enclave-data @enclave.json --tx
  
  # Sign raw bytes (hex encoded)
  snrd wallet sign 0xdeadbeef --enclave-data @enclave.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
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

			// Determine what to sign
			input := args[0]
			var dataToSign []byte

			isTransaction, _ := cmd.Flags().GetBool("tx")

			if input[0] == '@' {
				// Read from file
				fileData, err := os.ReadFile(input[1:])
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				if isTransaction {
					// For transaction, just use the raw bytes
					// In a real implementation, we'd extract sign bytes properly
					dataToSign = fileData
				} else {
					dataToSign = fileData
				}
			} else if len(input) > 2 && input[:2] == "0x" {
				// Hex encoded data
				dataToSign, err = hex.DecodeString(input[2:])
				if err != nil {
					return fmt.Errorf("failed to decode hex: %w", err)
				}
			} else {
				// Plain text message
				dataToSign = []byte(input)
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

			// Sign the data
			signReq := &plugin.SignDataRequest{
				Data: dataToSign,
			}
			signResp, err := motorPlugin.SignData(signReq)
			if err != nil {
				return fmt.Errorf("failed to sign data: %w", err)
			}

			if signResp.Error != "" {
				return fmt.Errorf("plugin signing error: %s", signResp.Error)
			}

			// Get issuer info for display
			issuerResp, _ := motorPlugin.GetIssuerDID()

			// Output the signature
			result := map[string]any{
				"signature": hex.EncodeToString(signResp.Signature),
				"signer": map[string]any{
					"did":     issuerResp.IssuerDID,
					"address": issuerResp.Address,
				},
				"data_hash": hex.EncodeToString(dataToSign[:min(32, len(dataToSign))]),
			}

			// Save to file if requested
			outputFile, _ := cmd.Flags().GetString("output-file")
			if outputFile != "" {
				jsonData, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal result: %w", err)
				}
				if err := os.WriteFile(outputFile, jsonData, 0o644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Signature saved to %s\n", outputFile)
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String("enclave-data", "", "Enclave data for signing (required)")
	cmd.Flags().Bool("tx", false, "Sign as transaction (parse JSON as tx)")
	cmd.Flags().String("output-file", "", "Save signature to file")
	cmd.MarkFlagRequired("enclave-data")

	return cmd
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
