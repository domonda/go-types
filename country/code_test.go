package country

import "testing"

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
