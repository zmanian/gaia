package stake

import (
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/stretchr/testify/assert"
)

var (
	validator = sdk.Actor{"testChain", "testapp", []byte("addressvalidator1")}
	empty     sdk.Actor

	coinPos          = coin.Coin{"atom", 1000}
	coinZero         = coin.Coin{"atom", 0}
	coinNeg          = coin.Coin{"atom", -10000}
	coinPosNotAtoms  = coin.Coin{"foo", 10000}
	coinZeroNotAtoms = coin.Coin{"foo", 0}
	coinNegNotAtoms  = coin.Coin{"foo", -10000}
)

func TestBondUpdateValidateBasic(t *testing.T) {
	type fields struct {
		Amount coin.Coin
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"basic good", fields{coinPos}, false},
		{"empty delegator", fields{coinPos}, false},
		{"zero coin", fields{coinZero}, true},
		{"neg coin", fields{coinNeg}, true},
		{"pos coin, non-atom denom", fields{coinPosNotAtoms}, true},
		{"zero coin, non-atom denom", fields{coinZeroNotAtoms}, true},
		{"neg coin, non-atom denom", fields{coinNegNotAtoms}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BondUpdate{
				Amount: tt.fields.Amount,
			}
			assert.Equal(t, tt.wantErr, tx.ValidateBasic() != nil, tt.name)
		})
	}
}
