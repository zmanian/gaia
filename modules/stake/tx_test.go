package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"

	crypto "github.com/tendermint/go-crypto"
)

var (
	validator = sdk.Actor{"testChain", "testapp", []byte("addressvalidator1")}
	empty     sdk.Actor

	coinPos          = coin.Coin{"fermion", 1000}
	coinZero         = coin.Coin{"fermion", 0}
	coinNeg          = coin.Coin{"fermion", -10000}
	coinPosNotAtoms  = coin.Coin{"foo", 10000}
	coinZeroNotAtoms = coin.Coin{"foo", 0}
	coinNegNotAtoms  = coin.Coin{"foo", -10000}
)

func TestBondUpdateValidateBasic(t *testing.T) {
	type fields struct {
		PubKey crypto.PubKey
		Bond   coin.Coin
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"basic good", fields{pk1, coinPos}, false},
		{"empty delegator", fields{crypto.PubKey{}, coinPos}, true},
		{"zero coin", fields{pk1, coinZero}, true},
		{"neg coin", fields{pk1, coinNeg}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := TxDelegate{BondUpdate{
				PubKey: tt.fields.PubKey,
				Bond:   tt.fields.Bond,
			}}
			assert.Equal(t, tt.wantErr, tx.ValidateBasic() != nil,
				"test: %v, tx.ValidateBasic: %v", tt.name, tx.ValidateBasic())
		})
	}
}
