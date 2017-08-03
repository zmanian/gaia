package commands

import (
	"github.com/spf13/cobra"

	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"

	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tendermint/basecoin/stack"
)

var (
	//CmdQueryValidator - CLI command to query the counter state
	CmdQueryValidator = &cobra.Command{
		Use:   "validator",
		Short: "Query a validator",
	}

	//CmdQueryValidatorSummary - CLI command to query the counter state
	CmdQueryValidatorSummary = &cobra.Command{
		Use:   "summary",
		Short: "Query a validator summary",
		RunE:  counterQueryCmd,
	}

	//CmdQueryValidatorDeligates - CLI command to query the counter state
	CmdQueryValidatorDeligates = &cobra.Command{
		Use:   "deligates",
		Short: "Query a validator's deligates",
		RunE:  counterQueryCmd,
	}

	//CmdQueryDeligator - CLI command to query the counter state
	CmdQueryDeligator = &cobra.Command{
		Use:   "deligator",
		Short: "Query a the validators of a delagator",
	}

	//CmdQueryDeligatorSummary - CLI command to query the counter state
	CmdQueryDeligatorSummary = &cobra.Command{
		Use:   "summary",
		Short: "Query a deligator summary",
		RunE:  counterQueryCmd,
	}

	//CmdQueryDeligatorValidators - CLI command to query the counter state
	CmdQueryDeligatorValidators = &cobra.Command{
		Use:   "validators",
		Short: "Query a the validators of a delagator",
		RunE:  counterQueryCmd,
	}
)

func init() {
	//combine the subcommands
	CmdQueryValidator.AddCommand(
		CmdQueryValidatorSummary,
		CmdQueryValidatorDeligates,
	)
	CmdQueryDeligator.AddCommand(
		CmdQueryDeligatorSummary,
		CmdQueryDeligatorValidatos,
	)
}

func counterQueryCmd(cmd *cobra.Command, args []string) error {
	key := stack.PrefixedKey(counter.NameCounter, counter.StateKey())

	var cp counter.State
	proof, err := proofcmd.GetAndParseAppProof(key, &cp)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(cp, proof.BlockHeight())
}
