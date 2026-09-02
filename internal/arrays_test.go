package internal

import (
	"reflect"
	"testing"

	"github.com/domonda/go-types/internal/pq"
)

func TestSplitArray(t *testing.T) {
	tests := []struct {
		name       string
		array      string
		wantFields []string
		wantErr    bool
	}{
		{
			name:       "null",
			array:      `null`,
			wantFields: nil,
		},
		{
			name:       "NULL",
			array:      `NULL`,
			wantFields: nil,
		},
		{
			name:       "empty[]",
			array:      `[]`,
			wantFields: nil,
		},
		{
			name:       "empty{}",
			array:      `{}`,
			wantFields: nil,
		},
		{
			name:       "empty{ }",
			array:      `{ }`,
			wantFields: nil,
		},
		{
			name:       `[a]`,
			array:      `[a]`,
			wantFields: []string{`a`},
		},
		{
			name:       `[a,b]`,
			array:      `[a,b]`,
			wantFields: []string{`a`, `b`},
		},
		{
			name:       `[a, b]`,
			array:      `[a, b]`,
			wantFields: []string{`a`, `b`},
		},
		{
			name:       `["[quoted", "{", "comma,string", "}"]`,
			array:      `["[quoted", "{", "comma,string", "}"]`,
			wantFields: []string{`"[quoted"`, `"{"`, `"comma,string"`, `"}"`},
		},
		{
			name:       `[[1,2,3], {"key": "comma,string"}, null]`,
			array:      `[[1,2,3], {"key": "comma,string"}, null]`,
			wantFields: []string{`[1,2,3]`, `{"key": "comma,string"}`, `null`},
		},
		{
			name:       `{{1,2,3},{4,5,6},{7,8,9}}`,
			array:      `{{1,2,3},{4,5,6},{7,8,9}}`,
			wantFields: []string{`{1,2,3}`, `{4,5,6}`, `{7,8,9}`},
		},
		{
			name:       `{{"meeting", "lunch"}, {"training", "presentation"}}`,
			array:      `{{"meeting", "lunch"}, {"training", "presentation"}}`,
			wantFields: []string{`{"meeting", "lunch"}`, `{"training", "presentation"}`},
		},
		{
			name:       `[['meeting', 'lunch'], ['training', 'presentation']]`,
			array:      `[['meeting', 'lunch'], ['training', 'presentation']]`,
			wantFields: []string{`['meeting', 'lunch']`, `['training', 'presentation']`},
		},
		{
			name:       `[['meeting', 'lunch'], 4, ['training', 'presentation']]`,
			array:      `[['meeting', 'lunch'], 4, ['training', 'presentation']]`,
			wantFields: []string{`['meeting', 'lunch']`, `4`, `['training', 'presentation']`},
		},
		{
			name:       "{bestellungen@example.com,if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com}",
			array:      "{bestellungen@example.com,if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com}",
			wantFields: []string{`bestellungen@example.com`, `if.need.of.a.''declaration.of.compliance''.please.contact.us@example.com`},
		},
		{
			name:       `["single ' quote", "within double quotes"]`,
			array:      `["single ' quote", "within double quotes"]`,
			wantFields: []string{`"single ' quote"`, `"within double quotes"`},
		},
		{
			name:       `{"single ' quote", "within double quotes"}`,
			array:      `{"single ' quote", "within double quotes"}`,
			wantFields: []string{`"single ' quote"`, `"within double quotes"`},
		},
		{
			name:       `{'double " quote', 'within single quotes'}`,
			array:      `{'double " quote', 'within single quotes'}`,
			wantFields: []string{`'double " quote'`, `'within single quotes'`},
		},
		{
			name:       "{single_quote'@example.com}",
			array:      "{single_quote'@example.com}",
			wantFields: []string{`single_quote'@example.com`},
		},
		{
			name:       "{'@example.com,service.wien@example.com,zaehler.wien@example.com}",
			array:      "{'@example.com,service.wien@example.com,zaehler.wien@example.com}",
			wantFields: []string{`'@example.com`, `service.wien@example.com`, `zaehler.wien@example.com`},
		},

		{
			// The escaped quote is not the end of the element
			name:       `{"a\"b"}`,
			array:      `{"a\"b"}`,
			wantFields: []string{`"a\"b"`},
		},
		{
			// An escaped backslash does not escape the closing quote,
			// this is what PostgreSQL outputs for the value `a\`
			name:       `{"a\\"}`,
			array:      `{"a\\"}`,
			wantFields: []string{`"a\\"`},
		},
		{
			name:       `{"a\\","b"}`,
			array:      `{"a\\","b"}`,
			wantFields: []string{`"a\\"`, `"b"`},
		},

		{
			// Backslash escaping applies only to double quoted elements,
			// a backslash in a single quoted element is a literal character
			name:       `['a\','b']`,
			array:      `['a\','b']`,
			wantFields: []string{`'a\'`, `'b'`},
		},
		{
			name:       `['a','b']`,
			array:      `['a','b']`,
			wantFields: []string{`'a'`, `'b'`},
		},

		// Invalid
		{
			name:    "empty",
			array:   ``,
			wantErr: true,
		},
		{
			name:    `empty ""`,
			array:   `""`,
			wantErr: true,
		},
		{
			name:    "empty elements {,}",
			array:   `{,}`,
			wantErr: true,
		},
		{
			name:    "empty elements {, ,}",
			array:   `{, ,}`,
			wantErr: true,
		},
		{
			name:    `e{}`,
			array:   `e{}`,
			wantErr: true,
		},
		{
			name:    `,{}`,
			array:   `,{}`,
			wantErr: true,
		},
		{
			name:    ` [a, b] `,
			array:   ` [a, b] `,
			wantErr: true,
		},
		{
			// The escape tracking must still report a genuinely
			// unclosed quote, not only stop reporting the closed ones
			name:    `{"abc}`,
			array:   `{"abc}`,
			wantErr: true,
		},
		{
			// The backslash escapes the closing quote, so this element
			// really is unclosed
			name:    `{"a\"}`,
			array:   `{"a\"}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFields, err := SplitArray(tt.array)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFields, tt.wantFields) {
				t.Errorf("SplitArray() = %#v, want %#v", gotFields, tt.wantFields)
			}
		})
	}
}

func TestSQLArrayLiteral(t *testing.T) {
	tests := []struct {
		name string
		s    []string
		want string
	}{
		{name: "nil", s: nil, want: `NULL`},
		{name: "empty", s: []string{}, want: `{}`},
		{name: "one", s: []string{`one`}, want: `{"one"}`},
		{name: "two", s: []string{`one`, `two`}, want: `{"one","two"}`},
		{name: "quoted", s: []string{`Hello "World"`}, want: `{"Hello \"World\""}`},
		// A backslash has to be escaped too, PostgreSQL reads {"a\b"}
		// back as the value `ab` and would silently lose it otherwise.
		{name: "backslash", s: []string{`a\b`}, want: `{"a\\b"}`},
		{name: "trailing backslash", s: []string{`a\`}, want: `{"a\\"}`},
		{name: "escaped quote", s: []string{`a\"b`}, want: `{"a\\\"b"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SQLArrayLiteral(tt.s)
			if got != tt.want {
				t.Errorf("SQLArrayLiteral() = %v, want %v", got, tt.want)
			}
			if tt.s != nil {
				val, _ := pq.StringArray(tt.s).Value()
				if val.(string) != got {
					t.Errorf("pq.StringArray() = %v, SQLArrayLiteral() = %v", val, got)
				}
			}
		})
	}
}

