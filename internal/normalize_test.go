package internal

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/domonda/go-errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upperCode normalizes to upper case and rejects "XX" and the empty
// string, like the code types of the domain packages.
type upperCode string

func (c upperCode) Normalized() (upperCode, error) {
	norm := upperCode(strings.ToUpper(strings.TrimSpace(string(c))))
	if norm == "" || norm == "XX" {
		return c, errors.New("invalid upperCode")
	}
	return norm, nil
}

// zeroIsBlank treats "0" as blank in addition to whitespace,
// like date.Date treats the zero dates.
func zeroIsBlank(c upperCode) bool {
	return IsBlank(c) || strings.TrimSpace(string(c)) == "0"
}

// NormalizedOrRaw applies normalize when it succeeds, returns empty for a
// blank value, and keeps the input when normalize fails, so no reading
// fails on content. Trim and TrimUpper always succeed, so they prettify
// a value Normalized would reject.
func TestTrim(t *testing.T) {
	trimmed, err := Trim(upperCode("  de "))
	require.NoError(t, err)
	assert.Equal(t, upperCode("de"), trimmed)
	upper, err := TrimUpper(upperCode("  de "))
	require.NoError(t, err)
	assert.Equal(t, upperCode("DE"), upper)
	kept, err := TrimUpper(upperCode("xx"))
	require.NoError(t, err)
	assert.Equal(t, upperCode("XX"), kept, "prettify does not reject")
}

func TestNormalizedOrRaw(t *testing.T) {
	assert.Equal(t, upperCode("DE"), NormalizedOrRaw(upperCode(" de "), upperCode.Normalized, nil))
	assert.Equal(t, upperCode("xx"), NormalizedOrRaw(upperCode("xx"), upperCode.Normalized, nil), "invalid input unchanged")
	assert.Equal(t, upperCode(""), NormalizedOrRaw(upperCode(" \t"), upperCode.Normalized, nil), "blank is empty")
	assert.Equal(t, upperCode("0"), NormalizedOrRaw(upperCode("0"), upperCode.Normalized, nil), "not blank with the default isBlank")
	assert.Equal(t, upperCode(""), NormalizedOrRaw(upperCode(" 0 "), upperCode.Normalized, zeroIsBlank), "blank for a custom isBlank")
}

