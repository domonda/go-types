package nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonEmptyString_MarshalJSON(t *testing.T) {
	type Struct struct {
		Null     NonEmptyString
		NullOmit NonEmptyString `json:",omitempty"`
		NonNull  NonEmptyString
	}
	s := Struct{NonNull: "Hello"}

	j, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, `{"Null":null,"NonNull":"Hello"}`, string(j))
}

func TestNonEmptyString_UnmarshalJSON(t *testing.T) {
	type Struct struct {
		Null     NonEmptyString
		NullOmit NonEmptyString `json:",omitempty"`
		Empty    NonEmptyString
		NonNull  NonEmptyString
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
