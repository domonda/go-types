package email

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"iter"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/domonda/go-errs"
	"github.com/domonda/go-types"
	"github.com/domonda/go-types/mapset"
	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/strutil"
)

// AddressSet is a set of unique email addresses.
//
// AddressSet is a collection: every format reads and writes the addresses
// as stored, because normalizing could merge two of them and the set would
// read back with fewer elements than it was written with. Use Normalized
// or NormalizedAddressList to normalize the addresses.
//
// Value writes the set as a PostgreSQL array text literal
// ({"a@example.com",...}), see
// https://www.postgresql.org/docs/current/arrays.html, which is understood
// by PostgreSQL and array-compatible databases such as CockroachDB and
// YugabyteDB. For a plain string column, which works with every database,
// write AddressList() instead. Scan reads a PostgreSQL array, a JSON array
// or object, and a plain address list, see Scan. MarshalJSON writes a
// sorted JSON array, MarshalText and so encoding/xml an address list.
// See "Reading and writing values" in the README of the module.
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
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) Contains(addr Address) bool {
	return mapset.Contains(set, addr)
}

// ContainsAny returns true if any of the passed addresses are in the set.
// It returns false for no addresses and is valid to call on a nil AddressSet.
func (set AddressSet) ContainsAny(addrs ...Address) bool {
	return slices.ContainsFunc(addrs, set.Contains)
}

// ContainsAll reports whether the set contains all addresses yielded by seq.
// It returns true for an empty sequence and is valid to call on a nil AddressSet.
func (set AddressSet) ContainsAll(seq iter.Seq[Address]) bool {
	return mapset.ContainsAll(set, seq)
}

// All returns an iterator over the addresses of the set in undefined order.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) All() iter.Seq[Address] {
	return mapset.All(set)
}

// Insert adds addr to the set and reports whether the set was changed.
//
// Unlike [AddressSet.Add] this method has a value receiver, which makes
// AddressSet implement the abstract set interface of the Go collections
// proposal, but it panics if the set is nil and addr is not already an
// element. Use Add to insert into a possibly nil set.
func (set AddressSet) Insert(addr Address) bool {
	return mapset.Insert(set, addr)
}

// InsertAll adds all addresses yielded by seq to the set
// and reports whether the set was changed.
// It panics if the set is nil and seq yields an address
// that is not already an element. There is no nil-safe sequence insert:
// on a possibly nil set collect the sequence first and use
// [AddressSet.AddSet], which takes an AddressSet rather than an iter.Seq.
func (set AddressSet) InsertAll(seq iter.Seq[Address]) bool {
	return mapset.InsertAll(set, seq)
}

// Add inserts the passed address into the set,
// allocating the underlying map if necessary.
//
// This is the nil-safe counterpart of [AddressSet.Insert]:
// it has a pointer receiver so that it can assign a newly
// allocated map to a nil AddressSet variable or struct field.
func (set *AddressSet) Add(addr Address) {
	if *set == nil {
		*set = AddressSet{addr: struct{}{}}
	} else {
		(*set)[addr] = struct{}{}
	}
}

// AddSet inserts all addresses from other into the set,
// allocating the underlying map if necessary.
//
// This is the nil-safe counterpart of [AddressSet.UnionWith]:
// it has a pointer receiver so that it can assign a newly
// allocated map to a nil AddressSet variable or struct field.
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

// Delete removes the passed address from the set
// and reports whether the set was changed.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) Delete(val Address) bool {
	return mapset.Delete(set, val)
}

// DeleteAll removes all addresses yielded by seq from the set
// and reports whether the set was changed.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) DeleteAll(seq iter.Seq[Address]) bool {
	return mapset.DeleteAll(set, seq)
}

// DeleteFunc removes every address for which del returns true
// and reports whether the set was changed.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) DeleteFunc(del func(Address) bool) bool {
	return mapset.DeleteFunc(set, del)
}

// Clear removes all addresses from the set.
// It is valid to call this method on a nil AddressSet.
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

// Union returns a new AddressSet with all addresses of set and other.
func (set AddressSet) Union(other AddressSet) AddressSet {
	return mapset.Union(set, other)
}

