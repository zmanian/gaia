package stake

import (
	"fmt"

	"github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
)

// Contains all of the tick processing to occure every block

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

		err = sendCoins(delegateeBond.Account, unbond.Account, payout)
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

// Process the validator commission history queue
// This function doesn't change the commission rate, the commission rate
// is changed instantaniously when modified, this queue allows for an
// accurate accounting of the recent commission history modifications to
// be held.
//func processQueueCommHistory(store state.SimpleDB, height uint64) error {
func processQueueCommHistory(queue *Queue, height uint64) error {
	//queue, err := LoadQueue(queueCommissionTypeByte, store)
	//if err != nil {
	//return err
	//}

	//Get the peek record from the queue
	var commission QueueElemCommChange
	bytes := queue.Peek()
	if bytes == nil { //exit if queue empty
		return nil
	}
	err := wire.ReadBinaryBytes(bytes, &commission)
	if err != nil {
		return err
	}

	i := 0
	for !commission.Delegatee.Empty() && commission.HeightAtInit+periodUnbonding <= height {
		i++
		queue.Pop()

		// check the next record in the queue record
		bytes := queue.Peek()
		if bytes == nil { //exit if queue empty
			return nil
		}
		err = wire.ReadBinaryBytes(bytes, &commission)
		if err != nil {
			return err
		}
	}
	//panic(fmt.Sprintf("debug %v\n", queue.length()))
	//panic(fmt.Sprintf("debug %v\n", queue.Peek()))
	return nil
}

// process all the validator rewards, and update the exchange rate for all validators
// NOTE this function assumes that the voting power for all the validators has
// already been appropriately updated, thus the total voting power must be passed in.
func processValidatorRewards(creditAcc func(receiver sdk.Actor, amount coin.Coins) error,
	store state.SimpleDB, height uint64, totalVotingPower Decimal) error {

	// Retrieve the list of validators
	delegateeBonds, err := loadDelegateeBonds(store)
	if err != nil {
		return err
	}

	// Get the total atoms
	totalAtoms, err := loadAtomSupply(store)
	if err != nil {
		return err
	}

	//Rewards per power
	rewardPerPower := (totalAtoms.Div(totalVotingPower)).Mul(inflationPerReward)

	for i, validator := range delegateeBonds {

		vp := validator.VotingPower
		if vp.Equal(Zero) { //is sorted so at first zero no more validators
			break
		}

		rewardCoins := vp.Mul(rewardPerPower)
		totalAtoms = totalAtoms.Add(rewardCoins)
		credit := coin.Coins{{bondDenom, rewardCoins.IntPart()}} //TODO make Decimal
		err = creditAcc(validator.Account, credit)
		if err != nil {
			return err
		}

		// Calculate the total amount of new tokens to be
		// assigned to the validator for the commission
		//
		// NOTE this can be a bit confusing best to work
		// on paper yourself, but the general proof to
		// arrive at the commTok2Val eqn is:
		//
		//   rate*(totalOldTok + newTok) = totalNewCoin
		//   rate*(totalOldTok) = totalNewCoin - commissionCoins
		//   :.
		//   newTok = ((totalNewCoin * TotalOldTok)
		//             /(totalNewCoin - commissionCoins))
		//             - totalOldTok

		//start by loading the bond account of the validator to itself
		delegators, err := loadDelegatorBonds(store, validator.Delegatee)
		if err != nil {
			return err
		}
		j, valSelfBond := delegators.Get(validator.Delegatee)

		coins1 := validator.TotalBondTokens                                     // total bonded coins before rewards
		coins2 := coins1.Add(rewardCoins)                                       // total bonded coins after rewards
		tok1 := validator.TotalBondTokens                                       // total tokens before rewards
		tok1Val := valSelfBond.BondTokens                                       // total tokens before rewards owned by the validator
		preRewardsDel := rewardCoins.Mul((tok1.Sub(tok1Val)).Div(tok1))         // pre-commission reward coins for delegators
		commCoin := preRewardsDel.Mul(validator.Commission)                     // commission coins taken on the preRewardsDel
		commTok2Val := ((coins2.Mul(tok1)).Div(coins2.Mul(commCoin))).Sub(tok1) // new tokens to be added to the validator bond account for commission

		//Add the new tokens to the validators self bond delegator account
		delegators[j].BondTokens = delegators[j].BondTokens.Add(commTok2Val)

		//save the updated delegator bond account for the validator
		saveDelegatorBonds(store, validator.Delegatee, delegators)

		//////////////////////////////////
		// Update the Validator's ExchangeRate and TotalBondTokens.
		//   The total number of tokens only increases slightly, and similarily
		//   the exchange rate should only change slightly due to the new tokens
		//   introduced from the commission rate to the validators (self-)delegator account

		// The new exchange rate can be calculated as following:
		// StartCoins = StartExchangeRate * StartTotalBondTokens
		// FinalExchangeRate = FinalCoins/FinalTotalBondTokens
		// EndExchangeRate = FinalCoins/FinalTotalBondTokens
		// EndExchangeRate = (StartCoins + RewardCoins)/FinalTotalBondTokens
		finalTotalBondTokens := delegateeBonds[i].TotalBondTokens.Add(commTok2Val)
		startCoins := delegateeBonds[i].ExchangeRate.Mul(delegateeBonds[i].TotalBondTokens)
		delegateeBonds[i].ExchangeRate = (startCoins.Add(rewardCoins)).Div(finalTotalBondTokens)

		//Also update the TotalBondTokens
		delegateeBonds[i].TotalBondTokens = finalTotalBondTokens
	}

	//save the updated delegateeBonds
	saveDelegateeBonds(store, delegateeBonds)

	//save the inflated total atom supply
	saveAtomSupply(store, totalAtoms)

	return nil
}
