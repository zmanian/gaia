package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
)

func TestRunTxBondUnbond(t *testing.T) {
	assert := assert.New(t)

	//set a store with a validator and delegator account
	store := state.NewMemKVStore()
	validator := sdk.Actor{"testChain", "testapp", []byte("validator")}
	holder := getHoldAccount(validator)
	validators := ValidatorBonds{&ValidatorBond{
		Validator:    validator,
		BondedTokens: 0,
		HoldAccount:  holder,
		VotingPower:  0,
	}}
	saveBonds(store, validators)

	//Setup a simple way to evaluate sending
	dummyAccs := make(map[string]uint64)
	dummyAccs[string(validator.Address)] = 1000
	dummyAccs[string(holder.Address)] = 0
	dummySend := func(sender, receiver sdk.Actor, amount coin.Coins) abci.Result {
		dummyAccs[string(sender.Address)] -= uint64(amount[0].Amount)
		dummyAccs[string(receiver.Address)] += uint64(amount[0].Amount)
		return abci.OK
	}

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
		{"test bond 1", args{store, txBond, validator}, abci.OK, 10, 990},
		{"test bond 2", args{store, txBond, validator}, abci.OK, 20, 980},
		{"test bond 3", args{store, txBond, validator}, abci.OK, 30, 970},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			holder := getHoldAccount(tt.args.sender)
			got := runTxBond(tt.args.store, tt.args.sender, holder, dummySend, tt.args.tx)

			//check the basics
			assert.True(got.IsSameCode(tt.wantRes), "runTxBondGuts(%v, %v, %v) = %v, want %v",
				tt.name, tt.args.tx, tt.args.sender, got, tt.wantRes)

			//Check that the accounts and the bond account have the appropriate values
			validators := LoadBonds(store)
			assert.Equal(tt.wantBonded, validators[0].BondedTokens, "%v, %v, %v",
				tt.name, tt.wantBonded, validators[0].BondedTokens)
			assert.Equal(tt.wantBonded, dummyAccs[string(holder.Address)], "%v, %v, %v",
				tt.name, tt.wantBonded, dummyAccs[string(holder.Address)])
			assert.Equal(tt.wantUnbonded, dummyAccs[string(validator.Address)], "%v, %v, %v",
				tt.name, tt.wantUnbonded, dummyAccs[string(validator.Address)])
		})
	}

	/////////////////////////////////////////////////////////////////////
	// Unbonding

	txUnbond := TxUnbond{
		Amount: coin.Coin{"atom", 10},
	}
	type args2 struct {
		store  state.SimpleDB
		tx     TxUnbond
		sender sdk.Actor
	}
	tests2 := []struct {
		name            string
		args            args2
		wantRes         abci.Result
		validatorExists bool
		wantBonded      uint64
		wantUnbonded    uint64
	}{
		{"test unbond 1", args2{store, txUnbond, validator}, abci.OK, true, 20, 980},
		{"test unbond 2", args2{store, txUnbond, validator}, abci.OK, true, 10, 990},
		{"test unbond 3", args2{store, txUnbond, validator}, abci.OK, false, 0, 1000},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {

			holder := getHoldAccount(tt.args.sender)
			got := runTxUnbond(tt.args.store, tt.args.sender, holder, dummySend, tt.args.tx)

			//check the basics
			assert.True(got.IsSameCode(tt.wantRes),
				"runTxUnbondGuts(%v, %v) = %v, want %v",
				tt.name, tt.args.tx, got, tt.wantRes)

			//Check that the accounts and the bond account have the appropriate values
			validators := LoadBonds(store)
			if tt.validatorExists {
				assert.Equal(tt.wantBonded, validators[0].BondedTokens, "%v, %v, %v",
					tt.name, tt.wantBonded, validators[0].BondedTokens)
			}
			assert.Equal(tt.wantBonded, dummyAccs[string(holder.Address)], "%v, %v, %v",
				tt.name, tt.wantBonded, dummyAccs[string(holder.Address)])
			assert.Equal(tt.wantUnbonded, dummyAccs[string(validator.Address)], "%v, %v, %v",
				tt.name, tt.wantUnbonded, dummyAccs[string(validator.Address)])
		})
	}
}
