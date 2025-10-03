// Package cli provides the DWN module CLI commands.
package cli

import (
	"github.com/spf13/cobra"
)

// AddWalletCmds adds wallet-specific commands to the root command
func AddWalletCmds(rootCmd *cobra.Command) {
	walletCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet operations",
	}

	walletCmd.AddCommand(
		SignCmd(),
		VerifyCmd(),
		SimulateCmd(),
		BroadcastCmd(),
	)

	// Add wallet commands
	rootCmd.AddCommand(walletCmd)
}
