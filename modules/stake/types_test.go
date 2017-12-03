package stake

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

var pks = []crypto.PubKey{newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
}

// NOTE: PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func candidatesFromActors(actors []sdk.Actor, store state.SimpleDB, amts []int) (candidates Candidates) {
	for i, a := range actors {
		c := &Candidate{
			PubKey:      pks[i],
			Owner:       a,
			Shares:      uint64(amts[i]),
			VotingPower: uint64(amts[i]),
		}
		candidates = append(candidates, c)
		saveCandidate(store, c)
	}

	return
}

func TestCandidatesMaxVals(t *testing.T) {
	params := defaultParams()
	assert := assert.New(t)
	store := state.NewMemKVStore()
	actors := newActors(3)
	bonds := candidatesFromActors(actors, store, []int{10, 300, 123})

	testCases := []struct {
		maxVals, expectedVals int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 3},
	}

	for _, testCase := range testCases {
		params.MaxVals = testCase.maxVals
		saveParams(store, params)
		UpdateValidatorSet(store)
		assert.Equal(testCase.expectedVals, len(bonds.GetValidators(store)), "%v", bonds.GetValidators(store))
	}
}

func TestCandidatesSort(t *testing.T) {
	params := defaultParams()
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()

	N := 5
	actors := newActors(N)
	bonds := candidatesFromActors(actors, store, []int{10, 300, 123, 4, 200})
	expectedOrder := []int{1, 4, 2, 0, 3}

	// test basic sort
	bonds.Sort()

	vals := bonds.GetValidators(store)
	require.Equal(N, len(vals))

	for i, val := range vals {
		expectedIdx := expectedOrder[i]
		assert.Equal(val.PubKey, pks[expectedIdx])
	}

	// now reduce the maxvals, ensure they're still ordered
	maxVals := 3
	params.MaxVals = maxVals
	saveParams(store, params)
	UpdateValidatorSet(store)
	vals = bonds.GetValidators(store)
	require.Equal(maxVals, len(vals))

	for i, val := range vals {
		expectedIdx := expectedOrder[i]
		assert.Equal(val.PubKey, pks[expectedIdx])
	}
}
func TestCandidatesUpdate(t *testing.T) {
	params := defaultParams()
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()

	actors := newActors(3)
	candidates := candidatesFromActors(actors, []int{10, 300, 123})
	candidates.Sort()

	maxVals := 2
	params.MaxVals = maxVals
	saveParams(store, params)

	// Change some of the bonded shares, get the new validator set
	vals1 := candidates.GetValidators(store)
	candidates[2].Shares = 1000
	candidates.updateVotingPower(store)
	vals2 := candidates.GetValidators(store)

	require.Equal(maxVals, len(vals2))

	expectedOrder := []int{0, 1, 2}
	for i, val := range vals2 {
		expectedIdx := expectedOrder[i]
		assert.Equal(val.PubKey, pks[expectedIdx])
	}

	// calculate the difference in the validator set from the original set
	diff := validatorsDiff(vals1, vals2, store)

	require.Equal(2, len(diff), "validator diff should have length 2, diff %v, val1 %v, val2 %v",
		diff, vals1, vals2)
	assert.True(diff[0].Power == 0)
	assert.True(diff[1].Power == 1000)
}
