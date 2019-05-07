package assign

import (
	"reflect"
	"time"

	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
)

type StringConfig struct {
	TrueStrings         []string
	FalseStrings        []string
	TimeFormats         []string
	MoneyAmountDecimals []int
	TypeAssigners       map[reflect.Type]StringAssigner
}

func NewStringConfig() *StringConfig {
	return &StringConfig{
		TrueStrings:  []string{"true", "TRUE", "yes", "YES"},
		FalseStrings: []string{"false", "FALSE", "no", "NO"},
		TimeFormats: []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
		},
		MoneyAmountDecimals: []int{2},
		TypeAssigners: map[reflect.Type]StringAssigner{
			reflect.TypeOf((*time.Time)(nil)).Elem():            StringAssignerFunc(assignTimeString),
			reflect.TypeOf((*time.Duration)(nil)).Elem():        StringAssignerFunc(assignDurationString),
			reflect.TypeOf((*date.Date)(nil)).Elem():            StringAssignerFunc(assignDateString),
			reflect.TypeOf((*date.NullableDate)(nil)).Elem():    StringAssignerFunc(assignNullableDateString),
			reflect.TypeOf((*money.Amount)(nil)).Elem():         StringAssignerFunc(assignMoneyAmountString),
			reflect.TypeOf((*money.CurrencyAmount)(nil)).Elem(): StringAssignerFunc(assignMoneyCurrencyAmountString),
		},
	}
}
