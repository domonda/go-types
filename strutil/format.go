package strutil

import (
	"strconv"

	"github.com/guregu/null"
)

// NullFloatToStringOrEmpty format the float value as string,
// or returns an empty string if f.Valid == false
func NullFloatToStringOrEmpty(f null.Float) string {
	if !f.Valid {
		return ""
	}
	return strconv.FormatFloat(f.Float64, 'f', -1, 64)
}

// NullIntToStringOrEmpty format the float value as string,
// or returns an empty string if i.Valid == false
func NullIntToStringOrEmpty(i null.Int) string {
	if !i.Valid {
		return ""
	}
	return strconv.FormatInt(i.Int64, 10)
}

// NullTimeToStringOrEmpty format the float value as string,
// or returns an empty string if t.Valid == false
func NullTimeToStringOrEmpty(t null.Time) string {
	if !t.Valid {
		return ""
	}
	return t.Time.String()
}
