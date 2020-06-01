package strfmt

import (
	"reflect"
	"time"
)

var DefaultScanConfig = NewScanConfig()

type ScanConfig struct {
	TrueStrings                 []string                 `json:"trueStrings"`
	FalseStrings                []string                 `json:"falseStrings"`
	TimeFormats                 []string                 `json:"timeFormats"`
	AcceptedMoneyAmountDecimals []int                    `json:"acceptedMoneyAmountDecimals,omitempty"`
	TypeScanners                map[reflect.Type]Scanner `json:"-"`
}

func NewScanConfig() *ScanConfig {
	c := &ScanConfig{
		TrueStrings:  []string{"true", "TRUE", "yes", "YES"},
		FalseStrings: []string{"false", "FALSE", "no", "NO"},
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
