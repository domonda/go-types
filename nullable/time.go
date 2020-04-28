package nullable

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Time represents a time.Time where the zero time instant
// (which is the empty default value of the type)
// is interpreted as SQL NULL and JSON null.
// It implements the sql.Scanner and driver.Valuer interfaces
// and also json.Marshaler and json.Unmarshaler.
// It assumes that zero time instant is never used
// in any real life application so it's free
// to be used as magical value for representing NULL.
// IsNull uses time.Time.IsZero internally.
type Time struct {
	time.Time
}

// TimeFrom returns a nullable.Time from a time.Time
func TimeFrom(t time.Time) Time {
	return Time{Time: t}
}

// TimeFromPtr returns a nullable.Time from a time.Time pointer
// with nil interpreted as null.
func TimeFromPtr(ptr *time.Time) Time {
	if ptr == nil {
		return Time{}
	}
	return Time{Time: *ptr}
}

// Ptr returns a pointer to Time or nil if IsNull
func (n Time) Ptr() *time.Time {
	if n.IsNull() {
		return nil
	}
	return &n.Time
}

// IsNull returns true if the Time is null.
// Uses time.Time.IsZero internally.
func (n Time) IsNull() bool {
	return n.Time.IsZero()
}

// String returns Time.String() or "NULL" if Time.IsZero().
func (n Time) String() string {
	return n.StringOr("NULL")
}

// StringOr returns Time.String() or the passed nullStr if Time.IsZero().
func (n Time) StringOr(nullStr string) string {
	if n.IsNull() {
		return nullStr
	}
	return n.Time.String()
}

// Scan implements the database/sql.Scanner interface.
func (n *Time) Scan(value interface{}) error {
	switch t := value.(type) {
	case nil:
		*n = Time{}
		return nil

	case time.Time:
		n.Time = t
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.Time", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (n Time) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Time, nil
}

// UnarshalJSON implements encoding/json.Unmarshaler.
// Interprets []byte(nil), []byte(""), []byte("null") as null.
func (n *Time) UnmarshalJSON(sourceJSON []byte) error {
	if len(sourceJSON) == 0 || bytes.Equal(sourceJSON, []byte("null")) /*|| bytes.Equal(sourceJSON, []byte(`"NULL"`))*/ {
		*n = Time{}
		return nil
	}
	return json.Unmarshal(sourceJSON, &n.Time)
}

// MarshalJSON implements encoding/json.Marshaler
func (n Time) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}
