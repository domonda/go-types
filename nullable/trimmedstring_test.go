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
	s := Struct{NonNull: "Hello"}

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
		"NullOmit": "here",
		"Empty": "",
		"NonNull": "Hello"
	}`
	expected := Struct{NullOmit: "here", NonNull: "Hello"}
	var result Struct

	err := json.Unmarshal([]byte(input), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
