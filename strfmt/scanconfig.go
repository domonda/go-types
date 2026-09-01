package strfmt

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/strutil"
)

// DefaultScanConfig is a package-level ScanConfig initialized with NewScanConfig defaults,
// suitable for general-purpose string scanning without further configuration.
var DefaultScanConfig = NewScanConfig()

// ScanConfig holds the settings that control how strings are interpreted
// when scanned into typed values by Scan. It defines which strings are
// treated as true, false, nil, and how date/time strings are parsed.
// An optional ValidateFunc can be used to validate values after scanning.
type ScanConfig struct {
	// TrueStrings are the source strings that IsTrue reports as true.
	// Scan compares them with a bool destination after trimming
	// surrounding whitespace from the source string.
	TrueStrings []string `json:"trueStrings"`

	// FalseStrings are the source strings that IsFalse reports as false.
	// Scan compares them with a bool destination after trimming
	// surrounding whitespace from the source string.
	FalseStrings []string `json:"falseStrings"`

	// NilStrings are the source strings that IsNil reports as nil,
	// meaning "no value" like an empty cell of a CSV file or a
	// spreadsheet. Scan compares the source string with surrounding
	// whitespace trimmed against them before scanning it in any other
	// way, see StrictEmptyStringParsing and the documentation of Scan.
	NilStrings []string `json:"nilStrings"`

	// TimeFormats are the layouts that ParseTime tries in order to parse
	// a source string as time.Time or nullable.Time, so they have to be
	// ordered from the most to the least specific one.
	TimeFormats []string `json:"timeFormats"`

	// AcceptedMoneyAmountDecimals are the numbers of decimal digits that
	// a money amount string may have, where an empty slice accepts any
	// number of them. Scan doesn't use it because a money.Amount
	// destination is scanned by its own ScanString method. It's part of
	// this config so that the same configuration can be passed on to
	// money.ParseAmount or money.NewAmountParser.
	AcceptedMoneyAmountDecimals []int `json:"acceptedMoneyAmountDecimals,omitempty"`

	// StrictEmptyStringParsing controls how Scan handles a nil source
	// string, like the empty string, for a destination type that can
	// neither hold nil nor be set to null nor hold the source string
	// itself, like a number, a bool or time.Time.
	//
	// If true, such a destination is reported as an error instead of
	// parsing an empty cell of a CSV file or a spreadsheet as the
	// number zero, which can't be told apart from a scanned zero.
	// A Scannable or encoding.TextUnmarshaler destination still gets
	// asked, so it can give a nil source string a meaning instead of an
	// error, and so does a Scanner registered at TypeScanners, which is
	// asked before any of this.
	//
	// If false, which is the default, the zero value of the destination
	// type is assigned, which keeps a row with empty cells scannable
	// into a column of any type.
	StrictEmptyStringParsing bool `json:"strictEmptyStringParsing"`

	// TypeScanners hold a Scanner per destination type, registered with
	// SetTypeScanner. Scan asks the Scanner of a destination type before
	// all of its built-in scanning logic, including the nil source string
	// handling, so a Scanner owns its type completely and can give a nil
	// source string a meaning of its own. The scanners registered by
	// NewScanConfig don't: they apply the same nil source string handling
	// that a type without a registered Scanner gets,
	// see StrictEmptyStringParsing.
	TypeScanners map[reflect.Type]Scanner `json:"-"`

	// ValidateFunc reports an invalid value as an error. Scan calls it
	// with a value that it scanned itself into one of the basic kinds.
	// A Scannable destination validates itself and only gets told
	// whether ValidateFunc is not nil, while a value from a registered
	// Scanner, from an UnmarshalText method, or set to nil, null or its
	// zero value for a nil source string is not validated. A string
	// destination assigned the source string of a nil string is
	// validated like any other scanned value.
	// Use nil to disable validation.
	ValidateFunc func(any) error `json:"-"`
}

// NewScanConfig returns a ScanConfig with sensible defaults:
// common true/false string variants, an empty string and "null"/"nil"
// variants as nil indicators, several time layouts ordered from most
// to least specific, accepted money decimal counts of 0, 2, and 4,
// non-strict empty string parsing, and the package-level types.Validate
// function as ValidateFunc.
// Type-specific scanners for time.Time, nullable.Time
// and time.Duration are pre-registered.
func NewScanConfig() *ScanConfig {
	c := &ScanConfig{
		TrueStrings:  []string{"true", "True", "TRUE", "yes", "Yes", "YES", "1"},
		FalseStrings: []string{"false", "False", "FALSE", "no", "No", "NO", "0"},
		NilStrings:   []string{"", "nil", "<nil>", "null", "NULL"},
		TimeFormats: []string{
			time.RFC3339Nano,
			time.RFC3339,
			time.DateOnly + " 15:04:05.999999999 -0700 MST", // Used by time.Time.String()
			time.DateTime,
			time.DateOnly + " 15:04",
			time.DateOnly + "T15:04", // Used by browser datetime-local input type
			time.DateOnly,
		},
		AcceptedMoneyAmountDecimals: []int{0, 2, 4},
		ValidateFunc:                types.Validate,
	}
	c.initTypeScanners()
	return c
}

