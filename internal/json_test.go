package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnmarshalJSONStringOrNull protects the UnmarshalJSON shared by the
// nullable string types of the domain packages: JSON null must reset a
// field that holds a value, because encoding/json leaves a field unchanged
// for null otherwise, and a string is assigned without normalization.
func TestUnmarshalJSONStringOrNull(t *testing.T) {
	type nullableString string

	n := nullableString("previous")
	require.NoError(t, UnmarshalJSONStringOrNull(&n, []byte(`null`)))
	assert.Equal(t, nullableString(""), n, "null resets the value")

	require.NoError(t, UnmarshalJSONStringOrNull(&n, []byte(`" de "`)))
	assert.Equal(t, nullableString(" de "), n, "a string is assigned unchanged")

	require.NoError(t, UnmarshalJSONStringOrNull(&n, []byte(`""`)))
	assert.Equal(t, nullableString(""), n)

	n = "previous"
	for _, data := range []string{`1`, `true`, `{}`, `["a"]`, `"unterminated`} {
		err := UnmarshalJSONStringOrNull(&n, []byte(data))
		assert.Error(t, err, "UnmarshalJSONStringOrNull(%s)", data)
		assert.Equal(t, nullableString("previous"), n, "unchanged on error for %s", data)
	}
	var typeErr *json.UnmarshalTypeError
	assert.ErrorAs(t, UnmarshalJSONStringOrNull(&n, []byte(`1`)), &typeErr, "same error type as encoding/json")
}
