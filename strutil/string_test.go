package strutil

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSlice(t *testing.T) {
	type StringType string

	out := ConvertSlice[StringType]([]string{"a", "b", "c"})
	require.Equal(t, []StringType{"a", "b", "c"}, out)

	out = ConvertSlice[StringType]([]string(nil))
	require.Nil(t, out)
}

func TestEmptyStringToNil(t *testing.T) {
	// The result is used for SQL arguments and JSON values where an empty
	// string must become NULL. It has to be an untyped nil interface,
	// a typed nil like (*string)(nil) would not compare equal to nil.
	assert.True(t, EmptyStringToNil("") == nil, "empty string must be an untyped nil")
	assert.Equal(t, "a", EmptyStringToNil("a"))
	// Whitespace is not trimmed, only the exact empty string is NULL
	assert.Equal(t, " ", EmptyStringToNil(" "))
}

func TestStringToPtrEmptyToNil(t *testing.T) {
	assert.Nil(t, StringToPtrEmptyToNil(""))
	// Whitespace is not trimmed, only the exact empty string is nil
	if p := StringToPtrEmptyToNil(" "); assert.NotNil(t, p) {
		assert.Equal(t, " ", *p)
	}
	if p := StringToPtrEmptyToNil("ä"); assert.NotNil(t, p) {
		assert.Equal(t, "ä", *p)
	}
}

func TestIndexInStrings(t *testing.T) {
	tests := []struct {
		name  string
		str   string
		slice []string
		want  int
	}{
		{name: "nil slice", str: "a", slice: nil, want: -1},
		{name: "empty string in nil slice", str: "", slice: nil, want: -1},
		{name: "not found", str: "x", slice: []string{"a", "b"}, want: -1},
		{name: "first", str: "a", slice: []string{"a", "b"}, want: 0},
		{name: "last", str: "b", slice: []string{"a", "b"}, want: 1},
		// The first match wins so the result can be used to look up
		// parallel slices without depending on later duplicates
		{name: "duplicates return first", str: "b", slice: []string{"a", "b", "b"}, want: 1},
		{name: "empty string element", str: "", slice: []string{"a", ""}, want: 1},
		// Comparison is exact, not case or Unicode normalization insensitive
		{name: "case sensitive", str: "A", slice: []string{"a"}, want: -1},
		{name: "multi-byte", str: "ü", slice: []string{"u", "ü"}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IndexInStrings(tt.str, tt.slice))
		})
	}
}

func TestDerefPtr(t *testing.T) {
	// A nil pointer must be safe to dereference and read as empty string
	assert.Equal(t, "", DerefPtr(nil))
	assert.Equal(t, "ä", DerefPtr(new("ä")))
}

func TestEqualPtrOrString(t *testing.T) {
	s := "a"
	tests := []struct {
		name string
		a, b *string
		want bool
	}{
		{name: "both nil", a: nil, b: nil, want: true},
		{name: "same pointer", a: &s, b: &s, want: true},
		{name: "different pointers to equal strings", a: new("a"), b: new("a"), want: true},
		{name: "different strings", a: new("a"), b: new("b"), want: false},
		// nil is not equal to a pointer to the empty string,
		// so NULL and "" stay distinguishable
		{name: "nil vs empty", a: nil, b: new(""), want: false},
		{name: "empty vs nil", a: new(""), b: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, EqualPtrOrString(tt.a, tt.b))
			assert.Equal(t, tt.want, EqualPtrOrString(tt.b, tt.a), "must be symmetric")
		})
	}
}

func TestJoin(t *testing.T) {
	// Join must behave exactly like strings.Join, just for any string type
	type Name string
	tests := []struct {
		name  string
		elems []Name
		sep   string
	}{
		{name: "nil", elems: nil, sep: ","},
		{name: "empty", elems: []Name{}, sep: ","},
		{name: "single", elems: []Name{"a"}, sep: ","},
		{name: "two", elems: []Name{"a", "b"}, sep: ", "},
		{name: "empty elements", elems: []Name{"", "", ""}, sep: "-"},
		{name: "empty separator", elems: []Name{"a", "b", "c"}, sep: ""},
		{name: "multi-byte", elems: []Name{"Grüße", "世界"}, sep: " → "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := strings.Join(ConvertSlice[string](tt.elems), tt.sep)
			assert.Equal(t, want, Join(tt.elems, tt.sep))
		})
	}
}

func TestCompareStringsShorterFirst(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{a: "", b: "", want: 0},
		{a: "a", b: "a", want: 0},
		{a: "", b: "a", want: -1},
		{a: "b", b: "aa", want: -1}, // shorter first even if lexicographically greater
		{a: "aa", b: "b", want: +1},
		{a: "ab", b: "ac", want: -1}, // equal length: lexicographic
		{a: "ac", b: "ab", want: +1},
		// Length is measured in bytes, not runes:
		// "ä" is 2 bytes and therefore longer than "z"
		{a: "z", b: "ä", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			assert.Equal(t, tt.want, CompareStringsShorterFirst(tt.a, tt.b))
			assert.Equal(t, -tt.want, CompareStringsShorterFirst(tt.b, tt.a), "must be antisymmetric")
		})
	}

	// Usable as slices.SortFunc comparison for any string type
	type Code string
	codes := []Code{"ccc", "b", "aa", "a", "bb"}
	slices.SortFunc(codes, CompareStringsShorterFirst)
	assert.Equal(t, []Code{"a", "b", "aa", "bb", "ccc"}, codes)
}
