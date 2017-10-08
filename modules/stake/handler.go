package stake

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
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
// TODO do we need name to be unexposed for security?
func Name() string {
	return stakingModuleName
}

var (
	allowedBondDenom string = "strings" // bondable coin denomination
)

// Params defines the parameters for bonding and unbonding
type Params struct {
	MaxVals int `json:"max_vals"` // maximum number of validators

	// gas costs for txs
	GasBond   uint64 `json:"gas_bond"`
	GasUnbond uint64 `json:"gas_unbond"`
}

func defaultParams() Params {
	return Params{
		MaxVals:   100,
		GasBond:   20,
		GasUnbond: 0,
	}
}

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
func (Handler) initState(module, key, value string, store state.SimpleDB) (err error) {
	if module != stakingModuleName {
		return errors.ErrUnknownModule(module)
	}
	switch key {
	case "allowed_bond_denom":
		allowedBondDenom = value
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
	sender, err := getTxSender(ctx)
	if err != nil {
		return res, err
	}
	_ = sender

	// return the fee for each tx type
	switch tx.Unwrap().(type) {
	case TxBond:
		// TODO: check the sender has enough coins to bond
		return sdk.NewCheck(globalParams.GasBond, ""), nil
	case TxUnbond:
		// TODO check the sender has coins to unbond
		return sdk.NewCheck(globalParams.GasUnbond, ""), nil
	}
	return res, errors.ErrUnknownTxType("GTH")
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	sender, abciRes := getTxSender(ctx)
	if abciRes.IsErr() {
		return res, abciRes
	}

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
		Data: abciRes.Data,
		Log:  abciRes.Log,
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

	// Get the validator bond account
	bonds, err := LoadBonds(store)
	if err != nil {
		return resErrLoadingValidators(err)
	}

	// Get the bond and index for this sender
	idx, bond := bonds.Get(sender)
	if bond == nil { //if it doesn't yet exist create it
		bond = NewValidatorBond(sender, holder, tx.PubKey)
		bonds = append(bonds, bond)
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
	bonds, err := LoadBonds(store)
	if err != nil {
		return resErrLoadingValidators(err)
	}

	idx, bond := bonds.Get(sender)
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

	if bond.BondedTokens == 0 {
		bonds, err = bonds.Remove(idx)
		if err != nil {
			cmn.PanicSanity(resBadRemoveValidator.Error())
		}
	}
	saveBonds(store, bonds)
	return abci.OK
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx sdk.Context) (sender sdk.Actor, res abci.Result) {
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 1 {
		return sender, resMissingSignature
	}
	// TODO: ensure senders[0] matches tx.pubkey ...
	return senders[0], abci.OK
}

func getHoldAccount(sender sdk.Actor) sdk.Actor {
	holdAddr := append([]byte{0x00}, sender.Address[1:]...) //shift and prepend a zero
	return sdk.NewActor(stakingModuleName, holdAddr)
}
