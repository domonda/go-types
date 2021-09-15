package strutil

import (
	"sort"
	"strings"
)

type StringSet map[string]struct{}

func NewStringSet(strings ...string) StringSet {
	set := make(StringSet, len(strings))
	for _, s := range strings {
		set[s] = struct{}{}
	}
	return set
}

func NewStringSetMergeSlices(slices ...[]string) StringSet {
	cap := 0
	for _, strings := range slices {
		cap += len(strings)
	}
	set := make(StringSet, cap)
	for _, strings := range slices {
		for _, s := range strings {
			set[s] = struct{}{}
		}
	}
	return set
}

func (set StringSet) Sorted() (s []string) {
	if count := len(set); count > 0 {
		s = make([]string, 0, count)
		for str := range set {
			s = append(s, str)
		}
		sort.Strings(s)
	}
	return s
}

// String implements the fmt.Stringer interface.
func (set StringSet) String() string {
	if len(set) == 0 {
		return `[]`
	}
	return `["` + strings.Join(set.Sorted(), `", "`) + `"]`
}

func (set StringSet) AddSlice(s []string) {
	for _, str := range s {
		set[str] = struct{}{}
	}
}

func (set StringSet) AddSet(other StringSet) {
	for str := range other {
		set[str] = struct{}{}
	}
}

func (set StringSet) Add(str string) {
	set[str] = struct{}{}
}

func (set StringSet) Contains(str string) bool {
	_, has := set[str]
	return has
}

func (set StringSet) ContainsAny(strs ...string) bool {
	for _, str := range strs {
		if set.Contains(str) {
			return true
		}
	}
	return false
}

// StringContainsAnyOfSet returns true if the passed string
// contains any of the strings of the StringSet.
func (set StringSet) StringContainsAnyOfSet(str string) bool {
	for s := range set {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}

func (set StringSet) Delete(str string) {
	delete(set, str)
}

func (set StringSet) DeleteAll() {
	for str := range set {
		delete(set, str)
	}
}

func (set StringSet) DeleteSlice(s []string) {
	for _, str := range s {
		delete(set, str)
	}
}

func (set StringSet) DeleteSet(other StringSet) {
	for str := range other {
		delete(set, str)
	}
}

func (set StringSet) Clone() StringSet {
	clone := make(StringSet, len(set))
	clone.AddSet(set)
	return clone
}

func (set StringSet) Diff(other StringSet) StringSet {
	diff := make(StringSet, len(set))
	for str := range set {
		if !other.Contains(str) {
			diff.Add(str)
		}
	}
	for str := range other {
		if !set.Contains(str) {
			diff.Add(str)
		}
	}
	return diff
}

func (set StringSet) Equal(other StringSet) bool {
	if len(set) != len(other) {
		return false
	}
	for str := range set {
		if !other.Contains(str) {
			return false
		}
	}
	return true
}
