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
type TxBond struct {
	Amount coin.Coin `json:"amount"`
	PubKey []byte    `json:"pubkey"`
}

// NewTxBond - new TxBond
func NewTxBond(amount coin.Coin, pubKey []byte) sdk.Tx {
	return TxBond{
		Amount: amount,
		PubKey: pubKey,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxBond) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, and valid coins
func (tx TxBond) ValidateBasic() error {
	return validateBasic(tx.Amount)
}

// TxUnbond - struct for unbonding transactions
type TxUnbond struct {
	Amount coin.Coin `json:"amount"`
}

// NewTxUnbond - new TxUnbond
func NewTxUnbond(amount coin.Coin) sdk.Tx {
	return TxUnbond{
		Amount: amount,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxUnbond) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// ValidateBasic - Check for non-empty actor, and valid coins
func (tx TxUnbond) ValidateBasic() error {
	return validateBasic(tx.Amount)
}

func validateBasic(amount coin.Coin) error {
	coins := coin.Coins{amount}
	if !coins.IsValidNonnegative() {
		return coin.ErrInvalidCoins()
	}

	if amount.Denom != bondDenom {
		return fmt.Errorf("Invalid coin denomination")
	}
	if amount.Amount <= 0 {
		return fmt.Errorf("Amount must be > 0")
	}
	return nil
}
