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

	delegator1Bonds := DelegatorBonds{
		{delegatee1, NewDecimal(2, 1)},
		{delegatee2, NewDecimal(3, 1)},
		{delegatee3, NewDecimal(4, 1)},
	}
	delegator2Bonds := DelegatorBonds{
		{delegatee1, NewDecimal(12, 1)},
		{delegatee2, NewDecimal(13, 1)},
		{delegatee3, NewDecimal(14, 1)},
	}
	delegator3Bonds := DelegatorBonds{
		{delegatee1, NewDecimal(22, 1)},
		{delegatee2, NewDecimal(23, 1)},
		{delegatee3, NewDecimal(24, 1)},
	}
	var delegatorNilBonds DelegatorBonds

	delegateeBonds := DelegateeBonds{
		DelegateeBond{
			Delegatee:       delegatee1,
			Commission:      NewDecimal(7, -2),
			ExchangeRate:    NewDecimal(8, -1),
			TotalBondTokens: NewDecimal(9, 1),
			Account:         sdk.Actor{"testChain", "testapp", []byte("addresslockedtoapp")},
		}}
	var delegateeNilBonds DelegateeBonds

	/////////////////////////////////////////////////////////////////////////
	// DelegatorBonds checks

	// Test a basic set and get
	saveDelegatorBonds(store, delegator1, delegator1Bonds)
	resBasicGet, err := loadDelegatorBonds(store, delegator1)
	require.Nil(err)
	assert.Equal(delegator1Bonds, resBasicGet)

	// Set two more records, get and check  them all
	saveDelegatorBonds(store, delegator2, delegator2Bonds)
	saveDelegatorBonds(store, delegator3, delegator3Bonds)

	resGet1, err := loadDelegatorBonds(store, delegator1)
	require.Nil(err)
	resGet2, err := loadDelegatorBonds(store, delegator2)
	require.Nil(err)
	resGet3, err := loadDelegatorBonds(store, delegator3)
	require.Nil(err)

	assert.Equal(delegator1Bonds, resGet1)
	assert.Equal(delegator2Bonds, resGet2)
	assert.Equal(delegator3Bonds, resGet3)

	// Delete one of the record, get and check them all
	removeDelegatorBonds(store, delegator2)

	resGet1, err = loadDelegatorBonds(store, delegator1)
	require.Nil(err)
	resGet2, err = loadDelegatorBonds(store, delegator2)
	require.Nil(err)
	resGet3, err = loadDelegatorBonds(store, delegator3)
	require.Nil(err)

	assert.Equal(delegator1Bonds, resGet1)
	assert.Equal(delegatorNilBonds, resGet2)
	assert.Equal(delegator3Bonds, resGet3)

	// Modify and Set a new delegator bond, check out the mods
	delegator3Bonds = DelegatorBonds{
		{delegatee1, NewDecimal(52, 1)},
		{delegatee2, NewDecimal(53, 1)},
		{delegatee3, NewDecimal(54, 1)},
	}
	saveDelegatorBonds(store, delegator3, delegator3Bonds)
	resBasicGet, err = loadDelegatorBonds(store, delegator3)
	require.Nil(err)
	assert.Equal(delegator3Bonds, resBasicGet)

	/////////////////////////////////////////////////////////////////////////
	// DelegateeBonds checks

	//check the empty store first
	resGet, err := loadDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeNilBonds, resGet)

	//Set and retrieve a record
	saveDelegateeBonds(store, delegateeBonds)
	resGet, err = loadDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeBonds, resGet)

	//modify a records, save, and retrieve
	delegateeBonds[0].Commission = NewDecimal(99, 1)
	saveDelegateeBonds(store, delegateeBonds)
	resGet, err = loadDelegateeBonds(store)
	require.Nil(err)
	assert.Equal(delegateeBonds, resGet)

}
