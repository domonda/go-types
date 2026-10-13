package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
)

// UnmarshalJSONStringOrNull implements encoding/json.Unmarshaler for a
// nullable string type whose null value is the empty string.
// The JSON null value sets *dst to "", so that decoding null into a field
// that holds a value resets it. A JSON string is assigned unchanged,
// without normalization or validation, like encoding/json does for a
// string type, any other JSON value is an error.
func UnmarshalJSONStringOrNull[S ~string](dst *S, data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		*dst = ""
		return nil
	}
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			// Name the type S instead of string like encoding/json
			// does for a string type without UnmarshalJSON
			return &json.UnmarshalTypeError{Value: typeErr.Value, Type: reflect.TypeFor[S](), Offset: typeErr.Offset}
		}
		return err
	}
	*dst = S(str)
	return nil
}
