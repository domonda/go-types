package assign

import (
	"reflect"
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
)

type StringParser struct {
	TrueStrings                 []string                        `json:"trueStrings"`
	FalseStrings                []string                        `json:"falseStrings"`
	TimeFormats                 []string                        `json:"timeFormats"`
	AcceptedMoneyAmountDecimals []int                           `json:"acceptedMoneyAmountDecimals,omitempty"`
	TypeAssigners               map[reflect.Type]StringAssigner `json:"-"`
}

func NewStringParser() *StringParser {
	p := &StringParser{
		TrueStrings:  []string{"true", "TRUE", "yes", "YES"},
		FalseStrings: []string{"false", "FALSE", "no", "NO"},
		TimeFormats: []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
		},
		AcceptedMoneyAmountDecimals: []int{0, 2, 4},
	}
	p.initTypeAssigners()
	return p
}

func (p *StringParser) initTypeAssigners() {
	p.TypeAssigners = map[reflect.Type]StringAssigner{
		reflect.TypeOf((*time.Time)(nil)).Elem():            StringAssignerFunc(assignTimeString),
		reflect.TypeOf((*time.Duration)(nil)).Elem():        StringAssignerFunc(assignDurationString),
		reflect.TypeOf((*date.Date)(nil)).Elem():            StringAssignerFunc(assignDateString),
		reflect.TypeOf((*date.NullableDate)(nil)).Elem():    StringAssignerFunc(assignNullableDateString),
		reflect.TypeOf((*money.Amount)(nil)).Elem():         StringAssignerFunc(assignMoneyAmountString),
		reflect.TypeOf((*money.CurrencyAmount)(nil)).Elem(): StringAssignerFunc(assignMoneyCurrencyAmountString),
	}
}
