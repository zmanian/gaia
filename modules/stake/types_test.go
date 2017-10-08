package stake

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func newActors(n int) (actors []sdk.Actor) {
	for i := 0; i < n; i++ {
		actors = append(actors, sdk.Actor{
			"testChain", "testapp", fmt.Sprintf("addr%d", i)})
	}
	return
}

//NOTE PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func bondsFromActors(actors []sdk.Actor, amts []int) (bonds []*ValidatorBond) {
	for i, a := range actors {
		bonds = append(bonds, &ValidatorBond{
			Validator:    a,
			PubKey:       a.Address.Bytes(),
			BondedTokens: amts[i],
			HoldAccount:  getHoldAccount(a),
			VotingPower:  amts[i],
		})
	}
	return

}

func TestValidatorBonds(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()

	actors := newActors(3)
	vals := ValidatorBonds(valsFromActors(actors, []int{10, 300, 123}))

	//test basic sort
	vals.Sort()
	//assert.True(validators[0].Validator.Equals(actor1), "not equal: %v, %v" validators[0].ValidatorV//)
	assert.Equal(vals[0].Validator, actors[1])
	assert.Equal(vals[1].Validator, actors[2])
	assert.Equal(vals[2].Validator, actors[0])

	//get the base validators set which will contain all the validators
	validators0 := validators.GetValidators()
	require.Equal(3, len(validators0))

	//test to see if the maxVal is functioning
	maxVal = 0
	validators.UpdateVotingPower(store)
	assert.Equal(0, len(validators.GetValidators()))
	maxVal = 1
	validators.UpdateVotingPower(store)
	assert.Equal(1, len(validators.GetValidators()))
	maxVal = 2
	validators.UpdateVotingPower(store)
	assert.Equal(2, len(validators.GetValidators()))

	//get/test the existing validator set
	validators1 := validators.GetValidators()
	require.Equal(2, len(validators1))
	assert.True(bytes.Equal(validators1[0].PubKey, actor2.Address.Bytes()),
		"%v, %v", validators1[0].PubKey, actor2.Address.Bytes())
	assert.True(bytes.Equal(validators1[1].PubKey, actor3.Address),
		"%v, %v", validators1[1].PubKey, actor3.Address)

	// Change some of the bonded tokens, get the new validator set
	validator1.BondedTokens = 1000
	validators.UpdateVotingPower(store)
	validators2 := validators.GetValidators()
	require.Equal(2, len(validators2))
	assert.True(bytes.Equal(validators2[0].PubKey, actor1.Address.Bytes()),
		"%v, %v", validators2[0].PubKey, actor1.Address)
	assert.True(bytes.Equal(validators2[1].PubKey, actor2.Address),
		"%v, %v", validators2[1].PubKey, actor2.Address)

	// calculate the difference in the validator set from the original set
	diff := ValidatorsDiff(validators1, validators2)
	require.Equal(2, len(diff), "validator diff should have length 2, diff %v, val1 %v, val2 %v",
		diff, validators1, validators2)
	assert.True(diff[0].Power == 0)
	assert.True(diff[1].Power == validators2[0].Power)
}
