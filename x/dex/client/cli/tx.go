package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/sonr-io/sonr/x/dex/types"
)

// NewTxCmd creates and returns the tx command
func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegisterDEXAccount(),
		CmdExecuteSwap(),
		CmdProvideLiquidity(),
		CmdRemoveLiquidity(),
		CmdCreateLimitOrder(),
		CmdCancelOrder(),
	)

	return cmd
}

// CmdRegisterDEXAccount returns a command to register a DEX account
func CmdRegisterDEXAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-account [did] [connection-id] [features]",
		Short: "Register a new ICA account for DEX operations",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]
			features := strings.Split(args[2], ",")

			msg := &types.MsgRegisterDEXAccount{
				Did:          did,
				ConnectionId: connectionID,
				Features:     features,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdExecuteSwap returns a command to execute a swap
func CmdExecuteSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [did] [connection-id] [token-in] [token-out-denom] [min-amount-out]",
		Short: "Execute a token swap through ICA (supports Noble USDC)",
		Long: `Execute a token swap on a remote chain via Interchain Accounts.

Examples:
  # Swap 1000000 uatom for USDC on Noble testnet
  snrd tx dex swap did:snr:user1 connection-0 1000000uatom uusdc 950000 --ucan-token="..." --timeout=60s

  # Swap USDC for ATOM on Osmosis
  snrd tx dex swap did:snr:user1 connection-1 1000000uusdc uatom 950000 --route="pool:1"
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]

			tokenIn, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid token-in: %w", err)
			}

			tokenOutDenom := args[3]

			minAmountOut, ok := math.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid min-amount-out: %s", args[4])
			}

			// Parse optional flags
			ucanToken, _ := cmd.Flags().GetString("ucan-token")
			route, _ := cmd.Flags().GetString("route")
			timeoutStr, _ := cmd.Flags().GetString("timeout")

			// Parse timeout duration
			timeoutDuration := 30 * time.Second // default
			if timeoutStr != "" {
				timeoutDuration, err = time.ParseDuration(timeoutStr)
				if err != nil {
					return fmt.Errorf("invalid timeout: %w", err)
				}
			}

			msg := &types.MsgExecuteSwap{
				Did:          did,
				ConnectionId: connectionID,
				SourceDenom:  tokenIn.Denom,
				TargetDenom:  tokenOutDenom,
				Amount:       tokenIn.Amount,
				MinAmountOut: minAmountOut,
				Route:        route,
				UcanToken:    ucanToken,
				Timeout:      time.Now().Add(timeoutDuration),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("ucan-token", "", "UCAN authorization token for permission delegation")
	cmd.Flags().String("route", "", "Optional specific swap route (e.g., 'pool:1' or 'noble:channel-0')")
	cmd.Flags().String("timeout", "30s", "Timeout duration for the swap (e.g., '30s', '1m')")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdProvideLiquidity returns a command to provide liquidity
func CmdProvideLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provide-liquidity [did] [connection-id] [pool-id] [token-a] [token-b] [min-shares]",
		Short: "Provide liquidity to a pool through ICA",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]

			poolID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool-id: %w", err)
			}

			tokenA, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return fmt.Errorf("invalid token-a: %w", err)
			}

			tokenB, err := sdk.ParseCoinNormalized(args[4])
			if err != nil {
				return fmt.Errorf("invalid token-b: %w", err)
			}

			minShares, ok := math.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("invalid min-shares: %s", args[5])
			}

			msg := &types.MsgProvideLiquidity{
				Did:          did,
				ConnectionId: connectionID,
				PoolId:       fmt.Sprintf("%d", poolID),
				Assets:       sdk.NewCoins(tokenA, tokenB),
				MinShares:    minShares,
				Timeout:      time.Now().Add(5 * time.Minute),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveLiquidity returns a command to remove liquidity
func CmdRemoveLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [did] [connection-id] [pool-id] [shares] [min-amount-a] [min-amount-b]",
		Short: "Remove liquidity from a pool through ICA",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]

			poolID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool-id: %w", err)
			}

			shares, ok := math.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid shares: %s", args[3])
			}

			minAmountA, ok := math.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid min-amount-a: %s", args[4])
			}

			minAmountB, ok := math.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("invalid min-amount-b: %s", args[5])
			}

			msg := &types.MsgRemoveLiquidity{
				Did:          did,
				ConnectionId: connectionID,
				PoolId:       fmt.Sprintf("%d", poolID),
				Shares:       shares,
				MinAmounts: sdk.NewCoins(
					sdk.NewCoin("token", minAmountA),
					sdk.NewCoin("token", minAmountB),
				),
				Timeout: time.Now().Add(5 * time.Minute),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateLimitOrder returns a command to create a limit order
func CmdCreateLimitOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-order [did] [connection-id] [token-in] [token-out-denom] [price]",
		Short: "Create a limit order through ICA",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]

			tokenIn, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid token-in: %w", err)
			}

			tokenOutDenom := args[3]

			price, err := math.LegacyNewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}

			msg := &types.MsgCreateLimitOrder{
				Did:          did,
				ConnectionId: connectionID,
				SellDenom:    tokenIn.Denom,
				BuyDenom:     tokenOutDenom,
				Amount:       tokenIn.Amount,
				Price:        price,
				Expiration:   time.Now().Add(24 * time.Hour),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCancelOrder returns a command to cancel an order
func CmdCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-order [did] [connection-id] [order-id]",
		Short: "Cancel an existing order through ICA",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			connectionID := args[1]
			orderID := args[2]

			msg := &types.MsgCancelOrder{
				Did:          did,
				ConnectionId: connectionID,
				OrderId:      orderID,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
