package stake

import (
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/go-wire"
)

//nolint
var BondKey = []byte{0x00}

// LoadValidatorBonds - loads the validator bond set
// TODO ultimately this function should be made unexported... being used right now
// for patchwork of tick functionality therefor much easier if exported until
// the new SDK is created
func LoadValidatorBonds(store state.SimpleDB) (validatorBonds ValidatorBonds, err error) {
	b := store.Get(BondKey)
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
	store.Set(BondKey, b)
}