// TestSQLArrayLiteralRoundTrip proves that every value survives being
// written as an SQL array literal and read back again. SQLArrayLiteral
// used to escape only '"' and not '\', which made PostgreSQL drop the
// backslashes of a value on reading it back.
func TestSQLArrayLiteralRoundTrip(t *testing.T) {
	// Values that PostgreSQL has to quote and/or escape in its array
	// output literal, plus values that are quote forcing for the writer
	// only. The production data loss that motivated these tests came
	// from elements containing '{' or '}'.
	roundTripValues := []string{
		`plain`,
		``,
		`a\b`,
		`a\`,
		`\`,
		`\\`,
		`a"b`,
		`"quoted"`,
		`"a,b"@example.com`, // valid RFC 5322 address with a quoted local part
		`in{o@example.com`,  // the shape that got corrupted in production
		`comma,separated`,
		`{braces}`,
		`  leading and trailing  `,
		"tab\tand\nnewline",
		`NULL`,
		`日本語`,
	}
	literal := SQLArrayLiteral(roundTripValues)

	t.Run("pq.StringArray.Value parity", func(t *testing.T) {
		// The writer has to agree with the one of the vendored lib/pq
		// for every value, not only for the simple ones, else the two
		// codecs of this module would write different literals.
		val, err := pq.StringArray(roundTripValues).Value()
		if err != nil {
			t.Fatalf("pq.StringArray.Value() error = %v", err)
		}
		if val.(string) != literal {
			t.Errorf("SQLArrayLiteral() = %q, pq.StringArray.Value() = %q", literal, val)
		}
	})

	t.Run("pq.StringArray.Scan", func(t *testing.T) {
		// The vendored lib/pq parser implements the PostgreSQL array
		// syntax and is what notnull.StringArray.Scan uses.
		var scanned pq.StringArray
		if err := scanned.Scan(literal); err != nil {
			t.Fatalf("pq.StringArray.Scan(%q) error = %v", literal, err)
		}
		if !reflect.DeepEqual([]string(scanned), roundTripValues) {
			t.Errorf("pq.StringArray.Scan(%q) = %#v, want %#v", literal, scanned, roundTripValues)
		}
	})

	t.Run("SplitArrayValues", func(t *testing.T) {
		values, err := SplitArrayValues(literal)
		if err != nil {
			t.Fatalf("SplitArrayValues(%q) error = %v", literal, err)
		}
		if !reflect.DeepEqual(values, roundTripValues) {
			t.Errorf("SplitArrayValues(%q) = %#v, want %#v", literal, values, roundTripValues)
		}
	})
}

