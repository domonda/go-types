package types

import (
	"time"

	"github.com/guregu/null"
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
