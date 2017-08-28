//nolint
package stake

//TODO move to tendermint/tmlibs?

import (
	"github.com/shopspring/decimal"
)

type Decimal struct {
	decimal.Decimal
}

var (
	Zero = NewDecimal(0, 1)
	One  = NewDecimal(1, 1)
)

func NewDecimal(value int64, exp int32) Decimal {
	return Decimal{decimal.New(value, exp)}
}

func NewDecimalFromString(value string) (Decimal, error) {
	d, err := decimal.NewFromString(value)
	out := Decimal{d}
	return out, err
}

func (d Decimal) String() string { return d.Decimal.String() }
func (d Decimal) IntPart() int64 { return d.Decimal.IntPart() }

func (d Decimal) Plus(d2 Decimal) Decimal  { return Decimal{d.Decimal.Add(d2.Decimal)} }
func (d Decimal) Minus(d2 Decimal) Decimal { return Decimal{d.Decimal.Sub(d2.Decimal)} }
func (d Decimal) Negative() Decimal        { return Decimal{d.Decimal.Neg()} }
func (d Decimal) Mul(d2 Decimal) Decimal   { return Decimal{d.Decimal.Mul(d2.Decimal)} }
func (d Decimal) Div(d2 Decimal) Decimal   { return Decimal{d.Decimal.Div(d2.Decimal)} }

func (d Decimal) Equal(d2 Decimal) bool { return d.Decimal.Equal(d2.Decimal) }
func (d Decimal) GT(d2 Decimal) bool    { return d.Decimal.GreaterThan(d2.Decimal) }
func (d Decimal) GTE(d2 Decimal) bool   { return d.Decimal.GreaterThanOrEqual(d2.Decimal) }
func (d Decimal) LT(d2 Decimal) bool    { return d.Decimal.LessThan(d2.Decimal) }
func (d Decimal) LTE(d2 Decimal) bool   { return d.Decimal.LessThanOrEqual(d2.Decimal) }
