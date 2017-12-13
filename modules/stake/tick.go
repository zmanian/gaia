package stake

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

// Tick - called at the end of every block
func Tick(store state.SimpleDB, ctx sdk.Context, dispatch sdk.Deliver) (change []*abci.Validator, err error) {

	//process the unbonding queue
	params := loadParams(store)
	height := ctx.BlockHeight()

	ctxUnbonding := ctx.WithPermissions(params.HoldUnbonding)
	transfer := coinSender{
		store:    store,
		dispatch: dispatch,
		ctx:      ctxUnbonding,
	}.transferFn

	p := processQueueUnbond{store, params, transfer}
	// Process Unbonding Queue
	processQueueUnbond(store, height, params.UnbondingPeriod)

	return UpdateValidatorSet(store)
}

//______________________________________________________________________________________

// Process all unbonding for the current block, note that the unbonding amounts
//   have already been subtracted from the bond account when they were added to the queue

//type QueueElem struct {
//Candidate  crypto.PubKey
//InitHeight uint64 // when the queue was initiated
//}
//// QueueElemUnbondDelegation - TODO
//type QueueElemUnbondDelegation struct {
//QueueElem
//Payout          sdk.Actor // account to pay out to
//Amount          uint64    // amount of shares which are unbonding
//StartSlashRatio uint64    // old candidate slash ratio at start of re-delegation
//}

type processQueue interface {
	process(height int64) error
}

type processQueueUnbond struct {
	store    state.SimpleDB
	params   Params
	transfer transferFn
}

func (p processQueueUnbond) process(height int64) error {
	queue, err := LoadQueue(queueUnbondTypeByte, store)
	if err != nil {
		return err
	}

	// loop through the list of unbonds in the queue starting from the queue peek
	var unbond QueueElemUnbondDelegation
	getNext := func() error {
		unbondBytes := queue.Peek()
		if unbondBytes == nil { //exit if queue empty
			return nil
		}
		err = wire.ReadBinaryBytes(unbondBytes, &unbond)
		if err != nil {
			return err
		}
	}
	if err = getNext(); err != nil {
		return err
	}
	for !unbond.Candidate.Empty() && unbond.InitHeight+unbondingPeriod <= height {
		queue.Pop()

		// get the candidate to unbond from
		candidate, err := loadCandidate(store, unbond.QueueElem.Candidate)
		if err != nil {
			return err
		}

		// check the slashing ratio to see if any slashings need to be applied here
		unbondQuantity := unbond.Amount
		if unbond.QueueElem.SlashRatio != candidate.SlashRatio {
			// TODO confirm math here/ test
			unbondQuantity = (unbond.Amount * unbond.QueueElem.SlashRatio) / candidate.SlashRatio
		}

		// send unbonded coins to queue account, based on current exchange rate
		amount := coin.Coins{{p.params.AllowedBondDenom, unbond.unbondQuantity}}
		err = p.transfer(delegateeBond.HoldUnbonding, unbond.Payout, amount)
		if err != nil {
			return err
		}

		// get next unbond record
		if err = getNext(); err != nil {
			return err
		}
	}
	return nil
}
