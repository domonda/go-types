package uu

import (
	"fmt"

	"github.com/domonda/go-errs"
)

const (
	ErrNilID errs.Sentinel = "Nil UUID"

	ErrInvalidVariant errs.Sentinel = "invalid UUID variant"
)

type ErrInvalidVersion int

func (e ErrInvalidVersion) Error() string {
	return fmt.Sprintf("invalid UUID version: %d", e)
}
