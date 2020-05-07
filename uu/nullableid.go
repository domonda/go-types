package uu

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

var IDNull = NullableID{IDNil}

// NullableID can be used with the standard sql package to represent a
// ID value that can be NULL in the database
type NullableID struct {
	ID
}

// NullableIDFrom returns a NullableID for an ID
func NullableIDFrom(id ID) NullableID {
	return NullableID{ID: id}
}

// NullableIDFromString creates a new valid NullableID
func NullableIDFromString(s string) (n NullableID, err error) {
	n.ID, err = IDFromString(s)
	if err != nil {
		return IDNull, err
	}
	return n, nil
}

// NullableIDFromBytes creates a new valid NullableID
func NullableIDFromBytes(s []byte) (n NullableID, err error) {
	n.ID, err = IDFromBytes(s)
	if err != nil {
		return IDNull, err
	}
	return n, nil
}

// NullableIDFromPtr creates a new NullableID that be null if ptr is nil.
func NullableIDFromPtr(ptr *ID) NullableID {
	if ptr == nil {
		return IDNull
	}
	return NullableID{*ptr}
}

// Set sets an ID for this NullableID
func (n *NullableID) Set(id ID) {
	n.ID = id
}

// SetNull sets the NullableID to null
func (n *NullableID) SetNull() {
	n.ID = IDNil
}

// Get returns the non nullable ID value
// or panics if the NullableID is null.
// Note: check with IsNull before using Get!
func (n *NullableID) Get() ID {
	if n.IsNull() {
		panic("NULL uu.ID")
	}
	return n.ID
}

// IsNull returns true if the NullableID is null
func (n NullableID) IsNull() bool {
	return n == IDNull
}

// String returns the ID as string or "NULL"
func (n NullableID) String() string {
	return n.StringOr("NULL")
}

// StringOr returns the ID as string or the passed nullStr
func (n NullableID) StringOr(nullStr string) string {
	if n.IsNull() {
		return nullStr
	}
	return n.ID.String()
}

// Valid returns if Variant and Version of this UUID are supported.
// A Nil UUID is also valid.
func (n NullableID) Valid() bool {
	return n == IDNull || n.ID.Valid()
}

// Validate returns an error if the Variant and Version of this UUID are not supported.
// A Nil UUID is also valid.
func (n NullableID) Validate() error {
	if n == IDNull {
		return nil
	}
	return n.ID.Validate()
}

// Ptr returns a pointer to this NullableID's value, or a nil pointer if this NullableID is null.
func (n NullableID) Ptr() *ID {
	if n == IDNull {
		return nil
	}
	return &n.ID
}

// Value implements the driver.Valuer interface.
func (n NullableID) Value() (driver.Value, error) {
	if n == IDNull {
		return nil, nil
	}
	// Delegate to ID Value function
	return n.ID.Value()
}

// Scan implements the sql.Scanner interface.
func (n *NullableID) Scan(src interface{}) error {
	if src == nil {
		*n = IDNull
		return nil
	}
	// Delegate to ID Scan function
	return n.ID.Scan(src)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input. Blank string input does not produce a null ID.
// It also supports unmarshalling a sql.NullString.
func (n *NullableID) UnmarshalJSON(data []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		n.ID, err = IDFromString(x)
	case map[string]interface{}:
		var ns sql.NullString
		err = json.Unmarshal(data, &ns)
		if ns.Valid {
			n.ID, err = IDFromString(ns.String)
		}
	case nil:
		*n = IDNull
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type uuid.NullString", reflect.TypeOf(v).Name())
	}
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if Valid == false.
func (n NullableID) MarshalJSON() ([]byte, error) {
	if n == IDNull {
		return []byte("null"), nil
	}
	return json.Marshal(n.ID.String())
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (n NullableID) MarshalText() ([]byte, error) {
	if n == IDNull {
		return []byte{}, nil
	}
	return []byte(n.ID.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (n *NullableID) UnmarshalText(text []byte) (err error) {
	if len(text) == 0 {
		*n = IDNull
		return nil
	}
	n.ID, err = IDFromBytes(text)
	return err
}
