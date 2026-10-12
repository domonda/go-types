package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// JSON is a []byte slice containing JSON text.
// JSON(nil) is interpreted as an empty json object: {}
// Implements the interfaces
// json.Marshaler, json.Unmarshaler, driver.Valuer and sql.Scanner.
// Use nullable.JSON if the JSON(nil) value should be
// interpreted as JSON "null" and SQL "NULL".
type JSON []byte

// MarshalJSON marshals source as JSON and returns it as a notnull JSON value.
func MarshalJSON(source any) (JSON, error) {
	return json.Marshal(source)
}

// MarshalFrom marshals source as JSON and sets it
// at j when there was no error.
func (j *JSON) MarshalFrom(source any) error {
	jsonBytes, err := json.Marshal(source)
	if err == nil {
		*j = jsonBytes
	}
	return err
}

// UnmarshalTo unmarshals the JSON of j to dest.
// A nil j is unmarshalled as the empty object {}.
func (j JSON) UnmarshalTo(dest any) error {
	if j == nil {
		return json.Unmarshal([]byte("{}"), dest)
	}
	return json.Unmarshal(j, dest)
}

// MarshalJSON returns j as the JSON encoding of j.
// A nil or empty j holds no JSON text and returns {}.
// MarshalJSON implements encoding/json.Marshaler
// See the package function MarshalJSON to marshal
// a struct into JSON
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of sourceJSON.
// The JSON null value is stored as the text null, not as nil.
// UnmarshalJSON implements encoding/json.Unmarshaler
// See method UnmarshalTo for unmarshalling into a struct.
func (j *JSON) UnmarshalJSON(sourceJSON []byte) error {
	if j == nil {
		return errors.New("UnmarshalJSON on nil pointer")
	}
	// Use append trick to make a copy of sourceJSON
	*j = append(JSON(nil), sourceJSON...)
	return nil
}

// Valid reports whether j is a valid JSON encoding.
func (j JSON) Valid() bool {
	if j == nil {
		return false
	}
	return json.Valid(j)
}

// Value returns j as a SQL value.
// A nil or empty JSON is written as the empty object {} like MarshalJSON writes it.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
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
// SQL NULL and an empty string or []byte, which hold
// no JSON text, store the empty object {}.
func (j *JSON) Scan(src any) error {
	switch x := src.(type) {
	case nil:
		*j = JSON("{}") // should we do this, set nil or error?

	case string:
		if x == "" {
			*j = JSON("{}")
		} else {
			// Converting from string does a copy
			*j = JSON(x)
		}

	case []byte:
		if len(x) == 0 {
			*j = JSON("{}")
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
	if j == nil {
		return "{}"
	}
	return string(j)
}

// GoString returns a Go syntax representation of j as a notnull.JSON literal.
// GoString implements the fmt.GoStringer interface.
func (j JSON) GoString() string {
	return fmt.Sprintf("notnull.JSON(`%s`)", j)
}

// PrettyPrint implements the pretty.PrintableWithResult interface
// returning the JSON as a Go literal.
func (j JSON) PrettyPrint(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "`%s`", j.String())
}

// Clone returns a copy of j.
// A nil j returns nil, so the clone also encodes as {}.
func (j JSON) Clone() JSON {
	if j == nil {
		return nil
	}
	clone := make(JSON, len(j))
	copy(clone, j)
	return clone
}
