package commands

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/gaia/modules/stake"
)

//nolint
var (
	CmdQueryCandidates = &cobra.Command{
		Use:   "candidates",
		Short: "Query for the set of validator-candidates pubkeys",
		RunE:  cmdQueryCandidates,
	}

	CmdQueryCandidate = &cobra.Command{
		Use:   "validator",
		Short: "Query a validator account",
		RunE:  cmdQueryCandidate,
	}
)

func init() {
	//Add Flags
	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagPubKey, "", "PubKey of the validator-candidate")

	CmdQueryCandidate.Flags().AddFlagSet(fsCandidate)
}

func cmdQueryCandidates(cmd *cobra.Command, args []string) error {

	var pks []crypto.PubKey

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	h, err := query.GetParsed(key, &pks, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(pks, h)
}

func cmdQueryCandidate(cmd *cobra.Command, args []string) error {

	var candidate stake.Candidate

	pk, err := getPubKey()
	if err != nil {
		return err
	}

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))
	h, err := query.GetParsed(key, &candidate, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(candidate, h)
}
