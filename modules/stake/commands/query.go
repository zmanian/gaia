package commands

import "github.com/spf13/cobra"

var (
	//CmdQueryDelegatee - CLI command to query the counter state
	CmdQueryDelegatee = &cobra.Command{
		Use:   "delegatee",
		Short: "Query a delegatee",
	}

	//CmdQueryDelegateeSummary - CLI command to query the counter state
	CmdQueryDelegateeSummary = &cobra.Command{
		Use:   "summary",
		Short: "Query a delegatee summary",
		RunE:  cmdQueryDelegateeSummary,
	}

	//CmdQueryDelegateeDeligators - CLI command to query the counter state
	CmdQueryDelegateeDeligators = &cobra.Command{
		Use:   "deligators",
		Short: "Query a delegatee's delegators",
		RunE:  cmdQueryDelegateeDeligates,
	}

	//CmdQueryDeligator - CLI command to query the counter state
	CmdQueryDeligator = &cobra.Command{
		Use:   "delegator",
		Short: "Query a delegator",
	}

	//CmdQueryDeligatorSummary - CLI command to query the counter state
	CmdQueryDeligatorSummary = &cobra.Command{
		Use:   "summary",
		Short: "Query a delegator summary",
		RunE:  cmdQueryDeligatorSummary,
	}

	//CmdQueryDeligatorDelegatees - CLI command to query the counter state
	CmdQueryDeligatorDelegatees = &cobra.Command{
		Use:   "delegatees",
		Short: "Query a delegator's delegatees",
		RunE:  cmdQueryDeligatorDelegatees,
	}
)

func init() {
	//combine the subcommands
	CmdQueryDelegatee.AddCommand(
		CmdQueryDelegateeSummary,
		CmdQueryDelegateeDeligates,
	)
	CmdQueryDeligator.AddCommand(
		CmdQueryDeligatorSummary,
		CmdQueryDeligatorValidators,
	)
}

//TODO complete functionality
func cmdQueryDelegateeSummary(cmd *cobra.Command, args []string) error {
	return nil
}

//TODO complete functionality
func cmdQueryDelegateeDeligates(cmd *cobra.Command, args []string) error {
	return nil
}

//TODO complete functionality
func cmdQueryDeligatorSummary(cmd *cobra.Command, args []string) error {
	return nil
}

//TODO complete functionality
func cmdQueryDeligatorDelegatees(cmd *cobra.Command, args []string) error {
	return nil
}

//func counterQueryCmd(cmd *cobra.Command, args []string) error {
//key := stack.PrefixedKey(counter.NameCounter, counter.StateKey())

//var cp counter.State
//proof, err := proofcmd.GetAndParseAppProof(key, &cp)
//if err != nil {
//return err
//}

//return proofcmd.OutputProof(cp, proof.BlockHeight())
//}
