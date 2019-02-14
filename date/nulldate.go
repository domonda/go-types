package date

// NullDate is identical to Date, except that IsZero() is considered valid
// by the Valid() and Validate() methods.
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
