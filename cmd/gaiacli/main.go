package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/tmlibs/cli"

	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/client/commands/proxy"
	"github.com/tendermint/basecoin/client/commands/seeds"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	authcmd "github.com/tendermint/basecoin/modules/auth/commands"
	basecmd "github.com/tendermint/basecoin/modules/base/commands"
	coincmd "github.com/tendermint/basecoin/modules/coin/commands"
	feecmd "github.com/tendermint/basecoin/modules/fee/commands"
	noncecmd "github.com/tendermint/basecoin/modules/nonce/commands"

	stkcmd "github.com/cosmos/gaia/modules/stake/commands"
)

// GaiaCli represents the base command when called without any subcommands
var GaiaCli = &cobra.Command{
	Use:   "gaiacli",
	Short: "Client for Cosmos-Gaia blockchain",
}

func main() {
	commands.AddBasicFlags(GaiaCli)

	// Prepare queries
	proofs.RootCmd.AddCommand(
		// These are default parsers, optional in your app
		proofs.TxQueryCmd,
		proofs.KeyQueryCmd,
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
	proofs.TxPresenters.Register("base", txcmd.BaseTxPresenter{})
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
		keycmd.RootCmd,
		seeds.RootCmd,
		proofs.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(GaiaCli, "GA", os.ExpandEnv("$HOME/.cosmos-gaia-cli"))
	cmd.Execute()
}
