package sqlarray

import (
	"github.com/lib/pq"
)

// Bools implements the sql.Scanner and driver.Valuer interfaces
// for a slice of bool.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
type Bools = pq.BoolArray
