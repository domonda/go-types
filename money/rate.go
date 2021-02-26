package money

import (
	"math"
	"math/big"

	"github.com/domonda/go-types/strfmt"
)

// Rate is a float64 underneath with additional methods
// useful for money conversion rates and percentages.
type Rate float64

// RateFromPtr dereferences ptr or returns nilVal if it is nil
func RateFromPtr(ptr *Rate, nilVal Rate) Rate {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for sourceWasNormalized and nil for err.
// ScanString implements the strfmt.Scannable interface.
func (r *Rate) ScanString(source string) (sourceWasNormalized bool, err error) {
	f, err := strfmt.ParseFloat(source)
	if err != nil {
		return false, err
	}
	*r = Rate(f)
	return true, nil
}

// RoundToInt returns the value rounded to an integer number
func (r Rate) RoundToInt() Rate {
	return Rate(math.Round(float64(r)))
}

// RoundToDecimals returns the value rounded
// to the passed number of decimal places.
func (r Rate) RoundToDecimals(decimals int) Rate {
	pow := math.Pow10(decimals)
	return Rate(math.Round(float64(r)*pow) / pow)
}

// Format formats the Rate similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
// Valid values for thousandsSep are [0, ',', '.']
// and thousandsSep must be different from decimalSep.
// The precision argument controls the number of digits (excluding the exponent).
// Note that the last digit is not rounded!
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
func (r Rate) Format(thousandsSep, decimalSep byte, precision int) string {
	return strfmt.FormatFloat(float64(r), thousandsSep, decimalSep, precision, true)
}

// BigFloat returns m as a new big.Float
func (r Rate) BigFloat() *big.Float {
	return big.NewFloat(float64(r))
}

func (r *Rate) Equal(other *Rate) bool {
	if r == other {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	return *r == *other
}

// Signbit reports whether a is negative or negative zero.
func (r Rate) Signbit() bool {
	return math.Signbit(float64(r))
}

// Copysign returns an Rate with the magnitude
// of a and with the sign of the sign argument.
func (r Rate) Copysign(sign Rate) Rate {
	return Rate(math.Copysign(float64(r), float64(sign)))
}

// Abs returns the absolute value of r.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (r Rate) Abs() Rate {
	return Rate(math.Abs(float64(r)))
}

// WithPosSign returns the value with a positive sign (abs) if true is passed,
// or with a negative sign if false is passed.
func (r Rate) WithPosSign(positive bool) Rate {
	if positive {
		return r.Copysign(+1)
	} else {
		return r.Copysign(-1)
	}
}

// WithNegSign returns the value with a negative sign if true is passed,
// or with a positive sign (abs) if false is passed.
func (r Rate) WithNegSign(negative bool) Rate {
	if negative {
		return r.Copysign(-1)
	} else {
		return r.Copysign(+1)
	}
}

// Inverse returns the inverse rate (1 / r)
func (r Rate) Inverse() Rate {
	return 1 / r
}

// Valid returns if a is not infinite or NaN
func (r Rate) Valid() bool {
	return !math.IsNaN(float64(r)) && !math.IsInf(float64(r), 0)
}

func (r Rate) ValidAndGreaterZero() bool {
	return r.Valid() && r > 0
}

func (r Rate) ValidAndSmallerZero() bool {
	return r.Valid() && r < 0
}

// ValidAndHasSign returns if r.Valid() and
// if it has the same sign than the passed int argument or any sign if 0 is passed.
func (r Rate) ValidAndHasSign(sign int) bool {
	if !r.Valid() {
		return false
	}
	switch {
	case sign > 0:
		return r > 0
	case sign < 0:
		return r < 0
	}
	return true
}
