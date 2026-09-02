package uu

import (
	"bytes"
	"database/sql/driver"
	"iter"
	"maps"
	"sort"
	"strings"

	"github.com/domonda/go-types/mapset"
	"github.com/domonda/go-types/strutil"
)

// IDSet is a set of uu.IDs.
// It is a map[ID]struct{} underneath.
// Implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// with the nil map value used as SQL NULL
//
// Value and Scan use the PostgreSQL array text format ({"id1","id2"}),
// see https://www.postgresql.org/docs/current/arrays.html. That format
// is understood by PostgreSQL and array-compatible databases such as
// CockroachDB and YugabyteDB; databases without a native array type
// (MySQL, MariaDB, SQLite, SQL Server, Oracle) are not supported.
type IDSet map[ID]struct{}

// MakeIDSet returns an IDSet with
// the optional passed ids added to it.
func MakeIDSet(ids ...ID) IDSet {
	return IDSlice(ids).AsSet()
}

// MakeIDSetFromStrings returns an IDSet with strs parsed as IDs
func MakeIDSetFromStrings(strs []string) (IDSet, error) {
	s := make(IDSet)
	for _, str := range strs {
		id, err := IDFromString(str)
		if err != nil {
			return nil, err
		}
		s.Insert(id)
	}
	return s, nil
}

// MakeIDSetMustFromStrings returns an IDSet with the
// passed strings as IDs or panics if there was an error.
func MakeIDSetMustFromStrings(strs ...string) IDSet {
	s, err := MakeIDSetFromStrings(strs)
	if err != nil {
		panic(err)
	}
	return s
}

// IDSetFromString parses a string created with IDSet.String()
func IDSetFromString(str string) (IDSet, error) {
	str = strings.TrimPrefix(str, "set")
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		str = str[1 : len(str)-1]
	} else if str == "null" || str == "NULL" {
		return nil, nil
	}
	if str == "" {
		return nil, nil
	}
	strs := strings.Split(str, ",")
	for i, s := range strs {
		strs[i] = strutil.TrimSpace(s)
	}
	return MakeIDSetFromStrings(strs)
}

// IDSetMust converts the passed values to an IDSet
// or panics if that's not possible or an ID is not valid.
// Returns nil if zero values are passed.
func IDSetMust[T IDSource](vals ...T) IDSet {
	if len(vals) == 0 {
		return nil
	}
	s := make(IDSet, len(vals))
	for _, val := range vals {
		s.Insert(IDMust(val))
	}
	return s
}

// String implements the fmt.Stringer interface.
func (s IDSet) String() string {
	return "set" + s.AsSortedSlice().String()
}

// Strings returns a slice with all IDs converted to strings
func (s IDSet) Strings() []string {
	if len(s) == 0 {
		return nil
	}
	ss := make([]string, len(s))
	i := 0
	for id := range s {
		ss[i] = id.String()
		i++
	}
	sort.Strings(ss)
	return ss
}

// PrettyString implements the pretty.Stringer interface
// using s.AsSortedSlice().PrettyString().
func (s IDSet) PrettyString() string {
	return s.AsSortedSlice().PrettyString()
}

// GetOne returns one ID in undefined order from the set
// or IDNil if the set is empty.
// Most useful to get the only ID in a set of size one.
func (s IDSet) GetOne() ID {
	for id := range s {
		return id
	}
	return IDNil
}

// AsSet returns s unchanged to implement the IDs interface.
func (s IDSet) AsSet() IDSet {
	return s
}

// AsSlice returns the IDs of the set as IDSlice with undefined order.
func (s IDSet) AsSlice() IDSlice {
	if len(s) == 0 {
		return nil
	}
	sl := make(IDSlice, len(s))
	i := 0
	for id := range s {
		sl[i] = id
		i++
	}
	return sl
}

