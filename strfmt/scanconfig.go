package strfmt

import (
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/domonda/go-types"
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
	TrueStrings                 []string `json:"trueStrings"`
	FalseStrings                []string `json:"falseStrings"`
	NilStrings                  []string `json:"nilStrings"`
	TimeFormats                 []string `json:"timeFormats"`
	AcceptedMoneyAmountDecimals []int    `json:"acceptedMoneyAmountDecimals,omitempty"`

	TypeScanners map[reflect.Type]Scanner `json:"-"`
	// Use nil to disable validation
	ValidateFunc func(any) error `json:"-"`
}

// NewScanConfig returns a ScanConfig with sensible defaults:
// common true/false string variants, an empty string and "null"/"nil"
// variants as nil indicators, several time layouts ordered from most
// to least specific, accepted money decimal counts of 0, 2, and 4,
// and the package-level types.Validate function as ValidateFunc.
// Type-specific scanners for time.Time and time.Duration are pre-registered.
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

func (c *ScanConfig) initTypeScanners() {
	c.TypeScanners = map[reflect.Type]Scanner{
		reflect.TypeFor[time.Time]():     ScannerFunc(scanTimeString),
		reflect.TypeFor[time.Duration](): ScannerFunc(scanDurationString),
	}
}

// SetTypeScanner registers a custom Scanner for the given reflect.Type,
// replacing any previously registered scanner for that type.
// The registered scanner takes priority over all built-in scanning logic.
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

func scanTimeString(dest reflect.Value, str string, config *ScanConfig) error {
	t, ok := config.ParseTime(strutil.TrimSpace(str))
	if !ok {
		return fmt.Errorf("can't scan %q as time.Time", str)
	}
	dest.Set(reflect.ValueOf(t))
	return nil
}

func scanDurationString(dest reflect.Value, str string, config *ScanConfig) error {
	d, err := time.ParseDuration(strutil.TrimSpace(str))
	if err != nil {
		return fmt.Errorf("can't scan %q as time.Duration because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}
