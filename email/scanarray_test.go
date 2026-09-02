package email

import (
	"strings"
	"testing"
)

// TestScanAddressArray covers the helper the three Scan methods share, so the
// array decoding and trimming rules are asserted once rather than three times
// through their callers. Before the extraction the same block was copy-pasted
// into all three, and the AddressSet copy had already drifted.
func TestScanAddressArray(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		want    []string
		wantErr bool
	}{
		// Quoted elements are unquoted, which is the whole point of using
		// the same codec that Value writes with.
		{name: "quoted", literal: `{"a@x.com","b@x.com"}`, want: []string{"a@x.com", "b@x.com"}},
		{name: "unquoted", literal: `{a@x.com,b@x.com}`, want: []string{"a@x.com", "b@x.com"}},
		// A quoted element may carry the separator and escaped quotes.
		{name: "escaped quotes and comma", literal: `{"\"Doe, John\" <j@x.com>"}`, want: []string{`"Doe, John" <j@x.com>`}},
		{name: "escaped backslash", literal: `{"a\\b@x.com"}`, want: []string{`a\b@x.com`}},
		// A hand written literal may pad its elements.
		{name: "padded", literal: `{ a@x.com , b@x.com }`, want: []string{"a@x.com", "b@x.com"}},
		{name: "whitespace only element", literal: `{ }`, want: []string{""}},
		{name: "empty array", literal: `{}`, want: []string{}},
		// A single quote is legal inside an unquoted element.
		{name: "single quote", literal: `{single_quote'@x.com}`, want: []string{"single_quote'@x.com"}},
		// Errors: a SQL NULL element, a trailing comma and a nested array.
		{name: "NULL element", literal: `{NULL}`, wantErr: true},
		{name: "trailing comma", literal: `{a@x.com,}`, wantErr: true},
		{name: "nested array", literal: `{{a@x.com}}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scanAddressArray(tt.literal, "TestType")
			if (err != nil) != tt.wantErr {
				t.Fatalf("scanAddressArray() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				// The type name is interpolated so each caller's error
				// names the type the user actually scanned into.
				if got := err.Error(); !strings.Contains(got, "email.TestType") {
					t.Errorf("error = %q, want it to name email.TestType", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("scanAddressArray() = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("scanAddressArray()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
