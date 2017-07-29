package stake

import (
	"fmt"

	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

const (
	stakeKeyPrefix       = "stake/"
	bondAccountKeyPrefix = stakeKeyPrefix + "ba/"
)

var bondValueKey = []byte(stakeKeyPrefix + "bv")

func bondAccountKey(delegatorAddress []byte, validatorPubKey []byte) []byte {
	return []byte(bondAccountKeyPrefix + fmt.Sprintf("%x/%x", delegatorAddress, validatorPubKey))
}

//------------------------------------------------------------------------

func loadBondAccount(store types.KVStore, delegatorAddress []byte, validatorPubKey []byte) (account *BondAccount) {
	accountKey := bondAccountKey(delegatorAddress, validatorPubKey)
	accountBytes := store.Get(accountKey)
	if accountBytes == nil {
		return nil
	}
	wire.ReadBinaryBytes(accountBytes, account)
	return
}

func storeBondAccount(store types.KVStore, delegatorAddress []byte, validatorPubKey []byte, account *BondAccount) {
	accountKey := bondAccountKey(delegatorAddress, validatorPubKey)
	accountBytes := wire.BinaryBytes(account)
	store.Set(accountKey, accountBytes)
}

func removeBondAccount(store types.KVStore, delegatorAddress []byte, validatorPubKey []byte) {
	// TODO: remove
	storeBondAccount(store, delegatorAddress, validatorPubKey, nil)
}

//------------------------------------------------------------------------

func loadBondValues(store types.KVStore) (bondValues BondValues) {
	bvBytes := store.Get(bondValueKey)
	if bvBytes == nil {
		return make(BondValues, 0)
	}
	wire.ReadBinaryBytes(bvBytes, bondValues)
	return
}

func storeBondValues(store types.KVStore, bondValues BondValues) {
	bvBytes := wire.BinaryBytes(bondValues)
	store.Set(bondValueKey, bvBytes)
}
