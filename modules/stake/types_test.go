package stake

//import (
//"testing"

//"github.com/stretchr/testify/assert"
//"github.com/stretchr/testify/require"

//"github.com/cosmos/cosmos-sdk"
//"github.com/cosmos/cosmos-sdk/modules/coin"
//)

//// TestState - test the delegatee and delegator bonds store
//func TestTypes(t *testing.T) {
//assert, require := assert.New(t), require.New(t)

//actor1 := sdk.Actor{"testChain", "testapp", []byte("address1")}
//actor2 := sdk.Actor{"testChain", "testapp", []byte("address2")}
//actor3 := sdk.Actor{"testChain", "testapp", []byte("address3")}
//holdActor1 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr1")}
//holdActor2 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr2")}
//holdActor3 := sdk.Actor{"testChain", "testapp", []byte("holdaccountAddr3")}

//delegatee1 :=  DelegateeBond  {
//Delegatee       : actor1,
//Commission      : NewDecimalFromString(0.0001),
//ExchangeRate      : NewDecimalFromString(1),
//TotalBondTokens      : NewDecimalFromString(10),
//Account        : holdActor1,
//VotingPower      : NewDecimalFromString(10),
//}
//delegatee2 :=  DelegateeBond  {
//Delegatee       : actor2,
//Commission      : NewDecimalFromString(0.0002),
//ExchangeRate      : NewDecimalFromString(100),
//TotalBondTokens      : NewDecimalFromString(3),
//Account        : holdActor2,
//VotingPower      : NewDecimalFromString(300),
//}
//delegatee3 :=  DelegateeBond  {
//Delegatee       : actor3,
//Commission      : NewDecimalFromString(0.0003),
//ExchangeRate      : NewDecimalFromString(1),
//TotalBondTokens      : NewDecimalFromString(123),
//Account        : holdActor3,
//VotingPower      : NewDecimalFromString(123),
//}

//delegatees := DelegateeBonds{delegatee1, delegatee2, delegatee3}

////test basic sort
//delegatees.Sort()
//delegatees[0].Delegatee.Equals(actor1)
//delegatees[1].Delegatee.Equals(actor3)
//delegatees[2].Delegatee.Equals(actor2)

////test getting the total voting power
//func (b DelegateeBonds) UpdateVotingPower() (totalPower Decimal) {
////get the existing validator set
//func (b DelegateeBonds) GetValidators(maxVal int) []*abci.Validator {
//
////change the exchange rate and update the voting power
//// resort
//// get the new validator set

//// calculate the difference in the validator set from the origional set
//func (b DelegateeBonds) ValidatorsDiff(previous []*abci.Validator, maxVal int) (diff []*abci.Validator) {

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
//
//func (b DelegatorBonds) Get(delegatee sdk.Actor) (int, *DelegatorBond) {
//func (b DelegatorBonds) Remove(i int) (DelegatorBonds, error) {
//}
