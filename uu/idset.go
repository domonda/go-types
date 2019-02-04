package uu

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// IDSet is a set of uu.IDs.
// It is a map[ID]struct{} underneath.
// Implements the database/sql.Scanner and database/sql/driver.Valuer interfaces
// with the nil map value used as SQL NULL
type IDSet map[ID]struct{}

// String implements the fmt.Stringer interface.
func (set IDSet) String() string {
	return fmt.Sprintf("set%v", set.SortedSlice())
}

// GetOne returns a random ID from the set or Nil if the set is empty.
// Most useful to get the only ID in a set of size one.
func (set IDSet) GetOne() ID {
	for id := range set {
		return id
	}
	return IDNil
}

func (set IDSet) Slice() IDSlice {
	s := make(IDSlice, len(set))
	i := 0
	for id := range set {
		s[i] = id
		i++
	}
	return s
}

func (set IDSet) SortedSlice() IDSlice {
	s := set.Slice()
	s.Sort()
	return s
}

func (set IDSet) AddSlice(s IDSlice) {
	for _, id := range s {
		set[id] = struct{}{}
	}
}

func (set IDSet) AddSet(other IDSet) {
	for id := range other {
		set[id] = struct{}{}
	}
}

func (set IDSet) Add(id ID) {
	set[id] = struct{}{}
}

func (set IDSet) Contains(id ID) bool {
	_, has := set[id]
	return has
}

func (set IDSet) Delete(id ID) {
	delete(set, id)
}

func (set IDSet) DeleteAll() {
	for id := range set {
		delete(set, id)
	}
}

func (set IDSet) DeleteSlice(s IDSlice) {
	for _, id := range s {
		delete(set, id)
	}
}

func (set IDSet) DeleteSet(other IDSet) {
	for id := range other {
		delete(set, id)
	}
}

func (set IDSet) Clone() IDSet {
	clone := make(IDSet)
	clone.AddSet(set)
	return clone
}

func (set IDSet) Diff(other IDSet) IDSet {
	diff := make(IDSet)
	for id := range set {
		if !other.Contains(id) {
			diff.Add(id)
		}
	}
	for id := range other {
		if !set.Contains(id) {
			diff.Add(id)
		}
	}
	return diff
}

// Scan implements the database/sql.Scanner interface
// with the nil map value used as SQL NULL.
// Does *set = make(Set) if *set == nil
// so it can be used with an not initialized Set variable
func (set *IDSet) Scan(value interface{}) (err error) {
	switch x := value.(type) {
	case string:
		return set.scanBytes([]byte(x))

	case []byte:
		return set.scanBytes(x)

	case nil:
		*set = nil
		return nil
	}

	return errors.Errorf("can't scan value '%#v' of type %T as uu.IDSet", value, value)
}

func (set *IDSet) scanBytes(src []byte) (err error) {
	if src == nil {
		*set = nil
		return nil
	}

	if len(src) < 2 || src[0] != '{' || src[len(src)-1] != '}' {
		return errors.Errorf("can't parse %#v as uu.IDSet", string(src))
	}

	ids := make(IDSlice, 0, 16)

	elements := bytes.Split(src[1:len(src)-1], []byte{','})
	for _, elem := range elements {
		elem = bytes.Trim(elem, `'"`)
		id, err := IDFromString(string(elem))
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	if *set == nil {
		*set = make(IDSet)
	} else {
		set.DeleteAll()
	}
	set.AddSlice(ids)

	return nil
}

// Value implements the driver database/sql/driver.Valuer interface
// with the nil map value used as SQL NULL
func (set IDSet) Value() (driver.Value, error) {
	if set == nil {
		return nil, nil
	}

	var b strings.Builder
	b.WriteByte('{')
	for i, id := range set.SortedSlice() {
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
