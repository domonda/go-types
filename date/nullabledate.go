package date

// Null is an empty string and will be treatet as SQL NULL.
// date.Null.IsZero() == true
var Null = NullableDate{""}

// NullableDate is identical to Date, except that IsZero() is considered valid
// by the Valid() and Validate() methods.
// NullableDate implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty/zero Date string as SQL NULL value.
// The main difference between Date and NullableDate is:
// Date("").Valid() == false
// NullableDate("").Valid() == true
type NullableDate struct {
	Date
}

func (n NullableDate) Valid() bool {
	return n.Validate() == nil
}

func (n NullableDate) Validate() error {
	if n.Date.IsZero() {
		return nil
	}
	return n.Date.Validate()
}
