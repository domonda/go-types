package types

import (
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

var (
	jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	emptyInterfaceTye = reflect.TypeOf((*interface{})(nil)).Elem()
)

// CanMarshalJSON returns if a type can be marshalled as JSON
func CanMarshalJSON(t reflect.Type) bool {
	if t == emptyInterfaceTye {
		return true
	}

	if t.Implements(jsonMarshalerType) {
		return true
	}
	kind := t.Kind()
	if kind != reflect.Ptr && reflect.PtrTo(t).Implements(jsonMarshalerType) {
		return true
	}

	if t.Implements(textMarshalerType) {
		return true
	}
	if kind != reflect.Ptr && reflect.PtrTo(t).Implements(textMarshalerType) {
		return true
	}

	if kind == reflect.Ptr {
		t = t.Elem()
		kind = t.Kind()
	}
	return kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice
}

// JSON is a []byte slice that implements the interfaces:
// json.Marshaler, json.Unmarshaler, driver.Value, sql.Scanner.
// Its nil value it is marshalled as the JSON value "null"
// and the SQL NULL value.
type JSON []byte

func MarshalJSON(source interface{}) (JSON, error) {
	return json.Marshal(source)
}

// MarshalFrom marshalles source as JSON and sets it
// at j when there was no error.
func (j *JSON) MarshalFrom(source interface{}) error {
	data, err := json.Marshal(source)
	if err == nil {
		*j = data
	}
	return err
}

// UnmarshalTo unmashalles the JSON of j to dest
func (j JSON) UnmarshalTo(dest interface{}) error {
	return json.Unmarshal(j, dest)
}

// MarshalJSON returns j as the JSON encoding of j.
// MarshalJSON implements encoding/json.Marshaler
// See the package function MarshalJSON to marshal
// a struct into JSON
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
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
	*j = append((*j)[0:0], sourceJSON...)
	return nil
}

// Valid reports whether j is a valid JSON encoding.
func (j JSON) Valid() bool {
	return json.Valid(j)
}

// Value returns j as a SQL value.
func (j JSON) Value() (driver.Value, error) {
	return []byte(j), nil
}

// Scan stores the src in *j. No validation is done.
func (j *JSON) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
		// Converting from string does a copy
		*j = JSON(t)
	case []byte:
		// Need to copy because, src will be gone after call
		*j = append((*j)[0:0], t...)
	case nil:
		*j = nil
	default:
		return errors.New("Incompatible type for JSON")
	}
	return nil
}

// String returns the JSON as string
func (j JSON) String() string {
	if j == nil {
		return "null"
	}
	return string(j)
}
