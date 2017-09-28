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
)

func Test_processQueueUnbond(t *testing.T) {

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
		{"before unbonding period", args{dummySend, store, 5}, false, 1000, 1000},
		{"just before unbonding period", args{dummySend, store, 9}, false, 1000, 1000},
		{"at unbonding period", args{dummySend, store, 10}, false, 1010, 990},
		{"after unbonding period but not queue", args{dummySend, store, 11}, false, 1010, 990},
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
				"name: %v, error: %v, sendCoins:%v, height: %v, wantDelegatorBal %v but %v, wantBondedBal %v got %v",
				tt.name, err, tt.args.sendCoins, tt.args.height,
				tt.wantDelegatorBal, dummyAccs["delegator"], tt.wantBondedBal, dummyAccs["bonded"])
		})
	}
}

func Test_processQueueCommHistory(t *testing.T) {
	type args struct {
		store  state.SimpleDB
		height uint64
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
			if err := processQueueCommHistory(tt.args.store, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("processQueueCommHistory(%v, %v) error = %v, wantErr %v", tt.args.store, tt.args.height, err, tt.wantErr)
			}
		})
	}
}

func Test_processValidatorRewards(t *testing.T) {
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
