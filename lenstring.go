package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// LenString holds a string together with a minimum and maximum length constraints.
// It implements validation to ensure the string length falls within the specified bounds.
// LenString implements the encoding.UnmarshalText, json.Unmarshaler,
// and strfmt.StringAssignable interfaces that will perform length validation.
type LenString struct {
	str string
	min int
	max int
}

// NewLenString returns a new LenString without validating it.
// Use this when you want to create a LenString and validate it separately.
func NewLenString(str string, min, max int) *LenString {
	return &LenString{
		str: str,
		min: min,
		max: max,
	}
}

// MustLenString returns a LenString or panics on errors from Validate.
// Use this when you're certain the string meets the length constraints.
func MustLenString(str string, min, max int) LenString {
	s := LenString{
		str: str,
		min: min,
		max: max,
	}
	err := s.Validate()
	if err != nil {
		panic(err)
	}
	return s
}

// Validate implements the ValidatErr interface.
// It checks that the minimum and maximum lengths are valid and that the string
// length falls within the specified bounds.
func (s *LenString) Validate() error {
	if s == nil {
		return errors.New("nil LenString")
	}
	if s.min < 0 {
		return fmt.Errorf("negative minimum length %d of LenString %q", s.min, s.str)
	}
	if s.max < 0 {
		return fmt.Errorf("negative maximum length %d of LenString %q", s.max, s.str)
	}
	if s.min > s.max {
		return fmt.Errorf("minimum length %d is greater than maximum length %d of LenString %q", s.min, s.max, s.str)
	}
	return s.validateLen(s.str)
}

// validateLen checks if the string length is within the min/max bounds.
func (s *LenString) validateLen(str string) error {
	l := len(str)
	if l < s.min {
		return fmt.Errorf("length %d of LenString %q is shorter than minimum of %d", l, str, s.min)
	}
	if l > s.max {
		return fmt.Errorf("length %d of LenString %q is longer than maximum of %d", l, str, s.max)
	}
	return nil
}

// String returns the string or "<nil>".
// String implements the fmt.Stringer interface.
func (s *LenString) String() string {
	if s == nil {
		return "<nil>"
	}
	return s.str
}

// SetString sets the string value after validating it against the length constraints.
// Returns an error if the string doesn't meet the length requirements.
func (s *LenString) SetString(str string) error {
	if err := s.validateLen(str); err != nil {
		return err
	}
	s.str = str
	return nil
}

// MinLen returns the minimum allowed length for the string.
func (s *LenString) MinLen() int {
	return s.min
}

// MaxLen returns the maximum allowed length for the string.
func (s *LenString) MaxLen() int {
	return s.max
}

// MarshalText implements the encoding.TextMarshaler interface.
func (s *LenString) MarshalText() (text []byte, err error) {
	if s == nil {
		return nil, nil
	}
	return []byte(s.str), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// It validates the text length before setting the string value.
func (s *LenString) UnmarshalText(text []byte) error {
	return s.SetString(string(text))
}

// MarshalJSON implements the json.Marshaler interface.
func (s *LenString) MarshalJSON() (text []byte, err error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.str)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It validates the string length after unmarshaling from JSON.
func (s *LenString) UnmarshalJSON(text []byte) error {
	if bytes.Equal(text, []byte("null")) {
		return nil // no-op
	}
	var str string
	err := json.Unmarshal(text, &s.str)
	if err != nil {
		return fmt.Errorf("can't unmarshal JSON to LenString because of: %w", err)
	}
	return s.SetString(str)
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
func (s *LenString) ScanString(source string, validate bool) error {
	if validate {
		if err := s.validateLen(source); err != nil {
			return err
		}
	}
	s.str = source
	return nil
}
