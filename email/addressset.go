package email

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"maps"
	"sort"
	"unsafe"

	"github.com/domonda/go-types"
	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/nullable"
)

// AddressSet is a set of unique email addresses.
//
// Value writes the set as a PostgreSQL array text literal
// ({"a@example.com",...}) and Scan reads that format, see
// https://www.postgresql.org/docs/current/arrays.html; Scan also
// accepts a plain non-array string as a single-element set.
// The array format is understood by PostgreSQL and array-compatible
// databases such as CockroachDB and YugabyteDB; databases without a
// native array type (MySQL, MariaDB, SQLite, SQL Server, Oracle) are
// not supported.
type AddressSet map[Address]struct{}

// Compile-time check that AddressSet implements types.NormalizableValidator[AddressSet]
var _ types.NormalizableValidator[AddressSet] = AddressSet{}

// MakeAddressSet returns an AddressSet containing the passed addresses.
func MakeAddressSet(addrs ...Address) AddressSet {
	set := make(AddressSet, len(addrs))
	for _, addr := range addrs {
		set[addr] = struct{}{}
	}
	return set
}

// NormalizedAddressSet returns an AddressSet containing the normalized
// form of the passed addresses, or an error if any address is invalid.
func NormalizedAddressSet(addrs ...Address) (AddressSet, error) {
	set := make(AddressSet, len(addrs))
	for _, addr := range addrs {
		norm, err := addr.Normalized()
		if err != nil {
			return nil, err
		}
		set[norm] = struct{}{}
	}
	return set, nil
}

// NormalizedAddressPartSet returns an AddressSet containing the normalized
// address parts (without name part) of the passed addresses,
// or an error if any address is invalid.
func NormalizedAddressPartSet(addrs ...Address) (AddressSet, error) {
	set := make(AddressSet, len(addrs))
	for _, addr := range addrs {
		norm, err := addr.AddressPart()
		if err != nil {
			return nil, err
		}
		set[norm] = struct{}{}
	}
	return set, nil
}

// Len returns the number of values in the set.
func (set AddressSet) Len() int {
	return len(set)
}

// IsEmpty returns true if the set is empty or nil.
func (set AddressSet) IsEmpty() bool {
	return len(set) == 0
}

// IsNull implements the nullable.Nullable interface
// by returning true if the set is nil.
func (set AddressSet) IsNull() bool {
	return set == nil
}

// Contains returns true if the set contains the passed address.
func (set AddressSet) Contains(addr Address) bool {
	_, ok := set[addr]
	return ok
}

// Add inserts the passed address into the set,
// allocating the underlying map if necessary.
func (set *AddressSet) Add(addr Address) {
	if *set == nil {
		*set = AddressSet{addr: struct{}{}}
	} else {
		(*set)[addr] = struct{}{}
	}
}

// AddSet inserts all addresses from other into the set,
// allocating the underlying map if necessary.
func (set *AddressSet) AddSet(other AddressSet) {
	if len(other) == 0 {
		return
	}
	if *set == nil {
		*set = make(AddressSet, len(other))
	}
	for addr := range other {
		(*set)[addr] = struct{}{}
	}
}

// AddNormalized normalizes the passed address and inserts it into the set.
// It returns an error if the address is invalid.
func (set *AddressSet) AddNormalized(addr Address) error {
	norm, err := addr.Normalized()
	if err != nil {
		return err
	}
	set.Add(norm)
	return nil
}

// AddAddressPart inserts the normalized address part (without name part)
// of the passed address into the set.
// It returns an error if the address is invalid.
func (set *AddressSet) AddAddressPart(addr Address) error {
	norm, err := addr.AddressPart()
	if err != nil {
		return err
	}
	set.Add(norm)
	return nil
}

// Delete removes the passed address from the set.
func (set AddressSet) Delete(val Address) {
	delete(set, val)
}

// DeleteSlice removes all addresses in the passed slice from the set.
func (set AddressSet) DeleteSlice(vals []Address) {
	for _, val := range vals {
		delete(set, val)
	}
}

