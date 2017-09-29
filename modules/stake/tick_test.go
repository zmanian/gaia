package stake

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

func TestProcessQueueUnbond(t *testing.T) {

	//set the unbonding period for testing
	periodUnbonding = 10

	//set a store with a delegatee and delegator account
	store := state.NewMemKVStore()
	actorDelegatee := sdk.Actor{"testChain", "testapp", []byte("delegatee")}
	actorDelegator := sdk.Actor{"testChain", "testapp", []byte("delegator")}
	actorBonded := sdk.Actor{"testChain", "testapp", []byte("bonded")}
	delegatees := DelegateeBonds{&DelegateeBond{
		Delegatee:       actorDelegatee,
		Commission:      NewDecimal(1, -4),
		ExchangeRate:    NewDecimal(1, 0),
		TotalBondTokens: NewDecimal(10000, 0),
		Account:         actorBonded,
		VotingPower:     NewDecimal(10000, 0),
	}}
	saveDelegateeBonds(store, delegatees)
	delegators := DelegatorBonds{&DelegatorBond{
		Delegatee:  actorDelegatee,
		BondTokens: NewDecimal(500, 0),
	}}
	saveDelegatorBonds(store, actorDelegator, delegators)

	//Setup a simple way to evaluate sending
	dummyAccs := make(map[string]int64)
	dummyAccs["delegator"] = 1000
	dummyAccs["bonded"] = 1000
	dummySend := func(sender, receiver sdk.Actor, amount coin.Coins) error {
		dummyAccs[string(sender.Address)] -= amount[0].Amount
		dummyAccs[string(receiver.Address)] += amount[0].Amount
		return nil
	}

	//Add some records to the unbonding queue
	tx := TxUnbond{BondUpdate{
		Delegatee: actorDelegatee,
		Amount:    coin.Coin{"atom", 10},
	}}
	res := runTxUnbondGuts(func() (sdk.Actor, abci.Result) { return actorDelegator, abci.OK }, store, tx, 0)
	require.True(t, res.IsOK(), "%v", res)
	res = runTxUnbondGuts(func() (sdk.Actor, abci.Result) { return actorDelegator, abci.OK }, store, tx, 15)
	require.True(t, res.IsOK(), "%v", res)
	res = runTxUnbondGuts(func() (sdk.Actor, abci.Result) { return actorDelegator, abci.OK }, store, tx, 50)
	require.True(t, res.IsOK(), "%v", res)

	type args struct {
		sendCoins func(sender, receiver sdk.Actor, amount coin.Coins) error
		store     state.SimpleDB
		height    uint64
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantDelegatorBal int64
		wantBondedBal    int64
	}{
		{"before unbonding period 1", args{dummySend, store, 5}, false, 1000, 1000},
		{"just before unbonding period 1", args{dummySend, store, 9}, false, 1000, 1000},
		{"at unbonding period 1", args{dummySend, store, 10}, false, 1010, 990},
		{"after unbonding period 1", args{dummySend, store, 11}, false, 1010, 990},
		{"before unbonding period 2", args{dummySend, store, 24}, false, 1010, 990},
		{"at unbonding period 2", args{dummySend, store, 25}, false, 1020, 980},
		{"after unbonding period 2", args{dummySend, store, 26}, false, 1020, 980},
		{"misc", args{dummySend, store, 40}, false, 1020, 980},
		{"before unbonding period 3", args{dummySend, store, 59}, false, 1020, 980},
		{"at unbonding period 3", args{dummySend, store, 60}, false, 1030, 970},
		{"after unbonding period 3 with empty queue", args{dummySend, store, 61}, false, 1030, 970},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processQueueUnbond(tt.args.sendCoins, tt.args.store, tt.args.height)
			if err == nil && dummyAccs["delegator"] != tt.wantDelegatorBal {
				err = fmt.Errorf("unexpected sender balance")
			}
			if err == nil && dummyAccs["bonded"] != tt.wantBondedBal {
				err = fmt.Errorf("unexpected receiver balance")
			}
			assert.Equal(t, tt.wantErr, err != nil,
				"name: %v, error: %v, sendCoins:%v, height: %v,"+
					"wantDelegatorBal %v got %v, wantBondedBal %v got %v",
				tt.name, err, tt.args.sendCoins, tt.args.height,
				tt.wantDelegatorBal, dummyAccs["delegator"], tt.wantBondedBal, dummyAccs["bonded"])
		})
	}
}

func TestProcessQueueCommHistory(t *testing.T) {

	// maximum total commission permitted across the queued commission history
	periodCommHistory = 10

	//set a store with a delegatee account
	store := state.NewMemKVStore()
	actorDelegatee := sdk.Actor{"testChain", "testapp", []byte("delegatee")}

	//Add some records to the unbonding queue
	queue, err := LoadQueue(queueCommissionTypeByte, store)
	require.Nil(t, err)
	addToQueue := func(commChange Decimal, height uint64) {
		queueElem := QueueElemCommChange{
			QueueElem:  QueueElem{actorDelegatee, height},
			CommChange: commChange,
		}
		bytes := wire.BinaryBytes(queueElem)
		queue.Push(bytes)
		//require.Equal(t, queue.Peek(), bytes)
	}
	addToQueue(NewDecimal(1, -1), 1)
	addToQueue(NewDecimal(2, -1), 2)
	addToQueue(NewDecimal(25, -2), 3)
	addToQueue(NewDecimal(-10, -2), 4)

	type args struct {
		store  state.SimpleDB
		height uint64
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		wantModHistory Decimal
	}{
		{"height 9, none popped", args{store, 9}, false, NewDecimal(45, -2)},
		{"height 10, none popped", args{store, 10}, false, NewDecimal(45, -2)},
		{"height 11, no changes", args{store, 11}, false, NewDecimal(35, -2)},
		{"height 12, no changes", args{store, 12}, false, NewDecimal(15, -2)},
		{"height 13, no changes", args{store, 13}, false, NewDecimal(-10, -2)},
		{"height 14, no changes", args{store, 14}, false, NewDecimal(0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processQueueCommHistory(queue, tt.args.height)
			//panic(fmt.Sprintf("%v", queue.Peek()))
			sum, res := getQueueSum(queue, actorDelegatee)
			if !res.IsOK() {
				err = res
			}
			if err == nil && !sum.Equal(tt.wantModHistory) {
				err = fmt.Errorf("unexpected sender balance")
			}
			assert.Equal(t, tt.wantErr, err != nil,
				"name: %v, error: %v, height: %v, wantModHistory %v got %v",
				tt.name, err, tt.args.height, tt.wantModHistory, sum)
		})
	}
}

func TestProcessValidatorRewards(t *testing.T) {
	type args struct {
		creditAcc        func(receiver sdk.Actor, amount coin.Coins) error
		store            state.SimpleDB
		height           uint64
		totalVotingPower Decimal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := processValidatorRewards(tt.args.creditAcc, tt.args.store, tt.args.height, tt.args.totalVotingPower); (err != nil) != tt.wantErr {
				t.Errorf("processValidatorRewards(%v, %v, %v, %v) error = %v, wantErr %v", tt.args.creditAcc, tt.args.store, tt.args.height, tt.args.totalVotingPower, err, tt.wantErr)
			}
		})
	}
}
