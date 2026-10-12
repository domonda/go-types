package account

import (
	"encoding/json"
	"encoding/xml"
	"math"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumber_TrimLeadingZeros(t *testing.T) {
	tests := []struct {
		no   Number
		want Number
	}{
		{"", ""},
		{"0", ""},
		{"00", ""},
		{"1", "1"},
		{"01", "1"},
		{"10", "10"},
		{"100", "100"},
		{"0001", "1"},
		{"00010", "10"},
		{"0x12", "x12"},
		{"Hello_World", "Hello_World"},
	}
	for _, tt := range tests {
		t.Run(string(tt.no), func(t *testing.T) {
			if got := tt.no.TrimLeadingZeros(); got != tt.want {
				t.Errorf("AccountNo.TrimLeadingZeros() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNumber_Valid(t *testing.T) {
	valid := []Number{
		"0",
		"0/",
		"0/0",
		"Hello_World",
		"a.b",
		"a:b",
		"a,b",
		"a;b",
		"a.b.",
		"a:b:",
		"a,b,",
		"a;b;",
		"D-2007019",
	}
	for _, n := range valid {
		t.Run(string(n), func(t *testing.T) {
			require.Truef(t, n.Valid(), "Number(%#v).Valid()", n)
		})
	}
	invalid := []Number{
		"",
	}
	for _, n := range invalid {
		t.Run(string(n), func(t *testing.T) {
			require.Falsef(t, n.Valid(), "Number(%#v).Valid()", n)
		})
	}
}

func TestNumberFrom(t *testing.T) {
	// Account numbers come from user input and document extraction,
	// so surrounding whitespace must be trimmed before validation.
	tests := []struct {
		name    string
		str     string
		want    Number
		wantErr bool
	}{
		{name: "digits", str: "123456", want: "123456"},
		{name: "trimmed", str: " \t123-456/7 \n", want: "123-456/7"},
		{name: "alphanumeric", str: "AT-2007019", want: "AT-2007019"},
		{name: "empty is invalid", str: "", wantErr: true},
		{name: "whitespace only is invalid", str: "   ", wantErr: true},
		{name: "must start alphanumeric", str: "-123", wantErr: true},
		{name: "inner space is invalid", str: "123 456", wantErr: true},
		{name: "unsupported character", str: "123#456", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NumberFrom(tt.str)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidNumber, "callers must be able to detect invalid numbers with errors.Is")
				assert.Equal(t, Number(""), got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNumber_Validate(t *testing.T) {
	require.NoError(t, Number("D-2007019").Validate())
	err := Number("").Validate()
	require.ErrorIs(t, err, ErrInvalidNumber)
	err = Number("/123").Validate()
	require.ErrorIs(t, err, ErrInvalidNumber)
	assert.Contains(t, err.Error(), `"/123"`, "error must name the offending number")
}

// NumberRegexp is documented as reassignable at startup to enforce an
// application specific format, so Valid/Validate must use it instead of
// the DefaultNumberRegex constant.
func TestNumberRegexp_Reassignable(t *testing.T) {
	orig := NumberRegexp
	t.Cleanup(func() { NumberRegexp = orig })

	NumberRegexp = regexp.MustCompile(`^[0-9]{6}$`)
	assert.True(t, Number("123456").Valid())
	assert.False(t, Number("ABC123").Valid(), "valid by default regex, invalid by the configured one")
	_, err := NumberFrom("1234567")
	assert.ErrorIs(t, err, ErrInvalidNumber)
}

func TestNumberFromUint(t *testing.T) {
	assert.Equal(t, Number("0"), NumberFromUint(0), "a non nullable Number keeps 0 as value")
	assert.Equal(t, Number("18446744073709551615"), NumberFromUint(math.MaxUint64))
}

func TestNumber_HasPrefixSuffix(t *testing.T) {
	n := Number("AT-123/4")
	assert.True(t, n.HasPrefix("AT-"))
	assert.False(t, n.HasPrefix("DE-"))
	assert.True(t, n.HasSuffix("/4"))
	assert.False(t, n.HasSuffix("/5"))
}

func TestNumber_IsNumeric(t *testing.T) {
	tests := []struct {
		n    Number
		want bool
	}{
		{"", false}, // empty must not count as numeric
		{"0", true},
		{"0123456789", true},
		{"12a", false},
		{"0x1F", false},
		{"1.5", false},
		{"-1", false},
		{"1/2", false},
		{"١٢٣", false}, // non ASCII digits are not accepted
	}
	for _, tt := range tests {
		t.Run(string(tt.n), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.n.IsNumeric())
			if tt.want {
				assert.NoError(t, tt.n.ValidateNumeric())
			} else {
				assert.ErrorIs(t, tt.n.ValidateNumeric(), ErrAlphanumericNumber)
			}
		})
	}
}

func TestNumber_UintInt(t *testing.T) {
	// Non digit strings like hex numbers or signed numbers must not be
	// parsed by strconv, they have to fail with ErrAlphanumericNumber.
	for _, n := range []Number{"", "0x10", "+1", "-1", "1e3", "12 34"} {
		t.Run(string(n), func(t *testing.T) {
			_, err := n.Uint()
			assert.ErrorIs(t, err, ErrAlphanumericNumber)
			p, err := n.UintPtr()
			assert.ErrorIs(t, err, ErrAlphanumericNumber)
			assert.Nil(t, p)
			_, err = n.Int()
			assert.ErrorIs(t, err, ErrAlphanumericNumber)
		})
	}

	u, err := Number("00042").Uint()
	require.NoError(t, err)
	assert.Equal(t, uint64(42), u, "leading zeros are not octal")
	p, err := Number("42").UintPtr()
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, uint64(42), *p)
	i, err := Number("00042").Int()
	require.NoError(t, err)
	assert.Equal(t, int64(42), i)

	// Overflow must be reported instead of silently truncated
	_, err = Number("18446744073709551616").Uint()
	assert.Error(t, err, "MaxUint64+1")
	p, err = Number("18446744073709551616").UintPtr()
	assert.Error(t, err, "MaxUint64+1")
	assert.Nil(t, p)
	_, err = Number("9223372036854775808").Int()
	assert.Error(t, err, "MaxInt64+1")
}

func TestNumber_Cut(t *testing.T) {
	before, after, found := Number("123/4").Cut("/")
	assert.True(t, found)
	assert.Equal(t, Number("123"), before)
	assert.Equal(t, Number("4"), after)

	before, after, found = Number("1234").Cut("/")
	assert.False(t, found)
	assert.Equal(t, Number("1234"), before)
	assert.Equal(t, Number(""), after)
}

func TestNumber_StringAndNullable(t *testing.T) {
	assert.Equal(t, "AT-1", Number("AT-1").String())
	assert.Equal(t, NullableNumber("AT-1"), Number("AT-1").Nullable())
	assert.True(t, Number("").Nullable().IsNull(), "empty Number converts to NULL")
}

func TestNumber_UnmarshalText(t *testing.T) {
	var n Number
	require.NoError(t, n.UnmarshalText([]byte(" 123/4 ")))
	assert.Equal(t, Number("123/4"), n, "text must be trimmed")

	err := n.UnmarshalText([]byte("#1"))
	require.ErrorIs(t, err, ErrInvalidNumber)
	assert.Equal(t, Number("123/4"), n, "failed unmarshalling must not modify the value")

	err = n.UnmarshalText(nil)
	require.ErrorIs(t, err, ErrInvalidNumber, "a non nullable Number can't be empty")
}

func TestNumber_Value(t *testing.T) {
	// An empty Number is written as SQL NULL, not as an empty string
	v, err := Number("").Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	v, err = Number(" 123 ").Value()
	require.NoError(t, err)
	assert.Equal(t, "123", v, "whitespace must not be written to the database")
}

func TestNumber_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    Number
		wantErr bool
	}{
		{name: "string", value: " 123-4 ", want: "123-4"},
		{name: "bytes", value: []byte("AB/12"), want: "AB/12"},
		{name: "int64", value: int64(1234567), want: "1234567"},
		{name: "int64 zero", value: int64(0), want: "0"},
		{name: "float64 integer", value: float64(1234567), want: "1234567"},
		{name: "large float64 not in exponent format", value: float64(1e21), want: "1000000000000000000000"},
		{name: "invalid string", value: "#1", wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid bytes", value: []byte("-1"), wantErr: true},
		{name: "negative int64", value: int64(-1), wantErr: true},
		{name: "negative float64", value: float64(-1), wantErr: true},
		// A float64 is only an account number if it is a non-negative
		// integer. "NaN" even matches the alphanumeric number pattern,
		// so it would be stored as a valid looking account number, and
		// a fraction means the column does not hold account numbers.
		{name: "NaN", value: math.NaN(), wantErr: true},
		{name: "+Inf", value: math.Inf(1), wantErr: true},
		{name: "-Inf", value: math.Inf(-1), wantErr: true},
		{name: "non-integer float64", value: float64(1.5), wantErr: true},
		// -0 equals 0, so it must scan like int64 zero instead of
		// becoming "-0" that fails validation when written back.
		// Note that the Go constant -0.0 is +0, so Copysign is needed.
		{name: "float64 negative zero", value: math.Copysign(0, -1), want: "0"},
		{name: "NULL", value: nil, wantErr: true},
		{name: "unsupported type", value: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Number("unchanged")
			err := n.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, Number("unchanged"), n, "failed scan must not modify the value")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestNumber_Scan_CopiesBytes(t *testing.T) {
	// database/sql may reuse the []byte buffer after Scan returns
	buf := []byte("12345")
	var n Number
	require.NoError(t, n.Scan(buf))
	copy(buf, "99999")
	assert.Equal(t, Number("12345"), n)
}

func TestNumber_JSON(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		for _, n := range []Number{"0", "123456", "AT-12/34", "a.b,c;d:e_f"} {
			j, err := json.Marshal(n)
			require.NoError(t, err)
			var got Number
			require.NoError(t, json.Unmarshal(j, &got))
			assert.Equal(t, n, got)
		}
	})

	t.Run("empty marshals as null", func(t *testing.T) {
		j, err := json.Marshal(Number(""))
		require.NoError(t, err)
		assert.Equal(t, `null`, string(j))
	})

	t.Run("unmarshal", func(t *testing.T) {
		tests := []struct {
			name    string
			json    string
			want    Number
			wantErr bool
		}{
			// Some systems export account numbers as JSON numbers
			{name: "JSON number", json: `1234567`, want: "1234567"},
			{name: "trimmed string", json: `" 123 "`, want: "123"},
			{name: "negative JSON number", json: `-1`, wantErr: true},
			{name: "fractional JSON number", json: `1.5`, wantErr: true},
			{name: "invalid string", json: `"#1"`, wantErr: true},
			{name: "empty string", json: `""`, wantErr: true},
			{name: "null", json: `null`, wantErr: true},
			{name: "object", json: `{}`, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var n Number
				err := json.Unmarshal([]byte(tt.json), &n)
				if tt.wantErr {
					require.ErrorIs(t, err, ErrInvalidNumber)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.want, n)
			})
		}
	})
}

func TestNumber_JSONSchema(t *testing.T) {
	s := Number("").JSONSchema()
	assert.Equal(t, "string", s.Type)
	assert.Equal(t, DefaultNumberRegex, s.Pattern)

	// NumberRegexp is reassignable for application specific formats,
	// the schema must publish the pattern that Validate actually uses,
	// else clients validating against the schema accept numbers that
	// the server rejects (or the other way round).
	orig := NumberRegexp
	t.Cleanup(func() { NumberRegexp = orig })
	NumberRegexp = regexp.MustCompile(`^[0-9]{6}$`)
	assert.Equal(t, `^[0-9]{6}$`, Number("").JSONSchema().Pattern)
}

func TestNumber_XML(t *testing.T) {
	type doc struct {
		Number Number `xml:"number"`
	}
	x, err := xml.Marshal(doc{Number: "AT-12/34"})
	require.NoError(t, err)
	assert.Equal(t, `<doc><number>AT-12/34</number></doc>`, string(x))

	var d doc
	require.NoError(t, xml.Unmarshal(x, &d))
	assert.Equal(t, Number("AT-12/34"), d.Number)

	require.NoError(t, xml.Unmarshal([]byte(`<doc><number> 123 </number></doc>`), &d))
	assert.Equal(t, Number("123"), d.Number, "XML text must be trimmed")

	// Marshalling trims like Value and MarshalJSON,
	// so every output format writes the same account number
	x, err = xml.Marshal(doc{Number: " 123 "})
	require.NoError(t, err)
	assert.Equal(t, `<doc><number>123</number></doc>`, string(x))

	err = xml.Unmarshal([]byte(`<doc><number>-1</number></doc>`), &d)
	assert.ErrorIs(t, err, ErrInvalidNumber)

	err = xml.Unmarshal([]byte(`<doc><number></number></doc>`), &d)
	assert.ErrorIs(t, err, ErrInvalidNumber, "a non nullable Number can't be empty")

	err = xml.Unmarshal([]byte(`<doc><number>1</doc>`), &d)
	assert.Error(t, err, "XML syntax errors must be propagated")
}
