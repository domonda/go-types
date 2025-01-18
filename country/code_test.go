package country

import "testing"

func TestCode_NormalizedWithAltCodes(t *testing.T) {
	tests := []struct {
		c       Code
		want    Code
		wantErr bool
	}{
		{c: "A", want: AT},
		{c: "AUT", want: AT},
		{c: "d", want: DE},
		{c: "Deu", want: DE},
		// Errors
		{c: "", want: "", wantErr: true},
		{c: "xxx", want: "xxx", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			got, err := tt.c.NormalizedWithAltCodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Code.NormalizedWithAltCodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Code.NormalizedWithAltCodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
