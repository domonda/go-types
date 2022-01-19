package float

import (
	"math"
	"testing"
)

func TestTolerant_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		json         []byte
		wantTolerant Tolerant
		wantErr      bool
	}{
		{name: "nil", json: nil, wantErr: true},
		{name: "JSON boolean", json: []byte(`true`), wantErr: true},
		{name: "JSON non number string", json: []byte(`"true"`), wantErr: true},

		{name: "JSON null", json: []byte(`null`), wantTolerant: 0},
		{name: "empty JSON string", json: []byte(`""`), wantTolerant: 0},
		{name: "0", json: []byte(`0`), wantTolerant: 0},
		{name: "-0.12345", json: []byte(`-0.12345`), wantTolerant: -0.12345},
		{name: "naked NaN", json: []byte(`NaN`), wantTolerant: Tolerant(math.NaN())},
		{name: "quoted NaN", json: []byte(`"NaN"`), wantTolerant: Tolerant(math.NaN())},
		{name: "naked Inf", json: []byte(`Inf`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "quoted Inf", json: []byte(`"Inf"`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "naked +Inf", json: []byte(`+Inf`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "quoted +Inf", json: []byte(`"+Inf"`), wantTolerant: Tolerant(math.Inf(1))},
		{name: "naked -Inf", json: []byte(`-Inf`), wantTolerant: Tolerant(math.Inf(-1))},
		{name: "quoted -Inf", json: []byte(`"-Inf"`), wantTolerant: Tolerant(math.Inf(-1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var amount Tolerant = 666 // Init with value different from default 0
			err := amount.UnmarshalJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tolerant.UnmarshalJSON(%s) error = %v, wantErr %v", tt.json, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalInclNaN(float64(amount), float64(tt.wantTolerant)) {
				t.Errorf("Tolerant.UnmarshalJSON(%s) got %f, want %f", tt.json, amount, tt.wantTolerant)
			}
		})
	}
}

func equalInclNaN(a, b float64) bool {
	if math.IsNaN(a) {
		return math.IsNaN(b)
	}
	if math.IsNaN(b) {
		return false
	}
	return a == b
}
