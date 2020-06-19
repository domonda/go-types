package txtfmt

import (
	"reflect"
	"testing"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/uu"
)

var caseSet = map[*FormatConfig]map[reflect.Value]string{
	NewEnglishFormatConfig(): {
		// nills/nulls
		reflect.ValueOf(voidStr()):                   "",
		reflect.ValueOf(voidFloat()):                 "",
		reflect.ValueOf(uu.NullableIDFrom(uu.IDNil)): "NULL",
		reflect.ValueOf(money.NullableCurrency("")):  "",
		reflect.ValueOf(bank.NullableIBAN("")):       "",
		reflect.ValueOf(bank.NullableBIC("")):        "",
		// booleans
		reflect.ValueOf(true):  "YES",
		reflect.ValueOf(false): "NO",
		// amounts
		reflect.ValueOf(money.Amount(123.456)):          "123.46",
		reflect.ValueOf(ptrMoneyAmount(178.456)):        "178.46",
		reflect.ValueOf(ptrPtrMoneyAmount(189.456)):     "189.46",
		reflect.ValueOf(money.Amount(123456.789)):       "123,456.79",
		reflect.ValueOf(ptrMoneyAmount(1789101.789)):    "1,789,101.79",
		reflect.ValueOf(ptrPtrMoneyAmount(1891011.789)): "1,891,011.79",
		// dates
		reflect.ValueOf(date.Date("2020-12-01")):      "01/12/2020",
		reflect.ValueOf(ptrDateDate("2021-12-01")):    "01/12/2021",
		reflect.ValueOf(ptrPtrDateDate("2022-12-01")): "01/12/2022",
	},
	NewGermanFormatConfig(): {
		// nills/nulls
		reflect.ValueOf(voidStr()):                   "",
		reflect.ValueOf(voidFloat()):                 "",
		reflect.ValueOf(uu.NullableIDFrom(uu.IDNil)): "NULL",
		reflect.ValueOf(money.NullableCurrency("")):  "",
		reflect.ValueOf(bank.NullableIBAN("")):       "",
		reflect.ValueOf(bank.NullableBIC("")):        "",
		// booleans
		reflect.ValueOf(true):  "JA",
		reflect.ValueOf(false): "NEIN",
		// amounts
		reflect.ValueOf(money.Amount(123.456)):          "123,46",
		reflect.ValueOf(ptrMoneyAmount(178.456)):        "178,46",
		reflect.ValueOf(ptrPtrMoneyAmount(189.456)):     "189,46",
		reflect.ValueOf(money.Amount(123456.789)):       "123.456,79",
		reflect.ValueOf(ptrMoneyAmount(1789101.789)):    "1.789.101,79",
		reflect.ValueOf(ptrPtrMoneyAmount(1891011.789)): "1.891.011,79",
		// dates
		reflect.ValueOf(date.Date("2020-12-01")):      "01.12.2020",
		reflect.ValueOf(ptrDateDate("2021-12-01")):    "01.12.2021",
		reflect.ValueOf(ptrPtrDateDate("2022-12-01")): "01.12.2022",
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

func voidStr() *string {
	return nil
}

func voidFloat() *float64 {
	return nil
}

func ptrMoneyAmount(a float64) *money.Amount {
	x := money.Amount(a)
	return &x
}

func ptrPtrMoneyAmount(a float64) **money.Amount {
	x := money.Amount(a)
	y := &x
	return &y
}

func ptrDateDate(d string) *date.Date {
	x := date.Date(d)
	return &x
}

func ptrPtrDateDate(d string) **date.Date {
	x := date.Date(d)
	y := &x
	return &y
}
