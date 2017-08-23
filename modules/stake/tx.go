package stake

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with it's validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxBond     = 0x55
	ByteTxUnbond   = 0x56
	ByteTxNominate = 0x57
	ByteTxModComm  = 0x58
	TypeTxBond     = name + "/bond"
	TypeTxUnbond   = name + "/unbond"
	TypeTxNominate = name + "/nominate"
	TypeTxModComm  = name + "/modComm" //modify commission rate
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxBond{}, TypeTxBond, ByteTxBond)
	sdk.TxMapper.RegisterImplementation(TxUnbond{}, TypeTxUnbond, ByteTxUnbond)
	sdk.TxMapper.RegisterImplementation(TxNominate{}, TypeTxNominate, ByteTxNominate)
	sdk.TxMapper.RegisterImplementation(TxModComm{}, TypeTxModComm, ByteTxModComm)
}

//Verify interface at compile time
var _, _, _, _ sdk.TxInner = &TxBond{}, &TxUnbond{}, &TxNominate{}, &TxModComm{}

/////////////////////////////////////////////////////////////////
// TxBond

// TxBond - struct for bonding transactions
type TxBond struct{ TxBonding }

// NewTxBond - new TxBond
func NewTxBond(validator sdk.Actor, amount coin.Coin) sdk.Tx {
	return TxBond{TxBonding{
		Validator: validator,
		Amount:    amount,
	}}.Wrap()
}

// TxUnbond - struct for unbonding transactions
type TxUnbond struct{ TxBonding }

// NewTxUnbond - new TxUnbond
func NewTxUnbond(validator sdk.Actor, amount coin.Coin) sdk.Tx {
	return TxUnbond{TxBonding{
		Validator: validator,
		Amount:    amount,
	}}.Wrap()
}

// TxBonding - struct for bonding or unbonding transactions
type TxBonding struct {
	Validator sdk.Actor `json:"validator"`
	Amount    coin.Coin `json:"amount"`
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxBonding) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check the bonding coins, Validator is non-empty
func (tx TxBonding) ValidateBasic() error {
	if tx.Validator.Empty() {
		return errValidatorEmpty
	}
	coins := coin.Coins{tx.Amount}
	if !coins.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !coins.IsNonnegative() {
		return coin.ErrInvalidCoins()
	}
	return nil
}

/////////////////////////////////////////////////////////////////
// TxNominate

// TxNominate - struct for all staking transactions
type TxNominate struct {
	Validator  sdk.Actor `json:"validator"`
	Amount     coin.Coin `json:"amount"`
	Commission uint64    `json:"commission"`
}

// NewTxNominate - return a new counter transaction struct wrapped as a sdk transaction
func NewTxNominate(validator sdk.Actor, amount coin.Coin, commission uint64) sdk.Tx {
	return TxNominate{
		Validator:  validator,
		Amount:     amount,
		Commission: commission,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxNominate) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check coins as well as that the validator is actually a validator
// TODO validate commission is not negative and valid
func (tx TxNominate) ValidateBasic() error {
	if tx.Validator.Empty() {
		return errValidatorEmpty
	}
	coins := coin.Coins{tx.Amount}
	if !coins.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !coins.IsNonnegative() {
		return coin.ErrInvalidCoins()
	}
	return nil
}

/////////////////////////////////////////////////////////////////
// TxModComm

// TxModComm - struct for all staking transactions
type TxModComm struct {
	Validator  sdk.Actor `json:"validator"`
	Commission uint64    `json:"commission"`
}

// NewTxModComm - return a new counter transaction struct wrapped as a sdk transaction
func NewTxModComm(validator sdk.Actor, commission uint64) sdk.Tx {
	return TxModComm{
		Validator:  validator,
		Commission: commission,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxModComm) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check coins as well as that the validator is actually a validator
// TODO validate commission is not negative and valid
func (tx TxModComm) ValidateBasic() error {
	if tx.Validator.Empty() {
		return errValidatorEmpty
	}
	return nil
}
