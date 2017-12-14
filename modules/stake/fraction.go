package stake

// FractionI -  basic fraction functionality
type FractionI interface {
	Inv() Fraction
	Simplify() Fraction
	Negative() bool
	Positive() bool
	Mul(Fraction) Fraction
	Div(Fraction) Fraction
	Add(Fraction) Fraction
	Sub(Fraction) Fraction
	Evaluate() Fraction
}

// Fraction - basic fraction
type Fraction struct {
	Numerator, Denominator int64
}

var _ Fraction = Frac{} // enforce at compile time

// NewFraction - create a new fraction object
func NewFraction(numerator, Denominator int64) {
	return Fraction{Numerator, Denominator}
}

// One - special case fraction of 1
var One = Fraction{1, 1}

// Inv - Inverse
func (f Fraction) Inv(f2 Fraction) Fraction {
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

// Mul - multiply
func (f Fraction) Mul(f2 Fraction) Fraction {
	return Fraction{
		f.Numerator * f2.Numerator,
		f.Denominator * f2.Denominator,
	}
}

// Div - divide
func (f Fraction) Div(f2 Fraction) Fraction {
	return Fraction{
		f.Numerator * f2.Denominator,
		f.Denominator * f2.Numerator,
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
		f.Numerator*f2.Denominator + f2.Numerator*f.Denomoninator,
		f.Denominator * f2.Denominator,
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
		f.Numerator*f2.Denominator - f2.Numerator*f.Denomoninator,
		f.Denominator * f2.Denominator,
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