func TestScanNormalized(t *testing.T) {
	var c upperCode
	require.NoError(t, ScanNormalized(&c, "de", upperCode.Normalized, nil))
	assert.Equal(t, upperCode("DE"), c)
	require.NoError(t, ScanNormalized(&c, []byte("xx"), upperCode.Normalized, nil))
	assert.Equal(t, upperCode("xx"), c, "an invalid value is kept without error")
	require.NoError(t, ScanNormalized(&c, nil, upperCode.Normalized, nil))
	assert.Equal(t, upperCode(""), c, "SQL NULL is the empty value")

	c = "DE"
	err := ScanNormalized(&c, int64(1), upperCode.Normalized, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal.upperCode", "the error names the type")
	assert.Equal(t, upperCode("DE"), c, "unchanged on error")
}

// ScanString with validate is the strict reading path of strfmt:
// an invalid or, for a non-nullable type, blank value is an error.
func TestScanStringNormalized(t *testing.T) {
	var c upperCode
	require.NoError(t, ScanStringNormalized(&c, "xx", false, false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("xx"), c)
	require.NoError(t, ScanStringNormalized(&c, "de", true, false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("DE"), c)
	assert.Error(t, ScanStringNormalized(&c, "xx", true, false, upperCode.Normalized, nil))
	assert.Error(t, ScanStringNormalized(&c, "", true, false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("DE"), c, "unchanged on error")

	// strfmt writes null as one of these words, which only a nullable
	// type reads as null; for another type they are just text
	for _, null := range []string{"NULL", "null", "nil"} {
		c = "DE"
		require.NoError(t, ScanStringNormalized(&c, null, true, true, upperCode.Normalized, nil))
		assert.Equal(t, upperCode(""), c, "nullable %q", null)
	}
	require.NoError(t, ScanStringNormalized(&c, "nil", true, false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("NIL"), c, "not nullable")
}

// Value is the one writing method that refuses an invalid value,
// because the database is where it would stay.
func TestValueNormalized(t *testing.T) {
	v, err := ValueNormalized(upperCode("de"), upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, driver.Value("DE"), v)

	v, err = ValueNormalized(upperCode("  "), upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Nil(t, v, "blank is SQL NULL")

	v, err = ValueNormalized(upperCode("0"), upperCode.Normalized, zeroIsBlank)
	require.NoError(t, err)
	assert.Nil(t, v, "blank for a custom isBlank is SQL NULL")

	_, err = ValueNormalized(upperCode("xx"), upperCode.Normalized, nil)
	assert.Error(t, err)
}

func TestMarshalNormalized(t *testing.T) {
	data, err := MarshalJSONNormalized(upperCode("de"), false, upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, `"DE"`, string(data))
	data, err = MarshalJSONNormalized(upperCode("xx"), false, upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, `"xx"`, string(data), "an invalid value is written unchanged")
	// A nullable type writes its blank value as JSON null, which reads
	// back as null, another type writes it as "" like a string
	data, err = MarshalJSONNormalized(upperCode(" "), true, upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, `null`, string(data))
	data, err = MarshalJSONNormalized(upperCode(" "), false, upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, `""`, string(data))
	data, err = MarshalJSONNormalized(upperCode("xx"), true, upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, `"xx"`, string(data), "only a blank value is null")
	text, err := MarshalTextNormalized(upperCode(" "), upperCode.Normalized, nil)
	require.NoError(t, err)
	assert.Equal(t, "", string(text))
}

// JSON null is a no-op for a non-nullable type, like encoding/json does
// for a string type, and resets a nullable type, which would else keep an
// old value when a partial update says there is none.
func TestUnmarshalJSONNormalized(t *testing.T) {
	c := upperCode("DE")
	require.NoError(t, UnmarshalJSONNormalized(&c, []byte(`null`), false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("DE"), c)
	require.NoError(t, UnmarshalJSONNormalized(&c, []byte(`null`), true, upperCode.Normalized, nil))
	assert.Equal(t, upperCode(""), c)

	require.NoError(t, UnmarshalJSONNormalized(&c, []byte(`" at "`), false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("AT"), c)
	require.NoError(t, UnmarshalJSONNormalized(&c, []byte(`"xx"`), false, upperCode.Normalized, nil))
	assert.Equal(t, upperCode("xx"), c)

	c = "DE"
	err := UnmarshalJSONNormalized(&c, []byte(`1`), false, upperCode.Normalized, nil)
	var typeErr *json.UnmarshalTypeError
	require.ErrorAs(t, err, &typeErr)
	assert.Contains(t, err.Error(), "internal.upperCode")
	assert.Equal(t, upperCode("DE"), c, "unchanged on error")
}

// plainJSONString skips json.Unmarshal for the common value, so it must
// accept only what decodes to itself and leave everything else, escapes,
// control and non-ASCII bytes and other JSON values, to json.Unmarshal.
func TestPlainJSONString(t *testing.T) {
	for _, data := range []string{
		`""`, `"DE"`, `" a b "`, `"2024-Q1"`, `"~!@#$%^&*()_+{}|:<>?"`,
		`"a\"b"`, `"a\\b"`, `"aA"`, `"a\nb"`, `"tab	in"`, `"ü"`, "\"\x7f\"", "\"\xff\"",
		`"`, `"unterminated`, `x"`, `1`, `null`, `["DE"]`, ` "DE"`, `"DE" `,
	} {
		got, ok := plainJSONString([]byte(data))
		var want string
		err := json.Unmarshal([]byte(data), &want)
		if !ok {
			continue
		}
		require.NoError(t, err, "plainJSONString accepted %q, which json.Unmarshal rejects", data)
		assert.Equal(t, want, got, "plainJSONString(%q)", data)
		assert.NotContains(t, data[1:len(data)-1], `\`, "plainJSONString accepted an escape in %q", data)
	}
	got, ok := plainJSONString([]byte(`"a b"`))
	require.True(t, ok)
	assert.Equal(t, "a b", got)
	for _, data := range []string{`"a\"b"`, `"ü"`, "\"\x7f\"", `1`, `null`} {
		_, ok := plainJSONString([]byte(data))
		assert.False(t, ok, "plainJSONString(%q) must leave it to json.Unmarshal", data)
	}
}

type xmlCode upperCode

func (c *xmlCode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return UnmarshalXMLNormalized((*upperCode)(c), d, start, upperCode.Normalized, nil)
}

func (c *xmlCode) UnmarshalXMLAttr(attr xml.Attr) error {
	return UnmarshalXMLAttrNormalized((*upperCode)(c), attr, upperCode.Normalized, nil)
}

func TestUnmarshalXMLNormalized(t *testing.T) {
	var doc struct {
		XMLName xml.Name `xml:"doc"`
		Attr    xmlCode  `xml:"a,attr"`
		Elem    xmlCode  `xml:"e"`
		Bad     xmlCode  `xml:"b"`
	}
	require.NoError(t, xml.Unmarshal([]byte(`<doc a=" de "><e>at</e><b>xx</b></doc>`), &doc))
	assert.Equal(t, xmlCode("DE"), doc.Attr)
	assert.Equal(t, xmlCode("AT"), doc.Elem)
	assert.Equal(t, xmlCode("xx"), doc.Bad)

	var c xmlCode
	assert.Error(t, xml.Unmarshal([]byte(`<e>unterminated`), &c))
}

// failText is a TextMarshaler that fails for the value "fail".
type failText string

func (f failText) MarshalText() ([]byte, error) {
	if f == "fail" {
		return nil, errs.New("can't marshal")
	}
	return []byte(f), nil
}

// A nil pointer gets no attribute, because encoding/xml would call the
// value receiver MarshalText through it and panic. A value gets the text
// of MarshalText, also "", like encoding/xml writes a TextMarshaler.
func TestMarshalXMLAttrText(t *testing.T) {
	name := xml.Name{Local: "a"}
	attr, err := MarshalXMLAttrText[failText](nil, name)
	require.NoError(t, err)
	assert.Equal(t, xml.Attr{}, attr, "encoding/xml leaves out an attribute without Name")

	for _, v := range []failText{"", "x"} {
		attr, err = MarshalXMLAttrText(&v, name)
		require.NoError(t, err)
		assert.Equal(t, xml.Attr{Name: name, Value: string(v)}, attr)
	}

	fail := failText("fail")
	_, err = MarshalXMLAttrText(&fail, name)
	assert.Error(t, err)
}