// DeleteSet removes all addresses contained in other from the set.
func (set AddressSet) DeleteSet(other AddressSet) {
	for str := range other {
		delete(set, str)
	}
}

// Clear removes all addresses from the set.
func (set AddressSet) Clear() {
	clear(set)
}

// Clone returns a copy of the set, or nil if the set is nil.
func (set AddressSet) Clone() AddressSet {
	if set == nil {
		return nil
	}
	return maps.Clone(set)
}

// GetOne returns one address of the set
// or an empty string if the set is empty.
func (set AddressSet) GetOne() Address {
	for addr := range set {
		return addr
	}
	return ""
}

// Sorted returns the addresses of the set as a sorted slice.
func (set AddressSet) Sorted() []Address {
	return types.SetToSortedSlice(set)
}

// Strings returns the addresses of the set as a sorted slice of strings.
func (set AddressSet) Strings() []string {
	switch len(set) {
	case 0:
		return nil
	case 1:
		for addr := range set {
			return []string{string(addr)}
		}
	}
	s := make([]string, len(set))
	i := 0
	for addr := range set {
		s[i] = string(addr)
		i++
	}
	sort.Strings(s)
	return s
}

// AddressList returns the sorted addresses of the set
// joined as a comma separated AddressList.
func (set AddressSet) AddressList() AddressList {
	return AddressListJoin(set.Sorted()...)
}

// String implements the fmt.Stringer interface returning
// the sorted addresses joined as a comma separated list.
func (set AddressSet) String() string {
	return string(set.AddressList())
}

// Normalized returns a new AddressSet with all addresses normalized,
// or the original set together with an error if any address is invalid.
func (set AddressSet) Normalized() (AddressSet, error) {
	if len(set) == 0 {
		return set, nil
	}
	normalized := make(AddressSet, len(set))
	for addr := range set {
		norm, err := addr.Normalized()
		if err != nil {
			return set, err
		}
		normalized.Add(norm)
	}
	return normalized, nil
}

// Validate returns the first error encountered
// validating the addresses of the set.
func (set AddressSet) Validate() error {
	for addr := range set {
		err := addr.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// Valid returns true if all addresses in the set are valid.
func (set AddressSet) Valid() bool {
	return set.Validate() == nil
}

// ValidAndNormalized returns true if all addresses in the set are valid and already normalized.
func (set AddressSet) ValidAndNormalized() bool {
	norm, err := set.Normalized()
	if err != nil {
		return false
	}
	if len(set) != len(norm) {
		return false
	}
	for addr := range set {
		if !norm.Contains(addr) {
			return false
		}
	}
	return true
}

// Scan implements the database/sql.Scanner interface.
// Supports scanning SQL strings and string arrays.
func (set *AddressSet) Scan(value any) error {
	switch s := value.(type) {
	case string:
		if s == "" {
			return errors.New("can't scan empty string as email.AddressSet")
		}
		if s[0] == '{' && s[len(s)-1] == '}' {
			array, err := nullable.SplitArray(s)
			if err != nil {
				// fmt.Printf("ARRAY: %#v\n", s)
				return fmt.Errorf("can't scan SQL array string %q as email.AddressSet because of: %w", s, err)
			}
			*set = make(AddressSet, len(array))
			for _, addr := range array {
				set.Add(Address(addr))
			}
		} else {
			*set = AddressSet{Address(s): struct{}{}}
		}
		return nil

	case []byte:
		return set.Scan(string(s))

	case nil:
		*set = nil
		return nil

	default:
		return fmt.Errorf("can't scan %T as email.AddressSet", value)
	}
}

// Value implements the driver database/sql/driver.Valuer interface.
func (set AddressSet) Value() (driver.Value, error) {
	if set == nil {
		return nil, nil
	}
	if len(set) == 0 {
		return "{}", nil
	}
	s := set.Sorted()
	return (*notnull.StringArray)(unsafe.Pointer(&s)).Value() //#nosec G103 -- unsafe OK
}
