package pq

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests in this file focus on BoolArray and StringArray because
// those are the vendored types used by nullable.BoolArray,
// nullable.StringArray and notnull.StringArray.
// Parts of the test cases are ported from github.com/lib/pq's array_test.go.

// TestParseArray checks the PostgreSQL array text format parser that
// BoolArray and StringArray scanning is built on. The format is the one
// emitted by the PostgreSQL backend: an unquoted NULL is SQL NULL (nil),
// quoted elements can contain delimiters, braces and backslash escapes.
func TestParseArray(t *testing.T) {
	tests := []struct {
		input string
		delim string
		dims  []int
		elems [][]byte
	}{
		{input: `{}`, delim: `,`, dims: nil, elems: [][]byte{}},
		{input: `{NULL}`, delim: `,`, dims: []int{1}, elems: [][]byte{nil}},
		{input: `{a}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte("a")}},
		{input: `{a,b}`, delim: `,`, dims: []int{2}, elems: [][]byte{[]byte("a"), []byte("b")}},
		{input: `{a,NULL,b}`, delim: `,`, dims: []int{3}, elems: [][]byte{[]byte("a"), nil, []byte("b")}},
		// A quoted "NULL" is the string NULL, not SQL NULL
		{input: `{"NULL"}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte("NULL")}},
		// Only the upper case unquoted NULL emitted by the backend is SQL NULL
		{input: `{null}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte("null")}},
		// The empty string must be quoted and is not nil
		{input: `{""}`, delim: `,`, dims: []int{1}, elems: [][]byte{{}}},
		{input: `{","}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte(",")}},
		{input: `{",",","}`, delim: `,`, dims: []int{2}, elems: [][]byte{[]byte(","), []byte(",")}},
		{input: `{"\"}"}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte(`"}`)}},
		{input: `{"\"","\""}`, delim: `,`, dims: []int{2}, elems: [][]byte{[]byte(`"`), []byte(`"`)}},
		{input: `{"a\\b"}`, delim: `,`, dims: []int{1}, elems: [][]byte{[]byte(`a\b`)}},
		{input: `{"a b","{c}"}`, delim: `,`, dims: []int{2}, elems: [][]byte{[]byte("a b"), []byte("{c}")}},
		{input: `{{a,b}}`, delim: `,`, dims: []int{1, 2}, elems: [][]byte{[]byte("a"), []byte("b")}},
		{input: `{{a},{b}}`, delim: `,`, dims: []int{2, 1}, elems: [][]byte{[]byte("a"), []byte("b")}},
		{input: `{{{a,b},{c,d},{e,f}}}`, delim: `,`, dims: []int{1, 3, 2}, elems: [][]byte{
			[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"), []byte("f"),
		}},
		{input: `{{"\"","\""}}`, delim: `,`, dims: []int{1, 2}, elems: [][]byte{[]byte(`"`), []byte(`"`)}},
		// Multi byte delimiters like the ';' of the box type are supported
		{input: `{axyzb}`, delim: `xyz`, dims: []int{2}, elems: [][]byte{[]byte("a"), []byte("b")}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			dims, elems, err := parseArray([]byte(tt.input), []byte(tt.delim))
			require.NoError(t, err)
			assert.Equal(t, tt.dims, dims, "dims")
			assert.Equal(t, tt.elems, elems, "elems")
		})
	}
}

