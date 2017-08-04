package stake

import (
	"fmt"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/go-wire"
)

const (
	bondAccountKeyPrefix = "ba/"
	bondValueKey         = []byte("bv")
)

func getDelegateeBondsKey(delegatorAddr, validatorPubKey []byte) []byte {
	return []byte(bondAccountKeyPrefix + fmt.Sprintf("%x/%x", delegatorAddr, validatorPubKey))
}

func setDelegateeBonds(store state.SimpleDB, delegatorAddr, validatorPubKey []byte, account *DelegateeBonds) {
	accountBytes := wire.BinaryBytes(account)
	store.Set(getDelegateeBondsKey(delegatorAddr, validatorPubKey), accountBytes)
}

func getDelegateeBonds(store state.SimpleDB, delegatorAddr,
	validatorPubKey []byte) (account *DelegateeBonds, err error) {

	accountBytes := store.Get(getDelegateeBondsKey(delegatorAddr, validatorPubKey))
	if accountBytes == nil {
		return nil
	}
	err := wire.ReadBinaryBytes(accountBytes, account)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}

func removeDelegateeBonds(store state.SimpleDB, delegatorAddr, validatorPubKey []byte) {
	store.Remove(getDelegateeBondsKey(delegatorAddr, validatorPubKey))
}

//////////////////////////////////////////////////////////////////////////////////////////

func setDelegatorBonds(store state.SimpleDB, bondValues AllDelegatorBonds) {
	bvBytes := wire.BinaryBytes(bondValues)
	store.Set(bondValueKey, bvBytes)
}

func getDelegatorBonds(store state.SimpleDB) (bondValues AllDelegatorBonds, err error) {
	bvBytes := store.Get(bondValueKey)
	if bvBytes == nil {
		return make(AllDelegatorBonds, 0)
	}
	err = wire.ReadBinaryBytes(bvBytes, bondValues)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}
