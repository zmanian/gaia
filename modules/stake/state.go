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

func getBondAccountKey(delegatorAddr, validatorPubKey []byte) []byte {
	return []byte(bondAccountKeyPrefix + fmt.Sprintf("%x/%x", delegatorAddr, validatorPubKey))
}

func setBondAccount(db state.SimpleDB, delegatorAddr, validatorPubKey []byte, account *BondAccount) {
	accountBytes := wire.BinaryBytes(account)
	db.Set(getBondAccountKey(delegatorAddr, validatorPubKey), accountBytes)
}

func getBondAccount(db state.SimpleDB, delegatorAddr, validatorPubKey []byte) (account *BondAccount, err error) {
	accountBytes := db.Get(getBondAccountKey(delegatorAddr, validatorPubKey))
	if accountBytes == nil {
		return nil
	}
	err := wire.ReadBinaryBytes(accountBytes, account)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}

func removeBondAccount(db state.SimpleDB, delegatorAddr, validatorPubKey []byte) {
	db.Remove(getBondAccountKey(delegatorAddr, validatorPubKey))
}

func setBondValues(db state.SimpleDB, bondValues BondValues) {
	bvBytes := wire.BinaryBytes(bondValues)
	db.Set(bondValueKey, bvBytes)
}

func getBondValues(db state.SimpleDB) (bondValues BondValues, err error) {
	bvBytes := db.Get(bondValueKey)
	if bvBytes == nil {
		return make(BondValues, 0)
	}
	err = wire.ReadBinaryBytes(bvBytes, bondValues)
	if err != nil {
		return errors.ErrDecoding()
	}
	return
}
