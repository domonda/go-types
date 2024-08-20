package account

import "testing"

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
