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
// notnull.SplitArrayValues decodes the same values from a PostgreSQL
// literal and additionally accepts JSON arrays and the unquoted [a,b]
// form. The pq parser is kept here because an email array is only ever
// read from a PostgreSQL column, for which the stricter parser is the
// right one: it rejects a NULL element or a multi dimensional array
// instead of passing their text through as an address.
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