// TestParseArrayError checks that malformed array literals are rejected
// with an error naming the offset instead of returning partial elements.
func TestParseArrayError(t *testing.T) {
	tests := []struct {
		input, err string
	}{
		{input: ``, err: "expected '{' at offset 0"},
		{input: `x`, err: "expected '{' at offset 0"},
		{input: `}`, err: "expected '{' at offset 0"},
		{input: `{`, err: "expected '}' at offset 1"},
		{input: `{{}`, err: "expected '}' at offset 3"},
		{input: `{}}`, err: "unexpected '}' at offset 2"},
		{input: `{,}`, err: "unexpected ',' at offset 1"},
		{input: `{,x}`, err: "unexpected ',' at offset 1"},
		{input: `{x,}`, err: "unexpected '}' at offset 3"},
		{input: `{x,{`, err: "unexpected '{' at offset 3"},
		{input: `{x},`, err: "unexpected ',' at offset 3"},
		{input: `{x}}`, err: "unexpected '}' at offset 3"},
		{input: `{{x}`, err: "expected '}' at offset 4"},
		{input: `{""x}`, err: "unexpected 'x' at offset 3"},
		{input: `{"abc}`, err: "expected '}' at offset 6"},
		{input: `{{a},{b,c}}`, err: "multidimensional arrays must have elements with matching dimensions"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, _, err := parseArray([]byte(tt.input), []byte{','})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

// TestBoolArray_Scan checks that PostgreSQL boolean array literals
// are scanned from both string and []byte sources, the two types
// database drivers deliver text values as.
func TestBoolArray_Scan(t *testing.T) {
	tests := []struct {
		src  string
		want BoolArray
	}{
		{src: `{}`, want: BoolArray{}},
		{src: `{t}`, want: BoolArray{true}},
		{src: `{f}`, want: BoolArray{false}},
		{src: `{f,t,t,f}`, want: BoolArray{false, true, true, false}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			arr := BoolArray{true, true, true}
			require.NoError(t, arr.Scan(tt.src), "string source")
			assert.Equal(t, tt.want, arr, "string source")

			arr = BoolArray{true, true, true}
			require.NoError(t, arr.Scan([]byte(tt.src)), "[]byte source")
			assert.Equal(t, tt.want, arr, "[]byte source")
		})
	}
}

// TestBoolArray_ScanNullAndEmpty checks the mapping that
// nullable.BoolArray documents: SQL NULL scans to a nil slice,
// the empty array {} to a non-nil empty slice, so the two stay
// distinguishable after a round-trip through the database.
func TestBoolArray_ScanNullAndEmpty(t *testing.T) {
	arr := BoolArray{true}
	require.NoError(t, arr.Scan(nil))
	assert.Nil(t, arr, "SQL NULL must scan to a nil slice")

	var empty BoolArray
	require.NoError(t, empty.Scan(`{}`))
	assert.NotNil(t, empty, "{} must scan to a non-nil slice")
	assert.Empty(t, empty)

	reused := BoolArray{true, false}
	require.NoError(t, reused.Scan(`{}`))
	assert.NotNil(t, reused, "{} must scan to a non-nil slice")
	assert.Empty(t, reused)
}

// TestBoolArray_ScanError checks that invalid input is rejected without
// modifying the destination. A []bool has no way to represent a NULL
// element, so that must be an error instead of silently becoming false.
func TestBoolArray_ScanError(t *testing.T) {
	tests := []struct {
		src any
		err string
	}{
		{src: 1, err: "cannot convert int to BoolArray"},
		{src: ``, err: "unable to parse array"},
		{src: `{`, err: "unable to parse array"},
		{src: `{{t},{f}}`, err: "cannot convert ARRAY[2][1] to BoolArray"},
		{src: `{NULL}`, err: `could not parse boolean array index 0: invalid boolean ""`},
		{src: `{t,NULL}`, err: `could not parse boolean array index 1: invalid boolean ""`},
		{src: `{true}`, err: `could not parse boolean array index 0: invalid boolean "true"`},
		{src: `{a}`, err: `could not parse boolean array index 0: invalid boolean "a"`},
		{src: `{t,f,cd}`, err: `could not parse boolean array index 2: invalid boolean "cd"`},
	}
	for _, tt := range tests {
		arr := BoolArray{true, true, true}
		err := arr.Scan(tt.src)
		require.Error(t, err, "Scan(%#v)", tt.src)
		assert.Contains(t, err.Error(), tt.err, "Scan(%#v)", tt.src)
		assert.Equal(t, BoolArray{true, true, true}, arr, "Scan(%#v) must not modify the destination", tt.src)
	}
}

// TestBoolArray_Value checks that nil is written as SQL NULL and a
// non-nil empty slice as the empty array {}, the distinction
// nullable.BoolArray documents.
func TestBoolArray_Value(t *testing.T) {
	tests := []struct {
		name string
		arr  BoolArray
		want driver.Value
	}{
		{name: "nil is SQL NULL", arr: nil, want: nil},
		{name: "empty", arr: BoolArray{}, want: `{}`},
		{name: "single", arr: BoolArray{true}, want: `{t}`},
		{name: "multiple", arr: BoolArray{false, true, false}, want: `{f,t,f}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.arr.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBoolArray_ValueScanRoundTrip(t *testing.T) {
	for _, arr := range []BoolArray{nil, {}, {true}, {false, true, false, false}} {
		v, err := arr.Value()
		require.NoError(t, err)
		var scanned BoolArray
		require.NoError(t, scanned.Scan(v))
		assert.Equal(t, arr, scanned, "round-trip of %#v via %#v", arr, v)
	}
}

// TestStringArray_Scan checks parsing of PostgreSQL text array literals
// as emitted by the backend: unquoted simple elements, quoted elements
// containing delimiters, spaces or braces, and backslash escapes.
func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		src  string
		want StringArray
	}{
		{src: `{}`, want: StringArray{}},
		{src: `{t}`, want: StringArray{"t"}},
		{src: `{f,1}`, want: StringArray{"f", "1"}},
		{src: `{"a\\b","c d",","}`, want: StringArray{`a\b`, "c d", ","}},
		{src: `{"say \"hi\"","{x}"}`, want: StringArray{`say "hi"`, "{x}"}},
		{src: `{"",a}`, want: StringArray{"", "a"}},
		// A quoted NULL is the string "NULL", not SQL NULL
		{src: `{"NULL",null}`, want: StringArray{"NULL", "null"}},
		{src: `{äöü,日本語}`, want: StringArray{"äöü", "日本語"}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			arr := StringArray{"x", "x", "x"}
			require.NoError(t, arr.Scan(tt.src), "string source")
			assert.Equal(t, tt.want, arr, "string source")

			arr = StringArray{"x", "x", "x"}
			require.NoError(t, arr.Scan([]byte(tt.src)), "[]byte source")
			assert.Equal(t, tt.want, arr, "[]byte source")
		})
	}
}

// TestStringArray_ScanNullAndEmpty checks the mapping that
// nullable.StringArray documents: SQL NULL scans to a nil slice,
// the empty array {} to a non-nil empty slice.
func TestStringArray_ScanNullAndEmpty(t *testing.T) {
	arr := StringArray{"x"}
	require.NoError(t, arr.Scan(nil))
	assert.Nil(t, arr, "SQL NULL must scan to a nil slice")

	var empty StringArray
	require.NoError(t, empty.Scan(`{}`))
	assert.NotNil(t, empty, "{} must scan to a non-nil slice")
	assert.Empty(t, empty)

	reused := StringArray{"a", "b"}
	require.NoError(t, reused.Scan([]byte(`{}`)))
	assert.NotNil(t, reused, "{} must scan to a non-nil slice")
	assert.Empty(t, reused)
}

// TestStringArray_ScanError checks that invalid input is rejected without
// modifying the destination. A []string can't represent a NULL element,
// so that must be an error instead of silently becoming "" which would be
// indistinguishable from a real empty string.
func TestStringArray_ScanError(t *testing.T) {
	tests := []struct {
		src any
		err string
	}{
		{src: true, err: "cannot convert bool to StringArray"},
		{src: ``, err: "unable to parse array"},
		{src: `{`, err: "unable to parse array"},
		{src: `{"a}`, err: "unable to parse array"},
		{src: `{{a},{b}}`, err: "cannot convert ARRAY[2][1] to StringArray"},
		{src: `{NULL}`, err: "parsing array element index 0: cannot convert nil to string"},
		{src: `{a,NULL}`, err: "parsing array element index 1: cannot convert nil to string"},
		{src: `{a,b,NULL}`, err: "parsing array element index 2: cannot convert nil to string"},
	}
	for _, tt := range tests {
		arr := StringArray{"x", "x", "x"}
		err := arr.Scan(tt.src)
		require.Error(t, err, "Scan(%#v)", tt.src)
		assert.Contains(t, err.Error(), tt.err, "Scan(%#v)", tt.src)
		assert.Equal(t, StringArray{"x", "x", "x"}, arr, "Scan(%#v) must not modify the destination", tt.src)
	}
}

// TestStringArray_Value checks that every element is quoted and that
// double quotes and backslashes are escaped, so no element content can
// be misread by PostgreSQL as a delimiter, a brace, the end of the
// element, or the NULL keyword.
func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name string
		arr  StringArray
		want driver.Value
	}{
		{name: "nil is SQL NULL", arr: nil, want: nil},
		{name: "empty", arr: StringArray{}, want: `{}`},
		{name: "single", arr: StringArray{"a"}, want: `{"a"}`},
		{name: "empty string element", arr: StringArray{""}, want: `{""}`},
		{name: "NULL string is quoted", arr: StringArray{"NULL"}, want: `{"NULL"}`},
		{name: "escaping", arr: StringArray{`a`, `\b`, `c"`, `d,e`}, want: `{"a","\\b","c\"","d,e"}`},
		{name: "braces and spaces", arr: StringArray{"{x}", " y "}, want: `{"{x}"," y "}`},
		{name: "only special chars", arr: StringArray{`"\"`}, want: `{"\"\\\""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.arr.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStringArray_ValueScanRoundTrip checks that every string survives
// the way to the database and back unchanged, including the ones that
// need quoting or escaping.
func TestStringArray_ValueScanRoundTrip(t *testing.T) {
	arrays := []StringArray{
		nil,
		{},
		{""},
		{"a", "b"},
		{"NULL", "null", ""},
		{`a\b`, `say "hi"`, `\`, `"`, `\"`, `\\`},
		{"a,b", "{c}", "d}", "{e", " f ", "\tg\n"},
		{"äöü", "日本語", "😀"},
	}
	for _, arr := range arrays {
		v, err := arr.Value()
		require.NoError(t, err)
		var scanned StringArray
		require.NoError(t, scanned.Scan(v), "Scan(%#v)", v)
		assert.Equal(t, arr, scanned, "round-trip of %#v via %#v", arr, v)
	}
}

func TestAppendArrayQuotedBytes(t *testing.T) {
	tests := []struct {
		prefix string
		v      string
		want   string
	}{
		{prefix: "", v: "", want: `""`},
		{prefix: "", v: "abc", want: `"abc"`},
		{prefix: "{", v: "abc", want: `{"abc"`},
		{prefix: "", v: `"`, want: `"\""`},
		{prefix: "", v: `\`, want: `"\\"`},
		{prefix: "", v: `a"b\c`, want: `"a\"b\\c"`},
		{prefix: "", v: `""\\`, want: `"\"\"\\\\"`},
		// Delimiters and braces need no escaping within quotes
		{prefix: "", v: `,{}`, want: `",{}"`},
	}
	for _, tt := range tests {
		got := appendArrayQuotedBytes([]byte(tt.prefix), []byte(tt.v))
		assert.Equal(t, tt.want, string(got), "appendArrayQuotedBytes(%q, %q)", tt.prefix, tt.v)
	}
}
