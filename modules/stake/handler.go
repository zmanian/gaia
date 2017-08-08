package stake

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/ibc"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

//nolint
const (
	Name = "stake"
	//Precision = 10e8
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
	stack.NopOption
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
	_, err = checkTx(ctx, tx)
	return
}
func checkTx(ctx basecoin.Context, tx basecoin.Tx) (ctr basecoin.Tx, err error) {
	ctr, ok := tx.Unwrap().(Tx)
	if !ok {
		return ctr, errors.ErrInvalidFormat(TypeTx, tx)
	}
	err = ctr.ValidateBasic()
	if err != nil {
		return ctr, err
	}
	return ctr, nil
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, dispatch basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	ctr, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	//start by processing the unbonding queue
	height := ctx.BlockHeight()
	processQueueUnbond(store, height)
	processQueueModComm(store, height, dispatch)

	//now actually run the transaction
	var tx Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	var abciRes abci.Result
	switch txType := tx.(type) {
	case TxBond:
		abciRes, err = sp.runTxBond(txType, store, ctx)
	case TxUnbond:
		abciRes, err = sp.runTxUnbond(txType, store, ctx, height)
	case TxNominate:
		abciRes, err = sp.runTxNominate(txType, store, ctx)
	case TxModComm:
		abciRes, err = sp.runTxModComm(txType, store, ctx, height)
	}

	//determine the validator set changes
	delegatorBonds := getDelegatorBonds(store)
	res = basecoin.DeliverResult{
		Data:    abciRes.Data,
		Log:     abciRes.Log,
		Diff:    delegatorBonds.Validators(), //FIXME this is the full set, need to just use the diff
		GasUsed: 0,                           //TODO add gas accounting
	}

	return res, err
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Plugin is a proof-of-stake plugin for Basecoin
type Plugin struct {
	Period2Unbond  uint64 // how long unbonding takes (measured in blocks)
	Period2ModComm uint64 // how long modifying a validator commission takes (measured in blocks)
	CoinDenom      string // bondable coin denomination
}

func (sp Plugin) runTxBond(tx TxBond, store state.SimpleDB, ctx types.CallContext) (res abci.Result) {
	if len(ctx.Coins) != 1 {
		return abci.ErrInternalError.AppendLog("Invalid coins")
	}
	if ctx.Coins[0].Denom != sp.CoinDenom {
		return abci.ErrInternalError.AppendLog("Invalid coin denomination")
	}

	// Get amount of coins to bond
	bondCoins := ctx.Coins[0].Amount
	if bondCoins.Amount <= 0 {
		return abci.ErrInternalError.AppendLog("Amount must be > 0")
	}

	// Get the delegator bond account
	delegatorBonds := getDelegatorBonds(store)
	_, delegatorBond := delegatorBonds.Get(tx.ValidatorPubKey)
	if delegatorBond == nil {
		return abci.ErrInternalError.AppendLog("Cannot bond to non-nominated account")
	}

	// Move coins from the deletatee account to the delegator lock account
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) == 0 {
		return res, errors.ErrMissingSignature()
	}
	send := coin.NewSendOneTx(senders[0], delegatorBond.Account, bondCoins)

	// If the deduction fails (too high), abort the command
	_, err = dispatch.DeliverTx(ctx, store, send)
	if err != nil {
		return res, err
	}

	// Get or create delegatee account
	delegateeBonds := getDelegateeBonds(store, ctx.CallerAddress, tx.ValidatorPubKey)
	if delegateeBonds == nil {
		delegateeBonds = &DelegateeBond{
			ValidatorPubKey: tx.ValidatorPubKey,
			Amount:          0,
		}
	}

	// Calculate amount of bond tokens to create, based on exchange rate
	//bondAmount := uint64(coinAmount) * Precision / delegatorBond.ExchangeRate
	bondAmount := uint64(coinAmount) / delegatorBond.ExchangeRate
	delegateeBonds.Amount += bondAmount

	// Save to store
	setDelegatorBonds(store, delegatorBonds)
	setDelegateeBonds(store, ctx.CallerAddress, tx.ValidatorPubKey, delegateeBonds)

	return abci.OK
}

