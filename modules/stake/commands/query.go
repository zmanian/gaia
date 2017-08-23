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

	//CmdQueryDelegateeDelegators - CLI command to query the counter state
	CmdQueryDelegateeDelegators = &cobra.Command{
		Use:   "deligators",
		Short: "Query a delegatee's delegators",
		RunE:  cmdQueryDelegateeDeligates,
	}

	//CmdQueryDelegator - CLI command to query the counter state
	CmdQueryDelegator = &cobra.Command{
		Use:   "delegator",
		Short: "Query a delegator",
	}

	//CmdQueryDelegatorSummary - CLI command to query the counter state
	CmdQueryDelegatorSummary = &cobra.Command{
		Use:   "summary",
		Short: "Query a delegator summary",
		RunE:  cmdQueryDelegatorSummary,
	}

	//CmdQueryDelegatorDelegatees - CLI command to query the counter state
	CmdQueryDelegatorDelegatees = &cobra.Command{
		Use:   "delegatees",
		Short: "Query a delegator's delegatees",
		RunE:  cmdQueryDelegatorDelegatees,
	}
)

func init() {
	//combine the subcommands
	CmdQueryDelegatee.AddCommand(
		CmdQueryDelegateeSummary,
		CmdQueryDelegateeDelegators,
	)
	CmdQueryDelegator.AddCommand(
		CmdQueryDelegatorSummary,
		CmdQueryDelegatorDelegatees,
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
func cmdQueryDelegatorSummary(cmd *cobra.Command, args []string) error {
	return nil
}

//TODO complete functionality
func cmdQueryDelegatorDelegatees(cmd *cobra.Command, args []string) error {
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
