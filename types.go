package stake

import (
	"bytes"
	"sort"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

const maximumValidators = 100

// BondValue defines an total amount of bond tokens and their exchange rate to
// coins, associated with a single validator. Interest increases the exchange
// rate, and slashing decreases it. When coins are delegated to this validator,
// the delegator is credited with bond tokens based on amount of coins
// delegated and the current exchange rate. Voting power can be calculated as
// total bonds multiplied by exchange rate.
type BondValue struct {
	ValidatorPubKey []byte

	// total bond tokens for this validator'
	Total uint64

	// exchange rate for this validator's bond tokens (in millionths of coins)
	ExchangeRate uint64
}

func (bc BondValue) VotingPower() uint64 {
	return bc.Total * bc.ExchangeRate / PRECISION
}

func (bc BondValue) Validator() *abci.Validator {
	return &abci.Validator{
		PubKey: bc.ValidatorPubKey,
		Power:  bc.VotingPower(),
	}
}

//--------------------------------------------------------------------------------

type BondValues []BondValue

func (bvs BondValues) Len() int {
	return len(bvs)
}

func (bvs BondValues) Less(i, j int) bool {
	vp1, vp2 := bvs[i].VotingPower(), bvs[j].VotingPower()
	if vp1 == vp2 {
		return bytes.Compare(bvs[i].ValidatorPubKey, bvs[j].ValidatorPubKey) == -1
	}
	return vp1 > vp2
}

func (bvs BondValues) Swap(i, j int) {
	bvs[i], bvs[j] = bvs[j], bvs[i]
}

func (bvs BondValues) Sort() {
	sort.Sort(bvs)
}

func (bvs BondValues) Validators() []*abci.Validator {
	validators := make([]*abci.Validator, 0, maximumValidators)
	for i, bv := range bvs {
		if i == maximumValidators {
			break
		}
		validators = append(validators, bv.Validator())
	}
	return validators
}

func (bvs BondValues) Get(validatorPubKey []byte) (int, *BondValue) {
	for i, bv := range bvs {
		if bytes.Equal(bv.ValidatorPubKey, validatorPubKey) {
			return i, &bv
		}
	}
	return 0, nil
}

func (bvs BondValues) Remove(i int) BondValues {
	return append(bvs[:i], bvs[i+1:]...)
}

//--------------------------------------------------------------------------------

// BondAccount defines an account of bond tokens. It is owned by one basecoin
// account, and is associated with one validator.
type BondAccount struct {
	Amount   uint64 // amount of bond tokens
	Sequence int
}

//--------------------------------------------------------------------------------

// Unbond defines an amount of bond tokens which are in the unbonding period
type Unbond struct {
	ValidatorPubKey []byte
	Address         []byte // basecoin account to pay out to
	BondAmount      uint64 // amount of bond tokens which are unbonding
	Height          uint64 // when the unbonding started
}

//--------------------------------------------------------------------------------

// Tx is the interface for stake transactions
type Tx interface{}

// BondTx bonds coins and receives bond tokens
type BondTx struct {
	ValidatorPubKey []byte
	Sequence        int
}

// UnbondTx places bond tokens into the unbonding period
type UnbondTx struct {
	ValidatorPubKey []byte
	BondAmount      uint64
	Sequence        int
}

func wireConcreteType(O interface{}, Byte byte) wire.ConcreteType {
	return wire.ConcreteType{O: O, Byte: Byte}
}

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wireConcreteType(BondTx{}, 0x01),
	wireConcreteType(UnbondTx{}, 0x02),
)
