package rest

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/common"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
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

type DelegateInput struct {
	Fees     *coin.Coin `json:"fees"`
	Sequence uint32     `json:"sequence"`

	Pubkey crypto.PubKey `json:"pubkey"`
	From   *sdk.Actor    `json:"from"`
	Amount coin.Coin     `json:"amount"`
}

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
	r.HandleFunc("/build/stake/delegate", delegate).Methods("POST")
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
			"unbond",
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

func prepareDelegateTx(di *DelegateInput) sdk.Tx {
	tx := stake.NewTxDelegate(di.Amount, di.Pubkey)
	// fees are optional
	if di.Fees != nil && !di.Fees.IsZero() {
		tx = fee.NewFee(tx, *di.Fees, *di.From)
	}
	// only add the actual digner to the nonce
	digners := []sdk.Actor{*di.From}
	tx = nonce.NewTx(di.Sequence, digners, tx)
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	tx = auth.NewSig(tx).Wrap()
	return tx
}

func delegate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	di := new(DelegateInput)
	if err := common.ParseRequestAndValidateJSON(r, di); err != nil {
		common.WriteError(w, err)
		return
	}

	var errsList []string
	if di.From == nil {
		errsList = append(errsList, `"from" cannot be nil`)
	}
	if di.Sequence <= 0 {
		errsList = append(errsList, `"sequence" must be > 0`)
	}
	if di.Pubkey.Empty() {
		errsList = append(errsList, `"pubkey" cannot be empty`)
	}
	if len(errsList) > 0 {
		code := http.StatusBadRequest
		err := &common.ErrorResponse{
			Err:  strings.Join(errsList, ", "),
			Code: code,
		}
		common.WriteCode(w, err, code)
		return
	}

	tx := prepareDelegateTx(di)
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
