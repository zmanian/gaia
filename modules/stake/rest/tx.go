package rest

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/gaia/modules/stake"
	scmds "github.com/cosmos/gaia/modules/stake/commands"
)

const (
	//parameters used in urls
	paramPubKey = "pubkey"
	paramAmount = "amount"
	paramShares = "shares"

	paramName    = "name"
	paramKeybase = "keybase"
	paramWebsite = "website"
	paramDetails = "details"
)

// RegisterDeclareCandidacy is a mux.Router handler that exposes
// POST method access on route /tx/stake/declare-candidacy to create a
// transaction for declaring candidacy
func RegisterDeclareCandidacy(r *mux.Router) error {
	r.HandleFunc(
		"/"+path.Join(
			"tx",
			"stake",
			"declare-candidacy",
			"{"+paramPubKey+"}",
			"{"+paramAmount+"}",
			"{"+paramName+"}",
			"{"+paramKeybase+"}",
			"{"+paramWebsite+"}",
			"{"+paramDetails+"}",
		),
		declareCandidacy,
	).Methods("POST")
	return nil
}

// RegisterEditCandidacy is a mux.Router handler that exposes
// POST method access on route /tx/stake/edit-candidacy to create a
// transaction for editing a candidate
func RegisterEditCandidacy(r *mux.Router) error {
	r.HandleFunc(
		"/"+path.Join(
			"tx",
			"stake",
			"edit-candidacy",
			"{"+paramPubKey+"}",
			"{"+paramName+"}",
			"{"+paramKeybase+"}",
			"{"+paramWebsite+"}",
			"{"+paramDetails+"}",
		),
		editCandidacy,
	).Methods("POST")
	return nil
}

// RegisterDelegate is a mux.Router handler that exposes
// POST method access on route /tx/stake/delegate to create a
// transaction for delegate to a candidaate/validator
func RegisterDelegate(r *mux.Router) error {
	r.HandleFunc(
		"/"+path.Join(
			"tx",
			"stake",
			"declare-candidacy",
			"{"+paramPubKey+"}",
			"{"+paramAmount+"}",
		),
		delegate,
	).Methods("POST")
	return nil
}

// RegisterUnbond is a mux.Router handler that exposes
// POST method access on route /tx/stake/unbond to create a
// transaction for unbonding delegated coins
func RegisterUnbond(r *mux.Router) error {
	r.HandleFunc(
		"/"+path.Join(
			"tx",
			"stake",
			"declare-candidacy",
			"{"+paramPubKey+"}",
			"{"+paramShares+"}",
		),
		unbond,
	).Methods("POST")
	return nil
}

func declareCandidacy(w http.ResponseWriter, r *http.Request) {
	// get the arguments object
	args := mux.Vars(r)

	// get the pubkey
	pkArg := args[paramPubKey]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the amount
	amountArg := args[paramAmount]
	amount, err := coin.ParseCoin(amountArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get description parameters
	description := stake.Description{
		Moniker: args[paramName],
		Keybase: args[paramKeybase],
		Website: args[paramWebsite],
		Details: args[paramDetails],
	}

	tx := stake.NewTxDeclareCandidacy(amount, pk, description)
	common.WriteSuccess(w, tx)
}

func editCandidacy(w http.ResponseWriter, r *http.Request) {

	// get the arguments object
	args := mux.Vars(r)

	// get the pubkey
	pkArg := args[paramPubKey]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get description parameters
	description := stake.Description{
		Moniker: args[paramName],
		Keybase: args[paramKeybase],
		Website: args[paramWebsite],
		Details: args[paramDetails],
	}

	tx := stake.NewTxEditCandidacy(pk, description)
	common.WriteSuccess(w, tx)
}

func delegate(w http.ResponseWriter, r *http.Request) {

	// get the arguments object
	args := mux.Vars(r)

	// get the pubkey
	pkArg := args[paramPubKey]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the amount
	amountArg := args[paramAmount]
	amount, err := coin.ParseCoin(amountArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	tx := stake.NewTxDelegate(amount, pk)
	common.WriteSuccess(w, tx)
}

func unbond(w http.ResponseWriter, r *http.Request) {
	// get the arguments object
	args := mux.Vars(r)

	// get the pubkey
	pkArg := args[paramPubKey]
	pk, err := scmds.GetPubKey(pkArg)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// get the shares
	sharesArg := args[paramShares]
	shares, err := strconv.ParseInt(sharesArg, 10, 64)
	if shares <= 0 {
		common.WriteError(w, fmt.Errorf("shares must be positive interger"))
		return
	}
	sharesU := uint64(shares)

	tx := stake.NewTxUnbond(sharesU, pk)
	common.WriteSuccess(w, tx)
}
