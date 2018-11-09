package sqlarray

import (
	"github.com/lib/pq"
)

// Strings implements the sql.Scanner and driver.Valuer interfaces
// for a slice of string.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type Strings pq.StringArray
