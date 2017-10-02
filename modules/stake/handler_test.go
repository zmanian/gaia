package stake

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
)

func TestRunTxBondGuts(t *testing.T) {

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
	dummySend := func(sender, receiver sdk.Actor, amount coin.Coins) abci.Result {
		dummyAccs[string(sender.Address)] -= amount[0].Amount
		dummyAccs[string(receiver.Address)] += amount[0].Amount
		return abci.OK
	}

	//Add some records to the unbonding queue
	tx := TxBond{BondUpdate{
		Delegatee: actorDelegatee,
		Amount:    coin.Coin{"atom", 10},
	}}

	type args struct {
		store  state.SimpleDB
		tx     TxBond
		sender sdk.Actor
	}
	tests := []struct {
		name    string
		args    args
		wantRes abci.Result
	}{
		{"test 1", args{store, tx, actorDelegator}, abci.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runTxBondGuts(dummySend, tt.args.store, tt.args.tx, tt.args.sender)

			//check the basics
			assert.True(t, got.IsSameCode(tt.wantRes), "runTxBondGuts(%v, %v) = %v, want %v",
				tt.args.tx, tt.args.sender, got, tt.wantRes)

			//TODO check that the accounts and the bond account have the appropriate values
		})
	}
}

func TestRunTxUnbondGuts(t *testing.T) {

	//queue, err := LoadQueue(queueCommissionTypeByte, store)
	//if err != nil {
	//return err
	//}

	type args struct {
		getSender func() (sdk.Actor, abci.Result)
		store     state.SimpleDB
		tx        TxUnbond
		height    uint64
	}
	tests := []struct {
		name    string
		args    args
		wantRes abci.Result
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := runTxUnbondGuts(tt.args.getSender, tt.args.store, tt.args.tx, tt.args.height); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("runTxUnbondGuts(%v, %v, %v, %v) = %v, want %v", tt.args.getSender, tt.args.store, tt.args.tx, tt.args.height, gotRes, tt.wantRes)
			}
		})
	}
}

func TestRunTxNominateGuts(t *testing.T) {
	type args struct {
		bondCoins func(bondAccount sdk.Actor, amount coin.Coins) abci.Result
		store     state.SimpleDB
		tx        TxNominate
	}
	tests := []struct {
		name    string
		args    args
		wantRes abci.Result
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := runTxNominateGuts(tt.args.bondCoins, tt.args.store, tt.args.tx); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("runTxNominateGuts(%v, %v, %v) = %v, want %v", tt.args.bondCoins, tt.args.store, tt.args.tx, gotRes, tt.wantRes)
			}
		})
	}
}

func TestRunTxModCommGuts(t *testing.T) {
	type args struct {
		store  state.SimpleDB
		tx     TxModComm
		height uint64
	}
	tests := []struct {
		name    string
		args    args
		wantRes abci.Result
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := runTxModCommGuts(tt.args.store, tt.args.tx, tt.args.height); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("runTxModCommGuts(%v, %v, %v) = %v, want %v", tt.args.store, tt.args.tx, tt.args.height, gotRes, tt.wantRes)
			}
		})
	}
}
