package stake

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	name = "stake"

	queueUnbondTypeByte = iota
	queueCommissionTypeByte
)

//nolint
var (
	periodUnbonding uint64 = 30     // queue blocks before unbond
	coinDenom       string = "atom" // bondable coin denomination
	maxVal          int    = 100    // maximum number of validators

	maxCommHistory           = NewDecimal(5, -2) // maximum total commission permitted across the queued commission history
	periodCommHistory uint64 = 28800             // 1 day @ 1 block/3 sec

	inflation Decimal = NewDecimal(7, -2) // inflation between (0 to 1)
)

// Name - simply the name TODO do we need name to be unexposed for security?
func Name() string {
	return name
}

// NewHandler returns a new counter transaction processing handler
func NewHandler(feeDenom string) sdk.Handler {
	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
	).
		IBC(ibc.NewMiddleware()).
		Apps(
			roles.NewMiddleware(),
			fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(roles.NewHandler()),
			stack.WrapHandler(ibc.NewHandler()),
		)
}

// Handler - the transaction processing handler
type Handler struct {
	stack.PassInitValidate
	validatorSet []*abci.Validator
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
	return "", initState(module, key, value)
}

//separated for testing
func initState(module, key, value string) (err error) {
	if module != name {
		return errors.ErrUnknownModule(module)
	}
	switch key {
	case "unbond_period":
		period, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("unbond period must be int, Error: %v", err.Error())
		}
		periodUnbonding = uint64(period)
	case "commhist_period":
		period, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("commhist period must be int, Error: %v", err.Error())
		}
		periodCommHistory = uint64(period)
	case "maxval":
		maxVal, err = strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("maxval must be int, Error: %v", err.Error())
		}
	case "bond_coin":
		coinDenom = value
	}
	return errors.ErrUnknownKey(key)
}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (res sdk.CheckResult, err error) {
	err = checkTx(ctx, tx)

	return
}
func checkTx(ctx sdk.Context, tx sdk.Tx) (err error) {
	err = tx.Unwrap().ValidateBasic()
	return
}

// Tick - Called every block even if no transaction, process all queues and validator rewards
func (h Handler) Tick(ctx sdk.Context, height uint64, store state.SimpleDB, dispatch sdk.Deliver) error {

	sendCoins := func(sender, receiver sdk.Actor, amount coin.Coins) (err error) {
		tx := coin.NewSendOneTx(sender, receiver, amount)
		_, err = dispatch.DeliverTx(ctx, store, tx)
		return
	}
	err := processQueueUnbond(sendCoins, store, height)
	if err != nil {
		return err
	}

	err = processQueueCommHistory(store, height)
	if err != nil {
		return err
	}

	creditCoins := func(receiver sdk.Actor, amount coin.Coins) (err error) {
		tx := coin.NewCreditTx(receiver, amount)
		_, err = dispatch.DeliverTx(ctx, store, tx)
		return err
	}
	err = processValidatorRewards(creditCoins, store, height)
	if err != nil {
		return err
	}
	return nil
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {
	err = checkTx(ctx, tx)
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
		abciRes = runTxUnbond(ctx, store, txType)
	case TxNominate:
		abciRes = runTxNominate(ctx, store, txType, dispatch)
	case TxModComm:
		abciRes = runTxModComm(ctx, store, txType)
	}

	//determine the validator set changes
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return res, err
	}
	diffVal, newVal := delegateeBonds.ValidatorsDiff(h.validatorSet, maxVal)
	h.validatorSet = newVal
	res = sdk.DeliverResult{
		Data:    abciRes.Data,
		Log:     abciRes.Log,
		Diff:    diffVal,
		GasUsed: 0, //TODO add gas accounting
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func runTxBond(ctx sdk.Context, store state.SimpleDB, tx TxBond,
	dispatch sdk.Deliver) (res abci.Result) {

	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 1 {
		return abci.ErrInternalError.AppendLog("Missing signature")
	}
	sender := senders[0]

	sendCoins := func(receiver sdk.Actor, amount coin.Coins) (res abci.Result) {
		// Move coins from the deletator account to the delegatee lock account
		send := coin.NewSendOneTx(sender, receiver, amount)

		// If the deduction fails (too high), abort the command
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return abci.ErrInternalError.AppendLog(err.Error())
		}
		return
	}
	return runTxBondGuts(sendCoins, store, tx, sender)
}

