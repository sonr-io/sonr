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

// VerifyCmd returns a command to verify signatures using the Motor plugin
func VerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [signature-file]",
		Short: "Verify a signature using Motor plugin",
		Long: `Verify a signature against data using the Motor plugin's MPC-based verification.

Examples:
  # Verify a signature file
  snrd wallet verify signature.json --data "Hello World" --enclave-data @enclave.json
  
  # Verify with data from file
  snrd wallet verify signature.json --data @message.txt --enclave-data @enclave.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Read signature file
			sigFile := args[0]
			sigData, err := os.ReadFile(sigFile)
			if err != nil {
				return fmt.Errorf("failed to read signature file: %w", err)
			}

			// Parse signature data
			var sigInfo struct {
				Signature string `json:"signature"`
				Signer    struct {
					DID     string `json:"did"`
					Address string `json:"address"`
				} `json:"signer"`
			}
			if err := json.Unmarshal(sigData, &sigInfo); err != nil {
				return fmt.Errorf("failed to parse signature file: %w", err)
			}

			// Get data to verify against
			dataInput, err := cmd.Flags().GetString("data")
			if err != nil {
				return err
			}

			if dataInput == "" {
				return fmt.Errorf("--data flag is required")
			}

			var dataToVerify []byte
			if dataInput[0] == '@' {
				// Read from file
				dataToVerify, err = os.ReadFile(dataInput[1:])
				if err != nil {
					return fmt.Errorf("failed to read data file: %w", err)
				}
			} else {
				dataToVerify = []byte(dataInput)
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

			// Decode signature
			signature, err := hex.DecodeString(sigInfo.Signature)
			if err != nil {
				return fmt.Errorf("failed to decode signature: %w", err)
			}

			// Verify the signature
			verifyReq := &plugin.VerifyDataRequest{
				Data:      dataToVerify,
				Signature: signature,
			}
			verifyResp, err := motorPlugin.VerifyData(verifyReq)
			if err != nil {
				return fmt.Errorf("failed to verify signature: %w", err)
			}

			if verifyResp.Error != "" {
				return fmt.Errorf("plugin verification error: %s", verifyResp.Error)
			}

			// Output verification result
			result := map[string]any{
				"valid": verifyResp.Valid,
				"signer": map[string]any{
					"did":     sigInfo.Signer.DID,
					"address": sigInfo.Signer.Address,
				},
				"signature_length": len(signature),
				"data_length":      len(dataToVerify),
			}

			if verifyResp.Valid {
				fmt.Println("✓ Signature is valid")
			} else {
				fmt.Println("✗ Signature is invalid")
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String("data", "", "Data to verify signature against (text or @file)")
	cmd.Flags().String("enclave-data", "", "Enclave data for verification (required)")
	cmd.MarkFlagRequired("data")
	cmd.MarkFlagRequired("enclave-data")

	return cmd
}
