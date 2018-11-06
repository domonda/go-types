package types

import (
	"time"

	"github.com/guregu/null"
	uuid "github.com/ungerik/go-uuid"
)

var (
	// Year2000 can be used for sanity checks of invoices
	Year2000 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

// TimeFromUnixMs returns a time.Time defined in millisonds
// from Unix Epoch with the given timezon location.
func TimeFromUnixMs(ms int64, loc *time.Location) time.Time {
	locBackup := time.Local
	time.Local = loc
	t := time.Unix(0, ms*1e6)
	time.Local = locBackup
	return t
}

// TimeOrNil returns *timePtr if timePtr is not nil and not the default time value,
// else nil is returned.
func TimeOrNil(timePtr *time.Time) interface{} {
	if timePtr == nil || timePtr.IsZero() {
		return nil
	}
	return *timePtr
}

func Int64OrNil(i *int64) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

// NullTimeFromPtr returns null.Time from a time.Time pointer.
// The zero time.Time value is also considere null.
func NullTimeFromPtr(timePtr *time.Time) (nt null.Time) {
	if timePtr != nil {
		return nt
	}
	nt.Time = *timePtr
	nt.Valid = !nt.Time.IsZero()
	return nt
}

// UUIDOrNil returns *uuidPtr if uuidPtr is not nil and not the default zero value,
// else nil is returned.
func UUIDOrNil(uuidPtr *uuid.UUID) interface{} {
	if uuidPtr == nil || *uuidPtr == uuid.Nil {
		return nil
	}
	return *uuidPtr
}
