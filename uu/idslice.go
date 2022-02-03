package uu

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"strings"
)

// IDSlice is a slice of uu.IDs.
// It is a []ID underneath.
// Implements the database/sql.Scanner and database/sql/driver.Valuer interfaces.
// Implements the encoding/json.Marshaler and Unmarshaler interfaces.
// with the nil slice value used as SQL NULL and JSON null.
type IDSlice []ID

// IDSliceFromString parses a string created with IDSlice.String()
func IDSliceFromString(str string) (IDSlice, error) {
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
		strs[i] = strings.TrimSpace(s)
	}
	return IDSliceFromStrings(strs)
}

// IDSliceFromStrings parses an IDSlice from strings
func IDSliceFromStrings(strs []string) (IDSlice, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	s := make(IDSlice, len(strs))
	for i, str := range strs {
		var err error
		s[i], err = IDFromString(str)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// IDSliceMustFromStrings parses an IDSlice from strings
// and panics in case of an error.
func IDSliceMustFromStrings(strs ...string) IDSlice {
	s, err := IDSliceFromStrings(strs)
	if err != nil {
		panic(err)
	}
	return s
}

// IDSliceMust converts the passed values to an IDSlice
// or panics if that's not possible or an ID is not valid.
// Supported types are string, []byte, [16]byte,
// ID, NullableID, and nil.
// Returns nil if zero values are passed.
func IDSliceMust(vals ...interface{}) IDSlice {
	if len(vals) == 0 {
		return nil
	}
	s := make(IDSlice, len(vals))
	for i, val := range vals {
		s[i] = IDMust(val)
	}
	return s
}

// AsSet returns the IDs of the slice as a IDSet.
func (s IDSlice) AsSet() IDSet {
	set := make(IDSet, len(s))
	set.AddSlice(s)
	return set
}

// AsSlice returns s unchanged to implement the IDs interface.
func (s IDSlice) AsSlice() IDSlice {
	return s
}

// ForEach calls the passed function for each ID.
// Any error from the callback function is returned
// by ForEach immediatly.
// Returning a sentinel error is a way to stop the loop
// with a known cause that might not be a real error.
func (s IDSlice) ForEach(callback func(ID) error) error {
	for _, id := range s {
		err := callback(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (s IDSlice) String() string {
	return "[" + strings.Join(s.Strings(), ",") + "]"
}

// PrettyPrint using IDSlice.String().
// Implements the pretty.Printable interface.
func (s IDSlice) PrettyPrint(w io.Writer) {
	w.Write([]byte(s.String())) //#nosec G104 -- go-pretty does not check write errors
}

// Strings returns a slice with all IDs converted to strings
func (s IDSlice) Strings() []string {
	if len(s) == 0 {
		return nil
	}
	ss := make([]string, len(s))
	for i, id := range s {
		ss[i] = id.String()
	}
	return ss
}

// Sort the slice in place.
func (s IDSlice) Sort() {
	sort.Sort(s)
}

// SortedClone returns a sorted clone of the slice.
func (s IDSlice) SortedClone() IDSlice {
	c := s.Clone()
	c.Sort()
	return c
}

// Len is the number of elements in the collection.
// One of the methods to implement sort.Interface.
func (s IDSlice) Len() int { return len(s) }

// Less reports whether the element with
// index i should sort before the element with index j.
// One of the methods to implement sort.Interface.
func (s IDSlice) Less(i, j int) bool { return s[i].Less(s[j]) }

// Swap swaps the elements with indexes i and j.
// One of the methods to implement sort.Interface.
func (s IDSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// IndexOf returns the index of the first occurrence of id
// in the slice, or -1 if id was not found.
func (s IDSlice) IndexOf(id ID) int {
	for i, curr := range s {
		if curr == id {
			return i
		}
	}
	return -1
}

func (s IDSlice) Contains(id ID) bool {
	for _, curr := range s {
		if curr == id {
			return true
		}
	}
	return false
}

func (s IDSlice) ContainsAny(other IDSlice) bool {
	for _, curr := range s {
		for _, id := range other {
			if curr == id {
				return true
			}
		}
	}
	return false
}

func (s IDSlice) Equal(other IDSlice) bool {
	if len(s) != len(other) {
		return false
	}
	for i := range s {
		if s[i] != other[i] {
			return false
		}
	}
	return true
}

func (s IDSlice) ContainsAnyFromSet(set IDSet) bool {
	for _, id := range s {
		if set.Contains(id) {
			return true
		}
	}
	return false
}

// RemoveFirst removes the first occurrence of id from the slice
// and returns its index or -1 if id was not found in the slice.
func (s *IDSlice) RemoveFirst(id ID) int {
	index := s.IndexOf(id)
	if index != -1 {
		s.RemoveAt(index)
	}
	return index
}

// RemoveAll removes the all occurrences of id from the slice
// and returns the count of removals.
func (s *IDSlice) RemoveAll(id ID) (count int) {
	for i := 0; i < len(*s); i++ {
		if (*s)[i] == id {
			*s = append((*s)[:i], (*s)[i+1:]...)
			count++
			i--
		}
	}
	return count
}

// RemoveAt removes the slice element at the given index.
// Will panic if the index is out of range.
func (s *IDSlice) RemoveAt(index int) {
	*s = append((*s)[:index], (*s)[index+1:]...)
}

// Clone returns a copy of the slice.
func (s IDSlice) Clone() IDSlice {
	if s == nil {
		return nil
	}
	clone := make(IDSlice, len(s))
	copy(clone, s)
	return clone
}

// MarshalText implements the encoding.TextMarshaler interface
func (s IDSlice) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
// by calling IDSliceFromString.
func (s *IDSlice) UnmarshalText(text []byte) error {
	parsed, err := IDSliceFromString(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// Scan implements the database/sql.Scanner interface
// with the nil map value used as SQL NULL.
// Does *s = make(Slice) if *s == nil
// so it can be used with an not initialized Slice variable
func (s *IDSlice) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case string:
		return s.scanBytes([]byte(x))

	case []byte:
		return s.scanBytes(x)

	case nil:
		*s = nil
		return nil
	}

	return fmt.Errorf("can't scan value '%#v' of type %T as uu.IDSlice", value, value)
}

func (s *IDSlice) scanBytes(src []byte) (err error) {
	if len(src) == 0 {
		*s = nil
		return nil
	}

	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		return fmt.Errorf("can't parse %q as uu.IDSlice", string(src))
	}

	ids := make(IDSlice, 0)

	if len(src) > 2 {
		elements := bytes.Split(src[1:len(src)-1], []byte{','})
		for _, elem := range elements {
			id, err := IDFromBytes(bytes.Trim(elem, `'"`))
			if err != nil {
				return err
			}
			ids = append(ids, id)
		}
	}

	*s = ids
	return nil
}

// Value implements the driver database/sql/driver.Valuer interface
// with the nil map value used as SQL NULL
func (s IDSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	l := len(s)
	if l == 0 {
		return "{}", nil
	}

	b := strings.Builder{}
	b.Grow(2 + (36 + 2) + (l-1)*(1+36+2))

	b.WriteString(`{"`)
	b.Write(s[0].StringBytes())
	for i := 1; i < l; i++ {
		b.WriteString(`","`)
		b.Write(s[i].StringBytes())
	}
	b.WriteString(`"}`)

	return b.String(), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (s IDSlice) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	l := len(s)
	if l == 0 {
		return []byte("[]"), nil
	}

	b := make([]byte, 0, 2+(36+2)+(l-1)*(1+36+2))

	b = append(b, `["`...)
	b = append(b, s[0].StringBytes()...)
	for i := 1; i < l; i++ {
		b = append(b, `","`...)
		b = append(b, s[i].StringBytes()...)
	}
	b = append(b, `"]`...)

	return b, nil
}

// UnmarshalJSON implements encoding/json.Unmarshaler
func (s *IDSlice) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		*s = nil
		return nil
	}
	if len(data) < 2 || data[0] != '[' || data[len(data)-1] != ']' {
		return fmt.Errorf("can't parse as uu.IDSlice because not a JSON array: %s", data)
	}

	ids := make(IDSlice, 0)

	if len(data) > 2 {
		for i, next := 1, 1; i < len(data); i++ {
			if data[i] == ',' || i == len(data)-1 {
				str := bytes.TrimSpace(data[next:i])
				if len(str) < 2 || str[0] != '"' || str[len(str)-1] != '"' {
					return fmt.Errorf("can't parse as uu.IDSlice because not a JSON string array: %s", data)
				}
				id, err := IDFromBytes(str[1 : len(str)-1])
				if err != nil {
					return fmt.Errorf("error parsing uu.IDSlice from JSON: %w", err)
				}

				ids = append(ids, id)
				next = i + 1
			}
		}
	}

	*s = ids
	return nil
}
