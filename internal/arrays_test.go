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