// UnionWith adds all addresses of other to set.
// It panics if set is nil and other has an address that is not already
// an element of set; use [AddressSet.AddSet] for a possibly nil set.
func (set AddressSet) UnionWith(other AddressSet) {
	mapset.UnionWith(set, other)
}

// Intersection returns a new AddressSet with the addresses
// that are in both set and other.
func (set AddressSet) Intersection(other AddressSet) AddressSet {
	return mapset.Intersection(set, other)
}

// IntersectionWith removes every address from set that is not also in other.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) IntersectionWith(other AddressSet) {
	mapset.IntersectionWith(set, other)
}

// Intersects reports whether set and other have at least one address in common.
func (set AddressSet) Intersects(other AddressSet) bool {
	return mapset.Intersects(set, other)
}

// Difference returns a new AddressSet with the addresses
// of set that are not in other.
func (set AddressSet) Difference(other AddressSet) AddressSet {
	return mapset.Difference(set, other)
}

// DifferenceWith removes every address of other from set.
// It is valid to call this method on a nil AddressSet.
func (set AddressSet) DifferenceWith(other AddressSet) {
	mapset.DifferenceWith(set, other)
}

// SymmetricDifference returns a new AddressSet containing the addresses
// that are in exactly one of set and other.
func (set AddressSet) SymmetricDifference(other AddressSet) AddressSet {
	return mapset.SymmetricDifference(set, other)
}

// SymmetricDifferenceWith replaces the addresses of set with the addresses
// that are in exactly one of set and other.
// It panics if set is nil and other has an address
// that is not already an element of set.
func (set AddressSet) SymmetricDifferenceWith(other AddressSet) {
	mapset.SymmetricDifferenceWith(set, other)
}

