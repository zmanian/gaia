package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// TestState - test the delegatee and delegator bonds store
func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := state.NewMemKVStore()

	delegatee1 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee1")}
	delegatee2 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee2")}
	delegatee3 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee3")}

	delegator1 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator1")}
	delegator2 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator2")}
	delegator3 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator3")}

	delegator1Bonds := DelegatorBonds{{delegatee1, 2}, {delegatee2, 3}, {delegatee3, 4}}
	delegator2Bonds := DelegatorBonds{{delegatee1, 12}, {delegatee2, 13}, {delegatee3, 14}}
	delegator3Bonds := DelegatorBonds{{delegatee1, 22}, {delegatee2, 23}, {delegatee3, 24}}
	var delegatorNilBonds DelegatorBonds

	delegateeBonds := DelegateeBonds{
		DelegateeBond{
			Delegatee:       delegatee1,
			Commission:      7,
			ExchangeRate:    8,
			TotalBondTokens: 9,
			Account:         sdk.Actor{"testChain", "testapp", []byte("addresslockedtoapp")},
		}}
	var delegateeNilBonds DelegateeBonds

	/////////////////////////////////////////////////////////////////////////
	// DelegatorBonds checks

	// Test a basic set and get
	setDelegatorBonds(store, delegator1, delegator1Bonds)
	resBasicGet, err := getDelegatorBonds(store, delegator1)
	require.Nil(err)
	assert.Equal(delegator1Bonds, resBasicGet)

	// Set two more records, get and check  them all
	setDelegatorBonds(store, delegator2, delegator2Bonds)
	setDelegatorBonds(store, delegator3, delegator3Bonds)

	resGet1, err := getDelegatorBonds(store, delegator1)
	require.Nil(err)
	resGet2, err := getDelegatorBonds(store, delegator2)
	require.Nil(err)
	resGet3, err := getDelegatorBonds(store, delegator3)
	require.Nil(err)

	assert.Equal(delegator1Bonds, resGet1)
	assert.Equal(delegator2Bonds, resGet2)
	assert.Equal(delegator3Bonds, resGet3)

	// Delete one of the record, get and check them all
	removeDelegatorBonds(store, delegator2)

	resGet1, err = getDelegatorBonds(store, delegator1)
	require.Nil(err)
	resGet2, err = getDelegatorBonds(store, delegator2)
	require.Nil(err)
	resGet3, err = getDelegatorBonds(store, delegator3)
	require.Nil(err)

	assert.Equal(delegator1Bonds, resGet1)
	assert.Equal(delegatorNilBonds, resGet2)
	assert.Equal(delegator3Bonds, resGet3)

	// Modify and Set a new delegator bond, check out the mods
	delegator3Bonds = DelegatorBonds{{delegatee1, 54}, {delegatee2, 53}, {delegatee3, 54}}
	setDelegatorBonds(store, delegator3, delegator3Bonds)
	resBasicGet, err = getDelegatorBonds(store, delegator3)
	require.Nil(err)
	assert.Equal(delegator3Bonds, resBasicGet)

	/////////////////////////////////////////////////////////////////////////
	// DelegateeBonds checks

	//check the empty store first
	resGet, err := getDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeNilBonds, resGet)

	//Set and retrieve a record
	setDelegateeBonds(store, delegateeBonds)
	resGet, err = getDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeBonds, resGet)

	//modify a records, save, and retrieve
	delegateeBonds[0].Commission = 99
	setDelegateeBonds(store, delegateeBonds)
	resGet, err = getDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeBonds, resGet)

}
