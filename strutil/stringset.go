package strutil

import (
	"sort"
	"strings"
)

type StringSet map[string]struct{}

func NewStringSet(strings ...string) StringSet {
	set := make(StringSet)
	for _, s := range strings {
		set[s] = struct{}{}
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
	clone := make(StringSet)
	clone.AddSet(set)
	return clone
}

func (set StringSet) Diff(other StringSet) StringSet {
	diff := make(StringSet)
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