func runTxBondGuts(sendCoins func(receiver sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxBond, sender sdk.Actor) abci.Result {

	// Get amount of coins to bond
	bondCoin := tx.Amount
	bondAmt := NewDecimal(bondCoin.Amount, 1)

	switch {
	case bondCoin.Denom != coinDenom:
		return abci.ErrInternalError.AppendLog("Invalid coin denomination")
	case bondAmt.LTE(Zero):
		return abci.ErrInternalError.AppendLog("Amount must be > 0")
	}

	// Get the delegatee bond account
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	_, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegateeBond == nil {
		return abci.ErrInternalError.AppendLog("Cannot bond to non-nominated account")
	}

	// Move coins from the deletator account to the delegatee lock account
	res := sendCoins(delegateeBond.Account, coin.Coins{bondCoin})
	if res.IsErr() {
		return res
	}

	// Get or create delegator bonds
	delegatorBonds, err := getDelegatorBonds(store, sender)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	if len(delegatorBonds) != 1 {
		delegatorBonds = DelegatorBonds{
			DelegatorBond{
				Delegatee:  tx.Delegatee,
				BondTokens: Zero,
			},
		}
	}

	// Calculate amount of bond tokens to create, based on exchange rate
	bondTokens := bondAmt.Div(delegateeBond.ExchangeRate)
	delegatorBonds[0].BondTokens = delegatorBonds[0].BondTokens.Add(bondTokens)

	// Save to store
	setDelegateeBonds(store, delegateeBonds)
	setDelegatorBonds(store, sender, delegatorBonds)

	return abci.OK
}

func runTxUnbond(ctx sdk.Context, store state.SimpleDB, tx TxUnbond) (res abci.Result) {

	getSender := func() (sender sdk.Actor, err error) {
		senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
		if len(senders) != 0 {
			err = fmt.Errorf("Missing signature")
			return
		}
		sender = senders[0]
		return
	}

	return runTxUnbondGuts(getSender, store, tx, ctx.BlockHeight())
}

func runTxUnbondGuts(getSender func() (sdk.Actor, error), store state.SimpleDB, tx TxUnbond,
	height uint64) (res abci.Result) {

	bondAmt := NewDecimal(tx.Amount.Amount, 1)

	if bondAmt.LTE(Zero) {
		return abci.ErrInternalError.AppendLog("Unbond amount must be > 0")
	}

	sender, err := getSender()
	if err != nil {
		return abci.ErrUnauthorized.AppendLog(err.Error())
	}

	delegatorBonds, err := getDelegatorBonds(store, sender)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	if delegatorBonds == nil {
		return abci.ErrBaseUnknownAddress.AppendLog("No bond account for this (address, validator) pair")
	}
	_, delegatorBond := delegatorBonds.Get(tx.Delegatee)
	if delegatorBond == nil {
		return abci.ErrInternalError.AppendLog("Delegator does not contain delegatee bond")
	}

	// subtract bond tokens from bond account
	if delegatorBond.BondTokens.LT(bondAmt) {
		return abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")
	}
	delegatorBond.BondTokens = delegatorBond.BondTokens.Sub(bondAmt)
	//New exchange rate = (new number of bonded atoms)/ total number of bondTokens for validator
	//delegateeBond.ExchangeRate := uint64(bondAmt) / bondTokens

	if delegatorBond.BondTokens.Equal(Zero) {
		removeDelegatorBonds(store, sender)
	} else {
		setDelegatorBonds(store, sender, delegatorBonds)
	}

	// subtract tokens from bond value
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	bvIndex, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegatorBond == nil {
		return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
	}
	delegateeBond.TotalBondTokens = delegateeBond.TotalBondTokens.Sub(bondAmt)
	if delegateeBond.TotalBondTokens.Equal(Zero) {
		delegateeBonds.Remove(bvIndex)
	}
	setDelegateeBonds(store, delegateeBonds)
	// TODO Delegatee bonds?

	// add unbond record to queue
	queueElem := QueueElemUnbond{
		QueueElem: QueueElem{
			Delegatee:    tx.Delegatee,
			HeightAtInit: height, // will unbond at `height + periodUnbonding`
		},
		Account:    sender,
		BondTokens: bondAmt,
	}
	queue, err := LoadQueue(queueUnbondTypeByte, store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

func runTxNominate(ctx sdk.Context, store state.SimpleDB, tx TxNominate,
	dispatch sdk.Deliver) (res abci.Result) {

	bondCoins := func(bondAccount sdk.Actor, amount coin.Coins) abci.Result {
		senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
		if len(senders) == 0 {
			return abci.ErrInternalError.AppendLog("Missing signature")
		}
		send := coin.NewSendOneTx(senders[0], bondAccount, amount)
		_, err := dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return abci.ErrInternalError.AppendLog(err.Error())
		}
		return abci.OK
	}
	return runTxNominateGuts(bondCoins, store, tx)
}

func runTxNominateGuts(bondCoins func(bondAccount sdk.Actor, amount coin.Coins) abci.Result,
	store state.SimpleDB, tx TxNominate) (res abci.Result) {

	// Create bond value object
	delegateeBond := DelegateeBond{
		Delegatee:    tx.Nominee,
		Commission:   tx.Commission,
		ExchangeRate: One,
	}

	// Bond the coins to the bond account
	res = bondCoins(delegateeBond.Account, coin.Coins{tx.Amount})
	if res.IsErr() {
		return res
	}

	// Append and store
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	delegateeBonds = append(delegateeBonds, delegateeBond)
	setDelegateeBonds(store, delegateeBonds)

	return abci.OK
}

func runTxModComm(ctx sdk.Context, store state.SimpleDB, tx TxModComm) (res abci.Result) {
	return runTxModCommGuts(store, tx, ctx.BlockHeight())
}

func runTxModCommGuts(store state.SimpleDB, tx TxModComm, height uint64) (res abci.Result) {

	// Retrieve the record to modify
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	record, delegateeBond := delegateeBonds.Get(tx.Delegatee)
	if delegateeBond == nil {
		return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
	}

	// Determine that the amount of change proposed is permissible according to the queue change amount
	queue, err := LoadQueue(queueCommissionTypeByte, store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}

	// First determine the sum of the changes in the queue
	// NOTE optimization opportunity: currently need to iterate through all records to pick out
	// the relevant records for the delegatee modifying their commission... Alternative would be to
	// have an array of all the commission change queues and process them individually in Tick
	// which would allow us to only grab one queue for here

	sumModCommQueue := Zero
	valuesBytes := queue.GetAll()
	for _, modCommBytes := range valuesBytes {

		var modComm QueueElemModComm
		err = wire.ReadBinaryBytes(modCommBytes, modComm)
		if err != nil {
			return abci.ErrInternalError.AppendLog(err.Error()) //should never occur under normal operation
		}
		if modComm.Delegatee.Equals(tx.Delegatee) {
			sumModCommQueue.Add(modComm.CommChange)
		}
	}

	commChange := tx.Commission.Sub(delegateeBond.Commission)
	if sumModCommQueue.Add(commChange).GT(maxCommHistory) {
		return abci.ErrUnauthorized.AppendLog(
			fmt.Sprintf("proposed change is greater than is permissible. \n"+
				"The maximum change is %v per %v blocks, the your proposed change will "+
				"bring your queued change to %v", maxCommHistory, periodCommHistory, commChange))
	}

	// Execute the change and add the change amount to the queue
	// Retrieve, Modify and save the commission
	delegateeBonds[record].Commission = tx.Commission
	setDelegateeBonds(store, delegateeBonds)

	// Add the commission modification the queue
	queueElem := QueueElemModComm{
		QueueElem: QueueElem{
			Delegatee:    tx.Delegatee,
			HeightAtInit: height,
		},
		CommChange: tx.Commission,
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// Process all unbonding for the current block, note that the unbonding amounts
//   have already been subtracted from the bond account when they were added to the queue
func processQueueUnbond(sendCoins func(sender, receiver sdk.Actor, amount coin.Coins) error,
	store state.SimpleDB, height uint64) error {
	queue, err := LoadQueue(queueUnbondTypeByte, store)
	if err != nil {
		return err
	}

	//Get the peek unbond record from the queue
	var unbond QueueElemUnbond
	unbondBytes := queue.Peek()
	err = wire.ReadBinaryBytes(unbondBytes, unbond)
	if err != nil {
		return err
	}

	for unbond.Delegatee.Address != nil && height-unbond.HeightAtInit > periodUnbonding {
		queue.Pop()

		// send unbonded coins to queue account, based on current exchange rate
		delegateeBonds, err := getDelegateeBonds(store)
		if err != nil {
			return err
		}
		_, delegateeBond := delegateeBonds.Get(unbond.Delegatee)
		if delegateeBond == nil {
			return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
		}
		coinAmount := unbond.BondTokens.Mul(delegateeBond.ExchangeRate)
		payout := coin.Coins{{coinDenom, coinAmount.IntPart()}} //TODO here coins must also be decimal!!!!

		err = sendCoins(delegateeBond.Account, unbond.Account, payout)
		if err != nil {
			return err
		}

		// get next unbond record
		unbondBytes := queue.Peek()
		err = wire.ReadBinaryBytes(unbondBytes, unbond)
		if err != nil {
			return err
		}
	}
	return nil

}

// Process all validator commission modification for the current block
func processQueueCommHistory(store state.SimpleDB, height uint64) error {
	queue, err := LoadQueue(queueCommissionTypeByte, store)
	if err != nil {
		return err
	}

	//Get the peek record from the queue
	var commission QueueElemModComm
	bytes := queue.Peek()
	err = wire.ReadBinaryBytes(bytes, commission)
	if err != nil {
		return err
	}

	for commission.Delegatee.Address != nil && height-commission.HeightAtInit > periodCommHistory {
		queue.Pop()

		// check the next record in the queue record
		bytes := queue.Peek()
		err = wire.ReadBinaryBytes(bytes, commission)
		if err != nil {
			return err
		}
	}
	return nil
}

//TODO add processing of the commission
func processValidatorRewards(creditAcc func(receiver sdk.Actor, amount coin.Coins) error, store state.SimpleDB, height uint64) error {

	// Retrieve the list of validators
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return err
	}
	validatorAccounts := delegateeBonds.ValidatorsActors(maxVal)

	for _, account := range validatorAccounts {

		credit := coin.Coins{{"atom", 10}} //TODO update to relative to the amount of coins held by validator

		err = creditAcc(account, credit)
		if err != nil {
			return err
		}
	}
	return nil
}
