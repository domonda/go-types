package uu

import (
	"bytes"
	"database/sql/driver"
	"io"
	"strings"
)

// IDSet is a set of uu.IDs.
// It is a map[ID]struct{} underneath.
// Implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// with the nil map value used as SQL NULL
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
		s.Add(id)
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
	}
	if str == "" {
		return nil, nil
	}
	return MakeIDSetFromStrings(strings.Split(str, ","))
}

// IDSetMust converts the passed values to an IDSet
// or panics if that's not possible or an ID is not valid.
// Supported types are string, []byte, [16]byte,
// ID, NullableID, and nil.
// Returns nil if zero values are passed.
func IDSetMust(vals ...interface{}) IDSet {
	if len(vals) == 0 {
		return nil
	}
	s := make(IDSet, len(vals))
	for _, val := range vals {
		s.Add(IDMust(val))
	}
	return s
}

// String implements the fmt.Stringer interface.
func (s IDSet) String() string {
	return "set" + s.AsSortedSlice().String()
}

// PrettyPrint using s.AsSortedSlice().PrettyPrint(w).
// Implements pretty.Stringer.
func (s IDSet) PrettyPrint(w io.Writer) {
	s.AsSortedSlice().PrettyPrint(w)
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

// Len returns the length of the IDSet.
func (s IDSet) Len() int {
	return len(s)
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

func (set IDSet) AddSlice(s IDSlice) {
	for _, id := range s {
		set[id] = struct{}{}
	}
}

func (s IDSet) AddSet(other IDSet) {
	for id := range other {
		s[id] = struct{}{}
	}
}

func (s IDSet) AddIDs(ids IDs) {
	ids.ForEach(func(id ID) error {
		s[id] = struct{}{}
		return nil
	})
}

func (s IDSet) Add(id ID) {
	s[id] = struct{}{}
}

// Contains returns true if the set contains the passed id.
// It is valid to call this method on a nil IDSet.
func (s IDSet) Contains(id ID) bool {
	if s == nil {
		return false
	}
	_, ok := s[id]
	return ok
}

func (s IDSet) Delete(id ID) {
	delete(s, id)
}

func (s IDSet) DeleteAll() {
	for id := range s {
		delete(s, id)
	}
}

func (s IDSet) DeleteSlice(sl IDSlice) {
	for _, id := range sl {
		delete(s, id)
	}
}

func (s IDSet) DeleteSet(other IDSet) {
	for id := range other {
		delete(s, id)
	}
}

func (s IDSet) Clone() IDSet {
	if s == nil {
		return nil
	}
	clone := make(IDSet)
	clone.AddSet(s)
	return clone
}

func (s IDSet) Diff(other IDSet) IDSet {
	diff := make(IDSet)
	for id := range s {
		if !other.Contains(id) {
			diff.Add(id)
		}
	}
	for id := range other {
		if !s.Contains(id) {
			diff.Add(id)
		}
	}
	return diff
}

func (s IDSet) Equal(other IDSet) bool {
	if len(s) != len(other) {
		return false
	}
	for id := range s {
		if !other.Contains(id) {
			return false
		}
	}
	return true
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
func (s *IDSet) Scan(value interface{}) error {
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
	return s.AsSortedSlice().Value()
}

// MarshalJSON implements encoding/json.Marshaler
func (s IDSet) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	return s.AsSortedSlice().MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// Id does assign a new IDSet to *set instead of modifying the existing map,
// so it can be used with uninitialized IDSet variable.
func (s *IDSet) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*s = nil
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
