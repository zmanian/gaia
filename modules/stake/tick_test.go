package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/rational"

	"github.com/cosmos/cosmos-sdk/state"
)

// XXX complete test
//func TestGetInflation(t *testing.T) {
//assert, require := assert.New(t), require.New(t)
//}

func TestProcessProvisions(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()
	params := loadParams(store)
	gs := loadGlobalState(store)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int64{400, 200, 100, 10, 1})
	for _, c := range candidates {
		saveCandidate(store, c)
	}

	// they should all already be validators
	change, err := UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(0, len(change), "%v", change) // change 1, remove 1, add 2

	// test the max value and test again
	params.MaxVals = 4
	saveParams(store, params)
	change, err = UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(1, len(change), "%v", change)
	testRemove(t, candidates[4].validator(), change[0])
	candidates = loadCandidates(store)
	assert.Equal(int64(0), candidates[4].VotingPower.Evaluate())

	// mess with the power's of the candidates and test
	candidates[0].Assets = rational.New(10)
	candidates[1].Assets = rational.New(600)
	candidates[2].Assets = rational.New(1000)
	candidates[3].Assets = rational.New(1)
	candidates[4].Assets = rational.New(10)
	for _, c := range candidates {
		saveCandidate(store, c)
	}
	change, err = UpdateValidatorSet(store, gs, params)
	require.Nil(err)
	require.Equal(5, len(change), "%v", change) //3 changed, 1 added, 1 removed
	candidates = loadCandidates(store)
	testChange(t, candidates[0].validator(), change[0])
	testChange(t, candidates[1].validator(), change[1])
	testChange(t, candidates[2].validator(), change[2])
	testRemove(t, candidates[3].validator(), change[3])
	testChange(t, candidates[4].validator(), change[4])
}
