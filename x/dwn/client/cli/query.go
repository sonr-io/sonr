package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sonr-io/sonr/x/dwn/types"
)

// GetQueryCmd returns the root query command for the DWN module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for " + types.ModuleName,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetCmdParams(),
		GetCmdEncryptionStatus(),
		GetCmdVRFContributions(),
		GetCmdEncryptedRecord(),
		GetWalletQueryCommands(),
	)
	return queryCmd
}

// GetCmdParams returns the command for querying module parameters
func GetCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show all module params",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdEncryptionStatus returns the command for querying encryption status
func GetCmdEncryptionStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encryption-status",
		Short: "Query current encryption key state and version",
		Long: `Query the current encryption status including:
- Current key version
- Validator set participating in consensus
- Single-node mode status
- Key rotation timestamps`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EncryptionStatus(
				context.Background(),
				&types.QueryEncryptionStatusRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdVRFContributions returns the command for querying VRF contributions
func GetCmdVRFContributions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrf-contributions [validator-address]",
		Short: "List VRF contributions for current consensus round",
		Long: `List VRF contributions for the current consensus round.
Optionally filter by validator address.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := &types.QueryVRFContributionsRequest{}

			// If validator address is provided, use it as filter
			if len(args) > 0 {
				req.ValidatorAddress = args[0]
			}

			// Get pagination from flags
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			req.Pagination = pageReq

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.VRFContributions(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "vrf-contributions")
	return cmd
}

// GetCmdEncryptedRecord returns the command for querying encrypted records
func GetCmdEncryptedRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypted-record [target-did] [record-id]",
		Short: "Query a specific encrypted record with automatic decryption",
		Long: `Query an encrypted DWN record and optionally decrypt it.
By default, the record data is decrypted if possible.
Use --return-encrypted to return the raw encrypted data instead.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			targetDID := args[0]
			recordID := args[1]

			// Check for return-encrypted flag
			returnEncrypted, err := cmd.Flags().GetBool("return-encrypted")
			if err != nil {
				return err
			}

			req := &types.QueryEncryptedRecordRequest{
				Target:          targetDID,
				RecordId:        recordID,
				ReturnEncrypted: returnEncrypted,
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EncryptedRecord(context.Background(), req)
			if err != nil {
				return err
			}

			// Add extra information about decryption status
			if !returnEncrypted {
				if res.WasDecrypted {
					fmt.Printf("✓ Record data was successfully decrypted\n\n")
				} else {
					fmt.Printf("⚠ Record data could not be decrypted or is not encrypted\n\n")
				}
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("return-encrypted", false, "Return encrypted data without decryption attempt")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
