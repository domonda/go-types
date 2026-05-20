package float

import (
	"fmt"
	"math"
	"strconv"
)

// Tolerant is a float64 underneath
// that is tolerant in parsing non standard JSON
type Tolerant float64

// AsFloatPtr returns f as *float64
func (f *Tolerant) AsFloatPtr() *float64 {
	return (*float64)(f)
}

// Valid returns if the float is neither infinite nor NaN
func (f Tolerant) Valid() bool {
	return !f.IsInf() && !f.IsNaN()
}

// ValidAndGreaterZero returns if the float is neither infinite nor NaN
// and greater than zero.
func (f Tolerant) ValidAndGreaterZero() bool {
	return f.Valid() && f > 0
}

// ValidAndSmallerZero returns if the float is neither infinite nor NaN
// and smaller than zero.
func (f Tolerant) ValidAndSmallerZero() bool {
	return f.Valid() && f < 0
}

// ValidAndHasSign returns if a.Valid() and
// if it has the same sign than the passed int argument or any sign if 0 is passed.
func (f Tolerant) ValidAndHasSign(sign int) bool {
	return ValidAndHasSign(float64(f), sign)
}

// IsNaN returns if the float is not a number (NaN)
func (f Tolerant) IsNaN() bool {
	return math.IsNaN(float64(f))
}

// IsNaN returns if the float is positive or negative infinity
func (f Tolerant) IsInf() bool {
	return math.IsInf(float64(f), 0)
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// and accepts numbers, strings, and null.
// JSON null and "" will set the amount to zero.
func (f *Tolerant) UnmarshalJSON(j []byte) error {
	s := string(j)

	if s == `null` || s == `""` {
		*f = 0
		return nil
	}

	// Unquote JSON string values so escaped contents are decoded correctly
	if len(s) > 0 && s[0] == '"' {
		unquoted, err := strconv.Unquote(s)
		if err != nil {
			return fmt.Errorf("can't unmarshal JSON(%s) as float.Tolerant because of: %w", j, err)
		}
		s = unquoted
	}

	parsed, err := Parse(s)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON(%s) as float.Tolerant because of: %w", j, err)
	}

	*f = Tolerant(parsed)
	return nil
}
