package nullable

import "github.com/lib/pq"

// StringArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of strings.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type StringArray = pq.StringArray
