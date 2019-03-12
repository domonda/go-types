package uu

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// IDSlice is a slice of uu.IDs.
// It is a []ID underneath.
// Implements the database/sql.Scanner and database/sql/driver.Valuer interfaces.
// Implements the encoding/json.Marshaler and Unmarshaler interfaces.
// with the nil slice value used as SQL NULL and JSON null.
type IDSlice []ID

func SliceFromStrings(strs []string) (s IDSlice, err error) {
	if strs == nil {
		return nil, nil
	}

	s = make(IDSlice, len(strs))
	for i, str := range strs {
		s[i], err = IDFromString(str)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func MustSliceFromStrings(strs ...string) IDSlice {
	s, err := SliceFromStrings(strs)
	if err != nil {
		panic(err)
	}
	return s
}

func (s IDSlice) Set() IDSet {
	set := make(IDSet, len(s))
	set.AddSlice(s)
	return set
}

// String implements the fmt.Stringer interface.
func (s IDSlice) String() string {
	return "[" + strings.Join(s.Strings(), ",") + "]"
}

func (s IDSlice) Strings() []string {
	ss := make([]string, len(s))
	for i, id := range s {
		ss[i] = id.String()
	}
	return ss
}

func IDCompareLess(a, b ID) bool {
	return bytes.Compare(a[:], b[:]) < 0
}

func (s IDSlice) Sort() {
	sort.Slice(s, func(i, j int) bool { return IDCompareLess(s[i], s[j]) })
}

func (s IDSlice) SortedClone() IDSlice {
	clone := s.Clone()
	clone.Sort()
	return clone
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

func (s IDSlice) Clone() IDSlice {
	clone := make(IDSlice, len(s))
	copy(clone, s)
	return clone
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

	return errors.Errorf("can't scan value '%#v' of type %T as uu.IDSlice", value, value)
}

func (s *IDSlice) scanBytes(src []byte) (err error) {
	if src == nil {
		*s = nil
		return nil
	}

	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("can't parse '%s' as uu.IDSlice", string(src))
	}

	ids := make([]ID, 0, 16)

	elements := bytes.Split(src[1:len(src)-1], []byte{','})
	for _, elem := range elements {
		elem = bytes.Trim(elem, `'"`)
		id, err := IDFromString(string(elem))
		if err != nil {
			return err
		}
		ids = append(ids, id)
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
func (s *IDSlice) UnmarshalJSON(data []byte) (err error) {
	if data == nil || string(data) == "null" {
		*s = nil
		return nil
	}

	// Can be unmarshaled directly because the slice elements are json.Unmarshaler
	return json.Unmarshal(data, s)

	// var strs []string
	// err = json.Unmarshal(data, &strs)
	// if err != nil {
	// 	return err
	// }
	// newS, err := SliceFromStrings(strs)
	// if err != nil {
	// 	return err
	// }

	// *s = newS
	// return nil
}
