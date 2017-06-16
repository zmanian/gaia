package main

import (
	"os"

	"github.com/spf13/cobra"

	// import _ to register the stake plugin to apptx
	_ "github.com/tendermint/basecoin-stake/cmd/stakecoin/commands"
	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/tmlibs/cli"
)

func main() {
	var RootCmd = &cobra.Command{
		Use: "gaia",
	}

	RootCmd.AddCommand(
		commands.InitCmd,
		commands.StartCmd,
		commands.TxCmd,
		commands.QueryCmd,
		commands.KeyCmd,
		commands.VerifyCmd,
		commands.BlockCmd,
		commands.AccountCmd,
		commands.UnsafeResetAllCmd,
		commands.QuickVersionCmd("0.1.0"),
	)

	cli.PrepareMainCmd(
		RootCmd,
		"GAIA", // envvar prefix
		os.ExpandEnv("$HOME/.cosmos-gaia"), // default home
	).Execute()
}