func (sp Plugin) runTxUnbond(tx TxUnbond, store state.SimpleDB,
	ctx types.CallContext, height uint64) (res abci.Result) {
	if tx.BondAmount <= 0 {
		return abci.ErrInternalError.AppendLog("Unbond amount must be > 0")
	}

	delegateeBonds := getDelegateeBonds(store, ctx.CallerAddress, tx.ValidatorPubKey)
	if delegateeBonds == nil {
		return abci.ErrBaseUnknownAddress.AppendLog("No bond account for this (address, validator) pair")
	}
	if delegateeBonds.Amount < tx.BondAmount {
		return abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")
	}

	// subtract tokens from bond account
	delegateeBonds.Amount -= tx.BondAmount
	if delegateeBonds.Amount == 0 {
		removeDelegateeBonds(store, ctx.CallerAddress, tx.ValidatorPubKey)
	} else {
		setDelegateeBonds(store, ctx.CallerAddress, tx.ValidatorPubKey, delegateeBonds)
	}

	// subtract tokens from bond value
	delegatorBonds := getDelegatorBonds(store)
	bvIndex, delegatorBond := delegatorBonds.Get(tx.ValidatorPubKey)
	delegatorBond.Total -= tx.BondAmount
	if delegatorBond.Total == 0 {
		delegatorBonds.Remove(bvIndex)
	}
	setDelegatorBonds(store, delegatorBonds)
	// TODO Delegator bonds?

	// add unbond record to queue
	queueElem := QueueElemUnbond{
		QueueElem: QueueElem{
			ValidatorPubKey: tx.ValidatorPubKey,
			HeightAtInit:    height, // will unbond at `height + Period2Unbond`
		},
		Address:    ctx.CallerAddress,
		BondAmount: tx.BondAmount,
	}
	queue := loadQueue(store)
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

func (sp Plugin) runNominate(tx TxNominate, store state.SimpleDB,
	ctx types.CallContext, dispatch basecoin.Deliver) (res abci.Result) {

	// Create bond value object
	delegatorBond := DelegatorBond{
		ValidatorPubKey: tx.PubKey,
		Commission:      tx.Commission,
		ExchangeRate:    1, // * Precision,
	}

	// Bond the tokens
	senders := ctx.GetPermissions("", auth.NameSigs) //XXX does auth need to be checked here?
	if len(senders) == 0 {
		return res, errors.ErrMissingSignature()
	}
	send := coin.NewSendOneTx(senders[0], delegatorBond.Account, tx.Amount)
	_, err = dispatch.DeliverTx(ctx, store, send)
	if err != nil {
		return res, err
	}

	// Append and store
	delegatorBonds := getDelegatorBonds(store)
	delegatorBonds = append(delegatorBonds, delegatorBond)
	setDelegatorBonds(store, delegatorBonds)

	return abci.OK
}

//TODO Update logic
func (sp Plugin) runModComm(tx TxModComm, store state.SimpleDB,
	ctx types.CallContext, height uint64) (res abci.Result) {

	// Retrieve the record to modify
	delegatorBonds, err := getDelegatorBonds(store)
	delegatorBond := delegatorBonds.Get()

	// Add the commission modification the queue
	queueElem := QueueElemModComm{
		QueueElem: QueueElem{
			ValidatorPubKey: tx.ValidatorPubKey,
			HeightAtInit:    height, // will unbond at `height + Period2Unbond`
		},
		Commission: tx.Commission,
	}
	queue := LoadQueue("commission", store)
	bytes := wire.BinaryBytes(queueElem)
	queue.Push(bytes)

	return abci.OK
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// Process all unbonding for the current block, note that the unbonding amounts
//   have already been subtracted from the bond account when they were added to the queue
func (sp Plugin) processQueueUnbond(store state.SimpleDB, height uint64, dispatch basecoin.Deliver) error {
	queue := LoadQueue("unbonding", store)

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

	for unbond != nil && height-unbond.HeightAtInit > sp.Period2Unbond {
		queue.Pop()

		// send unbonded coins to queue account, based on current exchange rate
		delegatorBonds, err := getDelegatorBonds(store)
		if err != nil {
			return err
		}
		_, delegatorBond := delegatorBonds.Get(unbond.ValidatorPubKey)
		coinAmount := unbond.Amount * delegatorBond.ExchangeRate // / Precision
		payout := coin.Coin{sp.CoinDenom, coinAmount}

		send := coin.NewSendOneTx(delegatorBond.Account, unbond.Account, payout)
		_, err = dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return res, err
		}

		// get next unbond record
		err = getUnbond()
		if err != nil {
			return err
		}
	}
}

// Process all validator commission modification for the current block
func (sp Plugin) processQueueModComm(store state.SimpleDB, height uint64) error {
	queue := LoadQueue("commission", store)

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

	for commission != nil && height-commission.HeightAtInit > sp.Period2ModComm {
		queue.Pop()

		// Retrieve, Modify and save the commission
		delegatorBonds, err := getDelegatorBonds(store)
		if err != nil {
			return err
		}
		record, _ := delegatorBonds.Get(commission.ValidatorPubKey)
		if err != nil {
			return err
		}
		delegatorBonds[record].Commission = commission.Commission
		setDelegatorBonds(store, delegatorBonds)

		// check the next record in the queue record
		err = getCommission()
		if err != nil {
			return err
		}
	}
}