// Equal returns true if set and other contain exactly the same addresses.
func (set AddressSet) Equal(other AddressSet) bool {
	return mapset.Equal(set, other)
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

// AddressList returns the sorted addresses of the set as stored
// joined as a comma separated AddressList.
func (set AddressSet) AddressList() AddressList {
	return AddressListJoin(set.Sorted()...)
}

// NormalizedAddressList returns the normalized addresses of the set,
// with addresses that normalize to the same one merged, sorted and joined
// as a comma separated AddressList, or an error if an address is invalid.
func (set AddressSet) NormalizedAddressList() (AddressList, error) {
	norm, err := set.Normalized()
	if err != nil {
		return "", err
	}
	return norm.AddressList(), nil
}

// String implements the fmt.Stringer interface returning
// the sorted addresses joined as a comma separated list,
// or "<nil>" for a nil set like [types.Set.String] does.
//
// Unlike the String methods of the address types, String does not
// normalize, because normalizing can merge addresses, and the string
// has to list every element of the set, like the String methods
// of the other set types.
//
// An allocated empty set renders as the empty string, which is what the
// empty address list is, so the nil rendering is what distinguishes the
// two. Use [AddressSet.AddressList] to get the joined list unconditionally.
func (set AddressSet) String() string {
	if set == nil {
		return "<nil>"
	}
	return string(set.AddressList())
}

// Normalized returns a new AddressSet with all addresses normalized,
// or the original set together with an error if any address is invalid.
func (set AddressSet) Normalized() (AddressSet, error) {
	if set == nil {
		return nil, nil
	}
	// An allocated empty set still gets its own map: returning the receiver
	// made the "normalized copy" an alias, so writing to it wrote through.
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
// The elements of a PostgreSQL array (from a text[] column) are trimmed,
// the elements of a JSON array and the keys of a JSON object (from a json
// or jsonb column) are kept exactly, and a plain address list or a JSON
// string is split with SplitAddressList. The addresses are not normalized
// or validated, so that one invalid value doesn't fail a whole query;
// Validate reports it. SQL NULL, a blank string and "null" scan as nil
// set, an empty array or object as empty set. Other SQL types are an error.
func (set *AddressSet) Scan(value any) error {
	var text string
	switch s := value.(type) {
	case nil:
		*set = nil
		return nil
	case string:
		text = s
	case []byte:
		text = string(s)
	default:
		return errs.Errorf("can't scan SQL value of type %T as email.AddressSet", value)
	}
	c, err := readAddressCollection(text, "AddressSet")
	if err != nil {
		return err
	}
	set.setCollection(c)
	return nil
}

// setCollection stores the address collection c
// read by Scan or UnmarshalJSON.
func (set *AddressSet) setCollection(c addressCollection) {
	switch {
	case c.null:
		*set = nil
	case c.entries != nil:
		*set = make(AddressSet, len(c.entries))
		for _, addr := range c.entries {
			(*set)[Address(addr)] = struct{}{}
		}
	default:
		entries := SplitAddressList(c.list)
		if entries == nil {
			*set = nil
			return
		}
		*set = MakeAddressSet(entries...)
	}
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// It reads a JSON array of strings, the elements kept exactly,
// so that MarshalJSON and UnmarshalJSON round-trip every set, also with
// invalid addresses. It also reads a JSON object with the addresses as
// keys, which MarshalJSON wrote before, and a JSON string with an address
// list split with SplitAddressList. JSON null and "" set the nil set,
// [] and {} the empty set, and any other JSON value is an error.
func (set *AddressSet) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		*set = nil
		return nil
	}
	c, err := unmarshalJSONAddressCollection(data, "AddressSet")
	if err != nil {
		return err
	}
	set.setCollection(c)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler, used by encoding/xml,
// by splitting text with SplitAddressList, without normalizing or
// validating the addresses. An empty text sets the nil set.
func (set *AddressSet) UnmarshalText(text []byte) error {
	entries := SplitAddressList(string(text))
	if entries == nil {
		*set = nil
		return nil
	}
	*set = MakeAddressSet(entries...)
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface
// by writing the addresses as stored in sorted order as PostgreSQL array,
// with the surrounding whitespace of valid addresses trimmed and duplicates
// that this creates removed, or an error if an address is invalid.
// A nil set is written as SQL NULL, an empty set as "{}".
func (set AddressSet) Value() (driver.Value, error) {
	if set == nil {
		return nil, nil
	}
	addrs, err := set.writableAddresses(false)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return "{}", nil
	}
	return (*notnull.StringArray)(&addrs).Value()
}

// writableAddresses returns the addresses of the set for Value and
// MarshalText: sorted, with the surrounding whitespace trimmed and the
// duplicates that this creates removed, or an error for an address that is
// invalid, or if asListEntry is true, would not read back as one entry of
// an address list with SplitAddressList.
func (set AddressSet) writableAddresses(asListEntry bool) ([]string, error) {
	addrs := make([]string, 0, len(set))
	for _, addr := range set.Sorted() {
		trimmed := strutil.TrimSpace(addr)
		if err := trimmed.Validate(); err != nil {
			return nil, errs.Errorf("invalid address %q in email.AddressSet: %w", string(addr), err)
		}
		if asListEntry {
			if entries := SplitAddressList(string(trimmed)); len(entries) != 1 || entries[0] != trimmed {
				return nil, errs.Errorf("address %q in email.AddressSet is no single entry of an address list", string(addr))
			}
		}
		addrs = append(addrs, string(trimmed))
	}
	slices.Sort(addrs)
	return slices.Compact(addrs), nil
}

// MarshalJSON implements encoding/json.Marshaler by writing the addresses
// as stored as sorted JSON array, also invalid ones, so that MarshalJSON
// and UnmarshalJSON round-trip every set. A nil set is JSON null,
// an empty set the empty array.
func (set AddressSet) MarshalJSON() ([]byte, error) {
	if set == nil {
		return []byte(`null`), nil
	}
	addrs := set.Strings()
	if addrs == nil {
		addrs = []string{}
	}
	return json.Marshal(addrs)
}

// MarshalText implements encoding.TextMarshaler, used by encoding/xml, by
// writing the addresses as stored in sorted order as address list, with
// the surrounding whitespace of valid addresses trimmed, or an error if an
// address is invalid or would not read back as one entry of the list.
// A nil or empty set is written as "".
func (set AddressSet) MarshalText() ([]byte, error) {
	addrs, err := set.writableAddresses(true)
	if err != nil {
		return nil, err
	}
	return []byte(strings.Join(addrs, ", ")), nil
}
