package stake

import (
	"fmt"
	"strconv"

	abci "github.com/tendermint/abci/types"
	bcs "github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

const PRECISION = 10e8

// Plugin is a proof-of-stake plugin for Basecoin
type Plugin struct {
	UnbondingPeriod uint64 // how long unbonding takes (measured in blocks)
	CoinDenom       string // bondable coin denomination
	height          uint64 // current block height
}

// Name returns the name of the stake plugin
func (sp Plugin) Name() string {
	return "stake"
}

// SetOption from ABCI
func (sp Plugin) SetOption(store types.KVStore, key string, value string) (log string) {
	if key == "unbondingPeriod" {
		var err error
		sp.UnbondingPeriod, err = strconv.ParseUint(value, 10, 64)
		if err != nil {
			panic(fmt.Errorf("Could not parse int: '%s'", value))
		}
		return
	}
	if key == "coinDenom" {
		sp.CoinDenom = value
		return
	}
	panic(fmt.Errorf("Unknown option key '%s'", key))
}

// RunTx from ABCI
func (sp Plugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	var tx Tx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	switch tx_ := tx.(type) {
	case BondTx:
		return sp.runBondTx(tx_, store, ctx)
	case UnbondTx:
		return sp.runUnbondTx(tx_, store, ctx)
	}
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
	bondAccount.Sequence += 1

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

// InitChain from ABCI
func (sp Plugin) InitChain(store types.KVStore, vals []*abci.Validator) {
	bondValues := loadBondValues(store)
	for _, val := range vals {
		// create bond value object
		bondValue := BondValue{
			Total:           val.Power,
			ExchangeRate:    1 * PRECISION,
			ValidatorPubKey: val.PubKey,
		}
		bondValues = append(bondValues, bondValue)

		// TODO: create bond account so initial bonds are unbondable
	}
	storeBondValues(store, bondValues)
}

// BeginBlock from ABCI
func (sp Plugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
	// process unbonds which have finished
	sp.height = header.GetHeight()
	queue := loadUnbondQueue(store)
	unbond := queue.Peek()
	for unbond != nil && sp.height-unbond.Height > sp.UnbondingPeriod {
		queue.Pop()

		// add unbonded coins to basecoin account, based on current exchange rate
		_, bondValue := loadBondValues(store).Get(unbond.ValidatorPubKey)
		coinAmount := unbond.BondAmount * bondValue.ExchangeRate / PRECISION
		account := bcs.GetAccount(store, unbond.Address)
		payout := makeCoin(coinAmount, sp.CoinDenom)
		account.Balance = account.Balance.Plus(payout)
		bcs.SetAccount(store, unbond.Address, account)

		// get next unbond record
		unbond = queue.Peek()
	}
}

// EndBlock from ABCI
func (sp Plugin) EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock {
	sp.height = height + 1
	bondValues := loadBondValues(store)
	return abci.ResponseEndBlock{bondValues.Validators()}
}

// convenience function, returns a []Coin with a single coin
func makeCoin(amount uint64, denom string) types.Coins {
	return types.Coins{
		types.Coin{
			Denom:  denom,
			Amount: int64(amount),
		},
	}
}
