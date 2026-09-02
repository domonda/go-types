package email

import "testing"

func TestNullableAddressList_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    NullableAddressList
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
		// SQL NULL is the empty list, unlike the empty string which is
		// rejected because it can only come from a non-NULL column.
		{name: "null", value: nil, want: ""},
		{name: "empty string", value: "", wantErr: true},
		{name: "invalid array", value: `{,}`, wantErr: true},
		{name: "unsupported type", value: 123, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullableAddressList
			err := n.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if n != tt.want {
				t.Errorf("Scan() = %q, want %q", n, tt.want)
			}
		})
	}
}

func TestNullableAddressList_ScanValueRoundTrip(t *testing.T) {
	// Value always writes the plain joined string, never an array, so a
	// value written by this type has to scan back unchanged.
	lists := []NullableAddressList{
		"",
		"a@example.com",
		"a@example.com, b@example.com",
		`"Doe, John" <john@example.com>, b@example.com`,
	}
	for _, want := range lists {
		t.Run(string(want), func(t *testing.T) {
			value, err := want.Value()
			if err != nil {
				t.Fatalf("Value returned %v", err)
			}
			var got NullableAddressList
			if err := got.Scan(value); err != nil {
				t.Fatalf("Scan(%v) returned %v", value, err)
			}
			if got != want {
				t.Errorf("Scan(Value()) = %q, want %q", got, want)
			}
		})
	}
}
