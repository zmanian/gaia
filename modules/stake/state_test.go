package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := state.NewMemKVStore()

	validator1 := sdk.Actor{"testChain", "testapp", []byte("addressvalidator1")}

	validatorBonds := ValidatorBonds{
		&ValidatorBond{
			Validator:    validator1,
			PubKey:       []byte{},
			BondedTokens: 9,
			HoldAccount:  sdk.Actor{"testChain", "testapp", []byte("addresslockedtoapp")},
		}}
	var validatorNilBonds ValidatorBonds

	/////////////////////////////////////////////////////////////////////////
	// ValidatorBonds checks

	//check the empty store first
	resGet, err := LoadValidatorBonds(store)
	require.Nil(err)
	assert.Equal(validatorNilBonds, resGet)

	//Set and retrieve a record
	saveValidatorBonds(store, validatorBonds)
	resGet, err = LoadValidatorBonds(store)
	require.Nil(err)
	assert.Equal(validatorBonds, resGet)

	//modify a records, save, and retrieve
	validatorBonds[0].BondedTokens = 99
	saveValidatorBonds(store, validatorBonds)
	resGet, err = LoadValidatorBonds(store)
	require.Nil(err)
	assert.Equal(validatorBonds, resGet)

}
