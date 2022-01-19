package money

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"github.com/domonda/go-types/float"
)

// Rate is a float64 underneath with additional methods
// useful for money conversion rates and percentages.
type Rate float64

// ParseRate parses a rate from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted.
// If a string ends with '%' then the parsed number part will be divided by 100.
func ParseRate(str string, acceptedDecimals ...int) (Rate, error) {
	var (
		s       = strings.TrimSpace(str)
		percent = false
	)
	if l := len(s); l > 0 && s[l-1] == '%' {
		percent = true
		s = strings.TrimRightFunc(s[:l-1], unicode.IsSpace)
	}
	f, _, _, decimals, err := float.ParseDetails(s)
	if err != nil {
		if s != str {
			err = fmt.Errorf("ParseRate(%q): %w", str, err)
		}
		return 0, err
	}
	if percent {
		f /= 100
	}
	if len(acceptedDecimals) == 0 || math.IsNaN(f) || math.IsInf(f, 0) {
		return Rate(f), nil
	}
	for _, accepted := range acceptedDecimals {
		if decimals == accepted {
			return Rate(f), nil
		}
	}
	return 0, fmt.Errorf("parsing %q returned %d decimals wich is not in accepted list of %v", s, decimals, acceptedDecimals)
}

// NewRate returns a pointer to a Rate
// with the passed value.
func NewRate(value float64) *Rate {
	r := new(Rate)
	*r = Rate(value)
	return r
}

// RateFromPtr dereferences ptr or returns defaultVal if it is nil
func RateFromPtr(ptr *Rate, defaultVal Rate) Rate {
	if ptr == nil {
		return defaultVal
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
	f, err := float.Parse(source)
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
func (r Rate) Format(thousandsSep, decimalSep rune, precision int) string {
	return float.Format(float64(r), thousandsSep, decimalSep, precision, true)
}

// StringOr returns strconv.FormatFloat the pointed to rate or defaultVal if ptr is nil.
func (ptr *Rate) StringOr(defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return strconv.FormatFloat(float64(*ptr), 'f', -1, 64)
}

// FloatOr returns the pointed to rate as float64 or defaultVal if ptr is nil.
func (ptr *Rate) FloatOr(defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return float64(*ptr)
}

// RateOr returns the pointed to rate or defaultVal if ptr is nil.
func (ptr *Rate) RateOr(defaultVal Rate) Rate {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
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

// AbsFloat returns the absolute value of r as float64.
//
// Special cases are:
//	AbsFloat(±Inf) = +Inf
//	AbsFloat(NaN) = NaN
func (r Rate) AbsFloat() float64 {
	return math.Abs(float64(r))
}

// Invert inverts the sign of the rate.
func (r *Rate) Invert() {
	*r = -*r
}

// Inverted returns the rate with inverted sign.
func (r Rate) Inverted() Rate {
	return -r
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

// UnmarshalJSON implements encoding/json.Unmarshaler
// and accepts numbers, strings, and null.
// JSON null or "" will set the rate to zero.
// If a string ends with '%' then the parsed number part will be divided by 100.
func (r *Rate) UnmarshalJSON(j []byte) error {
	s := string(j)

	if s == `null` || s == `""` {
		*r = 0
		return nil
	}

	// Strip quotes
	if l := len(s); l > 2 && s[0] == '"' && s[l-1] == '"' {
		s = s[1 : l-1]
	}

	rate, err := ParseRate(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as money.Rate because of: %w", j, err)
	}

	*r = rate
	return nil
}
