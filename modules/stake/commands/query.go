package commands

import "github.com/spf13/cobra"

var (
	//CmdQueryValidator - CLI command to query the counter state
	CmdQueryValidator = &cobra.Command{
		Use:   "validator",
		Short: "Query a validator account",
	}
)

//func init() {
//CmdQueryValidator
//}

//TODO complete functionality
func cmdQueryValidator(cmd *cobra.Command, args []string) error {
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
