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

	validatorBonds := ValidatorBonds{
		&ValidatorBond{
			Sender:       validator1,
			PubKey:       []byte{},
			BondedTokens: 9,
			HoldAccount:  sdk.Actor{"testChain", "testapp", []byte("addresslockedtoapp")},
		}}
	var validatorNilBonds ValidatorBonds

	/////////////////////////////////////////////////////////////////////////
	// ValidatorBonds checks

	//check the empty store first
	resGet := LoadBonds(store)
	assert.Equal(validatorNilBonds, resGet)

	//Set and retrieve a record
	saveBonds(store, validatorBonds)
	resGet = LoadBonds(store)
	assert.Equal(validatorBonds, resGet)

	//modify a records, save, and retrieve
	validatorBonds[0].BondedTokens = 99
	saveBonds(store, validatorBonds)
	resGet = LoadBonds(store)
	assert.Equal(validatorBonds, resGet)

}
