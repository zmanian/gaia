package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	crypto "github.com/tendermint/go-crypto"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy = 0x55
	ByteTxDelegate         = 0x56
	ByteTxUnbond           = 0x57
	TypeTxDeclareCandidacy = stakingModuleName + "/declareCandidacy"
	TypeTxDelegate         = stakingModuleName + "/delegate"
	TypeTxUnbond           = stakingModuleName + "/unbond"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxDelegate{}, TypeTxDelegate, ByteTxDelegate)
	sdk.TxMapper.RegisterImplementation(TxUnbond{}, TypeTxUnbond, ByteTxUnbond)
}

//Verify interface at compile time
var _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxDelegate{}, &TxUnbond{}

// BondUpdate - struct for bonding or unbonding transactions
type BondUpdate struct {
	PubKey crypto.PubKey `json:"pubKey"`
	Bond   coin.Coin     `json:"amount"`
}

// ValidateBasic - Check for non-empty actor, and valid coins
func (tx BondUpdate) ValidateBasic() error {
	if tx.PubKey.Empty() { // TODO will an empty validator actually have len 0?
		return errCandidateEmpty
	}

	coins := coin.Coins{tx.Bond}
	if !coins.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !coins.IsPositive() {
		return fmt.Errorf("Amount must be > 0")
	}
	return nil
}

// TxDeclareCandidacy - struct for unbonding transactions
type TxDeclareCandidacy struct{ BondUpdate }

// NewTxDeclareCandidacy - new TxDeclareCandidacy
func NewTxDeclareCandidacy(bond coin.Coin, pubKey crypto.PubKey) sdk.Tx {
	return TxDeclareCandidacy{BondUpdate{
		PubKey: pubKey,
		Bond:   bond,
	}}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxDelegate - struct for bonding transactions
type TxDelegate struct{ BondUpdate }

// NewTxDelegate - new TxDelegate
func NewTxDelegate(bond coin.Coin, pubKey crypto.PubKey) sdk.Tx {
	return TxDelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   bond,
	}}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDelegate) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxUnbond - struct for unbonding transactions
type TxUnbond struct {
	PubKey crypto.PubKey `json:"pubKey"`
	Shares uint64        `json:"amount"`
}

// NewTxUnbond - new TxUnbond
func NewTxUnbond(shares uint64, pubKey crypto.PubKey) sdk.Tx {
	return TxUnbond{
		PubKey: pubKey,
		Shares: shares,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxUnbond) Wrap() sdk.Tx { return sdk.Tx{tx} }

// ValidateBasic - Check for non-empty actor, positive shares
func (tx TxUnbond) ValidateBasic() error {
	if tx.PubKey.Empty() { // TODO will an empty validator actually have len 0?
		return errCandidateEmpty
	}

	if tx.Shares == 0 {
		return fmt.Errorf("Shares must be > 0")
	}
	return nil
}
