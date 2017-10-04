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
	ByteTxBond   = 0x55
	ByteTxUnbond = 0x56
	TypeTxBond   = name + "/bond"
	TypeTxUnbond = name + "/unbond"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxBond{}, TypeTxBond, ByteTxBond)
	sdk.TxMapper.RegisterImplementation(TxUnbond{}, TypeTxUnbond, ByteTxUnbond)
}

//Verify interface at compile time
var _, _ sdk.TxInner = &TxBond{}, &TxUnbond{}

/////////////////////////////////////////////////////////////////
// TxBond

// TxBond - struct for bonding transactions
type TxBond struct{ BondUpdate }

// NewTxBond - new TxBond
func NewTxBond(amount coin.Coin) sdk.Tx {
	return TxBond{BondUpdate{
		Amount: amount,
	}}.Wrap()
}

// TxUnbond - struct for unbonding transactions
type TxUnbond struct{ BondUpdate }

// NewTxUnbond - new TxUnbond
func NewTxUnbond(amount coin.Coin) sdk.Tx {
	return TxUnbond{BondUpdate{
		Amount: amount,
	}}.Wrap()
}

// BondUpdate - struct for bonding or unbonding transactions
type BondUpdate struct {
	Amount coin.Coin `json:"amount"`
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx BondUpdate) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, and valid coins
func (tx BondUpdate) ValidateBasic() error {
	coins := coin.Coins{tx.Amount}
	if !coins.IsValidNonnegative() {
		return coin.ErrInvalidCoins()
	}

	bondCoin := tx.Amount
	bondAmt := bondCoin.Amount
	if bondCoin.Denom != bondDenom {
		return fmt.Errorf("Invalid coin denomination")
	}
	if bondAmt <= 0 {
		return fmt.Errorf("Amount must be > 0")
	}
	return nil
}
