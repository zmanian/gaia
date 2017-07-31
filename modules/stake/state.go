package stake

import (
	"fmt"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/state"
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

/////////////////////////////////////////////////////////////

// CounterStore
//--------------------------------------------------------------------------------

// StoreActor - return the basecoin actor for the account
func StoreActor() basecoin.Actor {
	return basecoin.Actor{App: NameCounter, Address: []byte{0x04, 0x20}} //XXX what do these bytes represent? - should use typebyte variables
}

// State - state of the counter applicaton
type State struct {
	Counter   int        `json:"counter"`
	TotalFees coin.Coins `json:"total_fees"`
}

// StateKey - store key for the counter state
func StateKey() []byte {
	return []byte("state")
}

// LoadState - retrieve the counter state from the store
func LoadState(store state.SimpleDB) (state State, err error) {
	bytes := store.Get(StateKey())
	if len(bytes) > 0 {
		err = wire.ReadBinaryBytes(bytes, &state)
		if err != nil {
			return state, errors.ErrDecoding()
		}
	}
	return state, nil
}

// SaveState - save the counter state to the provided store
func SaveState(store state.SimpleDB, state State) error {
	bytes := wire.BinaryBytes(state)
	store.Set(StateKey(), bytes)
	return nil
}
