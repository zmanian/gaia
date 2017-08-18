package stake

import (
	"bytes"
	"sort"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
)

const maxValidators = 100

// DelegateeBond defines an total amount of bond tokens and their exchange rate to
// coins, associated with a single validator. Interest increases the exchange
// rate, and slashing decreases it. When coins are delegated to this validator,
// the delegatee is credited with bond tokens based on amount of coins
// delegated and the current exchange rate. Voting power can be calculated as
// total bonds multiplied by exchange rate.
type DelegateeBond struct {
	DelegateeAddr []byte
	Commission    uint64
	ExchangeRate  uint64         // Exchange rate for this validator's bond tokens (in millionths of coins)
	Account       basecoin.Actor // Account where the bonded tokens are held
}

// VotingPower - voting power based onthe bond value
func (bc DelegateeBond) VotingPower() uint64 {
	//TODO must query from the account
	return bc.Total * bc.ExchangeRate
	//return bc.Total * bc.ExchangeRate / Precision
}

// Validator - Get the validator from a bond value
func (bc DelegateeBond) Validator() *abci.Validator {
	return &abci.Validator{
		PubKey: bc.DelegateeAddr,
		Power:  bc.VotingPower(),
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
	if vp1 == vp2 {
		return bytes.Compare(b[i].DelegateeAddr, b[j].ValidatorPubKey) == -1
	}
	return vp1 > vp2
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

// Get - get a DelegateeBond for a specific validator from the DelegateeBonds
func (b DelegateeBonds) Get(delegateeAddr []byte) (int, *DelegateeBond) {
	for i, bv := range b {
		if bytes.Equal(bv.DelegateeAddr, delegateeAddr) {
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
	DelegateeAddr []byte
	BondTokens    uint64 // amount of bond tokens
}

// DelegatorBonds - all delegator bonds
type DelegatorBonds []DelegatorBond

// Get - get a DelegateeBond for a specific validator from the DelegateeBonds
func (b DelegatorBonds) Get(delegateeAddr []byte) (int, *DelegatorBond) {
	for i, bv := range b {
		if bytes.Equal(bv.DelegateeAddr, delegateeAddr) {
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
	DelegateeAddr []byte
	HeightAtInit  uint64 // when the queue was initiated
}

// QueueElemUnbond - the unbonding queue element
type QueueElemUnbond struct {
	QueueElem
	Account    basecoin.Actor // account to pay out to
	BondTokens uint64         // amount of bond tokens which are unbonding
}

// QueueElemModComm - the commission queue element
type QueueElemModComm struct {
	QueueElem
	Commission string // new commission for the
}
