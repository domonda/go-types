package money

import (
	"bytes"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/domonda/errors"
	"github.com/domonda/go-types/strfmt"
)

// Amount adds money related methods to float64
type Amount float64

// ParseAmount parses an amount from str accepting only certain decimal digit counts.
// If no acceptedDecimals are passed, then any decimal digit count is accepted.
func ParseAmount(str string, acceptedDecimals ...int) (Amount, error) {
	f, _, _, decimals, err := strfmt.ParseFloatDetails(str)
	if err != nil {
		return 0, err
	}
	if len(acceptedDecimals) == 0 {
		return Amount(f), nil
	}
	for _, accepted := range acceptedDecimals {
		if decimals == accepted {
			return Amount(f), nil
		}
	}
	return 0, errors.Errorf("parsing %q returned %d decimals wich is not in accepted list of %v", str, decimals, acceptedDecimals)
}

// AmountFromPtr dereferences ptr or returns nilVal if it is nil
func AmountFromPtr(ptr *Amount, nilVal Amount) Amount {
	if ptr == nil {
		return nilVal
	}
	return *ptr
}

// AssignString tries to parse and assign the passed
// source string as value of the implementing object.
// It returns an error if source could not be parsed.
// If the source string could be parsed, but was not
// in the expected normalized format, then false is
// returned for normalized and nil for err.
// AssignString implements strfmt.StringAssignable
func (a *Amount) AssignString(source string) (normalized bool, err error) {
	f, err := strfmt.ParseFloat(source)
	if err != nil {
		return false, err
	}
	*a = Amount(f)
	return true, nil
}

// Cents returns the amount rounded to cents
func (a Amount) Cents() int64 {
	return int64(math.Round(float64(a) * 100))
}

// RoundToInt returns the amount rounded to an integer number
func (a Amount) RoundToInt() Amount {
	return Amount(math.Round(float64(a)))
}

// RoundToCents returns the amount rounded to cents
func (a Amount) RoundToCents() Amount {
	s := strconv.FormatFloat(float64(a), 'f', -1, 64)

	p := strings.IndexByte(s, '.')
	if p == -1 {
		return a
	}
	if decim := len(s) - p - 1; decim <= 2 {
		return a
	}

	if s[p+3] < '5' {
		// If third decimal is smaller than 5 cut off rest
		s = s[:p+3]
	} else {
		// If third decimal is equal or larger than 5,
		// increase second decimal by one
		s = s[:p+2] + string(s[p+2]+1)
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return Amount(f)
}

// String returns the amount formatted to two decimal places
// String implements the fmt.Stringer interface.
func (a Amount) String() string {
	return a.Format(0, '.', 2)

	// neg := a < 0
	// s := strconv.FormatInt(a.Abs().Cents(), 10)

	// l := len(s) + 1
	// if l < 4 {
	// 	l = 4
	// }
	// if neg {
	// 	l++
	// }
	// var b strings.Builder
	// b.Grow(l)

	// if neg {
	// 	b.WriteByte('-')
	// }
	// switch len(s) {
	// case 1:
	// 	b.WriteString("0.0")
	// 	b.WriteString(s)
	// case 2:
	// 	b.WriteString("0.")
	// 	b.WriteString(s)
	// default:
	// 	b.WriteString(s[:len(s)-2])
	// 	b.WriteByte('.')
	// 	b.WriteString(s[len(s)-2:])
	// }

	// return b.String()
}

// StringOr returns ptr.String() or nilVal if ptr is nil.
func (ptr *Amount) StringOr(nilVal string) string {
	if ptr == nil {
		return nilVal
	}
	return ptr.String()
}

// Format formats the Amount similar to strconv.FormatFloat with the 'f' format option,
// but with decimalSep as decimal separator instead of a point
// and optional grouping of the integer part.
// Valid values for decimalSep are '.' and ','.
// If thousandsSep is not zero, then the integer part of the number is grouped
// with thousandsSep between every group of 3 digits.
// Valid values for thousandsSep are [0, ',', '.'] and thousandsSep must be different from  decimalSep.
// precision controls the number of digits (excluding the exponent).
// The special precision -1 uses the smallest number of digits
// necessary such that ParseFloat will return f exactly.
func (a Amount) Format(thousandsSep, decimalSep byte, precision int) string {
	return strfmt.FormatFloat(float64(a), thousandsSep, decimalSep, precision, true)
}

// BigFloat returns m as a new big.Float
func (a Amount) BigFloat() *big.Float {
	return big.NewFloat(float64(a))
}

func (a *Amount) Equal(b *Amount) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Copysign returns an Amount with the magnitude
// of a and with the sign of the sign argument.
func (a Amount) Copysign(sign Amount) Amount {
	return Amount(math.Copysign(float64(a), float64(sign)))
}

// Abs returns the absolute value of a.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Amount) Abs() Amount {
	return Amount(math.Abs(float64(a)))
}

// Valid returns if a is not infinite or NaN
func (a Amount) Valid() bool {
	return !math.IsNaN(float64(a)) && !math.IsInf(float64(a), 0)
}

func (a Amount) ValidAndGreaterZero() bool {
	return a.Valid() && a > 0
}

func (a Amount) ValidAndSmallerZero() bool {
	return a.Valid() && a < 0
}

// ValidAndHasSign returns if a.Valid() and
// if it has the same sign than the passed int argument or any sign if 0 is passed.
func (a Amount) ValidAndHasSign(sign int) bool {
	if !a.Valid() {
		return false
	}
	switch {
	case sign > 0:
		return a > 0
	case sign < 0:
		return a < 0
	}
	return true
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// and accepts numbers, strings and null.
// null will set the amout to zero.
func (a *Amount) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return errors.New("can't unmarshal empty JSON")
	}

	if bytes.Equal(b, []byte("null")) {
		*a = 0
		return nil
	}

	s := string(b)
	if len(b) > 2 && b[0] == '"' && b[len(b)-1] == '"' {
		s = s[1 : len(b)-1]
	}

	amount, err := ParseAmount(s)
	if err != nil {
		return errors.Wrapf(err, "can't unmarshal JSON as money.Amount: %s", b)
	}
	*a = amount
	return nil
}
