package stake

import (
	"testing"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/stretchr/testify/assert"
)

var (
	delegatee = sdk.Actor{"testChain", "testapp", []byte("addressdelegatee1")}
	empty     sdk.Actor

	coinPos          = coin.Coin{"atom", 1000}
	coinZero         = coin.Coin{"atom", 0}
	coinNeg          = coin.Coin{"atom", -10000}
	coinPosNotAtoms  = coin.Coin{"foo", 10000}
	coinZeroNotAtoms = coin.Coin{"foo", 0}
	coinNegNotAtoms  = coin.Coin{"foo", -10000}

	comm50    = NewDecimal(5, -1)
	commNeg50 = NewDecimal(-5, -1)
	comm0     = NewDecimal(0, 0)
	comm100   = NewDecimal(1, 0)
	comm110   = NewDecimal(11, -1)
)

// TestBondUpdateValidateBasic - sdfds
func TestBondUpdateValidateBasic(t *testing.T) {
	type fields struct {
		Delegatee sdk.Actor
		Amount    coin.Coin
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"basic good", fields{delegatee, coinPos}, false},
		{"empty delegator", fields{empty, coinPos}, true},
		{"zero coin", fields{delegatee, coinZero}, true},
		{"neg coin", fields{delegatee, coinNeg}, true},
		{"pos coin, non-atom denom", fields{delegatee, coinPosNotAtoms}, true},
		{"zero coin, non-atom denom", fields{delegatee, coinZeroNotAtoms}, true},
		{"neg coin, non-atom denom", fields{delegatee, coinNegNotAtoms}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := BondUpdate{
				Delegatee: tt.fields.Delegatee,
				Amount:    tt.fields.Amount,
			}
			assert.Equal(t, tt.wantErr, tx.ValidateBasic() != nil, tt.name)
		})
	}
}

func TestTxNominateValidateBasic(t *testing.T) {
	type fields struct {
		Nominee    sdk.Actor
		Amount     coin.Coin
		Commission Decimal
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"basic good", fields{delegatee, coinPos, comm50}, false},
		{"empty delegator", fields{empty, coinPos, comm50}, true},
		{"zero coin", fields{delegatee, coinZero, comm50}, true},
		{"neg coin", fields{delegatee, coinNeg, comm50}, true},
		{"pos coin, non-atom denom", fields{delegatee, coinPosNotAtoms, comm50}, true},
		{"zero coin, non-atom denom", fields{delegatee, coinZeroNotAtoms, comm50}, true},
		{"neg coin, non-atom denom", fields{delegatee, coinNegNotAtoms, comm50}, true},
		{"negative commission", fields{delegatee, coinPos, commNeg50}, true},
		{"zero commission", fields{delegatee, coinPos, comm0}, false},
		{"100% commission", fields{delegatee, coinPos, comm100}, false},
		{"110% commission", fields{delegatee, coinPos, comm110}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := TxNominate{
				Nominee:    tt.fields.Nominee,
				Amount:     tt.fields.Amount,
				Commission: tt.fields.Commission,
			}
			assert.Equal(t, tt.wantErr, tx.ValidateBasic() != nil, tt.name)
		})
	}
}

func TestTxModCommValidateBasic(t *testing.T) {
	type fields struct {
		Delegatee  sdk.Actor
		Commission Decimal
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"basic good", fields{delegatee, comm50}, false},
		{"empty delegator", fields{empty, comm50}, true},
		{"negative commission", fields{delegatee, commNeg50}, true},
		{"zero commission", fields{delegatee, comm0}, false},
		{"100% commission", fields{delegatee, comm100}, false},
		{"110% commission", fields{delegatee, comm110}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := TxModComm{
				Delegatee:  tt.fields.Delegatee,
				Commission: tt.fields.Commission,
			}
			assert.Equal(t, tt.wantErr, tx.ValidateBasic() != nil, tt.name)
		})
	}
}
