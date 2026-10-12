package strfmt

import (
	"errors"
	"math"
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

// TestFormatNilConfig protects that a nil FormatConfig formats with the
// defaults of NewFormatConfig instead of panicking on the first config field
// access, consistent with Scan, which scans with DefaultScanConfig for a nil
// ScanConfig, so a caller without locale preferences can pass nil to both.
// A panic deep inside an export loop is a crash where a default is harmless.
func TestFormatNilConfig(t *testing.T) {
	defaults := NewFormatConfig()
	for _, value := range []any{
		true,
		1234.5,
		float32(0.1),
		"text",
		666,
		money.Amount(1234.5),
		money.CurrencyAmount{Currency: money.EUR, Amount: 1234.5},
		date.Date("2024-03-05"),
		time.Date(2024, 3, 5, 6, 7, 8, 0, time.UTC),
		nullable.TimeNull,
		(*int)(nil),
		nil,
	} {
		var got string
		if assert.NotPanics(t, func() { got = Format(value, nil) }, "Format(%#v, nil)", value) {
			assert.Equal(t, Format(value, defaults), got, "Format(%#v, nil)", value)
		}
	}
	assert.Equal(t, defaults.Nil, FormatValue(reflect.Value{}, nil), "invalid reflect.Value")
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

// valueStringer implements fmt.Stringer with a value receiver.
type valueStringer struct{ Text string }

func (s valueStringer) String() string { return "from String: " + s.Text }

// TestFormatValueIntegerKinds checks that all integer kinds are formatted
// as plain decimal numbers without thousands separators or a decimal part,
// so that IDs, counts and years are written exactly as they are.
func TestFormatValueIntegerKinds(t *testing.T) {
	config := NewGermanFormatConfig()
	tests := []struct {
		value any
		want  string
	}{
		{value: 0, want: "0"},
		{value: 1234567, want: "1234567"},
		{value: int8(-128), want: "-128"},
		{value: int64(math.MinInt64), want: "-9223372036854775808"},
		{value: uint(0), want: "0"},
		{value: uint8(255), want: "255"},
		{value: uint16(65535), want: "65535"},
		{value: uint64(math.MaxUint64), want: "18446744073709551615"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Format(tt.value, config), "Format(%T(%v))", tt.value, tt.value)
	}

	u32 := uint32(4000000)
	assert.Equal(t, "4000000", Format(&u32, config), "pointer to uint32 is dereferenced")
}

// TestFormatValueFloat32 checks that a float32 is formatted with the
// shortest decimal that parses back to the same float32. Widening it to
// float64 first exposes the binary noise of the float32 value, which is
// 0.100000001490116119384765625 for float32(0.1), so "0.10000000149011612"
// would be written into a CSV or report instead of the "0.1" it holds.
func TestFormatValueFloat32(t *testing.T) {
	type namedFloat32 float32

	config := NewFormatConfig()
	tests := []struct {
		value any
		want  string
	}{
		{value: float32(0.1), want: "0.1"},
		{value: float32(-1.1), want: "-1.1"},
		{value: float32(3.14159), want: "3.14159"},
		{value: float32(16777216), want: "16777216"},
		{value: namedFloat32(0.2), want: "0.2"},
		{value: float32(math.Inf(-1)), want: "-Inf"},
		{value: float64(0.1), want: "0.1"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Format(tt.value, config), "Format(%T(%v))", tt.value, tt.value)
	}

	f := float32(0.3)
	assert.Equal(t, "0,3", Format(&f, NewGermanFormatConfig()), "pointer to float32 with German decimal comma")
}

// TestFormatValueStringerThroughPointers checks that a Stringer is still
// used when the value is passed through more than one level of pointers,
// where neither the value itself nor its address has the String method in
// its method set. Without it fmt.Sprint would print a pointer address.
func TestFormatValueStringerThroughPointers(t *testing.T) {
	config := NewFormatConfig()
	s := &valueStringer{Text: "text"}
	assert.Equal(t, "from String: text", Format(&s, config))
}

// TestFormatValueNullableType checks that a non-null nullable.Type[T] without
// a registered formatter is formatted like the T it wraps, with the locale of
// the config, and null as config.Nil. nullable.Type has no String or
// MarshalText method, so it fell back to fmt.Sprint and a spreadsheet cell
// got the struct internals "{1234 true}" instead of the number, which can't
// be read or scanned back.
func TestFormatValueNullableType(t *testing.T) {
	for _, config := range []*FormatConfig{NewGermanFormatConfig(), NewEnglishFormatConfig()} {
		config.Nil = "N/A"

		n := nullable.TypeFrom(1234)
		assert.Equal(t, "1234", Format(n, config), "nullable.Type[int]")
		assert.Equal(t, "1234", Format(&n, config), "*nullable.Type[int]")
		assert.Equal(t, "text", Format(nullable.TypeFrom("text"), config), "nullable.Type[string]")
		assert.Equal(t, config.True, Format(nullable.TypeFrom(true), config), "nullable.Type[bool]")
		assert.Equal(t, Format(1234.5, config), Format(nullable.TypeFrom(1234.5), config), "nullable.Type[float64]")
		assert.Equal(t, Format(money.Rate(0.19), config), Format(money.NullableRate(nullable.TypeFrom(money.Rate(0.19))), config), "money.NullableRate")
		someDate := date.Date("2024-03-05")
		assert.Equal(t, Format(someDate, config), Format(nullable.TypeFrom(someDate), config), "nullable.Type[date.Date]")

		assert.Equal(t, "N/A", Format(nullable.Type[int]{}, config), "null nullable.Type[int]")
		assert.Equal(t, "N/A", Format(money.NullableRate{}, config), "null money.NullableRate")
	}
}
