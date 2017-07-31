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

// Tx
//--------------------------------------------------------------------------------

// register the tx type with it's validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	Name      = "stake"
	ByteTx    = 0x55
	TypeTx    = NameCounter + "/count"
	PRECISION = 10e8
)

func init() {
	basecoin.TxMapper.RegisterImplementation(Tx{}, TypeTx, ByteTx)
}

// Tx - struct for all counter transactions
type Tx struct {
	Valid bool       `json:"valid"`
	Fee   coin.Coins `json:"fee"`
}

// NewTx - return a new counter transaction struct wrapped as a basecoin transaction
func NewTx(valid bool, fee coin.Coins) basecoin.Tx {
	return Tx{
		Valid: valid,
		Fee:   fee,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx, used to satisfy the XXX interface
func (c Tx) Wrap() basecoin.Tx {
	return basecoin.Tx{TxInner: c}
}

// ValidateBasic just makes sure the Fee is a valid, non-negative value
func (c Tx) ValidateBasic() error {
	if !c.Fee.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !c.Fee.IsNonnegative() {
		return coin.ErrInvalidCoins()
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////

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

// Handler the counter transaction processing handler
type Handler struct {
	stack.NopOption
}

var _ stack.Dispatchable = Handler{}

// Name - return stake namespace
func (Handler) Name() string {
	return Name
}

// AssertDispatcher - placeholder to satisfy XXX
func (Handler) AssertDispatcher() {}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, _ basecoin.Checker) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx basecoin.Context, store state.SimpleDB,
	tx basecoin.Tx, dispatch basecoin.Deliver) (res basecoin.Result, err error) {
	ctr, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}
	// note that we don't assert this on CheckTx (ValidateBasic),
	// as we allow them to be writen to the chain
	if !ctr.Valid {
		return res, ErrInvalidCounter()
	}

	// handle coin movement.... like, actually decrement the other account
	if !ctr.Fee.IsZero() {
		// take the coins and put them in out account!
		senders := ctx.GetPermissions("", auth.NameSigs)
		if len(senders) == 0 {
			return res, errors.ErrMissingSignature()
		}
		in := []coin.TxInput{{Address: senders[0], Coins: ctr.Fee}}
		out := []coin.TxOutput{{Address: StoreActor(), Coins: ctr.Fee}}
		send := coin.NewSendTx(in, out)
		// if the deduction fails (too high), abort the command
		_, err = dispatch.DeliverTx(ctx, store, send)
		if err != nil {
			return res, err
		}
	}

	// update the counter
	state, err := LoadState(store)
	if err != nil {
		return res, err
	}
	state.Counter++
	state.TotalFees = state.TotalFees.Plus(ctr.Fee)
	err = SaveState(store, state)

	return res, err
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (ctr Tx, err error) {
	//ctr, ok := tx.Unwrap().(Tx)
	//if !ok {
	//return ctr, errors.ErrInvalidFormat(TypeTx, tx)
	//}
	//err = ctr.ValidateBasic()
	//if err != nil {
	//return ctr, err
	//}
	//return ctr, nil

	var tx Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	switch txType := tx.(type) {
	case BondTx:
		return sp.runBondTx(txType, store, ctx)
	case UnbondTx:
		return sp.runUnbondTx(txType, store, ctx)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// Plugin is a proof-of-stake plugin for Basecoin
type Plugin struct {
	UnbondingPeriod uint64 // how long unbonding takes (measured in blocks)
	CoinDenom       string // bondable coin denomination
	height          uint64 // current block height
}

func (sp Plugin) runBondTx(tx BondTx, store types.KVStore, ctx types.CallContext) (res abci.Result) {
	if len(ctx.Coins) != 1 {
		return abci.ErrInternalError.AppendLog("Invalid coins")
	}
	if ctx.Coins[0].Denom != sp.CoinDenom {
		return abci.ErrInternalError.AppendLog("Invalid coin denomination")
	}

	// get amount of coins to bond
	coinAmount := ctx.Coins[0].Amount
	if coinAmount <= 0 {
		return abci.ErrInternalError.AppendLog("Amount must be > 0")
	}

	bondAccount := loadBondAccount(store, ctx.CallerAddress, tx.ValidatorPubKey)
	if bondAccount == nil {
		if tx.Sequence != 0 {
			return abci.ErrInternalError.AppendLog("Invalid sequence number")
		}
		// create new account for this (delegator, validator) pair
		bondAccount = &BondAccount{
			Amount:   0,
			Sequence: 0,
		}
	} else if tx.Sequence != (bondAccount.Sequence + 1) {
		return abci.ErrInternalError.AppendLog("Invalid sequence number")
	}

	// add tokens to validator's bond supply
	bondValues := loadBondValues(store)
	_, bondValue := bondValues.Get(tx.ValidatorPubKey)
	if bondValue == nil {
		// first bond for this validator, initialize a new BondValue
		bondValue = &BondValue{
			ValidatorPubKey: tx.ValidatorPubKey,
			Total:           0,
			ExchangeRate:    1 * PRECISION, // starts at one atom per bond token
		}
		bondValues = append(bondValues, *bondValue)
	}
	// calulcate amount of bond tokens to create, based on exchange rate
	bondAmount := uint64(coinAmount) * PRECISION / bondValue.ExchangeRate
	bondValue.Total += bondAmount
	bondAccount.Amount += bondAmount
	bondAccount.Sequence++

	// TODO: special rules for entering validator set

	storeBondValues(store, bondValues)
	storeBondAccount(store, ctx.CallerAddress, tx.ValidatorPubKey, bondAccount)

	return abci.OK
}

func (sp Plugin) runUnbondTx(tx UnbondTx, store types.KVStore, ctx types.CallContext) (res abci.Result) {
	if tx.BondAmount <= 0 {
		return abci.ErrInternalError.AppendLog("Unbond amount must be > 0")
	}

	bondAccount := loadBondAccount(store, ctx.CallerAddress, tx.ValidatorPubKey)
	if bondAccount == nil {
		return abci.ErrBaseUnknownAddress.AppendLog("No bond account for this (address, validator) pair")
	}
	if bondAccount.Amount < tx.BondAmount {
		return abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")
	}

	// subtract tokens from bond account
	bondAccount.Amount -= tx.BondAmount
	if bondAccount.Amount == 0 {
		removeBondAccount(store, ctx.CallerAddress, tx.ValidatorPubKey)
	} else {
		storeBondAccount(store, ctx.CallerAddress, tx.ValidatorPubKey, bondAccount)
	}

	// subtract tokens from bond value
	bondValues := loadBondValues(store)
	bvIndex, bondValue := bondValues.Get(tx.ValidatorPubKey)
	bondValue.Total -= tx.BondAmount
	if bondValue.Total == 0 {
		bondValues.Remove(bvIndex)
	}
	// will get sorted in EndBlock
	storeBondValues(store, bondValues)

	// add unbond record to queue
	unbond := Unbond{
		ValidatorPubKey: tx.ValidatorPubKey,
		BondAmount:      tx.BondAmount,
		Address:         ctx.CallerAddress,
		Height:          sp.height, // unbonds at `height + UnbondingPeriod`
	}
	unbondQueue := loadUnbondQueue(store)
	unbondQueue.Push(unbond)

	return abci.OK
}

//// SetOption from ABCI
//func (sp Plugin) SetOption(store types.KVStore, key string, value string) (log string) {
//if key == "unbondingPeriod" {
//var err error
//sp.UnbondingPeriod, err = strconv.ParseUint(value, 10, 64)
//if err != nil {
//panic(fmt.Errorf("Could not parse int: '%s'", value))
//}
//return
//}
//if key == "coinDenom" {
//sp.CoinDenom = value
//return
//}
//panic(fmt.Errorf("Unknown option key '%s'", key))
//}

//// InitChain from ABCI
//func (sp Plugin) InitChain(store types.KVStore, vals []*abci.Validator) {
//bondValues := loadBondValues(store)
//for _, val := range vals {
//// create bond value object
//bondValue := BondValue{
//Total:           val.Power,
//ExchangeRate:    1 * PRECISION,
//ValidatorPubKey: val.PubKey,
//}
//bondValues = append(bondValues, bondValue)

//// TODO: create bond account so initial bonds are unbondable
//}
//storeBondValues(store, bondValues)
//}

//// BeginBlock from ABCI
//func (sp Plugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {V
//// process unbonds which have finished
//sp.height = header.GetHeight()
//queue := loadUnbondQueue(store)
//unbond := queue.Peek()
//for unbond != nil && sp.height-unbond.Height > sp.UnbondingPeriod {
//queue.Pop()

//// add unbonded coins to basecoin account, based on current exchange rate
//_, bondValue := loadBondValues(store).Get(unbond.ValidatorPubKey)
//coinAmount := unbond.BondAmount * bondValue.ExchangeRate / PRECISION
//account := bcs.GetAccount(store, unbond.Address)
//payout := makeCoin(coinAmount, sp.CoinDenom)
//account.Balance = account.Balance.Plus(payout)
//bcs.SetAccount(store, unbond.Address, account)

//// get next unbond record
//unbond = queue.Peek()
//}
//}

//// EndBlock from ABCI
//func (sp Plugin) EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock {
//sp.height = height + 1
//bondValues := loadBondValues(store)
//return abci.ResponseEndBlock{bondValues.Validators()}
//}
