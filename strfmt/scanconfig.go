package strfmt

import (
	"reflect"
	"time"
)

var DefaultScanConfig = NewScanConfig()

type ScanConfig struct {
	TrueStrings                 []string                 `json:"trueStrings"`
	FalseStrings                []string                 `json:"falseStrings"`
	NilStrings                  []string                 `json:"nilStrings"`
	TimeFormats                 []string                 `json:"timeFormats"`
	AcceptedMoneyAmountDecimals []int                    `json:"acceptedMoneyAmountDecimals,omitempty"`
	TypeScanners                map[reflect.Type]Scanner `json:"-"`
}

func NewScanConfig() *ScanConfig {
	c := &ScanConfig{
		TrueStrings:  []string{"true", "TRUE", "yes", "YES", "1"},
		FalseStrings: []string{"false", "FALSE", "no", "NO", "0"},
		NilStrings:   []string{"", "nil", "<nil>", "null", "NULL"},
		TimeFormats: []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
		},
		AcceptedMoneyAmountDecimals: []int{0, 2, 4},
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
			return t, ok
		}
	}
	return time.Time{}, false
}
