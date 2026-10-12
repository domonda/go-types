package nullable

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TrimmedStringf(t *testing.T) {
	assert.Equal(t, TrimmedString("a=1"), TrimmedStringf("  a=%d ", 1))
	assert.Equal(t, TrimmedStringNull, TrimmedStringf("  %s ", ""))
}

func Test_TrimmedStringFrom(t *testing.T) {
	assert.Equal(t, TrimmedString("hello"), TrimmedStringFrom("  hello \n"))
	assert.Equal(t, TrimmedStringNull, TrimmedStringFrom("   "))
}

func Test_TrimmedStringFromPtr(t *testing.T) {
	assert.Equal(t, TrimmedStringNull, TrimmedStringFromPtr(nil))
	s := "  x "
	assert.Equal(t, TrimmedString("x"), TrimmedStringFromPtr(&s))
}

func Test_TrimmedStringFromError(t *testing.T) {
	assert.Equal(t, TrimmedStringNull, TrimmedStringFromError(nil))
	assert.Equal(t, TrimmedString("boom"), TrimmedStringFromError(errors.New("  boom ")))
}

func Test_TrimmedStringJoin(t *testing.T) {
	assert.Equal(t, TrimmedString("a,b,c"), TrimmedStringJoin(",", "a", " b ", "c"))
	assert.Equal(t, TrimmedString("a,c"), TrimmedStringJoin(",", "a", "  ", "c"))
	assert.Equal(t, TrimmedStringNull, TrimmedStringJoin(",", " ", ""))
}

func Test_TrimmedString_Ptr(t *testing.T) {
	assert.Nil(t, TrimmedStringNull.Ptr())
	p := TrimmedString("x").Ptr()
	require.NotNil(t, p)
	assert.Equal(t, "x", *p)
}

func Test_TrimmedString_IsNotNull(t *testing.T) {
	assert.False(t, TrimmedString("  ").IsNotNull())
	assert.True(t, TrimmedString("x").IsNotNull())
}

func Test_TrimmedString_StringOr(t *testing.T) {
	assert.Equal(t, "fallback", TrimmedString("  ").StringOr("fallback"))
	assert.Equal(t, "x", TrimmedString(" x ").StringOr("fallback"))
}

func Test_TrimmedString_ToValidUTF8(t *testing.T) {
	assert.Equal(t, TrimmedString("ab"), TrimmedString("a\xffb").ToValidUTF8(""))
}

func Test_TrimmedString_ToUpperLower(t *testing.T) {
	assert.Equal(t, TrimmedString("HELLO"), TrimmedString(" hello ").ToUpper())
	assert.Equal(t, TrimmedString("hello"), TrimmedString(" HELLO ").ToLower())
}

func Test_TrimmedString_Contains(t *testing.T) {
	s := TrimmedString(" hello world ")
	assert.True(t, s.Contains("lo wo"))
	assert.False(t, s.Contains("xyz"))
	assert.True(t, s.ContainsAny("xz o"))
	assert.False(t, s.ContainsAny("xyz"))
	assert.True(t, s.ContainsRune('h'))
	assert.False(t, s.ContainsRune('q'))
}

func Test_TrimmedString_PrefixSuffix(t *testing.T) {
	s := TrimmedString(" hello world ")
	assert.True(t, s.HasPrefix("hello"))
	assert.False(t, s.HasPrefix("world"))
	assert.True(t, s.HasSuffix("world"))
	assert.False(t, s.HasSuffix("hello"))
	assert.Equal(t, TrimmedString("world"), s.TrimPrefix("hello "))
	assert.Equal(t, TrimmedString("hello"), s.TrimSuffix(" world"))
}

func Test_TrimmedString_ReplaceAll(t *testing.T) {
	assert.Equal(t, TrimmedString("a-b-c"), TrimmedString(" a b c ").ReplaceAll(" ", "-"))
}

func Test_TrimmedString_Split(t *testing.T) {
	got := TrimmedString(" a , b , c ").Split(",")
	assert.Equal(t, []TrimmedString{"a", "b", "c"}, got)
}

func Test_TrimmedString_Get(t *testing.T) {
	assert.Panics(t, func() { TrimmedString("  ").Get() })
	assert.Equal(t, "x", TrimmedString(" x ").Get())
}

func Test_TrimmedString_SetAndSetNull(t *testing.T) {
	var s TrimmedString
	s.Set("  hello  ")
	assert.Equal(t, TrimmedString("hello"), s)
	s.SetNull()
	assert.True(t, s.IsNull())
}

func Test_TrimmedString_Value(t *testing.T) {
	val, err := TrimmedString("  ").Value()
	require.NoError(t, err)
	assert.Nil(t, val, "blank TrimmedString returns SQL NULL")

	val, err = TrimmedString(" x ").Value()
	require.NoError(t, err)
	assert.Equal(t, "x", val)
}

func Test_TrimmedString_Scan(t *testing.T) {
	var s TrimmedString

	require.NoError(t, s.Scan(nil))
	assert.True(t, s.IsNull())

	require.NoError(t, s.Scan("  hello  "))
	assert.Equal(t, TrimmedString("hello"), s)

	require.NoError(t, s.Scan([]byte("  bytes  ")))
	assert.Equal(t, TrimmedString("bytes"), s)

	assert.Error(t, s.Scan(123))
}

func Test_TrimmedString_MarshalText(t *testing.T) {
	b, err := TrimmedString("  ").MarshalText()
	require.NoError(t, err)
	assert.Nil(t, b)

	b, err = TrimmedString(" x ").MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "x", string(b))
}

