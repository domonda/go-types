package nullable

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// JSON is a []byte slice containing JSON text or nil
// as the representation of the JSON "null" value.
// that implements the interfaces:
// json.Marshaler, json.Unmarshaler, driver.Value, sql.Scanner.
// The nil value of the type JSON is marshalled as
// the JSON "null" and SQL NULL values.
// The JSON "null" and SQL NULL values are unmarshalled
// as nil value ot the type JSON.
type JSON []byte

func MarshalJSON(source any) (JSON, error) {
	return json.Marshal(source)
}

// IsNull returns true if j is nil.
// IsNull implements the Nullable interface.
func (j JSON) IsNull() bool { return j == nil }

// MarshalFrom marshalles source as JSON and sets it
// at j when there was no error.
func (j *JSON) MarshalFrom(source any) error {
	jsonBytes, err := json.Marshal(source)
	if err == nil {
		if bytes.Equal(jsonBytes, []byte("null")) {
			*j = nil
		} else {
			*j = jsonBytes
		}
	}
	return err
}

// UnmarshalTo unmashalles the JSON of j to dest
func (j JSON) UnmarshalTo(dest any) error {
	return json.Unmarshal(j, dest)
}

// MarshalJSON returns j as the JSON encoding of j.
// MarshalJSON implements encoding/json.Marshaler
// See the package function MarshalJSON to marshal
// a struct into JSON
func (j JSON) MarshalJSON() ([]byte, error) {
	if j.IsNull() {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of sourceJSON.
// UnarshalJSON implements encoding/json.Unmarshaler
// See method Unmarshal for unmarshalling into a struct.
func (j *JSON) UnmarshalJSON(sourceJSON []byte) error {
	if j == nil {
		return errors.New("UnmarshalJSON on nil pointer")
	}
	if sourceJSON == nil || bytes.Equal(sourceJSON, []byte("null")) {
		*j = nil
	} else {
		// Use append trick to make a copy of sourceJSON
		*j = append(JSON(nil), sourceJSON...)
	}
	return nil
}

// Valid reports whether j is a valid JSON encoding.
func (j JSON) Valid() bool {
	if j.IsNull() {
		return true
	}
	return json.Valid(j)
}

// Value returns j as a SQL value.
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return []byte(j), nil
}

// IsEmpty returns true if j is nil, or an empty JSON value like "", "{}", or "[]"
func (j JSON) IsEmpty() bool {
	switch string(j) {
	case "", "{}", "[]":
		return true
	}
	return false
}

// Scan stores the src in *j. No validation is done.
func (j *JSON) Scan(src any) error {
	switch x := src.(type) {
	case nil:
		*j = nil

	case string:
		if x == "null" {
			*j = nil
		} else {
			// Converting from string does a copy
			*j = JSON(x)
		}

	case []byte:
		if bytes.Equal(x, []byte("null")) {
			*j = nil
		} else {
			// Need to copy because, src will be gone after call.
			// Use append trick to make a copy of src bytes
			*j = append(JSON(nil), x...)
		}

	default:
		return errors.New("Incompatible type for JSON")
	}
	return nil
}

// String returns the JSON as string.
// String implements the fmt.Stringer interface.
func (j JSON) String() string {
	if j.IsNull() {
		return "null"
	}
	return string(j)
}

func (j JSON) GoString() string {
	return fmt.Sprintf("nullable.JSON(`%s`)", j)
}

// PrettyPrint implements the pretty.PrintableWithResult interface
// returning the JSON as a Go literal.
func (j JSON) PrettyPrint(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "`%s`", j.String())
}

// Clone returns a copy of j
func (j JSON) Clone() JSON {
	if j == nil {
		return nil
	}
	clone := make(JSON, len(j))
	copy(clone, j)
	return clone
}
