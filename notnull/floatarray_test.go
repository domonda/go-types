package notnull

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FloatArray_Value(t *testing.T) {
	val, err := FloatArray(nil).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{}", val, "FloatArray(nil).Value() returns empty SQL array")

	val, err = FloatArray([]float64{}).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{}", val, "FloatArray([]float64{}).Value() returns empty SQL array")

	val, err = FloatArray([]float64{1, 2, 3}).Value()
	assert.NoError(t, err, "FloatArray.Value")
	assert.Equal(t, "{1,2,3}", val, "FloatArray.Value")
}

func Test_FloatArray_MarshalJSON(t *testing.T) {
	val, err := json.Marshal(FloatArray(nil))
	assert.NoError(t, err, "json.Marshal(FloatArray(nil))")
	assert.Equal(t, "[]", string(val), "json.Marshal(FloatArray(nil)) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{}))")
	assert.Equal(t, "[]", string(val), "json.Marshal(FloatArray([]float64{})) returns empty JSON array")

	val, err = json.Marshal(FloatArray([]float64{1, 2, 3}))
	assert.NoError(t, err, "json.Marshal(FloatArray([]float64{1, 2, 3}))")
	assert.Equal(t, "[1,2,3]", string(val), "json.Marshal(FloatArray([]float64{1, 2, 3}))")
}
