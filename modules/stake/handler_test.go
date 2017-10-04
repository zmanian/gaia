package stake

import (
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
)

func TestRunTxBondUnbondGuts(t *testing.T) {

	//set a store with a validator and delegator account
	store := state.NewMemKVStore()
	actorValidator := sdk.Actor{"testChain", "testapp", []byte("validator")}
	actorBonded := sdk.Actor{"testChain", "testapp", []byte("holding")}
	validators := ValidatorBonds{&ValidatorBond{
		Validator:    actorValidator,
		BondedTokens: 0,
		HoldAccount:  actorBonded,
		VotingPower:  0,
	}}
	saveValidatorBonds(store, validators)

	//Setup a simple way to evaluate sending
	dummyAccs := make(map[string]uint64)
	dummyAccs["validator"] = 1000
	dummyAccs["holding"] = 0
	dummySend := func(sender, receiver sdk.Actor, amount coin.Coins) abci.Result {
		dummyAccs[string(sender.Address)] -= uint64(amount[0].Amount)
		dummyAccs[string(receiver.Address)] += uint64(amount[0].Amount)
		return abci.OK
	}

	//Add some records to the unbonding queue
	txBond := TxBond{
		Amount: coin.Coin{"atom", 10},
	}

	type args struct {
		store  state.SimpleDB
		tx     TxBond
		sender sdk.Actor
	}
	tests := []struct {
		name         string
		args         args
		wantRes      abci.Result
		wantBonded   uint64
		wantUnbonded uint64
	}{
		{"test Bond 1", args{store, txBond, actorValidator}, abci.OK, 10, 990},
		{"test Bond 2", args{store, txBond, actorValidator}, abci.OK, 20, 980},
		{"test Bond 3", args{store, txBond, actorValidator}, abci.OK, 30, 970},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runTxBondGuts(dummySend, tt.args.store, tt.args.tx, tt.args.sender)

			validators, err := LoadValidatorBonds(store)
			require.Nil(t, err)

			//check the basics
			assert.True(t, got.IsSameCode(tt.wantRes), "runTxBondGuts(%v, %v, %v) = %v, want %v",
				tt.name, tt.args.tx, tt.args.sender, got, tt.wantRes)

			//Check that the accounts and the bond account have the appropriate values
			assert.Equal(t, tt.wantBonded, validators[0].BondedTokens)
			assert.Equal(t, tt.wantBonded, dummyAccs["holding"])
			assert.Equal(t, tt.wantUnbonded, dummyAccs["validator"])
		})
	}

	/////////////////////////////////////////////////////////////////////
	// Unbonding

	getSender := func() (sender sdk.Actor, res abci.Result) {
		return actorValidator, abci.OK
	}

	//Add some records to the unbonding queue
	txUnbond := TxUnbond{
		Amount: coin.Coin{"atom", 10},
	}
	type args2 struct {
		store state.SimpleDB
		tx    TxUnbond
	}
	tests2 := []struct {
		name         string
		args         args2
		wantRes      abci.Result
		wantBonded   uint64
		wantUnbonded uint64
	}{
		{"test unbond 1", args2{store, txUnbond}, abci.OK, 20, 980},
		{"test unbond 2", args2{store, txUnbond}, abci.OK, 10, 990},
		{"test unbond 3", args2{store, txUnbond}, abci.OK, 0, 1000},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			gotRes := runTxUnbondGuts(getSender, dummySend,
				tt.args.store, tt.args.tx)

			validators, err := LoadValidatorBonds(store)
			require.Nil(t, err)

			//check the basics
			assert.True(t, gotRes.IsSameCode(tt.wantRes),
				"runTxUnbondGuts(%v, %v) = %v, want %v",
				tt.name, tt.args.tx, gotRes, tt.wantRes)

			//Check that the accounts and the bond account have the appropriate values
			assert.Equal(t, tt.wantBonded, validators[0].BondedTokens)
			assert.Equal(t, tt.wantBonded, dummyAccs["holding"])
			assert.Equal(t, tt.wantUnbonded, dummyAccs["validator"])
		})
	}
}
