package stake

import wire "github.com/tendermint/go-wire"

// XXX test fractions!

// FractionI -  basic fraction functionality
// TODO better name that FractionI?
type FractionI interface {
	Inv() FractionI
	SetNumerator(int64) FractionI
	SetDenominator(int64) FractionI
	GetNumerator() int64
	GetDenominator() int64
	Simplify() FractionI
	Negative() bool
	Positive() bool
	GT(FractionI) bool
	LT(FractionI) bool
	GTint(int64) bool
	LTint(int64) bool
	Equal(FractionI) bool
	Mul(FractionI) FractionI
	Div(FractionI) FractionI
	Add(FractionI) FractionI
	Sub(FractionI) FractionI
	MulInt(int64) FractionI
	DivInt(int64) FractionI
	AddInt(int64) FractionI
	SubInt(int64) FractionI
	Evaluate() int64
}

// Fraction - basic fraction
type Fraction struct {
	Numerator, Denominator int64
}

var _ FractionI = Fraction{} // enforce at compile time
var _ = wire.RegisterInterface(struct{ FractionI }{}, wire.ConcreteType{Fraction{}, 0x01})

// NewFraction - create a new fraction object
func NewFraction(Numerator int64, Denominator ...int64) Fraction {
	var denom int64 = 1
	if len(Denominator) > 0 {
		denom = Denominator[0]
	}
	return Fraction{Numerator, denom}
}

// SetNumerator - return a fraction with a new Numerator
func (f Fraction) SetNumerator(Numerator int64) FractionI {
	return Fraction{Numerator, f.Denominator}
}

// SetDenominator - return a fraction with a new Denominator
func (f Fraction) SetDenominator(Denominator int64) FractionI {
	return Fraction{f.Numerator, Denominator}
}

// GetNumerator - return the Numerator
func (f Fraction) GetNumerator() int64 {
	return f.Numerator
}

// GetDenominator - return the Denominator
func (f Fraction) GetDenominator() int64 {
	return f.Denominator
}

// TODO define faster operations (mul, add, etc) on One and Zero
// nolint special predefined fractions
var One = Fraction{1, 1}
var Zero = Fraction{0, 1}

// Inv - Inverse
func (f Fraction) Inv() FractionI {
	return Fraction{f.Denominator, f.Numerator}
}

// Simplify - find the greatest common Denominator, divide
func (f Fraction) Simplify() FractionI {

	gcd := f.Numerator

	for d := f.Denominator; d != 0; {
		gcd, d = d, gcd%d
	}

	return Fraction{f.Numerator / gcd, f.Denominator / gcd}
}

// Negative - is the fractior negative
func (f Fraction) Negative() bool {
	switch {
	case f.Numerator > 0:
		if f.Denominator > 0 {
			return false
		}
		return true
	case f.Numerator < 0:
		if f.Denominator < 0 {
			return false
		}
		return true
	}
	return false
}

// Positive - is the fraction positive
func (f Fraction) Positive() bool {
	switch {
	case f.Numerator > 0:
		if f.Denominator > 0 {
			return true
		}
		return false
	case f.Numerator < 0:
		if f.Denominator < 0 {
			return true
		}
		return false
	}
	return false
}

// GT - greater than
func (f Fraction) GT(f2 FractionI) bool {
	return f.Sub(f2).Positive()
}

// LT - less than
func (f Fraction) LT(f2 FractionI) bool {
	return f.Sub(f2).Negative()
}

// GTint - greater than integer
func (f Fraction) GTint(i int64) bool {
	return f.SubInt(i).Positive()
}

// LTint - less than integer
func (f Fraction) LTint(i int64) bool {
	return f.SubInt(i).Negative()
}

// Equal - test if two Fractions are equal, does not simplify
func (f Fraction) Equal(f2 FractionI) bool {
	if f.Numerator == 0 {
		return f2.GetNumerator() == 0
	}
	return ((f.Numerator == f2.GetNumerator()) && (f.Denominator == f2.GetDenominator()))
}

// Mul - multiply
func (f Fraction) Mul(f2 FractionI) FractionI {
	return Fraction{
		f.Numerator * f2.GetNumerator(),
		f.Denominator * f2.GetDenominator(),
	}.Simplify()
}

// MulInt - multiply fraction by integer
func (f Fraction) MulInt(i int64) FractionI {
	return Fraction{
		f.Numerator * i,
		f.Denominator,
	}.Simplify()
}

// Div - divide
func (f Fraction) Div(f2 FractionI) FractionI {
	return Fraction{
		f.Numerator * f2.GetDenominator(),
		f.Denominator * f2.GetNumerator(),
	}.Simplify()
}

// DivInt - divide fraction by and integer
func (f Fraction) DivInt(i int64) FractionI {
	return Fraction{
		f.Numerator,
		f.Denominator * i,
	}.Simplify()
}

// Add - add without simplication
func (f Fraction) Add(f2 FractionI) FractionI {
	if f.Denominator == f2.GetDenominator() {
		return Fraction{
			f.Numerator + f2.GetNumerator(),
			f.Denominator,
		}.Simplify()
	}
	return Fraction{
		f.Numerator*f2.GetDenominator() + f2.GetNumerator()*f.Denominator,
		f.Denominator * f2.GetDenominator(),
	}.Simplify()
}

// AddInt - add fraction with integer, no simplication
func (f Fraction) AddInt(i int64) FractionI {
	return Fraction{
		f.Numerator + i*f.Denominator,
		f.Denominator,
	}.Simplify()
}

// Sub - subtract without simplication
func (f Fraction) Sub(f2 FractionI) FractionI {
	if f.Denominator == f2.GetDenominator() {
		return Fraction{
			f.Numerator - f2.GetNumerator(),
			f.Denominator,
		}.Simplify()
	}
	return Fraction{
		f.Numerator*f2.GetDenominator() - f2.GetNumerator()*f.Denominator,
		f.Denominator * f2.GetDenominator(),
	}.Simplify()
}

// SubInt - subtract fraction with integer, no simplication
func (f Fraction) SubInt(i int64) FractionI {
	return Fraction{
		f.Numerator - i*f.Denominator,
		f.Denominator,
	}.Simplify()
}

// Evaluate - evaluate the fraction using bankers rounding
func (f Fraction) Evaluate() int64 {

	d := f.Numerator / f.Denominator // always drops the decimal
	if f.Numerator%f.Denominator == 0 {
		return d
	}

	// evaluate the remainder using bankers rounding
	remainderDigit := (f.Numerator * 10 / f.Denominator) - (d * 10) // get the first remainder digit
	isFinalDigit := (f.Numerator*10%f.Denominator == 0)             // is this the final digit in the remainder?
	if isFinalDigit && remainderDigit == 5 {
		return d + (d % 2) // always rounds to the even number
	}
	if remainderDigit >= 5 {
		d++
	}
	return d
}
