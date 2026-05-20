package nullable

import "github.com/domonda/go-types/internal/pq"

// StringArray implements the sql.Scanner and driver.Valuer interfaces
// for a slice of strings.
// A nil slice is mapped to the SQL NULL value,
// and a non nil zero length slice to an empty SQL array '{}'.
//
// Value and Scan use the PostgreSQL array text format ({"a","b"}), see
// https://www.postgresql.org/docs/current/arrays.html. That format is
// understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type StringArray = pq.StringArray
