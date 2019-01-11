package strutil

import "sort"

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
