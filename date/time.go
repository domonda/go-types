package date

import "time"

var ParseTimeDefaultLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05",
}

// ParseTime is a non Date related helper function
// that parses the passed string as time.Time.
// It uses time.Parse with the passed layouts
// and returns the first valid parsed time.
// If no layouts are passed, then ParseTimeDefaultLayouts will be used.
// The bool result indicates if a valid time could be parsed.
func ParseTime(str string, layouts ...string) (time.Time, bool) {
	if len(layouts) == 0 {
		layouts = ParseTimeDefaultLayouts
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
