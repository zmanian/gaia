package stake

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
)

// ValidatorBond defines the total amount of bond tokens and their exchange rate to
// coins, associated with a single validator. Accumulation of interest is modelled
// as an in increase in the exchange rate, and slashing as a decrease.
// When coins are delegated to this validator, the validator is credited
// with a DelegatorBond whose number of bond tokens is based on the amount of coins
// delegated divided by the current exchange rate. Voting power can be calculated as
// total bonds multiplied by exchange rate.
type ValidatorBond struct {
	Validator    sdk.Actor
	BondedTokens uint64    // Total number of bond tokens for the validator
	HoldAccount  sdk.Actor // Account where the bonded coins are held. Controlled by the app
	VotingPower  uint64    // Total number of bond tokens for the validator
}

// ABCIValidator - Get the validator from a bond value
func (b ValidatorBond) ABCIValidator() *abci.Validator {

	//var PubKey crypto.PubKeyEd25519 = make([]byte, 32)
	//b.Validator.Address
	return &abci.Validator{
		PubKey: b.Validator.Address, //XXX needs to actually be wire.BinaryBytes(ValidatorPubKey)
		Power:  b.VotingPower,       //TODO could be a problem sending in truncated IntPart here
	}
}

//--------------------------------------------------------------------------------

// ValidatorBonds - the set of all ValidatorBonds
type ValidatorBonds []*ValidatorBond

var _ sort.Interface = ValidatorBonds{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (b ValidatorBonds) Len() int      { return len(b) }
func (b ValidatorBonds) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ValidatorBonds) Less(i, j int) bool {
	vp1, vp2 := b[i].VotingPower, b[j].VotingPower
	d1, d2 := b[i].Validator, b[j].Validator
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
func (b ValidatorBonds) Sort() {
	sort.Sort(b)
}

// UpdateVotingPower - voting power based on bond tokens and exchange rate
// TODO make not a function of ValidatorBonds as validatorbonds can be loaded from the store
func (b ValidatorBonds) UpdateVotingPower(store state.SimpleDB) {

	for _, bv := range b {
		bv.VotingPower = bv.BondedTokens
	}

	// Now sort and truncate the power
	b.Sort()
	for i, bv := range b {
		if i >= maxVal {
			bv.VotingPower = 0
		}
	}
	saveValidatorBonds(store, b)
	return
}

// GetValidators - get the most recent updated validator set from the
// ValidatorBonds. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func (b ValidatorBonds) GetValidators() []*abci.Validator {
	validators := make([]*abci.Validator, 0, maxVal)
	for _, bv := range b {
		if bv.VotingPower == 0 { //exit as soon as the first Voting power set to zero is found
			break
		}
		validators = append(validators, bv.ABCIValidator())
	}
	return validators
}

// ValidatorsDiff - get the difference in the validator set from the input validator set
func ValidatorsDiff(previous, new []*abci.Validator) (diff []*abci.Validator) {

	//calculate any differences from the previous to the new validator set
	// first loop through the previous validator set, and then catch any
	// missed records in the new validator set
	diff = make([]*abci.Validator, 0, maxVal)
	for _, prevVal := range previous {
		found := false
		for _, newVal := range new {
			if bytes.Equal(prevVal.PubKey, newVal.PubKey) {
				found = true
				if newVal.Power != prevVal.Power {
					diff = append(diff, &abci.Validator{newVal.PubKey, newVal.Power})
					break
				}
			}
		}
		if !found {
			diff = append(diff, &abci.Validator{prevVal.PubKey, 0})
		}
	}
	for _, newVal := range new {
		found := false
		for _, prevVal := range previous {
			if bytes.Equal(prevVal.PubKey, newVal.PubKey) {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &abci.Validator{newVal.PubKey, newVal.Power})
		}
	}
	return
}

// Get - get a ValidatorBond for a specific validator from the ValidatorBonds
func (b ValidatorBonds) Get(validator sdk.Actor) (int, *ValidatorBond) {
	for i, bv := range b {
		if bv.Validator.Equals(validator) {
			return i, bv
		}
	}
	return 0, nil
}

// Remove - remove validator from the validator list
func (b ValidatorBonds) Remove(i int) (ValidatorBonds, error) {
	switch {
	case i < 0:
		return b, fmt.Errorf("Cannot remove a negative element")
	case i >= len(b):
		return b, fmt.Errorf("Element is out of upper bound")
	default:
		return append(b[:i], b[i+1:]...), nil
	}
}
