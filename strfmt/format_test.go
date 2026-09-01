package strfmt

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// textMarshalerStringer implements both encoding.TextMarshaler and fmt.Stringer
// with different results to tell which one FormatValue used.
type textMarshalerStringer struct{}

func (textMarshalerStringer) MarshalText() ([]byte, error) { return []byte("from MarshalText"), nil }
func (textMarshalerStringer) String() string               { return "from String" }

// failingTextMarshalerStringer returns a text together with an error
// from MarshalText, which must not be used by FormatValue.
type failingTextMarshalerStringer struct{}

func (failingTextMarshalerStringer) MarshalText() ([]byte, error) {
	return []byte("incomplete text"), errors.New("MarshalText failed")
}
func (failingTextMarshalerStringer) String() string { return "from String" }

// TestFormatTextMarshaler checks the documented precedence of
// encoding.TextMarshaler over the generic type conversions:
// a successfully marshalled text is used, a failed one is not.
func TestFormatTextMarshaler(t *testing.T) {
	config := NewFormatConfig()

	got := Format(textMarshalerStringer{}, config)
	if got != "from MarshalText" {
		t.Errorf("Format(textMarshalerStringer{}) = %q, want %q", got, "from MarshalText")
	}

	got = Format(failingTextMarshalerStringer{}, config)
	if got != "from String" {
		t.Errorf("Format(failingTextMarshalerStringer{}) = %q, want %q", got, "from String")
	}

	// Observable effect for a go-types type: the MarshalText method of
	// nullable.TrimmedString normalizes the value, while the string kind
	// of its underlying type still holds the untrimmed source string.
	got = Format(nullable.TrimmedString(" trimmed "), config)
	if got != "trimmed" {
		t.Errorf("Format(nullable.TrimmedString(%q)) = %q, want %q", " trimmed ", got, "trimmed")
	}
}

// ptrStringer implements fmt.Stringer with a pointer receiver, so the method
// is only reachable through the address of an addressable value.
type ptrStringer struct{ Text string }

func (s *ptrStringer) String() string { return "from *String: " + s.Text }

// TestFormatValueStringerFallbacks checks the fallbacks below the kind
// conversions: a Stringer reachable only through the address of the value,
// and a []byte formatted as the string it holds. Neither has a conversion
// based on its kind, so without them both would end up in fmt.Sprint.
func TestFormatValueStringerFallbacks(t *testing.T) {
	config := NewFormatConfig()

	// A value receiver can't reach a pointer receiver method, so this
	// only works through the val.Addr() branch of FormatValue
	strct := struct{ S ptrStringer }{ptrStringer{Text: "text"}}
	val := reflect.ValueOf(&strct).Elem().Field(0)
	require.True(t, val.CanAddr(), "test needs an addressable value")
	assert.Equal(t, "from *String: text", FormatValue(val, config))

	// The same value passed by value has no address to reach the method
	assert.Equal(t, "{text}", Format(ptrStringer{Text: "text"}, config))

	assert.Equal(t, "bytes", Format([]byte("bytes"), config))
}

// TestFormatValueUnexportedField checks that the value of an addressable
// unexported struct field is formatted like an exported one instead of
// panicking in reflect.Value.Interface or falling back to the raw internals
// of a struct or array kind. A reflection driven caller walking a struct
// must neither crash on one of its fields nor write "{0 63902963045 <nil>}"
// into a cell where a date belongs.
func TestFormatValueUnexportedField(t *testing.T) {
	config := NewFormatConfig()
	strct := struct {
		i  int
		s  string
		f  float64
		b  bool
		t  time.Time
		a  money.Amount
		d  date.Date
		id uu.ID
	}{
		i:  666,
		s:  "text",
		f:  6.66,
		b:  true,
		t:  time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC),
		a:  money.Amount(1234.5),
		d:  date.Date("2026-01-02"),
		id: uu.IDMustFromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
	}
	val := reflect.ValueOf(&strct).Elem()

	assert.Equal(t, "666", FormatValue(val.FieldByName("i"), config))
	assert.Equal(t, "text", FormatValue(val.FieldByName("s"), config))
	assert.Equal(t, "6.66", FormatValue(val.FieldByName("f"), config))
	assert.Equal(t, "true", FormatValue(val.FieldByName("b"), config))
	// A registered TypeFormatter is used, not the raw struct fields
	assert.Equal(t, "2026-01-02T15:04:05Z", FormatValue(val.FieldByName("t"), config))
	assert.Equal(t, "1,234.50", FormatValue(val.FieldByName("a"), config))
	assert.Equal(t, "2026-01-02", FormatValue(val.FieldByName("d"), config))
	// A MarshalText method is used, not the raw byte array
	assert.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", FormatValue(val.FieldByName("id"), config))

	// The same values of an exported field format identically
	exported := struct {
		T  time.Time
		A  money.Amount
		ID uu.ID
	}{strct.t, strct.a, strct.id}
	exportedVal := reflect.ValueOf(&exported).Elem()
	for _, name := range []string{"T", "A", "ID"} {
		unexportedName := strings.ToLower(name)
		assert.Equal(t,
			FormatValue(exportedVal.FieldByName(name), config),
			FormatValue(val.FieldByName(unexportedName), config),
			"unexported field %q formats like exported field %q", unexportedName, name,
		)
	}

	// A nil value is formatted as config.Nil because
	// nullable.ReflectIsNull doesn't need to interface it
	nilFields := struct {
		p *int
		s []int
	}{}
	nilVal := reflect.ValueOf(&nilFields).Elem()
	assert.Equal(t, config.Nil, FormatValue(nilVal.FieldByName("p"), config))
	assert.Equal(t, config.Nil, FormatValue(nilVal.FieldByName("s"), config))
}

// TestFormatValueNonAddressableUnexportedField checks that the value of an
// unexported field of a struct that was not passed as a pointer, which can
// neither be interfaced nor read through an address, falls back to fmt.Sprint
// instead of panicking. There is no way to call a String or MarshalText
// method on such a value, so fmt is the last sane formatting available.
func TestFormatValueNonAddressableUnexportedField(t *testing.T) {
	config := NewFormatConfig()
	strct := struct {
		i int
		t time.Time
	}{
		i: 666,
		t: time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC),
	}
	val := reflect.ValueOf(strct) // not a pointer, so not addressable
	require.False(t, val.Field(0).CanAddr(), "test needs a non-addressable value")

	// The conversions based on the kind still work
	assert.Equal(t, "666", FormatValue(val.Field(0), config))
	// A struct kind has none, so fmt.Sprint is the fallback
	assert.NotPanics(t, func() { FormatValue(val.Field(1), config) })
	assert.NotEmpty(t, FormatValue(val.Field(1), config))
}
