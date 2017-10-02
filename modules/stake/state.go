package stake

import (
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/go-wire"
)

func loadValidatorBonds(store state.SimpleDB) (validatorBonds ValidatorBonds, err error) {
	b := store.Get([]byte{0x00})
	if b == nil {
		return
	}
	err = wire.ReadBinaryBytes(b, &validatorBonds)
	if err != nil {
		err = errors.ErrDecoding()
	}
	return
}

func saveValidatorBonds(store state.SimpleDB, validatorBonds ValidatorBonds) {
	b := wire.BinaryBytes(validatorBonds)
	store.Set([]byte{0x00}, b)
}
