package main

import (
	"os"

	"cosmossdk.io/log"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	_ "github.com/joho/godotenv/autoload"

	"github.com/sonr-io/snrd/app"
	"github.com/sonr-io/snrd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
		os.Exit(1)
	}
}
