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
// as abstract container of IDs.
// It is intended to pass IDs as input arguments,
// but does not support scanning or unmarshalling.
type IDs interface {
	fmt.Stringer
	pretty.Printer
	driver.Valuer
	json.Marshaler

	// AsSet returns the contained IDs as IDSet.
	AsSet() IDSet

	// AsSlice returns the contained IDs as IDSlice.
	AsSlice() IDSlice
}
