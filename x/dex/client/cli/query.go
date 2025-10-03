package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/sonr-io/sonr/x/dex/types"
)

// NewQueryCmd creates and returns the query command
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryAccount(),
		CmdQueryAccounts(),
		CmdQueryBalance(),
		CmdQueryPool(),
		CmdQueryOrders(),
		CmdQueryHistory(),
	)

	return cmd
}

// CmdQueryParams queries the module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current DEX module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAccount queries a DEX account
func CmdQueryAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [did] [connection-id]",
		Short: "Query a DEX account by DID and connection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Account(context.Background(), &types.QueryAccountRequest{
				Did:          did,
				ConnectionId: connectionID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAccounts queries all DEX accounts
func CmdQueryAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Query all DEX accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			did, _ := cmd.Flags().GetString("did")

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Accounts(context.Background(), &types.QueryAccountsRequest{
				Did:        did,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("did", "", "Filter by DID")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "accounts")
	return cmd
}

// CmdQueryBalance queries remote chain balances
func CmdQueryBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [did] [connection-id]",
		Short: "Query remote chain balances for a DID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]
			denom, _ := cmd.Flags().GetString("denom")

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Balance(context.Background(), &types.QueryBalanceRequest{
				Did:          did,
				ConnectionId: connectionID,
				Denom:        denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("denom", "", "Filter by specific denomination")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPool queries pool information
func CmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [connection-id] [pool-id]",
		Short: "Query pool information on a remote chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			connectionID := args[0]
			poolID := args[1]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Pool(context.Background(), &types.QueryPoolRequest{
				ConnectionId: connectionID,
				PoolId:       poolID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryOrders queries orders for a DID
func CmdQueryOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders [did] [connection-id]",
		Short: "Query orders for a DID on a specific connection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]
			status, _ := cmd.Flags().GetString("status")

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Orders(context.Background(), &types.QueryOrdersRequest{
				Did:          did,
				ConnectionId: connectionID,
				Status:       status,
				Pagination:   pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by order status (pending|open|filled|cancelled)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	return cmd
}

// CmdQueryHistory queries transaction history
func CmdQueryHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [did]",
		Short: "Query transaction history for a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID, _ := cmd.Flags().GetString("connection")
			operationType, _ := cmd.Flags().GetString("type")

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.History(context.Background(), &types.QueryHistoryRequest{
				Did:           did,
				ConnectionId:  connectionID,
				OperationType: operationType,
				Pagination:    pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("connection", "", "Filter by connection ID")
	cmd.Flags().String("type", "", "Filter by transaction type (swap|liquidity|order)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "history")
	return cmd
}
