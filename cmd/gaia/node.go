package main

import (
	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	//"github.com/tendermint/tmlibs/cli"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	basecmd "github.com/cosmos/cosmos-sdk/server/commands"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"

	"github.com/cosmos/gaia/modules/stake"
	"github.com/cosmos/gaia/version"
)

// RootCmd is the entry point for this binary
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "The Cosmos Network delegation-game blockchain test",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Tick - Called every block even if no transaction,
// process all queues, validator rewards, and calculate the validator set difference
func tickFn(ctx sdk.Context, store state.SimpleDB) (diffVal []*abci.Validator, err error) {
	// First need to prefix the store, at this point it's a global store
	store = stack.PrefixedStore(stake.Name(), store)

	// Determine the validator set changes
	candidates := stake.LoadCandidates(store)
	startVal := candidates.GetValidators(store)
	changed := candidates.UpdateVotingPower(store)
	if !changed {
		return
	}
	newVal := candidates.GetValidators(store)
	diffVal = stake.ValidatorsDiff(startVal, newVal, store)
	return
}

func cmdNode(cmd *cobra.Command, args []string) error {
	// require all fees in strings - change this in your app!
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
			fee.NewSimpleFeeMiddleware(coin.Coin{"strings", 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
			stake.NewHandler(),
		)

	NodeCmd.AddCommand(
		basecmd.GetInitCmd("fermion", []string{"stake/allowed_bond_denom/fermion"}),
		basecmd.GetTickStartCmd(sdk.TickerFunc(tickFn)),
		basecmd.UnsafeResetAllCmd,
		version.VersionCmd,
	)
	basecmd.SetUpRoot(NodeCmd)

	return nil
	// XXX zr ??
	//cmd := cli.PrepareMainCmd(NodeCmd, "GA", os.ExpandEnv("$HOME/.cosmos-gaia"))
}
