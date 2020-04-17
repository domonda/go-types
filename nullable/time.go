package nullable

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time implements the sql.Scanner and driver.Valuer interfaces
// and represents a time.Time where the zero time instant
// (which is the empty default value of the type)
// is interpreted as SQL NULL.
// It assumes that zero time instant is never used
// in any real life application so it's free
// to be used as magical value for representing NULL.
// As with time.Time use the IsZero method the check for NULL
// instead of comparing with the default value of the type.
type Time struct {
	time.Time
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
