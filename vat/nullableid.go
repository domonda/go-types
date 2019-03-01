package vat

// Null is an empty string and will be treatet as SQL NULL.
var Null NullableID

// NullableID is identical to ID, except that the Null value (empty string)
// is considered valid by the Valid() and Validate() methods.
type NullableID struct {
	ID
}

func (n NullableID) Valid() bool {
	return n.Validate() == nil
}

func (n NullableID) Validate() error {
	if n.ID == "" {
		return nil
	}
	return n.ID.Validate()
}
