package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

// Tick - called at the end of every block
func Tick(store state.SimpleDB, ctx sdk.Context) (change []*abci.Validator, err error) {

	//process the unbonding queue

	return UpdateValidatorSet(store)
}

//______________________________________________________________________________________

// Process all unbonding for the current block, note that the unbonding amounts
//   have already been subtracted from the bond account when they were added to the queue

func processQueueUnbond(store state.SimpleDB, height int64, periodUnbonding int64) error {
	queue, err := LoadQueue(queueUnbondTypeByte, store)
	if err != nil {
		return err
	}

	//Get the peek unbond record from the queue
	var unbond QueueElemUnbond
	unbondBytes := queue.Peek()
	if unbondBytes == nil { //exit if queue empty
		return nil
	}
	err = wire.ReadBinaryBytes(unbondBytes, &unbond)
	if err != nil {
		return err
	}

	// here a few variables used in the loop
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return err
	}

	for !unbond.Delegatee.Empty() && unbond.HeightAtInit+periodUnbonding <= height {
		queue.Pop()

		// send unbonded coins to queue account, based on current exchange rate
		_, delegateeBond := delegateeBonds.Get(unbond.Delegatee)
		if delegateeBond == nil {
			// This error should never really happen
			return fmt.Errorf("Attempted to retrieve a non-existent delegatee during validator reward processing")
		}
		coinAmount := unbond.BondTokens.Mul(delegateeBond.ExchangeRate)
		payout := coin.Coins{{bondDenom, coinAmount.IntPart()}} //TODO update coins to decimal

		err = sendCoins(delegateeBond.HoldAccount, unbond.Account, payout)
		if err != nil {
			return err
		}

		// get next unbond record
		unbondBytes := queue.Peek()
		if unbondBytes == nil { //exit if queue empty
			return nil
		}
		err = wire.ReadBinaryBytes(unbondBytes, &unbond)
		if err != nil {
			return err
		}
	}
	return nil

}
