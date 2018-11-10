package uu

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// NullID can be used with the standard sql package to represent a
// ID value that can be NULL in the database
type NullID struct {
	ID    ID
	Valid bool
}

// NullIDFrom creates a new valid NullID
func NullIDFrom(id ID) NullID {
	return NewNullID(id, true)
}

// NullIDFromString creates a new valid NullID
func NullIDFromString(s string) (n NullID, err error) {
	n.ID, err = IDFromString(s)
	if err != nil {
		return NullID{}, err
	}
	n.Valid = true
	return n, nil
}

// NullIDFromBytes creates a new valid NullID
func NullIDFromBytes(s []byte) (n NullID, err error) {
	n.ID, err = IDFromBytes(s)
	if err != nil {
		return NullID{}, err
	}
	n.Valid = true
	return n, nil
}

// NullIDFromPtr creates a new NullID that be null if ptr is nil.
func NullIDFromPtr(ptr *ID) NullID {
	if ptr == nil {
		return NullID{}
	}
	return NewNullID(*ptr, true)
}

// NewNullID creates a new NullID
func NewNullID(u ID, valid bool) NullID {
	return NullID{
		ID:    u,
		Valid: valid,
	}
}

// SetValid changes this NullID's value and also sets it to be non-null.
func (u *NullID) SetValid(v ID) {
	u.ID = v
	u.Valid = true
}

// Ptr returns a pointer to this NullID's value, or a nil pointer if this NullID is null.
func (u NullID) Ptr() *ID {
	if !u.Valid {
		return nil
	}
	return &u.ID
}

// Value implements the driver.Valuer interface.
func (u NullID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	// Delegate to ID Value function
	return u.ID.Value()
}

// Scan implements the sql.Scanner interface.
func (u *NullID) Scan(src interface{}) error {
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
func (u *NullID) UnmarshalJSON(data []byte) (err error) {
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
func (u NullID) MarshalJSON() ([]byte, error) {
	if !u.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(u.ID.String())
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (u NullID) MarshalText() ([]byte, error) {
	if !u.Valid {
		return []byte{}, nil
	}
	return []byte(u.ID.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (u *NullID) UnmarshalText(text []byte) (err error) {
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
func (u NullID) String() string {
	if !u.Valid {
		return "null"
	}
	return u.ID.String()
}
