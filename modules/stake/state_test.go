package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestState(t *testing.T) {
	assert := assert.New(t)

	store := state.NewMemKVStore()

	validator1 := sdk.Actor{"testChain", "testapp", []byte("addressvalidator1")}

	candidates := Candidates{
		&Candidate{
			Owner:  validator1,
			PubKey: pk1,
			Shares: 9,
		}}
	var validatorNilBonds Candidates

	/////////////////////////////////////////////////////////////////////////
	// Candidates checks

	//check the empty store first
	resGet := LoadCandidates(store)
	assert.Equal(validatorNilBonds, resGet)

	//Set and retrieve a record
	saveCandidate(store, candidates[0])
	resGet = LoadCandidates(store)
	assert.Equal(candidates, resGet)

	//modify a records, save, and retrieve
	candidates[0].Shares = 99
	saveCandidate(store, candidates[0])
	resGet = LoadCandidates(store)
	assert.Equal(candidates, resGet)
}
