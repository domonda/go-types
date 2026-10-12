package money

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/domonda/go-types/float"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// Rate is a float64 underneath with additional methods
// useful for money conversion rates and percentages.
type Rate float64

// NullableRate is a nullable wrapper around Rate,
// where the zero value of the underlying nullable.Type represents NULL.
type NullableRate = nullable.Type[Rate]

// NullableRateFrom returns a not null nullable rate from a value
// using the value as the non-null value.
func NullableRateFrom(value float64) NullableRate {
	return NullableRate(nullable.TypeFrom(Rate(value)))
}

// NullableRateFromPtr returns a nullable rate from a pointer
// using nil as the null value.
func NullableRateFromPtr(ptr *Rate) NullableRate {
	return NullableRate(nullable.TypeFromPtr(ptr))
}

// ParseRate parses a rate from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted
// and infinity and NaN are parsed and returned without error.
// If acceptedDecimals are passed, then infinity and NaN are errors
// because they have no decimals that could be accepted.
// If a string ends with '%' then the parsed number part will be divided by 100.
func ParseRate(str string, acceptedDecimals ...int) (Rate, error) {
	var (
		s       = strutil.TrimSpace(str)
		percent = false
	)
	if l := len(s); l > 0 && s[l-1] == '%' {
		percent = true
		s = strings.TrimRightFunc(s[:l-1], strutil.IsSpace)
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
	if len(acceptedDecimals) == 0 {
		return Rate(f), nil
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, errNonFiniteNotAccepted(str, acceptedDecimals)
	}
	if slices.Contains(acceptedDecimals, decimals) {
		return Rate(f), nil
	}
	return 0, fmt.Errorf("parsing %q returned %d decimals which is not in accepted list of %v", s, decimals, acceptedDecimals)
}

// RateFromPtr dereferences ptr or returns defaultVal if it is nil.
// See also Rate.Ptr.
func RateFromPtr(ptr *Rate, defaultVal Rate) Rate {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Ptr returns a new pointer to a copy of the rate value.
// See also RateFromPtr.
func (r Rate) Ptr() *Rate {
	return &r
}

// ScanString tries to parse and assign the passed
// source string as value of the implementing type.
//
// If validate is true, the source string is checked
// for validity before it is assigned to the type.
//
// If validate is false and the source string
// can still be assigned in some non-normalized way
// it will be assigned without returning an error.
//
// The source string is parsed with ParseRate,
// so a trailing '%' divides the number by 100.
func (r *Rate) ScanString(source string, validate bool) error {
	rate, err := ParseRate(source)
	if err != nil {
		return err
	}
	if validate && !rate.Valid() {
		return fmt.Errorf("invalid rate: %q", source)
	}
	*r = rate
	return nil
}

// GoString returns the rate as string
// in full float64 precision for debugging
func (r Rate) GoString() string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.200f", float64(r)), "0"), ".")
}

// RoundToInt returns the value rounded to an integer number
func (r Rate) RoundToInt() Rate {
	return Rate(math.Round(float64(r)))
}

// RoundToDecimals returns the value rounded half away from zero
// to the passed number of decimal places.
// Negative decimals round to tens, hundreds, and so on.
// Like Amount.RoundToDecimals it rounds the shortest decimal
// representation of the float64 value, so Rate(0.145).RoundToDecimals(2)
// is 0.15 and not the 0.14 of its binary value.
// NaN and infinite rates are returned unchanged.
// See float.RoundToDecimals.
func (r Rate) RoundToDecimals(decimals int) Rate {
	return float.RoundToDecimals(r, decimals)
}

// FormatSep formats the Rate similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
// Valid values for thousandsSep are 0 (no grouping), ',', '.', ' '
// and the apostrophe,
// and thousandsSep must be different from decimalSep.
// The precision argument is the number of decimal places, padded with zeros.
// The rate is rounded to them half away from zero like Rate.RoundToDecimals
// does, so Rate(0.145) formats as "0.15" with precision 2, see float.Format.
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return the rate exactly.
func (r Rate) FormatSep(thousandsSep, decimalSep rune, precision int) string {
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

// Equal returns true if two Rate pointers point to rates with equal values
// or are the same pointer. Returns false if either pointer is nil while the other is not.
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
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (r Rate) Abs() Rate {
	return Rate(math.Abs(float64(r)))
}

// AbsFloat returns the absolute value of r as float64.
//
// Special cases are:
//
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

// ValidAndPositive returns true if the rate is neither infinite nor NaN
// and greater than or equal to zero.
func (r Rate) ValidAndPositive() bool {
	return r.Valid() && r >= 0
}

// ValidAndGreaterZero returns true if the rate is neither infinite nor NaN
// and strictly greater than zero.
func (r Rate) ValidAndGreaterZero() bool {
	return r.Valid() && r > 0
}

// ValidAndSmallerZero returns true if the rate is neither infinite nor NaN
// and strictly less than zero.
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

	// Decode a JSON string through the standard library so \uXXXX and other
	// escape sequences are resolved, not just the surrounding quotes stripped.
	if len(s) > 0 && s[0] == '"' {
		if err := json.Unmarshal(j, &s); err != nil {
			return fmt.Errorf("can't unmarshal JSON(%s) as money.Rate because of: %w", j, err)
		}
	}

	rate, err := ParseRate(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as money.Rate because of: %w", j, err)
	}

	*r = rate
	return nil
}
