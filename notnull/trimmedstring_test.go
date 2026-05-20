package notnull

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TrimmedString_Set(t *testing.T) {
	var s TrimmedString
	s.Set("  hello  ")
	assert.Equal(t, TrimmedString("hello"), s)
}

func Test_TrimmedString_FromAndFmt(t *testing.T) {
	assert.Equal(t, TrimmedString("hi"), TrimmedStringFrom("  hi\t"))
	assert.Equal(t, TrimmedString("a=1"), TrimmedStringf(" a=%d ", 1))
}

func Test_TrimmedString_IsEmpty(t *testing.T) {
	assert.True(t, TrimmedString("").IsEmpty())
	assert.True(t, TrimmedString("   ").IsEmpty())
	assert.False(t, TrimmedString("x").IsEmpty())
	assert.False(t, TrimmedString("  x  ").IsEmpty())
}

func Test_TrimmedString_Value(t *testing.T) {
	v, err := TrimmedString("  abc  ").Value()
	assert.NoError(t, err)
	assert.Equal(t, "abc", v)
}

func Test_TrimmedString_Scan(t *testing.T) {
	var s TrimmedString
	assert.NoError(t, s.Scan("  hi  "))
	assert.Equal(t, TrimmedString("hi"), s)

	assert.NoError(t, s.Scan([]byte("  bye  ")))
	assert.Equal(t, TrimmedString("bye"), s)
}

func Test_TrimmedString_JSON(t *testing.T) {
	b, err := json.Marshal(TrimmedString(" hello "))
	assert.NoError(t, err)
	assert.Equal(t, `"hello"`, string(b))

	var s TrimmedString
	assert.NoError(t, json.Unmarshal([]byte(`" world "`), &s))
	assert.Equal(t, TrimmedString("world"), s)
}

func Test_TrimmedString_Join(t *testing.T) {
	got := TrimmedStringJoin(", ", " a ", "b", "  c  ")
	assert.Equal(t, TrimmedString("a, b, c"), got)
}

func Test_TrimmedString_StringMethods(t *testing.T) {
	s := TrimmedString("  Hello World  ")
	assert.Equal(t, TrimmedString("HELLO WORLD"), s.ToUpper())
	assert.Equal(t, TrimmedString("hello world"), s.ToLower())
	assert.True(t, s.Contains("World"))
	assert.True(t, s.HasPrefix("Hello"))
	assert.True(t, s.HasSuffix("World"))
}

func Test_TrimmedString_String(t *testing.T) {
	assert.Equal(t, "hello", TrimmedString("  hello  ").String())
	assert.Equal(t, "", TrimmedString("   ").String())
	assert.Equal(t, "", TrimmedString("").String())
}

func Test_TrimmedString_ToValidUTF8(t *testing.T) {
	// Valid UTF-8 stays unchanged (also trimmed)
	assert.Equal(t, TrimmedString("hello"), TrimmedString("  hello  ").ToValidUTF8("?"))
	// Invalid UTF-8 byte gets replaced
	invalid := TrimmedString("  ab\xffcd  ")
	assert.Equal(t, TrimmedString("ab?cd"), invalid.ToValidUTF8("?"))
	// Empty replacement drops invalid bytes
	assert.Equal(t, TrimmedString("abcd"), invalid.ToValidUTF8(""))
}

func Test_TrimmedString_ContainsAny(t *testing.T) {
	s := TrimmedString("  Hello World  ")
	assert.True(t, s.ContainsAny("xyzW"))
	assert.False(t, s.ContainsAny("xyz"))
	assert.False(t, s.ContainsAny(""))
}

func Test_TrimmedString_ContainsRune(t *testing.T) {
	s := TrimmedString("  Hello  ")
	assert.True(t, s.ContainsRune('e'))
	assert.False(t, s.ContainsRune('z'))
	// Trimmed whitespace must not be found
	assert.False(t, s.ContainsRune(' '))
}

func Test_TrimmedString_TrimPrefix(t *testing.T) {
	assert.Equal(t, TrimmedString("World"), TrimmedString("  Hello World  ").TrimPrefix("Hello "))
	// No match returns unchanged (but trimmed)
	assert.Equal(t, TrimmedString("Hello World"), TrimmedString("  Hello World  ").TrimPrefix("xyz"))
}

func Test_TrimmedString_TrimSuffix(t *testing.T) {
	assert.Equal(t, TrimmedString("Hello"), TrimmedString("  Hello World  ").TrimSuffix(" World"))
	// No match returns unchanged (but trimmed)
	assert.Equal(t, TrimmedString("Hello World"), TrimmedString("  Hello World  ").TrimSuffix("xyz"))
}

func Test_TrimmedString_ReplaceAll(t *testing.T) {
	assert.Equal(t, TrimmedString("a-b-c"), TrimmedString("  a.b.c  ").ReplaceAll(".", "-"))
	// No match returns unchanged (but trimmed)
	assert.Equal(t, TrimmedString("abc"), TrimmedString("  abc  ").ReplaceAll("x", "y"))
	// Replacement that introduces whitespace at the edges is trimmed away
	assert.Equal(t, TrimmedString("X"), TrimmedString(" a ").ReplaceAll("a", " X "))
}

func Test_TrimmedString_Split(t *testing.T) {
	got := TrimmedString("  a , b , c  ").Split(",")
	assert.Equal(t, []TrimmedString{"a", "b", "c"}, got)

	// No separator present returns single-element slice
	single := TrimmedString("  abc  ").Split(",")
	assert.Equal(t, []TrimmedString{"abc"}, single)

	// Empty separator splits after each UTF-8 sequence
	runes := TrimmedString("abc").Split("")
	assert.Equal(t, []TrimmedString{"a", "b", "c"}, runes)
}

func Test_TrimmedString_Text(t *testing.T) {
	b, err := TrimmedString("  hello  ").MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "hello", string(b))

	var s TrimmedString
	require.NoError(t, s.UnmarshalText([]byte("  world  ")))
	assert.Equal(t, TrimmedString("world"), s)
}

func Test_TrimmedString_UnmarshalJSON_Error(t *testing.T) {
	var s TrimmedString
	err := s.UnmarshalJSON([]byte(`{not valid`))
	assert.Error(t, err, "UnmarshalJSON with invalid JSON")
}

func Test_TrimmedString_XML(t *testing.T) {
	type wrapper struct {
		XMLName xml.Name      `xml:"wrapper"`
		Value   TrimmedString `xml:"value"`
	}

	out, err := xml.Marshal(wrapper{Value: TrimmedString("  hello  ")})
	require.NoError(t, err, "xml.Marshal")
	assert.Equal(t, `<wrapper><value>hello</value></wrapper>`, string(out))

	var w wrapper
	require.NoError(t, xml.Unmarshal([]byte(`<wrapper><value>  world  </value></wrapper>`), &w), "xml.Unmarshal")
	assert.Equal(t, TrimmedString("world"), w.Value)

	// Unmarshal error path: malformed XML
	var bad wrapper
	err = xml.Unmarshal([]byte(`<wrapper><value>oops</wrapper>`), &bad)
	assert.Error(t, err, "xml.Unmarshal with malformed XML")
}

func Test_TrimmedString_Scan_Error(t *testing.T) {
	var s TrimmedString
	err := s.Scan(123)
	assert.Error(t, err, "Scan with unsupported type")
}
