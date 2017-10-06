package stake

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestTypes(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	store := state.NewMemKVStore()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")
	actor1 := sdk.Actor{"testChain", "testapp", addr1}
	actor2 := sdk.Actor{"testChain", "testapp", addr2}
	actor3 := sdk.Actor{"testChain", "testapp", addr3}
	holdActor1 := sdk.Actor{"testChain", "testapp", []byte("hold1")}
	holdActor2 := sdk.Actor{"testChain", "testapp", []byte("hol2")}
	holdActor3 := sdk.Actor{"testChain", "testapp", []byte("hol3")}

	//NOTE PubKey is supposed to be the binaryBytes of the crypto.PubKey
	// instead this is just being set the address here for testing purposes
	validator1 := &ValidatorBond{
		Validator:    actor1,
		PubKey:       actor1.Address.Bytes(),
		BondedTokens: 10,
		HoldAccount:  holdActor1,
		VotingPower:  10,
	}
	validator2 := &ValidatorBond{
		Validator:    actor2,
		PubKey:       actor2.Address.Bytes(),
		BondedTokens: 300,
		HoldAccount:  holdActor2,
		VotingPower:  300,
	}
	validator3 := &ValidatorBond{
		Validator:    actor3,
		PubKey:       actor3.Address.Bytes(),
		BondedTokens: 123,
		HoldAccount:  holdActor3,
		VotingPower:  123,
	}

	validators := ValidatorBonds{validator1, validator2, validator3}

	//test basic sort
	validators.Sort()
	//assert.True(validators[0].Validator.Equals(actor1), "not equal: %v, %v" validators[0].ValidatorV//)
	assert.Equal(validators[0].Validator, actor2)
	assert.Equal(validators[1].Validator, actor3)
	assert.Equal(validators[2].Validator, actor1)

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