func TestSplitArrayValues(t *testing.T) {
	tests := []struct {
		name       string
		array      string
		wantValues []string
		wantErr    bool
	}{
		{
			name:       "null",
			array:      `null`,
			wantValues: nil,
		},
		{
			name:       "empty",
			array:      `{}`,
			wantValues: nil,
		},
		{
			name:       "unquoted SQL elements",
			array:      `{a,b}`,
			wantValues: []string{`a`, `b`},
		},
		{
			// PostgreSQL quotes only the elements that need it,
			// so a single array mixes both forms.
			name:       "mixed quoted and unquoted SQL elements",
			array:      `{a,"in{o@example.com",b}`,
			wantValues: []string{`a`, `in{o@example.com`, `b`},
		},
		{
			name:       "SQL quote escape",
			array:      `{"\"a,b\"@example.com"}`,
			wantValues: []string{`"a,b"@example.com`},
		},
		{
			name:       "SQL backslash escape",
			array:      `{"a\\b","a\\"}`,
			wantValues: []string{`a\b`, `a\`},
		},
		{
			// A backslash in a PostgreSQL array literal escapes the
			// following character whatever it is, so this is not a
			// newline but the letter n.
			name:       "SQL backslash escapes any character",
			array:      `{"a\nb"}`,
			wantValues: []string{`anb`},
		},
		{
			// JSON on the other hand interprets the escape sequences.
			name:       "JSON escapes",
			array:      `["a\nb", "a\tb", "ä", "a\\b", "a\"b"]`,
			wantValues: []string{"a\nb", "a\tb", "ä", `a\b`, `a"b`},
		},
		{
			name:       "JSON objects are not touched",
			array:      `[{"a":1}, {"b":"x,y"}, "str", 3, null]`,
			wantValues: []string{`{"a":1}`, `{"b":"x,y"}`, `str`, `3`, `null`},
		},
		{
			// Neither valid SQL nor valid JSON array syntax,
			// documented as being returned unchanged.
			name:       "single quoted elements are not unquoted",
			array:      `{'a','b'}`,
			wantValues: []string{`'a'`, `'b'`},
		},
		{
			// Documented limitation: a SQL NULL element and the string
			// "NULL" are the same value here, only SplitArray keeps
			// the quotes that tell them apart.
			name:       "SQL NULL is indistinguishable from the string NULL",
			array:      `{a,NULL,"NULL"}`,
			wantValues: []string{`a`, `NULL`, `NULL`},
		},

		{
			// An element that is not a valid JSON string is returned
			// unchanged, quotes included, so that it fails visibly
			// downstream instead of being half unescaped
			name:       "invalid JSON escape",
			array:      `["a\q"]`,
			wantValues: []string{`"a\q"`},
		},

		{
			name:    "not an array",
			array:   `nope`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValues, err := SplitArrayValues(tt.array)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitArrayValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("SplitArrayValues() = %#v, want %#v", gotValues, tt.wantValues)
			}
		})
	}
}

// TestUnquoteArrayElemDefensive covers the guards that SplitArray itself
// can never trigger, because it only ever hands over an element whose
// quotes are balanced. They exist so that a direct caller, or a future
// change to the splitter, can't make the unquoting read past the end of
// the element or drop a character.
func TestUnquoteArrayElemDefensive(t *testing.T) {
	tests := []struct {
		name       string
		elem       string
		jsonSyntax bool
		want       string
	}{
		{name: "too short", elem: `"`, want: `"`},
		{name: "empty", elem: ``, want: ``},
		{name: "opening quote only", elem: `"a`, want: `"a`},
		{name: "closing quote only", elem: `a"`, want: `a"`},
		{name: "too short JSON", elem: `"`, jsonSyntax: true, want: `"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquoteArrayElem(tt.elem, tt.jsonSyntax)
			if got != tt.want {
				t.Errorf("unquoteArrayElem(%q, %v) = %q, want %q", tt.elem, tt.jsonSyntax, got, tt.want)
			}
		})
	}
}

func TestUnescapeSQLArrayElem(t *testing.T) {
	tests := []struct {
		name string
		elem string
		want string
	}{
		{name: "no escape", elem: `abc`, want: `abc`},
		{name: "escaped quote", elem: `a\"b`, want: `a"b`},
		{name: "escaped backslash", elem: `a\\b`, want: `a\b`},
		{name: "escapes any character", elem: `a\nb`, want: `anb`},
		// A trailing backslash can't reach here through SplitArray,
		// which would have reported an unclosed quote for it, so this
		// only pins that the guard keeps the last byte and doesn't read
		// past the end.
		{name: "trailing lone backslash", elem: `a\`, want: `a\`},
		{name: "lone backslash", elem: `\`, want: `\`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeSQLArrayElem(tt.elem)
			if got != tt.want {
				t.Errorf("unescapeSQLArrayElem(%q) = %q, want %q", tt.elem, got, tt.want)
			}
		})
	}
}
