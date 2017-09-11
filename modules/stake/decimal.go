//nolint
package stake

//TODO move to tendermint/tmlibs?

import (
	"github.com/rigelrozanski/decimal"
)

const places int32 = 18 //number of decimal places

// NOTE we need to use custom Value and Exp instead of
// just using decimal.Decimal here to maintain exposed
// variables for go-wire serialization
type Decimal struct {
	Value int64 // Number = Value * 10 ^ Exp
	Exp   int32
}

var (
	Zero = NewDecimal(0, 1)
	One  = NewDecimal(1, 1)
)

func NewDecimal(value int64, exp int32) Decimal {
	return Decimal{value, exp}
}

func NewDecimalFromString(value string) (Decimal, error) {
	d, err := decimal.NewFromString(value)
	out := get(d)
	return out, err
}

//coversion to/from the shopspring decimal format
func get(d decimal.Decimal) Decimal {
	d = d.RoundBank(places)
	return Decimal{d.IntPart(), d.Exponent()}
}
func set(d Decimal) decimal.Decimal {
	return decimal.New(d.Value, d.Exp)
}

func (d Decimal) String() string { return set(d).StringFixedBank() }
func (d Decimal) IntPart() int64 { return set(d).IntPart() }

func (d Decimal) Plus(d2 Decimal) Decimal  { return get(set(d).Add(set(d2))) }
func (d Decimal) Minus(d2 Decimal) Decimal { return get(set(d).Sub(set(d2))) }
func (d Decimal) Negative() Decimal        { return get(set(d).Neg()) }
func (d Decimal) Mul(d2 Decimal) Decimal   { return get(set(d).Mul(set(d2))) }
func (d Decimal) Div(d2 Decimal) Decimal   { return get(set(d).Div(set(d2))) }

func (d Decimal) Equal(d2 Decimal) bool { return set(d).Equal(set(d2)) }
func (d Decimal) GT(d2 Decimal) bool    { return set(d).GreaterThan(set(d2)) }
func (d Decimal) GTE(d2 Decimal) bool   { return set(d).GreaterThanOrEqual(set(d2)) }
func (d Decimal) LT(d2 Decimal) bool    { return set(d).LessThan(set(d2)) }
func (d Decimal) LTE(d2 Decimal) bool   { return set(d).LessThanOrEqual(set(d2)) }
