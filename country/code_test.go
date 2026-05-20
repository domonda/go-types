package country

import "testing"

func TestParseCode(t *testing.T) {
	tests := []struct {
		in      string
		want    Code
		wantErr bool
	}{
		// 1. Short path: already a canonical alpha-2 code.
		{in: "DE", want: DE},
		{in: "FR", want: FR},

		// 2. Codes and names resolved via Normalized.
		{in: "de", want: DE},
		{in: " AUT ", want: AT},
		{in: "A", want: AT},
		{in: "SUI", want: CH},
		{in: "Deutschland", want: DE},
		{in: "Österreich", want: AT},

		// 3. English country name.
		{in: "Germany", want: DE},
		{in: "germany", want: DE},
		{in: "  United Kingdom  ", want: GB},
		{in: "United States", want: US},

		// Errors: input is returned unchanged.
		{in: "", want: "", wantErr: true},
		{in: "xxx", want: "xxx", wantErr: true},
		{in: "Atlantis", want: "Atlantis", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseCode(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCode(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseCode(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCode_Normalized(t *testing.T) {
	tests := []struct {
		c       Code
		want    Code
		wantErr bool
	}{
		{c: "A", want: AT},
		{c: "AUT", want: AT},
		{c: "d", want: DE},
		{c: "Deu", want: DE},
		{c: "Deutschland", want: DE},
		{c: "Österreich", want: AT},
		{c: "Oesterreich", want: AT},
		// Errors
		{c: "", want: "", wantErr: true},
		{c: "xxx", want: "xxx", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			got, err := tt.c.Normalized()
			if (err != nil) != tt.wantErr {
				t.Errorf("Code.Normalized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Code.Normalized() = %v, want %v", got, tt.want)
			}
		})
	}
}
