package stake

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func newActors(n int) (actors []sdk.Actor) {
	for i := 0; i < n; i++ {
		actors = append(actors, sdk.Actor{
			"testChain", "testapp", []byte(fmt.Sprintf("addr%d", i))})
	}

	return
}

var pks = []crypto.PubKey{
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
}

// NOTE: PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func candidatesFromActors(actors []sdk.Actor, amts []int) (candidates Candidates) {
	for i := 0; i < len(actors); i++ {
		c := &Candidate{
			PubKey:      pks[i],
			Owner:       actors[i],
			Shares:      uint64(amts[i]),
			VotingPower: uint64(amts[i]),
		}
		candidates = append(candidates, c)
	}

	return
}

// helper function test if Candidate is changed asabci.Validator
func testChange(t *testing.T, val Validator, chg *abci.Validator) {
	assert := assert.New(t)
	assert.Equal(val.PubKey.Bytes(), chg.PubKey)
	assert.Equal(val.VotingPower, chg.Power)
}

// helper function test if Candidate is removed as abci.Validator
func testRemove(t *testing.T, val Validator, chg *abci.Validator) {
	assert := assert.New(t)
	assert.Equal(val.PubKey.Bytes(), chg.PubKey)
	assert.Equal(uint64(0), chg.Power)
}

//___________________________________________________________________________________

func TestCandidatesSort(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int{10, 300, 123, 4, 200})
	expectedOrder := []int{1, 4, 2, 0, 3}

	// test basic sort
	candidates.Sort()

	vals := candidates.Validators()
	require.Equal(N, len(vals))

	for i, val := range vals {
		expectedIdx := expectedOrder[i]
		assert.Equal(val.PubKey, pks[expectedIdx])
	}
}

func TestUpdateVotingPower(t *testing.T) {
	assert := assert.New(t)
	store := state.NewMemKVStore()

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int{400, 200, 100, 10, 1})

	// test a basic change in voting power
	candidates[0].Shares = 500
	candidates.updateVotingPower(store)
	assert.Equal(uint64(500), candidates[0].VotingPower, "%v", candidates[0])

	// test a swap in voting power
	candidates[1].Shares = 600
	candidates.updateVotingPower(store)
	assert.Equal(uint64(600), candidates[0].VotingPower, "%v", candidates[0])
	assert.Equal(uint64(500), candidates[1].VotingPower, "%v", candidates[1])

	// test the max validators term
	params := loadParams(store)
	params.MaxVals = 4
	saveParams(store, params)
	candidates.updateVotingPower(store)
	assert.Equal(uint64(0), candidates[4].VotingPower, "%v", candidates[4])
}

func TestGetValidators(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int{400, 200, 0, 0, 0})

	validators := candidates.Validators()
	require.Equal(2, len(validators))
	assert.Equal(candidates[0].PubKey, validators[0].PubKey)
	assert.Equal(candidates[1].PubKey, validators[1].PubKey)
}

func TestValidatorsChanged(t *testing.T) {
	require := require.New(t)

	v1 := (&Candidate{PubKey: pks[0], VotingPower: 10}).validator()
	v2 := (&Candidate{PubKey: pks[1], VotingPower: 10}).validator()
	v3 := (&Candidate{PubKey: pks[2], VotingPower: 10}).validator()
	v4 := (&Candidate{PubKey: pks[3], VotingPower: 10}).validator()
	v5 := (&Candidate{PubKey: pks[4], VotingPower: 10}).validator()

	// test from nothing to something
	vs1 := Validators{}
	vs2 := Validators{v1, v2}
	changed := vs1.validatorsChanged(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[1], changed[1])

	// test from something to nothing
	vs1 = Validators{v1, v2}
	vs2 = Validators{}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[0], changed[0])
	testRemove(t, vs1[1], changed[1])

	// test identical
	vs1 = Validators{v1, v2, v4}
	vs2 = Validators{v1, v2, v4}
	changed = vs1.validatorsChanged(vs2)
	require.Zero(len(changed))

	// test single value change
	vs2[2].VotingPower = 1
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test multiple value change
	vs2[0].VotingPower = 11
	vs2[2].VotingPower = 5
	changed = vs1.validatorsChanged(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[0], changed[0])
	testChange(t, vs2[2], changed[1])

	// test validator added at the beginning
	vs1 = Validators{v2, v4}
	vs2 = Validators{v1, v2, v4}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[0], changed[0])

	// test validator added in the middle
	vs1 = Validators{v1, v2, v4}
	vs2 = Validators{v1, v2, v3, v4}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[2], changed[0])

	// test validator added at the end
	vs2 = Validators{v1, v2, v4, v5}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testChange(t, vs2[3], changed[0])

	// test multiple validators added
	vs2 = Validators{v1, v2, v3, v4, v5}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(2, len(changed))
	testChange(t, vs2[2], changed[0])
	testChange(t, vs2[4], changed[1])

	// test validator removed at the beginning
	vs2 = Validators{v2, v4}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[0], changed[0])

	// test validator removed in the middle
	vs2 = Validators{v1, v4}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[1], changed[0])

	// test validator removed at the end
	vs2 = Validators{v1, v2}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(1, len(changed))
	testRemove(t, vs1[2], changed[0])

	// test multiple validators removed
	vs2 = Validators{v1}
	changed = vs1.validatorsChanged(vs2)
	require.Equal(2, len(changed))
	testRemove(t, vs1[1], changed[0])
	testRemove(t, vs1[2], changed[1])

	// test many types of changes
	vs2 = Validators{v1, v3, v4, v5}
	vs2[2].VotingPower = 11
	changed = vs1.validatorsChanged(vs2)
	require.Equal(4, len(changed), "%v", changed) // change 1, remove 1, add 2
	testRemove(t, vs1[1], changed[0])
	testChange(t, vs2[1], changed[1])
	testChange(t, vs2[2], changed[2])
	testChange(t, vs2[3], changed[3])

}

func TestUpdateValidatorSet(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int{400, 200, 100, 10, 1})
	for _, c := range candidates {
		saveCandidate(store, c)
	}

	// They should all already be validators
	change, err := UpdateValidatorSet(store)
	require.Nil(err)
	require.Equal(0, len(change), "%v", change) // change 1, remove 1, add 2

	// test the max value and test again
	params := loadParams(store)
	params.MaxVals = 4
	saveParams(store, params)
	change, err = UpdateValidatorSet(store)
	require.Nil(err)
	require.Equal(1, len(change), "%v", change)
	testRemove(t, candidates[4].validator(), change[0])
	candidates = loadCandidates(store)
	assert.Equal(uint64(0), candidates[4].VotingPower)

	//mess with the power's of the candidates and test
	candidates[0].Shares = 10
	candidates[1].Shares = 200
	candidates[2].Shares = 1000
	candidates[3].Shares = 1
	candidates[4].Shares = 10
	for _, c := range candidates {
		saveCandidate(store, c)
	}
	change, err = UpdateValidatorSet(store)
	require.Nil(err)
	require.Equal(5, len(change), "%v", change) //3 changed, 1 added, 1 removed
	//testRemove(t, candidates[3].validator(), change[0])
	//candidates = loadCandidates(store)
	//assert.Equal(uint64(0), candidates[4].VotingPower)

}
