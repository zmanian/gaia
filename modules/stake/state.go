package stake

import (
	"fmt"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/go-wire"
)

const (
	delegateeKeyPrefix = "de/"
	delegatorKey       = []byte("dr")
)

func getDelegateeBondsKey(delegatorAddr, validatorPubKey []byte) []byte {
	return []byte(delegateeKeyPrefix + fmt.Sprintf("%x/%x", delegatorAddr, validatorPubKey))
}

func setDelegateeBonds(store state.SimpleDB, delegatorAddr, validatorPubKey []byte, delegatee *DelegateeBonds) {
	delegateeBytes := wire.BinaryBytes(delegatee)
	store.Set(getDelegateeBondsKey(delegatorAddr, validatorPubKey), delegateeBytes)
}

func getDelegateeBonds(store state.SimpleDB, delegatorAddr,
	validatorPubKey []byte) (delegatee *DelegateeBonds, err error) {

	delegateeBytes := store.Get(getDelegateeBondsKey(delegatorAddr, validatorPubKey))
	if delegateeBytes == nil {
		return nil
	}
	err := wire.ReadBinaryBytes(delegateeBytes, delegatee)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}

func removeDelegateeBonds(store state.SimpleDB, delegatorAddr, validatorPubKey []byte) {
	store.Remove(getDelegateeBondsKey(delegatorAddr, validatorPubKey))
}

//////////////////////////////////////////////////////////////////////////////////////////

func setDelegatorBonds(store state.SimpleDB, delegatorBonds DelegatorBonds) {
	bvBytes := wire.BinaryBytes(delegatorBonds)
	store.Set(delegatorKey, bvBytes)
}

func getDelegatorBonds(store state.SimpleDB) (delegatorBonds DelegatorBonds, err error) {
	bvBytes := store.Get(delegatorKey)
	if bvBytes == nil {
		return make(DelegatorBonds, 0)
	}
	err = wire.ReadBinaryBytes(bvBytes, delegatorBonds)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}
