package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/tmlibs/cli"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/proxy"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/client/commands/seeds"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	authcmd "github.com/cosmos/cosmos-sdk/modules/auth/commands"
	"github.com/cosmos/cosmos-sdk/modules/base"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	coincmd "github.com/cosmos/cosmos-sdk/modules/coin/commands"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	feecmd "github.com/cosmos/cosmos-sdk/modules/fee/commands"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	noncecmd "github.com/cosmos/cosmos-sdk/modules/nonce/commands"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"

	stkcmd "github.com/cosmos/gaia/modules/stake/commands"
)

// GaiaCli represents the base command when called without any subcommands
var GaiaCli = &cobra.Command{
	Use:   "gaiacli",
	Short: "Client for Cosmos-Gaia blockchain",
}

// BuildApp constructs the stack we want to use for this app
func BuildApp(feeDenom string) sdk.Handler {
	return stack.New(
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
			fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
		)
}

func main() {
	commands.AddBasicFlags(GaiaCli)

	// Prepare queries
	query.RootCmd.AddCommand(
		// These are default parsers, optional in your app
		query.TxQueryCmd,
		query.KeyQueryCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,

		// Staking commands
		stkcmd.CmdQueryDelegatee,
		stkcmd.CmdQueryDelegator,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		feecmd.FeeWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	// Prepare transactions
	txcmd.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,

		// Staking commands
		stkcmd.CmdBond,
		stkcmd.CmdUnbond,
		stkcmd.CmdNominate,
		stkcmd.CmdModComm,
	)

	// Set up the various high-level commands to use
	GaiaCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		commands.VersionCmd, //TODO update to custom version command
		keycmd.RootCmd,
		seeds.RootCmd,
		query.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(GaiaCli, "GA", os.ExpandEnv("$HOME/.cosmos-gaia-cli"))
	cmd.Execute()
}
