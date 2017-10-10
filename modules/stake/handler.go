package stake

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	stakingModuleName = "stake"
)

// Name is the name of the modules.
func Name() string {
	return stakingModuleName
}

// Params defines the parameters for bonding and unbonding
type Params struct {
	MaxVals          int    `json:"max_vals"`           // maximum number of validators
	AllowedBondDenom string `json:"allowed_bond_denom"` // bondable coin denomination

	// gas costs for txs
	GasBond   uint64 `json:"gas_bond"`
	GasUnbond uint64 `json:"gas_unbond"`
}

func defaultParams() Params {
	return Params{
		MaxVals:          100,
		AllowedBondDenom: "strings",
		GasBond:          20,
		GasUnbond:        0,
	}
}

// TODO sync with the state
// global for now
var globalParams = defaultParams()

// Handler - the transaction processing handler
type Handler struct {
	stack.PassInitValidate
}

// NewHandler returns a new Handler with the default Params.
func NewHandler() Handler {
	return Handler{}
}

var _ stack.Dispatchable = Handler{} // enforce interface at compile time

// Name - return stake namespace
func (Handler) Name() string {
	return stakingModuleName
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

// InitState - set genesis parameters for staking
func (h Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	return "", h.initState(module, key, value, store)
}

//separated for testing
func (Handler) initState(module, key, value string, store state.SimpleDB) error {
	if module != stakingModuleName {
		return errors.ErrUnknownModule(module)
	}

	switch key {
	case "allowed_bond_denom":
		globalParams.AllowedBondDenom = value
		return nil
	case "max_vals",
		"gas_bond",
		"gas_unbond":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}
		switch key {
		case "max_val":
			globalParams.MaxVals = i
		case "gas_bond":
			globalParams.GasBond = uint64(i)
		case "gas_unbound":
			globalParams.GasUnbond = uint64(i)
		}
		return nil
	}
	return errors.ErrUnknownKey(key)
}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (res sdk.CheckResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	// get the sender
	sender, abciRes := getTxSender(ctx)
	if abciRes.IsErr() {
		return res, abciRes
	}

	// return the fee for each tx type
	switch txInner := tx.Unwrap().(type) {
	case TxBond:
		return sdk.NewCheck(globalParams.GasBond, ""),
			checkTxBond(txInner, sender, store)
	case TxUnbond:
		return sdk.NewCheck(globalParams.GasUnbond, ""),
			checkTxUnbond(txInner, sender, store)
	}
	return res, errors.ErrUnknownTxType("GTH")
}

func checkTxBond(tx TxBond, sender sdk.Actor, store state.SimpleDB) error {
	// TODO check the sender has enough coins to bond
	//acc := coin.Account{}
	//// vvv this causes nil pointer ref error INSIDE of GetParsed
	//_, err := query.GetParsed(sender.Address, &acc, true) //NOTE we are not using proof queries
	//if err != nil {
	//return err
	//}
	//if acc.Coins.IsGTE(coin.Coins{tx.Amount}) {
	//return fmt.Errorf("not enough coins to bond, have %v, trying to bond %v",
	//acc.Coins, tx.Amount)
	//}

	// check to see if the pubkey has been registered before,
	//  if it has been used ensure that the validator account is same
	//  to prevent accidentally bonding to validator other than you
	bonds := LoadBonds(store)
	_, bond := bonds.GetByPubKey(tx.PubKey)
	if bond != nil {
		if !bond.Validator.Equals(sender) {
			return fmt.Errorf("cannot bond tokens to pubkey used by another validator"+
				" PubKey %v already registered with %v validator address",
				bond.PubKey, bond.Validator)
		}
	}
	return nil
}

