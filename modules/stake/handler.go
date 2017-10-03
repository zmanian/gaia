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
	name = "stake"
)

//nolint
var (
	//TODO should all these global parameters be moved to the state?
	bondDenom string = "atom" // bondable coin denomination
	maxVal    int    = 100    // maximum number of validators

	// GasAllocations per staking transaction
	costBond   = uint64(20)
	costUnbond = uint64(0)
)

// Name - simply the name TODO do we need name to be unexposed for security?
func Name() string {
	return name
}

// Handler - the transaction processing handler
type Handler struct {
	stack.PassInitValidate
}

var _ stack.Dispatchable = Handler{} // enforce interface at compile time

// Name - return stake namespace
func (Handler) Name() string {
	return name
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

// InitState - set genesis parameters for staking
func (Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	return "", initState(module, key, value, store)
}

//separated for testing
func initState(module, key, value string, store state.SimpleDB) (err error) {
	if module != name {
		return errors.ErrUnknownModule(module)
	}
	switch key {
	case "bond_coin":
		bondDenom = value
	case "maxval",
		"cost_bond",
		"cost_unbond":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}
		switch key {
		case "maxval":
			maxVal = i
		case "cost_bond":
			costBond = uint64(i)
		case "cost_unbond":
			costUnbond = uint64(i)
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

	switch t := tx.Unwrap().(type) {
	case TxBond:
		// TODO could also check for existence of validator here? (already checked in deliverTx)

		//check for suffient funds to send
		sender, abciRes := getSingleSender(ctx)
		if abciRes.IsErr() {
			return res, abciRes
		}
		_, err := coin.CheckCoins(store, sender, coin.Coins{t.Amount}.Negative())
		if err != nil {
			return res, err
		}
		return sdk.NewCheck(costBond, ""), nil

	case TxUnbond:
		// TODO check for sufficient validator tokens to unbond here? (already checked in deliverTx)
		return sdk.NewCheck(costUnbond, ""), nil
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// Tick - Called every block even if no transaction,
//   process all queues, validator rewards, and calculate the validator set difference
func (h Handler) Tick(ctx sdk.Context, height uint64, store state.SimpleDB,
	dispatch sdk.Deliver) (diffVal []*abci.Validator, err error) {

	// Determine the validator set changes
	validatorBonds, err := loadValidatorBonds(store)
	if err != nil {
		return
	}
	startVal := validatorBonds.GetValidators()
	validatorBonds.UpdateVotingPower(store)
	newVal := validatorBonds.GetValidators()
	diffVal = ValidatorsDiff(startVal, newVal)

	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	_, err = h.CheckTx(ctx, store, tx, nil)
	if err != nil {
		return
	}

	//Run the transaction
	unwrap := tx.Unwrap()
	var abciRes abci.Result
	switch txType := unwrap.(type) {
	case TxBond:
		abciRes = runTxBond(ctx, store, txType, dispatch)
	case TxUnbond:
		abciRes = runTxUnbond(ctx, store, txType, dispatch)
	}

	res = sdk.DeliverResult{
		Data: abciRes.Data,
		Log:  abciRes.Log,
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func runTxBond(ctx sdk.Context, store state.SimpleDB, tx TxBond,
	dispatch sdk.Deliver) (res abci.Result) {

	sender, res := getSingleSender(ctx)
	if res.IsErr() {
		return res
	}

	return runTxBondGuts(getSendFunc(ctx, store, dispatch), store, tx, sender)
}

// sendCoins is a helper function
func getSendFunc(ctx sdk.Context, store state.SimpleDB,
	dispatch sdk.Deliver) func(sender, receiver sdk.Actor, amount coin.Coins) (res abci.Result) {

	return func(sender, receiver sdk.Actor, amount coin.Coins) (res abci.Result) {
		// Move coins from the deletator account to the validator lock account
		send := coin.NewSendOneTx(sender, receiver, amount)

		// If the deduction fails (too high), abort the command
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return abci.ErrInsufficientFunds.AppendLog(err.Error())
		}
		return
	}
}

func getSingleSender(ctx sdk.Context) (sender sdk.Actor, res abci.Result) {
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 1 {
		return sender, resMissingSignature
	}
	return senders[0], abci.OK
}

func runTxBondGuts(sendCoins func(sender, receiver sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxBond, sender sdk.Actor) abci.Result {

	// Get amount of coins to bond
	bondCoin := tx.Amount
	bondAmt := uint64(bondCoin.Amount)

	// Get the validator bond account
	validatorBonds, err := loadValidatorBonds(store)
	if err != nil {
		return resErrLoadingValidators(err)
	}
	i, validatorBond := validatorBonds.Get(tx.Validator)
	if validatorBond == nil {
		return resBondNotNominated
	}

	// Move coins from the delegator account to the validator lock account
	res := sendCoins(sender, validatorBond.HoldAccount, coin.Coins{bondCoin})
	if res.IsErr() {
		return res
	}

	validatorBonds[i].BondedTokens += bondAmt

	// NOTE the exchange rate does not change due to the bonding process,
	// all tokens are assigned AT the exchange rate
	// (but does change when validator rewards are processed)

	// Save to store
	saveValidatorBonds(store, validatorBonds)

	return abci.OK
}

func runTxUnbond(ctx sdk.Context, store state.SimpleDB, tx TxUnbond,
	dispatch sdk.Deliver) (res abci.Result) {

	getSender := func() (sender sdk.Actor, res abci.Result) {
		res = abci.OK
		senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
		if len(senders) != 0 {
			res = resMissingSignature
			return
		}
		sender = senders[0]
		return
	}

	return runTxUnbondGuts(getSender, getSendFunc(ctx, store, dispatch), store, tx)
}

func runTxUnbondGuts(getSender func() (sdk.Actor, abci.Result),
	sendCoins func(sender, receiver sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxUnbond) (res abci.Result) {

	bondAmt := uint64(tx.Amount.Amount)

	sender, res := getSender()
	if res.IsErr() {
		return res
	}

	//get validator bond
	validatorBonds, err := loadValidatorBonds(store)
	if err != nil {
		return resErrLoadingValidators(err)
	}
	bvIndex, validatorBond := validatorBonds.Get(tx.Validator)
	if validatorBond == nil {
		return resNoValidatorForAddress
	}

	// subtract tokens from validatorBonds
	validatorBond.BondedTokens -= bondAmt
	if validatorBond.BondedTokens == 0 {
		validatorBonds.Remove(bvIndex)
	}

	saveValidatorBonds(store, validatorBonds)

	// send unbonded coins to queue account, based on current exchange rate
	res = sendCoins(validatorBond.HoldAccount, sender, coin.Coins{tx.Amount})
	if res.IsErr() {
		return res
	}

	return abci.OK
}
