package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/commits"
	"github.com/cosmos/cosmos-sdk/client/commands/keys"
	"github.com/cosmos/cosmos-sdk/client/commands/proxy"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	rpccmd "github.com/cosmos/cosmos-sdk/client/commands/rpc"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	authcmd "github.com/cosmos/cosmos-sdk/modules/auth/commands"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	coincmd "github.com/cosmos/cosmos-sdk/modules/coin/commands"
	feecmd "github.com/cosmos/cosmos-sdk/modules/fee/commands"
	ibccmd "github.com/cosmos/cosmos-sdk/modules/ibc/commands"
	noncecmd "github.com/cosmos/cosmos-sdk/modules/nonce/commands"
	rolecmd "github.com/cosmos/cosmos-sdk/modules/roles/commands"

	stakecmd "github.com/cosmos/gaia/modules/stake/commands"
	"github.com/cosmos/gaia/version"
)

// GaiaCmd is the entry point for this binary
var GaiaCmd = &cobra.Command{
	Use:   "gaia",
	Short: "The Cosmos Network delegation-game blockchain test",
	Long:  "", //TODO re-work
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute - execute the main commands
func Execute() {
	addGlobalCommands()
	GaiaCmd.Execute()
}

func addGlobalCommands() {

	nodeCommands()
	GaiaCmd.AddCommand(nodeCmd)

	serverCommands()
	GaiaCmd.AddCommand(serverCmd)

	cliCommands()

	addGlobalFlags()
}

func addGlobalFlags() {
	commands.AddBasicFlags(GaiaCmd)
	//GaiaCmd.PersistentFlags().StringVarP()
}

func cliCommands() {
	// Prepare queries
	query.RootCmd.AddCommand(
		// These are default parsers, but optional in your app (you can remove key)
		query.TxQueryCmd,
		query.KeyQueryCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,
		rolecmd.RoleQueryCmd,
		ibccmd.IBCQueryCmd,

		//stakecmd.CmdQueryValidator,
		stakecmd.CmdQueryCandidates,
		stakecmd.CmdQueryCandidate,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		feecmd.FeeWrapper{},
		rolecmd.RoleWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	// you will always want this for the base send command
	txcmd.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,
		coincmd.CreditTxCmd,
		// this enables creating roles
		rolecmd.CreateRoleTxCmd,
		// these are for handling ibc
		ibccmd.RegisterChainTxCmd,
		ibccmd.UpdateChainTxCmd,
		ibccmd.PostPacketTxCmd,

		stakecmd.CmdDeclareCandidacy,
		stakecmd.CmdDelegate,
		stakecmd.CmdUnbond,
		stakecmd.CmdDeclareCandidacy,
	)

	// Set up the various commands to use
	GaiaCmd.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keys.RootCmd,
		commits.RootCmd,
		rpccmd.RootCmd,
		query.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
		version.VersionCmd,
		//auto.AutoCompleteCmd,
	)

	_ = cli.PrepareMainCmd(GaiaCmd, "GA", os.ExpandEnv("$HOME/.cosmos-gaia-cli"))
}
