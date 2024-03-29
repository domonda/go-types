package date

import (
	"fmt"
	"time"
)

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
func ParseTime(str string, layouts ...string) (time.Time, error) {
	if len(layouts) == 0 {
		layouts = ParseTimeDefaultLayouts
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q as time with layouts %v", str, layouts)
}
