package stake

import (
	"fmt"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/go-wire"
)

var (
	delegatorKeyPrefix = "de/"
	delegateeKey       = []byte("dr")
)

func getDelegatorBondsKey(delegateeAddr, validatorPubKey []byte) []byte {
	return []byte(delegatorKeyPrefix + fmt.Sprintf("%x/%x", delegateeAddr, validatorPubKey))
}

func setDelegatorBonds(store state.SimpleDB, delegateeAddr, validatorPubKey []byte, delegator DelegatorBonds) {
	delegatorBytes := wire.BinaryBytes(&delegator)
	store.Set(getDelegatorBondsKey(delegateeAddr, validatorPubKey), delegatorBytes)
}

func getDelegatorBonds(store state.SimpleDB, delegateeAddr,
	validatorPubKey []byte) (delegator DelegatorBonds, err error) {

	delegatorBytes := store.Get(getDelegatorBondsKey(delegateeAddr, validatorPubKey))
	if delegatorBytes == nil {
		return
	}
	err = wire.ReadBinaryBytes(delegatorBytes, delegator)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func removeDelegatorBonds(store state.SimpleDB, delegateeAddr, validatorPubKey []byte) {
	store.Remove(getDelegatorBondsKey(delegateeAddr, validatorPubKey))
}

//////////////////////////////////////////////////////////////////////////////////////////

func setDelegateeBonds(store state.SimpleDB, delegateeBonds DelegateeBonds) {
	bvBytes := wire.BinaryBytes(delegateeBonds)
	store.Set(delegateeKey, bvBytes)
}

func getDelegateeBonds(store state.SimpleDB) (delegateeBonds DelegateeBonds, err error) {
	bvBytes := store.Get(delegateeKey)
	if bvBytes == nil {
		return make(DelegateeBonds, 0), nil
	}
	err = wire.ReadBinaryBytes(bvBytes, delegateeBonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}
