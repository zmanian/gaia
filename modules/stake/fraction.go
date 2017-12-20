package stake

// XXX test fractions!

// FractionI -  basic fraction functionality
// TODO better name that FractionI?
type FractionI interface {
	Inv() FractionI
	SetNumerator(int64) FractionI
	SetDenominator(int64) FractionI
	Numerator() int64
	Denominator() int64
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
	numerator, denominator int64
}

var _ FractionI = Fraction{} // enforce at compile time

// NewFraction - create a new fraction object
func NewFraction(numerator int64, denominator ...int64) Fraction {
	var denom int64 = 1
	if len(denominator) > 0 {
		denom = denominator[0]
	}
	return Fraction{numerator, denom}
}

// SetNumerator - return a fraction with a new numerator
func (f Fraction) SetNumerator(numerator int64) FractionI {
	return Fraction{numerator, f.denominator}
}

// SetDenominator - return a fraction with a new denominator
func (f Fraction) SetDenominator(denominator int64) FractionI {
	return Fraction{f.numerator, denominator}
}

// Numerator - return the numerator
func (f Fraction) Numerator() int64 {
	return f.numerator
}

// Denominator - return the denominator
func (f Fraction) Denominator() int64 {
	return f.denominator
}

// TODO define faster operations (mul, add, etc) on One and Zero
// nolint special predefined fractions
var One = Fraction{1, 1}
var Zero = Fraction{0, 1}

// Inv - Inverse
func (f Fraction) Inv() FractionI {
	return Fraction{f.denominator, f.numerator}
}

// Simplify - find the greatest common denominator, divide
func (f Fraction) Simplify() FractionI {

	gcd := f.numerator

	for d := f.denominator; d != 0; {
		gcd, d = d, gcd%d
	}

	return Fraction{f.numerator / gcd, f.denominator / gcd}
}

// Negative - is the fractior negative
func (f Fraction) Negative() bool {
	switch {
	case f.numerator > 0:
		if f.denominator > 0 {
			return false
		}
		return true
	case f.numerator < 0:
		if f.denominator < 0 {
			return false
		}
		return true
	}
	return false
}

// Positive - is the fraction positive
func (f Fraction) Positive() bool {
	switch {
	case f.numerator > 0:
		if f.denominator > 0 {
			return true
		}
		return false
	case f.numerator < 0:
		if f.denominator < 0 {
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
	if f.numerator == 0 {
		return f2.Numerator() == 0
	}
	return ((f.numerator == f2.Numerator()) && (f.denominator == f2.Denominator()))
}

// Mul - multiply
func (f Fraction) Mul(f2 FractionI) FractionI {
	return Fraction{
		f.numerator * f2.Numerator(),
		f.denominator * f2.Denominator(),
	}
}

// MulInt - multiply fraction by integer
func (f Fraction) MulInt(i int64) FractionI {
	return Fraction{
		f.numerator * i,
		f.denominator,
	}
}

// Div - divide
func (f Fraction) Div(f2 FractionI) FractionI {
	return Fraction{
		f.numerator * f2.Denominator(),
		f.denominator * f2.Numerator(),
	}
}

// DivInt - divide fraction by and integer
func (f Fraction) DivInt(i int64) FractionI {
	return Fraction{
		f.numerator,
		f.denominator * i,
	}
}

// Add - add without simplication
func (f Fraction) Add(f2 FractionI) FractionI {
	if f.denominator == f2.Denominator() {
		return Fraction{
			f.numerator + f2.Numerator(),
			f.denominator,
		}
	}
	return Fraction{
		f.numerator*f2.Denominator() + f2.Numerator()*f.denominator,
		f.denominator * f2.Denominator(),
	}
}

// AddInt - add fraction with integer, no simplication
func (f Fraction) AddInt(i int64) FractionI {
	return Fraction{
		f.numerator + i*f.denominator,
		f.denominator,
	}
}

// Sub - subtract without simplication
func (f Fraction) Sub(f2 FractionI) FractionI {
	if f.denominator == f2.Denominator() {
		return Fraction{
			f.numerator - f2.Numerator(),
			f.denominator,
		}
	}
	return Fraction{
		f.numerator*f2.Denominator() - f2.Numerator()*f.denominator,
		f.denominator * f2.Denominator(),
	}
}

// SubInt - subtract fraction with integer, no simplication
func (f Fraction) SubInt(i int64) FractionI {
	return Fraction{
		f.numerator - i*f.denominator,
		f.denominator,
	}
}

// Evaluate - evaluate the fraction using bankers rounding
func (f Fraction) Evaluate() int64 {

	d := f.numerator / f.denominator // always drops the decimal
	if f.numerator%f.denominator == 0 {
		return d
	}

	// evaluate the remainder using bankers rounding
	remainderDigit := (f.numerator * 10 / f.denominator) - (d * 10) // get the first remainder digit
	isFinalDigit := (f.numerator*10%f.denominator == 0)             // is this the final digit in the remainder?
	if isFinalDigit && remainderDigit == 5 {
		return d + (d % 2) // always rounds to the even number
	}
	if remainderDigit >= 5 {
		d++
	}
	return d
}
