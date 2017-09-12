package stake

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

const (
	delegatorKeyPrefix = iota
	delegateeKey
	validatorSetKey
)

func getDelegatorBondsKey(delegator sdk.Actor) []byte {
	delegatorBytes := wire.BinaryBytes(&delegator)
	return append([]byte{delegatorKeyPrefix}, delegatorBytes...)
}

func setDelegatorBonds(store state.SimpleDB, delegator sdk.Actor, bonds DelegatorBonds) {
	bondsBytes := wire.BinaryBytes(bonds)
	store.Set(getDelegatorBondsKey(delegator), bondsBytes)
}

func getDelegatorBonds(store state.SimpleDB,
	delegator sdk.Actor) (bonds DelegatorBonds, err error) {

	delegatorBytes := store.Get(getDelegatorBondsKey(delegator))
	if delegatorBytes == nil {
		return
	}
	err = wire.ReadBinaryBytes(delegatorBytes, &bonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func removeDelegatorBonds(store state.SimpleDB, delegator sdk.Actor) {
	store.Remove(getDelegatorBondsKey(delegator))
}

//////////////////////////////////////////////////////////////////////////////////////////

func getDelegateeBonds(store state.SimpleDB) (delegateeBonds DelegateeBonds, err error) {
	b := store.Get([]byte{delegateeKey})
	if b == nil {
		return
	}
	err = wire.ReadBinaryBytes(b, &delegateeBonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func setDelegateeBonds(store state.SimpleDB, delegateeBonds DelegateeBonds) {
	b := wire.BinaryBytes(delegateeBonds)
	store.Set([]byte{delegateeKey}, b)
}

//////////////////////////////////////////////////////////////////////////////////////////

func getValidatorSet(store state.SimpleDB) (validators []*abci.Validator, err error) {
	b := store.Get([]byte{validatorSetKey})
	if b == nil {
		return
	}
	err = wire.ReadBinaryBytes(b, &validators)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func setValidatorSet(store state.SimpleDB, validators []*abci.Validator) {
	b := wire.BinaryBytes(validators)
	store.Set([]byte{validatorSetKey}, b)
}
