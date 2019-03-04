package uu

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// NullableID can be used with the standard sql package to represent a
// ID value that can be NULL in the database
type NullableID struct {
	ID    ID
	Valid bool
}

// NullableIDFrom creates a new valid NullableID
func NullableIDFrom(id ID) NullableID {
	return NewNullableID(id, true)
}

// NullableIDFromString creates a new valid NullableID
func NullableIDFromString(s string) (n NullableID, err error) {
	n.ID, err = IDFromString(s)
	if err != nil {
		return NullableID{}, err
	}
	n.Valid = true
	return n, nil
}

// NullableIDFromBytes creates a new valid NullableID
func NullableIDFromBytes(s []byte) (n NullableID, err error) {
	n.ID, err = IDFromBytes(s)
	if err != nil {
		return NullableID{}, err
	}
	n.Valid = true
	return n, nil
}

// NullableIDFromPtr creates a new NullableID that be null if ptr is nil.
func NullableIDFromPtr(ptr *ID) NullableID {
	if ptr == nil {
		return NullableID{}
	}
	return NewNullableID(*ptr, true)
}

// NewNullableID creates a new NullableID
func NewNullableID(u ID, valid bool) NullableID {
	return NullableID{
		ID:    u,
		Valid: valid,
	}
}

// SetValid changes this NullableID's value and also sets it to be non-null.
func (u *NullableID) SetValid(v ID) {
	u.ID = v
	u.Valid = true
}

// Ptr returns a pointer to this NullableID's value, or a nil pointer if this NullableID is null.
func (u NullableID) Ptr() *ID {
	if !u.Valid {
		return nil
	}
	return &u.ID
}

// Value implements the driver.Valuer interface.
func (u NullableID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	// Delegate to ID Value function
	return u.ID.Value()
}

// Scan implements the sql.Scanner interface.
func (u *NullableID) Scan(src interface{}) error {
	if src == nil {
		u.ID, u.Valid = IDNil, false
		return nil
	}

	// Delegate to ID Scan function
	u.Valid = true
	return u.ID.Scan(src)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input. Blank string input does not produce a null ID.
// It also supports unmarshalling a sql.NullString.
func (u *NullableID) UnmarshalJSON(data []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		u.ID, err = IDFromString(x)
	case map[string]interface{}:
		var n sql.NullString
		err = json.Unmarshal(data, &n)
		if n.Valid {
			u.ID, err = IDFromString(n.String)
		}
	case nil:
		u.ID = IDNil
		u.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type uuid.NullString", reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if Valid == false.
func (u NullableID) MarshalJSON() ([]byte, error) {
	if !u.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(u.ID.String())
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (u NullableID) MarshalText() ([]byte, error) {
	if !u.Valid {
		return []byte{}, nil
	}
	return []byte(u.ID.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (u *NullableID) UnmarshalText(text []byte) (err error) {
	if len(text) == 0 {
		u.ID = IDNil
		u.Valid = false
		return nil
	}
	u.ID, err = IDFromBytes(text)
	u.Valid = err == nil
	return err
}

// String returns the ID as string if Valid == true,
// or "null" if Valid == false.
func (u NullableID) String() string {
	if !u.Valid {
		return "null"
	}
	return u.ID.String()
}
