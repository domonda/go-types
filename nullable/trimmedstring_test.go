package nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimmedString_MarshalJSON(t *testing.T) {
	type Struct struct {
		Null     TrimmedString
		NullOmit TrimmedString `json:",omitempty"`
		NonNull  TrimmedString
	}
	s := Struct{NonNull: "\tHello \n"}

	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, `{"Null":null,"NonNull":"Hello"}`, string(j))
}

func TestTrimmedString_UnmarshalJSON(t *testing.T) {
	type Struct struct {
		Null     TrimmedString
		NullOmit TrimmedString `json:",omitempty"`
		Empty    TrimmedString
		NonNull  TrimmedString
	}
	input := `{
		"Null": null,
		"NullOmit": " here ",
		"Empty": "",
		"NonNull": "Hello   "
	}`
	expected := Struct{NullOmit: "here", NonNull: "Hello"}
	var result Struct

	err := json.Unmarshal([]byte(input), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTrimmedString_IsNull(t *testing.T) {
	tests := []struct {
		s    TrimmedString
		want bool
	}{
		{s: "", want: true},
		{s: " ", want: true},
		{s: " \n\t", want: true},
		{s: "a", want: false},
		{s: "NULL", want: false},
		{s: "  NULL  ", want: false},
		{s: "null", want: false},
		{s: "nil", want: false},
		{s: "<nil>", want: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.s), func(t *testing.T) {
			if got := tt.s.IsNull(); got != tt.want {
				t.Errorf("TrimmedString.IsNull(%#v) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
