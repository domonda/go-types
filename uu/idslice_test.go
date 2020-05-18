package uu

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDSlice(t *testing.T) {
	expected := IDSliceMustFromStrings(
		"ec449f0f-e10c-4edb-8b59-0e6c896fdca5",
		"2d6a2c10-e4a6-45a3-a705-8115214a3778",
		"f3e52e97-e976-4a4c-a602-294310bcf935",
		"cc5873e6-286d-48cd-ae88-bda3a1e986e3",
	)

	jsonStr := `["ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"]`

	j, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.Equal(t, jsonStr, string(j))

	var parsed IDSlice
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	err = json.Unmarshal([]byte(`null`), &parsed)
	assert.NoError(t, err)
	assert.Nil(t, parsed)

	j, err = json.Marshal(nil)
	assert.NoError(t, err)
	assert.Equal(t, `null`, string(j))

	parsed = nil
	err = json.Unmarshal([]byte(`[]`), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, IDSlice{}, parsed)

	j, err = json.Marshal(make(IDSlice, 0))
	assert.NoError(t, err)
	assert.Equal(t, `[]`, string(j))

	tests := []string{
		jsonStr,
		`"ec449f0f-e10c-4edb-8b59-0e6c896fdca5","2d6a2c10-e4a6-45a3-a705-8115214a3778","f3e52e97-e976-4a4c-a602-294310bcf935","cc5873e6-286d-48cd-ae88-bda3a1e986e3"`,
		`ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3`,
		`[ec449f0f-e10c-4edb-8b59-0e6c896fdca5,2d6a2c10-e4a6-45a3-a705-8115214a3778,f3e52e97-e976-4a4c-a602-294310bcf935,cc5873e6-286d-48cd-ae88-bda3a1e986e3]`,
	}
	for _, str := range tests {
		t.Run(str, func(t *testing.T) {
			got, err := IDSliceFromString(str)
			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}

	_, err = IDSliceFromString("")
	assert.Error(t, err)

	got, err := IDSliceFromStrings(nil)
	assert.NoError(t, err)
	assert.Nil(t, got)
}
