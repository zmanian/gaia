package stake

import (
	"bytes"
	"sort"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/modules/coin"
)

//TODO make genesis parameter
const maxValidators = 100

// BondValue defines an total amount of bond tokens and their exchange rate to
// coins, associated with a single validator. Interest increases the exchange
// rate, and slashing decreases it. When coins are delegated to this validator,
// the delegator is credited with bond tokens based on amount of coins
// delegated and the current exchange rate. Voting power can be calculated as
// total bonds multiplied by exchange rate.
type BondValue struct {
	ValidatorPubKey []byte
	Commission      string
	Total           uint64 // Number of bond tokens for this validator
	ExchangeRate    uint64 // Exchange rate for this validator's bond tokens (in millionths of coins)
}

// VotingPower - voting power based onthe bond value
func (bc BondValue) VotingPower() uint64 {
	return bc.Total * bc.ExchangeRate / Precision
}

// Validator - Get the validator from a bond value
func (bc BondValue) Validator() *abci.Validator {
	return &abci.Validator{
		PubKey: bc.ValidatorPubKey,
		Power:  bc.VotingPower(),
	}
}

//--------------------------------------------------------------------------------

// BondValues - the set of all BondValues
type BondValues []BondValue

var _ sort.Interface = BondValues{}

// nolint - sort interface functions
func (bvs BondValues) Len() int      { return len(bvs) }
func (bvs BondValues) Swap(i, j int) { bvs[i], bvs[j] = bvs[j], bvs[i] }
func (bvs BondValues) Less(i, j int) bool {
	vp1, vp2 := bvs[i].VotingPower(), bvs[j].VotingPower()
	if vp1 == vp2 {
		return bytes.Compare(bvs[i].ValidatorPubKey, bvs[j].ValidatorPubKey) == -1
	}
	return vp1 > vp2
}

// Sort - Sort the array of bonded values
func (bvs BondValues) Sort() {
	sort.Sort(bvs)
}

// Validators - get the active validator list from the array of BondValues
func (bvs BondValues) Validators() []*abci.Validator {
	validators := make([]*abci.Validator, 0, maxValidators)
	for i, bv := range bvs {
		if i == maxValidators {
			break
		}
		validators = append(validators, bv.Validator())
	}
	return validators
}

// Get - get a BondValue for a specific validator from the BondValues
func (bvs BondValues) Get(validatorPubKey []byte) (int, *BondValue) {
	for i, bv := range bvs {
		if bytes.Equal(bv.ValidatorPubKey, validatorPubKey) {
			return i, &bv
		}
	}
	return 0, nil
}

//TODO remove this block if not used
//func (bvs BondValues) Remove(i int) BondValues {
//return append(bvs[:i], bvs[i+1:]...)
//}

//--------------------------------------------------------------------------------

// BondAccount defines an account of bond tokens. It is owned by one basecoin
// account, and is associated with one validator.
type BondAccount struct {
	Amount coin.Coins // amount of bond tokens
}

//TODO remove this block if not used
// Unbond defines an amount of bond tokens which are in the unbonding period
//type Unbond struct {
//ValidatorPubKey []byte
//Address         []byte // account to pay out to
//BondAmount      uint64 // amount of bond tokens which are unbonding
//Height          uint64 // when the unbonding started
//}
