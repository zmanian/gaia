package stake

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/go-wire"
)

const (
	delegatorKeyPrefix = iota
	delegateeKey
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

func setDelegateeBonds(store state.SimpleDB, delegateeBonds DelegateeBonds) {
	bvBytes := wire.BinaryBytes(delegateeBonds)
	store.Set([]byte{delegateeKey}, bvBytes)
}

func getDelegateeBonds(store state.SimpleDB) (delegateeBonds DelegateeBonds, err error) {
	bvBytes := store.Get([]byte{delegateeKey})
	if bvBytes == nil {
		return make(DelegateeBonds, 0), nil
	}
	err = wire.ReadBinaryBytes(bvBytes, delegateeBonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}