// ForEach calls the passed function for each ID.
// Any error from the callback function is returned
// by ForEach immediatly.
// Returning a sentinel error is a way to stop the loop
// with a known cause that might not be a real error.
func (s IDSet) ForEach(callback func(ID) error) error {
	for id := range s {
		err := callback(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// AsSortedSlice returns the IDs of the set as sorted IDSlice.
func (s IDSet) AsSortedSlice() IDSlice {
	sl := s.AsSlice()
	sl.Sort()
	return sl
}

// AddSlice adds all IDs from s to the set.
//
// Deprecated: use set.InsertAll(slices.Values(s)).
func (set IDSet) AddSlice(s IDSlice) {
	for _, id := range s {
		set[id] = struct{}{}
	}
}

// AddSet adds all IDs from other into the set.
//
// Deprecated: use [IDSet.UnionWith].
func (s IDSet) AddSet(other IDSet) {
	for id := range other {
		s[id] = struct{}{}
	}
}

// AddIDs adds all IDs yielded by ids into the set.
func (s IDSet) AddIDs(ids IDs) {
	ids.ForEach(func(id ID) error {
		s[id] = struct{}{}
		return nil
	}) //#nosec G104 -- always returns nil
}

// Add inserts id into the set.
//
// Deprecated: use [IDSet.Insert] which additionally reports
// whether the set was changed.
func (s IDSet) Add(id ID) {
	s[id] = struct{}{}
}

// Insert adds id to the set and reports whether the set was changed.
// It panics if the set is nil and id is not already an element.
func (s IDSet) Insert(id ID) bool {
	return mapset.Insert(s, id)
}

// InsertAll adds all IDs yielded by seq to the set
// and reports whether the set was changed.
// It panics if the set is nil and seq yields an ID
// that is not already an element.
func (s IDSet) InsertAll(seq iter.Seq[ID]) bool {
	return mapset.InsertAll(s, seq)
}

// All returns an iterator over the IDs of the set in undefined order.
// It is valid to call this method on a nil IDSet.
func (s IDSet) All() iter.Seq[ID] {
	return mapset.All(s)
}

// Contains returns true if the set contains the passed id.
// It is valid to call this method on a nil IDSet.
func (s IDSet) Contains(id ID) bool {
	return mapset.Contains(s, id)
}

// ContainsAll reports whether the set contains all IDs yielded by seq.
// It returns true for an empty sequence and is valid to call on a nil IDSet.
func (s IDSet) ContainsAll(seq iter.Seq[ID]) bool {
	return mapset.ContainsAll(s, seq)
}

// Delete removes id from the set and reports whether the set was changed.
// It is valid to call this method on a nil IDSet.
func (s IDSet) Delete(id ID) bool {
	return mapset.Delete(s, id)
}

// DeleteAll removes all IDs yielded by seq from the set
// and reports whether the set was changed.
// It is valid to call this method on a nil IDSet.
func (s IDSet) DeleteAll(seq iter.Seq[ID]) bool {
	return mapset.DeleteAll(s, seq)
}

// DeleteFunc removes every ID for which del returns true
// and reports whether the set was changed.
// It is valid to call this method on a nil IDSet.
func (s IDSet) DeleteFunc(del func(ID) bool) bool {
	return mapset.DeleteFunc(s, del)
}

// Clear removes all IDs from the set, leaving it empty.
// It is valid to call this method on a nil IDSet.
func (s IDSet) Clear() {
	clear(s)
}

// DeleteSlice removes every ID in sl from the set.
//
// Deprecated: use set.DeleteAll(slices.Values(sl)).
func (s IDSet) DeleteSlice(sl IDSlice) {
	for _, id := range sl {
		delete(s, id)
	}
}

// DeleteSet removes every ID contained in other from the set.
//
// Deprecated: use [IDSet.DifferenceWith].
func (s IDSet) DeleteSet(other IDSet) {
	for id := range other {
		delete(s, id)
	}
}

// Clone returns a shallow copy of the set, or nil if the set is nil.
func (s IDSet) Clone() IDSet {
	if s == nil {
		return nil
	}
	return maps.Clone(s)
}

// Union returns a new IDSet with all IDs of s and other.
func (s IDSet) Union(other IDSet) IDSet {
	return mapset.Union(s, other)
}

// UnionWith adds all IDs of other to s.
// It panics if s is nil and other has an ID
// that is not already an element of s.
func (s IDSet) UnionWith(other IDSet) {
	mapset.UnionWith(s, other)
}

// Intersection returns a new IDSet with the IDs that are in both s and other.
func (s IDSet) Intersection(other IDSet) IDSet {
	return mapset.Intersection(s, other)
}

// IntersectionWith removes every ID from s that is not also in other.
// It is valid to call this method on a nil IDSet.
func (s IDSet) IntersectionWith(other IDSet) {
	mapset.IntersectionWith(s, other)
}

// Intersects reports whether s and other have at least one ID in common.
func (s IDSet) Intersects(other IDSet) bool {
	return mapset.Intersects(s, other)
}

// Difference returns a new IDSet with the IDs of s that are not in other.
//
// This replaces the former Diff method, which returned the symmetric
// difference and is now [IDSet.SymmetricDifference].
func (s IDSet) Difference(other IDSet) IDSet {
	return mapset.Difference(s, other)
}

// DifferenceWith removes every ID of other from s.
// It is valid to call this method on a nil IDSet.
func (s IDSet) DifferenceWith(other IDSet) {
	mapset.DifferenceWith(s, other)
}

// SymmetricDifference returns a new IDSet containing all IDs that are in s
// but not in other, and all IDs that are in other but not in s.
func (s IDSet) SymmetricDifference(other IDSet) IDSet {
	return mapset.SymmetricDifference(s, other)
}

// SymmetricDifferenceWith replaces the IDs of s with the IDs
// that are in exactly one of s and other.
// It panics if s is nil and other has an ID
// that is not already an element of s.
func (s IDSet) SymmetricDifferenceWith(other IDSet) {
	mapset.SymmetricDifferenceWith(s, other)
}

// Equal reports whether s and other contain exactly the same set of IDs.
func (s IDSet) Equal(other IDSet) bool {
	return mapset.Equal(s, other)
}

// Len returns the length of the IDSet.
func (s IDSet) Len() int {
	return len(s)
}

// IsEmpty returns true if the set is empty or nil.
func (s IDSet) IsEmpty() bool {
	return len(s) == 0
}

// IsNull implements the nullable.Nullable interface
// by returning true if the set is nil.
func (s IDSet) IsNull() bool {
	return s == nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (s IDSet) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *IDSet) UnmarshalText(text []byte) error {
	parsed, err := IDSetFromString(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// Scan implements the database/sql.Scanner interface
// with the nil map value used as SQL NULL.
// Id does assign a new IDSet to *set instead of modifying the existing map,
// so it can be used with uninitialized IDSet variable.
func (s *IDSet) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	var idSlice IDSlice
	err := idSlice.Scan(value)
	if err != nil {
		return err
	}
	*s = idSlice.AsSet()
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface
// with the nil map value used as SQL NULL
func (s IDSet) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	// AsSortedSlice returns a nil IDSlice for an empty set, whose Value is
	// SQL NULL, so the empty array has to be written here: a nil set is
	// NULL, an allocated empty set is the empty array, and Scan reads both
	// back to the state they came from.
	if len(s) == 0 {
		return "{}", nil
	}
	return s.AsSortedSlice().Value()
}

// MarshalJSON implements encoding/json.Marshaler
func (s IDSet) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	// Same reason as in Value: AsSortedSlice is nil for an empty set and
	// marshals as null, so the empty array has to be written here to keep
	// JSON null and the empty array distinct.
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	return s.AsSortedSlice().MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
// It assigns a new IDSet for a JSON array, so it can be used with an
// uninitialized IDSet variable, and empties an already allocated set
// in place for JSON null. It never assigns nil: a nil map panics when
// a key is set.
func (s *IDSet) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		// JSON null empties the set rather than setting it to nil.
		// A nil map panics when a key is set, so returning one here
		// would crash the next Insert on a freshly unmarshalled set.
		// [IDSlice.UnmarshalJSON] does return nil for null, because a
		// nil slice can be appended to. An allocated empty set still
		// marshals back to null.
		if *s == nil {
			*s = make(IDSet)
		} else {
			s.Clear()
		}
		return nil
	}
	var idSlice IDSlice
	err := idSlice.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	*s = idSlice.AsSet()
	return nil
}
