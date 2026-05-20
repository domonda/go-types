package nullable

import (
	"github.com/domonda/go-types/internal/pq"
)

// BoolArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of bool.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
//
// Value and Scan use the PostgreSQL array text format ({t,f,t}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type BoolArray = pq.BoolArray
