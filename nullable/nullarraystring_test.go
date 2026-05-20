package nullable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// String method coverage for NullIntArray and NullFloatArray.
// The Value/Scan/IsNull behavior is covered in
// nullintarray_test.go and nullfloatarray_test.go.

func Test_NullIntArray_String(t *testing.T) {
	assert.Equal(t, "NullIntArray<nil>", NullIntArray(nil).String())
	assert.Equal(t, "NullIntArray{}", NullIntArray{}.String())
	assert.Equal(t, "NullIntArray{1,NULL,3}", NullIntArray{
		TypeFrom[int64](1), {}, TypeFrom[int64](3),
	}.String())
}

func Test_NullFloatArray_String(t *testing.T) {
	assert.Equal(t, "NullFloatArray<nil>", NullFloatArray(nil).String())
	assert.Equal(t, "NullFloatArray{}", NullFloatArray{}.String())
	assert.Equal(t, "NullFloatArray{1.5,NULL,3}", NullFloatArray{
		TypeFrom(1.5), {}, TypeFrom(3.0),
	}.String())
}
