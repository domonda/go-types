package uu

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/domonda/go-pretty"
)

var (
	_ IDs = IDSet{}
	_ IDs = IDSlice{}
)

// IDs is an interface implemented by IDSet and IDSlice
// as abstract collection of IDs.
// It is intended for passing IDs as input arguments,
// but does not support scanning or unmarshalling.
type IDs interface {
	fmt.Stringer
	pretty.Printer
	driver.Valuer
	json.Marshaler

	// Len returns the length of the ID collection.
	Len() int

	// AsSet returns the contained IDs as IDSet.
	AsSet() IDSet

	// AsSlice returns the contained IDs as IDSlice.
	AsSlice() IDSlice

	// ForEach calls the passed function for each ID.
	// Any error from the callback function is returned
	// by ForEach immediatly.
	// Returning a sentinel error is a way to stop the loop
	// with a known cause that might not be a real error.
	ForEach(func(ID) error) error
}
