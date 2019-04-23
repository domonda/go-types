package nullable

import (
	"github.com/lib/pq"
)

// BoolArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of bool.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type BoolArray = pq.BoolArray
