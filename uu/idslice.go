package uu

import (
	"bytes"
	"database/sql/driver"
	"fmt"
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
	}
	return IDSliceFromStrings(strings.Split(str, ","))
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

// MakeSet returns an IDSet with the IDs from the IDSlice.
func (s IDSlice) MakeSet() IDSet {
	set := make(IDSet, len(s))
	set.AddSlice(s)
	return set
}

// String implements the fmt.Stringer interface.
func (s IDSlice) String() string {
	return "[" + strings.Join(s.Strings(), ",") + "]"
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
	sort.Slice(s, func(i, j int) bool { return IDCompare(s[i], s[j]) < 0 })
}

// SortedClone returns a sorted clone of the slice.
func (s IDSlice) SortedClone() IDSlice {
	clone := s.Clone()
	clone.Sort()
	return clone
}

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
	if src == nil {
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

	var b strings.Builder
	b.WriteByte('{')
	for i, id := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(id.String())
		b.WriteByte('"')
	}
	b.WriteByte('}')

	return b.String(), nil
}

// MarshalJSON implements encoding/json.Marshaler
func (s IDSlice) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	var b bytes.Buffer
	b.WriteByte('[')
	for i, id := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(id.String())
		b.WriteByte('"')
	}
	b.WriteByte(']')

	return b.Bytes(), nil
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
