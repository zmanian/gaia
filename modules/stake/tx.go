package stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	crypto "github.com/tendermint/go-crypto"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDelegate   = 0x55
	ByteTxUndelegate = 0x56
	TypeTxDelegate   = stakingModuleName + "/bond"
	TypeTxUndelegate = stakingModuleName + "/unbond"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDelegate{}, TypeTxDelegate, ByteTxDelegate)
	sdk.TxMapper.RegisterImplementation(TxUndelegate{}, TypeTxUndelegate, ByteTxUndelegate)
}

//Verify interface at compile time
var _, _ sdk.TxInner = &TxDelegate{}, &TxUndelegate{}

// BondUpdate - struct for bonding or unbonding transactions
type BondUpdate struct {
	PubKey crypto.PubKey `json:"pubKey"`
	Bond   coin.Coin     `json:"amount"`
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx BondUpdate) Wrap() sdk.Tx {
	return sdk.Tx{tx}
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

// TxDelegate - struct for bonding transactions
type TxDelegate struct{ BondUpdate }

// NewTxDelegate - new TxDelegate
func NewTxDelegate(bond coin.Coin, pubKey crypto.PubKey) sdk.Tx {
	return TxDelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   bond,
	}}.Wrap()
}

// TxUndelegate - struct for unbonding transactions
type TxUndelegate struct{ BondUpdate }

// NewTxUndelegate - new TxUndelegate
func NewTxUndelegate(bond coin.Coin, pubKey crypto.PubKey) sdk.Tx {
	return TxUndelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   bond,
	}}.Wrap()
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

// TxRevokeCandidacy - struct for unbonding transactions
type TxRevokeCandidacy struct {
	PubKey crypto.PubKey `json:"pubKey"`
}
