package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// this makes sure that txs are rejected with invalid data or permissions
func TestState(t *testing.T) {
	//assert, require := assert.New(t), require.New(t)
	assert := assert.New(t)

	store := state.NewMemKVStore()

	delegatee1 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee1")}
	delegatee2 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee2")}
	delegatee3 := sdk.Actor{"testChain", "testapp", []byte("addressdelegatee3")}

	delegator1 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator1")}
	//delegator2 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator2")}
	//delegator3 := sdk.Actor{"testChain", "testapp", []byte("addressdelegator3")}

	delegator1Bonds := DelegatorBonds{{delegatee1, 2}, {delegatee2, 3}, {delegatee3, 4}}
	setDelegatorBonds(store, delegator1, delegator1Bonds)
	res, err := getDelegatorBonds(store, delegator1)
	assert.Nil(err)
	assert.Equal(delegator1Bonds, res)
}
