package uu

import (
	"fmt"

	"github.com/domonda/go-errs"
)

const (
	// ErrNilID is returned when a nil (all-zero) UUID is used where a valid UUID is required.
	ErrNilID errs.Sentinel = "Nil UUID"

	// ErrInvalidVariant is returned when a UUID has an unrecognised layout variant.
	ErrInvalidVariant errs.Sentinel = "invalid UUID variant"
)

// ErrInvalidVersion is returned when a UUID carries an unsupported version number.
type ErrInvalidVersion int

// Error implements the error interface, returning a message that includes
// the invalid version number.
func (e ErrInvalidVersion) Error() string {
	return fmt.Sprintf("invalid UUID version: %d", e)
}
