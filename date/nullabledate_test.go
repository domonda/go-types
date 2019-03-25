package date

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NullableDate(t *testing.T) {
	var n NullableDate
	assert.True(t, n.Valid(), "empty NullableDate is valid")
	assert.NoError(t, n.Validate(), "empty NullableDate is valid")

	n = "0001-01-01"
	assert.True(t, n.Valid(), "empty NullableDate is valid")
	assert.NoError(t, n.Validate(), "empty NullableDate is valid")

	assert.Empty(t, n.NormalizedOrNull())
}

func Test_NullableDate_UnmarshalJSON(t *testing.T) {
	sourceJSON := `{
		"empty": "",
		"null": null,
		"notNull": "2012-12-12",
		"invalid": "Not a date!"
	}`
	s := struct {
		Empty   NullableDate `json:"empty"`
		Null    NullableDate `json:"null"`
		NotNull NullableDate `json:"notNull"`
		Invalid NullableDate `json:"invalid"`
	}{}
	err := json.Unmarshal([]byte(sourceJSON), &s)
	assert.NoError(t, err, "json.Unmarshal")
	assert.Equal(t, Null, s.Empty, "empty JSON string is Null")
	assert.Equal(t, Null, s.Null, "JSON null value as Null")
	assert.Equal(t, NullableDate("2012-12-12"), s.NotNull, "valid NullableDate")
	assert.Equal(t, NullableDate("Not a date!"), s.Invalid, "invalid NullableDate parsed as is, without error")
	assert.False(t, s.Invalid.Valid(), "invalid NullableDate parsed as is, not valid")
}
