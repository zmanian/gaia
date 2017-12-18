package stake

// XXX test fractions!

// FractionI -  basic fraction functionality
type FractionI interface {
	Inv() Fraction
	Simplify() Fraction
	Negative() bool
	Positive() bool
	GT(Fraction) bool
	LT(Fraction) bool
	Equal(Fraction) bool
	Mul(Fraction) Fraction
	Div(Fraction) Fraction
	Add(Fraction) Fraction
	Sub(Fraction) Fraction
	Evaluate() int64
	MulInt(int64) Fraction
	DivInt(int64) Fraction
	AddInt(int64) Fraction
	SubInt(int64) Fraction
}

// Fraction - basic fraction
type Fraction struct {
	Numerator, Denominator int64
}

var _ FractionI = Fraction{} // enforce at compile time

// NewFraction - create a new fraction object
func NewFraction(numerator, denominator int64) Fraction {
	return Fraction{numerator, denominator}
}

// nolint sprecial predefined fractions
var One = Fraction{1, 1}
var Zero = Fraction{0, 1}

// Inv - Inverse
func (f Fraction) Inv() Fraction {
	return Fraction{f.Denominator, f.Numerator}
}

// Simplify - find the greatest common Denominator, divide
func (f Fraction) Simplify() Fraction {

	gcd := f.Numerator

	for d := f.Denominator; d != 0; {
		gcd, d = d, gcd%d
	}

	return Fraction{f.Numerator / gcd, f.Denominator / gcd}
}

// Negative - is negative TODO make more efficient?
func (f Fraction) Negative() bool {
	return (f.Numerator / f.Denominator) < 0
}

// Positive - is negative TODO make more efficient?
func (f Fraction) Positive() bool {
	return (f.Numerator / f.Denominator) > 0
}

// GT - greater than
func (f Fraction) GT(f2 Fraction) bool {
	return f.Sub(f2).Positive()
}

// LT - less than
func (f Fraction) LT(f2 Fraction) bool {
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
func (f Fraction) Equal(f2 Fraction) bool {
	if f.Numerator == 0 {
		return f2.Numerator == 0
	}
	return ((f.Numerator == f2.Numerator) && (f.Denominator == f2.Denominator))
}

// Mul - multiply
func (f Fraction) Mul(f2 Fraction) Fraction {
	return Fraction{
		f.Numerator * f2.Numerator,
		f.Denominator * f2.Denominator,
	}
}

// MulInt - multiply fraction by integer
func (f Fraction) MulInt(i int64) Fraction {
	return Fraction{
		f.Numerator * i,
		f.Denominator,
	}
}

// Div - divide
func (f Fraction) Div(f2 Fraction) Fraction {
	return Fraction{
		f.Numerator * f2.Denominator,
		f.Denominator * f2.Numerator,
	}
}

// DivInt - divide fraction by and integer
func (f Fraction) DivInt(i int64) Fraction {
	return Fraction{
		f.Numerator,
		f.Denominator * i,
	}
}

// Add - add without simplication
func (f Fraction) Add(f2 Fraction) Fraction {
	if f.Denominator == f2.Denominator {
		return Fraction{
			f.Numerator + f2.Numerator,
			f.Denominator,
		}
	}
	return Fraction{
		f.Numerator*f2.Denominator + f2.Numerator*f.Denominator,
		f.Denominator * f2.Denominator,
	}
}

// AddInt - add fraction with integer, no simplication
func (f Fraction) AddInt(i int64) Fraction {
	return Fraction{
		f.Numerator + i*f.Denominator,
		f.Denominator,
	}
}

// Sub - subtract without simplication
func (f Fraction) Sub(f2 Fraction) Fraction {
	if f.Denominator == f2.Denominator {
		return Fraction{
			f.Numerator - f2.Numerator,
			f.Denominator,
		}
	}
	return Fraction{
		f.Numerator*f2.Denominator - f2.Numerator*f.Denominator,
		f.Denominator * f2.Denominator,
	}
}

// SubInt - subtract fraction with integer, no simplication
func (f Fraction) SubInt(i int64) Fraction {
	return Fraction{
		f.Numerator - i*f.Denominator,
		f.Denominator,
	}
}

// Evaluate - evaluate the fraction using bankers rounding
func (f Fraction) Evaluate() int64 {

	d := f.Numerator / f.Denominator // always drops the decimal
	if f.Numerator%f.Denominator == 0 {
		return d
	}

	// evaluate the remainder
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
