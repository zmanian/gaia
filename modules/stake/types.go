package stake

import (
	"bytes"
	"sort"

	"github.com/cosmos/cosmos-sdk"
	abci "github.com/tendermint/abci/types"
)

const maxValidators = 100

// DelegateeBond defines an total amount of bond tokens and their exchange rate to
// coins, associated with a single validator. Interest increases the exchange
// rate, and slashing decreases it. When coins are delegated to this validator,
// the delegatee is credited with bond tokens based on amount of coins
// delegated and the current exchange rate. Voting power can be calculated as
// total bonds multiplied by exchange rate.
type DelegateeBond struct {
	Delegatee       sdk.Actor
	Commission      Decimal
	ExchangeRate    Decimal   // Exchange rate for this validator's bond tokens (in millionths of coins)
	TotalBondTokens Decimal   // Total number of bond tokens in the account
	Account         sdk.Actor // Account where the bonded tokens are held
}

// VotingPower - voting power based onthe bond value
func (b DelegateeBond) VotingPower() uint64 {
	decPower := b.TotalBondTokens.Mul(b.ExchangeRate)

	//in order to pass the voting power as an uint64 with some precision multiple by a large number
	return uint64(decPower.Mul(NewDecimal(1, 10)).IntPart())
}

// Validator - Get the validator from a bond value
func (b DelegateeBond) Validator() *abci.Validator {
	return &abci.Validator{
		PubKey: b.Delegatee.Address,
		Power:  b.VotingPower(),
	}
}

//--------------------------------------------------------------------------------

// DelegateeBonds - the set of all DelegateeBonds
type DelegateeBonds []DelegateeBond

var _ sort.Interface = DelegateeBonds{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (b DelegateeBonds) Len() int      { return len(b) }
func (b DelegateeBonds) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b DelegateeBonds) Less(i, j int) bool {
	vp1, vp2 := b[i].VotingPower(), b[j].VotingPower()
	d1, d2 := b[i].Delegatee, b[j].Delegatee
	switch {
	case vp1 != vp2:
		return vp1 > vp2
	case d1.ChainID < d2.ChainID:
		return true
	case d1.ChainID > d2.ChainID:
		return false
	case d1.App < d2.App:
		return true
	case d1.App > d2.App:
		return false
	default:
		return bytes.Compare(d1.Address, d2.Address) == -1
	}
}

// Sort - Sort the array of bonded values
func (b DelegateeBonds) Sort() {
	sort.Sort(b)
}

// Validators - get the active validator list from the array of DelegateeBonds
func (b DelegateeBonds) Validators() []*abci.Validator {
	validators := make([]*abci.Validator, 0, maxValidators)
	for i, bv := range b {
		if i == maxValidators {
			break
		}
		validators = append(validators, bv.Validator())
	}
	return validators
}

// ValidatorsDiff - get the difference in the validator set from the input validator set
func (b DelegateeBonds) ValidatorsDiff(previous []*abci.Validator) []*abci.Validator {

	//Get the current validator set
	current := b.Validators()

	//calculate any differences from the previous to the current validator set
	diff := make([]*abci.Validator, 0, maxValidators)
	for _, prev := range previous {

		//test for a difference between the validator power of validator
		currentPower := getValidatorPower(current, prev.PubKey)
		if currentPower != prev.Power {
			diff = append(diff, &abci.Validator{prev.PubKey, currentPower})
		}
	}
	return diff
}

// getValidatorPower - return the validator power with the matching pubKey from the validator list
func getValidatorPower(set []*abci.Validator, pubKey []byte) uint64 {
	for _, validator := range set {
		if bytes.Equal(validator.PubKey, pubKey) {
			return validator.Power
		}
	}
	return 0 // no power if not found
}

// ValidatorsActors - get the actors of the active validator list from the array of DelegateeBonds
func (b DelegateeBonds) ValidatorsActors() []sdk.Actor {
	accounts := make([]sdk.Actor, 0, maxValidators)
	for i, bv := range b {
		if i == maxValidators {
			break
		}
		accounts = append(accounts, bv.Account)
	}

	return accounts
}

// Get - get a DelegateeBond for a specific validator from the DelegateeBonds
func (b DelegateeBonds) Get(delegatee sdk.Actor) (int, *DelegateeBond) {
	for i, bv := range b {
		if bytes.Equal(bv.Delegatee.Address, delegatee.Address) &&
			bv.Delegatee.ChainID == delegatee.ChainID &&
			bv.Delegatee.App == delegatee.App {
			return i, &bv
		}
	}
	return 0, nil
}

// Remove - remove delegatee from the delegatee list
func (b DelegateeBonds) Remove(i int) DelegateeBonds {
	return append(b[:i], b[i+1:]...)
}

////////////////////////////////////////////////////////////////////////////////

// DelegatorBond defines an account of bond tokens. It is owned by one delegator
// account, and is associated with one delegatee account
type DelegatorBond struct {
	Delegatee  sdk.Actor
	BondTokens Decimal // amount of bond tokens
}

// DelegatorBonds - all delegator bonds existing with multiple delegatees
type DelegatorBonds []DelegatorBond

// Get - get a DelegateeBond for a specific validator from the DelegateeBonds
func (b DelegatorBonds) Get(delegatee sdk.Actor) (int, *DelegatorBond) {
	for i, bv := range b {
		if bytes.Equal(bv.Delegatee.Address, delegatee.Address) &&
			bv.Delegatee.ChainID == delegatee.ChainID &&
			bv.Delegatee.App == delegatee.App {
			return i, &bv
		}
	}
	return 0, nil
}

// Remove - remove delegatee from the delegatee list
func (b DelegatorBonds) Remove(i int) DelegatorBonds {
	return append(b[:i], b[i+1:]...)
}

////////////////////////////////////////////////////////////////////////////////

// QueueElem - queue element, the basis of a queue interaction with a delegatee/validator
type QueueElem struct {
	Delegatee    sdk.Actor
	HeightAtInit uint64 // when the queue was initiated
}

// QueueElemUnbond - the unbonding queue element
type QueueElemUnbond struct {
	QueueElem
	Account    sdk.Actor // account to pay out to
	BondTokens Decimal   // amount of bond tokens which are unbonding
}

// QueueElemModComm - the commission queue element
type QueueElemModComm struct {
	QueueElem
	Commission Decimal // new commission for the
}
