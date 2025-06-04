package account

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumber_TrimLeadingZeros(t *testing.T) {
	tests := []struct {
		no   Number
		want Number
	}{
		{"", ""},
		{"0", ""},
		{"00", ""},
		{"1", "1"},
		{"01", "1"},
		{"10", "10"},
		{"100", "100"},
		{"0001", "1"},
		{"00010", "10"},
		{"0x12", "x12"},
		{"Hello_World", "Hello_World"},
	}
	for _, tt := range tests {
		t.Run(string(tt.no), func(t *testing.T) {
			if got := tt.no.TrimLeadingZeros(); got != tt.want {
				t.Errorf("AccountNo.TrimLeadingZeros() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNumber_Valid(t *testing.T) {
	valid := []Number{
		"0",
		"0/",
		"0/0",
		"Hello_World",
		"a.b",
		"a:b",
		"a,b",
		"a;b",
		"a.b.",
		"a:b:",
		"a,b,",
		"a;b;",
		"D-2007019",
	}
	for _, n := range valid {
		t.Run(string(n), func(t *testing.T) {
			require.Truef(t, n.Valid(), "Number(%#v).Valid()", n)
		})
	}
	invalid := []Number{
		"",
	}
	for _, n := range invalid {
		t.Run(string(n), func(t *testing.T) {
			require.Falsef(t, n.Valid(), "Number(%#v).Valid()", n)
		})
	}
}
