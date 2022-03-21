package types

import (
	"reflect"
	"testing"
)

func TestSetToSortedSlice(t *testing.T) {
	{
		set := map[int]struct{}{
			3: {},
			2: {},
			1: {},
		}
		want := []int{1, 2, 3}
		got := SetToSortedSlice(set)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SetToSortedSlice() = %#v, want %#v", got, want)
		}
	}
	{
		set := map[string]struct{}{
			"3": {},
			"2": {},
			"1": {},
		}
		want := []string{"1", "2", "3"}
		got := SetToSortedSlice(set)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SetToSortedSlice() = %#v, want %#v", got, want)
		}
	}
}
