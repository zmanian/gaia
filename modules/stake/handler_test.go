package stake

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
)

func TestRunTxBondGuts(t *testing.T) {
	type args struct {
		sendCoins func(receiver sdk.Actor, amount coin.Coins) abci.Result
		store     state.SimpleDB
		tx        TxBond
		sender    sdk.Actor
	}
	tests := []struct {
		name string
		args args
		want abci.Result
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runTxBondGuts(tt.args.sendCoins, tt.args.store, tt.args.tx, tt.args.sender); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runTxBondGuts(%v, %v, %v, %v) = %v, want %v", tt.args.sendCoins, tt.args.store, tt.args.tx, tt.args.sender, got, tt.want)
			}
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
