package txtfmt

import (
	"reflect"
	"testing"

	"github.com/domonda/go-types/money"
)

var caseSet = map[*FormatConfig]map[reflect.Value]string{
	NewEnglishFormatConfig(): {
		reflect.ValueOf(true):                     "YES",
		reflect.ValueOf(false):                    "NO",
		reflect.ValueOf(money.Amount(123.456)):    "123.46",
		reflect.ValueOf(money.Amount(123456.789)): "123,456.79",
	},
	NewGermanFormatConfig(): {
		reflect.ValueOf(true):                     "JA",
		reflect.ValueOf(false):                    "NEIN",
		reflect.ValueOf(money.Amount(123.456)):    "123,46",
		reflect.ValueOf(money.Amount(123456.789)): "123.456,79",
	},
}

func Test_FormatValue(t *testing.T) {
	for config, cases := range caseSet {
		for val, expected := range cases {
			got := FormatValue(val, config)
			if expected != got {
				t.Fatalf("expected %s got %s", expected, got)
			}
		}
	}
}
