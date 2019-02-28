package date

// NullDate is identical to Date, except that IsZero() is considered valid
// by the Valid() and Validate() methods.
// NullDate implements the database/sql.Scanner and database/sql/driver.Valuer interfaces,
// and will treat an empty/zero Date string as SQL NULL value.
// The main difference between Date and NullDate is:
// Date("").Valid() == false
// NullDate("").Valid() == true
type NullDate struct {
	Date
}

func (n NullDate) Valid() bool {
	return n.Validate() == nil
}

func (n NullDate) Validate() error {
	if n.Date.IsZero() {
		return nil
	}
	return n.Date.Validate()
}
