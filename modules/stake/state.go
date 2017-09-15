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
	atomSupplyKey
)

func loadDelegatorBondsKey(delegator sdk.Actor) []byte {
	delegatorBytes := wire.BinaryBytes(&delegator)
	return append([]byte{delegatorKeyPrefix}, delegatorBytes...)
}

func saveDelegatorBonds(store state.SimpleDB, delegator sdk.Actor, bonds DelegatorBonds) {
	bondsBytes := wire.BinaryBytes(bonds)
	store.Set(loadDelegatorBondsKey(delegator), bondsBytes)
}

func loadDelegatorBonds(store state.SimpleDB,
	delegator sdk.Actor) (bonds DelegatorBonds, err error) {

	delegatorBytes := store.Get(loadDelegatorBondsKey(delegator))
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
	store.Remove(loadDelegatorBondsKey(delegator))
}

//////////////////////////////////////////////////////////////////////////////////////////

func loadDelegateeBonds(store state.SimpleDB) (delegateeBonds DelegateeBonds, err error) {
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

func saveDelegateeBonds(store state.SimpleDB, delegateeBonds DelegateeBonds) {
	b := wire.BinaryBytes(delegateeBonds)
	store.Set([]byte{delegateeKey}, b)
}

//////////////////////////////////////////////////////////////////////////////////////////

//AtomSupply is the total atom supply of the Hub

func loadAtomSupply(store state.SimpleDB) (atoms Decimal, err error) {
	b := store.Get([]byte{atomSupplyKey})
	if b == nil {
		return
	}
	err = wire.ReadBinaryBytes(b, &atoms)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

//TODO need to init the very first atom supply somewhere
func saveAtomSupply(store state.SimpleDB, atoms Decimal) {
	b := wire.BinaryBytes(atoms)
	store.Set([]byte{atomSupplyKey}, b)
}
