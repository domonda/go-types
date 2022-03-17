package strfmt

import (
	"reflect"
	"testing"
	"time"

	"github.com/domonda/go-types/bank"
	"github.com/domonda/go-types/date"
	"github.com/domonda/go-types/money"
	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/uu"
)

var caseSet = map[*FormatConfig]map[any]string{
	NewEnglishFormatConfig(): {
		// nil and zero
		"":                           "",
		(*string)(nil):               "",
		(*float64)(nil):              "",
		reflect.ValueOf([]byte(nil)): "",
		uu.IDNull:                    "",
		money.NullableCurrency(""):   "",
		bank.NullableIBAN(""):        "",
		bank.NullableBIC(""):         "",
		time.Time{}:                  "",
		new(time.Time):               "",
		nullable.TimeNull:            "",
		reflect.Value{}:              "",
		any(nil):                     "",
		// booleans
		true:  "yes",
		false: "no",
		// amounts
		money.Amount(123.456):          "123.46",
		ptrMoneyAmount(178.456):        "178.46",
		ptrPtrMoneyAmount(189.456):     "189.46",
		money.Amount(123456.789):       "123,456.79",
		ptrMoneyAmount(1789101.789):    "1,789,101.79",
		ptrPtrMoneyAmount(1891011.789): "1,891,011.79",
		// date / time
		date.Date("2020-12-01"):                          "01/12/2020",
		ptrDateDate("2021-12-01"):                        "01/12/2021",
		ptrPtrDateDate("2022-12-01"):                     "01/12/2022",
		time.Date(2022, 02, 10, 14, 15, 59, 0, time.UTC): "10/02/2022 14:15:59 UTC",
	},
	NewGermanFormatConfig(): {
		// nil and zero
		"":                           "",
		(*string)(nil):               "",
		(*float64)(nil):              "",
		reflect.ValueOf([]byte(nil)): "",
		uu.IDNull:                    "",
		money.NullableCurrency(""):   "",
		bank.NullableIBAN(""):        "",
		bank.NullableBIC(""):         "",
		time.Time{}:                  "",
		new(time.Time):               "",
		nullable.TimeNull:            "",
		reflect.Value{}:              "",
		any(nil):                     "",
		// booleans
		true:  "ja",
		false: "nein",
		// amounts
		money.Amount(123.456):          "123,46",
		ptrMoneyAmount(178.456):        "178,46",
		ptrPtrMoneyAmount(189.456):     "189,46",
		money.Amount(123456.789):       "123.456,79",
		ptrMoneyAmount(1789101.789):    "1.789.101,79",
		ptrPtrMoneyAmount(1891011.789): "1.891.011,79",
		// date / time
		date.Date("2020-12-01"):                          "01.12.2020",
		ptrDateDate("2021-12-01"):                        "01.12.2021",
		ptrPtrDateDate("2022-12-01"):                     "01.12.2022",
		time.Date(2022, 02, 10, 14, 15, 59, 0, time.UTC): "10.02.2022 14:15:59 UTC",
	},
}

func TestFormat(t *testing.T) {
	for config, cases := range caseSet {
		for val, expected := range cases {
			got := Format(val, config)
			if expected != got {
				t.Fatalf("Format(%#v) = %s, expected = %s", val, got, expected)
			}
		}
	}
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
