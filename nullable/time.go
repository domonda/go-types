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
// As with time.Time use the IsZero method the check for NULL
// instead of comparing with the default value of the type.
type Time struct {
	time.Time
}

// String returns Time.String() or "NULL" if Time.IsZero().
func (nt Time) String() string {
	if nt.Time.IsZero() {
		return "NULL"
	}
	return nt.Time.String()
}

// Scan implements the database/sql.Scanner interface.
func (nt *Time) Scan(value interface{}) error {
	switch t := value.(type) {
	case nil:
		*nt = Time{}
		return nil

	case time.Time:
		nt.Time = t
		return nil

	default:
		return fmt.Errorf("can't scan %T as nullable.Time", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (nt Time) Value() (driver.Value, error) {
	if nt.Time.IsZero() {
		return nil, nil
	}
	return nt.Time, nil
}

// UnarshalJSON implements encoding/json.Unmarshaler.
// Interprets []byte(nil), []byte(""), []byte("null") as null.
func (nt *Time) UnmarshalJSON(sourceJSON []byte) error {
	if len(sourceJSON) == 0 || bytes.Equal(sourceJSON, []byte("null")) {
		*nt = Time{}
		return nil
	}
	return json.Unmarshal(sourceJSON, &nt.Time)
}

// MarshalJSON implements encoding/json.Marshaler
func (nt Time) MarshalJSON() ([]byte, error) {
	if nt.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}
