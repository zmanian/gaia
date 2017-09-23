package stake

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk"
)

// TestState - test the delegatee and delegator bonds store
func TestTypes(t *testing.T) {
	//assert, require := assert.New(t), require.New(t)
	assert := assert.New(t)

	actor1 := sdk.Actor{"testChain", "testapp", []byte("address1")}
	actor2 := sdk.Actor{"testChain", "testapp", []byte("address2")}
	actor3 := sdk.Actor{"testChain", "testapp", []byte("address3")}
	holdActor1 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr1")}
	holdActor2 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr2")}
	holdActor3 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr3")}

	delegatee1 := &DelegateeBond{
		Delegatee:       actor1,
		Commission:      NewDecimal(1, -4),
		ExchangeRate:    NewDecimal(1, 0),
		TotalBondTokens: NewDecimal(10, 0),
		Account:         holdActor1,
		VotingPower:     NewDecimal(10, 0),
	}
	delegatee2 := &DelegateeBond{
		Delegatee:       actor2,
		Commission:      NewDecimal(2, -4),
		ExchangeRate:    NewDecimal(100, 0),
		TotalBondTokens: NewDecimal(3, 0),
		Account:         holdActor2,
		VotingPower:     NewDecimal(300, 0),
	}
	delegatee3 := &DelegateeBond{
		Delegatee:       actor3,
		Commission:      NewDecimal(3, -4),
		ExchangeRate:    NewDecimal(1, 0),
		TotalBondTokens: NewDecimal(123, 0),
		Account:         holdActor3,
		VotingPower:     NewDecimal(123, 0),
	}

	delegatees := DelegateeBonds{delegatee1, delegatee2, delegatee3}

	//test basic sort
	delegatees.Sort()
	//assert.True(delegatees[0].Delegatee.Equals(actor1), "not equal: %v, %v" delegatees[0].DelegateeV//)
	assert.Equal(delegatees[0].Delegatee, actor2)
	assert.Equal(delegatees[1].Delegatee, actor3)
	assert.Equal(delegatees[2].Delegatee, actor1)

	//get the base validators set which will contain all the delegatees
	validators0 := delegatees.GetValidators()
	assert.Equal(3, len(validators0))

	//test to see if the minValBond is functioning
	minValBond = NewDecimal(10000, 0)
	delegatees.UpdateVotingPower()
	assert.Equal(0, len(delegatees.GetValidators()), "%v", delegatees.GetValidators())
	minValBond = NewDecimal(0, 0)
	delegatees.UpdateVotingPower()
	assert.Equal(3, len(delegatees.GetValidators()), "%v", delegatees.GetValidators())
	minValBond = NewDecimal(0, 0)
	delegatees.UpdateVotingPower()

	assert.Equal(NewDecimal(300, 0), delegatees[0].TotalBondTokens.Mul(delegatees[0].ExchangeRate))
	assert.Equal(2, len(delegatees.GetValidators()), "%v, %v, %v,", delegatees[0], delegatees[1], delegatees[2])

	//test to see if the maxVal is functioning
	maxVal = 0
	delegatees.UpdateVotingPower()
	assert.Equal(0, len(delegatees.GetValidators()))
	maxVal = 1
	delegatees.UpdateVotingPower()
	assert.Equal(1, len(delegatees.GetValidators()))
	maxVal = 2
	delegatees.UpdateVotingPower()
	assert.Equal(2, len(delegatees.GetValidators()))

	//test getting the total voting power
	assert.True(delegatees.UpdateVotingPower().Equal(NewDecimal(423, 0)))

	//get/test the existing validator set
	validators1 := delegatees.GetValidators()
	assert.Equal(2, len(validators1))
	assert.True(bytes.Equal(validators1[0].PubKey, actor3.Address))
	assert.True(bytes.Equal(validators1[1].PubKey, actor2.Address))

	//change the exchange rate and update the voting power
	delegatees[0].ExchangeRate = NewDecimal(1000, 0)
	assert.True(delegatees.UpdateVotingPower().Equal(NewDecimal(1300, 0)))

	// resort
	delegatees.Sort()

	// get the new validator set
	validators2 := delegatees.GetValidators()
	assert.Equal(2, len(validators2))
	assert.True(bytes.Equal(validators2[0].PubKey, actor1.Address))
	assert.True(bytes.Equal(validators2[1].PubKey, actor3.Address))

	// calculate the difference in the validator set from the origional set
	diff := delegatees.ValidatorsDiff(validators1)
	assert.Equal(2, len(diff))
	assert.True(diff[0].Power == 0)
	assert.True(diff[1].Power == validators2[2].Power)

	//// test get and remove one of the validators
	//func (b DelegateeBonds) Get(delegatee sdk.Actor) (int, *DelegateeBond) {
	//func (b DelegateeBonds) Remove(i int) (DelegateeBonds, error) {
	////get the final validator set

	////init a few delegator bonds
	//// Get them
	//// Remove one
	//// attempt to get a removed delegator

	//type DelegatorBond struct {
	//Delegatee  sdk.Actor
	//BondTokens Decimal // amount of bond tokens
	//}

	//func (b DelegatorBonds) Get(delegatee sdk.Actor) (int, *DelegatorBond) {
	//func (b DelegatorBonds) Remove(i int) (DelegatorBonds, error) {
}
