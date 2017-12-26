package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/rational"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := state.NewMemKVStore()

	delegator := sdk.Actor{"testChain", "testapp", []byte("addressdelegator")}
	validator := sdk.Actor{"testChain", "testapp", []byte("addressvalidator")}

	pk := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")

	//----------------------------------------------------------------------
	// Candidate checks

	// XXX expand to include both liabilities and assets use/test all candidate fields
	candidate := &Candidate{
		Owner:       validator,
		PubKey:      pk,
		Liabilities: rational.New(9),
		VotingPower: rational.New(0),
	}

	// check the empty store first
	resCand := loadCandidate(store, pk)
	assert.Nil(resCand)
	resPks := loadCandidatesPubKeys(store)
	assert.Zero(len(resPks))

	// set and retrieve a record
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.Equal(candidate, resCand)

	// modify a records, save, and retrieve
	candidate.Liabilities = rational.New(99)
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.Equal(candidate, resCand)

	// also test that the pubkey has been added to pubkey list
	resPks = loadCandidatesPubKeys(store)
	require.Equal(1, len(resPks))
	assert.Equal(pk, resPks[0])

	//----------------------------------------------------------------------
	// Bond checks

	bond := &DelegatorBond{
		PubKey: pk,
		Shares: rational.New(9),
	}

	//check the empty store first
	resBond := loadDelegatorBond(store, delegator, pk)
	assert.Nil(resBond)

	//Set and retrieve a record
	saveDelegatorBond(store, delegator, bond)
	resBond = loadDelegatorBond(store, delegator, pk)
	assert.Equal(bond, resBond)

	//modify a records, save, and retrieve
	bond.Shares = rational.New(99)
	saveDelegatorBond(store, delegator, bond)
	resBond = loadDelegatorBond(store, delegator, pk)
	assert.Equal(bond, resBond)

	//----------------------------------------------------------------------
	// Param checks

	params := defaultParams()

	//check that the empty store loads the default
	resParams := loadParams(store)
	assert.Equal(params, resParams)

	//modify a params, save, and retrieve
	params.MaxVals = 777
	saveParams(store, params)
	resParams = loadParams(store)
	assert.Equal(params, resParams)
}
