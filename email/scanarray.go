package email

import (
	"fmt"

	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/strutil"
)

// scanAddressArray decodes a PostgreSQL array text literal into its elements
// for the Scan methods of AddressSet, AddressList and NullableAddressList.
//
// It uses the same array codec that AddressSet.Value writes with, so that a
// quoted element, or one containing the separator or an escaped quote, is
// decoded rather than carried through with its quotes attached. Every element
// is then trimmed, because a hand written literal may pad them with spaces;
// callers skip the elements that are empty afterwards.
//
// typeName is the email type being scanned, used only in the error message.
func scanAddressArray(literal, typeName string) ([]string, error) {
	var array notnull.StringArray
	err := array.Scan(literal)
	if err != nil {
		return nil, fmt.Errorf("can't scan SQL array string %q as email.%s because of: %w", literal, typeName, err)
	}
	for i, addr := range array {
		array[i] = strutil.TrimSpace(addr)
	}
	return array, nil
}
