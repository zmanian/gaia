package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/cli"
	tmflags "github.com/tendermint/tmlibs/cli/flags"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	basecmd "github.com/cosmos/cosmos-sdk/server/commands"
	"github.com/cosmos/cosmos-sdk/stack"
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
	basecmd.Handler = stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
	).
		IBC(ibc.NewMiddleware()).
		Apps(
			roles.NewMiddleware(),
			fee.NewSimpleFeeMiddleware(coin.Coin{"mycoin", 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
		)

	RootCmd.AddCommand(
		commands.InitCmd,
		basecmd.StartCmd,
		basecmd.UnsafeResetAllCmd,
		commands.VersionCmd,
	)

	cmd := cli.PrepareMainCmd(RootCmd, "GA", os.ExpandEnv("$HOME/.cosmos-gaia"))
	cmd.Execute()
}
