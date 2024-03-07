package strfmt

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	types "github.com/domonda/go-types"
)

var DefaultScanConfig = NewScanConfig()

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

func NewScanConfig() *ScanConfig {
	c := &ScanConfig{
		TrueStrings:  []string{"true", "TRUE", "yes", "YES", "1"},
		FalseStrings: []string{"false", "FALSE", "no", "NO", "0"},
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
		reflect.TypeOf((*time.Time)(nil)).Elem():     ScannerFunc(scanTimeString),
		reflect.TypeOf((*time.Duration)(nil)).Elem(): ScannerFunc(scanDurationString),
	}
}

func (c *ScanConfig) SetTypeScanner(t reflect.Type, s Scanner) {
	c.TypeScanners[t] = s
}

func (c *ScanConfig) IsTrue(str string) bool {
	for _, val := range c.TrueStrings {
		if str == val {
			return true
		}
	}
	return false
}

func (c *ScanConfig) IsFalse(str string) bool {
	for _, val := range c.FalseStrings {
		if str == val {
			return true
		}
	}
	return false
}

func (c *ScanConfig) IsNil(str string) bool {
	for _, val := range c.NilStrings {
		if str == val {
			return true
		}
	}
	return false
}

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
	t, ok := config.ParseTime(strings.TrimSpace(str))
	if !ok {
		return fmt.Errorf("can't scan %q as time.Time", str)
	}
	dest.Set(reflect.ValueOf(t))
	return nil
}

func scanDurationString(dest reflect.Value, str string, config *ScanConfig) error {
	d, err := time.ParseDuration(strings.TrimSpace(str))
	if err != nil {
		return fmt.Errorf("can't scan %q as time.Duration because %w", str, err)
	}
	dest.Set(reflect.ValueOf(d))
	return nil
}
