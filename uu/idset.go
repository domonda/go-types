package uu

import (
	"database/sql/driver"
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
	return IDSlice(ids).MakeSet()
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
	return MakeIDSetFromStrings(strings.Split(str, ","))
}

// String implements the fmt.Stringer interface.
func (s IDSet) String() string {
	return "set" + s.SortedSlice().String()
}

// PrettyString returns s.SortedSlice().PrettyString().
// Implements pretty.Stringer.
func (s IDSet) PrettyString() string {
	return s.SortedSlice().PrettyString()
}

// GetOne returns a random ID from the set or IDNil if the set is empty.
// Most useful to get the only ID in a set of size one.
func (s IDSet) GetOne() ID {
	for id := range s {
		return id
	}
	return IDNil
}

func (s IDSet) Slice() IDSlice {
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

func (s IDSet) SortedSlice() IDSlice {
	sl := s.Slice()
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

func (s IDSet) Add(id ID) {
	s[id] = struct{}{}
}

func (s IDSet) Contains(id ID) bool {
	_, has := s[id]
	return has
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
	var idSlice IDSlice
	err := idSlice.Scan(value)
	if err != nil {
		return err
	}
	*s = idSlice.MakeSet()
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface
// with the nil map value used as SQL NULL
func (s IDSet) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	return s.SortedSlice().Value()
}

// MarshalJSON implements encoding/json.Marshaler
func (s IDSet) MarshalJSON() ([]byte, error) {
	return s.SortedSlice().MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler
// Id does assign a new IDSet to *set instead of modifying the existing map,
// so it can be used with uninitialized IDSet variable.
func (s *IDSet) UnmarshalJSON(data []byte) error {
	var idSlice IDSlice
	err := idSlice.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	*s = idSlice.MakeSet()
	return nil
}
