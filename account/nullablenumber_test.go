package account

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullableNumberFrom(t *testing.T) {
	// Unlike NumberFrom, an empty or whitespace only string is NULL, not an error
	tests := []struct {
		name    string
		str     string
		want    NullableNumber
		wantErr bool
	}{
		{name: "empty is NULL", str: "", want: NumberNull},
		{name: "whitespace is NULL", str: " \t\n", want: NumberNull},
		{name: "trimmed", str: " 123/4 ", want: "123/4"},
		{name: "invalid", str: "#1", wantErr: true},
		{name: "must start alphanumeric", str: "_1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NullableNumberFrom(tt.str)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidNumber)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNullableNumberFromUint(t *testing.T) {
	// 0 is interpreted as NULL for the nullable type
	assert.Equal(t, NumberNull, NullableNumberFromUint(0))
	assert.Equal(t, NullableNumber("42"), NullableNumberFromUint(42))
}

func TestNullableNumber_Valid(t *testing.T) {
	assert.True(t, NumberNull.Valid(), "NULL is a valid NullableNumber")
	assert.NoError(t, NumberNull.Validate())
	assert.True(t, NullableNumber("AT-1").Valid())
	assert.NoError(t, NullableNumber("AT-1").Validate())
	assert.False(t, NullableNumber(" ").Valid(), "only the empty string is NULL")
	assert.ErrorIs(t, NullableNumber("-1").Validate(), ErrInvalidNumber)
}

func TestNullableNumber_NullMethods(t *testing.T) {
	var n NullableNumber
	assert.True(t, n.IsNull(), "zero value must be NULL")
	assert.False(t, n.IsNotNull())
	assert.Equal(t, Number("default"), n.GetOr("default"))
	assert.Panics(t, func() { n.Get() }, "Get on NULL must panic instead of returning an invalid Number")

	n.Set("123")
	assert.True(t, n.IsNotNull())
	assert.False(t, n.IsNull())
	assert.Equal(t, Number("123"), n.Get())
	assert.Equal(t, Number("123"), n.GetOr("default"))
	assert.Equal(t, "123", n.String())

	n.SetNull()
	assert.True(t, n.IsNull())
	assert.Equal(t, "", n.String())
}

func TestNullableNumber_StringHelpers(t *testing.T) {
	n := NullableNumber("00AT-123/4")
	assert.True(t, n.HasPrefix("00AT"))
	assert.False(t, n.HasPrefix("DE"))
	assert.True(t, n.HasSuffix("/4"))
	assert.False(t, n.HasSuffix("/5"))
	assert.Equal(t, NullableNumber("AT-123/4"), n.TrimLeadingZeros())
	assert.Equal(t, NumberNull, NullableNumber("000").TrimLeadingZeros(), "all zeros trims to NULL")
	assert.Equal(t, NumberNull, NumberNull.TrimLeadingZeros())

	before, after, found := n.Cut("/")
	assert.True(t, found)
	assert.Equal(t, NullableNumber("00AT-123"), before)
	assert.Equal(t, NullableNumber("4"), after)
	_, _, found = n.Cut("#")
	assert.False(t, found)
}

func TestNullableNumber_Numeric(t *testing.T) {
	assert.True(t, NullableNumber("123").IsNumeric())
	assert.NoError(t, NullableNumber("123").ValidateNumeric())
	assert.False(t, NullableNumber("12a").IsNumeric())
	assert.ErrorIs(t, NullableNumber("12a").ValidateNumeric(), ErrAlphanumericNumber)
	assert.False(t, NumberNull.IsNumeric(), "NULL is not numeric")
	assert.ErrorIs(t, NumberNull.ValidateNumeric(), ErrAlphanumericNumber)
}

func TestNullableNumber_UintInt(t *testing.T) {
	// NULL converts to the zero value without an error
	u, err := NumberNull.Uint()
	require.NoError(t, err)
	assert.Equal(t, uint64(0), u)
	p, err := NumberNull.UintPtr()
	require.NoError(t, err)
	assert.Nil(t, p, "NULL must convert to a nil pointer, not a pointer to 0")
	i, err := NumberNull.Int()
	require.NoError(t, err)
	assert.Equal(t, int64(0), i)

	u, err = NullableNumber("42").Uint()
	require.NoError(t, err)
	assert.Equal(t, uint64(42), u)
	p, err = NullableNumber("42").UintPtr()
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, uint64(42), *p)
	i, err = NullableNumber("42").Int()
	require.NoError(t, err)
	assert.Equal(t, int64(42), i)

	// Non NULL alphanumeric values are still an error
	_, err = NullableNumber("0x10").Uint()
	assert.ErrorIs(t, err, ErrAlphanumericNumber)
	p, err = NullableNumber("0x10").UintPtr()
	assert.ErrorIs(t, err, ErrAlphanumericNumber)
	assert.Nil(t, p)
	_, err = NullableNumber("0x10").Int()
	assert.ErrorIs(t, err, ErrAlphanumericNumber)
}

func TestNullableNumber_UnmarshalText(t *testing.T) {
	n := NullableNumber("123")
	require.NoError(t, n.UnmarshalText([]byte("  ")))
	assert.True(t, n.IsNull(), "empty text must set NULL")

	require.NoError(t, n.UnmarshalText([]byte(" 1/2 ")))
	assert.Equal(t, NullableNumber("1/2"), n)

	err := n.UnmarshalText([]byte("#1"))
	require.ErrorIs(t, err, ErrInvalidNumber)
	assert.Equal(t, NullableNumber("1/2"), n, "failed unmarshalling must not modify the value")
}

func TestNullableNumber_Value(t *testing.T) {
	// NULL must be written to SQL as nil, not as an empty string
	v, err := NumberNull.Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	v, err = NullableNumber("  ").Value()
	require.NoError(t, err)
	assert.Nil(t, v, "whitespace only must also be written as NULL")

	v, err = NullableNumber(" 123 ").Value()
	require.NoError(t, err)
	assert.Equal(t, "123", v)
}

func TestNullableNumber_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    NullableNumber
		wantErr bool
	}{
		{name: "SQL NULL", value: nil, want: NumberNull},
		{name: "empty string is NULL", value: "", want: NumberNull},
		{name: "string", value: " 123-4 ", want: "123-4"},
		{name: "empty bytes is NULL", value: []byte{}, want: NumberNull},
		{name: "bytes", value: []byte("AB/12"), want: "AB/12"},
		// 0 is NULL, consistent with NullableNumberFromUint
		{name: "int64 zero is NULL", value: int64(0), want: NumberNull},
		{name: "int64", value: int64(1234567), want: "1234567"},
		{name: "float64 zero is NULL", value: float64(0), want: NumberNull},
		{name: "float64", value: float64(1234567), want: "1234567"},
		{name: "invalid string", value: "#1", wantErr: true},
		{name: "invalid bytes", value: []byte("-1"), wantErr: true},
		{name: "negative int64", value: int64(-1), wantErr: true},
		{name: "negative float64", value: float64(-0.5), wantErr: true},
		{name: "unsupported type", value: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NullableNumber("unchanged")
			err := n.Scan(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, NullableNumber("unchanged"), n, "failed scan must not modify the value")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestNullableNumber_JSON(t *testing.T) {
	type obj struct {
		N NullableNumber `json:"n"`
	}

	t.Run("NULL round trip", func(t *testing.T) {
		j, err := json.Marshal(obj{N: NumberNull})
		require.NoError(t, err)
		assert.Equal(t, `{"n":null}`, string(j), "NULL must be marshalled as JSON null, not as empty string")

		got := obj{N: "previous"}
		require.NoError(t, json.Unmarshal(j, &got))
		assert.True(t, got.N.IsNull(), "JSON null must overwrite a previous value with NULL")
	})

	t.Run("whitespace marshals as null", func(t *testing.T) {
		j, err := json.Marshal(NullableNumber("  "))
		require.NoError(t, err)
		assert.Equal(t, `null`, string(j))
	})

	t.Run("value round trip", func(t *testing.T) {
		j, err := json.Marshal(obj{N: "AT-12/34"})
		require.NoError(t, err)
		assert.Equal(t, `{"n":"AT-12/34"}`, string(j))
		var got obj
		require.NoError(t, json.Unmarshal(j, &got))
		assert.Equal(t, NullableNumber("AT-12/34"), got.N)
	})

	t.Run("JSON number", func(t *testing.T) {
		var n NullableNumber
		require.NoError(t, json.Unmarshal([]byte(`1234567`), &n))
		assert.Equal(t, NullableNumber("1234567"), n)
	})

	t.Run("invalid", func(t *testing.T) {
		var n NullableNumber
		assert.ErrorIs(t, json.Unmarshal([]byte(`"#1"`), &n), ErrInvalidNumber)
		assert.ErrorIs(t, json.Unmarshal([]byte(`-1`), &n), ErrInvalidNumber)
	})
}

func TestNullableNumber_JSONSchema(t *testing.T) {
	// The schema must accept JSON null because NULL is marshalled as null
	s := NumberNull.JSONSchema()
	require.Len(t, s.OneOf, 2)
	assert.Equal(t, "string", s.OneOf[0].Type)
	assert.Equal(t, DefaultNumberRegex, s.OneOf[0].Pattern)
	assert.Equal(t, "null", s.OneOf[1].Type)
}

func TestNullableNumber_XML(t *testing.T) {
	type doc struct {
		Number NullableNumber `xml:"number"`
	}

	x, err := xml.Marshal(doc{Number: "AT-12/34"})
	require.NoError(t, err)
	assert.Equal(t, `<doc><number>AT-12/34</number></doc>`, string(x))
	var d doc
	require.NoError(t, xml.Unmarshal(x, &d))
	assert.Equal(t, NullableNumber("AT-12/34"), d.Number)

	// NULL round trips through an empty element
	x, err = xml.Marshal(doc{Number: NumberNull})
	require.NoError(t, err)
	assert.Equal(t, `<doc><number></number></doc>`, string(x))
	d.Number = "previous"
	require.NoError(t, xml.Unmarshal(x, &d))
	assert.True(t, d.Number.IsNull())

	err = xml.Unmarshal([]byte(`<doc><number>-1</number></doc>`), &d)
	assert.ErrorIs(t, err, ErrInvalidNumber)

	err = xml.Unmarshal([]byte(`<doc><number>1</doc>`), &d)
	assert.Error(t, err, "XML syntax errors must be propagated")
}
