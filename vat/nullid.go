package vat

// ZeroID is identical to ID, except that the Null value (empty string)
// is considered valid by the Valid() and Validate() methods.
type NullID struct {
	ID
}

func (n NullID) Valid() bool {
	return n.Validate() == nil
}

func (n NullID) Validate() error {
	if n.ID == "" {
		return nil
	}
	return n.ID.Validate()
}
