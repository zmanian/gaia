package stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
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
type TxBond struct{ BondUpdate }

// NewTxBond - new TxBond
func NewTxBond(delegatee sdk.Actor, amount coin.Coin) sdk.Tx {
	return TxBond{BondUpdate{
		Delegatee: delegatee,
		Amount:    amount,
	}}.Wrap()
}

// TxUnbond - struct for unbonding transactions
type TxUnbond struct{ BondUpdate }

// NewTxUnbond - new TxUnbond
func NewTxUnbond(delegatee sdk.Actor, amount coin.Coin) sdk.Tx {
	return TxUnbond{BondUpdate{
		Delegatee: delegatee,
		Amount:    amount,
	}}.Wrap()
}

// BondUpdate - struct for bonding or unbonding transactions
type BondUpdate struct {
	Delegatee sdk.Actor `json:"delegatee"`
	Amount    coin.Coin `json:"amount"`
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx BondUpdate) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, and valid coins
func (tx BondUpdate) ValidateBasic() error {
	if tx.Delegatee.Empty() {
		return errValidatorEmpty
	}

	coins := coin.Coins{tx.Amount}
	if !coins.IsValidNonnegative() {
		return coin.ErrInvalidCoins()
	}

	bondCoin := tx.Amount
	bondAmt := NewDecimal(bondCoin.Amount, 1)
	if bondCoin.Denom != bondDenom {
		return fmt.Errorf("Invalid coin denomination")
	}
	if bondAmt.LTE(Zero) {
		return fmt.Errorf("Amount must be > 0")
	}

	return nil
}

/////////////////////////////////////////////////////////////////
// TxNominate

// TxNominate - struct for all staking transactions
type TxNominate struct {
	Nominee    sdk.Actor `json:"nominee"`
	Amount     coin.Coin `json:"amount"`
	Commission Decimal   `json:"commission"`
}

// NewTxNominate - return a new transaction for validator self-nomination
func NewTxNominate(nominee sdk.Actor, amount coin.Coin, commission Decimal) sdk.Tx {
	return TxNominate{
		Nominee:    nominee,
		Amount:     amount,
		Commission: commission,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxNominate) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, valid coins, and valid commission range
func (tx TxNominate) ValidateBasic() error {
	if tx.Nominee.Empty() {
		return errValidatorEmpty
	}
	coins := coin.Coins{tx.Amount}
	if !coins.IsValidNonnegative() {
		return coin.ErrInvalidCoins()
	}
	if tx.Commission.LT(NewDecimal(0, 1)) {
		return errCommissionNegative
	}
	if tx.Commission.GT(NewDecimal(1, 1)) {
		return errCommissionHuge
	}
	return nil
}

/////////////////////////////////////////////////////////////////
// TxModComm

// TxModComm - struct for all staking transactions
type TxModComm struct {
	Delegatee  sdk.Actor `json:"delegatee"`
	Commission Decimal   `json:"commission"`
}

// NewTxModComm - return a new counter transaction struct wrapped as a sdk transaction
func NewTxModComm(delegatee sdk.Actor, commission Decimal) sdk.Tx {
	return TxModComm{
		Delegatee:  delegatee,
		Commission: commission,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxModComm) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, and valid commission range
func (tx TxModComm) ValidateBasic() error {
	if tx.Delegatee.Empty() {
		return errValidatorEmpty
	}
	if tx.Commission.LT(NewDecimal(0, 1)) {
		return errCommissionNegative
	}
	if tx.Commission.GT(NewDecimal(1, 1)) {
		return errCommissionHuge
	}
	return nil
}
