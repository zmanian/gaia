package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/gaia/modules/stake"
	scmds "github.com/cosmos/gaia/modules/stake/commands"
)

// RegisterDeclareCandidacy is a mux.Router handler that exposes
// POST method access on route /tx/stake/declare-candidacy to create a
// transaction for declaring candidacy
func RegisterDeclareCandidacy(r *mux.Router) error {
	r.HandleFunc("/tx/stake/declare-candidacy/{pubkey}/{amount}", declareCandidacy).Methods("POST")
	return nil
}

// RegisterDelegate is a mux.Router handler that exposes
// POST method access on route /tx/stake/delegate to create a
// transaction for delegate to a candidaate/validator
func RegisterDelegate(r *mux.Router) error {
	r.HandleFunc("/tx/stake/delegate/{pubkey}/{amount}", delegate).Methods("POST")
	return nil
}

// RegisterUnbond is a mux.Router handler that exposes
// POST method access on route /tx/stake/unbond to create a
// transaction for unbonding delegated coins
func RegisterUnbond(r *mux.Router) error {
	r.HandleFunc("/tx/stake/unbond/{pubkey}/{amount}", unbond).Methods("POST")
	return nil
}

func declareCandidacy(w http.ResponseWriter, r *http.Request) {
	bondUpdate(w, r, stake.NewTxDeclareCandidacy)
}
func delegate(w http.ResponseWriter, r *http.Request) {
	bondUpdate(w, r, stake.NewTxDelegate)
}
func unbond(w http.ResponseWriter, r *http.Request) {
	bondUpdate(w, r, stake.NewTxUnbond)
}

func bondUpdate(w http.ResponseWriter, r *http.Request, makeTx scmds.MakeTx) {
	// get the arguments object
	args := mux.Vars(r)

	// get the pubkey
	pkArg := args["pubkey"]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the amount
	amountArg := args["amount"]
	amount, err := coin.ParseCoin(viper.GetString(amountArg))
	if err != nil {
		common.WriteError(w, err)
		return
	}

	tx := makeTx(amount, pk)
	common.WriteSuccess(w, tx)
}
