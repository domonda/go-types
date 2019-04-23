package notnull

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON is a []byte slice containing JSON text.
// JSON(nil) is interpreted as an empty json object: {}
// Implements the interfaces:
// json.Marshaler, json.Unmarshaler, driver.Value, sql.Scanner.
// Use nullable.JSON if the JSON(nil) value should be
// interpreted as JSON "null" and SQL "NULL".
type JSON []byte

func MarshalJSON(source interface{}) (JSON, error) {
	return json.Marshal(source)
}

// MarshalFrom marshalles source as JSON and sets it
// at j when there was no error.
func (j *JSON) MarshalFrom(source interface{}) error {
	jsonBytes, err := json.Marshal(source)
	if err == nil {
		*j = jsonBytes
	}
	return err
}

// UnmarshalTo unmashalles the JSON of j to dest
func (j JSON) UnmarshalTo(dest interface{}) error {
	if j == nil {
		return json.Unmarshal([]byte("{}"), dest)
	}
	return json.Unmarshal(j, dest)
}

// MarshalJSON returns j as the JSON encoding of j.
// MarshalJSON implements encoding/json.Marshaler
// See the package function MarshalJSON to marshal
// a struct into JSON
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
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
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

// IsEmpty returns true if j is nil, or an empty JSON value like "", "{}", or "[]"
func (j JSON) IsEmpty() bool {
	return j == nil || string(j) == "{}" || string(j) == "[]"
}

// Scan stores the src in *j. No validation is done.
func (j *JSON) Scan(src interface{}) error {
	switch x := src.(type) {
	case nil:
		*j = JSON("{}") // should we do this, set nil or error?

	case string:
		// Converting from string does a copy
		*j = JSON(x)

	case []byte:
		// Need to copy because, src will be gone after call.
		// Use append trick to make a copy of src bytes
		*j = append(JSON(nil), x...)

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

// Clone returns a copy of j
func (j JSON) Clone() JSON {
	clone := make(JSON, len(j))
	copy(clone, j)
	return clone
}
