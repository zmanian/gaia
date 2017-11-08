package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/gaia/modules/stake"
	scmds "github.com/cosmos/gaia/modules/stake/commands"

	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/tmlibs/common"
)

// RegisterQueryCandidate is a mux.Router handler that exposes GET
// method access on route /query/account/{signature} to query accounts
func RegisterQueryCandidate(r *mux.Router) error {
	r.HandleFunc("/query/stake/candidate/{pubkey}", doQueryCandidate).Methods("GET")
	return nil
}

// doQueryCandidate is the HTTP handlerfunc to query a candidate
// it expects a query string
func doQueryCandidate(w http.ResponseWriter, r *http.Request) {

	// get the arguments
	args := mux.Vars(r)
	pkArg := args["pubkey"]
	prove := !viper.GetBool(commands.FlagTrustNode) // from viper because defined when starting server

	// get the pubkey
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the candidate
	var candidate stake.Candidate
	key := stack.PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))
	height, err := query.GetParsed(key, &candidate, query.GetHeight(), prove)
	if lightclient.IsNoDataErr(err) {
		err := fmt.Errorf("candidate bytes are empty for pubkey: %q", pkArg)
		common.WriteError(w, err)
		return
	} else if err != nil {
		common.WriteError(w, err)
		return
	}

	// write the output
	err = query.FoutputProof(w, candidate, height)
	if err != nil {
		common.WriteError(w, err)
	}
}
