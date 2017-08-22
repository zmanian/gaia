package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/client/commands"
	basecmd "github.com/cosmos/cosmos-sdk/cmd/basecoin/commands"
	"github.com/cosmos/gaia/modules/stake"
)

//nolint
const (
	defaultLogLevel = "error"
	FlagLogLevel    = "log_level"
)

var (
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
)

// RootCmd - main node command
var RootCmd = &cobra.Command{
	Use:   "gaia",
	Short: "The Cosmos Network delegation-game blockchain test",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		level := viper.GetString(FlagLogLevel)
		logger, err = tmflags.ParseLogLevel(level, logger, defaultLogLevel)
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().String(FlagLogLevel, defaultLogLevel, "Log level")
}

func main() {
	// require all fees in mycoin - change this in your app!
	basecmd.Handler = stake.NewHandler("mycoin")

	RootCmd.AddCommand(
		commands.InitCmd,
		basecmd.StartCmd,
		basecmd.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(RootCmd, "GA", os.ExpandEnv("$HOME/.cosmos-gaia"))
	cmd.Execute()
}