func Test_TrimmedString_UnmarshalText(t *testing.T) {
	var s TrimmedString
	require.NoError(t, s.UnmarshalText([]byte("  hello  ")))
	assert.Equal(t, TrimmedString("hello"), s)
}

func Test_TrimmedString_UnmarshalJSON_Error(t *testing.T) {
	var s TrimmedString
	assert.Error(t, s.UnmarshalJSON([]byte(`{not a string}`)))
}

func Test_TrimmedString_JSONSchema(t *testing.T) {
	schema := TrimmedString("").JSONSchema()
	require.NotNil(t, schema)
	assert.Equal(t, "Nullable Trimmed String", schema.Title)
	require.Len(t, schema.OneOf, 2)
	assert.Equal(t, "string", schema.OneOf[0].Type)
	assert.Equal(t, "null", schema.OneOf[1].Type)
}

func Test_TrimmedString_XML(t *testing.T) {
	type wrapper struct {
		S TrimmedString
	}
	out, err := xml.Marshal(wrapper{S: " hello "})
	require.NoError(t, err)
	assert.Equal(t, "<wrapper><S>hello</S></wrapper>", string(out))

	var w wrapper
	require.NoError(t, xml.Unmarshal([]byte("<wrapper><S>  world  </S></wrapper>"), &w))
	assert.Equal(t, TrimmedString("world"), w.S)

	assert.Error(t, xml.Unmarshal([]byte("<wrapper><S>unclosed"), &w))
}

func Test_TrimmedString_RoundTrip_SQL(t *testing.T) {
	original := TrimmedString("round-trip")
	val, err := original.Value()
	require.NoError(t, err)
	var scanned TrimmedString
	require.NoError(t, scanned.Scan(val))
	assert.Equal(t, original, scanned)
}

func TestTrimmedString_MarshalJSON(t *testing.T) {
	type Struct struct {
		Null     TrimmedString
		NullOmit TrimmedString `json:",omitempty"`
		NonNull  TrimmedString
	}
	s := Struct{NonNull: "\tHello \n"}

	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, `{"Null":null,"NonNull":"Hello"}`, string(j))
}

func TestTrimmedString_UnmarshalJSON(t *testing.T) {
	type Struct struct {
		Null     TrimmedString
		NullOmit TrimmedString `json:",omitempty"`
		Empty    TrimmedString
		NonNull  TrimmedString
	}
	input := `{
		"Null": null,
		"NullOmit": " here ",
		"Empty": "",
		"NonNull": "Hello   "
	}`
	expected := Struct{NullOmit: "here", NonNull: "Hello"}
	var result Struct

	err := json.Unmarshal([]byte(input), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTrimmedString_IsNull(t *testing.T) {
	tests := []struct {
		s    TrimmedString
		want bool
	}{
		{s: "", want: true},
		{s: " ", want: true},
		{s: " \n\t", want: true},
		{s: "a", want: false},
		{s: "NULL", want: false},
		{s: "  NULL  ", want: false},
		{s: "null", want: false},
		{s: "nil", want: false},
		{s: "<nil>", want: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.s), func(t *testing.T) {
			if got := tt.s.IsNull(); got != tt.want {
				t.Errorf("TrimmedString(%#v).IsNull = %#v, want %#v", tt.s, got, tt.want)
			}
		})
	}
}

// A TrimmedString whose trimmed form is empty can't hold a value,
// so IsNull must report it as null and every encoding must write
// the null value. Otherwise a string of only runes that TrimSpace
// removes but IsSpace doesn't report, like the BOM, a zero width
// joiner or an invalid UTF-8 byte, was not null but wrote the empty
// string "" to SQL and JSON instead of NULL and null.
func TestTrimmedString_IsNullAgreesWithTrimmedValue(t *testing.T) {
	for _, s := range []TrimmedString{
		"\xef\xbb\xbf",                 // BOM U+FEFF
		"\xe2\x80\x8d",                 // ZERO WIDTH JOINER U+200D
		"\xe2\x81\xa0",                 // WORD JOINER U+2060
		"\xff",                         // invalid UTF-8
		"\x00\x7f",                     // control characters
		" \xef\xbb\xbf\t\xe2\x80\x8b ", // mixed with whitespace and ZERO WIDTH SPACE
	} {
		t.Run(string(s), func(t *testing.T) {
			require.Equal(t, "", s.String(), "precondition: trimmed form is empty")
			assert.True(t, s.IsNull(), "IsNull")
			assert.False(t, s.IsNotNull(), "IsNotNull")
			assert.Nil(t, s.Ptr(), "Ptr")
			assert.Equal(t, "default", s.StringOr("default"), "StringOr")

			val, err := s.Value()
			require.NoError(t, err)
			assert.Nil(t, val, "Value must return SQL NULL")

			j, err := s.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, "null", string(j), "MarshalJSON must return JSON null")

			text, err := s.MarshalText()
			require.NoError(t, err)
			assert.Nil(t, text, "MarshalText")
		})
	}

	// A value surrounded by such runes is not null and written trimmed
	s := TrimmedString("\xef\xbb\xbfvalue\xe2\x80\x8d")
	assert.False(t, s.IsNull())
	val, err := s.Value()
	require.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestTrimmedString_String(t *testing.T) {
	assert.False(t, unicode.IsSpace('\u200b'), "unicode.IsSpace does not report ZERO WIDTH SPACE")
	tests := []struct {
		s    TrimmedString
		want string
	}{
		{s: "\u200bZERO WIDTH SPACE\r\n", want: "ZERO WIDTH SPACE"},
	}
	for _, tt := range tests {
		t.Run(string(tt.s), func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("TrimmedString(%#v).String() = %#v, want %#v", tt.s, got, tt.want)
			}
		})
	}
}
