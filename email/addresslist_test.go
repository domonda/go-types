package email

import (
	"net/mail"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardComparison(t *testing.T) {
	validEmailAddresses := map[string]*mail.Address{
		`"Unger, Erik" <u.erik@domonda.com>`:        {Name: "Unger, Erik", Address: "u.erik@domonda.com"},
		`"Unger, Erik" <"Unger, Erik"@domonda.com>`: {Name: "Unger, Erik", Address: "Unger, Erik@domonda.com"},
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			result, err := mail.ParseAddress(addr)
			assert.NoError(t, err, "valid email address")
			assert.Equal(t, expected, result, "expected: %s", expected)
		})
	}

	for addr, expected := range validEmailAddresses {
		t.Run(addr, func(t *testing.T) {
			results, err := mail.ParseAddressList(addr)
			assert.NoError(t, err, "valid email address")
			assert.Len(t, results, 1, "list of one address")
			assert.Equal(t, expected, results[0], "expected: %s", expected)
		})
	}
}

func TestAddressList_Split(t *testing.T) {
	tests := []struct {
		l       AddressList
		want    []Address
		wantErr bool
	}{
		{l: ``, want: nil},
		{l: `<hello@example.com>,`, want: []Address{`hello@example.com`}},
		{l: `<Hello@example.com>, World@example.com`, want: []Address{`hello@example.com`, `world@example.com`}},
		// Partial result: parsable addresses are returned alongside the error.
		{l: `alice@example.com, broken address, bob@example.com`, want: []Address{`alice@example.com`, `bob@example.com`}, wantErr: true},
		{l: `@broken`, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.l), func(t *testing.T) {
			got, err := tt.l.Split()
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressList(%#v).Split() error = %v, wantErr %v", string(tt.l), err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressList(%#v).Split() = %v, want %v", string(tt.l), got, tt.want)
			}
		})
	}
}

func TestAddressList_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    AddressList
		wantErr bool
	}{
		// The array branch reads a PostgreSQL text[] column. Its elements
		// are quoted and escaped, so they have to be decoded rather than
		// joined verbatim, otherwise the quotes end up inside an address.
		{
			name:  "quoted array elements",
			value: `{"a@example.com","b@example.com"}`,
			want:  "a@example.com, b@example.com",
		},
		{
			name:  "quoted element with a comma and escaped quotes",
			value: `{"\"Doe, John\" <john@example.com>","b@example.com"}`,
			want:  `"Doe, John" <john@example.com>, b@example.com`,
		},
		{
			name:  "unquoted array elements",
			value: `{a@example.com,b@example.com}`,
			want:  "a@example.com, b@example.com",
		},
		{name: "empty array", value: `{}`, want: ""},
		// A hand written literal, unlike PostgreSQL output, may pad its
		// elements. The padding is not part of the address.
		{name: "padded elements", value: `{ a@example.com , b@example.com }`, want: "a@example.com, b@example.com"},
		{name: "whitespace only array", value: `{ }`, want: ""},
		// Regression cases for the array codec swap (nullable.SplitArray ->
		// notnull.StringArray). Each behaves differently than before, so each
		// is pinned rather than left to be discovered in production.
		// A SQL NULL element used to become the address literally spelled NULL.
		{name: "NULL element", value: `{NULL}`, wantErr: true},
		// Quoted, it is the ordinary string "NULL", not a SQL NULL.
		{name: "quoted NULL element", value: `{"NULL"}`, want: "NULL"},
		// A trailing comma used to be accepted and yield one element.
		{name: "trailing comma", value: `{a@example.com,}`, wantErr: true},
		// Backslash escapes are unescaped now; before, both backslashes stayed.
		{name: "escaped backslash", value: `{"a\\b@example.com"}`, want: `a\b@example.com`},
		// A nested literal used to yield the inner braces as a single address.
		{name: "nested array", value: `{{a@example.com}}`, wantErr: true},
		// A non-array string is an already joined list, so it is taken as is.
		{name: "plain string", value: "a@example.com, b@example.com", want: "a@example.com, b@example.com"},
		{name: "bytes", value: []byte(`{"a@example.com"}`), want: "a@example.com"},
		// AddressList is the not-null type, so SQL NULL is an error here.
		// NullableAddressList is what scans NULL into the empty list.
		{name: "null", value: nil, wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid array", value: `{,}`, wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l AddressList
			err := l.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if l != tt.want {
				t.Errorf("Scan() = %q, want %q", l, tt.want)
			}
		})
	}
}