func checkTxUnbond(tx TxUnbond, sender sdk.Actor, store state.SimpleDB) error {

	//check if have enough tokens to unbond
	bonds := LoadBonds(store)
	_, bond := bonds.Get(sender)
	if bond.BondedTokens < uint64(tx.Amount.Amount) {
		return fmt.Errorf("not enough bond tokens to unbond, have %v, trying to unbond %v",
			bond.BondedTokens, tx.Amount)
	}
	return nil
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	// TODO: remove redunandcy
	// also we don't need to check the res - gas is already deducted in sdk
	_, err = h.CheckTx(ctx, store, tx, dispath)
	if err != nil {
		return
	}

	sender, abciRes := getTxSender(ctx)
	if abciRes.IsErr() {
		return res, abciRes
	}

	// get the holding account for the sender's bond.
	// holding account is just an sdk.Actor, with the sender's address shifted one byte right.
	holder := getHoldAccount(sender)

	//Run the transaction
	switch _tx := tx.Unwrap().(type) {
	case TxBond:
		fn := defaultTransferFn(ctx, store, dispatch)
		abciRes = runTxBond(store, sender, holder, fn, _tx)
	case TxUnbond:
		//context with hold account permissions
		ctx2 := ctx.WithPermissions(holder)
		fn := defaultTransferFn(ctx2, store, dispatch)
		abciRes = runTxUnbond(store, sender, holder, fn, _tx)
	}

	res = sdk.DeliverResult{
		Data:    abciRes.Data,
		Log:     abciRes.Log,
		GasUsed: globalParams.GasBond,
	}

	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////
// these functions assume everything has been authenticated,
// now we just bond or unbond and save

func runTxBond(store state.SimpleDB, sender, holder sdk.Actor,
	transferFn transferFn, tx TxBond) (res abci.Result) {

	// Get amount of coins to bond
	bondCoin := tx.Amount

	// Get the validator bond accounts, and bond and index for this sender
	bonds := LoadBonds(store)
	idx, bond := bonds.Get(sender)
	if bond == nil { //if it doesn't yet exist create it
		bonds = bonds.Add(NewValidatorBond(sender, holder, tx.PubKey))
		idx = len(bonds) - 1
	}

	// Move coins from the sender account to the holder account
	res = transferFn(sender, holder, coin.Coins{bondCoin})
	if res.IsErr() {
		return res
	}

	// Update the bond and save to store
	bonds[idx].BondedTokens += uint64(bondCoin.Amount)
	saveBonds(store, bonds)

	return abci.OK
}

func runTxUnbond(store state.SimpleDB, sender, holder sdk.Actor,
	transferFn transferFn, tx TxUnbond) (res abci.Result) {

	//get validator bond
	bonds := LoadBonds(store)
	_, bond := bonds.Get(sender)
	if bond == nil {
		return resNoValidatorForAddress
	}

	// transfer coins back to account
	unbondCoin := tx.Amount
	unbondAmt := uint64(unbondCoin.Amount)
	res = transferFn(holder, sender, coin.Coins{unbondCoin})
	if res.IsErr() {
		return res
	}

	bond.BondedTokens -= unbondAmt

	saveBonds(store, bonds)
	return abci.OK
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx sdk.Context) (sender sdk.Actor, res abci.Result) {
	senders := ctx.GetPermissions("", auth.NameSigs)
	if len(senders) != 1 {
		return sender, resMissingSignature
	}
	// TODO: ensure senders[0] matches tx.pubkey ...
	// NOTE on TODO..  right now the PubKey doesn't need to match the sender
	// and we actually don't have the means to construct the priv_validator.json
	// with its private key with current keys tooling in SDK so needs to be
	// a second key... This is still secure because you will only be able to
	// unbond to the first married account, although, you could hypotheically
	// bond some coins to somebody elses account (effectively giving them coins)
	// maybe that is worth checking more. Validators should probably be allowed
	// to use two different keys, one for validating and one with coins on it...
	// so this point may never be relevant
	return senders[0], abci.OK
}

func getHoldAccount(sender sdk.Actor) sdk.Actor {
	holdAddr := append([]byte{0x00}, sender.Address[1:]...) //shift and prepend a zero
	return sdk.NewActor(stakingModuleName, holdAddr)
}