// initTypeScanners registers the built-in scanners.
// None of them handles a nil source string in a special way: Scan either
// resolves it before calling a scanner or passes it on so that the type
// can report it, see StrictEmptyStringParsing.
func (c *ScanConfig) initTypeScanners() {
	c.TypeScanners = map[reflect.Type]Scanner{
		reflect.TypeFor[time.Time]():     ScannerFunc(scanTimeString),
		reflect.TypeFor[nullable.Time](): ScannerFunc(scanNullableTimeString),
		reflect.TypeFor[time.Duration](): ScannerFunc(scanDurationString),
	}
}

// SetTypeScanner registers a custom Scanner for the given reflect.Type,
// replacing any previously registered scanner for that type.
// The registered scanner takes precedence over all built-in scanning logic
// for a destination of that type, including the handling of a nil source
// string, which it therefore has to handle itself if the default of
// ScanConfig.StrictEmptyStringParsing is not what it wants.
func (c *ScanConfig) SetTypeScanner(t reflect.Type, s Scanner) {
	c.TypeScanners[t] = s
}

// IsTrue reports whether str is one of the configured true strings.
func (c *ScanConfig) IsTrue(str string) bool {
	return slices.Contains(c.TrueStrings, str)
}

// IsFalse reports whether str is one of the configured false strings.
func (c *ScanConfig) IsFalse(str string) bool {
	return slices.Contains(c.FalseStrings, str)
}

// IsNil reports whether str is one of the configured nil strings.
func (c *ScanConfig) IsNil(str string) bool {
	return slices.Contains(c.NilStrings, str)
}

// ParseTime tries to parse str using each of the configured TimeFormats in order,
// returning the first successfully parsed time.Time and ok=true.
// If no format matches, it returns the zero time.Time and ok=false.
func (c *ScanConfig) ParseTime(str string) (t time.Time, ok bool) {
	for _, format := range c.TimeFormats {
		t, err := time.Parse(format, str)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// scanNilStringToNonOptional applies the nil source string handling of Scan
// to a destination type that can neither hold nil nor be set to null, like
// time.Time or time.Duration, and reports whether it did.
//
// Scan asks a registered Scanner before its own nil source string handling,
// so that a Scanner can give a nil source string its own meaning. The
// scanners registered by this package don't, they apply the same handling
// as a destination type without a registered Scanner would get.
func scanNilStringToNonOptional(dest reflect.Value, str string, config *ScanConfig) (handled bool, err error) {
	if !config.IsNil(strutil.TrimSpace(str)) {
		return false, nil
	}
	if config.StrictEmptyStringParsing {
		return true, fmt.Errorf("can't scan nil string %q into non-optional destination type %s with strict empty string parsing", str, dest.Type())
	}
	dest.SetZero()
	return true, nil
}

// scanTimeString scans str as time.Time using the TimeFormats of the config.
func scanTimeString(dest reflect.Value, str string, config *ScanConfig) error {
	if handled, err := scanNilStringToNonOptional(dest, str, config); handled {
		return err
	}
	t, ok := config.ParseTime(strutil.TrimSpace(str))
	if !ok {
		return fmt.Errorf("can't scan %q as time.Time", str)
	}
	dest.Set(reflect.ValueOf(t))
	return nil
}

// scanNullableTimeString scans str as nullable.Time using the TimeFormats
// of the config, which the UnmarshalText method of nullable.Time would
// ignore in favor of the RFC3339 format alone.
func scanNullableTimeString(dest reflect.Value, str string, config *ScanConfig) error {
	// The zero nullable.Time is its null value, and unlike time.Time it can
	// represent "no value", so a nil source string is never an error for it
	if config.IsNil(strutil.TrimSpace(str)) {
		dest.SetZero()
		return nil
	}
	t, ok := config.ParseTime(strutil.TrimSpace(str))
	if !ok {
		return fmt.Errorf("can't scan %q as nullable.Time", str)
	}
	dest.Set(reflect.ValueOf(nullable.TimeFrom(t)))
	return nil
}

// scanDurationString scans str as time.Duration.
func scanDurationString(dest reflect.Value, str string, config *ScanConfig) error {
	if handled, err := scanNilStringToNonOptional(dest, str, config); handled {
		return err
	}
	trimmed := strutil.TrimSpace(str)
	d, err := time.ParseDuration(trimmed)
	if err != nil {
		// A number without a unit means nanoseconds, which is how
		// a time.Duration is represented as its underlying int64 kind
		// and how it's marshalled as JSON
		ns, nsErr := strconv.ParseInt(trimmed, 10, 64)
		if nsErr != nil {
			return fmt.Errorf("can't scan %q as time.Duration because %w", str, err)
		}
		d = time.Duration(ns)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}
