package stake

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/ibc"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/go-wire"
)

//nolint
const (
	Name = "stake"
	//Precision = 10e8
	Period2Unbond  uint64 = 30     // how long unbonding takes (measured in blocks)
	Period2ModComm uint64 = 30     // how long modifying a validator commission takes (measured in blocks)
	Inflation      uint   = 1      // inflation in percent (1 to 100)
	CoinDenom      string = "atom" // bondable coin denomination

	queueUnbondTB = iota
	queueCommissionTB
)

// NewHandler returns a new counter transaction processing handler
func NewHandler(feeDenom string) basecoin.Handler {
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

// Handler the transaction processing handler
type Handler struct {
	stack.PassInitState
	stack.PassInitValidate
}

var _ stack.Dispatchable = Handler{} //enforce interface at compile time

// Name - return stake namespace
func (Handler) Name() string {
	return Name
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, _ basecoin.Checker) (res basecoin.CheckResult, err error) {
	err = checkTx(ctx, tx)
	return
}
func checkTx(ctx basecoin.Context, tx basecoin.Tx) (err error) {
	err = tx.Unwrap().ValidateBasic()
	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, dispatch basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	err = checkTx(ctx, tx)
	if err != nil {
		return
	}

	//start by processing the unbonding queue
	height := ctx.BlockHeight()
	err = processQueueUnbond(ctx, store, height, dispatch)
	if err != nil {
		return
	}
	err = processQueueModComm(ctx, store, height)
	if err != nil {
		return
	}
	err = processValidatorRewards(ctx, store, height, dispatch)
	if err != nil {
		return
	}

	//now actually run the transaction
	unwrap := tx.Unwrap()
	var abciRes abci.Result
	switch txType := unwrap.(type) {
	case TxBond:
		abciRes = runTxBond(ctx, store, txType, dispatch)
	case TxUnbond:
		abciRes = runTxUnbond(ctx, store, txType, height)
	case TxNominate:
		abciRes = runTxNominate(ctx, store, txType, dispatch)
	case TxModComm:
		abciRes = runTxModComm(ctx, store, txType, height)
	}

	//determine the validator set changes
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return res, err
	}
	res = basecoin.DeliverResult{
		Data:    abciRes.Data,
		Log:     abciRes.Log,
		Diff:    delegateeBonds.ValidatorsDiff(nil), //TODO add the previous validator set instead of nil
		GasUsed: 0,                                  //TODO add gas accounting
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func runTxBond(ctx basecoin.Context, store state.SimpleDB, tx TxBond,
	dispatch basecoin.Deliver) (res abci.Result) {

	// Get amount of coins to bond
	bondCoin := tx.Amount
	bondAmt := bondCoin.Amount
	valAddr := tx.Validator.Address

	switch {
	case bondCoin.Denom != CoinDenom:
		return abci.ErrInternalError.AppendLog("Invalid coin denomination")
	case bondAmt <= 0:
		return abci.ErrInternalError.AppendLog("Amount must be > 0")
	}

	// Get the delegatee bond account
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	_, delegateeBond := delegateeBonds.Get(valAddr)
	if delegateeBond == nil {
		return abci.ErrInternalError.AppendLog("Cannot bond to non-nominated account")
	}

	// Move coins from the deletatee account to the delegatee lock account
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 1 {
		return abci.ErrInternalError.AppendLog("Missing signature")
	}
	sender := senders[0]
	send := coin.NewSendOneTx(sender, delegateeBond.Account, coin.Coins{bondCoin})

	// If the deduction fails (too high), abort the command
	_, err = dispatch.DeliverTx(ctx, store, send)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}

	// Get or create delegator account
	delegatorBonds, err := getDelegatorBonds(store, sender.Address, valAddr)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	if len(delegatorBonds) != 1 {
		delegatorBonds = DelegatorBonds{
			DelegatorBond{
				DelegateeAddr: valAddr,
				BondTokens:    0,
			},
		}
	}

	// Calculate amount of bond tokens to create, based on exchange rate
	//bondTokenAmt := uint64(coinAmount) * Precision / delegateeBond.ExchangeRate
	bondTokens := uint64(bondAmt) / delegateeBond.ExchangeRate
	delegatorBonds[0].BondTokens += bondTokens

	// Save to store
	setDelegateeBonds(store, delegateeBonds)
	setDelegatorBonds(store, sender.Address, valAddr, delegatorBonds)

	return abci.OK
}

func runTxUnbond(ctx basecoin.Context, store state.SimpleDB, tx TxUnbond,
	height uint64) (res abci.Result) {

	bondAmt := uint64(tx.Amount.Amount)
	valAddr := tx.Validator.Address

	if bondAmt <= 0 {
		return abci.ErrInternalError.AppendLog("Unbond amount must be > 0")
	}

	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) != 0 {
		return abci.ErrInternalError.AppendLog("Missing signature")
	}
	sender := senders[0]

	delegatorBonds, err := getDelegatorBonds(store, sender.Address, valAddr)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	if delegatorBonds == nil {
		return abci.ErrBaseUnknownAddress.AppendLog("No bond account for this (address, validator) pair")
	}
	_, delegatorBond := delegatorBonds.Get(valAddr)
	if delegatorBond == nil {
		return abci.ErrInternalError.AppendLog("Delegator does not contain delegatee bond")
	}

	// subtract bond tokens from bond account
	if delegatorBond.BondTokens < bondAmt {
		return abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")
	}
	delegatorBond.BondTokens -= bondAmt
	//New exchange rate = (new number of bonded atoms)/ total number of bondTokens for validator
	//delegateeBond.ExchangeRate := uint64(bondAmt) / bondTokens

	if delegatorBond.BondTokens == 0 {
		removeDelegatorBonds(store, sender.Address, valAddr)
	} else {
		setDelegatorBonds(store, sender.Address, valAddr, delegatorBonds)
	}

	// subtract tokens from bond value
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	bvIndex, delegateeBond := delegateeBonds.Get(valAddr)
	if delegatorBond == nil {
		return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
	}
	delegateeBond.TotalBondTokens -= bondAmt
	if delegateeBond.TotalBondTokens == 0 {
		delegateeBonds.Remove(bvIndex)
	}
	setDelegateeBonds(store, delegateeBonds)
	// TODO Delegatee bonds?

	// add unbond record to queue
	queueElem := QueueElemUnbond{
		QueueElem: QueueElem{
			DelegateeAddr: valAddr,
			HeightAtInit:  height, // will unbond at `height + Period2Unbond`
		},
		Account:    sender,
		BondTokens: bondAmt,
	}
	queue, err := LoadQueue(queueUnbondTB, store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

func runTxNominate(ctx basecoin.Context, store state.SimpleDB, tx TxNominate,
	dispatch basecoin.Deliver) (res abci.Result) {

	// Create bond value object
	delegateeBond := DelegateeBond{
		DelegateeAddr: tx.Validator.Address,
		Commission:    tx.Commission,
		ExchangeRate:  1, // * Precision,
	}

	// Bond the tokens
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) == 0 {
		return abci.ErrInternalError.AppendLog("Missing signature")
	}
	send := coin.NewSendOneTx(senders[0], delegateeBond.Account, coin.Coins{tx.Amount})
	_, err := dispatch.DeliverTx(ctx, store, send)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
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

//TODO Update logic
func runTxModComm(ctx basecoin.Context, store state.SimpleDB, tx TxModComm,
	height uint64) (res abci.Result) {

	// Retrieve the record to modify
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	delegateeAddr := tx.Validator.Address
	_, delegateeBond := delegateeBonds.Get(delegateeAddr)
	if delegateeBond == nil {
		return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
	}

	// Add the commission modification the queue
	queueElem := QueueElemModComm{
		QueueElem: QueueElem{
			DelegateeAddr: tx.Validator.Address,
			HeightAtInit:  height, // will unbond at `height + Period2Unbond`
		},
		Commission: tx.Commission,
	}
	queue, err := LoadQueue(queueCommissionTB, store)
	if err != nil {
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// Process all unbonding for the current block, note that the unbonding amounts
//   have already been subtracted from the bond account when they were added to the queue
func processQueueUnbond(ctx basecoin.Context, store state.SimpleDB,
	height uint64, dispatch basecoin.Deliver) error {
	queue, err := LoadQueue(queueUnbondTB, store)
	if err != nil {
		return err
	}

	//Get the peek unbond record from the queue
	var unbond QueueElemUnbond
	getUnbond := func() error {
		unbondBytes := queue.Peek()
		return wire.ReadBinaryBytes(unbondBytes, unbond)
	}
	err = getUnbond()
	if err != nil {
		return err
	}

	for unbond.DelegateeAddr != nil && height-unbond.HeightAtInit > Period2Unbond {
		queue.Pop()

		// send unbonded coins to queue account, based on current exchange rate
		delegateeBonds, err := getDelegateeBonds(store)
		if err != nil {
			return err
		}
		_, delegateeBond := delegateeBonds.Get(unbond.DelegateeAddr)
		if delegateeBond == nil {
			return abci.ErrInternalError.AppendLog("Delegatee does not exist for that address")
		}
		coinAmount := unbond.BondTokens * delegateeBond.ExchangeRate // / Precision
		payout := coin.Coins{{CoinDenom, int64(coinAmount)}}

		send := coin.NewSendOneTx(delegateeBond.Account, unbond.Account, payout)
		_, err = dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return err
		}

		// get next unbond record
		err = getUnbond()
		if err != nil {
			return err
		}
	}
	return nil
}

// Process all validator commission modification for the current block
func processQueueModComm(ctx basecoin.Context, store state.SimpleDB, height uint64) error {
	queue, err := LoadQueue(queueCommissionTB, store)
	if err != nil {
		return err
	}

	//Get the peek record from the queue
	var commission QueueElemModComm
	getCommission := func() error {
		bytes := queue.Peek()
		return wire.ReadBinaryBytes(bytes, commission)
	}
	err = getCommission()
	if err != nil {
		return err
	}

	for commission.DelegateeAddr != nil && height-commission.HeightAtInit > Period2ModComm {
		queue.Pop()

		// Retrieve, Modify and save the commission
		delegateeBonds, err := getDelegateeBonds(store)
		if err != nil {
			return err
		}
		record, _ := delegateeBonds.Get(commission.DelegateeAddr)
		if err != nil {
			return err
		}
		delegateeBonds[record].Commission = commission.Commission
		setDelegateeBonds(store, delegateeBonds)

		// check the next record in the queue record
		err = getCommission()
		if err != nil {
			return err
		}
	}
	return nil
}

//TODO add processing of the commission
func processValidatorRewards(ctx basecoin.Context, store state.SimpleDB,
	height uint64, dispatch basecoin.Deliver) error {

	// Retrieve the list of validators
	delegateeBonds, err := getDelegateeBonds(store)
	if err != nil {
		return err
	}
	validatorAccounts := delegateeBonds.ValidatorsActors()

	for _, account := range validatorAccounts {

		credit := coin.Coins{{"atom", 10}} //TODO update to relative to the amount of coins held by validator

		creditTx := coin.NewCreditTx(account, credit)
		_, err = dispatch.DeliverTx(ctx, store, creditTx)
		if err != nil {
			return err
		}
	}
	return nil
}
