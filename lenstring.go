package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// LenString holds a string together with a minimum and maximum length.
// Validate returns an error if the string length does not fit the minium-maxium length.
// LenString implements the encoding.UnmarshalText, json.Unmarshaler,
// and strfmt.StringAssignable interfaces that will do length validation.
type LenString struct {
	str string
	min int
	max int
}

// NewLenString returns a new LenString without validating it.
func NewLenString(str string, min, max int) *LenString {
	return &LenString{
		str: str,
		min: min,
		max: max,
	}
}

// MustLenString returns a LenString or panics on errors from Validate.
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

// Validate implements the ValidatErr interface
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

func (s *LenString) validateLen(str string) error {
	l := len(s.str)
	if l < s.min {
		return fmt.Errorf("length %d of LenString %q is shorter than minimum of %d", l, s.str, s.min)
	}
	if l > s.max {
		return fmt.Errorf("length %d of LenString %q is longer than maximum of %d", l, s.str, s.max)
	}
	return nil
}

func (s *LenString) String() string {
	if s == nil {
		return "<nil>"
	}
	return s.str
}

func (s *LenString) SetString(str string) error {
	if err := s.validateLen(str); err != nil {
		return err
	}
	s.str = str
	return nil
}

func (s *LenString) MinLen() int {
	return s.min
}

func (s *LenString) MaxLen() int {
	return s.max
}

// MarshalText implements the encoding.TextMarshaler interface
func (s *LenString) MarshalText() (text []byte, err error) {
	if s == nil {
		return nil, nil
	}
	return []byte(s.str), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *LenString) UnmarshalText(text []byte) error {
	return s.SetString(string(text))
}

// MarshalText implements the json.Marshaler interface
func (s *LenString) MarshalJSON() (text []byte, err error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.str)
}

// UnmarshalJSON implements the json.Unmarshaler interface
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

// AssignString implements the strfmt.StringAssignable interface
func (s *LenString) AssignString(source string) (normalized bool, err error) {
	err = s.SetString(source)
	if err != nil {
		return false, err
	}
	return true, nil
}
